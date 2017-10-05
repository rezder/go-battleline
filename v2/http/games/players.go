package games

import (
	"fmt"
	"github.com/pkg/errors"
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"io"
	"strconv"
	"time"
)

const (
	ACTIDMess       = 1
	ACTIDInvite     = 2
	ACTIDInvAccept  = 3
	ACTIDInvDecline = 4
	ACTIDInvRetract = 5
	ACTIDMove       = 6
	ACTIDQuit       = 7
	ACTIDWatch      = 8
	ACTIDWatchStop  = 9
	ACTIDList       = 10
	ACTIDSave       = 11

	JTMess         = 1
	JTInvite       = 2
	JTPlaying      = 3
	JTWatching     = 4
	JTList         = 5
	JTCloseCon     = 6
	JTClearInvites = 7

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
	// not join the players server when he is booted out. Unlikely but possible.
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
					ID:        p.id,
					Name:      p.name,
					InviteCh:  inviteCh,
					DoneComCh: p.doneComCh,
					MessageCh: messCh,
					BootCh:    p.bootCh}
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

// Serve serves a player.
// The server will not wait for its started go routine to finsh but it will send
// the kill signal to them.
func (player *Player) Serve() {
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

	playerGameCh := make(chan *PlayingChData)

	gameState := new(GameState)

	watchGameCh := make(chan *playerWatchGameData)
	watchGames := make(map[int]*JoinWatchChCl)
	sendCh <- readList
	sendSysMess(sendCh, "Welcome back to Battleline!")
Loop:
	for {
		select {
		case <-player.bootCh:
			handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
				receivedInvites, sendInvites, sendCh, wrtDoneCh, false)
			break Loop
		case <-wrtBrookCh:
			log.Printf(log.DebugMsg, "Player %v received write broken.", player.id)
			handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
				receivedInvites, sendInvites, sendCh, wrtDoneCh, false)
			break Loop
		case act, open := <-actChan:
			if open {
				upd := handlePlayerAction(act, readList, sendCh, player, sendInvites,
					receivedInvites, inviteResponseCh, gameState, playerGameCh, watchGames, watchGameCh)

				if upd {
					readList = player.pubList.Read()
					sendCh <- readList
				}
			} else {
				handleCloseDown(player.doneComCh, gameState, watchGames, player.id,
					receivedInvites, sendInvites, sendCh, wrtDoneCh, true)
				break Loop
			}

		case playingChData := <-playerGameCh:
			readList = handleGameReceive(playingChData, sendCh, player.pubList, readList, gameState,
				receivedInvites, sendInvites, player.id)
		case wgd := <-watchGameCh:
			if wgd.joinWatchChCl == nil { //set chan to nil for done
				delete(watchGames, wgd.watchingPlayerID)
				readList = player.pubList.Read()
				sendCh <- readList
			} else {
				watchGames[wgd.watchingPlayerID] = wgd.joinWatchChCl
				sendCh <- wgd.watchingChData
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
					playerGameCh, player.tableStChCl)
			} else {
				if response.PlayingCh != nil {
					close(response.PlayingCh)
				}
			}
		case message := <-player.messCh:
			readList = playerExist(player.pubList, readList, message.SenderID, sendCh)
			_, found := readList[strconv.Itoa(message.SenderID)]
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
	playerGameCh chan<- *PlayingChData,
	watchGames map[int]*JoinWatchChCl,
	watchGameCh chan<- *playerWatchGameData) (isUpd bool) {

	switch act.ActType {
	case ACTIDMess:
		isUpd = actMessage(act, readList, sendCh, player.id, player.name)
	case ACTIDInvite:
		isUpd = actSendInvite(sendInvites, inviteResponseCh, player.doneComCh,
			act, readList, sendCh, player.id, player.name, gameState)
	case ACTIDInvAccept:
		isUpd = actAccInvite(recievedInvites, act, sendCh, playerGameCh, player.doneComCh,
			player.id, gameState)
	case ACTIDInvDecline:
		isUpd = actDeclineInvite(recievedInvites, act, player.id)
	case ACTIDInvRetract:
		actRetractInvite(sendInvites, act)
	case ACTIDMove:
		log.Printf(log.DebugMsg, "handle move for player id: %v Name: %v, Move: %v", player.id, player.name, act)
		actMove(act, gameState, sendCh, player.errCh, player.id)
	case ACTIDSave:
		if gameState.waitingForClient() || gameState.waitingForServer() {
			gameState.closeChannel()
		} else {
			errTxt := "Requesting game save with no game active!"
			sendSysMess(sendCh, errTxt)
			player.errCh <- errors.Wrap(NewPlayerErr(errTxt, player.id), log.ErrNo(19))
		}
	case ACTIDQuit:
		if gameState.waitingForClient() {
			gameState.respCh <- SMQuit
		} else {
			sendSysMess(sendCh, "Quitting game out of turn is not possible.")
			player.errCh <- errors.Wrap(NewPlayerErr("Quitting game out of turn", player.id), log.ErrNo(20))
		}
	case ACTIDWatch:
		isUpd = actWatch(watchGames, act, watchGameCh, player.doneComCh, readList,
			sendCh, player.id)
	case ACTIDWatchStop:
		actWatchStop(watchGames, act, sendCh, player.id)
	case ACTIDList:
		isUpd = true
	default:
		player.errCh <- errors.Wrap(NewPlayerErr("Action do not exist", player.id), log.ErrNo(21))
	}
	return isUpd
}

