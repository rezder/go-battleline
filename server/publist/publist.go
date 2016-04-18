/*publist is used to keep public list of player information.
The list is synchronized with read write lock. Data comes from the two servers.
The data from the server should be a snapshot(copy) and the Read data is a pointer
to the current list and that list is never update, but replaced when updated.
This give all the players that read the list a reference to the same list (so no modify!!) until
the list is updated and any new reads can give differnt list.
*/
package publist

import (
	bat "rezder.com/game/card/battleline"
	"strconv"
	"sync"
)

//List is the structure that maintain the public data.
type List struct {
	lock    *sync.RWMutex
	games   map[int]*GameData
	players map[int]*PlayerData
	list    map[string]*Data
}

//New create a list.
func New() (list *List) {
	list = new(List)
	list.lock = new(sync.RWMutex)
	list.games = make(map[int]*GameData)
	list.players = make(map[int]*PlayerData)
	list.list = make(map[string]*Data)
	return list
}

//UpdateGames update the public data with game information.
//newgames is used directly so it must be a copy.
func (list *List) UpdateGames(newgames map[int]*GameData) {
	list.lock.Lock()
	list.games = newgames
	list.update()
	list.lock.Unlock()
}

//UpdatePlayers update the public data with player information.
//newplayers is used directly so it must be a copy.
func (list *List) UpdatePlayers(newplayers map[int]*PlayerData) {
	list.lock.Lock()
	list.players = newplayers
	list.update()
	list.lock.Unlock()
}

//Read get the current public list. The list is a multible used map.
//So no change.
func (list *List) Read() (publist map[string]*Data) {
	list.lock.RLock()
	publist = list.list
	list.lock.RUnlock()
	return publist
}

//update updates the list with data from players and games.
func (list *List) update() {
	var data *Data
	var found bool
	var gdata *GameData
	publist := make(map[string]*Data)
	for key, v := range list.players {
		data = new(Data)
		data.Id = v.Id
		data.Name = v.Name
		data.Invite = v.Invite
		data.DoneCom = v.DoneCom
		data.Message = v.Message
		gdata, found = list.games[key]
		if found {
			data.Opp = gdata.Opp
			data.OppName = list.players[data.Opp].Name
			data.Watch = gdata.Watch
		}
		publist[strconv.Itoa(key)] = data
	}
	list.list = publist
}

//GameData is the game information in the game map.
//Every game have two enteries one for every player.
type GameData struct {
	Opp   int
	Watch *WatchChCl
}

//WatchData is the information send to a table to start watching a game.
type WatchData struct {
	Id   int               //This me
	Send chan<- *MoveBench //Send here. Remember to close.
}

//The watch channel and its close channel.
type WatchChCl struct {
	Channel chan *WatchData
	Close   chan struct{}
}

func NewWatchChCl() (w *WatchChCl) {
	w = new(WatchChCl)
	w.Channel = make(chan *WatchData)
	w.Close = make(chan struct{})
	return w
}

//MoveBench the data send to watchers.
type MoveBench struct {
	Mover     int
	Move      bat.Move
	NextMover int
}

//PlayerData the data send between logon server and the player server.
type PlayerData struct {
	Id      int
	Name    string
	Invite  chan<- *Invite
	DoneCom chan struct{}   //Used by all send to player
	Message chan<- *MesData //never closed
	BootCh  chan struct{}   //For server to boot player
}

//Invite the invite data.
//The invite can be retracted before the receive channel have stoped listening.
//This make the standard select unreliable. Use a select with default to check if retracted and
//the receiver must count on receiving retracted responses.
type Invite struct {
	Inviter   int
	Name      string
	Response  chan<- *InviteResponse `json:"-"` //Common for all invitaion
	Retract   chan struct{}          `json:"-"` //Per invite
	DoneComCh chan struct{}          `json:"-"`
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
	GameCh    chan<- *MoveView //nil when decline
}

//MoveView the information send from the table to the players.
//It should contain all information to animate a player move.
//Its is non symetrics, each player get his own.
type MoveView struct {
	Mover bool
	Move  bat.Move
	Card  int //This is the return card from deck moves, zero when not used.
	*Turn
	MoveCh chan<- [2]int `json:"-"`
}

// Turn a player turn.
type Turn struct {
	MyTurn    bool
	State     int
	Moves     []bat.Move
	MovesHand map[string][]bat.Move
	MovesPass bool
}

func NewTurn(turn *bat.Turn, playerix int) (disp *Turn) {
	disp = new(Turn)
	disp.MyTurn = turn.Player == playerix
	disp.State = turn.State
	if turn.Moves != nil {
		disp.Moves = make([]bat.Move, len(turn.Moves))
		copy(disp.Moves, turn.Moves)
	}
	if turn.MovesHand != nil {
		disp.MovesHand = make(map[string][]bat.Move)
		for k, v := range turn.MovesHand {
			m := make([]bat.Move, len(v))
			copy(m, v)
			disp.MovesHand[strconv.Itoa(k)] = m
		}
	}

	disp.MovesPass = turn.MovePass
	return disp
}

func (t *Turn) Equal(other *Turn) (equal bool) {
	if other == nil && t == nil {
		equal = true
	} else if other != nil && t != nil {
		if t == other {
			equal = true
		} else if t.MyTurn == other.MyTurn && t.State == other.State && t.MovesPass == other.MovesPass {
			mequal := false
			if len(other.Moves) == 0 && len(t.Moves) == 0 {
				mequal = true
			} else if len(other.Moves) == len(t.Moves) {
				mequal = true
				for i, v := range other.Moves {
					if !v.MoveEqual(t.Moves[i]) {
						mequal = false
						break
					}
				}
			}
			if mequal {
				mhequal := false
				if len(other.MovesHand) == 0 && len(t.MovesHand) == 0 {
					mhequal = true
				} else if len(other.MovesHand) == len(t.MovesHand) {
					mhequal = true
				Card:
					for cardix, moves := range other.MovesHand {
						turnMoves, found := t.MovesHand[cardix]
						if found && len(moves) == len(turnMoves) {
							for i, v := range moves {
								if !v.MoveEqual(turnMoves[i]) {
									mhequal = false
									break Card
								}
							}
						} else {
							mhequal = false
							break
						}
					}
				}
				if mhequal {
					equal = true
				}
			}
		}
	}
	return equal
}

//Data the public list data a combination of tabel and player information.
type Data struct {
	Id      int
	Name    string
	Invite  chan<- *Invite
	DoneCom chan struct{} //Used by the player
	Message chan<- *MesData
	Opp     int // maybe this is not need
	OppName string
	Watch   *WatchChCl
}
