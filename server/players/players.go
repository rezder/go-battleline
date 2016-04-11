package players

import (
	"fmt"
	"golang.org/x/net/websocket"
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
)

func Start(join chan *Player, pubList *pub.List, startGameCh *tables.StartGameChan, finished chan struct{}) {
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
				list[p.id] = &pub.PlayerData{p.id, p.name, invite, p.doneCom, mess}
				publish(list, pubList)
			} else {
				if len(list) > 0 {
					for _, player := range list {
						close(player.DoneCom) //no send invite or message
						close(player.Invite)  //close only invite should be ennough of a signal
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
	closed    bool
}

func NewPlayer(id int, name string, ws *websocket.Conn, errChan chan<- error) (p *Player) {
	p = new(Player)
	p.id = id
	p.name = name
	p.ws = ws
	p.doneCom = make(chan struct{})
	p.err = errChan
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
func (player *Player) closeConn() {
	if !player.closed {
		err := player.ws.Close()
		if err != nil {
			player.err <- err
		}
	}
}
func (player *Player) Start() {
	defer player.closeConn()
	sendChan := make(chan interface{}, 10)
	brokeConn := make(chan struct{})
	slowConn := make(chan struct{})
	readList := player.pubList.Read()
	go netSend(player.ws, sendChan, player.err, brokeConn, slowConn, readList)

	c := make(chan *Action, 1)
	go netReceive(player.ws, c, player.err, brokeConn)
	var actChan <-chan *Action
	actChan = c

	inviteResponse := make(chan *pub.InviteResponse)

	sendInvites := make(map[int]*pub.Invite)
	receivedInvites := make(map[int]*pub.Invite)

	gameChReceive := make(chan *pub.MoveView) //if already sat game is finished and
	//gameChan should be sat to nil
	var gameChResp chan<- [2]int
	var gameMove *pub.MoveView

	watchGameChan := make(chan *startWatchGameData)
	var watchGames map[int]*pub.WatchChan

	initDoneCh := make(chan struct{})
	stopCom := false
	stopClient := false
Loop:
	for {
		select {
		case <-brokeConn: //TODO i do not think we need it here only in netwrite close on action should be ok (if ws.close trikker iof)
			stop = true
			//TODO
		case <-slowConn:
			player.ws.Close() // is only here to do line below else it could be in write
			player.closed = true
		case act := <-actChan:
			upd := false
			switch act.ActType { //TODO Add open. When closed client disabled. Request disable com if not already dissable if both disable finsih Check stop games
			//stop Com and stop client / stop game, stop watches, clear invites/ finsihCheck
			case ACT_MESS:
				upd = actMessage(act, readList, sendChan, player.id, player.name)
			case ACT_INVITE:
				upd = actSendInvite(sendInvites, inviteResponse, act, readList, sendChan, player.id,
					player.name)
			case ACT_INVACCEPT:
				actAccInvite(receivedInvites, act, sendChan, gameChReceive, initDoneCh, player.id)
			case ACT_INVDECLINE:
				actDeclineInvite(receivedInvites, act, player.id)
			case ACT_MOVE:
				actMove(act, gameChResp, gameMove, sendChan)
			case ACT_QUIT:
				if gameChResp != nil {
					close(gameChResp)
				} else {
					sendSysMess(sendChan, "Quitting game with no game active")
				}
			case ACT_WATCH:
				upd = actWatch(watchGames, act, watchGameChan, initDoneCh, readList, sendChan, player.id)
			case ACT_WATCHSTOP:
				actWatchStop(watchGames, act, sendChan, player.id)
			default:
				sendSysMess(sendChan, "No existen action")
			}
			if upd {
				sendChan <- player.pubList.Read()
			}
			//else //closed stopClient=true if stopCom clearInvites and finsihCheck else request stopcom
		case move := <-gameChReceive: //TODO sep chan for init
			if move == nil {
				gameChResp = nil
				gameMove = nil
				if finishCheck(stop, gameChResp, watchGames) {
					break Loop
				}
			} else {
				if gameChResp == nil { //Init move
					gameChResp = move.MoveChan
					gameMove = move
					clearInvites(receivedInvites, sendInvites, player.id)
					sendChan <- move
					readList = player.pubList.Read()
					sendChan <- readList
				} else {
					gameChResp = move.MoveChan
					gameMove = move
					sendChan <- move
				}

			}
		case wgd := <-watchGameChan:
			if wgd.channel == nil { //sep chan for done
				delete(watchGames, wgd.player)
				if finishCheck(stopCom, stopClient, gameChResp, watchGames) {
					break Loop
				}
			} else {
				watchGames[wgd.player] = wgd.channel
				sendChan <- wgd.initMove
			}
		case invite, open := <-player.invite:
			if open {
				readList = playerExist(player.pubList, readList, invite.Inviter, sendChan)
				_, found := readList[strconv.Itoa(invite.Inviter)]
				if found {
					receivedInvites[invite.Inviter] = invite
					sendChan <- invite
				}
			} else {
				//TODO message still open????
				//Send stop message to client if not stopClient
				stopCom = true
				//if stopClient clearInvites(receivedInvites, sendInvites, player.id) and finishCheck else close ws stop games
				// TODO retract invites do not work we must add doneCom and reject accept reponse to invite that does not exist and do a imperfect check before accept game ? or maybe something else

				if gameChResp != nil { //stop game
					close(gameChResp)
				}
				close(initDoneCh) //stop init game and watch game
				//TODO stop all watch games

				//TODO stop every thing
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
					sendChan <- mess
				} else {
					moveRecCh := make(chan *pub.MoveView, 1)
					tableData := new(tables.StartGameData)
					tableData.PlayerIds = [2]int{player.id, response.Responder}
					tableData.PlayerChans = [2]chan<- *pub.MoveView{moveRecCh, response.GameChan}
					select {
					case player.tableStCh.Channel <- tableData:
						go gameListen(gameChReceive, initDoneCh, moveRecCh, sendChan, invite.Name)
					case <-player.tableStCh.Close:
						txt := fmt.Sprintf("Failed to start game with %v as server no longer accept games",
							response.Name)
						sendSysMess(sendChan, txt)
					}
				}
			} else {
				//TODO new way of handle retracted invite
				// close(response.GameChan) at least this
			}
		case message := <-player.mess:
			readList = playerExist(player.pubList, readList, message.Sender, sendChan)
			_, found := readList[strconv.Itoa(message.Sender)]
			if found {
				sendChan <- message
			}
		}
	}
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
			txt := fmt.Sprintf("Player id: %v do not have a active")
			sendSysMess(sendCh, txt)
			updList = true
		}
	}
	return updList
}
func watchGameListen(watchCh *pub.WatchChan, benchCh <-chan *pub.MoveBench,
	startWatchGameCh chan<- *startWatchGameData, initDoneCh chan struct{}, sendCh chan<- interface{},
	watchId int, playerId int) {

	initMove, initOpen := <-benchCh
	if initOpen {
		stData := new(startWatchGameData)
		stData.player = watchId
		stData.initMove = initMove
		stData.channel = watchCh
		select {
		case startWatchGameCh <- stData:
		case <-initDoneCh:
			watch := new(pub.WatchData)
			watch.Id = playerId
			watch.Send = nil //stop
			select {
			case watchCh.Channel <- watch:
			case <-watchCh.Close:
			}
		}
		for {
			move, open := <-benchCh
			if open {
				sendCh <- move
			} else {
				stData = new(startWatchGameData)
				stData.player = watchId
				startWatchGameCh <- stData //TODO new channel
			}
		}
	} else {
		txt := fmt.Sprintf("Failed to start watching game with player id %v", watchId)
		sendSysMess(sendCh, txt) //TODO request a update list here and game
	}

}
func actWatchStop(watchGames map[int]*pub.WatchChan, act *Action,
	sendCh chan<- interface{}, playerId int) {
	watchCh, found := watchGames[act.Id]
	if found {
		watch := new(pub.WatchData)
		watch.Id = playerId
		watch.Send = nil //stop
		select {
		case watchCh.Channel <- watch:
		case <-watchCh.Close:
			delete(watchGames, act.Id)
		}
	} else {
		txt := fmt.Sprintf("Stop watching player %v failed", act.Id)
		sendSysMess(sendCh, txt)
	}
}

