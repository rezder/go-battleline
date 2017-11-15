package prob

import (
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"os"
	"testing"
)

func TestMoves(t *testing.T) {
	filePath := "_test/testdb.db2"
	bdb := testOpenDb(filePath, t)
	gameix := 0
	log.InitLog(log.Min)
	testGameix := -1 //-1
	_, _, err := bdb.Search(func(hist *game.Hist) bool {
		if testGameix == -1 || gameix == testGameix {
			g := game.NewGame()
			g.LoadHist(hist)
			winner, hasNext := g.ScrollForward() //initMove
			for ; hasNext; winner, hasNext = g.ScrollForward() {
				viewPos := game.NewViewPos(g.Pos, game.ViewAll.Players[0], winner)
				if len(viewPos.Moves) == 0 {
					viewPos = game.NewViewPos(g.Pos, game.ViewAll.Players[1], winner)
				}
				if len(viewPos.Moves) != 0 {
					moveix := testMoveix(viewPos)
					newMove := viewPos.Moves[moveix]
					oldMove := g.Hist.Moves[viewPos.LastMoveIx+1]
					isDeck := newMove.MoveType == game.MoveTypeAll.Deck ||
						newMove.MoveType.IsScout()
					if isDeck && oldMove.MoveType == newMove.MoveType {
						mix := 0
						if newMove.MoveType == game.MoveTypeAll.Scout1 {
							mix = 1
						}
						oldCard := card.Card(oldMove.Moves[mix].Index)
						newCard := card.Card(newMove.Moves[mix].Index)
						if oldCard.IsTac() && newCard != card.BACKTac ||
							oldCard.IsTroop() && newCard != card.BACKTroop {
							t.Errorf("Game ts: %v, Gameix %v,Moveix %v, Old deck move: %v deviates from new move: %v", g.Hist.Time, gameix, g.Pos.LastMoveIx, oldMove, newMove)
						}
					} else if !newMove.IsEqual(oldMove) {
						if strenght(newMove) > 0 && strenght(newMove) == strenght(oldMove) {
						} else {
							t.Errorf("Game ts:%v,Game Index: %v,Moveix: %v, Old move: %v deviates from new move: %v", g.Hist.Time, gameix, g.Pos.LastMoveIx, oldMove, newMove)
						}
					}
				}
			}
		}
		gameix++
		return false
	}, nil)
	if err != nil {
		t.Errorf("Failed searching file %v with error: %v", filePath, err)
	}
	err = bdb.Close()
	if err != nil {
		t.Errorf("Closing failed file %v with error: %v", filePath, err)
	}
}
func strenght(move *game.Move) int {
	if len(move.Moves) > 0 && card.Card(move.Moves[0].Index).IsTroop() {
		return card.Troop(move.Moves[0].Index).Strenght()
	}
	return 0
}
func testMoveix(viewPos *game.ViewPos) (moveix int) {
	moveix = -1
	firstMove := viewPos.Moves[0]
	switch firstMove.MoveType {
	case game.MoveTypeAll.Cone:
		moveix = MoveClaim(viewPos)
	case game.MoveTypeAll.Scout2:
		fallthrough
	case game.MoveTypeAll.Scout3:
		fallthrough
	case game.MoveTypeAll.Deck:
		moveix = MoveDeck(viewPos)
	case game.MoveTypeAll.ScoutReturn:
		moveix = MoveScoutReturn(viewPos)
	default:
		moveix = MoveHand(viewPos)
	}
	if log.Level() == log.Debug {
		fmt.Printf("#%v,%v\n", viewPos.LastMoveIx, moveix)
	}
	return moveix
}
func testOpenDb(filePath string, t *testing.T) (bdb *dbhist.Db) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		t.Fatalf("Open database file: %v failed: %v", filePath, err)
	}

	bdb = dbhist.New(dbhist.KeyPlayersTime, db, 25)
	err = bdb.Init()
	if err != nil {
		bdb.Close()
		t.Fatalf("Init database failed: %v", err)
	}
	return bdb
}
func BenchmarkMove(b *testing.B) {
	filePath := "_test/speedgame.batt2"
	g, err := testLoadGobGame(filePath)
	if err != nil {
		b.Fatalf("Load file %v failed with %v", filePath, err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		winner, hasNext := g.ScrollForward() //initMove
		for ; hasNext; winner, hasNext = g.ScrollForward() {
			viewPos := game.NewViewPos(g.Pos, game.ViewAll.Players[0], winner)
			if len(viewPos.Moves) == 0 {
				viewPos = game.NewViewPos(g.Pos, game.ViewAll.Players[1], winner)
			}
			if len(viewPos.Moves) != 0 {
				testMoveix(viewPos)
			}
		}
	}

}
func testLoadGobGame(filePath string) (g *game.Game, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return g, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var hist game.Hist
	deCoder := gob.NewDecoder(file)
	err = deCoder.Decode(&hist)
	if err != nil {
		return g, err
	}
	g = game.NewGame()
	g.LoadHist(&hist)
	return g, err
}
func TestMudReduce(t *testing.T) {
	cardPos := [71]pos.Card{0, 21, 22, 22, 0, 0, 0, 1, 1, 1, 1, 22, 21, 21, 22, 22, 21, 11, 11, 11, 11, 0, 0, 0, 21, 22, 0, 0, 0, 21, 0, 13, 0, 15, 15, 0, 0, 0, 0, 0, 15, 13, 0, 0, 5, 0, 0, 0, 21, 12, 0, 0, 0, 0, 4, 4, 4, 0, 14, 22, 0, 23, 23, 23, 23, 10, 23, 23, 23, 23, 23}
	posCards := game.NewPosCards(cardPos)
	posCards = mudTrim(posCards, cardPos[card.TCMud])
	dish0 := pos.CardAll.Players[0].Dish
	dish1 := pos.CardAll.Players[1].Dish
	t.Log(posCards[dish0], posCards[dish1])
	if len(posCards[dish0]) != 2 && posCards[dish0][1] != 7 {
		t.Error("Mud trim failed expected card 7 to be dished")
	}
	if len(posCards[dish1]) != 1 && posCards[dish1][0] != 17 {
		t.Error("Mud trim failed expected card 17 to be dished")
	}
}
