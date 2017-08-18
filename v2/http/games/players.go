package games

import (
	"fmt"
	"github.com/pkg/errors"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"io"
	"strconv"
	"time"
)

const (
	actIDMess       = 1
	actIDInvite     = 2
	actIDInvAccept  = 3
	actIDInvDecline = 4
	actIDInvRetract = 5
	actIDMove       = 6
	actIDQuit       = 7
	actIDWatch      = 8
	actIDWatchStop  = 9
	actIDList       = 10
	actIDSave       = 11

	jtMess         = 1
	jtInvite       = 2
	jtMove         = 3
	jtBenchMove    = 4
	jtList         = 5
	jtCloseCon     = 6
	jtClearInvites = 7

	wrtBuffSIZE  = 10
	wrtBuffLIMIT = 8
)

//PlayersServer a server that keep tract of the players.
type PlayersServer struct {
	JoinCh        chan *Player
	DisableCh     chan *PlayersDisData
	pubList       *PubList
	startGameChCl *StartGameChCl
	finishedCh    chan struct{}
}

//NewPlayersServer create a Players server.
func NewPlayersServer(pubList *PubList, startGameChCl *StartGameChCl) (s *PlayersServer) {
	s = new(PlayersServer)
	s.pubList = pubList
	s.startGameChCl = startGameChCl
	s.JoinCh = make(chan *Player)
	s.DisableCh = make(chan *PlayersDisData)
	s.finishedCh = make(chan struct{})
	return s
}

//Start starts the players server.
func (s *PlayersServer) Start() {
	go playersServe(s.JoinCh, s.DisableCh, s.pubList, s.startGameChCl, s.finishedCh)
}

//Stop stops the players server may take a while all player have close there games
//and log out.
func (s *PlayersServer) Stop() {
	log.Print(log.DebugMsg, "Closing players join channel")
	close(s.JoinCh)
	<-s.finishedCh
	log.Print(log.DebugMsg, "Receiving players finished")
}

