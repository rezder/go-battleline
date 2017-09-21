package dbhist

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
)

const (
	//TimeFormat is the RFC3339Nano with trailing zeros.
	TimeFormat = "2006-01-02T15:04:05.000000000Z07:00"
)

//Db is a battleline database
type Db struct {
	keyf       func(*game.Hist) []byte
	db         *bolt.DB
	bucketID   []byte
	maxFetchNo int
}

//New  create a battleline database.
func New(keyf func(*game.Hist) []byte, db *bolt.DB, maxFetchNo int) *Db {
	bdb := new(Db)
	bdb.db = db
	bdb.keyf = keyf
	bdb.bucketID = []byte("BattBucket")
	bdb.maxFetchNo = maxFetchNo
	return bdb
}

//Close the bolt database
func (bdb *Db) Close() error {
	return bdb.db.Close()
}

//MaxFetchNo returns the max. number of records fetched in search.
func (bdb *Db) MaxFetchNo() int {
	return bdb.maxFetchNo
}

//Key returns the key.
func (bdb *Db) Key(hist *game.Hist) []byte {
	return bdb.keyf(hist)
}

//Init inits the battleline bucket if does not exist.
func (bdb *Db) Init() error {
	err := bdb.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bdb.bucketID)
		if err != nil {
			return errors.Wrapf(err, log.ErrNo(1)+"Creating bucket %v", string(bdb.bucketID))
		}
		return nil
	})
	return err
}

func encode(hist *game.Hist) (bs []byte, err error) {
	var histBuf bytes.Buffer
	encoder := gob.NewEncoder(&histBuf)
	err = encoder.Encode(hist)
	if err != nil {
		return bs, err
	}
	bs = histBuf.Bytes()
	return bs, err
}
func decode(bs []byte) (hist *game.Hist, err error) {
	buf := bytes.NewBuffer(bs)
	decoder := gob.NewDecoder(buf)
	h := *new(game.Hist)
	err = decoder.Decode(&h)
	if err != nil {
		return hist, err
	}
	hist = &h
	return hist, err
}

//Put adds a history of a game to the bucket.
func (bdb *Db) Put(hist *game.Hist) (err error) {
	key := bdb.keyf(hist)
	bs, err := encode(hist)
	if err != nil {
		return err
	}

	err = bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucketID)
		updErr := b.Put(key, bs)
		return updErr
	})
	return err
}

//Puts adds muliple histories to bucket.
func (bdb *Db) Puts(hists []*game.Hist) (err error) {
	keys := make([][]byte, len(hists))
	bss := make([][]byte, len(hists))
	for i, hist := range hists {
		keys[i] = bdb.keyf(hist)
		bss[i], err = encode(hist)
		if err != nil {
			return err
		}
	}
	err = bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucketID)
		for i := 0; i < len(keys); i++ {
			updErr := b.Put(keys[i], bss[i])
			if updErr != nil {
				return updErr
			}
		}
		return nil
	})
	return err
}

//Delete deletes a key from database.
func (bdb *Db) Delete(key []byte) (err error) {
	err = bdb.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bdb.bucketID)
		deleteErr := bucket.Delete(key)
		return deleteErr
	})
	return err
}

//Get fetches a game history from the database.
func (bdb *Db) Get(key []byte) (hist *game.Hist, err error) {
	var cbs []byte
	err = bdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bdb.bucketID)
		cbs = bucket.Get(key)
		return nil
	})
	if err != nil {
		return hist, err
	}
	if cbs != nil {
		hist, err = decode(cbs)
	}
	return hist, err
}

