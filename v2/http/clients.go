package http

import (
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/http/games"
	"github.com/rezder/go-battleline/v2/http/login"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	//pwCOST the password time cost, because of future improvement in hardware.
	pwCOST = 5
)

var (
	//NAMESIZE the name character size limit.
	nameSIZE = [2]int{4, 20}
	//PWSIZE the password size limit.
	pwSIZE        = 8
	dbClientsFILE = "server/data/clients.db" // can be change for test
)

//Client the login object. Hold information of the user including
//loged-in information.
type Client struct {
	Name      string
	ID        int
	Pw        []byte
	IsDisable bool
	//Filled when logIn
	sid     string
	sidTime time.Time
	//Filled when establish websocket. Just because they login does
	// not garantie they etstablish a web socket
	ws *websocket.Conn
	//Filled when asked to logout
	isBootOut bool
}

//Copy copies a client.
func (c *Client) Copy() *Client {
	cv := *c
	if c.Pw != nil {
		pwh := make([]byte, len(c.Pw))
		copy(pwh, c.Pw)
		cv.Pw = pwh
	}
	return &cv
}

//IsEqual tests for equal.
func (c *Client) IsEqual(o *Client) (isEqual bool) {
	if c == o {
		isEqual = true
	} else {
		if c.Name == o.Name &&
			c.ID == o.ID &&
			c.IsDisable == o.IsDisable &&
			c.sid == o.sid &&
			c.sidTime == o.sidTime &&
			c.ws == o.ws &&
			c.isBootOut == o.isBootOut {
			if len(c.Pw) == len(o.Pw) {
				isEqual = true
				for i, b := range c.Pw {
					if b != o.Pw[i] {
						isEqual = false
						break
					}
				}
			}
		}
	}

	return isEqual
}

func NewClient(name string, pwTxt string) (client *Client, err error) {
	pwh, err := bcrypt.GenerateFromPassword([]byte(pwTxt), pwCOST) //TODO we need salt some day, so player with same password does not have same hash.
	if err != nil {
		return client, err
	}
	client = new(Client)
	client.Name = name
	client.Pw = pwh
	return client, err
}

//Clients the clients list.
type Clients struct {
	mu         *sync.RWMutex
	logIns     map[string]*Client
	cdb        *CDb
	gameServer *games.Server
}

//NewClients creates new clients.
func NewClients(games *games.Server) (clients *Clients, err error) {
	clients = new(Clients)
	clients.gameServer = games
	clients.mu = new(sync.RWMutex)
	clients.logIns = make(map[string]*Client)
	db, err := NewCdb(dbClientsFILE)
	if err != nil {
		_ = clients.gameServer.Cancel()
		err = errors.Wrapf(err, "Init data base file %v failed", dbClientsFILE)
		return clients, err
	}
	clients.cdb = db
	return clients, err
}

//CancelGameServer cancels the game server.
func (clients *Clients) CancelGameServer() (err error) {
	if clients.gameServer != nil {
		err = clients.gameServer.Cancel()
	}
	return err
}

//Close closes clients database.
func (clients *Clients) Close() (err error) {
	err = clients.cdb.Close()
	return err
}

