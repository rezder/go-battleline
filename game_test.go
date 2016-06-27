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

func TestGameSave(t *testing.T) {
	fileNamePos := "test/savePos.gob"
	fileName := "test/save.gob"
	game := New([2]int{1, 2})
	game.Start(1)
	filePos, errPos := os.Create(fileNamePos)
	file, err := os.Create(fileName)
	//f,err:=os.Open(fileName)
	GobRegistor()
	if err == nil && errPos == nil {
		err = Save(game, file, false)
		errPos = Save(game, filePos, true)
		file.Close()
		filePos.Close()
		if err != nil {
			t.Errorf("Save game file error. File :%v, Error: %v", fileName, err.Error())
		} else {
			file, err = os.Open(fileName)
			loadGame, err := Load(file)
			if err != nil {
				t.Errorf("Load game file error. File :%v, Error: %v", fileName, err.Error())
			} else {
				t.Logf("Comparing game in file: %v", fileName)
				compGames(game, loadGame, t)
				os.Remove(fileName)
			}
		}
		if errPos != nil {
			t.Errorf("Save game file error. File :%v, Error: %v", fileNamePos, errPos.Error())
		} else {
			filePos, err = os.Open(fileNamePos)
			loadGamePos, err := Load(filePos)
			if err != nil {
				t.Errorf("Load game file error. File :%v, Error: %v", fileNamePos, err.Error())
			} else {
				t.Logf("Comparing game in file: %v", fileNamePos)
				compGames(game, loadGamePos, t)
				os.Remove(fileNamePos)
			}
		}
	} else {
		t.Errorf("Create file error. Files :%v,%v Error: %v", fileName, fileNamePos, err.Error())
	}
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
func saveGame(fileName string, game *Game) (err error) {
	file, err := os.Create(fileName)
	if err == nil {
		GobRegistor()
		err = Save(game, file, true)
		file.Close()
	}
	return err
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
	if err == nil {
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
		for fileName, _ := range fileMap {
			file, err := os.Open(filepath.Join(dirName, fileName[3:]))
			defer file.Close()
			if err == nil {
				game, err := Load(file)
				if err == nil {
					posFile, err := os.Open(filepath.Join(dirName, fileName))
					defer posFile.Close()
					decoder := gob.NewDecoder(posFile)
					testPos := *new(TestPos)
					err = decoder.Decode(&testPos)
					if err == nil {
						gameMoveLoop(game, func(gamePos *GamePos, move int) {
							pos, found := testPos.Pos[move]
							if found {
								if !game.Pos.Equal(pos) {
									t.Errorf("Move: %v failed in file: %v", move, fileName)
								}
							}
						})
					}
				} else {
					t.Errorf("Error loading file: %v. Error: %v", fileName, err)
				}
			} else {
				t.Errorf("Error testing file: %v. Error: %v", fileName, err)
			}
		}
	} else {
		t.Errorf("Error listing directory: %v. Error: %v", dirName, err)
	}
}
func gameMoveLoop(game *Game, posFunc func(*GamePos, int)) {
	game.Pos = NewGamePos()
	game.Pos.DeckTroop = *game.InitDeckTroop.Copy()
	game.Pos.DeckTac = *game.InitDeckTac.Copy()
	deal(&game.Pos.Hands, &game.Pos.DeckTroop)
	game.Pos.Turn.start(game.Starter, game.Pos.Hands[game.Starter], &game.Pos.Flags,
		&game.Pos.DeckTac, &game.Pos.DeckTroop, &game.Pos.Dishs)
	for i, move := range game.Moves {
		if move[0] == -1 && move[1] == -1 {
			game.Quit(game.Pos.Player)
		} else if move[0] == 0 && move[0] == -1 {
			game.Pass()
		} else if move[0] > 0 {
			game.MoveHand(move[0], move[1])
		} else {
			game.Move(move[1])
		}
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
		file, err := os.Open(filepath.Join(dirName, fileName))
		defer file.Close()
		if err == nil {
			game, err := Load(file)
			if err == nil {
				Save(game, file, false)
				t.Logf("File: %v\nGame Pos:%v\nDeck:%v", fileName, game.Pos, game.InitDeckTroop)
				testPos := createPos(game)
				fileNamePos := ("pos" + fileName)
				posFile, err := os.Create(filepath.Join(dirName, fileNamePos))
				defer posFile.Close()
				if err == nil {
					err = testPos.Save(posFile)
					if err == nil {
						fileMap[fileNamePos] = true
					} else {
						t.Errorf("Error saving file: %v. Error: %v", fileName, err)
					}
				} else {
					t.Errorf("Error creating file: %v. Error: %v", fileName, err)
				}
			} else {
				t.Errorf("Error loading file: %v. Error: %v", fileName, err)
			}
		} else {
			t.Errorf("Error opening file: %v. Error: %v", fileName, err)
		}
	}
}