func actMove(act *Action, gameChResp chan<- [2]int, gameMove *pub.MoveView,
	sendChan chan<- interface{}) {
	if gameChResp != nil && gameMove != nil {
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
			gameChResp <- act.Move
		} else {
			txt := "Illegal Move"
			sendSysMess(sendChan, txt) //TODO All action error should be loged also
			//as it could be sign of hacking and it should not be possible to crash!!!
		}
	} else {
		txt := "Making a move with out an active game"
		sendSysMess(sendChan, txt)
	}
}
func actAccInvite(recInvites map[int]*pub.Invite, act *Action, sendCh chan<- interface{},
	startGame chan<- *pub.MoveView, gamesDoneCh chan struct{}, id int) {
	invite, found := recInvites[act.Id]
	if found {
		resp := new(pub.InviteResponse)
		resp.Responder = id
		moveRecCh := make(chan *pub.MoveView, 1)
		resp.GameChan = moveRecCh
		select {
		case invite.Response <- resp:
			go gameListen(startGame, gamesDoneCh, moveRecCh, sendCh, invite.Name)
			delete(recInvites, invite.Inviter)
		case <-invite.Retract: //TODO do not work as reponse is not closed
			txt := fmt.Sprintf("Accept invite to id: %v failed invitation retracted", act.Id)
			sendSysMess(sendCh, txt)
		}
	} else {
		txt := fmt.Sprintf("Invite id %v do not exist.", act.Id)
		sendSysMess(sendCh, txt)
	}
}
func gameListen(gameCh chan<- *pub.MoveView, initDoneCh chan struct{}, moveRecCh <-chan *pub.MoveView,
	sendCh chan<- interface{}, oppName string) {
	initMove, initOpen := <-moveRecCh
	if initOpen {
		select {
		case gameCh <- initMove: //TODO need another channel
		case <-initDoneCh:
			close(initMove.MoveChan)
		}
		for {
			move, open := <-moveRecCh
			if open {
				if move.MyTurn {
					gameCh <- move
				} else {
					sendCh <- move
				}
			} else {
				gameCh <- nil
			}
		}
	} else {
		txt := fmt.Sprintf("Failed to start game with %v", oppName)
		sendSysMess(sendCh, txt)
	}
}
func actSendInvite(invites map[int]*pub.Invite, respCh chan<- *pub.InviteResponse, act *Action,
	readList map[string]*pub.Data, sendCh chan<- interface{}, id int, name string) (upd bool) {
	p, found := readList[strconv.Itoa(act.Id)]
	if found {
		invite := new(pub.Invite)
		invite.Inviter = id
		invite.Name = name
		invite.Response = respCh
		invite.Retract = make(chan struct{})
		p.Invite <- invite
	} else {
		m := fmt.Sprintf("Invite to Id: %v failed", act.Id)
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
		p.Message <- mess
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
		resp := new(pub.InviteResponse)
		resp.Responder = playerId
		resp.GameChan = nil
		select {
		case invite.Response <- resp:
		case <-invite.Retract:
		}
		delete(receivedInvites, invite.Inviter)
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
func netSend(ws *websocket.Conn, toSend <-chan interface{}, err chan<- error,
	brokenConn chan struct{}, slow chan struct{}, readList map[string]*pub.Data) {
	//TODO handle web send.
	// I think when slow and broken received send should take place but the channel should drain the toSend channel
}
func netReceive(ws *websocket.Conn, accChan chan<- *Action, errChan chan<- error,
	brokenConn chan struct{}) {
	//TODO handle web read
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
