package battdb

import (
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
)

func TestUpdateTimePlayers(t *testing.T) {
	fileName := "batttp.db"
	putGetTest(fileName, KeyTimePlayers, t)

}
func TestUpdatePlayersTime(t *testing.T) {
	fileName := "batttp.db"
	putGetTest(fileName, KeyPlayersTime, t)

}
func TestUpdate(t *testing.T) {
	fileName := "battplayers.db"
	putGetTest(fileName, KeyPlayers, t)
}

func TestSearch(t *testing.T) {
	name := "_test/testdb.db"
	//err := createTestDb(name) //adds 30 game
	//if err != nil {
	//		t.Fatalf("Create database failed: %v", err)
	//	}
	db, err := bolt.Open(name, 0600, nil)
	if err != nil {
		t.Fatalf("Open database file failed: %v", err)
	}
	defer db.Close()
	bdb := New(KeyPlayersTime, db, 1000)
	err = bdb.Init()
	if err != nil {
		t.Fatalf("Init database failed: %v", err)
	}
	ids := [2]int{1, 2}
	prefix := KeyPlayerIds(ids)
	opPrefix := KeyPlayerIds([2]int{2, 1})
	for i, _ := range prefix {
		if prefix[i] != opPrefix[i] {
			t.Error("The id order should not matter")
			break
		}
	}

	prefixGames, _, err := bdb.ScannPrefix(nil, prefix)
	if err != nil {
		t.Fatalf("Scann failed for prefix without function ids %v, error: %v", ids, err)
	}
	noSearch := 0
	searchGames, _, err := bdb.Search(func(game *bat.Game, _ []byte) bool {
		noSearch++
		if game.PlayerIds[0] > 2 || game.PlayerIds[0] < 1 ||
			game.PlayerIds[1] > 2 || game.PlayerIds[1] < 1 {
			return false
		}
		return true
	})
	if err != nil {
		t.Fatalf("Scann failed for seach with function %v", err)
	}
	if len(prefixGames) != len(searchGames) || len(prefixGames) < 10 {
		t.Errorf("Empty condition Scann failed found %v,%v games", len(prefixGames), len(searchGames))
	}
	noSearchExp := 30
	if noSearch != noSearchExp {
		t.Errorf("Empty condition Scann failed found: %v, expected: %v games", noSearch, noSearchExp)
	}
	noSearchLoop := 0
	searchLoopGames, _, err := bdb.SearchLoop(func(game *bat.Game, _ []byte) bool {
		noSearchLoop++
		if game.PlayerIds[0] > 2 || game.PlayerIds[0] < 1 ||
			game.PlayerIds[1] > 2 || game.PlayerIds[1] < 1 {
			return false
		}
		return true
	}, nil)
	if err != nil {
		t.Fatalf("Scann failed for seach loop with function %v", err)
	}
	if len(searchGames) != len(searchLoopGames) {
		t.Errorf("Search and SearchLoop deviates %v,%v games", len(searchLoopGames), len(searchGames))
	}
	if noSearchLoop != noSearchExp {
		t.Errorf("Search and SearchLoop deviates %v,%v on loops", noSearch, noSearchExp)
	}

	prefixGames, _, err = bdb.ScannPrefix(testSearch, prefix)
	if err != nil {
		t.Fatalf("Scann failed for prefix %v with test function, error: %v", ids, err)
	}
	if len(prefixGames) != 7 {
		t.Errorf("Scann failed found %v games expected %v", len(prefixGames), 7)
	}
}
func testSearch(game *bat.Game, _ []byte) bool {
	game.CalcPos()
	for _, cardix := range game.Pos.Hands[0].Troops {
		troop, _ := cards.DrTroop(cardix)
		if troop.Value() > 8 {
			return true
		}
	}
	return false
}
func createTestDb(name string) (err error) {
	db, err := bolt.Open(name, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	bdb := New(KeyPlayersTime, db, 1000)
	err = bdb.Init()
	if err != nil {
		return err
	}
	p1 := 1
	p2 := 2
	p3 := 3
	for i := 0; i < 10; i++ {
		game := createGame(p1, p2, 6)
		_, err = bdb.Put(game)
		if err != nil {
			return err
		}
		time.Sleep(time.Second / 2)
		game = createGame(p3, p2, 8)
		_, err = bdb.Put(game)
		if err != nil {
			return err
		}
		game = createGame(p3, p1, 10)
		_, err = bdb.Put(game)
		if err != nil {
			return err
		}
		time.Sleep(time.Second / 2)
	}
	return err
}
func addMoves(no int, game *bat.Game) {
	game.Start(0)
	for i := 0; i < no; i++ {
		if len(game.Pos.MovesHand) > 0 {
			cardix := 0
			for ix := range game.Pos.MovesHand {
				cardix = ix
				break
			}
			game.MoveHand(cardix, len(game.Pos.MovesHand[cardix])-1)
		} else {
			game.Move(len(game.Pos.Moves) - 1)
		}
	}
}
func createGame(id1, id2, moveNo int) (game *bat.Game) {
	game = bat.New([2]int{id1, id2})
	addMoves(moveNo, game)
	return game
}
func putGetTest(dbname string, keyf func(*bat.Game) []byte, t *testing.T) {
	db, err := bolt.Open(dbname, 0600, nil)
	if err != nil {
		t.Fatalf("Open database failed: %v", err)
	}
	defer os.Remove(dbname)
	defer db.Close()
	bdb := New(keyf, db, 1000)
	err = bdb.Init()
	if err != nil {
		t.Fatalf("Init bucket failed: %v", err)
	}

	game := createGame(1, 2, 3)
	key, err := bdb.Put(game)
	if err != nil {
		t.Fatalf("Save game failed: %v", err)
	}
	gameSaved, err := bdb.Get(key)
	if err != nil {
		t.Fatalf("Fetching game failed: %v", err)
	}
	gameSaved.CalcPos()
	if !gameSaved.Equal(game) {
		t.Error("Fetch corupt")
	}
	game2 := createGame(5, 2, 2)
	game3 := createGame(4, 2, 7)
	games := []*bat.Game{game, game2, game3}
	keys, err := bdb.Puts(games)
	if err != nil {
		t.Fatalf("Save games failed: %v", err)
	}
	savedGames, err := bdb.Gets(keys)
	if err != nil {
		t.Fatalf("Load games failed: %v", err)
	}
	for i, g := range savedGames {
		if g == nil {
			t.Fatalf("Load failed key: %v", string(keys[i]))
		}
	}
}