// playersServe serves the players.
func playersServe(
	joinCh <-chan *Player,
	disableCh <-chan *PlayersDisData,
	pubList *PubList, startGameChCl *StartGameChCl,
	finishedCh chan struct{}) {

	leaveCh := make(chan int)
	list := make(map[int]*PlayerData)
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
			publishPlayers(list, pubList)
		case p, open := <-joinCh:
			if open {
				inviteCh := make(chan *Invite)
				messCh := make(chan *MesData)
				p.joinServer(inviteCh, messCh, leaveCh, pubList, startGameChCl)
				p.joinedCh <- p
				list[p.id] = &PlayerData{
					ID:      p.id,
					Name:    p.name,
					Invite:  inviteCh,
					DoneCom: p.doneComCh,
					Message: messCh,
					BootCh:  p.bootCh}
				publishPlayers(list, pubList)
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
					delete(disPlayers, disData.PlayerID)
				} else {
					if !disPlayers[disData.PlayerID] {
						disPlayers[disData.PlayerID] = true
						p, found := list[disData.PlayerID]
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

// publishPlayers publish the players map.
// Asyncronized.
func publishPlayers(list map[int]*PlayerData, pubList *PubList) {
	copy := make(map[int]*PlayerData)
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
	tableStChCl *StartGameChCl
	leaveCh     chan<- int
	pubList     *PubList
	inviteCh    <-chan *Invite
	doneComCh   chan struct{}
	messCh      <-chan *MesData
	ws          *websocket.Conn
	errCh       chan<- error
	bootCh      chan struct{} //server boot channel. used to kick player out.
	joinedCh    chan<- *Player
}

// NewPlayer creates a new player.
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
func (player *Player) joinServer(inviteCh <-chan *Invite, messCh <-chan *MesData,
	leaveCh chan<- int, pubList *PubList, startGameChCl *StartGameChCl) {
	player.leaveCh = leaveCh
	player.pubList = pubList
	player.inviteCh = inviteCh
	player.messCh = messCh
	player.tableStChCl = startGameChCl
}

// Start a player server.
// The server will not wait for its started go routine to finsh but it will send
// the kill signal to them.
func (player *Player) Start() {
	sendCh := make(chan interface{}, wrtBuffSIZE)
	wrtBrookCh := make(chan struct{})
	wrtDoneCh := make(chan struct{})
	readList := player.pubList.Read()
	go netWrite(player.ws, sendCh, player.errCh, wrtBrookCh, wrtDoneCh)

	c := make(chan *Action, 1)
	go netRead(player.ws, c, player.doneComCh, player.errCh)
	var actChan <-chan *Action
	actChan = c

	inviteResponseCh := make(chan *InviteResponse)

	sendInvites := make(map[int]*Invite)
	receivedInvites := make(map[int]*Invite)

	gameRecieveCh := make(chan *MoveView)

	gameState := new(GameState)

	watchGameCh := make(chan *startWatchGameData)
	var watchGames map[int]*WatchChCl
	sendCh <- readList
	sendSysMess(sendCh, "Welcome to Battleline!\nWill you play a game?")
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
				upd := handlePlayerAction(act, readList, sendCh, player, sendInvites,
					receivedInvites, inviteResponseCh, gameState, gameRecieveCh, watchGames, watchGameCh)

				if upd {
					readList = player.pubList.Read()
					sendCh <- readList
				}
			} else {
				handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
					receivedInvites, sendInvites, sendCh, wrtDoneCh, true)
				break Loop
			}

		case move := <-gameRecieveCh:
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
			readList = playerExist(player.pubList, readList, invite.InvitorID, sendCh)
			_, found := readList[strconv.Itoa(invite.InvitorID)]
			if found {
				receivedInvites[invite.InvitorID] = invite
				sendCh <- invite
			}

		case response := <-inviteResponseCh:
			invite, found := sendInvites[response.Responder]
			if found {
				handleInviteResponse(response, invite, sendInvites, sendCh, player.id, player.doneComCh,
					gameRecieveCh, player.tableStChCl)
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
func handlePlayerAction(
	act *Action,
	readList map[string]*PubData,
	sendCh chan<- interface{},
	player *Player,
	sendInvites, recievedInvites map[int]*Invite,
	inviteResponseCh chan<- *InviteResponse,
	gameState *GameState,
	gameRecieveCh chan<- *MoveView,
	watchGames map[int]*WatchChCl,
	watchGameCh chan<- *startWatchGameData) (upd bool) {

	switch act.ActType {
	case actIDMess:
		upd = actMessage(act, readList, sendCh, player.id, player.name)
	case actIDInvite:
		upd = actSendInvite(sendInvites, inviteResponseCh, player.doneComCh,
			act, readList, sendCh, player.id, player.name, gameState)
	case actIDInvAccept:
		upd = actAccInvite(recievedInvites, act, sendCh, gameRecieveCh, player.doneComCh,
			player.id, gameState)
	case actIDInvDecline:
		upd = actDeclineInvite(recievedInvites, act, player.id)
	case actIDInvRetract:
		actRetractInvite(sendInvites, act)
	case actIDMove:
		log.Printf(log.DebugMsg, "handle move for player id: %v Name: %v\n Move: %v", player.id, player.name, act)
		actMove(act, gameState, sendCh, player.errCh, player.id)
	case actIDSave:
		if gameState.waitingForClient() || gameState.waitingForServer() {
			gameState.closeChannel()
		} else {
			errTxt := "Requesting game save with no game active!"
			sendSysMess(sendCh, errTxt)
			player.errCh <- errors.Wrap(NewPlayerErr(errTxt, player.id), log.ErrNo(19))
		}
	case actIDQuit:
		if gameState.waitingForClient() {
			gameState.respCh <- [2]int{0, SMQuit}
		} else {
			sendSysMess(sendCh, "Quitting game out of turn is not possible.")
			player.errCh <- errors.Wrap(NewPlayerErr("Quitting game out of turn", player.id), log.ErrNo(20))
		}
	case actIDWatch:
		upd = actWatch(watchGames, act, watchGameCh, player.doneComCh, readList,
			sendCh, player.id)
	case actIDWatchStop:
		actWatchStop(watchGames, act, sendCh, player.id)
	case actIDList:
		upd = true
	default:
		player.errCh <- errors.Wrap(NewPlayerErr("Action do not exist", player.id), log.ErrNo(21))
	}
	return upd
}

//GameState keeps track of the current game.
type GameState struct {
	respCh   chan<- [2]int
	lastMove *MoveView
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
func (state *GameState) addGame(respCh chan<- [2]int, move *MoveView) {
	state.respCh = respCh
	state.lastMove = move
}

//receiveMove receives a move.
func (state *GameState) receiveMove(move *MoveView) {
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
func actRetractInvite(sendInvites map[int]*Invite, act *Action) {
	invite, found := sendInvites[act.ID]
	if found {
		close(invite.Retract)
		delete(sendInvites, act.ID)
	}
}

//handleInviteResponse handle a invite response.
//# sendInvites
func handleInviteResponse(response *InviteResponse, invite *Invite,
	sendInvites map[int]*Invite, sendCh chan<- interface{}, playerID int, doneComCh chan struct{},
	gameReceiveCh chan<- *MoveView, tableStChCl *StartGameChCl) {
	_, found := sendInvites[response.Responder] //if not found we rejected but he did not get the msg yet
	if found {
		delete(sendInvites, response.Responder)
		if response.GameCh == nil {
			invite.Rejected = true
			sendCh <- invite
		} else {
			moveRecCh := make(chan *MoveView, 1)
			tableData := new(StartGameData)
			tableData.PlayerIds = [2]int{playerID, response.Responder}
			tableData.PlayerChs = [2]chan<- *MoveView{moveRecCh, response.GameCh}
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
func handleGameReceive(move *MoveView, sendCh chan<- interface{}, pubList *PubList, readList map[string]*PubData, gameState *GameState, receivedInvites map[int]*Invite,
	sendInvites map[int]*Invite, playerID int) (updReadList map[string]*PubData) {
	updReadList = readList
	if move == nil {
		gameState.removeGame()
	} else {
		if !gameState.hasGame() { //Init move
			gameState.addGame(move.MoveCh, move)
			clearInvites(receivedInvites, sendInvites, playerID)
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
	watchGames map[int]*WatchChCl, playerID int, receivedInvites map[int]*Invite,
	sendInvites map[int]*Invite, sendCh chan<- interface{}, wrtDoneCh chan struct{}, playerCloseConn bool) {
	close(doneComCh)
	gameState.closeChannel()
	if len(watchGames) > 0 {
		for _, ch := range watchGames {
			stopWatch(ch, playerID)
		}
	}
	clearInvites(receivedInvites, sendInvites, playerID)

	sendCh <- CloseCon{playerCloseConn, "Server close the connection"}
	<-wrtDoneCh
}

//actWatch handle client action watch a game.
//# watchGames
func actWatch(watchGames map[int]*WatchChCl, act *Action,
	startWatchGameCh chan<- *startWatchGameData, gameDoneCh chan struct{}, readList map[string]*PubData, sendCh chan<- interface{}, playerID int) (updList bool) {
	_, found := watchGames[act.ID] // This test only works for game started, the bench server should reject
	// any repeat request to start game.
	if found {
		txt := fmt.Sprintf("Start watching id: %v failed as you are already watching", act.ID)
		sendSysMess(sendCh, txt)
	} else {
		watchChCl := findWatchCh(readList, act.ID)
		if watchChCl != nil {
			watch := new(WatchData)
			benchCh := make(chan *MoveBench, 1)
			watch.ID = playerID
			watch.Send = benchCh
			select {
			case watchChCl.Channel <- watch:
				go watchGameListen(watchChCl, benchCh, startWatchGameCh, gameDoneCh, sendCh, act.ID, playerID)
			case <-watchChCl.Close:
				txt := fmt.Sprintf("Player id: %v do not have a active game", act.ID)
				sendSysMess(sendCh, txt)
				updList = true
			}
		} else {
			txt := fmt.Sprintf("Player id: %v do not have a active game", act.ID)
			sendSysMess(sendCh, txt)
			updList = true
		}
	}
	return updList
}
func findWatchCh(pubList map[string]*PubData, watchID int) (ch *WatchChCl) {
	pubData, found := pubList[strconv.Itoa(watchID)]
	if found {
		ch = pubData.Watch
	}
	return ch
}

// watchGameListen handle game information from a watch game.
// If doneCh is closed it will close the game connection in init stage, and
// in normal state it will stop resending but keep reading until connection is closed.
func watchGameListen(watchChCl *WatchChCl, benchCh <-chan *MoveBench,
	startWatchGameCh chan<- *startWatchGameData, doneCh chan struct{}, sendCh chan<- interface{},
	watchID int, playerID int) {

	initMove, initOpen := <-benchCh
	if initOpen {
		stData := new(startWatchGameData)
		stData.player = watchID
		stData.initMove = initMove
		stData.chCl = watchChCl
		isDone := sendStartWatchData(startWatchGameCh, doneCh, stData)
		if isDone {
			stopWatch(watchChCl, playerID)
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
					stData.player = watchID
					sendStartWatchData(startWatchGameCh, doneCh, stData)
					break Loop
				}
			} else {
				if !open {
					break Loop
				}
			}
		}
	} else {
		txt := fmt.Sprintf("Failed to start watching game with player id %v\n.There could be two reason for this one your already started watching or the game is finished.", watchID)
		sendSysMessGo(sendCh, doneCh, txt)
	}

}
func sendStartWatchData(startWatchGameCh chan<- *startWatchGameData,
	doneCh chan struct{},
	stData *startWatchGameData) (isDone bool) {
	select {
	case startWatchGameCh <- stData:
	case <-doneCh:
		isDone = true
	}
	return isDone
}

// actWatchStop stop watching a game.
func actWatchStop(watchGames map[int]*WatchChCl, act *Action,
	sendCh chan<- interface{}, playerID int) {
	watchChCl, found := watchGames[act.ID]
	if found {
		stopWatch(watchChCl, playerID)
		delete(watchGames, act.ID)
	} else {
		txt := fmt.Sprintf("Stop watching player %v failed", act.ID)
		sendSysMess(sendCh, txt)
	}
}

// stopWatch sends the stop watch signal if possibel.
func stopWatch(watchChCl *WatchChCl, playerID int) {
	watch := new(WatchData)
	watch.ID = playerID
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
			if act.Move[1] == SMPass {
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
			log.Print(log.DebugMsg, "Sending move to table")
			gameState.sendMove(act.Move)
		} else {
			txt := "Illegal Move" //TODO giveup
			sendSysMess(sendCh, txt)
			errCh <- errors.Wrap(NewPlayerErr("Illegal move", id), log.ErrNo(22))
		}
	} else {
		txt := "Is not your time to move.!"
		sendSysMess(sendCh, txt)
		errCh <- errors.Wrap(NewPlayerErr("Move out of turn", id), log.ErrNo(23))
	}
}

// actAccInvite accept a invite.
func actAccInvite(recInvites map[int]*Invite, act *Action, sendCh chan<- interface{},
	startGame chan<- *MoveView, doneCh chan struct{}, id int, gameState *GameState) (upd bool) {

	invite, found := recInvites[act.ID]
	if found {
		if !gameState.hasGame() {
			resp := new(InviteResponse)
			resp.Responder = id
			moveRecCh := make(chan *MoveView, 1)
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
			delete(recInvites, invite.InvitorID)
		} else {
			txt := fmt.Sprintf("Invite from %v was not accepted as game is in progress", invite.InvitorName)
			sendSysMess(sendCh, txt)
		}
	} else {
		txt := fmt.Sprintf("Invite id %v do not exist.", act.ID)
		sendSysMess(sendCh, txt)
	}
	return upd
}

// gameListen listen for game moves.
// If doneCh is closed and init state the game response channel is closed else
// the listener keep listening until the channel is close but it do not resend the moves.
//
func gameListen(gameCh chan<- *MoveView, doneCh chan struct{}, moveRecCh <-chan *MoveView,
	sendCh chan<- interface{}, invite *Invite) {
	initMove, initOpen := <-moveRecCh
	if initOpen {
		isDone := sendMoveGameCh(gameCh, doneCh, initMove)
		if isDone {
			close(initMove.MoveCh)
		}
		var writeStop bool
	Loop:
		for {
			move, open := <-moveRecCh
			if !writeStop {
				if open {
					if move.MyTurn {
						writeStop = sendMoveGameCh(gameCh, doneCh, move)
					} else {
						writeStop = sendMoveSendCh(sendCh, doneCh, move)
					}
				} else {
					sendMoveGameCh(gameCh, doneCh, nil)
					break Loop
				}
			} else {
				if !open {
					break Loop
				}
			}
		}
	} else {
		invite.Rejected = true
		sendCh <- invite
	}
}
func sendMoveSendCh(
	sendCh chan<- interface{},
	doneCh chan struct{},
	move *MoveView) (isDone bool) {
	select {
	case <-doneCh: //sendCh is buffered and closed at a later stage
		isDone = true
	default:
		select {
		case sendCh <- move:
		case <-doneCh:
			isDone = true
		}
	}
	return isDone
}
func sendMoveGameCh(
	gameCh chan<- *MoveView,
	doneCh chan struct{},
	move *MoveView) (isDone bool) {
	select {
	case gameCh <- move:
	case <-doneCh:
		isDone = true
	}
	return isDone

}

//actSendInvite send a invite.
func actSendInvite(invites map[int]*Invite, respCh chan<- *InviteResponse, doneCh chan struct{},
	act *Action, readList map[string]*PubData, sendCh chan<- interface{}, id int, name string,
	gameState *GameState) (upd bool) {
	invite := new(Invite)
	invite.InvitorID = id
	invite.InvitorName = name
	p, found := readList[strconv.Itoa(act.ID)]
	if found {
		invite.ReceiverID = p.ID
		if !gameState.hasGame() {
			invite.Response = respCh
			invite.Retract = make(chan struct{})
			invite.DoneComCh = doneCh
			select {
			case p.Invite <- invite:
				invites[p.ID] = invite
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
		m := fmt.Sprintf("Invite to Id: %v failed could not find player", act.ID)
		sendSysMess(sendCh, m)
		upd = true
	}
	return upd
}

// actMessage send a message.
func actMessage(act *Action, readList map[string]*PubData, sendCh chan<- interface{},
	id int, name string) (upd bool) {
	p, found := readList[strconv.Itoa(act.ID)]
	if found {
		mess := new(MesData)
		mess.Sender = id
		mess.Name = name
		mess.Message = act.Mess
		select {
		case p.Message <- mess:
		case <-p.DoneCom:
			m := fmt.Sprintf("Message to Id: %v failed", act.ID)
			sendSysMess(sendCh, m)
			upd = true
		}
	} else {
		m := fmt.Sprintf("Message to Id: %v failed", act.ID)
		sendSysMess(sendCh, m)
		upd = true
	}
	return upd
}

// actDeclineInvite decline a invite.
func actDeclineInvite(receivedInvites map[int]*Invite, act *Action, playerID int) (upd bool) {
	invite, found := receivedInvites[act.ID]
	if found {
		upd = declineInvite(invite, playerID)
		delete(receivedInvites, invite.InvitorID)
	}
	return upd
}

// declineInvite send the decline signal to opponent.
func declineInvite(invite *Invite, playerID int) (upd bool) {
	resp := new(InviteResponse)
	resp.Responder = playerID
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
func playerExist(pubList *PubList, readList map[string]*PubData, id int,
	sendCh chan<- interface{}) (upd map[string]*PubData) {
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
	mess := new(MesData)
	mess.Sender = -1
	mess.Name = "System"
	mess.Message = txt
	ch <- mess
}

// sendSysMessGo send a system message if possible.
func sendSysMessGo(sendCh chan<- interface{}, doneCh chan struct{}, txt string) {
	mess := new(MesData)
	mess.Sender = -1
	mess.Name = "System"
	mess.Message = txt
	select {
	case sendCh <- mess:
	case <-doneCh:
	}
}

// clearInvites clear invites. Retract all send and cancel all recieved.
func clearInvites(receivedInvites map[int]*Invite, sendInvites map[int]*Invite,
	playerID int) {
	if len(receivedInvites) != 0 {
		resp := new(InviteResponse)
		resp.Responder = playerID
		resp.GameCh = nil
		for _, invite := range receivedInvites {
			select {
			case invite.Response <- resp:
			case <-invite.Retract:
			}
		}
		for id := range receivedInvites {
			delete(receivedInvites, id)
		}
	}
	if len(sendInvites) != 0 {
		for _, invite := range sendInvites {
			close(invite.Retract)
		}
		for id := range sendInvites {
			delete(receivedInvites, id)
		}
	}
}

// CloseCon close connection signal to netWrite.
type CloseCon struct {
	player bool
	Reason string
}

// ClearInvites signal
type ClearInvites string

// netWrite keep reading until overflow/broken line or done message is send.
// overflow/broken line disable the write to net but keeps draining the pipe.
func netWrite(
	ws *websocket.Conn,
	dataCh <-chan interface{},
	errCh chan<- error,
	brokenConn chan struct{},
	doneCh chan struct{}) {

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
			err := websocket.JSON.Send(ws, netWriteAddJSONType(data))
			if err != nil {
				errCh <- errors.Wrap(err, log.ErrNo(24)+"Websocket send")
				broke = true
			} else {
				if len(dataCh) > wrtBuffLIMIT {
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

//netWriteAddJsonType adds the json type to interface values.
func netWriteAddJSONType(data interface{}) (jdata *JsonData) {
	jdata = new(JsonData)
	jdata.Data = data
	switch data.(type) {
	case map[string]*PubData:
		jdata.JsonType = jtList
	case *Invite:
		jdata.JsonType = jtInvite
	case *MesData:
		jdata.JsonType = jtMess
	case *MoveView:
		jdata.JsonType = jtMove
	case *MoveBench:
		jdata.JsonType = jtBenchMove
	case CloseCon:
		jdata.JsonType = jtCloseCon
	case ClearInvites:
		jdata.JsonType = jtClearInvites
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
		ts := time.Now().Add(10 * time.Minute)
		err := ws.SetReadDeadline(ts)
		if err == nil {
			err = websocket.JSON.Receive(ws, &act)
			log.Printf(log.DebugMsg, "Action received: %v, Error: %v\n", act, err)
		}
		if err == io.EOF {
			break Loop
		} else if err != nil {
			select {
			case <-doneCh:
			default:
				errCh <- errors.Wrap(err, log.ErrNo(25)+"Websocket receive")
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
	ID      int
	Move    [2]int
	Mess    string
}

// NewAction creates a new action.
func NewAction(actType int) (a *Action) {
	a = new(Action)
	a.ActType = actType
	a.Move = [2]int{0, SMNone}
	return a
}

// startWatchGameData the data send on startWatchGameCh
// to inform player server about starting and stoping to watch a game.
type startWatchGameData struct {
	chCl     *WatchChCl
	initMove *MoveBench
	player   int
}

//PlayerErr a error message that include the player id.
type PlayerErr struct {
	PlayerID int
	Txt      string
}

//NewPlayerErr creates a new player specific error.
func NewPlayerErr(txt string, id int) (err error) {
	return &PlayerErr{id, txt}
}
func (err *PlayerErr) Error() string {
	return fmt.Sprintf(err.Txt+" Error reported on player id: %v", err.PlayerID)
}

// PlayersDisData is the disable or enable player information, used to send over channel.
type PlayersDisData struct {
	Disable  bool
	PlayerID int
}

//JsonData a json wrapper for export of interface values via json.
type JsonData struct {
	JsonType int
	Data     interface{}
}
