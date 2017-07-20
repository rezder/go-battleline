package battleline

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func saveGame(t *testing.T, name string, game *Game, savePos bool) (err error) {
	file, err := os.Create(name)
	if err != nil {
		t.Errorf("Create file error. File :%v Error: %v", name, err.Error())
		return err
	}

	err = Save(game, file, savePos)
	if err != nil {
		t.Errorf("Save game file error. File :%v, Error: %v", file, err.Error())
		file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Closing file error. File :%v, Error: %v", file, err.Error())
	}
	return err
}

func loadGame(t *testing.T, name string) (game *Game, err error) {
	file, err := os.Open(name)
	if err != nil {
		t.Errorf("Open file error. File :%v, Error: %v", name, err.Error())
		return game, err
	}
	game, err = Load(file)
	if err != nil {
		t.Errorf("Load game file error. File :%v, Error: %v", name, err.Error())
		file.Close()
		return game, err
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Closing file error. File :%v, Error: %v", file, err.Error())
		return game, err
	}
	return game, err
}

// compareSavedGame compare a saved game with the game and WARNING delete the file
func compareSavedGame(t *testing.T, name string, game *Game) {
	loadGame, err := loadGame(t, name)
	if err != nil {
		return
	}
	err = os.Remove(name)
	if err != nil {
		t.Errorf("Remove game file error. File :%v, Error: %v", name, err.Error())
		return
	}
	t.Logf("Comparing game in file: %v", name)
	compGames(game, loadGame, t)

	return

}
func TestGameSave(t *testing.T) {
	fileNamePos := "test/savePos.gob"
	fileName := "test/save.gob"
	game := New([2]int{1, 2})
	game.Start(1)
	GobRegistor()
	err := saveGame(t, fileName, game, false)
	if err != nil {
		return
	}
	err = saveGame(t, fileNamePos, game, true)
	if err != nil {
		return
	}
	compareSavedGame(t, fileNamePos, game)
	compareSavedGame(t, fileName, game)

}
func TestDecodeGame(t *testing.T) {
	GobRegistor()
	game := New([2]int{1, 2})
	game.Start(0)
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(game)
	if err != nil {
		t.Errorf("Error encoding: %v", err)
		return
	}

	var loadGame Game
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err = d.Decode(&loadGame)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	} else {
		compGames(game, &loadGame, t)
	}
}
func compGames(save *Game, load *Game, t *testing.T) {
	if !save.Pos.Turn.Equal(&load.Pos.Turn) {
		t.Logf("turn:\n%v\n load turn:\n%v", save.Pos.Turn, load.Pos.Turn)
		t.Error("Save and load turn not equal")
	} else if !save.Pos.Equal(load.Pos) {
		t.Logf("Pos.Hands:\n%v\n, load pos.Hands:\n%v", save.Pos.Hands, load.Pos.Hands)
		t.Error("Save and load pos not equal")
	} else if !save.Equal(load) {
		t.Logf("Before game: %v", save)
		t.Logf("After game: %v", load)
		t.Error("Save and load game not equal")
	}
}

func playerClaimFlag(game *Game, flag int) {
	var move Move
	if flag == -1 {
		move = *NewMoveClaim([]int{}) //make([]int,0)
	} else {
		move = *NewMoveClaim([]int{flag})
	}
	game.Move(game.Pos.GetMoveix(0, move))

}
func moveHandDeck(game *Game, card int, flag int, deck int) {
	var move Move
	move = *NewMoveCardFlag(flag)
	ix := game.Pos.GetMoveix(card, move)
	game.MoveHand(card, ix)
	move = *NewMoveDeck(deck)
	game.Move(game.Pos.GetMoveix(0, move))

}

type TestPos struct {
	Pos map[int]*GamePos
}

func (pos *TestPos) Save(file *os.File) (err error) {
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(pos)
	return err
}

