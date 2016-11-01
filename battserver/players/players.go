//The players server. Keeps track of login players.
package players

import (
	"fmt"
	bat "github.com/rezder/go-battleline/battleline"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-error/cerrors"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"strconv"
)

const (
	ACT_MESS       = 1
	ACT_INVITE     = 2
	ACT_INVACCEPT  = 3
	ACT_INVDECLINE = 4
	ACT_INVRETRACT = 5
	ACT_MOVE       = 6
	ACT_QUIT       = 7
	ACT_WATCH      = 8
	ACT_WATCHSTOP  = 9
	ACT_LIST       = 10
	ACT_SAVE       = 11

	JT_Mess         = 1
	JT_Invite       = 2
	JT_Move         = 3
	JT_BenchMove    = 4
	JT_List         = 5
	JT_CloseCon     = 6
	JT_ClearInvites = 7

	WRTBUFF_SIZE = 10
	WRTBUFF_LIM  = 8
)

type Server struct {
	JoinCh        chan *Player
	DisableCh     chan *DisData
	pubList       *pub.List
	startGameChCl *pub.StartGameChCl
	finishedCh    chan struct{}
}

func New(pubList *pub.List, startGameChCl *pub.StartGameChCl) (s *Server) {
	s = new(Server)
	s.pubList = pubList
	s.startGameChCl = startGameChCl
	s.JoinCh = make(chan *Player)
	s.DisableCh = make(chan *DisData)
	s.finishedCh = make(chan struct{})
	return s
}
func (s *Server) Start() {
	go start(s.JoinCh, s.DisableCh, s.pubList, s.startGameChCl, s.finishedCh)
}
func (s *Server) Stop() {
	if cerrors.IsVerbose() {
		log.Println("Closing players join channel")
	}
	close(s.JoinCh)
	_ = <-s.finishedCh
	if cerrors.IsVerbose() {
		log.Println("Receiving players finished")
	}
}

// Start starts the players server.
func start(joinCh <-chan *Player, disableCh <-chan *DisData, pubList *pub.List, startGameChCl *pub.StartGameChCl,
	finishedCh chan struct{}) {
	leaveCh := make(chan int)
	list := make(map[int]*pub.PlayerData)
	done := false
	// disPlayers is need because is possible for a player to be logged in but
	// not join the players server when he is booted out. Unlikely but possibel.
	disPlayers := make(map[int]bool)
Loop:
	for {
		select {
		case id := <-leaveCh:
			delete(list, id)
			if done && len(list) == 0 {
				break Loop
			}
			publish(list, pubList)
		case p, open := <-joinCh:
			if open {
				inviteCh := make(chan *pub.Invite)
				messCh := make(chan *pub.MesData)
				p.joinServer(inviteCh, messCh, leaveCh, pubList, startGameChCl)
				p.joinedCh <- p
				list[p.id] = &pub.PlayerData{p.id, p.name, inviteCh, p.doneComCh, messCh, p.bootCh}
				publish(list, pubList)
				if disPlayers[p.id] {
					close(p.bootCh)
				}
			} else {
				done = true
				if len(list) > 0 {
					for _, player := range list {
						close(player.BootCh)
					}
					joinCh = nil // do not listen any more
				} else {
					break Loop
				}
			}
		case disData := <-disableCh:
			if !done {
				if !disData.Disable {
					delete(disPlayers, disData.PlayerId)
				} else {
					if !disPlayers[disData.PlayerId] {
						disPlayers[disData.PlayerId] = true
						p, found := list[disData.PlayerId]
						if found {
							close(p.BootCh)
						}
					}
				}
			}
		}
	}
	close(finishedCh)
}

// publish publish the players map.
// Asyncronized.
func publish(list map[int]*pub.PlayerData, pubList *pub.List) {
	copy := make(map[int]*pub.PlayerData)
	for key, v := range list {
		copy[key] = v
	}
	go pubList.UpdatePlayers(copy)
}