//GameState keeps track of the current game.
type GameState struct {
	respCh      chan<- int
	lastViewPos *bg.ViewPos
	isClosed    bool //chanel closed
	hasMoved    bool
}

//waitingForServer the player is waiting for the server to return a move.
func (state *GameState) waitingForServer() (res bool) {
	if state.respCh != nil &&
		state.lastViewPos.Winner == bg.NoPlayer &&
		state.lastViewPos.LastMoveType != bg.MoveTypeAll.Pause {
		if len(state.lastViewPos.Moves) > 0 { //players turn
			if state.hasMoved || state.isClosed {
				res = true
			}
		} else { //Waiting for opponent
			res = true
		}
	}
	return res
}

//waitingForClient the client to make a move.
func (state *GameState) waitingForClient() (res bool) {
	if state.respCh != nil && len(state.lastViewPos.Moves) > 0 {
		if !state.hasMoved || !state.isClosed {
			res = true
		}
	}
	return res
}

//removeGame removes game.
func (state *GameState) removeGame() {
	state.respCh = nil
	state.lastViewPos = nil
	state.isClosed = false
}

//addGame adds a game.
func (state *GameState) addGame(data *PlayingChData) {
	state.respCh = data.MoveCh
	state.lastViewPos = data.ViewPos
}

//receiveNewViewPos receives a viewPos.
func (state *GameState) receiveNewViewPos(viewPos *bg.ViewPos) {
	state.lastViewPos = viewPos
	state.hasMoved = false //Player do not receives views when opponent moves
	// they are send directly to client.
}

//sendMove to table.
func (state *GameState) sendMove(moveix int) {
	state.hasMoved = true
	state.respCh <- moveix
}
func (state *GameState) hasGame() bool {
	return state.respCh != nil
}

// closeChannel close the table move channel to signal save game.
func (state *GameState) closeChannel() {
	if state.waitingForClient() || state.waitingForServer() {
		close(state.respCh)
		state.isClosed = true
	}
}

//actRetractInvite retract an invite
func actRetractInvite(sendInvites map[int]*Invite, act *Action) {
	invite, found := sendInvites[act.ID]
	if found {
		close(invite.RetractCh)
		delete(sendInvites, act.ID)
	}
}

