// Package html contain the http server.
package html

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rezder/go-battleline/battserver/games"
	"github.com/rezder/go-battleline/battserver/players"
	"github.com/rezder/go-error/cerrors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	fileDown         = "html/pages/down.html"
	fileLogIn        = "html/pages/login.html"
	fileCreateClient = "html/pages/client.html"
)

//Server a http server.
type Server struct {
	errCh       chan<- error
	netListener *net.TCPListener
	clients     *Clients
	doneCh      chan struct{}
	port        string
	pages       *Pages
}

//New creates a new Server.
func New(errCh chan<- error, port string, save bool, saveDir string, archiverPort int) (s *Server, err error) {
	s = new(Server)
	pages := NewPages()
	pages.addFile(fileLogIn)
	pages.addFile(fileCreateClient)
	err = pages.load()
	if err != nil {
		return s, err
	}
	s.pages = pages
	s.port = port
	s.errCh = errCh
	dport := ":80"
	if len(port) != 0 {
		dport = port
	}
	laddr, err := net.ResolveTCPAddr("tcp", dport) //TODO CHECK this look strange
	if err != nil {
		err = cerrors.Wrap(err, 3, "Resolve address")
		return s, err
	}
	var netListener *net.TCPListener
	netListener, err = net.ListenTCP("tcp", laddr)
	//netListener, err = net.ListenTCP("tcp", ":8181")
	if err == nil {
		s.netListener = netListener
		var gameServer *games.Server
		gameServer, err = games.New(errCh, save, saveDir, archiverPort)
		if err == nil {
			var clients *Clients
			clients, err = loadClients(gameServer)
			if err == nil {
				s.clients = clients
				s.doneCh = make(chan struct{})
			}
		}
	}
	return s, err
}

//Start starts the server.
func (s *Server) Start() {
	s.clients.gameServer.Start()
	go start(s.errCh, s.netListener, s.clients, s.doneCh, s.port, s.pages)
}

//Stop stops the server.
func (s *Server) Stop() {
	gameServer := s.clients.SetGameServer(nil) //Prevent new players
	if gameServer != nil {
		gameServer.Stop() //kick players out
	}
	if cerrors.IsVerbose() {
		log.Println("Close net listner.")
	}
	err := s.netListener.Close()
	if err != nil {
		s.errCh <- err
	}
	<-s.doneCh
	if cerrors.IsVerbose() {
		log.Println("Recieve done from http server.")
	}
	err = s.clients.save() //no lock is used.
	if err != nil {
		s.errCh <- err
	}
}

// Start the server.
func start(
	errCh chan<- error,
	netListener *net.TCPListener,
	clients *Clients,
	doneCh chan struct{},
	port string,
	pages *Pages) {
	http.Handle("/", &logInHandler{clients, fileDown, fileLogIn})
	http.Handle("/client", &clientHandler{clients, fileDown, fileCreateClient})
	http.Handle("/in/game", &gameHandler{clients, errCh, port, fileDown})
	http.Handle("/form/login", &logInPostHandler{clients, pages, errCh, port, fileDown, fileLogIn})
	http.Handle("/form/client", &clientPostHandler{clients, pages, errCh, port, fileDown, fileCreateClient})
	http.Handle("/in/gamews", *createWsHandler(clients, errCh))
	http.Handle("/static/", http.FileServer(http.Dir("./html")))

	server := &http.Server{Addr: "game.rezder.com" + port} //address is not used
	err := server.Serve(tcpKeepAliveListener{netListener})
	err = cerrors.Wrap(err, 4, "Http server serves")
	errCh <- err
	close(doneCh)
}

