package machine

import (
	"bufio"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	bat "github.com/rezder/go-battleline/battleline"
)

func TestPosDb(t *testing.T) {
	posDbFile := "_test/testPos.db"
	mposDb, err := bolt.Open(posDbFile, 0600, nil)
	if err != nil {
		t.Fatalf("Faild to open file %v got error %v", posDbFile, err)
	}
	defer testCleanFile(t, posDbFile)
	defer mposDb.Close()
	mdb := NewDbPos(mposDb, 100)
	err = mdb.Init()
	if err != nil {
		t.Fatalf("Init database %v failed", posDbFile)
	}

	gameFile := "_test/game1vs21038.gob"
	game, err := testLoadGame(t, gameFile)
	if err != nil {
		t.Fatalf("Loading game file %v failed", gameFile)
	}
	err = mdb.AddGame(game, []byte("test"))
	if err != nil {
		t.Errorf("Adding game %v failed with error: %v", gameFile, err)
	}
	stdWriter := bufio.NewWriter(os.Stdout)
	err = PrintMachineData(OutTypeMoveCard, false, stdWriter, mdb, 10)
	if err != nil {
		t.Errorf("Error writing to std out failed: %v", err)
	}
	err = stdWriter.Flush()
	if err != nil {
		t.Errorf("Error flusing to std out failed: %v", err)
	}
	//TODO Load some data maybe
}
func testLoadGame(t *testing.T, name string) (game *bat.Game, err error) {
	file, err := os.Open(name)
	if err != nil {
		t.Errorf("Open file error. File :%v, Error: %v", name, err.Error())
		return game, err
	}
	game, err = bat.Load(file)
	if err != nil {
		t.Errorf("Load game file error. File :%v, Error: %v", name, err.Error())
		_ = file.Close() // ignore one error should be enough
		return game, err
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Closing file error. File :%v, Error: %v", file, err.Error())
		return game, err
	}
	return game, err
}
func testCleanFile(t *testing.T, file string) {
	err := os.Remove(file)
	if err != nil {
		t.Errorf("Deleting file %v failed with %v", file, err)
	}
}