//handleInviteResponse handle a invite response.
//# sendInvites
func handleInviteResponse(
	response *InviteResponse,
	invite *Invite,
	sendInvites map[int]*Invite,
	sendCh chan<- interface{},
	playerID int,
	playerDoneComCh chan struct{},
	playerGameCh chan<- *PlayingChData,
	startGameChCl *StartGameChCl) {

	_, found := sendInvites[response.Responder] //if not found we rejected but he did not get the msg yet
	if found {
		delete(sendInvites, response.Responder)
		if response.PlayingCh == nil {
			invite.IsRejected = true
			sendCh <- invite
		} else {
			playingRecCh := make(chan *PlayingChData, 1)
			startData := new(StartGameChData)
			startData.PlayerIds = [2]int{playerID, response.Responder}
			startData.PlayerChs = [2]chan<- *PlayingChData{playingRecCh, response.PlayingCh}
			select {
			case startGameChCl.Channel <- startData:
				go gameListen(playerGameCh, playerDoneComCh, playingRecCh, sendCh, invite)
			case <-startGameChCl.Close:
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
func handleGameReceive(
	playingChData *PlayingChData,
	sendCh chan<- interface{},
	pubList *PubList,
	readList map[string]*PubData,
	gameState *GameState,
	receivedInvites map[int]*Invite,
	sendInvites map[int]*Invite, playerID int) (updReadList map[string]*PubData) {

	updReadList = readList
	if playingChData == nil {
		gameState.removeGame()
		updReadList = pubList.Read()
		sendCh <- updReadList
	} else {
		if !gameState.hasGame() { //Init data
			gameState.addGame(playingChData)
			clearInvites(receivedInvites, sendInvites, playerID)
			sendCh <- ClearInvites("All invites was clear as game starts.")
			updReadList = pubList.Read()
			sendCh <- updReadList
			sendCh <- playingChData
		} else {
			if len(playingChData.ViewPos.Moves) > 0 {
				gameState.receiveNewViewPos(playingChData.ViewPos)
			}
			sendCh <- playingChData
		}
	}
	return updReadList
}

// handleCloseDown close down the player.
//# receivedInvites
//# sendInvites
//# gameState
func handleCloseDown(playerDoneComCh chan struct{}, gameState *GameState,
	watchGames map[int]*JoinWatchChCl, playerID int, receivedInvites map[int]*Invite,
	sendInvites map[int]*Invite, sendCh chan<- interface{}, wrtDoneCh chan struct{}, isPlayerCloseConn bool) {
	log.Printf(log.DebugMsg, "Closing player id: %v done comunication channel.", playerID)
	close(playerDoneComCh)
	if gameState != nil {
		log.Printf(log.DebugMsg, "Player id: %v closing game.", playerID)
		gameState.closeChannel()
	}
	if len(watchGames) > 0 {
		for watchingID, ch := range watchGames {
			log.Printf(log.DebugMsg, "Player id: %v stops watching player is: %v", playerID, watchingID)
			stopWatch(ch, playerID)
		}
	}
	clearInvites(receivedInvites, sendInvites, playerID)
	log.Printf(log.DebugMsg, "Player id: %v sending close connetion with is player %v", playerID, isPlayerCloseConn)
	sendCh <- CloseCon{isPlayerCloseConn, "Server close the connection"}
	log.Printf(log.DebugMsg, "Player id: %v Waiting for write done.", playerID)
	<-wrtDoneCh
	log.Printf(log.DebugMsg, "Player id: %v received write done.", playerID)
}

//actWatch handle client action watch a game.
//# watchGames
func actWatch(
	watchGames map[int]*JoinWatchChCl,
	act *Action,
	playerWatchGameCh chan<- *playerWatchGameData,
	playerDoneComCh chan struct{},
	readList map[string]*PubData,
	sendCh chan<- interface{},
	playerID int) (isUpdList bool) {
	_, found := watchGames[act.ID] // This test only works for game started, the bench server should reject
	// any repeat request to start game.
	if found {
		txt := fmt.Sprintf("Start watching id: %v failed as you are already watching", act.ID)
		sendSysMess(sendCh, txt)
	} else {
		joinWatchChCl := findJoinWatchCh(readList, act.ID)
		if joinWatchChCl != nil {
			watch := new(JoinWatchChData)
			watch.ID = playerID
			watchingCh := make(chan *WatchingChData, 1)
			watch.SendCh = watchingCh
			select {
			case joinWatchChCl.Channel <- watch:
				go watchGameListen(joinWatchChCl, watchingCh, playerWatchGameCh, playerDoneComCh, sendCh, act.ID, playerID)
			case <-joinWatchChCl.Close:
				txt := fmt.Sprintf("Player id: %v do not have a active game", act.ID)
				sendSysMess(sendCh, txt)
				isUpdList = true
			}
		} else {
			txt := fmt.Sprintf("Player id: %v do not have a active game", act.ID)
			sendSysMess(sendCh, txt)
			isUpdList = true
		}
	}
	return isUpdList
}
func findJoinWatchCh(pubList map[string]*PubData, watchID int) (ch *JoinWatchChCl) {
	pubData, found := pubList[strconv.Itoa(watchID)]
	if found {
		ch = pubData.JoinWatchChCl
	}
	return ch
}

// watchGameListen handle game information from a watch game.
// If playerDoneComCh is closed it will close the game connection in init stage, and
// in normal state it will stop resending but keep reading until connection is closed.
func watchGameListen(
	joinWatchChCl *JoinWatchChCl,
	watchingCh <-chan *WatchingChData,
	playerWatchGameCh chan<- *playerWatchGameData,
	playerDoneComCh chan struct{},
	sendCh chan<- interface{},
	watchID, playerID int) {

	initWatchingChData, initOpen := <-watchingCh
	if initOpen {
		initWatchingChData.WatchingID = watchID
		isDoneCom := sendToPlayerStartWatch(watchID, initWatchingChData, joinWatchChCl, playerWatchGameCh, playerDoneComCh)
		if isDoneCom {
			stopWatch(joinWatchChCl, playerID)
		}
		writeStop := false
	Loop:
		for {
			watchingData, open := <-watchingCh
			if !writeStop {
				if open {
					watchingData.WatchingID = watchID
					select {
					case <-playerDoneComCh:
						writeStop = true
					default:
						select {
						case sendCh <- watchingData:
						case <-playerDoneComCh:
							writeStop = true
						}
					}
				} else { //Game is finished.
					sendToPlayerStopWatch(watchID, playerWatchGameCh, playerDoneComCh)
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
		sendSysMessGo(sendCh, playerDoneComCh, txt)
	}

}
func sendToPlayerStartWatch(
	watchID int,
	initWatchingChData *WatchingChData,
	joinWatchChCl *JoinWatchChCl,
	playerWatchGameCh chan<- *playerWatchGameData,
	playerDoneComCh chan struct{}) (isDone bool) {

	pwgData := new(playerWatchGameData)
	pwgData.watchingPlayerID = watchID
	pwgData.watchingChData = initWatchingChData
	pwgData.joinWatchChCl = joinWatchChCl
	return sendToPlayerWatchData(playerWatchGameCh, playerDoneComCh, pwgData)
}
func sendToPlayerStopWatch(
	watchID int,
	playerWatchGameCh chan<- *playerWatchGameData,
	playerDoneComCh chan struct{}) {

	pwgData := new(playerWatchGameData)
	pwgData.watchingPlayerID = watchID
	_ = sendToPlayerWatchData(playerWatchGameCh, playerDoneComCh, pwgData)
}
func sendToPlayerWatchData(playerWatchGameCh chan<- *playerWatchGameData,
	playerDoneComCh chan struct{},
	data *playerWatchGameData) (isDone bool) {
	select {
	case playerWatchGameCh <- data:
	case <-playerDoneComCh:
		isDone = true
	}
	return isDone
}

// actWatchStop stop watching a game.
func actWatchStop(watchGames map[int]*JoinWatchChCl, act *Action,
	sendCh chan<- interface{}, playerID int) {
	joinWatchChCl, found := watchGames[act.ID]
	if found {
		stopWatch(joinWatchChCl, playerID)
		delete(watchGames, act.ID) //TODO MAYBE send stop to the player
	} else {
		txt := fmt.Sprintf("Stop watching player %v failed", act.ID)
		sendSysMess(sendCh, txt)
	}
}

// stopWatch sends the stop watch signal if possibel.
func stopWatch(joinWatchChCl *JoinWatchChCl, playerID int) {
	watchData := new(JoinWatchChData)
	watchData.ID = playerID
	watchData.SendCh = nil //stop
	select {
	case joinWatchChCl.Channel <- watchData:
	case <-joinWatchChCl.Close:
	}
}

// actMove makes a game move if the move is valid.
func actMove(act *Action, gameState *GameState, sendCh chan<- interface{}, errCh chan<- error, id int) {
	if gameState.waitingForClient() {
		lastPos := gameState.lastViewPos
		if act.Moveix >= 0 && act.Moveix < len(lastPos.Moves) {
			log.Print(log.DebugMsg, "Sending move to table")
			gameState.sendMove(act.Moveix)
		} else {
			txt := "Illegal Move"
			sendSysMess(sendCh, txt)
			errCh <- errors.Wrap(NewPlayerErr("Illegal move", id), log.ErrNo(22))
			gameState.sendMove(SMQuit)
		}
	} else {
		txt := "Is not your time to move.!"
		sendSysMess(sendCh, txt)
		errCh <- errors.Wrap(NewPlayerErr("Move out of turn", id), log.ErrNo(23))
	}
}

// actAccInvite accept a invite.
func actAccInvite(recInvites map[int]*Invite, act *Action, sendCh chan<- interface{},
	playerGameCh chan<- *PlayingChData, playerDoneComCh chan struct{}, id int, gameState *GameState) (isUpd bool) {

	invite, found := recInvites[act.ID]
	if found {
		if !gameState.hasGame() {
			resp := new(InviteResponse)
			resp.Responder = id
			playingRecCh := make(chan *PlayingChData, 1)
			resp.PlayingCh = playingRecCh
			select {
			case <-invite.RetractCh:
				txt := fmt.Sprintf("Accepting invite from %v failed invitation retracted", invite.InvitorName)
				sendSysMess(sendCh, txt)
			default:
				select {
				case invite.ResponseCh <- resp:
					go gameListen(playerGameCh, playerDoneComCh, playingRecCh, sendCh, invite)
				case <-invite.DoneComCh:
					txt := fmt.Sprintf("Accepting invite from %v failed player done", invite.InvitorName)
					sendSysMess(sendCh, txt)
					isUpd = true
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
	return isUpd
}

// gameListen listen for game moves.
// If playerDoneComCh is closed and init state the game response channel is closed else
// the listener keep listening until the channel is close but it do not resend the moves.
//
func gameListen(playerGameCh chan<- *PlayingChData, playerDoneComCh chan struct{}, playingRecCh <-chan *PlayingChData,
	sendCh chan<- interface{}, invite *Invite) {
	initPlayingChData, initOpen := <-playingRecCh
	if initOpen {
		isDone := sendToPlayerGameCh(playerGameCh, playerDoneComCh, initPlayingChData)
		if isDone {
			close(initPlayingChData.MoveCh)
		}
		var isWriteStop bool
	Loop:
		for {
			playingChData, open := <-playingRecCh
			if !isWriteStop {
				if open {
					isWriteStop = sendToPlayerGameCh(playerGameCh, playerDoneComCh, playingChData)
				} else {
					sendToPlayerGameCh(playerGameCh, playerDoneComCh, nil)
					break Loop
				}
			} else {
				if !open {
					break Loop
				}
			}
		}
	} else {
		invite.IsRejected = true
		sendCh <- invite
	}
}

//Use to send pos view directly when the player do not need to take action, but the the views got out of order.
/*func sendToClientPlayingData(
	sendCh chan<- interface{},
	playerDoneComCh chan struct{},
	playingChData *PlayingChData) (isDone bool) {
	select {
	case <-playerDoneComCh: //sendCh is buffered and closed at a later stage
		isDone = true
	default:
		select {
		case sendCh <- playingChData:
		case <-playerDoneComCh:
			isDone = true
		}
	}
	return isDone
}*/
func sendToPlayerGameCh(
	playerGameCh chan<- *PlayingChData,
	playerDoneComCh chan struct{},
	playingChData *PlayingChData) (isDone bool) {
	select {
	case playerGameCh <- playingChData:
	case <-playerDoneComCh:
		isDone = true
	}
	return isDone

}

//actSendInvite send a invite.
func actSendInvite(invites map[int]*Invite, respCh chan<- *InviteResponse, playerDoneComCh chan struct{},
	act *Action, readList map[string]*PubData, sendCh chan<- interface{}, id int, name string,
	gameState *GameState) (isUpd bool) {
	invite := new(Invite)
	invite.InvitorID = id
	invite.InvitorName = name
	p, found := readList[strconv.Itoa(act.ID)]
	if found {
		invite.ReceiverID = p.ID
		invite.ReceiverName = p.Name
		if !gameState.hasGame() {
			invite.ResponseCh = respCh
			invite.RetractCh = make(chan struct{})
			invite.DoneComCh = playerDoneComCh
			select {
			case p.InviteCh <- invite:
				invites[p.ID] = invite
			case <-p.DoneComCh:
				invite.IsRejected = true
				sendCh <- invite
				m := fmt.Sprintf("Invite to %v failed player done", p.Name)
				sendSysMess(sendCh, m)
				isUpd = true
			}
		} else {
			invite.IsRejected = true
			sendCh <- invite
			m := fmt.Sprintf("Invite to %v cannot be extended while playing", p.Name)
			sendSysMess(sendCh, m)
		}
	} else {
		invite.IsRejected = true
		sendCh <- invite
		m := fmt.Sprintf("Invite to Id: %v failed could not find player", act.ID)
		sendSysMess(sendCh, m)
		isUpd = true
	}
	return isUpd
}

// actMessage send a message.
func actMessage(act *Action, readList map[string]*PubData, sendCh chan<- interface{},
	id int, name string) (isUpd bool) {
	p, found := readList[strconv.Itoa(act.ID)]
	if found {
		mess := new(MesData)
		mess.SenderID = id
		mess.SenderName = name
		mess.Message = act.Mess
		select {
		case p.MessageCh <- mess:
		case <-p.DoneComCh:
			m := fmt.Sprintf("Message to Id: %v failed", act.ID)
			sendSysMess(sendCh, m)
			isUpd = true
		}
	} else {
		m := fmt.Sprintf("Message to Id: %v failed", act.ID)
		sendSysMess(sendCh, m)
		isUpd = true
	}
	return isUpd
}

// actDeclineInvite decline a invite.
func actDeclineInvite(receivedInvites map[int]*Invite, act *Action, playerID int) (isUpd bool) {
	invite, found := receivedInvites[act.ID]
	if found {
		isUpd = declineInvite(invite, playerID)
		delete(receivedInvites, invite.InvitorID)
	}
	return isUpd
}

// declineInvite send the decline signal to opponent.
func declineInvite(invite *Invite, playerID int) (isUpd bool) {
	resp := new(InviteResponse)
	resp.Responder = playerID
	resp.PlayingCh = nil
	select {
	case <-invite.RetractCh:
	default:
		select {
		case invite.ResponseCh <- resp:
		case <-invite.DoneComCh:
			isUpd = true
		}
	}
	return isUpd
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
	mess.SenderID = -1
	mess.SenderName = "System"
	mess.Message = txt
	ch <- mess
}

// sendSysMessGo send a system message if possible.
func sendSysMessGo(sendCh chan<- interface{}, playerDoneComCh chan struct{}, txt string) {
	mess := new(MesData)
	mess.SenderID = -1
	mess.SenderName = "System"
	mess.Message = txt
	select {
	case sendCh <- mess:
	case <-playerDoneComCh:
	}
}

// clearInvites clear invites. Retract all send and cancel all recieved.
func clearInvites(receivedInvites map[int]*Invite, sendInvites map[int]*Invite,
	playerID int) {
	if len(receivedInvites) != 0 {
		resp := new(InviteResponse)
		resp.Responder = playerID
		resp.PlayingCh = nil
		for _, invite := range receivedInvites {
			select {
			case invite.ResponseCh <- resp:
			case <-invite.RetractCh:
			}
		}
		for id := range receivedInvites {
			delete(receivedInvites, id)
		}
	}
	if len(sendInvites) != 0 {
		for _, invite := range sendInvites {
			close(invite.RetractCh)
		}
		for id := range sendInvites {
			delete(receivedInvites, id)
		}
	}
}

// CloseCon close connection signal to netWrite.
type CloseCon struct {
	isPlayer bool
	Reason   string
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
			if closeCon.isPlayer {
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
				log.Print(log.DebugMsg, "closing write broken channel")
				close(brokenConn)
			}
		}
		if stop {
			break Loop
		}
	}
	log.Print(log.DebugMsg, "Closing write done")
	close(doneCh)
}

//netWriteAddJsonType adds the json type to interface values.
func netWriteAddJSONType(data interface{}) (jdata *JsonData) {
	jdata = new(JsonData)
	jdata.Data = data
	switch data.(type) {
	case map[string]*PubData:
		jdata.JsonType = JTList
	case *Invite:
		jdata.JsonType = JTInvite
	case *MesData:
		jdata.JsonType = JTMess
	case *PlayingChData:
		jdata.JsonType = JTPlaying
	case *WatchingChData:
		jdata.JsonType = JTWatching
	case CloseCon:
		jdata.JsonType = JTCloseCon
	case ClearInvites:
		jdata.JsonType = JTClearInvites
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
func netRead(ws *websocket.Conn, accCh chan<- *Action, playerDoneComCh chan struct{}, errCh chan<- error) {

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
			case <-playerDoneComCh:
			default:
				errCh <- errors.Wrap(err, log.ErrNo(25)+"Websocket receive")
			}
			break Loop //maybe to harsh
		} else {
			select {
			case accCh <- &act:
			case <-playerDoneComCh:
				break Loop
			}
		}
	}
	log.Print(log.DebugMsg, "Closing action channel")
	close(accCh)
}

// Action the client action.
type Action struct {
	ActType int
	ID      int
	Moveix  int
	Mess    string
}

// NewAction creates a new action.
func NewAction(actType int) (a *Action) {
	a = new(Action)
	a.ActType = actType
	a.Moveix = SMNone
	return a
}

// playerWatchGameData the data send on playerWatchGameCh
// to inform player server about starting and stoping to watch a game.
type playerWatchGameData struct {
	joinWatchChCl    *JoinWatchChCl
	watchingChData   *WatchingChData
	watchingPlayerID int
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
