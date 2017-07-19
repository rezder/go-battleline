package machine

import (
	"encoding/binary"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-error/log"
)

var (
	buckStd     = []byte("Std") // Card to flag and pass.
	buckDeck    = []byte("Deck")
	buckClaim   = []byte("Claim")
	buckSpecial = []byte("Special") //Traitor,Scout,Deserter, Redeploy, scoutReturn.
	buckGames   = []byte("Game")
	buckWin     = []byte("Win")
	buckLose    = []byte("Lose")
	//TODO we need a LegalMoves bucket except the made move which is on pos, remember pass move may be trouble
	keyMeta = []byte("Meta")
)

// DbPos a machine position database.
type DbPos struct {
	db         *bolt.DB
	maxFetchNo int
}

// MaxFetchNo the max number of records to fetch.
func (dbp *DbPos) MaxFetchNo() int {
	return dbp.maxFetchNo
}

//NewDbPos  create a battleline database.
func NewDbPos(db *bolt.DB, maxFetchNo int) *DbPos {
	dbp := new(DbPos)
	dbp.db = db
	dbp.maxFetchNo = maxFetchNo
	return dbp
}

//Init init the battleline bucket if does not exist.
func (dbp *DbPos) Init() error {
	err := dbp.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(buckGames)
		if err != nil {
			return errors.Wrapf(err, log.ErrNo(1)+"Creating bucket %v", string(buckGames))
		}
		return nil
	})
	return err
}

// AddGame adds all position from a game to the database.
func (dbp *DbPos) AddGame(game *bat.Game, id []byte) (err error) {
	meta, winnerMoves, loserMoves := extractMposs(game)
	err = dbp.db.Update(func(tx *bolt.Tx) error {
		gamesb := tx.Bucket(buckGames)
		gameBucket, updErr := gamesb.CreateBucket(id)
		if updErr != nil {
			return updErr
		}
		metaB, updErr := MetaEncode(meta)
		if updErr != nil {
			return updErr
		}
		updErr = gameBucket.Put(keyMeta, metaB)
		if updErr != nil {
			return updErr
		}
		updErr = updateMpos(gameBucket, buckWin, winnerMoves)
		if updErr != nil {
			return updErr
		}
		updErr = updateMpos(gameBucket, buckLose, loserMoves)
		return updErr
	})
	return err
}
func updateMpos(gameBucket *bolt.Bucket, id []byte, moves []*MPosJoin) (err error) {
	bucket, err := gameBucket.CreateBucket(id)
	if err != nil {
		return err
	}
	bucketNames := [][]byte{buckStd, buckDeck, buckClaim, buckSpecial}
	mposBuckets, err := createMposBuckets(bucket, bucketNames)
	if err != nil {
		return err
	}
	mposMoves := sortMpos(moves)
	for i, bucketName := range bucketNames {
		err = putMposList(mposBuckets[i], mposMoves[string(bucketName)])
		if err != nil {
			return err
		}
	}
	return err
}
func createMposBuckets(bucket *bolt.Bucket, bucketNames [][]byte) (buckets []*bolt.Bucket, err error) {
	buckets = make([]*bolt.Bucket, len(bucketNames))
	for i, buckID := range bucketNames {
		buckets[i], err = bucket.CreateBucket(buckID)
		if err != nil {
			return nil, err
		}
	}
	return buckets, err
}
func sortMpos(mPosJoins []*MPosJoin) (mposMoves map[string][]*MPosJoin) {
	n := (len(mPosJoins) + 2) / 3
	specialMoves := make([]*MPosJoin, 0, 4)
	stdMoves := make([]*MPosJoin, 0, n)
	claimMoves := make([]*MPosJoin, 0, n)
	deckMoves := make([]*MPosJoin, 0, n+3)
	for _, mPosJoin := range mPosJoins {
		mMove := mPosJoin.pos[len(mPosJoin.pos)-4:]
		if mMove[pMoveSpecialCard] == SPCClaimFlag {
			claimMoves = append(claimMoves, mPosJoin)
		} else if mMove[pMoveSpecialCard] == SPCDeck {
			deckMoves = append(deckMoves, mPosJoin)
		} else if mMove[pMoveSecondCard] != 0 {
			specialMoves = append(specialMoves, mPosJoin)
		} else if mMove[pMoveFirstCard] == cards.TCScout {
			deckMoves = append(deckMoves, mPosJoin)
			specialMoves = append(specialMoves, mPosJoin)
		} else if mMove[pMoveFirstCard] == 0 && mMove[pMoveSecondCard] == 0 { //pass
			specialMoves = append(specialMoves, mPosJoin)
		} else {
			stdMoves = append(stdMoves, mPosJoin)
		}
	}
	mposMoves = make(map[string][]*MPosJoin)
	mposMoves[string(buckClaim)] = claimMoves
	mposMoves[string(buckDeck)] = deckMoves
	mposMoves[string(buckSpecial)] = specialMoves
	mposMoves[string(buckStd)] = stdMoves
	return mposMoves
}
func putMposList(bucket *bolt.Bucket, mPosJoins []*MPosJoin) (err error) {
	for _, mPosJoin := range mPosJoins {
		key := make([]byte, 4)
		copy(key[:2], itob(uint16(mPosJoin.pos[0])))
		copy(key[2:], itob(0))
		err = bucket.Put(key, mPosJoin.pos)
		if err != nil {
			return err
		}
		for i, mMove := range mPosJoin.moves {
			copy(key[:2], itob(uint16(mPosJoin.pos[0])))
			copy(key[2:], itob(uint16(i+1)))
			err = bucket.Put(key, mMove)
			if err != nil {
				return err
			}
		}
	}
	return err
}
func extractMposs(game *bat.Game) (meta *Meta, winnerMPosJoins, loserMPosJoins []*MPosJoin) {
	var scoutReturnMove bat.Move
	scoutReturnMover := 0
	mPosJoins := [2][]*MPosJoin{make([]*MPosJoin, 0, 90), make([]*MPosJoin, 0, 90)}
	meta = NewMeta(game.PlayerIds, game.Starter)
	mover := 0
	game.GameMoveLoop(func(moveGameix int, pos *bat.GamePos, moveCardix, moveix int, move bat.Move, isGiveUp, isPass bool) {
		mover = pos.Player
		if isGiveUp {
			meta.GiveUp = true
		} else {
			mPos := CreatePos(pos, move, scoutReturnMove, isPass, moveCardix, mover, scoutReturnMover, moveGameix)
			mPosJoin := NewMPosJoin(mPos, pos.MovesHand, pos.MovePass, moveCardix, moveix, isPass)
			mPosJoins[mover] = append(mPosJoins[mover], mPosJoin)
			if moveCardix != 0 {
				meta.Players[mover].AddHandMove(move)
			}
			_, ok := move.(bat.MoveScoutReturn)
			if ok {
				scoutReturnMove = move
				scoutReturnMover = mover
			}
		}
	})
	meta.AddLastPosInfo(game.Pos)
	winner := mover
	noMoves := len(game.Moves)
	if meta.GiveUp {
		winner = opponent(winner)
		noMoves = noMoves - 1
	}
	meta.SetWinner(winner)
	meta.SetNoMoves(noMoves)
	winnerMPosJoins = mPosJoins[winner]
	loserMPosJoins = mPosJoins[opponent(winner)]

	return meta, winnerMPosJoins, loserMPosJoins
}

func opponent(playerix int) (opp int) {
	if playerix == 0 {
		opp = 1
	} else {
		opp = 0
	}
	return opp
}
func itob(v uint16) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint16(b, v)
	return b
}
