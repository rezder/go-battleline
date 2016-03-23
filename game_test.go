package battleline

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
)

func TestGameSave(t *testing.T) {
	fileName := "test/save.gob"
	game := New(1, [2]int{1, 2})
	game.Start(1)
	file, err := os.Create(fileName)
	//f,err:=os.Open(fileName)
	GobRegistor()
	if err == nil {
		err = Save(game, file)
		file.Close()
		if err != nil {
			t.Errorf("Save game file error. File :%v, Error: %v", fileName, err.Error())
		} else {
			file, err = os.Open(fileName)
			loadGame, err := Load(file)
			if err != nil {
				t.Errorf("Load game file error. File :%v, Error: %v", fileName, err.Error())
			} else {
				checkGames(game, loadGame, t)
			}
		}
	} else {
		t.Errorf("Create file error. File :%v, Error: %v", fileName, err.Error())
	}
}
func TestDecodeGame(t *testing.T) {
	GobRegistor()
	game := New(1, [2]int{1, 2})
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
		checkGames(game, &loadGame, t)
	}
}
func checkGames(save *Game, load *Game, t *testing.T) {
	if !save.Pos.Turn.Equal(&load.Pos.Turn) {
		t.Logf("turn:\n%v\n load turn:\n%v", save.Pos.Turn, load.Pos.Turn)
		t.Error("Save and load turn not equal")
	} else if !save.Pos.Equal(&load.Pos) {
		t.Logf("Pos.Hands:\n%v\n, load pos.Hands:\n%v", save.Pos.Hands, load.Pos.Hands)
		t.Error("Save and load pos not equal")
	} else if true { //!save.Equal(load) {
		t.Logf("Before:\n%v\n", *save)
		t.Logf("After:\n%v", *load)
		t.Error("Save and load game not equal")
	}
}