// Player it is the player information used to exchange between the logon server and
// the players server.
type Player struct {
	id          int
	name        string
	tableStChCl *pub.StartGameChCl
	leaveCh     chan<- int
	pubList     *pub.List
	inviteCh    <-chan *pub.Invite
	doneComCh   chan struct{}
	messCh      <-chan *pub.MesData
	ws          *websocket.Conn
	errCh       chan<- error
	bootCh      chan struct{} //server boot channel. used to kick player out.
	joinedCh    chan<- *Player
}

func NewPlayer(id int, name string, ws *websocket.Conn, errCh chan<- error,
	joinedCh chan<- *Player) (p *Player) {
	p = new(Player)
	p.id = id
	p.name = name
	p.ws = ws
	p.doneComCh = make(chan struct{})
	p.errCh = errCh
	p.bootCh = make(chan struct{})
	p.joinedCh = joinedCh
	return p
}

// joinServer add the players server information.
func (p *Player) joinServer(inviteCh <-chan *pub.Invite, messCh <-chan *pub.MesData,
	leaveCh chan<- int, pubList *pub.List, startGameChCl *pub.StartGameChCl) {
	p.leaveCh = leaveCh
	p.pubList = pubList
	p.inviteCh = inviteCh
	p.messCh = messCh
	p.tableStChCl = startGameChCl
}

// Start a player server.
// The server will not wait for its started go routine to finsh but it will send
// the kill signal to them.
func (player *Player) Start() {
	sendCh := make(chan interface{}, WRTBUFF_SIZE)
	wrtBrookCh := make(chan struct{})
	wrtDoneCh := make(chan struct{})
	readList := player.pubList.Read()
	go netWrite(player.ws, sendCh, player.errCh, wrtBrookCh, wrtDoneCh)

	c := make(chan *Action, 1)
	go netRead(player.ws, c, player.doneComCh, player.errCh)
	var actChan <-chan *Action
	actChan = c

	inviteResponse := make(chan *pub.InviteResponse)

	sendInvites := make(map[int]*pub.Invite)
	receivedInvites := make(map[int]*pub.Invite)

	gameReceiveCh := make(chan *pub.MoveView)

	gameState := new(GameState)

	watchGameCh := make(chan *startWatchGameData)
	var watchGames map[int]*pub.WatchChCl
	sendCh <- readList
	sendSysMess(sendCh, "Welcome to Battleline!\n Will you play a game?")
Loop:
	for {
		select {
		case <-player.bootCh:
			handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
				receivedInvites, sendInvites, sendCh, wrtDoneCh, false)
			break Loop
		case <-wrtBrookCh:
			handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
				receivedInvites, sendInvites, sendCh, wrtDoneCh, false)
			break Loop
		case act, open := <-actChan:
			if open {
				upd := false
				switch act.ActType {
				case ACT_MESS:
					upd = actMessage(act, readList, sendCh, player.id, player.name)
				case ACT_INVITE:
					upd = actSendInvite(sendInvites, inviteResponse, player.doneComCh,
						act, readList, sendCh, player.id, player.name, gameState)
				case ACT_INVACCEPT:
					upd = actAccInvite(receivedInvites, act, sendCh, gameReceiveCh, player.doneComCh,
						player.id, gameState)
				case ACT_INVDECLINE:
					upd = actDeclineInvite(receivedInvites, act, player.id)
				case ACT_INVRETRACT:
					actRetractInvite(sendInvites, act)
				case ACT_MOVE:
					if cerrors.IsVerbose() {
						log.Printf("handle move for player id: %v Name: %v\n Move: %v", player.id, player.name, act)
					}
					actMove(act, gameState, sendCh, player.errCh, player.id)
				case ACT_SAVE:
					if gameState.waitingForClient() || gameState.waitingForServer() {
						gameState.closeChannel()
					} else {
						errTxt := "Requesting game save with no game active!"
						sendSysMess(sendCh, errTxt)
						player.errCh <- cerrors.Wrap(NewPlayerErr(errTxt, player.id), 19, "")
					}
				case ACT_QUIT:
					if gameState.waitingForClient() {
						gameState.respCh <- [2]int{0, pub.SM_Quit}
					} else {
						sendSysMess(sendCh, "Quitting game out of turn is not possible.")
						player.errCh <- cerrors.Wrap(NewPlayerErr("Quitting game out of turn", player.id), 20, "")
					}
				case ACT_WATCH:
					upd = actWatch(watchGames, act, watchGameCh, player.doneComCh, readList,
						sendCh, player.id)
				case ACT_WATCHSTOP:
					actWatchStop(watchGames, act, sendCh, player.id)
				case ACT_LIST:
					upd = true
				default:
					player.errCh <- cerrors.Wrap(NewPlayerErr("Action do not exist", player.id), 21, "")
				}
				if upd {
					readList = player.pubList.Read()
					sendCh <- readList
				}
			} else {
				handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
					receivedInvites, sendInvites, sendCh, wrtDoneCh, true)
				break Loop
			}

		case move := <-gameReceiveCh:
			readList = handleGameReceive(move, sendCh, player.pubList, readList, gameState,
				receivedInvites, sendInvites, player.id)
		case wgd := <-watchGameCh:
			if wgd.chCl == nil { //set chan to nil for done
				delete(watchGames, wgd.player)
			} else {
				watchGames[wgd.player] = wgd.chCl
				sendCh <- wgd.initMove
			}
		case invite := <-player.inviteCh:
			readList = playerExist(player.pubList, readList, invite.InvitorId, sendCh)
			_, found := readList[strconv.Itoa(invite.InvitorId)]
			if found {
				receivedInvites[invite.InvitorId] = invite
				sendCh <- invite
			}

		case response := <-inviteResponse:
			invite, found := sendInvites[response.Responder]
			if found {
				handleInviteResponse(response, invite, sendInvites, sendCh, player.id, player.doneComCh,
					gameReceiveCh, player.tableStChCl)
			} else {
				if response.GameCh != nil {
					close(response.GameCh)
				}
			}
		case message := <-player.messCh:
			readList = playerExist(player.pubList, readList, message.Sender, sendCh)
			_, found := readList[strconv.Itoa(message.Sender)]
			if found {
				sendCh <- message
			}
		}
	}
	player.leaveCh <- player.id
}

