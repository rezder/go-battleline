package machine

import (
	"fmt"

	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
)

type MMove []uint8

const (
	mMoveNoBytes = 4
)

var (
	pMoveFirstCard      = 0 //1851 71 does not include special SPCClaimFlag or SPCDeck
	pMoveFirstCardDest  = 1 //1922 11 pos:deck,flag,dish
	pMoveSecondCard     = 2 //1933 70 no scout
	pMoveSecondCardDest = 3 //2003 11 pos:deck,dish,flag. Features 2013
	pMoveSpecialCard    = 0
	pMoveDeck           = 1
	pMoveClaimFlag9     = 1
	pMoveClaimFlags     = 2
)

func NewMMove() MMove {
	mMove := make([]uint8, mMoveNoBytes)
	return mMove
}
func CreateMMove(cardix int, move bat.Move, pass bool) (mMove MMove) {
	mMove = NewMMove()
	if !pass {
		if cardix == 0 {
			switch move := move.(type) {
			case bat.MoveDeck:
				mMove[pMoveSpecialCard] = SPCDeck
				mMove[pMoveDeck] = deckMoveToByte(move.Deck)
			case bat.MoveClaim:
				mMove[pMoveSpecialCard] = SPCClaimFlag
				mMove[pMoveClaimFlag9], mMove[pMoveClaimFlags] = claimMoveToByte(move.Flags)
			case bat.MoveScoutReturn:
				mMove[pMoveFirstCard], mMove[pMoveSecondCard] = scoutReturnToByte(move.Tac, move.Troop)
				mMove[pMoveFirstCardDest] = MoveDestAll.Deck
				mMove[pMoveSecondCardDest] = MoveDestAll.Deck
			}
		} else {
			mMove[pMoveFirstCard] = uint8(cardix)
			switch cardMove := move.(type) {
			case bat.MoveCardFlag:
				mMove[pMoveFirstCardDest] = MoveDestAll.Flag[cardMove.Flagix]
			case bat.MoveDeserter:
				mMove[pMoveFirstCardDest] = MoveDestAll.Dish
				mMove[pMoveSecondCard] = uint8(cardMove.Card)
				mMove[pMoveSecondCardDest] = MoveDestAll.Dish
			case bat.MoveRedeploy:
				mMove[pMoveFirstCardDest] = MoveDestAll.Dish
				mMove[pMoveSecondCard] = uint8(cardMove.OutCard)
				if cardMove.InFlag == -1 {
					mMove[pMoveSecondCardDest] = MoveDestAll.Dish
				} else {
					mMove[pMoveSecondCardDest] = MoveDestAll.Flag[cardMove.InFlag]
				}
			case bat.MoveTraitor:
				mMove[pMoveFirstCardDest] = MoveDestAll.Dish
				mMove[pMoveSecondCard] = uint8(cardMove.OutCard)
				mMove[pMoveSecondCardDest] = MoveDestAll.Flag[cardMove.InFlag]
			case bat.MoveDeck: //scout
				mMove[pMoveFirstCardDest] = MoveDestAll.Dish
				mMove[pMoveDeck] = deckMoveToByte(cardMove.Deck)
			}

		}
	}
	return mMove
}
func (mMove MMove) cardPos(cardix uint8) (cp uint8) {
	return mMove[pCard+int(cardix)]

}
func flagixFromCardPos(cardPos uint8) (flagix int) {
	flagix, ok := CardPosAll.Flagix(cardPos)
	if !ok {
		panic(fmt.Sprintf("Card position %v is not a flag", CardPosAll.ValueToTxt(cardPos)))
	}
	return flagix
}
func (mMove MMove) Reverse() (cardix int, move bat.Move, pass bool) {
	if mMove[pMoveFirstCard] == 0 && mMove[pMoveSecondCard] == 0 {
		pass = true
	} else if mMove[pMoveSpecialCard] == SPCDeck {
		deck := deckMoveToInt(mMove[pMoveDeck])
		move = bat.NewMoveDeck(deck)
	} else if mMove[pMoveSpecialCard] == SPCClaimFlag {
		claimFlags := claimMoveToFlags(mMove[pMoveClaimFlags], mMove[pMoveClaimFlag9]) //TODO check if nil is ok
		move = bat.NewMoveClaim(claimFlags)
	} else if mMove[pMoveFirstCardDest] == MoveDestAll.Deck && mMove[pMoveFirstCard] != 0 {
		troops, tacs := scoutReturnToCards(mMove[pMoveFirstCard], mMove[pMoveSecondCard])
		move = bat.NewMoveScoutReturn(tacs, troops)
	} else {
		cardix = int(mMove[pMoveFirstCard])
		switch cardix {
		case cards.TCDeserter:
			cardPos := mMove.cardPos(mMove[pMoveSecondCard])
			flagix := flagixFromCardPos(cardPos)
			move = bat.NewMoveDeserter(flagix, int(mMove[pMoveSecondCard]))
		case cards.TCTraitor:
			move = mMove.doubleMove(true)
		case cards.TCScout:
			deck := deckMoveToInt(mMove[pMoveDeck])
			move = bat.NewMoveDeck(deck)
		case cards.TCRedeploy:
			move = mMove.doubleMove(false)
		default:
			dest := mMove[pMoveFirstCardDest]
			flagix := flagixFromCardPos(dest)
			move = bat.NewMoveCardFlag(flagix)
		}
	}

	return cardix, move, pass
}
func (mMove MMove) doubleMove(isTraitor bool) (move bat.Move) {
	cardPos := mMove.cardPos(mMove[pMoveSecondCard])
	outFlagix := flagixFromCardPos(cardPos)
	dest := mMove[pMoveSecondCardDest]
	inFlagix := flagixFromCardPos(dest)
	if isTraitor {
		move = bat.NewMoveRedeploy(outFlagix, int(mMove[pMoveSecondCard]), inFlagix)
	} else {
		move = bat.NewMoveTraitor(outFlagix, int(mMove[pMoveSecondCard]), inFlagix)
	}
	return move
}
func scoutReturnToByte(tacs, troops []int) (firstReturnedCard, secondReturnedCard uint8) {
	if len(tacs) == 1 {
		firstReturnedCard = uint8(tacs[0])
		if len(troops) == 1 {
			secondReturnedCard = uint8(troops[0])
		}
	} else if len(troops) == 2 {
		firstReturnedCard = uint8(troops[0])
		secondReturnedCard = uint8(troops[1])
	} else if len(tacs) == 2 {
		firstReturnedCard = uint8(tacs[0])
		secondReturnedCard = uint8(tacs[1])
	}
	return firstReturnedCard, secondReturnedCard
}
func scoutReturnToCards(firstCard, secondCard uint8) (tacs, troops []int) {
	tacs, troops = scoutReturnToCardsAdd(firstCard, tacs, troops)
	tacs, troops = scoutReturnToCardsAdd(secondCard, tacs, troops)
	return tacs, troops
}
func scoutReturnToCardsAdd(card uint8, tacs, troops []int) ([]int, []int) {
	if cards.IsTac(int(card)) {
		tacs = append(tacs, int(card))
	} else {
		troops = append(troops, int(card))
	}
	return tacs, troops
}
func deckMoveToByte(deck int) uint8 {
	if deck == bat.DECKTac {
		return 1
	}
	return 0
}
func deckMoveToInt(isTac uint8) int {
	if isTac == 1 {
		return bat.DECKTac
	}
	return bat.DECKTroop
}
func claimMoveToByte(flags []int) (flag9, bitFlags uint8) {
	var flaglxs [8]bool
	for _, flagix := range flags {
		if flagix == 8 {
			flag9 = 1
		} else {
			flaglxs[flagix] = true
		}
	}
	bitFlags = convertBitFlagToByte(flaglxs)
	return flag9, bitFlags
}

