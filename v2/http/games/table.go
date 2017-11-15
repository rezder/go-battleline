package games

import (
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	dpos "github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"math/rand"
	"time"
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
	finishCh chan *bg.Game,
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
	playingChDatas, watchingChData, moves := initChData(game.Pos, game.Hist.PlayerIDs, game.Hist.Time, moveixChs)
	playerChs[0] <- playingChDatas[0]
	playerChs[1] <- playingChDatas[1]
	benchCh <- watchingChData
	var moveix int
	var isOpen bool
	var mover int
	winner := dpos.NoPlayer

	for winner == dpos.NoPlayer {
		var failedClaimedExs [9][]card.Card
		mover = moves[0].Mover
		log.Printf(log.DebugMsg, "Waiting for mover ix: %v id: %v", mover, ids[mover])
		moveix, isOpen = <-moveixChs[mover]
		if isOpen {
			log.Printf(log.DebugMsg, "Recived move ix:%v from mover ix: %v id:%v", moveix, mover, ids[mover])
		} else {
			log.Print(log.DebugMsg, "Recived move channel closed")
		}
		if !isOpen {
			winner = game.Pause(moves)
		} else if moveix == SMQuit {
			winner = game.GiveUp(moves)
		} else {
			winner, failedClaimedExs = game.Move(moves[moveix])
		}
		playingChDatas, watchingChData, moves = createChData(game.Pos, game.Hist.PlayerIDs, game.Hist.Time, winner, failedClaimedExs)
		log.Printf(log.DebugMsg, "Sending view to playerid: %v\n%v\n%v", ids[0], playingChDatas[0].ViewPos, failedClaimedExs)
		playerChs[0] <- playingChDatas[0]
		log.Printf(log.DebugMsg, "Sending view to playerid: %v\n%v\n%v", ids[1], playingChDatas[1].ViewPos, failedClaimedExs)
		playerChs[1] <- playingChDatas[1]
		benchCh <- watchingChData
		if !isOpen {
			break
		}
	}
	close(playerChs[0])
	close(playerChs[1])
	close(benchCh)
	finishCh <- game // may be buffered tables will close bench before closing tables
}

//createChData
func createChData(
	pos *bg.Pos,
	ids [2]int,
	gameTs time.Time,
	winner int,
	failedClaimedExs [9][]card.Card) (playingChDatas [2]*PlayingChData, watchingChData *WatchingChData, moves []*bg.Move) {
	for i := range ids {
		playingChDatas[i] = &PlayingChData{
			ViewPos:          bg.NewViewPos(pos, bg.ViewAll.Players[i], winner),
			PlayingIDs:       ids,
			GameTs:           gameTs,
			FailedClaimedExs: failedClaimedExs,
		}
		if len(playingChDatas[i].ViewPos.Moves) > 0 {
			moves = playingChDatas[i].ViewPos.Moves
		}
	}

	watchingChData = &WatchingChData{
		ViewPos:    bg.NewViewPos(pos, bg.ViewAll.Spectator, winner),
		PlayingIDs: ids,
		GameTs:     gameTs,
	}

	return playingChDatas, watchingChData, moves
}

//initChData the first views
func initChData(
	pos *bg.Pos,
	ids [2]int,
	gameTs time.Time,
	moveChs [2]chan int) (playingChDatas [2]*PlayingChData, watchingChData *WatchingChData, moves []*bg.Move) {

	var failedClaimedExs [9][]card.Card
	playingChDatas, watchingChData, moves = createChData(pos, ids, gameTs, dpos.NoPlayer, failedClaimedExs)
	for i, data := range playingChDatas {
		data.MoveCh = moveChs[i]
	}

	return playingChDatas, watchingChData, moves
}