//LogOut logs out the client.
func (clients *Clients) LogOut(name string) {
	clients.mu.Lock()
	defer clients.mu.Unlock()
	client, isFound := clients.logIns[name]
	if isFound {
		if client.isBootOut {
			if clients.gameServer != nil {
				clients.gameServer.BootPlayerStop(client.ID)
			}
		}
		delete(clients.logIns, name)
	}
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

//JoinGameServer add a client to the game server.
//ok: True: if request succeded the player is returned on the joined channel
// when ready.
//isJoined: True: if the client is already loged-in.
func (clients *Clients) JoinGameServer(name string, sid string, ws *websocket.Conn,
	errCh chan<- error, joinedCh chan<- *games.Player) (ok, isJoined bool) {
	clients.mu.Lock()
	defer clients.mu.Unlock()
	if clients.gameServer != nil {
		client, found := clients.logIns[name]
		if found {
			if client.sid == sid { //I do not think this is necessary because of the handshake
				if client.ws == nil {
					client.ws = ws
					clients.gameServer.JoinClient(client.ID, client.Name, client.ws, errCh, joinedCh)
					ok = true
				} else {
					isJoined = true
				}
			}
		}
	}
	return ok, isJoined
}

// VerifySid verify name and session id,
//before ws is sat.
func (clients *Clients) VerifySid(name, sid string) (ok, isDown bool) {
	clients.mu.RLock()
	defer clients.mu.RUnlock()
	isDown = clients.gameServer == nil
	client, found := clients.logIns[name]
	if found {
		if client.sid == sid && client.ws == nil {
			ok = true
		}
	}
	return ok, isDown
}

//LogIn log-in a client.
func (clients *Clients) LogIn(name string, pw string) (status login.Status, sid string, err error) {
	client, isFound, err := clients.cdb.GetName(name)
	if err != nil {
		err = errors.Wrapf(err, "Failed loading client %v from database", name)
		return status, sid, err
	}
	if !isFound {
		status = login.StatusAll.InValid
	} else {
		cerr := bcrypt.CompareHashAndPassword(client.Pw, []byte(pw))
		if cerr == bcrypt.ErrMismatchedHashAndPassword {
			status = login.StatusAll.InValid
		} else if cerr != nil {
			return status, sid, cerr
		} else {
			clients.mu.Lock()
			defer clients.mu.Unlock()
			if clients.gameServer == nil {
				status = login.StatusAll.Down
			} else {
				client, isFound, err = clients.cdb.GetID(client.ID)
				if err != nil {
					err = errors.Wrapf(err, "Failed loading client name,id: %v,%v from database", name, client.ID)
					return status, sid, err
				}
				if !isFound {
					err = errors.Errorf("Failed to load just verified client name,id %v,%v from data base, this should never happen", name, client.ID)
					return status, sid, err
				}
				inClient, isIn := clients.logIns[name]
				if isIn && inClient.ws != nil || isIn && time.Since(inClient.sidTime) < time.Minute*3 {
					status = login.StatusAll.Exist
				} else {
					if isIn {
						delete(clients.logIns, name)
					}
					if client.IsDisable {
						status = login.StatusAll.Disabled
					} else {
						client.sid = sessionID()
						client.sidTime = time.Now()
						sid = client.sid
						clients.logIns[name] = client
						status = login.StatusAll.Ok
					}
				}
			}
		}
	}
	return status, sid, err
}

//UpdateDisable disable/enable a client.
func (clients *Clients) UpdateDisable(id int, isDisable bool) (err error) {
	name, isUpd, err := clients.cdb.UpdDisable(id, isDisable)
	if err != nil {
		return err
	}
	if isUpd {
		clients.bootPlayer(name)
	}
	return err
}
func (clients *Clients) bootPlayer(name string) {
	client, found := clients.logIns[name]
	if found {
		client.isBootOut = true
		if clients.gameServer != nil {
			clients.gameServer.BootPlayer(client.ID)
		}
	}
}

//IsGameServerDown checks if the game server is down.
func (clients *Clients) IsGameServerDown() bool {
	return clients.gameServer == nil
}

//AddNew create and log-in a new client.
func (clients *Clients) AddNew(name string, pwTxt string) (status login.Status, sid string, err error) {
	if checkNamePwSize(name, pwTxt) {
		client, err := NewClient(name, pwTxt)
		if err != nil {
			return status, sid, err
		}
		var isUpd bool
		client, isUpd, err = clients.cdb.UpdInsert(client)
		if err != nil {
			return status, sid, err
		}
		if isUpd {
			clients.mu.Lock()
			defer clients.mu.Unlock()
			if clients.gameServer == nil {
				status = login.StatusAll.Down
			} else {
				var isFound bool
				client, isFound, err = clients.cdb.GetID(client.ID)
				if err != nil {
					err = errors.Wrapf(err, "Failed loading client name,id: %v,%v from database", name, client.ID)
					return status, sid, err
				}
				if !isFound {
					err = errors.Errorf("Failed to load just verified client name,id %v,%v from data base, this should never happen", name, client.ID)
					return status, sid, err
				}
				client.sid = sessionID()
				client.sidTime = time.Now()
				sid = client.sid
				clients.logIns[name] = client
				status = login.StatusAll.Ok
			}
		} else {
			status = login.StatusAll.Exist
		}
	} else {
		status = login.StatusAll.InValid
	}
	return status, sid, err
}

//create a random session id.
func sessionID() (txt string) {
	i := rand.Int()
	txt = strconv.Itoa(i)
	return txt
}

// checkNamePwSize check client name and password information for size when
//creating a new client.
func checkNamePwSize(name string, pw string) (ok bool) {
	if len(name) >= nameSIZE[0] &&
		len(name) <= nameSIZE[1] &&
		len(pw) >= pwSIZE {
		ok = true
	}
	return ok
}
