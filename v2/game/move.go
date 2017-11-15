package game

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/pos"
	"math/rand"
)

const (
	//NOHandInit  the starting numbers of cards.
	NOHandInit = 7
)

var (
	//BoardPieceAll all the board piece types
	//Cone or Card.
	BoardPieceAll = newBoardPieceAllST()
)

// Move a battleline move.
type Move struct {
	Mover int
	MoveType
	Moves []*BoardPieceMove
}

// NewMove create a Move with out any board piece moves.
func NewMove(mover int, moveType MoveType) (move *Move) {
	move = new(Move)
	move.Mover = mover
	move.MoveType = moveType
	return move
}
func (move *Move) String() string {
	return fmt.Sprintf("Move{Mover:%v,MoveType:%v,Moves:%v}", move.Mover, move.MoveType, move.Moves)
}

//GetMoverAndType returns mover and move type.
//Works with nil.
func (move *Move) GetMoverAndType() (int, MoveType) {
	if move == nil {
		return pos.NoPlayer, MoveTypeAll.None
	}
	return move.Mover, move.MoveType
}

// IsEqual checks if two moves are equal.
func (move *Move) IsEqual(o *Move) bool {
	if o == move {
		return true
	}
	if (o == nil && move != nil) || (o != nil && move == nil) {
		return false
	}
	isEqual := false
	if move.Mover == o.Mover &&
		move.MoveType == o.MoveType &&
		len(move.Moves) == len(o.Moves) {
		isEqual = true
		for i, bpMove := range move.Moves {
			if !bpMove.IsEqual(o.Moves[i]) {
				isEqual = false
				break
			}
		}
	}
	return isEqual
}

//Copy makes a copy of move.
func (move *Move) Copy() (copy *Move) {
	if move != nil {
		copy = new(Move)
		copy.Mover = move.Mover
		copy.MoveType = move.MoveType
		if move.Moves != nil {
			copy.Moves = make([]*BoardPieceMove, len(move.Moves))
			for i, refBPMove := range move.Moves {
				bpMove := *refBPMove
				copy.Moves[i] = &bpMove
			}
		}
	}
	return copy
}

// BoardPieceMove a single board piece move.
type BoardPieceMove struct {
	Index  int
	NewPos uint8
	OldPos uint8
	BoardPiece
}

func (b *BoardPieceMove) String() string {
	return fmt.Sprintf("BPMove{Index:%v,NewPos:%v,OldPos:%v}", b.Index, b.NewPos, b.OldPos)
}

//IsEqual checks is to Board Piece Move is equal.
func (b *BoardPieceMove) IsEqual(o *BoardPieceMove) bool {
	if b == o {
		return true
	}
	if (b == nil && o != nil) || (b != nil && o == nil) {
		return false
	}
	if b.Index == o.Index &&
		b.NewPos == o.NewPos &&
		b.OldPos == o.OldPos &&
		b.BoardPiece == o.BoardPiece {
		return true
	}
	return false
}

// BoardPiece the board piece type card or cone.
type BoardPiece uint8

// IsCone checks if the board piece is cone.
func (b BoardPiece) IsCone() bool {
	return b == BoardPieceAll.Cone
}

// IsCard checks if the board piece is a Card.
func (b BoardPiece) IsCard() bool {
	return b == BoardPieceAll.Card
}
func (b BoardPiece) String() string {
	return BoardPieceAll.Names()[int(b)]
}

// BoardPieceAllST the singleton of all the board pieces.
type BoardPieceAllST struct {
	Cone BoardPiece
	Card BoardPiece
}

// newBoardPieceAllST creats the Board piece all singleton
func newBoardPieceAllST() (b BoardPieceAllST) {
	b.Cone = 1
	b.Card = 0
	return b
}

// Names returns all the names of the board pieces Card or Cone.
func (m BoardPieceAllST) Names() []string {
	return []string{"Card", "Cone"}
}

// All return all the board piece Card and Cone
func (m BoardPieceAllST) All() []BoardPiece {
	return []BoardPiece{0, 1}

}

// moveCreateInit creates the first move, deal 7 cards to both players.
func moveCreateInit(mover int) (move *Move) {
	move = new(Move)
	move.Mover = mover
	move.MoveType = MoveTypeAll.Init
	move.Moves = make([]*BoardPieceMove, 0, 14)
	perm := rand.Perm(60)
	for i := 0; i < NOHandInit; i++ {
		move0 := BoardPieceMove{
			BoardPiece: BoardPieceAll.Card,
			Index:      perm[i] + 1,
			OldPos:     uint8(pos.CardAll.DeckTroop),
			NewPos:     uint8(pos.CardAll.Players[opp(mover)].Hand),
		}
		move.Moves = append(move.Moves, &move0)
		move1 := BoardPieceMove{
			BoardPiece: BoardPieceAll.Card,
			Index:      perm[7+i] + 1,
			OldPos:     uint8(pos.CardAll.DeckTroop),
			NewPos:     uint8(pos.CardAll.Players[mover].Hand),
		}
		move.Moves = append(move.Moves, &move1)
	}
	return move
}
