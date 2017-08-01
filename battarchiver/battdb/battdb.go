package battdb

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
)

//Db is a battleline database
type Db struct {
	keyf       func(*bat.Game) []byte
	db         *bolt.DB
	bucket     []byte
	maxFetchNo int
}

//New  create a battleline database.
func New(keyf func(*bat.Game) []byte, db *bolt.DB, maxFetchNo int) *Db {
	bdb := new(Db)
	bdb.db = db
	bdb.keyf = keyf
	bdb.bucket = []byte("BattBucket")
	bdb.maxFetchNo = maxFetchNo
	return bdb
}
func (bdb *Db) MaxFetchNo() int {
	return bdb.maxFetchNo
}

//Init init the battleline bucket if does not exist.
func (bdb *Db) Init() error {
	err := bdb.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bdb.bucket)
		if err != nil {
			return errors.Wrapf(err, log.ErrNo(1)+"Creating bucket %v", string(bdb.bucket))
		}
		return nil
	})
	return err
}

func encode(game *bat.Game) (value []byte, err error) {
	pos := game.Pos
	game.Pos = nil
	var gameBuf bytes.Buffer
	encoder := gob.NewEncoder(&gameBuf)
	err = encoder.Encode(game)
	game.Pos = pos
	if err != nil {
		return value, err
	}
	value = gameBuf.Bytes()
	return value, err
}
func decode(value []byte) (game *bat.Game, err error) {
	buf := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buf)
	g := *new(bat.Game)
	err = decoder.Decode(&g)
	if err != nil {
		return game, err
	}
	game = &g
	return game, err
}

//Put adds a game to the bucket.
func (bdb *Db) Put(game *bat.Game) (key []byte, err error) {
	key = bdb.keyf(game)
	value, err := encode(game)
	if err != nil {
		return key, err
	}

	err = bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		updErr := b.Put(key, value)
		return updErr
	})
	return key, err
}

//Puts adds muliple games to bucket.
func (bdb *Db) Puts(games []*bat.Game) (keys [][]byte, err error) {
	keys = make([][]byte, len(games))
	values := make([][]byte, len(games))
	for i, game := range games {
		keys[i] = bdb.keyf(game)
		values[i], err = encode(game)
		if err != nil {
			return keys, err
		}
	}
	err = bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		for i := 0; i < len(keys); i++ {
			updErr := b.Put(keys[i], values[i])
			if updErr != nil {
				return updErr
			}
		}
		return nil
	})
	return keys, err
}
func copyBytes(b []byte) (cb []byte) {
	if b != nil {
		cb = make([]byte, len(b))
		copy(cb, b)
	}
	return cb
}

//Get fetch a game from the database.
func (bdb *Db) Get(key []byte) (game *bat.Game, err error) {
	var cb []byte
	err = bdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bdb.bucket)
		bv := bucket.Get(key)
		cb = copyBytes(bv)
		return nil
	})
	if err != nil {
		return game, err
	}
	if cb != nil {
		game, err = decode(cb)
	}
	return game, err
}

//Gets fetch multiple games from database.
func (bdb *Db) Gets(keys [][]byte) (games []*bat.Game, err error) {
	bv := make([][]byte, len(keys))
	games = make([]*bat.Game, len(keys))
	err = bdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bdb.bucket)
		for i, key := range keys {
			bv[i] = copyBytes(bucket.Get(key))
		}
		return nil
	})
	if err != nil {
		return games, err
	}
	var game *bat.Game
	for i, b := range bv {
		if b != nil {
			game, err = decode(b)
			if err != nil {
				return games, err
			}
			games[i] = game
		}
	}
	return games, err
}
func filterGame(games []*bat.Game,
	game *bat.Game,
	key []byte,
	filterF func(*bat.Game, []byte) bool,
	maxFetch int) ([]*bat.Game, bool) {
	isMaxFetch := false

	if filterF == nil || filterF(game, key) {
		games = append(games, game)
		if len(games) == maxFetch {
			isMaxFetch = true
		}
	}
	return games, isMaxFetch
}