//createWsHandler create the websocket handler.
func createWsHandler(clients *Clients, errCh chan<- error) (server *websocket.Server) {
	wsHandshake := func(ws *websocket.Config, r *http.Request) (err error) {
		name, sid, err := getCookies(r)
		if err != nil {
			err = cerrors.Wrap(err, 5, "Websocket handshake")
			errCh <- err
			return err
		}
		ok, down := clients.verifySid(name, sid)
		if down {
			err = errors.New("Game server down")
		} else if !ok {
			err = fmt.Errorf("Failed session id! Ip: %v", r.RemoteAddr)
			err = cerrors.Wrap(err, 6, "Websocket handshake")
			errCh <- err
		}
		return err
	}
	wsHandler := func(ws *websocket.Conn) {
		joinCh := make(chan *players.Player)
		name, sid, err := getCookies(ws.Request())
		ok, _, joined := clients.joinGameServer(name, sid, ws, errCh, joinCh)
		if ok {
			player := <-joinCh
			player.Start()
			err = ws.Close()
			clients.logOut(name)
			if err != nil {
				err = cerrors.Wrap(err, 8, "Player is finish closing websocket")
				errCh <- err
			}
		} else {
			if !joined {
				clients.logOut(name)
			}
			err = ws.Close()
			if err != nil {
				err = cerrors.Wrap(err, 9, "Player faild to join closing websocket")
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
			err = fmt.Errorf("Missing cookie! Ip: %v", r.RemoteAddr)
		}
	} else {
		err = fmt.Errorf("Missing cookie! Ip: %v", r.RemoteAddr)
	}
	return name, sid, err
}

//logInHandler the login page handler.
type logInHandler struct {
	clients   *Clients
	fileDown  string
	fileLogIn string
}

func serveFile(clients *Clients, file, fileDown string, w http.ResponseWriter, r *http.Request) {
	clients.mu.RLock()
	down := clients.gameServer == nil // not atomic
	clients.mu.RUnlock()
	if !down {
		http.ServeFile(w, r, file)
	} else {
		http.ServeFile(w, r, fileDown)
	}
}
func (l *logInHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveFile(l.clients, l.fileLogIn, l.fileDown, w, r)
}

//logInPostHandler the login post handler.
type logInPostHandler struct {
	clients   *Clients
	pages     *Pages
	errCh     chan<- error
	port      string
	fileDown  string
	fileLogIn string
}

func (g *logInPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	sid, err := g.clients.logIn(name, pw)
	if err != nil {
		_, ok := err.(*ErrDown)
		if ok {
			http.ServeFile(w, r, g.fileDown)
		} else {
			txt := fmt.Sprintf("Login failed! %v", err.Error())
			addToForm(txt, g.fileLogIn, g.pages, w)
			err = fmt.Errorf("Login failed! %v Ip: %v", err.Error(), r.RemoteAddr)
			g.errCh <- cerrors.Wrap(err, 10, "")
		}
	} else {
		setCookies(w, name, sid)
		http.Redirect(w, r, "/in/game", 303)
	}
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

//gameHandler the game handler this handler return our game page.
type gameHandler struct {
	clients  *Clients
	errCh    chan<- error
	port     string
	fileDown string
}

func (g *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, sid, err := getCookies(r)
	if err != nil {
		http.Redirect(w, r, "/", 303)
		g.errCh <- cerrors.Wrap(err, 11, "Serving game html file")
		return
	}
	ok, down := g.clients.verifySid(name, sid)
	if ok {
		http.ServeFile(w, r, "html/pages/game.html")
	} else if down {
		http.ServeFile(w, r, g.fileDown)
	} else {
		http.Redirect(w, r, "/", 303)
		err = fmt.Errorf("Failed session id! Ip: %v", r.RemoteAddr)
		g.errCh <- cerrors.Wrap(err, 12, "Serving game html file")
	}
}

//clientHandler The create new client page handler.
type clientHandler struct {
	clients          *Clients
	fileDown         string
	fileCreateClient string
}

func (c *clientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveFile(c.clients, c.fileCreateClient, c.fileDown, w, r)
}

//clientPostHandler the new client post handler.
type clientPostHandler struct {
	clients          *Clients
	pages            *Pages
	errCh            chan<- error
	port             string
	fileDown         string
	fileCreateClient string
}

func (handler *clientPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	sid, err := handler.clients.addNew(name, pw)
	if err != nil {
		switch err := err.(type) {
		case *ErrDown:
			http.ServeFile(w, r, handler.fileDown)
		case *ErrExist:
			addToForm(err.Error(), handler.fileCreateClient, handler.pages, w)
		case *ErrSize:
			w.WriteHeader(http.StatusBadRequest)
			errSize := fmt.Errorf("Data was submited with out our page validation! Ip: %v", r.RemoteAddr)
			handler.errCh <- cerrors.Wrap(errSize, 13, "Serve new client")
		default:
			addToForm("Unexpected error.", handler.fileCreateClient, handler.pages, w)
			handler.errCh <- cerrors.Wrap(err, 14, "Serve new client")
		}
	} else {
		setCookies(w, name, sid)
		http.Redirect(w, r, "/in/game", 303)
	}
}

//addToForm add a paragraph with a text message to a page.
func addToForm(txt string, fileName string, pages *Pages, w http.ResponseWriter) {
	body := pages.readPage(fileName)
	reader := bytes.NewReader(body)
	startNode, err := html.Parse(reader)
	if err != nil {
		panic(err.Error())
	}
	found := addPTextNode(startNode, createPTextNode(txt))
	if !found {
		panic("html file do not have a body")
	} else {
		err = html.Render(w, startNode)
		if err != nil {
			panic(err.Error())
		}
	}
}

// createPTextNode create a paragraph with a child text node.
func createPTextNode(txt string) (node *html.Node) {
	nodeTxt := new(html.Node)
	nodeTxt.Type = html.TextNode
	nodeTxt.Data = txt
	node = new(html.Node)
	node.Type = html.ElementNode
	node.Data = "p"
	node.DataAtom = atom.P
	node.AppendChild(nodeTxt)
	return node
}

// addPTextNode add a node as the last child.
func addPTextNode(node *html.Node, addNode *html.Node) (found bool) {
	if node.Type == html.ElementNode && node.DataAtom == atom.Body {
		found = true
		node.AppendChild(addNode)
	}
	if !found {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			found = addPTextNode(c, addNode)
		}
	}
	return found
}

//Add keep a live for 3 minute to a tcp handler.
//This is deafult but id do not know how good an idea this is.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return c, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
