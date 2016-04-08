/*publist is used to keep public list of player information.
The list is synchronized with read write lock. Data comes from the two servers.
The data from the server should be a snapshot(copy) and the Read data is a pointer
to the current list and that list is never update, but replaced when updated.
This give all the players that read the list a reference to the same list (so no modify!!) until
the list is updated and any new reads will give differnt list.
*/
package publist

import (
	bat "rezder.com/game/card/battleline"
	"strconv"
	"sync"
)

type List struct {
	lock    *sync.RWMutex
	games   map[int]*GameData
	players map[int]*PlayerData
	list    map[string]*Data
}

func New() (list *List) {
	list = new(List)
	list.lock = new(sync.RWMutex)
	list.games = make(map[int]*GameData)
	list.players = make(map[int]*PlayerData)
	list.list = make(map[string]*Data)
	return list
}

func (list *List) UpdateGames(newgames map[int]*GameData) {
	list.lock.Lock()
	list.games = newgames
	list.update()
	list.lock.Unlock()
}
func (list *List) UpdatePlayers(newplayers map[int]*PlayerData) {
	list.lock.Lock()
	list.players = newplayers
	list.update()
	list.lock.Unlock()
}

func (list *List) Read() (publist map[string]*Data) {
	list.lock.RLock()
	publist = list.list
	list.lock.RUnlock()
	return publist
}

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
		data.DoneInvite = v.DoneInvite
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

type GameData struct {
	Opp   int
	Watch *WatchChan
}
type WatchData struct {
	Id   int              //This me
	Send chan<- MoveBench //send here. Remember to close.
}
type WatchChan struct {
	Channel chan *WatchData
	Close   chan struct{}
}

func NewWatchChan() (w *WatchChan) {
	w = new(WatchChan)
	w.Channel = make(chan *WatchData)
	w.Close = make(chan struct{})
	return w
}

type MoveBench struct {
	Mover     int
	Move      bat.Move
	NextMover int
}

type PlayerData struct {
	Id         int
	Name       string
	Invite     chan<- *Invite //Closed by the server
	DoneInvite chan struct{}  //Used by the server
}
type Invite struct {
	Inviter  int
	Response chan *InviteResponse //Common for all invitaion
	Retract  chan struct{}        //Per invite
}
type InviteResponse struct {
	Responder int
	GameChan  chan *MoveView //nil when decline
}
type MoveView struct {
	Mover bool
	Move  bat.Move
	Card  int //This is the return card from deck moves, zero when not used.
	*Turn
	MoveChan chan<- [2]int `json:"-"`
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

type Data struct {
	Id         int
	Name       string
	Invite     chan<- *Invite
	DoneInvite chan struct{} //Used by the player
	Opp        int           // maybe this is not need
	OppName    string
	Watch      *WatchChan
}