//GameState keeps track of the current game.
type GameState struct {
	respCh   chan<- [2]int
	lastMove *pub.MoveView
	closed   bool //chanel closed
	hasMoved bool
}

//waitingForServer the player is waiting for the server to return a move.
func (state *GameState) waitingForServer() (res bool) {
	if state.respCh != nil && state.lastMove.State != bat.TURNFinish &&
		state.lastMove.State != bat.TURNQuit {
		if state.lastMove.MyTurn {
			if state.hasMoved || state.closed {
				res = true
			}
		} else {
			res = true
		}
	}
	return res
}

//waitingForClient the client to make a move.
func (state *GameState) waitingForClient() (res bool) {
	if state.respCh != nil && state.lastMove.State != bat.TURNFinish &&
		state.lastMove.State != bat.TURNQuit {
		if state.lastMove.MyTurn {
			if !state.hasMoved || !state.closed {
				res = true
			}
		}
	}
	return res
}

//removeGame removes game.
func (state *GameState) removeGame() {
	state.respCh = nil
	state.lastMove = nil
	state.closed = false
}

//addGame adds a game.
func (state *GameState) addGame(respCh chan<- [2]int, move *pub.MoveView) {
	state.respCh = respCh
	state.lastMove = move
}

//receiveMove receives a move.
func (state *GameState) receiveMove(move *pub.MoveView) {
	state.lastMove = move
	state.hasMoved = false
}

//sendMove to table.
func (state *GameState) sendMove(move [2]int) {
	state.hasMoved = true
	state.respCh <- move
}
func (state *GameState) hasGame() bool {
	return state.respCh != nil
}

// closeChannel close the table move channel to signal save game.
func (state *GameState) closeChannel() {
	if state.waitingForClient() || state.waitingForServer() {
		close(state.respCh)
		state.closed = true
	}
}