//Gets fetches multiple games from database.
func (bdb *Db) Gets(keys [][]byte) (hists []*game.Hist, err error) {
	bss := make([][]byte, len(keys))
	hists = make([]*game.Hist, len(keys))
	err = bdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bdb.bucketID)
		for i, key := range keys {
			bss[i] = bucket.Get(key)
		}
		return nil
	})
	if err != nil {
		return hists, err
	}
	var hist *game.Hist
	for i, bs := range bss {
		if bs != nil {
			hist, err = decode(bs)
			if err != nil {
				return hists, err
			}
			hists[i] = hist
		}
	}
	return hists, err
}
func filter(
	hists []*game.Hist,
	hist *game.Hist,
	key []byte,
	filterFunc func(*game.Hist) bool,
	maxFetch int) ([]*game.Hist, bool) {

	isMaxFetch := false
	if filterFunc == nil || filterFunc(hist) {
		hists = append(hists, hist)
		if len(hists) == maxFetch {
			isMaxFetch = true
		}
	}
	return hists, isMaxFetch
}

//ScannStartEnd scans a interval of keys.
//The keys must be bytes comparable and the start key must exist.
func (bdb *Db) ScannStartEnd(
	filterFunc func(game *game.Hist) bool,
	start []byte,
	end []byte) (hists []*game.Hist, isMaxFetch bool, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucketID)
		c := b.Cursor()
		for k, bs := c.Seek(start); k != nil && bytes.Compare(k, end) <= 0; k, bs = c.Next() {
			hist, decodeErr := decode(bs)
			if decodeErr != nil {
				return decodeErr
			}
			hists, isMaxFetch = filter(hists, hist, k, filterFunc, bdb.maxFetchNo)
			if isMaxFetch {
				break
			}
		}
		return nil
	})

	return hists, isMaxFetch, err
}

//ScannPrefix makes a prefix scann.
func (bdb *Db) ScannPrefix(
	filterFunc func(hist *game.Hist) bool,
	prefix []byte) (hists []*game.Hist, isMaxFetch bool, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucketID)
		c := b.Cursor()
		for k, bs := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, bs = c.Next() {
			hist, decodeErr := decode(bs)
			if decodeErr != nil {
				return decodeErr
			}
			hists, isMaxFetch = filter(hists, hist, k, filterFunc, bdb.maxFetchNo)
			if isMaxFetch {
				break
			}
		}
		return nil
	})

	return hists, isMaxFetch, err
}

//Search the database from the start key
//if next key is not empty more enteries exist,than could be loaded
func (bdb *Db) Search(
	filterFunc func(hist *game.Hist) bool,
	startKey []byte) (hists []*game.Hist, nextKey []byte, err error) {

	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bdb.bucketID)
		c := b.Cursor()
		isMaxFetch := false
		var k, bs []byte
		if len(startKey) > 0 {
			k, bs = c.Seek(startKey)
		} else {
			k, bs = c.Last()
		}
		for ; k != nil; k, bs = c.Prev() {
			hist, decodeErr := decode(bs)
			if decodeErr != nil {
				return decodeErr
			}
			hists, isMaxFetch = filter(hists, hist, k, filterFunc, bdb.maxFetchNo)
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
	return hists, nextKey, err
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

//KeyPlayerIDs define a key as player ids where the smalles player
//id comes first.
func KeyPlayerIDs(ids [2]int) (key []byte) {
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

//KeyPlayersTime define a key as timestamp 2006-01-02T15:04:05.000000000Z07:00
//plus player ids where the smalles player id comes first.
func KeyPlayersTime(hist *game.Hist) (key []byte) {
	key = KeyPlayers(hist)
	ts3339NanoBs := []byte(hist.Time.Format(TimeFormat))
	key = append(key, ts3339NanoBs...)
	return key
}

//KeyTimePlayers define a key as player ids where the smalles player
//id comes first plus timestamp 2006-01-02T15:04:05.000000000Z07:00
func KeyTimePlayers(hist *game.Hist) (key []byte) {
	key = []byte(hist.Time.Format(TimeFormat))
	key = append(key, KeyPlayers(hist)...)
	return key
}

//KeyPlayers define a key as player ids where the smalles player
//id comes first.
func KeyPlayers(hist *game.Hist) (b []byte) {
	return KeyPlayerIDs(hist.PlayerIDs)
}

//BackupHandleFunc handles http back up requests.
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
