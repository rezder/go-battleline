// Package http contain the http server.
package http

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/http/games"
	"github.com/rezder/go-battleline/v2/http/login"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"time"
)

//Server a http server.
type Server struct {
	errServer   *ErrServer
	netListener *net.TCPListener
	clients     *Clients
	port        string
	doneCh      chan struct{}
	rootDir     string
}

//New creates a new Server.
func New(port string, archPokePort int, archAddr, rootDir string) (s *Server, err error) {
	s = new(Server)
	s.rootDir = rootDir
	s.port = port
	tcpPort := port
	if len(port) == 0 {
		tcpPort = ":80"
	}
	s.port = port
	laddr, err := net.ResolveTCPAddr("tcp", tcpPort) //TODO CHECK this look strange
	if err != nil {
		err = errors.Wrap(err, log.ErrNo(3)+"Resolve address")
		return s, err
	}
	var netListener *net.TCPListener
	netListener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		err = errors.Wrapf(err, "TCP listen on %v failed", laddr)
		return s, err
	}
	s.netListener = netListener
	var gameServer *games.Server
	gameServer, err = games.New(archPokePort, archAddr)
	if err != nil {
		return s, err
	}
	var clients *Clients
	clients, err = NewClients(gameServer)
	if err != nil {
		return s, err
	}
	s.clients = clients
	s.doneCh = make(chan struct{})
	s.errServer = NewErrServer(clients)

	return s, err
}

//Cancel cancels the server, if the server is created without errors
// it must be canceled or started.
func (s *Server) Cancel() (err error) {
	err = s.clients.Close()
	if err == nil {
		err = s.clients.CancelGameServer()
	} else {
		_ = s.clients.CancelGameServer()
	}
	return err
}

//Start starts the server.
func (s *Server) Start() {
	s.errServer.Start()
	s.clients.gameServer.Start(s.errServer.Ch())
	go start(s.errServer.Ch(), s.netListener, s.clients, s.doneCh, s.port, s.rootDir)
}

//Stop stops the server.
func (s *Server) Stop() {
	gameServer := s.clients.SetGameServer(nil) //Prevent new players
	if gameServer != nil {
		gameServer.Stop() //kick players out
	}
	log.Println(log.DebugMsg, "Close net listner.")
	err := s.netListener.Close()
	if err != nil {
		err = errors.Wrap(err, "Closing net listner")
		log.PrintErr(err)
	}
	<-s.doneCh
	log.Println(log.DebugMsg, "Recieve done from http server.")
	s.errServer.Stop() //Stop before saving so all errors may modify clients if need.
	err = s.clients.Close()
	if err != nil {
		err = errors.Wrap(err, "Closing clients failed")
		log.PrintErr(err)
	}
}

// Start the server.
func start(
	errCh chan<- error,
	netListener *net.TCPListener,
	clients *Clients,
	doneCh chan struct{},
	port string,
	rootDir string) {
	http.Handle("/post/login", &logInPostHandler{clients, errCh})
	http.Handle("/post/client", &clientPostHandler{clients, errCh})
	http.Handle("/in/gamews", *createWsHandler(clients, errCh))
	http.Handle("/ping", &pingHandler{clients, errCh})
	http.Handle("/", http.FileServer(http.Dir(rootDir)))

	server := &http.Server{Addr: "game.rezder.com" + port} //address is not used
	err := server.Serve(tcpKeepAliveListener{netListener})
	err = errors.Wrap(err, log.ErrNo(4)+"Http server serves")
	errCh <- err
	close(doneCh)
}