func TestGames(t *testing.T) {
	GobRegistor()
	dirName := "test"
	fileMap := make(map[string]bool)
	newGames := make([]string, 0, 0)
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		t.Errorf("Error listing directory: %v. Error: %v", dirName, err)
		return
	}

	for _, fileInfo := range files {
		if strings.Contains(fileInfo.Name(), "pos") {
			fileMap[fileInfo.Name()] = true
		}
	}
	for _, fileInfo := range files {
		if !strings.Contains(fileInfo.Name(), "pos") {
			if !fileMap["pos"+fileInfo.Name()] {
				newGames = append(newGames, fileInfo.Name())
			}
		}
	}
	createGameTest(fileMap, newGames, dirName, t)
	for fileName := range fileMap {
		game, err := loadGame(t, filepath.Join(dirName, fileName[3:]))
		if err == nil {
			testPos, err := loadGamePositions(t, filepath.Join(dirName, fileName))
			if err == nil {
				gameMoveLoop(game, func(gamePos *GamePos, move int) {
					pos, found := testPos.Pos[move]
					if found {
						if !game.Pos.Equal(pos) {
							for i := range pos.Flags {
								if !pos.Flags[i].Equal(game.Pos.Flags[i]) {
									t.Logf("ix: %v, oldFlag %v, newflag %v", i, pos.Flags[i], game.Pos.Flags[i])
								}
							}
							t.Errorf("Move: %v failed in file: %v", move, fileName)
						}
					}
				})
			}
		}
	}

}
func loadGamePositions(t *testing.T, name string) (pos *TestPos, err error) {
	posFile, err := os.Open(name)
	if err != nil {
		t.Errorf("Error opning file: %v. Error: %v", name, err)
		return pos, err
	}
	decoder := gob.NewDecoder(posFile)
	testPos := *new(TestPos)
	err = decoder.Decode(&testPos)
	if err != nil {
		posFile.Close()
		t.Errorf("Error decoding file: %v. Error: %v", name, err)
		return pos, err
	}
	err = posFile.Close()
	if err != nil {
		t.Errorf("Error closing file: %v. Error: %v", name, err)
	}
	return &testPos, err
}

func gameMoveLoop(game *Game, posFunc func(*GamePos, int)) {
	moves := game.ResetGame()
	for i, move := range moves {
		histMove(move, game)
		posFunc(game.Pos, i)
	}
}

func createPos(game *Game) (testPos *TestPos) {
	testPos = new(TestPos)
	testPos.Pos = make(map[int]*GamePos)

	gameMoveLoop(game, func(gamePos *GamePos, move int) {
		testPos.Pos[move] = gamePos.Copy()
	})

	return testPos
}
func createGameTest(fileMap map[string]bool, newGames []string, dirName string, t *testing.T) {
	for _, fileName := range newGames {
		//file, err := os.OpenFile(filepath.Join(dirName, fileName), os.O_RDWR, 0666)
		game, err := loadGame(t, filepath.Join(dirName, fileName))
		if err == nil {
			err = saveGame(t, filepath.Join(dirName, fileName), game, false) //remove possition
			if err == nil {
				//	t.Logf("File: %v\nGame Pos:%v\nDeck:%v", fileName, game.Pos, game.InitDeckTroop)
				testPos := createPos(game)
				fileNamePos := ("pos" + fileName)
				err = saveTestPos(t, filepath.Join(dirName, fileNamePos), testPos)
				if err == nil {
					fileMap[fileNamePos] = true
				}
			} else {
				t.Errorf("Error saveing game:%v", fileName)
			}
		} else {
			t.Errorf("Error loading game:%v", fileName)
		}
	}
}
func saveTestPos(t *testing.T, name string, testPos *TestPos) (err error) {
	file, err := os.Create(name)
	if err != nil {
		t.Errorf("Create file error. File :%v Error: %v", name, err.Error())
		return err
	}

	err = testPos.Save(file)
	if err != nil {
		t.Errorf("Save test possition file error. File :%v, Error: %v", file, err.Error())
		file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Closing file error. File :%v, Error: %v", file, err.Error())
	}
	return err
}

//dataConversion was use with a data error
func dataConversion(game *Game) {
	for i, move := range game.Moves {
		if move[1] >= 0 {
			if move[0] == -1 {
				move[0] = 0
				game.Moves[i] = move
			}
		} else {
			if move[0] == -1 {
				move[1] = -2
				move[0] = 0
				game.Moves[i] = move
			}
		}
	}
}
func TestClaimCombi(t *testing.T) {
	flags := []int{0}
	m := claimCombi(flags)
	no := claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2, 3}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2, 3, 4}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2, 3, 4, 5}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}

	flags = []int{0, 1, 2, 3, 4, 5, 6}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2, 3, 4, 5, 6, 7}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
	flags = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	m = claimCombi(flags)
	no = claimCombiNo(len(flags))
	if len(m) != no {
		t.Errorf("%v flags expected %v, got %v", len(flags), len(m), no)
	}
}
