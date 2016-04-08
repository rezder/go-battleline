package players

import (
	"golang.org/x/net/websocket"
	pub "rezder.com/game/card/battleline/server/publist"
)

func Start(join chan *Player, pubList *pub.List, finished chan struct{}) {
	leave := make(chan int)
	list := make(map[int]*pub.PlayerData)
	done := false
Loop:
	for {
		select {
		case id := <-leave:
			delete(list, id)
			if done && len(list) == 0 {
				break Loop
			}
			publish(list, pubList)
		case p, open := <-join:
			if open {
				invite := make(chan *pub.Invite)
				p.joinServer(invite, leave, pubList)
				list[p.id] = &pub.PlayerData{p.id, p.name, invite, p.doneInvite}
				publish(list, pubList)
			} else {
				if len(list) > 0 {
					for _, player := range list {
						close(player.DoneInvite)
						close(player.Invite)
					}
				} else {
					break Loop
				}
			}
		}
	}
	close(finished)
}
func publish(list map[int]*pub.PlayerData, pubList *pub.List) {
	copy := make(map[int]*pub.PlayerData)
	for key, v := range list {
		copy[key] = v
	}
	go pubList.UpdatePlayers(copy)
}

type Player struct {
	id         int
	name       string
	leave      chan<- int
	pubList    *pub.List
	invite     <-chan *pub.Invite
	doneInvite chan struct{}
	ws         *websocket.Conn
	err        chan<- error
	closed     bool
}

func NewPlayer(id int, name string, ws *websocket.Conn, errChan chan<- error) (p *Player) {
	p = new(Player)
	p.id = id
	p.name = name
	p.ws = ws
	p.doneInvite = make(chan struct{})
	p.err = errChan
	return p
}
func (p *Player) joinServer(invite <-chan *pub.Invite, leave chan<- int, pubList *pub.List) {
	p.leave = leave
	p.pubList = pubList
	p.invite = invite
}
func (player *Player) Start() {
	defer player.closeConn()
	sendChan := make(chan interface{}, 10)
	brokeConn := make(chan struct{})
	slowConn := make(chan struct{})
	go netSend(player.ws, sendChan, player.err, brokeConn, slowConn)

	c := make(chan *Action, 1)
	go netReceive(c, player.err, brokeConn)
	var actChan <-chan *Action
	actChan = c

	inviteResponse := make(chan *pub.InviteResponse)

	sendInvites := make(map[int]*pub.Invite)
	receivedInvites := make(map[int]*pub.Invite)

	startGame := make(chan chan<- [2]int) //if already sat game is finished and gameChan should be sat to nil
	var gameChan chan<- [2]int

	startWatchGameChan := make(chan *startWatchGameData)
	var watchGames map[int]<-chan *pub.WatchChan
Loop:
	for {
		select {
		case <-brokeConn:
		case act := <-actChan:
		case gc := <-startGame:
		case wgd := <-startWatchGameChan:
		case invite, open := <-player.invite:
		case response := <-inviteResponse:
		}
	}
}

func netSend(ws *websocket.Conn, toSend <-chan interface{}, err chan<- error, brokenConn chan struct{}, slow chan struct{}) {

}
func netReceive(accChan chan<- *Action, errChan chan<- error, brokenConn chan struct{}) {

}

type Action struct {
}

type startWatchGameData struct {
	channel <-chan *pub.WatchChan
	player  int
}

func (player *Player) closeConn() {
	if !player.closed {
		err := player.ws.Close()
		if err != nil {
			player.err <- err
		}
	}
}
