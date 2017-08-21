package games

import (
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-error/log"
	"math/rand"
)

const (
	//SMQuit the special move quit
	SMQuit = -1
	//SMNone used to track fails in json or other file serialization
	//It is not a move.
	SMNone = -2
)

//tableServe runs a table with a game.
//If resumeGame is nil a new game is started.
func tableServe(
	ids [2]int,
	playerChs [2]chan<- *PlayingChData,
	joinWatchChCl *JoinWatchChCl,
	resumeGame *bg.Game,
	finishCh chan *FinishTableData,
	errCh chan<- error) {

	var moveixChs [2]chan int
	moveixChs[0] = make(chan int)
	moveixChs[1] = make(chan int)
	benchCh := make(chan *WatchingChData, 1)
	go benchServe(joinWatchChCl, benchCh)
	game := resumeGame
	if game == nil {
		game = bg.NewGame()
		dealer := rand.Intn(2)
		game.Start(ids, dealer)
	}
	playingChDatas, watchingChData := initChData(game.Pos, ids, moveixChs)
	playerChs[0] <- playingChDatas[0]
	playerChs[1] <- playingChDatas[1]
	benchCh <- watchingChData
	var moveix int
	var isOpen bool
	var mover int
	moves := game.Pos.CalcMoves()
	winner := bg.NoPlayer

	for winner != bg.NoPlayer {
		var failedClaimedExs [9][]card.Move
		mover = moves[0].Mover
		log.Printf(log.DebugMsg, "Waiting for mover ix: %v id: %v", mover, ids[mover])
		moveix, isOpen = <-moveixChs[mover]
		log.Printf(log.DebugMsg, "Recived move ix:%v from mover ix: %v id:%v", moveix, mover, ids[mover])
		if !isOpen {
			winner = game.Pause(moves)
		} else if moveix == SMQuit {
			winner = game.GiveUp(moves)
		} else {
			winner, failedClaimedExs = game.Move(moves[moveix])
		}
		playingChDatas, watchingChData = createChData(game.Pos, ids, winner, failedClaimedExs)
		log.Printf(log.DebugMsg, "Sending view to playerid: %v\n%v\n", ids[0], playingChDatas[0].ViewPos)
		playerChs[0] <- playingChDatas[0]
		log.Printf(log.DebugMsg, "Sending view to playerid: %v\n%v\n", ids[1], playingChDatas[1].ViewPos)
		playerChs[1] <- playingChDatas[1]
		benchCh <- watchingChData
		if !isOpen {
			break
		}
	}
	close(playerChs[0])
	close(playerChs[1])
	close(benchCh)
	finData := new(FinishTableData)
	finData.ids = ids
	finData.game = game
	finishCh <- finData // may be buffered tables will close bench before closing tables
}

//createChData
func createChData(
	pos *bg.Pos,
	ids [2]int,
	winner int,
	failedClaimedExs [9][]card.Move) (playingChDatas [2]*PlayingChData, watchingChData *WatchingChData) {
	for i := range ids {
		playingChDatas[i] = &PlayingChData{
			ViewPos:          bg.NewViewPos(pos, bg.ViewAll.Players[i], winner),
			PlayingIDs:       ids,
			FailedClaimedExs: failedClaimedExs,
		}
	}

	watchingChData = &WatchingChData{
		ViewPos:    bg.NewViewPos(pos, bg.ViewAll.Spectator, winner),
		PlayingIDs: ids,
	}

	return playingChDatas, watchingChData
}

//initChData the first views
func initChData(
	pos *bg.Pos,
	ids [2]int,
	moveChs [2]chan int) (playingChDatas [2]*PlayingChData, watchingChData *WatchingChData) {

	var failedClaimedExs [9][]card.Move
	playingChDatas, watchingChData = createChData(pos, ids, bg.NoPlayer, failedClaimedExs)
	for i, data := range playingChDatas {
		data.MoveCh = moveChs[i]
	}

	return playingChDatas, watchingChData
}
