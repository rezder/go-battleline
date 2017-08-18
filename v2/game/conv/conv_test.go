package conv

import (
	"encoding/gob"
	bold "github.com/rezder/go-battleline/battleline"
	boldFlag "github.com/rezder/go-battleline/battleline/flag"
	"github.com/rezder/go-battleline/battserver/tables"
	"github.com/rezder/go-battleline/v2/game"
	dPos "github.com/rezder/go-battleline/v2/game/pos"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	dirTestGames = "test"
)

func TestGameFile(t *testing.T) {
	suffix := ".batt2"
	files, err := ioutil.ReadDir(dirTestGames)
	if err != nil {
		t.Errorf("Error reding directory: %v Err: %v", dirTestGames, err)
		return
	}
	for _, fileInfo := range files {
		if strings.HasSuffix(fileInfo.Name(), ".gob") {
			src := filepath.Join(dirTestGames, fileInfo.Name())
			err = GameFile(src, src+suffix)
			if err != nil {
				t.Errorf("Converting file: %v failed, Error: %v", fileInfo.Name(), err)
			}
		}
	}
	files, err = ioutil.ReadDir(dirTestGames)
	if err != nil {
		t.Errorf("Reding  directory: %v failed, Error: %v", dirTestGames, err)
		return
	}

	for _, fileInfo := range files {
		if strings.HasSuffix(fileInfo.Name(), suffix) {
			dest := filepath.Join(dirTestGames, fileInfo.Name())
			hist := testLoadHist(t, dest)
			src := dest[:len(dest)-len(suffix)]
			oldGame, err := loadOldGame(src)
			if err != nil {
				t.Errorf("Error while loading file: %v, Error %v", src, err)
			}
			err = os.Remove(dest)
			if err != nil {
				t.Errorf("Error while deleteing file: %v, Error %v", dest, err)
			}
			if hist != nil && oldGame != nil {
				newGame := game.NewGame()
				newGame.LoadHist(hist)
				oldGame.GameMoveLoop(func(
					gameMoveix int,
					pos *bold.GamePos,
					moveCardix, dealtix, moveix int,
					move bold.Move,
					isGiveUpMove, isPassMove bool,
					claimFailExs [9][]int) {

					newGame.ScrollForward() //Init Move
					newMoves := newGame.Pos.CalcMoves()
					compareMoves(pos.Moves, pos.MovesHand, pos.MovePass, newMoves, move, pos, newGame.Pos, t)
					compareCones(pos.Flags, newGame.Pos.ConePos, t)
				})
			}
		}
	}
}
func compareMoves(oldMoves []bold.Move,
	oldHandMoves map[int][]bold.Move,
	isPassMove bool,
	newMoves []*game.Move,
	madeOldMove bold.Move,
	oldPos *bold.GamePos,
	newPos *game.Pos,
	t *testing.T) {
	passMoves := 0
	if isPassMove {
		passMoves = 1
	}
	noOldMoves := len(oldMoves) + passMoves
	if len(oldHandMoves) > 0 {
		for _, moves := range oldHandMoves {
			noOldMoves = noOldMoves + len(moves)
		}
	}

	if noOldMoves == len(newMoves) {

	} else {
		t.Logf("\n\nNew Move: %v, Old Move: %v", newMoves[0], madeOldMove)
		handNo := make(map[int]int)
		for k, v := range oldHandMoves {
			handNo[k] = handNo[k] + len(v)
		}
		t.Logf("Pass Move: %v", isPassMove)
		t.Logf("newPos%v", *newPos)
		t.Logf("TacPos: %v", newPos.CardPos[61:])
		t.Logf("Old pos:%v", oldPos)
		t.Logf("Hand: %v", oldPos.Hands)
		t.Logf("falgs: %v", oldPos.Flags)
		t.Logf("Hand No: %v", handNo)
		t.Errorf("No. old moves and new moves differ: Old,New: %v,%v", noOldMoves, len(newMoves))
	}

}
func compareCones(
	flags [bold.NOFlags]*boldFlag.Flag,
	conePos [10]dPos.Cone,
	t *testing.T) {
	for flagix, flag := range flags {
		if flag.Claimed() {
			isWinner := flag.Won()
			if isWinner[0] {
				if conePos[flagix+1] != dPos.ConeAll.Players[0] {
					t.Errorf("Flag: %v deviates old won by: %v, new won by: %v",
						flagix+1, 0, conePos[flagix+1].Winner())
				}
			} else {
				if conePos[flagix+1] != dPos.ConeAll.Players[1] {
					t.Errorf("Flag: %v deviates old won by: %v, new won by: %v",
						flagix+1, 1, conePos[flagix+1].Winner())
				}
			}
		} else {
			if conePos[flagix+1] != dPos.ConeAll.None {
				t.Errorf("Flag: %v deviates old not won new won by: %v",
					flagix+1, conePos[flagix+1].Winner())
			}
		}

	}
}
func testLoadHist(t *testing.T, filePath string) (hist *game.Hist) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Errorf("Error loading file: %v Error: %v", filePath, err)
		return hist
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			t.Errorf("Error closing file: %v Error: %v", filePath, cerr)
		}
	}()
	var loadHist game.Hist
	err = gob.NewDecoder(file).Decode(&loadHist)
	if err != nil {
		t.Errorf("Error decoding file: %v Error: %v", filePath, err)
		return hist
	}
	return &loadHist
}
func TestDbGameFile(t *testing.T) {
	srcFilePath := filepath.Join(dirTestGames, "game.db")
	dstFilePath := filepath.Join(dirTestGames, "hist.db")
	err := DbFile(srcFilePath, dstFilePath)
	if err != nil {
		t.Errorf("Converting database file: %v failed. Errot: %v", srcFilePath, err)
		return
	}
	err = os.Remove(dstFilePath)
	if err != nil {
		t.Errorf("Deleting database file: %v failed. Errot: %v", dstFilePath, err)
		return
	}
}
func TestGamesFile(t *testing.T) {
	t.SkipNow() // Only used to create the test data once
	dir := "/home/rho/go/src/github.com/rezder/go-battleline/battserver/data"
	fileName := "savegames.gob"
	oldGames, err := tables.LoadSaveGames(filepath.Join(dir, fileName))
	if err != nil {
		t.Errorf("Loading file: %v failed, Error:%v", fileName, err)
		return
	}
	for id, oldGame := range oldGames.Games {
		fileName := "pauseGameNo" + id + ".gob"
		file, err := os.Create(filepath.Join(dirTestGames, fileName))
		if err != nil {
			t.Errorf("Creating file: %v failed, Error %v", fileName, err)
			break
		}
		err = bold.Save(oldGame, file, false)
		if err != nil {
			_ = file.Close()
			if err != nil {
				t.Errorf("Saving file: %v failed, Error %v", fileName, err)
				break
			}
		}
		err = file.Close()
		if err != nil {
			t.Errorf("Closing file: %v failed, Error %v", fileName, err)
			break
		}
	}
}