//claimMoveToFlags return a list of claimed flags
// nil if no claim.
func claimMoveToFlags(flag9, bitFlags uint8) (flags []int) {
	boolFlags := convertByteToBoolFlag(bitFlags)
	for i, isClaim := range boolFlags {
		if isClaim {
			flags = append(flags, i)
		}
	}
	if flag9 == 1 {
		flags = append(flags, 8)
	}
	return flags //TODO check if nil is ok
}

func convertByteToBoolFlag(x uint8) (flags [8]bool) {
	for i := uint(0); i < 8; i++ {
		flags[i] = x&(1<<i)>>i == 1
	}
	return flags
}
func convertBitFlagToByte(flags [8]bool) (x uint8) {
	for i, b := range flags {
		if b {
			x = x | (1 << uint8(i))
		}
	}
	return x
}

// Special move
const (
	SPCClaimFlag = 100
	SPCDeck      = 101
)

type MoveDestAllSingleton struct {
	SimpleDom
	Dish uint8
	Flag [9]uint8
	Deck uint8
}

var (
	MoveDestAll *MoveDestAllSingleton
)

func init() {
	MoveDestAll = newMoveDestAllSingleton()
}

func newMoveDestAllSingleton() (md *MoveDestAllSingleton) {
	md = new(MoveDestAllSingleton)
	md.SimpleDom = *new(SimpleDom)
	md.name = "Move Destinations"
	md.txtValues = make([]string, 11)
	md.values = make([]uint8, 11)

	md.Dish = 0
	md.values[0] = 0
	md.txtValues[0] = "Dish"
	for i := 0; i < 9; i++ {
		md.Flag[i] = uint8(i + 1)
		md.txtValues[i] = fmt.Sprintf("Flag%v", i+1)
		md.values[i+1] = uint8(i) + 1
	}
	md.Deck = 10
	md.txtValues[10] = "Deck"
	md.values[10] = 10
	return md
}

type MPosJoin struct {
	pos   MPos
	moves []MMove
}

func NewMPosJoin(mPos MPos, movesHand map[int][]bat.Move, passPossible bool, moveCardix, moveix int, movePass bool) (mpj *MPosJoin) {
	mpj = new(MPosJoin)
	mpj.pos = mPos
	if len(movesHand) > 0 {
		mpj.moves = make([]MMove, 0, 50)
		for cardix, moves := range movesHand {
			for ix, move := range moves {
				if movePass || (moveCardix != cardix && moveix != ix) {
					mMove := CreateMMove(cardix, move, movePass)
					mpj.moves = append(mpj.moves, mMove)
				}
			}
		}
		if passPossible && !movePass {
			var mPass [4]byte
			mpj.moves = append(mpj.moves, mPass[:])
		}
	}
	return mpj
}
