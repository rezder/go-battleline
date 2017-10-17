package dbhist

import (
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/rezder/go-battleline/v2/game"
)

func TestUpdateTimePlayers(t *testing.T) {
	fileName := "batttp.db"
	testPutGet(fileName, KeyTimePlayers, t)

}
func TestUpdatePlayersTime(t *testing.T) {
	fileName := "batttp.db"
	testPutGet(fileName, KeyPlayersTime, t)

}
func TestUpdate(t *testing.T) {
	fileName := "battplayers.db"
	testPutGet(fileName, KeyPlayers, t)
}

func TestSearch(t *testing.T) {
	name := "_test/testdb.db"
	//err := createTestDb(name) //adds 30 game histories
	//if err != nil {
	//		t.Fatalf("Create database failed: %v", err)
	//}

	bdb := testOpenDb(name, KeyPlayersTime, t)
	if bdb != nil {
		defer func() {
			if cerr := bdb.db.Close(); cerr != nil {
				t.Errorf("Close database file: %v failed: %v", name, cerr)
			}
		}()
	}
	id := 1
	prefix := itob(id)
	prefixHists, _, err := bdb.ScannPrefix(nil, prefix)
	if err != nil {
		t.Fatalf("Scann failed for prefix without function ids %v, error: %v", id, err)
	}
	if len(prefixHists) != 20 {
		t.Errorf("Scann failed for prefix found,expected: %v,%v no of game histories", len(prefixHists), 20)
	}
	noSearch := 0
	searchHists, _, err := bdb.Search(func(hist *game.Hist) bool {
		noSearch++
		if hist.PlayerIDs[0] > 2 || hist.PlayerIDs[0] < 1 ||
			hist.PlayerIDs[1] > 2 || hist.PlayerIDs[1] < 1 {
			return false
		}
		return true
	}, nil)
	if err != nil {
		t.Fatalf("Scann failed for seach with function %v", err)
	}

	noSearchExp := 30
	if noSearch != noSearchExp {
		t.Errorf("Empty condition Scann failed found: %v, expected: %v no. of games histories", noSearch, noSearchExp)
	}

	prefixFilterHists, _, err := bdb.ScannPrefix(testSearch, prefix)
	if err != nil {
		t.Fatalf("Scann failed for prefix with filter prefix: %v, error: %v", id, err)
	}
	if len(prefixFilterHists) != 10 {
		t.Errorf("Scann result failed found %v game histories expected %v", len(prefixFilterHists), 20)
	}
	if len(prefixFilterHists) != len(searchHists) || len(prefixFilterHists) != 10 {
		t.Errorf("Empty condition Scann failed found %v,%v no of game histories", len(prefixHists), len(searchHists))
	}
}
func testSearch(hist *game.Hist) bool {
	return len(hist.Moves) == 11
}
func testOpenDb(filePath string, keyf func(*game.Hist) []byte, t *testing.T) (bdb *Db) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		t.Fatalf("Open database file: %v failed: %v", filePath, err)
	}

	bdb = New(keyf, db, 25)
	err = bdb.Init()
	if err != nil {
		t.Fatalf("Init database failed: %v", err)
	}
	return bdb
}
func createTestDb(name string, t *testing.T) {
	bdb := testOpenDb(name, KeyPlayersTime, t)
	if bdb != nil {
		defer func() {
			if cerr := bdb.db.Close(); cerr != nil {
				t.Errorf("Close database file: %v failed: %v", name, cerr)
			}
		}()
	}
	p1 := 1
	p2 := 2
	p3 := 3
	for i := 0; i < 10; i++ {
		hist := createHist(p1, p2, 6)
		err := bdb.Put(hist)
		if err != nil {
			t.Errorf("Database file: %v Put 1. failed with Error: %v ", name, err)
		}
		hist = createHist(p3, p2, 8)
		err = bdb.Put(hist)
		if err != nil {
			t.Errorf("Database file: %v Put 2. failed with Error: %v ", name, err)
		}
		hist = createHist(p3, p1, 10)
		err = bdb.Put(hist)
		if err != nil {
			t.Errorf("Database file: %v Put 3. failed with Error: %v ", name, err)
		}
	}
}
func addMoves(no int, battGame *game.Game) {
	for i := 0; i < no; i++ {
		battGame.Move(battGame.Pos.CalcMoves()[0])
	}
}
func createHist(id1, id2, moveNo int) (hist *game.Hist) {
	battGame := game.NewGame()
	battGame.Start([2]int{id1, id2}, 1)
	addMoves(moveNo, battGame)
	return battGame.Hist
}
func testPutGet(dbname string, keyf func(*game.Hist) []byte, t *testing.T) {
	bdb := testOpenDb(dbname, keyf, t)
	if bdb != nil {
		defer func() {
			if cerr := bdb.db.Close(); cerr != nil {
				t.Errorf("Close database file: %v failed: %v", dbname, cerr)
			}
		}()
	}
	defer func() {
		if rerr := os.Remove(dbname); rerr != nil {
			t.Errorf("Delting file: %v, failed with Error: %v", dbname, rerr)
		}
	}()
	testSaveGame(bdb, t)
	testSaveGame(bdb, t)
}
func testSaveGames(bdb *Db, hist *game.Hist, t *testing.T) {
	hist2 := createHist(5, 2, 2)
	hist3 := createHist(4, 2, 7)
	hists := []*game.Hist{hist, hist2, hist3}
	err := bdb.Puts(hists)
	if err != nil {
		t.Fatalf("Save game histories failed: %v", err)
	}
	keys := make([][]byte, 0, len(hists))
	for _, hist := range hists {
		keys = append(keys, bdb.Key(hist))
	}
	savedHists, err := bdb.Gets(keys)
	if err != nil {
		t.Fatalf("Load game histories failed: %v", err)
	}
	for i, g := range savedHists {
		if g == nil {
			t.Fatalf("Load failed key: %v", string(keys[i]))
		}
	}
}
func testSaveGame(bdb *Db, t *testing.T) {
	hist := createHist(1, 2, 3)
	err := bdb.Put(hist)
	if err != nil {
		t.Fatalf("Save game history failed: %v", err)
	}
	histSaved, err := bdb.Get(bdb.Key(hist))
	if err != nil {
		t.Fatalf("Fetching game failed: %v", err)
	}
	if !histSaved.IsEqual(hist) {
		t.Error("Fetch corupt")
	}
}
