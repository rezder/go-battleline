package http

import (
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/http/games"
	"github.com/rezder/go-error/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	//COST the password time cost, because of future improvement in hardware.
	pwCOST          = 5
	clientsFileNAME = "server/data/clients.gob"
)

var (
	//NAMESIZE the name character size limit.
	nameSIZE = [2]int{4, 20}
	//PWSIZE the password size limit.
	pwSIZE = [2]int{8, 20}
)

//Client the login object. Hold information of the user including
//loged-in information.
type Client struct {
	Name    string
	ID      int
	Pw      []byte
	Disable bool
	mu      *sync.Mutex
	//Filled when logIn
	sid     string
	sidTime time.Time
	//Filled when establish websocket. Just because they login does
	// not garantie they etstablish a web socket
	ws *websocket.Conn
}

//createClient creates a new client and log the client in.
func createClient(name string, id int, pw []byte) (client *Client) {
	client = new(Client)
	client.Name = name
	client.ID = id
	client.Pw = pw
	client.sid = sessionID()
	client.sidTime = time.Now()
	client.mu = new(sync.Mutex)
	return client
}

//Clients the clients list.
type Clients struct {
	mu         *sync.RWMutex
	List       map[string]*Client
	NextID     int
	gameServer *games.Server
}

func NewClients(games *games.Server) (clients *Clients) {
	clients = new(Clients)
	clients.gameServer = games
	clients.mu = new(sync.RWMutex)
	clients.List = make(map[string]*Client)
	clients.NextID = 1
	return clients
}

//loadClients loads client list from a file and adds
//the mutexs.
func loadClients(games *games.Server) (clients *Clients, err error) {
	file, err := os.Open(clientsFileNAME)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			clients = NewClients(games) //first start
		} else {
			err = errors.Wrapf(err, log.ErrNo(1)+"Open clients file %v", clientsFileNAME)
			return clients, err
		}
	} else {
		defer file.Close()
		decoder := gob.NewDecoder(file)
		lc := *NewClients(games)
		err = decoder.Decode(&lc)
		if err != nil {
			err = errors.Wrapf(err, "Decoding user file %v failed", clientsFileNAME)
			return clients, err
		}
		clients = &lc
		for _, client := range clients.List {
			client.mu = new(sync.Mutex)
		}
	}
	return clients, err
}

