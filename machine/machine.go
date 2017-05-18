package machine

import (
	"io"
	"sort"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
)

const (
	OutTypeMoveCard = 0
	OutTypeDeck     = 2
	OutTypeClaim    = 3
)

func AddGames(bdb *battdb.Db, mdb *DbPos) (err error) {
	var games []*bat.Game
	var nextKey []byte
	for {
		keys := make([][]byte, 0, bdb.MaxFetchNo())
		games, nextKey, err = bdb.SearchLoop(func(_ *bat.Game, key []byte) bool {
			keys = append(keys, key)
			return true
		}, nextKey)
		if err != nil {
			err = errors.Wrap(err, "Search game database failed")
			return err
		}
		log.Printf(log.Min, "Loaded %v games", len(games))
		for i, game := range games {
			err = mdb.AddGame(game, keys[i])
			if err != nil {
				if err == bolt.ErrBucketExists {
					log.Printf(log.Debug, "Game %v was not added as it allready existed", string(keys[i]))
				} else {
					err = errors.Wrapf(err, "Fail to add game %v", string(keys[i]))
					return err
				}
			}
		}
		if len(nextKey) == 0 {
			break
		}
	}
	return err
}
func PrintMachineData(outType int, sparse bool, writer io.Writer, posDb *DbPos, gameLimit int) (err error) {
	//TODO more move types
	switch outType {
	case OutTypeMoveCard:
		err = writePos(writer, posDb, gameLimit, MPosCreateFeatureFlds(), sparse)
	case OutTypeDeck:
	case OutTypeClaim:
	}
	return err
}
func writePos(writer io.Writer, posDb *DbPos, gameLimit int, flds []Fld, sparse bool) (err error) {
	db := posDb.db
	err = db.View(func(tx *bolt.Tx) (verr error) {
		gamesBucket := tx.Bucket(buckGames)
		gamesCursor := gamesBucket.Cursor()
		noGame := 0
		for gameKey, _ := gamesCursor.Last(); gameKey != nil; gameKey, _ = gamesCursor.Prev() {
			gameBucket := gamesBucket.Bucket(gameKey)
			winBucket := gameBucket.Bucket(buckWin)
			readMovePosBucket(winBucket, buckStd, writer, flds, sparse)
			readMovePosBucket(winBucket, buckSpecial, writer, flds, sparse)
			noGame = noGame + 1
			if noGame == gameLimit {
				log.Print(log.Min, "Reach game limit")
				break
			}
		}
		log.Printf(log.Min, "Wrote %v games", noGame)
		return nil
	})
	return err
}
func readMovePosBucket(bucket *bolt.Bucket, posBucketKey []byte, writer io.Writer, flds []Fld, sparse bool) {
	posBucket := bucket.Bucket(posBucketKey)
	posCursor := posBucket.Cursor()
	var mMove []uint8
	mPos := make([]uint8, mPosNoBytes)
	count := 0
	for posKey, mData := posCursor.First(); posKey != nil; posKey, mData = posCursor.Next() {
		if len(mData) == mPosNoBytes {
			copy(mPos, mData)
			mMove = nil
			count = count + 1
		} else {
			mMove = mData
		}
		printlnOutTypeMove(writer, mPos, mMove, flds, sparse)
	}
}

func printlnOutTypeMove(writer io.Writer, mpos, mMove []uint8, flds []Fld, sparse bool) {

	if sparse {
		y, x := ExtractSparseRow(mpos, mMove, flds)
		lnDelimiter := []byte("\n")
		delimiter := []byte(" ")
		ixDelimiter := []byte(":")
		writer.Write([]byte(strconv.FormatFloat(y, 'f', -1, 64)))
		sortKeys := make([]int, 0, len(x))
		for i, _ := range x {
			sortKeys = append(sortKeys, i)
		}
		sort.Ints(sortKeys)
		for _, key := range sortKeys {
			writer.Write(delimiter)
			writer.Write([]byte(strconv.Itoa(key)))
			writer.Write(ixDelimiter)
			writer.Write([]byte(strconv.FormatFloat(x[key], 'f', -1, 64)))
		}
		writer.Write(lnDelimiter)
	} else {
		y, x := ExtractRow(mpos, mMove, flds)
		lnDelimiter := []byte("\n")
		delimiter := []byte(",")
		for _, v := range x {
			writer.Write([]byte(strconv.Itoa(int(v))))
			writer.Write(delimiter)
		}
		writer.Write([]byte(strconv.Itoa(int(y))))
		writer.Write(lnDelimiter)
	}
}