//ScannStartEnd scans a interval of keys.
//The keys must be bytes comparable and the start key must exist.
func (bdb *Db) ScannStartEnd(
	filterF func(game *bat.Game, key []byte) bool,
	start []byte,
	end []byte) (games []*bat.Game, isMaxFetch bool, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		c := b.Cursor()
		for k, v := c.Seek(start); k != nil && bytes.Compare(k, end) <= 0; k, v = c.Next() {
			game, decodeErr := decode(v)
			if decodeErr != nil {
				return decodeErr
			}
			games, isMaxFetch = filterGame(games, game, k, filterF, bdb.maxFetchNo)
			if isMaxFetch {
				break
			}
		}
		return nil
	})

	return games, isMaxFetch, err
}

//ScannPrefix makes a prefix scann.
func (bdb *Db) ScannPrefix(
	filterF func(game *bat.Game, key []byte) bool,
	prefix []byte) (games []*bat.Game, isMaxFetch bool, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			game, decodeErr := decode(v)
			if decodeErr != nil {
				return decodeErr
			}
			games, isMaxFetch = filterGame(games, game, k, filterF, bdb.maxFetchNo)
			if isMaxFetch {
				break
			}
		}
		return nil
	})

	return games, isMaxFetch, err
}

//Search the full database from the last entry.
//The search stops if the max numbers of records is reach.
func (bdb *Db) Search(
	filterF func(game *bat.Game, key []byte) bool) (games []*bat.Game, nextKey []byte, err error) {

	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		c := b.Cursor()
		isMaxFetch := false
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			game, decodeErr := decode(v)
			if decodeErr != nil {
				return decodeErr
			}
			games, isMaxFetch = filterGame(games, game, k, filterF, bdb.maxFetchNo)
			if isMaxFetch {
				k, _ = c.Prev()
				if k != nil {
					nextKey = k
				}
				break
			}
		}
		return nil
	})
	return games, nextKey, err
}

//Search the full database from the last entry.
//The search stops if the max numbers of records is reach.
func (bdb *Db) SearchLoop(
	filterF func(game *bat.Game, key []byte) bool,
	startKey []byte) (games []*bat.Game, nextKey []byte, err error) {

	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucket)
		c := b.Cursor()
		isMaxFetch := false
		var k, v []byte
		if len(startKey) > 0 {
			k, v = c.Seek(startKey)
		} else {
			k, v = c.Last()
		}
		for ; k != nil; k, v = c.Prev() {
			game, decodeErr := decode(v)
			if decodeErr != nil {
				return decodeErr
			}
			games, isMaxFetch = filterGame(games, game, k, filterF, bdb.maxFetchNo)
			if isMaxFetch {
				k, _ = c.Prev()
				if k != nil {
					nextKey = k
				}
				break
			}
		}
		return nil
	})
	return games, nextKey, err
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

//KeyPlayerIds define a key as player ids where the smalles player
//id comes first.
func KeyPlayerIds(ids [2]int) (key []byte) {
	key = make([]byte, 16)
	if ids[0] > ids[1] {
		copy(key, itob(ids[1]))
		copy(key[8:], itob(ids[0]))
	} else {
		copy(key, itob(ids[0]))
		copy(key[8:], itob(ids[1]))
	}
	return key
}

//KeyPlayersTime define a key as timestamp 2006-01-02T15:04:05Z07:00
//plus player ids where the smalles player id comes first.
func KeyPlayersTime(game *bat.Game) (key []byte) {
	key = KeyPlayers(game)
	ts := time.Now()
	ts3339 := []byte(ts.Format(time.RFC3339))
	key = append(key, ts3339...)
	frac100 := ts.Nanosecond() / 1000000
	key = append(key, itob(frac100)...)
	return key
}

//KeyTimePlayers define a key as player ids where the smalles player
//id comes first plus timestamp 2006-01-02T15:04:05Z07:00.
func KeyTimePlayers(game *bat.Game) (key []byte) {
	ts := time.Now()
	key = []byte(ts.Format(time.RFC3339))
	key = append(key, KeyPlayers(game)...)
	frac100 := ts.Nanosecond() / 1000000
	key = append(key, itob(frac100)...)
	return key
}

//KeyPlayers define a key as player ids where the smalles player
//id comes first.
func KeyPlayers(game *bat.Game) (b []byte) {
	return KeyPlayerIds(game.PlayerIds)
}
func (bdb *Db) BackupHandleFunc(w http.ResponseWriter, req *http.Request) {
	err := bdb.db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
