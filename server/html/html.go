// The http server
package html

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"rezder.com/game/card/battleline/server/players"
	"time"
)

// Start the server.
func Start(errCh chan<- error, netListener *net.TCPListener, clients *Clients,
	finCh chan struct{}) {
	pages := NewPages("html/pages")
	pages.load()
	http.Handle("/", &logInHandler{clients, pages})
	http.Handle("/client", &clientHandler{clients, pages})
	http.Handle("/in/game", &gameHandler{clients, pages, errCh})
	http.Handle("/form/login", &logInPostHandler{clients, pages, errCh})
	http.Handle("/form/client", &clientPostHandler{clients, pages, errCh})
	http.Handle("/in/gamews", *createWsHandler(clients, errCh))
	http.Handle("/static/", http.FileServer(http.Dir("./html")))

	server := &http.Server{Addr: "game.rezder.com:8181"} //address is not used
	err := server.Serve(tcpKeepAliveListener{netListener})
	errCh <- err
	close(finCh)
}

//createWsHandler create the websocket handler.
func createWsHandler(clients *Clients, errCh chan<- error) (server *websocket.Server) {
	wsHandshake := func(ws *websocket.Config, r *http.Request) (err error) {
		name, sid, err := getSidCookies(r)
		if err == nil {
			ok, down := clients.verifySid(name, sid)
			if down {
				err = errors.New("Game server down")
			} else if !ok {
				err = errors.New(fmt.Sprintf("Failed session id! Ip: %v", r.RemoteAddr))
				errCh <- err
			}
		} else {
			errCh <- err
		}
		return err
	}
	wsHandler := func(ws *websocket.Conn) {
		joinCh := make(chan *players.Player)
		name, sid, err := getSidCookies(ws.Request())
		ok, _, joined := clients.joinGameServer(name, sid, ws, errCh, joinCh)
		if ok {
			player := <-joinCh
			player.Start()
			err = ws.Close()
			clients.logOut(name)
		} else {
			if !joined {
				clients.logOut(name)
			}
			err = ws.Close()
		}
		if err != nil {
			errCh <- err
		}
	}
	server = &websocket.Server{Handler: wsHandler, Handshake: wsHandshake}
	return server
}

//getSidCookies extract the name and session cookies.
func getSidCookies(r *http.Request) (name string, sid string, err error) {
	nameC, err := r.Cookie("name")
	if err == nil {
		name = nameC.Value
		sidC, err := r.Cookie("sid")
		if err == nil {
			sid = sidC.Value
		} else {
			err = errors.New(fmt.Sprintf("Missing cookie! Ip: %v", r.RemoteAddr))
		}
	} else {
		err = errors.New(fmt.Sprintf("Missing cookie! Ip: %v", r.RemoteAddr))
	}
	return name, sid, err
}

//logInHandler the login page handler.
type logInHandler struct {
	clients *Clients
	pages   *Pages
}

func (l *logInHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.clients.RLock()
	down := l.clients.Games == nil // not atomic
	l.clients.RUnlock()
	if !down {
		w.Write(l.pages.readPage("login.html"))
	} else {
		w.Write(l.pages.readPage("down.html"))
	}
}

//logInPostHandler the login post handler.
type logInPostHandler struct {
	clients *Clients
	pages   *Pages
	errCh   chan<- error
}

func (g *logInPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	sid, err := g.clients.logIn(name, pw)
	if err != nil {
		_, ok := err.(*ErrDown)
		if ok {
			w.Write(g.pages.readPage("down.html"))
		} else {
			txt := fmt.Sprintf("Login failed! %v", err.Error())
			addToForm(txt, "login.html", g.pages, w)
			g.errCh <- errors.New(fmt.Sprintf("Login failed! %v Ip: %v", err.Error(), r.RemoteAddr))
		}
	} else {
		setSidCookies(w, name, sid)
		http.Redirect(w, r, "http://game.rezder.com:8181/in/game", 303)
	}
}

//setSidCookies set the name and session id cookies.
func setSidCookies(w http.ResponseWriter, name string, sid string) {
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
	clients *Clients
	pages   *Pages
	errCh   chan<- error
}

func (g *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, sid, err := getSidCookies(r)
	if err == nil {
		ok, down := g.clients.verifySid(name, sid)
		if ok {
			w.Write(g.pages.readPage("game.html"))
		} else if down {
			w.Write(g.pages.readPage("down.html"))
		} else {
			w.WriteHeader(http.StatusBadRequest) // TODO this status did not give the expected result
			g.errCh <- errors.New(fmt.Sprintf("Failed session id! Ip: %v", r.RemoteAddr))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		g.errCh <- err
	}
}

//clientHandler The create new client page handler.
type clientHandler struct {
	clients *Clients
	pages   *Pages
}

func (c *clientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.clients.RLock()
	down := c.clients.Games == nil // not atomic
	c.clients.RUnlock()
	if !down {
		w.Write(c.pages.readPage("client.html"))
	} else {
		w.Write(c.pages.readPage("down.html"))
	}
}

//clientPostHandler the new client post handler.
type clientPostHandler struct {
	clients *Clients
	pages   *Pages
	errCh   chan<- error
}

func (handler *clientPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("txtUserName")
	pw := r.FormValue("pwdPassword")
	sid, err := handler.clients.addNew(name, pw)
	if err != nil {
		switch err := err.(type) {
		case *ErrDown:
			w.Write(handler.pages.readPage("down.html"))
		case *ErrExist:
			addToForm(err.Error(), "client.html", handler.pages, w)
		case *ErrSize:
			w.WriteHeader(http.StatusBadRequest)
			handler.errCh <- errors.New(fmt.Sprintf("Data was submited with out our page validation! Ip: %v", r.RemoteAddr))
		default:
			addToForm("Unexpected error.", "client.html", handler.pages, w)
			handler.errCh <- err
		}
	} else {
		setSidCookies(w, name, sid)
		http.Redirect(w, r, "http://game.rezder.com:8181/in/game", 303)
	}
}

//addToForm add a paragraph with a text message to a page.
func addToForm(txt string, fileName string, pages *Pages, w http.ResponseWriter) {
	body := pages.readPage(fileName)
	reader := bytes.NewReader(body)
	startNode, err := html.Parse(reader)
	if err != nil {
		panic(err.Error())
	} else {
		found := addPTextNode(startNode, createPTextNode(txt))
		if !found {
			panic("html file do not have a body")
		} else {
			html.Render(w, startNode)
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
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