//actRetractInvite retract an invite
func actRetractInvite(sendInvites map[int]*pub.Invite, act *Action) {
	invite, found := sendInvites[act.Id]
	if found {
		close(invite.Retract)
		delete(sendInvites, act.Id)
	}
}

//handleInviteResponse handle a invite response.
//# sendInvites
func handleInviteResponse(response *pub.InviteResponse, invite *pub.Invite,
	sendInvites map[int]*pub.Invite, sendCh chan<- interface{}, playerId int, doneComCh chan struct{},
	gameReceiveCh chan<- *pub.MoveView, tableStChCl *pub.StartGameChCl) {
	_, found := sendInvites[response.Responder] //if not found we rejected but he did not get the msg yet
	if found {
		delete(sendInvites, response.Responder)
		if response.GameCh == nil {
			invite.Rejected = true
			sendCh <- invite
		} else {
			moveRecCh := make(chan *pub.MoveView, 1)
			tableData := new(pub.StartGameData)
			tableData.PlayerIds = [2]int{playerId, response.Responder}
			tableData.PlayerChs = [2]chan<- *pub.MoveView{moveRecCh, response.GameCh}
			select {
			case tableStChCl.Channel <- tableData:
				go gameListen(gameReceiveCh, doneComCh, moveRecCh, sendCh, invite)
			case <-tableStChCl.Close:
				txt := fmt.Sprintf("Failed to start game with %v as server no longer accept games",
					response.Name)
				sendSysMess(sendCh, txt)
			}
		}
	}
}

//handleGameReceive handles the recieve move from the game server.
// #receivedInvites
// #sendInvites
// #gameState
func handleGameReceive(move *pub.MoveView, sendCh chan<- interface{}, pubList *pub.List,
	readList map[string]*pub.Data, gameState *GameState, receivedInvites map[int]*pub.Invite,
	sendInvites map[int]*pub.Invite, playerId int) (updReadList map[string]*pub.Data) {
	updReadList = readList
	if move == nil {
		gameState.removeGame()
	} else {
		if !gameState.hasGame() { //Init move
			gameState.addGame(move.MoveCh, move)
			clearInvites(receivedInvites, sendInvites, playerId)
			sendCh <- ClearInvites("All invites was clear as game starts.")
			sendCh <- move
			updReadList = pubList.Read()
			sendCh <- updReadList
		} else {
			gameState.receiveMove(move)
			sendCh <- move
		}
	}
	return updReadList
}

// handleCloseDown close down the player.
//# receivedInvites
//# sendInvites
//# gameState
func handleCloseDown(doneComCh chan struct{}, gameState *GameState,
	watchGames map[int]*pub.WatchChCl, playerId int, receivedInvites map[int]*pub.Invite,
	sendInvites map[int]*pub.Invite, sendCh chan<- interface{}, wrtDoneCh chan struct{}, playerCloseConn bool) {
	close(doneComCh)
	gameState.closeChannel()
	if len(watchGames) > 0 {
		for _, ch := range watchGames {
			stopWatch(ch, playerId)
		}
	}
	clearInvites(receivedInvites, sendInvites, playerId)

	sendCh <- CloseCon{playerCloseConn, "Server close the connection"}
	_ = <-wrtDoneCh
}

//actWatch handle client action watch a game.
//# watchGames
func actWatch(watchGames map[int]*pub.WatchChCl, act *Action,
	startWatchGameCh chan<- *startWatchGameData, gameDoneCh chan struct{}, readList map[string]*pub.Data,
	sendCh chan<- interface{}, playerId int) (updList bool) {
	_, found := watchGames[act.Id] // This test only works for game started, the bench server should reject
	// any repeat request to start game.
	if found {
		txt := fmt.Sprintf("Start watching id: %v faild as you are already watching", act.Id)
		sendSysMess(sendCh, txt)
	} else {
		watch := new(pub.WatchData)
		benchCh := make(chan *pub.MoveBench, 1)
		watch.Id = playerId
		watch.Send = benchCh
		watchChCl := pub.NewWatchChCl()
		select {
		case watchChCl.Channel <- watch:
			go watchGameListen(watchChCl, benchCh, startWatchGameCh, gameDoneCh, sendCh, act.Id, playerId)
		case <-watchChCl.Close:
			txt := fmt.Sprintf("Player id: %v do not have a active game", act.Id)
			sendSysMess(sendCh, txt)
			updList = true
		}
	}
	return updList
}

