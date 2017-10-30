package prob

import (
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
)

//Moves game moves.
type Moves []*game.Move

//FindDeck finds a deck move, if not panics.
func (moves Moves) FindDeck(deckPos pos.Card) (moveix int) {
	moveix = -1
	for i, move := range moves {
		if move.Moves[0].OldPos == uint8(deckPos) {
			moveix = i
			break
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return moveix
}

//FindScoutReturn finds a scout return moves, if not panics.
func (moves Moves) FindScoutReturn(tacs []card.Card, troops []card.Troop) (moveix int) {
	moveix = -1
	cardsNo := len(tacs) + len(troops)
	if len(moves) > 0 && moves[0].MoveType == game.MoveTypeAll.ScoutReturn {
	Loop:
		for ix, move := range moves {
			if len(move.Moves) == cardsNo {
				for bpi, bpMove := range move.Moves {
					if bpi < len(tacs) {
						if bpMove.Index != int(tacs[bpi]) {
							continue Loop
						}
					} else {
						if bpMove.Index != int(troops[bpi-len(tacs)]) {
							continue Loop
						}
					}
				}
				moveix = ix
				break
			}
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return moveix
}

//FindCone finds a cone if not panics.
func (moves Moves) FindCone(coneixs []int) (moveix int) {
	moveix = -1
	for i, move := range moves {
		if len(move.Moves) == len(coneixs) {
			moveix = i
			for _, bpMove := range move.Moves {
				isFound := false
				for _, coneix := range coneixs {
					if bpMove.Index == coneix {
						isFound = true
					}
				}
				if !isFound {
					moveix = -1
					break
				}
			}
			if moveix != -1 {
				break
			}
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return moveix
}

//FindScoutMove finds a scout move, if not panics.
func (moves Moves) FindScoutMove(deckPos pos.Card) (moveix int) {
	moveix = -1
	for ix, move := range moves {
		if len(move.Moves) > 0 && move.Moves[0].Index == card.TCScout {
			if move.Moves[1].OldPos == uint8(deckPos) {
				moveix = ix
				break
			}
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return moveix
}

//FindDeserterMove finds a deserter move if not panics.
func (moves Moves) FindDeserterMove(killCard card.Card) (dMove *game.Move, moveix int) {
	moveix = -1
	for ix, move := range moves {
		if len(move.Moves) > 0 && move.Moves[0].Index == card.TCDeserter {
			if move.Moves[1].Index == int(killCard) {
				moveix = ix
				dMove = move
				break
			}
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return dMove, moveix
}

//FindDoubleFlags finds double flag moves redeploy and traitor.
func (moves Moves) FindDoubleFlags(guile card.Guile) (dbMoves []*game.Move, moveixs []int) {
	for ix, move := range moves {
		if len(move.Moves) > 0 && move.Moves[0].Index == int(guile) {
			dbMoves = append(dbMoves, move)
			moveixs = append(moveixs, ix)
		}
	}
	return dbMoves, moveixs
}

//SearchPass finds a pass move.
func (moves Moves) SearchPass() (moveix int) {
	moveix = -1
	for ix, move := range moves {
		if len(move.Moves) == 0 {
			moveix = ix
			break
		}
	}
	return moveix
}

//FindHandFlag finds a hand to flag move, if not panics.
func (moves Moves) FindHandFlag(flagix int, cardMove card.Card) (handFlagMove *game.Move, moveix int) {
	moveix = -1
	for i, move := range moves {
		if len(move.Moves) > 0 {
			bpMove := move.Moves[0]
			if flagix == pos.Card(bpMove.NewPos).Flagix() &&
				bpMove.Index == int(cardMove) &&
				pos.Card(bpMove.OldPos).IsOnHand() {
				moveix = i
				handFlagMove = move
				break
			}
		}
	}
	if moveix == -1 {
		panic("Move should exist") //all find should panic I think
	}
	return handFlagMove, moveix
}
