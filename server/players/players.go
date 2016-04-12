package players

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	pub "rezder.com/game/card/battleline/server/publist"
	"rezder.com/game/card/battleline/server/tables"
	"strconv"
)

const (
	ACT_MESS       = 1
	ACT_INVITE     = 2
	ACT_INVACCEPT  = 3
	ACT_INVDECLINE = 4
	ACT_MOVE       = 5
	ACT_QUIT       = 6
	ACT_WATCH      = 7
	ACT_WATCHSTOP  = 8

	WRTBUFF_SIZE = 10
	WRTBUFF_LIM  = 8
)

func Start(join chan<- *Player, joined <-chan *Player, pubList *pub.List, startGameCh *tables.StartGameChan, finished chan struct{}) {
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
				mess := make(chan *pub.MesData)
				p.joinServer(invite, mess, leave, pubList, startGameCh)
				joined <- p
				list[p.id] = &pub.PlayerData{p.id, p.name, invite, p.doneCom, mess, p.bootCh}
				publish(list, pubList)
			} else {
				if len(list) > 0 {
					for _, player := range list {
						close(player.BootCh)
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
	id        int
	name      string
	tableStCh *tables.StartGameChan
	leave     chan<- int
	pubList   *pub.List
	invite    <-chan *pub.Invite
	doneCom   chan struct{}
	mess      <-chan *pub.MesData
	ws        *websocket.Conn
	err       chan<- error
	bootCh    chan struct{} //server boot channel. used to kick player out.
}

func NewPlayer(id int, name string, ws *websocket.Conn, errChan chan<- error) (p *Player) {
	p = new(Player)
	p.id = id
	p.name = name
	p.ws = ws
	p.doneCom = make(chan struct{})
	p.err = errChan
	p.bootCh = make(chan struct{})
	return p
}
func (p *Player) joinServer(invite <-chan *pub.Invite, mess <-chan *pub.MesData,
	leave chan<- int, pubList *pub.List, startGameCh *tables.StartGameChan) {
	p.leave = leave
	p.pubList = pubList
	p.invite = invite
	p.mess = mess
	p.tableStCh = startGameCh
}

func (player *Player) Start() {
	sendCh := make(chan interface{}, WRTBUFF_SIZE)
	wrtBrookCh := make(chan struct{})
	readList := player.pubList.Read()
	go netWrite(player.ws, sendCh, player.err, wrtBrookCh)

	c := make(chan *Action, 1)
	go netRead(player.ws, c, player.doneCom, player.err)
	var actChan <-chan *Action
	actChan = c

	inviteResponse := make(chan *pub.InviteResponse)

	sendInvites := make(map[int]*pub.Invite)
	receivedInvites := make(map[int]*pub.Invite)

	gameReceiveCh := make(chan *pub.MoveView) //if already sat game is finished and
	//gameChan should be sat to nil
	var gameRespCh chan<- [2]int
	var gameMove *pub.MoveView

	watchGameCh := make(chan *startWatchGameData)
	var watchGames map[int]*pub.WatchChan

Loop:
	for {
		select {
		case <-player.bootCh:
			closeDown(player.doneCom, gameRespCh, watchGames, player.id, receivedInvites, sendInvites, sendCh)
			break Loop
		case <-wrtBrookCh:
			closeDown(player.doneCom, gameRespCh, watchGames, player.id, receivedInvites, sendInvites, sendCh)
			break Loop
		case act, open := <-actChan:
			if open {
				upd := false
				switch act.ActType { //TODO Add open. When closed client disabled. Request disable com if not already dissable if both disable finsih Check stop games
				//stop Com and stop client / stop game, stop watches, clear invites/ finsihCheck
				case ACT_MESS:
					upd = actMessage(act, readList, sendCh, player.id, player.name)
				case ACT_INVITE:
					upd = actSendInvite(sendInvites, inviteResponse, player.doneCom,
						act, readList, sendCh, player.id, player.name)
				case ACT_INVACCEPT:
					actAccInvite(receivedInvites, act, sendCh, gameReceiveCh, player.doneCom, player.id)
				case ACT_INVDECLINE:
					actDeclineInvite(receivedInvites, act, player.id)
				case ACT_MOVE:
					actMove(act, gameRespCh, gameMove, sendCh)
				case ACT_QUIT:
					if gameRespCh != nil {
						close(gameRespCh)
					} else {
						sendSysMess(sendCh, "Quitting game with no game active")
					}
				case ACT_WATCH:
					upd = actWatch(watchGames, act, watchGameCh, player.doneCom, readList,
						sendCh, player.id)
				case ACT_WATCHSTOP:
					actWatchStop(watchGames, act, sendCh, player.id)
				default:
					sendSysMess(sendCh, "No existen action")
				}
				if upd {
					sendCh <- player.pubList.Read()
				}
			} else {
				closeDown(player.doneCom, gameRespCh, watchGames, player.id, receivedInvites, sendInvites, sendCh)
				break Loop
			}

		case move := <-gameReceiveCh: //TODO sep chan for init
			if move == nil {
				gameRespCh = nil
				gameMove = nil
			} else {
				if gameRespCh == nil { //Init move
					gameRespCh = move.MoveChan
					gameMove = move
					clearInvites(receivedInvites, sendInvites, player.id)
					sendCh <- move
					readList = player.pubList.Read()
					sendCh <- readList
				} else {
					gameRespCh = move.MoveChan
					gameMove = move
					sendCh <- move
				}
			}
		case wgd := <-watchGameCh:
			if wgd.channel == nil { //sep chan for done
				delete(watchGames, wgd.player)
			} else {
				watchGames[wgd.player] = wgd.channel
				sendCh <- wgd.initMove
			}
		case invite := <-player.invite:

			readList = playerExist(player.pubList, readList, invite.Inviter, sendCh)
			_, found := readList[strconv.Itoa(invite.Inviter)]
			if found {
				receivedInvites[invite.Inviter] = invite
				sendCh <- invite
			}

		case response := <-inviteResponse:
			invite, found := sendInvites[response.Responder]
			if found {
				delete(sendInvites, response.Responder)
				if response.GameChan == nil {
					mess := new(pub.MesData)
					mess.Sender = response.Responder
					mess.Name = response.Name
					mess.Message = "Sorry I must decline your offer to play a game."
					sendCh <- mess
				} else {
					moveRecCh := make(chan *pub.MoveView, 1)
					tableData := new(tables.StartGameData)
					tableData.PlayerIds = [2]int{player.id, response.Responder}
					tableData.PlayerChans = [2]chan<- *pub.MoveView{moveRecCh, response.GameChan}
					select {
					case player.tableStCh.Channel <- tableData:
						go gameListen(gameReceiveCh, player.doneCom, moveRecCh, sendCh, invite.Name)
					case <-player.tableStCh.Close:
						txt := fmt.Sprintf("Failed to start game with %v as server no longer accept games",
							response.Name)
						sendSysMess(sendCh, txt)
					}
				}
			} else {
				if response.GameChan != nil {
					close(response.GameChan)
				}
			}
		case message := <-player.mess:
			readList = playerExist(player.pubList, readList, message.Sender, sendCh)
			_, found := readList[strconv.Itoa(message.Sender)]
			if found {
				sendCh <- message
			}
		}
	}
	player.leave <- player.id
}
func closeDown(doneCom chan struct{}, gameRespCh chan<- [2]int, watchGames map[int]*pub.WatchChan,
	playerId int, receivedInvites map[int]*pub.Invite, sendInvites map[int]*pub.Invite, sendCh chan<- interface{}) {
	close(doneCom)
	if gameRespCh == nil {
		close(gameRespCh)
	}
	if len(watchGames) > 0 {
		for _, ch := range watchGames {
			stopWatch(ch, playerId)
		}
	}
	clearInvites(receivedInvites, sendInvites, playerId)
	sendCh <- CloseCon("Server close the connection")
}
func actWatch(watchGames map[int]*pub.WatchChan, act *Action,
	startWatchGameCh chan<- *startWatchGameData, gameDoneCh chan struct{}, readList map[string]*pub.Data,
	sendCh chan<- interface{}, playerId int) (updList bool) {
	watchCh, found := watchGames[act.Id]
	if found {
		txt := fmt.Sprintf("Start watching id: %v faild as you are already watching", act.Id)
		sendSysMess(sendCh, txt)
	} else {
		watch := new(pub.WatchData)
		benchCh := make(chan *pub.MoveBench, 1)
		watch.Id = playerId
		watch.Send = benchCh
		select {
		case watchCh.Channel <- watch:
			go watchGameListen(watchCh, benchCh, startWatchGameCh, gameDoneCh, sendCh, act.Id, playerId)
		case <-watchCh.Close:
			txt := fmt.Sprintf("Player id: %v do not have a active game")
			sendSysMess(sendCh, txt)
			updList = true
		}
	}
	return updList
}
func watchGameListen(watchCh *pub.WatchChan, benchCh <-chan *pub.MoveBench,
	startWatchGameCh chan<- *startWatchGameData, doneCh chan struct{}, sendCh chan<- interface{},
	watchId int, playerId int) {

	initMove, initOpen := <-benchCh
	if initOpen {
		stData := new(startWatchGameData)
		stData.player = watchId
		stData.initMove = initMove
		stData.channel = watchCh
		select {
		case startWatchGameCh <- stData:
		case <-doneCh:
			watch := new(pub.WatchData)
			watch.Id = playerId
			watch.Send = nil //stop
			select {
			case watchCh.Channel <- watch:
			case <-watchCh.Close:
			}
		}
		writeStop := false
	Loop:
		for {
			move, open := <-benchCh
			if !writeStop {
				if open {
					select {
					case <-doneCh:
						writeStop = true
					default:
						select {
						case sendCh <- move:
						case <-doneCh:
							writeStop = true
						}
					}
				} else {
					stData = new(startWatchGameData)
					stData.player = watchId
					select {
					case startWatchGameCh <- stData:
					case <-doneCh:
					}
					break Loop
				}
			}
		}
	} else {
		txt := fmt.Sprintf("Failed to start watching game with player id %v", watchId)
		sendSysMessGo(sendCh, doneCh, txt) //TODO request a update list here and game
	}

}
func actWatchStop(watchGames map[int]*pub.WatchChan, act *Action,
	sendCh chan<- interface{}, playerId int) {
	watchCh, found := watchGames[act.Id]
	if found {
		stopWatch(watchCh, playerId)
		delete(watchGames, act.Id)
	} else {
		txt := fmt.Sprintf("Stop watching player %v failed", act.Id)
		sendSysMess(sendCh, txt)
	}
}
func stopWatch(watchCh *pub.WatchChan, playerId int) {
	watch := new(pub.WatchData)
	watch.Id = playerId
	watch.Send = nil //stop
	select {
	case watchCh.Channel <- watch:
	case <-watchCh.Close:
	}
}

func actMove(act *Action, gameRespCh chan<- [2]int, gameMove *pub.MoveView,
	sendCh chan<- interface{}) {
	if gameRespCh != nil && gameMove != nil {
		valid := false
		if gameMove.MovesPass {
			if act.Move[0] == -1 && act.Move[1] == -1 {
				valid = true
			}
		} else if gameMove.MovesHand != nil {
			handmoves, found := gameMove.MovesHand[strconv.Itoa(act.Move[0])]
			if found && act.Move[1] >= 0 && act.Move[1] < len(handmoves) {
				valid = true
			}
		} else if gameMove.Moves != nil {
			if act.Move[1] >= 0 && act.Move[1] < len(gameMove.Moves) {
				valid = true
			}
		}
		if valid {
			gameRespCh <- act.Move
		} else {
			txt := "Illegal Move"
			sendSysMess(sendCh, txt) //TODO All action error should be loged also
			//as it could be sign of hacking and it should not be possible to crash!!!
		}
	} else {
		txt := "Making a move with out an active game"
		sendSysMess(sendCh, txt)
	}
}
func actAccInvite(recInvites map[int]*pub.Invite, act *Action, sendCh chan<- interface{},
	startGame chan<- *pub.MoveView, doneCh chan struct{}, id int) {
	invite, found := recInvites[act.Id]
	if found {
		resp := new(pub.InviteResponse)
		resp.Responder = id
		moveRecCh := make(chan *pub.MoveView, 1)
		resp.GameChan = moveRecCh
		select {
		case <-invite.Retract:
			txt := fmt.Sprintf("Accept invite to id: %v failed invitation retracted", act.Id)
			sendSysMess(sendCh, txt)
		default:
			select {
			case invite.Response <- resp:
				go gameListen(startGame, doneCh, moveRecCh, sendCh, invite.Name)
			case <-invite.DoneComCh:
				txt := fmt.Sprintf("Accept invite to id: %v failed player done", act.Id)
				sendSysMess(sendCh, txt)
			}
		}
		delete(recInvites, invite.Inviter)
	} else {
		txt := fmt.Sprintf("Invite id %v do not exist.", act.Id)
		sendSysMess(sendCh, txt)
	}
}
func gameListen(gameCh chan<- *pub.MoveView, doneCh chan struct{}, moveRecCh <-chan *pub.MoveView,
	sendCh chan<- interface{}, oppName string) {
	initMove, initOpen := <-moveRecCh
	if initOpen {
		select {
		case gameCh <- initMove:
		case <-doneCh:
			close(initMove.MoveChan)
		}
		var writeStop bool
	Loop:
		for {
			move, open := <-moveRecCh
			if !writeStop {
				if open {
					if move.MyTurn {
						select {
						case gameCh <- move:
						case <-doneCh:
							writeStop = true
						}
					} else {
						select {
						case <-doneCh: //sendCh is buffered and closed at a later stage
							writeStop = true
						default:
							select {
							case sendCh <- move:
							case <-doneCh:
								writeStop = true
							}
						}
					}
				} else {
					select {
					case gameCh <- nil:
					case <-doneCh:
					}
					break Loop
				}
			}
		}
	} else {
		txt := fmt.Sprintf("Failed to start game with %v", oppName)
		sendSysMessGo(sendCh, doneCh, txt)
	}
}
func actSendInvite(invites map[int]*pub.Invite, respCh chan<- *pub.InviteResponse, doneCh chan struct{}, act *Action,
	readList map[string]*pub.Data, sendCh chan<- interface{}, id int, name string) (upd bool) {
	p, found := readList[strconv.Itoa(act.Id)]
	if found {
		invite := new(pub.Invite)
		invite.Inviter = id
		invite.Name = name
		invite.Response = respCh
		invite.Retract = make(chan struct{})
		invite.DoneComCh = doneCh
		select {
		case p.Invite <- invite:
			invites[id] = invite
		case <-p.DoneCom:
			m := fmt.Sprintf("Invite to Id: %v failed player done", act.Id)
			sendSysMess(sendCh, m)
			upd = true
		}
	} else {
		m := fmt.Sprintf("Invite to Id: %v failed could not find player", act.Id)
		sendSysMess(sendCh, m)
		upd = true
	}
	return upd
}
func actMessage(act *Action, readList map[string]*pub.Data, sendCh chan<- interface{},
	id int, name string) (upd bool) {
	p, found := readList[strconv.Itoa(act.Id)]
	if found {
		mess := new(pub.MesData)
		mess.Sender = id
		mess.Name = name
		mess.Message = act.Mess
		select {
		case p.Message <- mess:
		case <-p.DoneCom:
			m := fmt.Sprintf("Message to Id: %v failed", act.Id)
			sendSysMess(sendCh, m)
			upd = true
		}
	} else {
		m := fmt.Sprintf("Message to Id: %v failed", act.Id)
		sendSysMess(sendCh, m)
		upd = true
	}
	return upd
}
func actDeclineInvite(receivedInvites map[int]*pub.Invite, act *Action, playerId int) {
	invite, found := receivedInvites[act.Id]
	if found {
		declineInvite(invite, playerId)
		delete(receivedInvites, invite.Inviter)
	}
}
func declineInvite(invite *pub.Invite, playerId int) {
	resp := new(pub.InviteResponse)
	resp.Responder = playerId
	resp.GameChan = nil
	select {
	case <-invite.Retract:
	default:
		select {
		case invite.Response <- resp:
		case <-invite.DoneComCh:
		}
	}
}
func playerExist(pubList *pub.List, readList map[string]*pub.Data, id int,
	sendCh chan<- interface{}) (upd map[string]*pub.Data) {
	_, found := readList[strconv.Itoa(id)]
	if !found {
		upd = pubList.Read()
		sendCh <- upd
	} else {
		upd = readList
	}
	return upd
}

func sendSysMess(ch chan<- interface{}, txt string) {
	mess := new(pub.MesData)
	mess.Sender = -1
	mess.Name = "System"
	mess.Message = txt
	ch <- mess
}
func sendSysMessGo(sendCh chan<- interface{}, doneCh chan struct{}, txt string) {
	mess := new(pub.MesData)
	mess.Sender = -1
	mess.Name = "System"
	mess.Message = txt
	select {
	case sendCh <- mess:
	case <-doneCh:
	}
}
func clearInvites(receivedInvites map[int]*pub.Invite, sendInvites map[int]*pub.Invite,
	playerId int) {
	if len(receivedInvites) != 0 {
		resp := new(pub.InviteResponse)
		resp.Responder = playerId
		resp.GameChan = nil
		for _, invite := range receivedInvites {
			select {
			case invite.Response <- resp:
			case <-invite.Retract:
			}
		}
		for id, _ := range receivedInvites {
			delete(receivedInvites, id)
		}
	}
	if len(sendInvites) != 0 {
		for _, invite := range sendInvites {
			close(invite.Retract)
		}
		for id, _ := range sendInvites {
			delete(receivedInvites, id)
		}
	}
}

type CloseCon string

// netWrite keep reading until overflow/broken line or done message is send.
// overflow/broken line disable the write to net but keeps draining the pipe.
func netWrite(ws *websocket.Conn, dataCh <-chan interface{}, errCh chan<- error, brokenConn chan struct{}) {
	broke := false
	stop := false
Loop:
	for {
		data := <-dataCh
		_, stop = data.(CloseCon)
		if !broke {
			err := websocket.JSON.Send(ws, data)
			if err != nil {
				errCh <- err
				broke = true
			} else {
				if len(dataCh) > WRTBUFF_LIM {
					broke = true
				}
			}
			if broke {
				close(brokenConn)
			}
		}
		if stop {
			break Loop
		}
	}
}

//netRead reading data from a websocket.
//Keep reading until eof,an error or done. Done can not breake the read stream so to make sure to end loop
//ws must be closed.
//Close the channel before leaving
func netRead(ws *websocket.Conn, accCh chan<- *Action, doneCh chan struct{}, errCh chan<- error) {

Loop:
	for {
		var act Action
		err := websocket.JSON.Receive(ws, act)
		if err == io.EOF {

			break Loop
		} else if err != nil {
			errCh <- err
			break Loop //maybe to harsh
		} else {
			select {
			case accCh <- &act:
			case <-doneCh:
				break Loop
			}
		}
	}
	close(accCh)
}

type Action struct {
	ActType int
	Id      int
	Move    [2]int
	Mess    string
}

func NewAction() (a *Action) {
	a = new(Action)
	a.Id = -1
	a.Move = [2]int{-1, -1}
	return a
}

type startWatchGameData struct {
	channel  *pub.WatchChan
	initMove *pub.MoveBench
	player   int
}