// watchGameListen handle game information from a watch game.
// If doneCh is closed it will close the game connection in init stage, and
// in normal state it will stop resending but keep reading until connection is closed.
func watchGameListen(watchChCl *pub.WatchChCl, benchCh <-chan *pub.MoveBench,
	startWatchGameCh chan<- *startWatchGameData, doneCh chan struct{}, sendCh chan<- interface{},
	watchId int, playerId int) {

	initMove, initOpen := <-benchCh
	if initOpen {
		stData := new(startWatchGameData)
		stData.player = watchId
		stData.initMove = initMove
		stData.chCl = watchChCl
		select {
		case startWatchGameCh <- stData:
		case <-doneCh:
			stopWatch(watchChCl, playerId)
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
				} else { //Game is finished.
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
		txt := fmt.Sprintf("Failed to start watching game with player id %v\n.There could be two reason for this one your already started watching or the game is finished.", watchId)
		sendSysMessGo(sendCh, doneCh, txt)
	}

}

// actWatchStop stop watching a game.
func actWatchStop(watchGames map[int]*pub.WatchChCl, act *Action,
	sendCh chan<- interface{}, playerId int) {
	watchChCl, found := watchGames[act.Id]
	if found {
		stopWatch(watchChCl, playerId)
		delete(watchGames, act.Id)
	} else {
		txt := fmt.Sprintf("Stop watching player %v failed", act.Id)
		sendSysMess(sendCh, txt)
	}
}

// stopWatch sends the stop watch signal if possibel.
func stopWatch(watchChCl *pub.WatchChCl, playerId int) {
	watch := new(pub.WatchData)
	watch.Id = playerId
	watch.Send = nil //stop
	select {
	case watchChCl.Channel <- watch:
	case <-watchChCl.Close:
	}
}

// actMove makes a game move if the move is valid.
func actMove(act *Action, gameState *GameState, sendCh chan<- interface{}, errCh chan<- error, id int) {
	if gameState.waitingForClient() {
		valid := false
		lastMove := gameState.lastMove
		if lastMove.MovesPass {
			if act.Move[1] == pub.SM_Pass {
				valid = true
			}
		}
		if lastMove.MovesHand != nil {
			handmoves, found := lastMove.MovesHand[strconv.Itoa(act.Move[0])]
			if found && act.Move[1] >= 0 && act.Move[1] < len(handmoves) {
				valid = true
			}
		} else if lastMove.Moves != nil {
			if act.Move[1] >= 0 && act.Move[1] < len(lastMove.Moves) {
				valid = true
			}
		}
		if valid {
			if cerrors.IsVerbose() {
				log.Println("Sending move to table")
			}
			gameState.sendMove(act.Move)
		} else {
			txt := "Illegal Move"
			sendSysMess(sendCh, txt)
			errCh <- cerrors.Wrap(NewPlayerErr("Illegal move", id), 22, "")
		}
	} else {
		txt := "Is not your time to move.!"
		sendSysMess(sendCh, txt)
		errCh <- cerrors.Wrap(NewPlayerErr("Move out of turn", id), 23, "")
	}
}

// actAccInvite accept a invite.
func actAccInvite(recInvites map[int]*pub.Invite, act *Action, sendCh chan<- interface{},
	startGame chan<- *pub.MoveView, doneCh chan struct{}, id int, gameState *GameState) (upd bool) {

	invite, found := recInvites[act.Id]
	if found {
		if !gameState.hasGame() {
			resp := new(pub.InviteResponse)
			resp.Responder = id
			moveRecCh := make(chan *pub.MoveView, 1)
			resp.GameCh = moveRecCh
			select {
			case <-invite.Retract:
				txt := fmt.Sprintf("Accepting invite from %v failed invitation retracted", invite.InvitorName)
				sendSysMess(sendCh, txt)
			default:
				select {
				case invite.Response <- resp:
					go gameListen(startGame, doneCh, moveRecCh, sendCh, invite)
				case <-invite.DoneComCh:
					txt := fmt.Sprintf("Accepting invite from %v failed player done", invite.InvitorName)
					sendSysMess(sendCh, txt)
					upd = true
				}
			}
			delete(recInvites, invite.InvitorId)
		} else {
			txt := fmt.Sprintf("Invite from %v was not accepted as game is in progress", invite.InvitorName)
			sendSysMess(sendCh, txt)
		}
	} else {
		txt := fmt.Sprintf("Invite id %v do not exist.", act.Id)
		sendSysMess(sendCh, txt)
	}
	return upd
}

// gameListen listen for game moves.
// If doneCh is closed and init state the game response channel is closed else
// the listener keep listening until the channel is close but it do not resend the moves.
//
func gameListen(gameCh chan<- *pub.MoveView, doneCh chan struct{}, moveRecCh <-chan *pub.MoveView,
	sendCh chan<- interface{}, invite *pub.Invite) {
	initMove, initOpen := <-moveRecCh
	if initOpen {
		select {
		case gameCh <- initMove:
		case <-doneCh:
			close(initMove.MoveCh)
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
		invite.Rejected = true
		sendCh <- invite
	}
}

//actSendInvite send a invite.
func actSendInvite(invites map[int]*pub.Invite, respCh chan<- *pub.InviteResponse, doneCh chan struct{},
	act *Action, readList map[string]*pub.Data, sendCh chan<- interface{}, id int, name string,
	gameState *GameState) (upd bool) {
	invite := new(pub.Invite)
	invite.InvitorId = id
	invite.InvitorName = name
	p, found := readList[strconv.Itoa(act.Id)]
	if found {
		invite.ReceiverId = p.Id
		if !gameState.hasGame() {
			invite.Response = respCh
			invite.Retract = make(chan struct{})
			invite.DoneComCh = doneCh
			select {
			case p.Invite <- invite:
				invites[p.Id] = invite
			case <-p.DoneCom:
				invite.Rejected = true
				sendCh <- invite
				m := fmt.Sprintf("Invite to %v failed player done", p.Name)
				sendSysMess(sendCh, m)
				upd = true
			}
		} else {
			invite.Rejected = true
			sendCh <- invite
			m := fmt.Sprintf("Invite to %v cannot be extended while playing", p.Name)
			sendSysMess(sendCh, m)
		}
	} else {
		invite.Rejected = true
		sendCh <- invite
		m := fmt.Sprintf("Invite to Id: %v failed could not find player", act.Id)
		sendSysMess(sendCh, m)
		upd = true
	}
	return upd
}

// actMessage send a message.
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

// actDeclineInvite decline a invite.
func actDeclineInvite(receivedInvites map[int]*pub.Invite, act *Action, playerId int) (upd bool) {
	invite, found := receivedInvites[act.Id]
	if found {
		upd = declineInvite(invite, playerId)
		delete(receivedInvites, invite.InvitorId)
	}
	return upd
}

// declineInvite send the decline signal to opponent.
func declineInvite(invite *pub.Invite, playerId int) (upd bool) {
	resp := new(pub.InviteResponse)
	resp.Responder = playerId
	resp.GameCh = nil
	select {
	case <-invite.Retract:
	default:
		select {
		case invite.Response <- resp:
		case <-invite.DoneComCh:
			upd = true
		}
	}
	return upd
}

// playerExist check if a player exist in the public list if not update the list.
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

// sendSysMess send a system message, there send channel is not check for closed,
// it is assumed to be open. Use go instead if not sure.
func sendSysMess(ch chan<- interface{}, txt string) {
	mess := new(pub.MesData)
	mess.Sender = -1
	mess.Name = "System"
	mess.Message = txt
	ch <- mess
}

// sendSysMessGo send a system message if possible.
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

// clearInvites clear invites. Retract all send and cancel all recieved.
func clearInvites(receivedInvites map[int]*pub.Invite, sendInvites map[int]*pub.Invite,
	playerId int) {
	if len(receivedInvites) != 0 {
		resp := new(pub.InviteResponse)
		resp.Responder = playerId
		resp.GameCh = nil
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

// CloseCon close connection signal to netWrite.
type CloseCon struct {
	player bool
	Reason string
}

// Clear invites signal
type ClearInvites string

// netWrite keep reading until overflow/broken line or done message is send.
// overflow/broken line disable the write to net but keeps draining the pipe.
func netWrite(ws *websocket.Conn, dataCh <-chan interface{}, errCh chan<- error,
	brokenConn chan struct{}, doneCh chan struct{}) {
	broke := false
	serverStop := false
	playerStop := false
	stop := false
	var closeCon CloseCon
Loop:
	for {
		data := <-dataCh
		closeCon, stop = data.(CloseCon)
		if stop {
			if closeCon.player {
				playerStop = true
			} else {
				serverStop = true
			}
		}
		if !broke && !playerStop {
			err := websocket.JSON.Send(ws, netWrite_AddJsonType(data))
			if err != nil {
				errCh <- cerrors.Wrap(err, 24, "Websocket send")
				broke = true
			} else {
				if len(dataCh) > WRTBUFF_LIM {
					broke = true
				}
			}
			if broke && !serverStop {
				close(brokenConn)
			}
		}
		if stop {
			break Loop
		}
	}
	close(doneCh)
}

//netWrite_AddJsonType adds the json type to interface values.
func netWrite_AddJsonType(data interface{}) (jdata *JsonData) {
	jdata = new(JsonData)
	jdata.Data = data
	switch data.(type) {
	case map[string]*pub.Data:
		jdata.JsonType = JT_List
	case *pub.Invite:
		jdata.JsonType = JT_Invite
	case *pub.MesData:
		jdata.JsonType = JT_Mess
	case *pub.MoveView:
		jdata.JsonType = JT_Move
	case *pub.MoveBench:
		jdata.JsonType = JT_BenchMove
	case CloseCon:
		jdata.JsonType = JT_CloseCon
	case ClearInvites:
		jdata.JsonType = JT_ClearInvites
	default:
		txt := fmt.Sprintf("Message not implemented yet: %v\n", data)
		panic(txt)
	}
	return jdata
}

//netRead reading data from a websocket.
//Keep reading until eof,an error or done. Done can not breake the read stream
//so to make sure to end loop websocket must be closed.
//It always close the channel before leaving.
func netRead(ws *websocket.Conn, accCh chan<- *Action, doneCh chan struct{}, errCh chan<- error) {

Loop:
	for {
		var act Action
		err := websocket.JSON.Receive(ws, &act)
		if cerrors.IsVerbose() {
			log.Printf("Action received: %v, Error: %v\n", act, err)
		}
		if err == io.EOF {
			break Loop
		} else if err != nil {
			select {
			case <-doneCh:
			default:
				errCh <- cerrors.Wrap(err, 25, "Websocket receive")
			}
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

// Action the client action.
type Action struct {
	ActType int
	Id      int
	Move    [2]int
	Mess    string
}

func NewAction() (a *Action) {
	a = new(Action)
	a.Move = [2]int{0, pub.SM_None}
	return a
}

// startWatchGameData the data send on startWatchGameCh
// to inform player server about starting and stoping to watch a game.
type startWatchGameData struct {
	chCl     *pub.WatchChCl
	initMove *pub.MoveBench
	player   int
}

//PlayerErr a error message that include the player id.
type PlayerErr struct {
	PlayerId int
	Txt      string
}

func NewPlayerErr(txt string, id int) (err error) {
	return &PlayerErr{id, txt}
}
func (err *PlayerErr) Error() string {
	return fmt.Sprintf(err.Txt+" Error reported on player id: %v", err.PlayerId)
}

type DisData struct {
	Disable  bool
	PlayerId int
}

//JsonData a json wrapper for export of interface values via json.
type JsonData struct {
	JsonType int
	Data     interface{}
}