// save saves the client list to file.
// No lock is used.
func (clients *Clients) save() (err error) {
	file, err := os.Create(clientsFileNAME)
	if err != nil {
		err = errors.Wrap(err, log.ErrNo(2)+"Create clients file")
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(clients)
	return err
}

//logOut logout the client. Two locks are use the map and client.
func (clients *Clients) logOut(name string) {
	clients.mu.RLock()
	client := clients.List[name]
	clients.mu.RUnlock()
	client.mu.Lock()
	clearLogIn(client)
	client.mu.Unlock()
}

//clearLogIn clear the login information. No locks used.
func clearLogIn(client *Client) {
	client.sid = ""
	client.sidTime = *new(time.Time)
	client.ws = nil
}

//SetGameServer set the game server. The old game server is return,
//if set to nil all http server will return game server down.
//lock is used.
func (clients *Clients) SetGameServer(games *games.Server) (oldGames *games.Server) {
	oldGames = clients.gameServer
	clients.mu.Lock()
	clients.gameServer = games
	clients.mu.Unlock()
	return oldGames
}

//joinGameServer add a client to the game server.
//ok: True: if succes.
//down: True: if game server down.
//joined: True: if the client is already loged-in.
func (clients *Clients) joinGameServer(name string, sid string, ws *websocket.Conn,
	errCh chan<- error, joinCh chan<- *games.Player) (ok, down, joined bool) {
	clients.mu.RLock()
	if clients.gameServer != nil {
		client, found := clients.List[name]
		if found {
			client.mu.Lock()
			if client.sid != sid { //I do not think this is necessary because of the handshake
				client.mu.Unlock()
			} else {
				if client.ws == nil {
					client.ws = ws
					player := games.NewPlayer(client.ID, name, ws, errCh, joinCh)
					client.mu.Unlock()
					clients.gameServer.PlayersJoinCh() <- player
					ok = true
				} else {
					joined = true
					client.mu.Unlock()
				}
			}
		}
	} else {
		down = true
	}
	clients.mu.RUnlock()
	return ok, down, joined
}

// verifySid verify name and session id.
func (clients *Clients) verifySid(name, sid string) (ok, down bool) {
	ok = true
	clients.mu.RLock()
	down = clients.gameServer == nil
	client, found := clients.List[name]
	if found {
		client.mu.Lock()
		if !down {
			if client.sid != sid || client.ws != nil {
				ok = false
			}
		} else { //down
			ok = false
			clearLogIn(client) //Logout with out lock
		}
		client.mu.Unlock()
	}

	clients.mu.RUnlock()
	return ok, down
}

//logIn log-in a client.
func (clients *Clients) logIn(name string, pw string) (sid string, err error) {
	clients.mu.RLock()
	defer clients.mu.RUnlock()
	if clients.gameServer == nil {
		err = NewErrDown("Game server down")
		return sid, err
	}
	client, found := clients.List[name]
	if !found {
		err = errors.New("Name password combination do not exist")
		return sid, err
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.Disable {
		err = errors.New("Account disabled.")
		return sid, err
	}
	if client.sid != "" {
		if client.ws != nil {
			err = errors.New("Allready loged in.")
			return sid, err
		} else {
			if time.Now().Sub(client.sidTime) > time.Minute*3 {
				clearLogIn(client)
			} else {
				err = errors.New("Allready Loging in.")
				return sid, err
			}
		}
	}
	err = bcrypt.CompareHashAndPassword(client.Pw, []byte(pw))
	if err != nil {
		err = errors.New("Name password combination do not exist")
		return sid, err
	}
	client.sid = sessionID()
	client.sidTime = time.Now()
	sid = client.sid

	return sid, err
}

//disable disable/enable a client.
func (clients *Clients) updateDisable(name string, disable bool) {
	clients.mu.RLock() //TODO change to wr lock and save file.
	defer clients.mu.RUnlock()
	client, found := clients.List[name]
	if found {
		client.mu.Lock()
		defer client.mu.Unlock()
		if client.Disable != disable {
			client.Disable = disable
			if clients.gameServer != nil {
				clients.gameServer.PlayersDisableCh() <- &games.PlayersDisData{Disable: disable, PlayerID: client.ID}
			}
		}
	}

}

//addNew create and log-in a new client.
// Errors: ErrExist,ErrDown,ErrSize and bcrypt errors.
func (clients *Clients) addNew(name string, pwTxt string) (sid string, err error) {
	err = checkNamePwSize(name, pwTxt)
	if err != nil {
		return sid, err
	}
	clients.mu.Lock()
	if clients.gameServer != nil {
		client, found := clients.List[name]
		if !found {
			var pwh []byte
			pwh, err = bcrypt.GenerateFromPassword([]byte(pwTxt), pwCOST) //TODO we need salt some day, so player with same password does not have same hash.
			if err == nil {
				client = createClient(name, clients.NextID, pwh)
				clients.NextID = clients.NextID + 1
				clients.List[client.Name] = client
				sid = client.sid //TODO add save file
			}
		} else {
			err = NewErrExist("Name is used.")
		}
	} else {
		err = NewErrDown("Game server down")
	}
	clients.mu.Unlock()

	return sid, err
}

//create a random session id.
func sessionID() (txt string) {
	i := rand.Int()
	txt = strconv.Itoa(i)
	return txt
}

// checkNamePwSize check client name and password information for size when
//creating a new client.
func checkNamePwSize(name string, pw string) (err error) {
	nameSize := -1
	pwSize := -1

	if len(name) < nameSIZE[0] || len(name) > nameSIZE[1] {
		nameSize = len(name)
	}
	if len(pw) < pwSIZE[0] || len(pw) > pwSIZE[1] {
		pwSize = len(pw)
	}
	if nameSize != -1 || pwSize != -1 {
		err = NewErrSize(nameSize, pwSize)
	}

	return err
}

//ErrDown err when server is down.
type ErrDown struct {
	reason string
}

func NewErrDown(reason string) (e *ErrDown) {
	e = new(ErrDown)
	e.reason = reason
	return e
}

func (e *ErrDown) Error() string {
	return e.reason
}

//ErrSize err when password or name do meet size limits.
type ErrSize struct {
	name int
	pw   int
}

func NewErrSize(name int, pw int) (e *ErrSize) {
	e = new(ErrSize)
	e.name = name
	e.pw = pw
	return e
}
func (e *ErrSize) Error() string {
	var txt string
	switch {
	case e.name >= 0 && e.pw >= 0:
		txt = fmt.Sprintf("The lengh of Name: %v is illegal and the lenght of Password: %v is illegal.", e.name, e.pw)
	case e.name >= 0:
		txt = fmt.Sprintf("The lenght of Name: %v is illegal.", e.name)
	case e.pw >= 0:
		txt = fmt.Sprintf("The lenght of Password: %v is illegal", e.pw)
	}
	return txt
}

// ErrExist when user name and password combination do no match or user
//do not exist.
type ErrExist struct {
	txt string
}

func NewErrExist(txt string) (e *ErrExist) {
	e = new(ErrExist)
	e.txt = txt
	return e
}
func (e *ErrExist) Error() string {
	return e.txt
}