//createWsHandler create the websocket handler.
func createWsHandler(clients *Clients, errCh chan<- error) (server *websocket.Server) {
	wsHandshake := func(ws *websocket.Config, r *http.Request) (err error) {
		name, sid, err := getCookies(r)
		if err != nil {
			err = errors.Wrap(err, log.ErrNo(5)+"Websocket handshake")
			errCh <- err
			return err
		}
		ok, down := clients.VerifySid(name, sid)
		if down {
			if ok {
				clients.LogOut(name)
			}
			err = errors.New("Game server down")
		} else if !ok {
			err = errors.New(fmt.Sprintf(log.ErrNo(6)+"Failed session id! Ip: %v", r.RemoteAddr))
			errCh <- err
		}
		return err
	}
	wsHandler := func(ws *websocket.Conn) {
		joinedCh := make(chan *games.Player)
		name, sid, err := getCookies(ws.Request())
		ok, isJoined := clients.JoinGameServer(name, sid, ws, errCh, joinedCh)
		if ok {
			player := <-joinedCh
			player.Serve()
			err = ws.Close()
			clients.LogOut(name)
			if err != nil {
				err = errors.Wrap(err, log.ErrNo(8)+"Player is finish closing websocket")
				errCh <- err
			}
		} else {
			if !isJoined {
				clients.LogOut(name)
			}
			err = ws.Close()
			if err != nil {
				err = errors.Wrap(err, log.ErrNo(9)+"Player failed to join closing websocket failed")
				errCh <- err
			}
		}
	}
	server = &websocket.Server{Handler: wsHandler, Handshake: wsHandshake}
	return server
}

//getCookies extract the name and session cookies.
func getCookies(r *http.Request) (name string, sid string, err error) {
	nameC, err := r.Cookie("name")
	if err == nil {
		name = nameC.Value
		sidC, errC := r.Cookie("sid")
		if errC == nil {
			sid = sidC.Value
		} else {
			err = errors.New(fmt.Sprintf("Missing cookie! Ip: %v", r.RemoteAddr))
		}
	} else {
		err = errors.New(fmt.Sprintf("Missing cookie! Ip: %v", r.RemoteAddr))
	}
	return name, sid, err
}

//pingHandler handle check for game server is running.
type pingHandler struct {
	clients *Clients
	errCh   chan<- error
}

func (p *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isUp := true
	if p.clients.IsGameServerDown() {
		isUp = false
	}
	err := httpWrite(struct{ IsUp bool }{IsUp: isUp}, w)
	p.errCh <- err
}

//logInPostHandler the login post handler.
type logInPostHandler struct {
	clients *Clients
	errCh   chan<- error
}

func (g *logInPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	status, sid, err := g.clients.LogIn(name, pw)
	if err != nil {
		g.errCh <- err
		status = login.StatusAll.Err
	} else if status.IsOk() {
		setCookies(w, name, sid)
	} else if status.IsInValid() {
		err = errors.New(fmt.Sprintf(log.ErrNo(10)+"Login failed! %v Ip: %v", status, r.RemoteAddr))
		g.errCh <- err
	}
	err = httpWrite(struct{ LogInStatus login.Status }{LogInStatus: status}, w)
	if err != nil {
		g.errCh <- err
	}
}
func httpWrite(v interface{}, w http.ResponseWriter) (err error) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	cap := 0
	if len(js) > 100 {
		cap = len(js) - 100
	}
	log.Printf(log.DebugMsg, "Respond example: %v", string(js[cap:]))
	return err
}

//setCookies set the name and session id cookies.
func setCookies(w http.ResponseWriter, name string, sid string) {
	nameC := new(http.Cookie)
	nameC.Name = "name"
	nameC.Value = name
	nameC.Path = "/in"
	sidC := new(http.Cookie)
	sidC.Name = "sid"
	sidC.Value = sid
	sidC.Path = "/in"
	http.SetCookie(w, sidC)
	http.SetCookie(w, nameC)
}

//clientPostHandler the new client post handler.
type clientPostHandler struct {
	clients *Clients
	errCh   chan<- error
}

func (handler *clientPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	status, sid, err := handler.clients.AddNew(name, pw)
	if err != nil {
		status = login.StatusAll.Err
		handler.errCh <- errors.WithMessage(err, fmt.Sprintf("Failed to add client: %v", name))
	} else {
		if status.IsInValid() {
			errTxt := fmt.Sprintf("Data was submited with out our page validation! Ip: %v", r.RemoteAddr)
			errSize := errors.New(log.ErrNo(13) + errTxt)
			handler.errCh <- errSize
		} else if status.IsOk() {
			setCookies(w, name, sid)
		}
	}
	err = httpWrite(struct{ LogInStatus login.Status }{LogInStatus: status}, w)
	if err != nil {
		handler.errCh <- err
	}
}

//Add keep a live for 3 minute to a tcp handler.
//TODO CHECK why this is default but id do not know how good an idea this is.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return c, err
	}
	err = tc.SetKeepAlive(true)
	if err != nil {
		return c, err
	}
	err = tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
