package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	botpos "github.com/rezder/go-battleline/battbot/gamepos"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	pub "github.com/rezder/go-battleline/battserver/publist"
	tab "github.com/rezder/go-battleline/battserver/tables"
	"github.com/rezder/go-error/log"
	"testing"
)

func TestBot(t *testing.T) {
	log.InitLog(log.Min)
	testGameix := -1 //Set if debugging a game
	src := "_test/bdb1.db"
	db, err := bolt.Open(src, 0600, nil)
	if err != nil {
		t.Fatalf("Open database: %v failed with %v", src, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			t.Fatalf("Close database %v failed with %v", src, cerr)
		}
	}()

	gameDb := battdb.New(battdb.KeyPlayersTime, db, 1000)
	err = gameDb.Init()
	if err != nil {
		t.Fatalf("Init database %v failed with %v", src, err)
	}
	var nextKey []byte
	var games []*bat.Game
	for {
		games, nextKey, err = gameDb.SearchLoop(nil, nextKey)
		if err != nil {
			t.Fatalf("Search game database: %v failed with %v", src, err)
		}
		t.Logf("Loaded %v games", len(games))
		for i, game := range games {
			if testGameix == -1 || i == testGameix { //TODO  remove
				testGame(game, i, t)
			}
		}
		if len(nextKey) == 0 {
			break
		}
	}
}

func testGame(game *bat.Game, gameix int, t *testing.T) {
	var simMoveixs [2]int
	var simMove bat.Move
	var botPoss [2]*botpos.Pos
	game.GameMoveLoop(func(
		gameMoveix int,
		prePos, postPos *bat.GamePos,
		moveCardix, dealtix, moveix int,
		move bat.Move,
		isGiveUp, isPass bool,
		claimFailExs [9][]int,
		mudDishixs []int,
	) {
		if gameMoveix == 0 {
			for i := 0; i < 2; i++ {
				botPoss[i] = botpos.New()
				initMove := new(pub.MoveView)
				initMove.Turn = pub.NewTurn(&prePos.Turn, i)
				h := make([]int, len(prePos.Hands[i].Troops))
				copy(h, prePos.Hands[i].Troops)
				initMove.Move = *tab.NewMoveInit(h)
				_ = botPoss[i].UpdMove(initMove)
				if prePos.Player == i {
					simMoveixs = botPoss[i].MakeMove()
					simMove = testGetMove(simMoveixs, prePos.MovesHand, prePos.Moves)
				}
			}
		}
		testCompareMove(simMoveixs, simMove, moveix, moveCardix, gameMoveix, gameix, move, t)
		if isGiveUp {
			move = *tab.NewMoveQuit()
		} else if isPass {
			move = *tab.NewMovePass()
		} else {
			switch battMove := move.(type) {
			case bat.MoveDeserter:
				move = *tab.NewMoveDeserterView(&battMove, mudDishixs)
			case bat.MoveRedeploy:
				move = *tab.NewMoveRedeployView(&battMove, mudDishixs)
			case bat.MoveScoutReturn:
				move = *tab.NewMoveScoutReturnView(battMove)
			case bat.MoveClaim:
				move = *tab.NewMoveClaimView(battMove, postPos, claimFailExs)
			}
		}
		mover := prePos.Player
		for i, botPos := range botPoss {
			moveView := new(pub.MoveView)
			moveView.Mover = i == mover
			moveView.Move = move
			moveView.MoveCardix = moveCardix
			moveView.Turn = pub.NewTurn(&postPos.Turn, i)
			if dealtix != 0 {
				if mover == i {
					moveView.DeltCardix = dealtix
				}
			}
			if !botPos.UpdMove(moveView) && botPos.IsBotTurn() {
				simMoveixs = botPos.MakeMove()
				simMove = testGetMove(simMoveixs, postPos.MovesHand, postPos.Moves)
			}
		}
		if log.Level() == log.Debug {
			fmt.Printf("#%v,%v\n", gameMoveix, simMoveixs)
		}
	})
}
func testGetMove(moveixs [2]int, handMoves map[int][]bat.Move, moves []bat.Move) (move bat.Move) {
	switch {
	case moveixs[1] == bat.SMGiveUp:
	case moveixs[1] == bat.SMPass:
	case moveixs[1] >= 0:
		if moveixs[0] > 0 {
			move = handMoves[moveixs[0]][moveixs[1]]
		} else {
			move = moves[moveixs[1]]
		}
	default:
		panic("This should not happen. Move data is corrupt")
	}
	return move
}
func testCompareMove(
	simMoveixs [2]int,
	simMove bat.Move,
	moveix, moveCardix, gameMoveix, gameix int,
	move bat.Move,
	t *testing.T,
) {
	if simMoveixs[0] != moveCardix || simMoveixs[1] != moveix {
		simFlagMove, isSimOk := simMove.(bat.MoveCardFlag)
		moveFlag, isOk := move.(bat.MoveCardFlag)
		isLog := true
		if isOk && isSimOk {
			if moveFlag.Flagix == simFlagMove.Flagix {
				simTroop, isSimOk := cards.DrTroop(simMoveixs[0])
				troop, isOk := cards.DrTroop(moveCardix)
				if isSimOk && isOk {
					if simTroop.Value() == troop.Value() {
						isLog = false
					}
				}
			}
		}
		if isLog {
			t.Errorf("Game %v deviates on move %v: Oldmove: %v,OldCardix: %v NewMove: %v,NewCardix: %v", gameix, gameMoveix, move, moveCardix, simMove, simMoveixs[0])
		}
	}
}
