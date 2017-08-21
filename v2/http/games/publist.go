package games

import (
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"strconv"
	"sync"
)

/*PubList is the structure that maintain the public data.
The list is synchronized with read write lock. Data comes from the two servers.
The data from the server should be a snapshot(copy) and the Read data is a pointer
to the current list and that list is never update, but replaced when updated.
This give all the players that read the list a reference to the same list (so no modify!!) until
the list is updated and any new reads can give differnt list.
*/
type PubList struct {
	lock    *sync.RWMutex
	games   map[int]*GameData
	players map[int]*PlayerData
	list    map[string]*PubData
}

//NewList create a list.
func NewList() (list *PubList) {
	list = new(PubList)
	list.lock = new(sync.RWMutex)
	list.games = make(map[int]*GameData)
	list.players = make(map[int]*PlayerData)
	list.list = make(map[string]*PubData)
	return list
}

//UpdateGames update the public data with game information.
//newgames is used directly so it must be a copy.
func (list *PubList) UpdateGames(newgames map[int]*GameData) {
	list.lock.Lock()
	list.games = newgames
	list.update()
	list.lock.Unlock()
}

//UpdatePlayers update the public data with player information.
//newplayers is used directly so it must be a copy.
func (list *PubList) UpdatePlayers(newplayers map[int]*PlayerData) {
	list.lock.Lock()
	list.players = newplayers
	list.update()
	list.lock.Unlock()
}

//Read get the current public list. The list is a multible used map.
//So no change.
func (list *PubList) Read() (publist map[string]*PubData) {
	list.lock.RLock()
	publist = list.list
	list.lock.RUnlock()
	return publist
}

//update updates the list with data from players and games.
func (list *PubList) update() {
	var data *PubData
	var isGameFound bool
	var gdata *GameData
	var isOppFound bool
	var opp *PlayerData
	publist := make(map[string]*PubData)
	for key, v := range list.players {
		data = new(PubData)
		data.ID = v.ID
		data.Name = v.Name
		data.InviteCh = v.InviteCh
		data.DoneComCh = v.DoneComCh
		data.MessageCh = v.MessageCh
		gdata, isGameFound = list.games[key]
		if isGameFound {
			opp, isOppFound = list.players[gdata.Opp] // opponent may have left
			if isOppFound {
				data.Opp = gdata.Opp
				data.OppName = opp.Name
				data.JoinWatchChCl = gdata.JoinWatchChCl
			}
		}
		publist[strconv.Itoa(key)] = data
	}
	list.list = publist
}

//PubData the public list data a combination of tabel and player information.
type PubData struct {
	ID            int
	Name          string
	InviteCh      chan<- *Invite  `json:"-"`
	DoneComCh     chan struct{}   `json:"-"` //Used by the player
	MessageCh     chan<- *MesData `json:"-"`
	Opp           int
	OppName       string
	JoinWatchChCl *JoinWatchChCl `json:"-"`
}

//GameData is the game information in the game map.
//Every game have two enteries one for every player.
type GameData struct {
	Opp           int
	JoinWatchChCl *JoinWatchChCl
}

//JoinWatchChData is the information send to a table to start
//or stop watching a game.
type JoinWatchChData struct {
	ID     int                    //This me
	SendCh chan<- *WatchingChData //Send here. Remember to close.
}

//JoinWatchChCl watch channel and its close channel.
type JoinWatchChCl struct {
	Channel chan *JoinWatchChData
	Close   chan struct{}
}

//NewJoinWatchChCl creates a new watch channel.
func NewJoinWatchChCl() (w *JoinWatchChCl) {
	w = new(JoinWatchChCl)
	w.Channel = make(chan *JoinWatchChData)
	w.Close = make(chan struct{})
	return w
}

//WatchingChData the data send to watchers.
type WatchingChData struct {
	ViewPos    *bg.ViewPos
	PlayingIDs [2]int
}

//PlayerData the public list player information.
type PlayerData struct {
	ID        int
	Name      string
	InviteCh  chan<- *Invite
	DoneComCh chan struct{}   //Used by all send to player
	MessageCh chan<- *MesData //never closed
	BootCh    chan struct{}   //For server to boot player
}

//Invite the invite data.
//The invite can be retracted before the receive channel have stoped listening.
//This make the standard select unreliable. Use a select with default to check if retracted and
//the receiver must count on receiving retracted responses.
type Invite struct {
	InvitorID   int
	InvitorName string
	ReceiverID  int
	IsRejected  bool                   //TODO MAYBE add reason
	ResponseCh  chan<- *InviteResponse `json:"-"` //Common for all invitaion
	RetractCh   chan struct{}          `json:"-"` //Per invite
	DoneComCh   chan struct{}          `json:"-"`
}

//MesData message data.
type MesData struct {
	Sender  int
	Name    string
	Message string
}

//InviteResponse the response to a invitation.
//GameCh is nil when decline.
type InviteResponse struct {
	Responder int
	Name      string
	PlayingCh chan<- *PlayingChData //nil when decline
}

//PlayingChData the information send from the table to the players.
//The players view of the game position and the move channel
// only set ones on the first send.
type PlayingChData struct {
	ViewPos          *bg.ViewPos
	PlayingIDs       [2]int
	FailedClaimedExs [9][]card.Move
	MoveCh           chan<- int `json:"-"`
}
