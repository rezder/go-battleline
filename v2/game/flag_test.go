package game

import (
	"github.com/rezder/go-battleline/v2/game/pos"
	"testing"
)

func TestFlagFailedClaimed(t *testing.T) {
	cardPos := [71]pos.Card{0, 21, 22, 22, 0, 0, 0, 0, 0, 0, 0, 22, 21, 21, 22, 22, 21, 0, 0, 0, 0, 0, 0, 0, 21, 22, 0, 0, 0, 21, 0, 13, 0, 15, 15, 0, 0, 0, 0, 0, 15, 13, 0, 0, 5, 0, 0, 0, 21, 12, 0, 0, 0, 0, 4, 4, 4, 0, 14, 22, 0, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23}
	conePos := [10]pos.Cone{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	posCards := NewPosCards(cardPos)
	deckTroops := posCards.SimDeckTroops()
	flags := FlagsCreate(posCards, conePos)
	for flagix, flag := range flags {
		t.Logf("Flagix:%v, Flag:%v", flagix, flag)
		for playerix := 0; playerix < 2; playerix++ {
			if flag.HasFormation(playerix) {
				isClaimable, exCards := flag.IsClaimable(playerix, deckTroops)
				if isClaimable {
					t.Errorf("Claim should fail for playerix: %v flag:%v, deck:%v", playerix, flag, deckTroops)
				} else {
					if len(exCards) == 0 {
						t.Errorf("Claim should fail for playerix with an example: %v flag:%v, deck:%v", playerix, flag, deckTroops)
					}
				}
			}
		}
	}
}
func TestRemoveFailedClaimed(t *testing.T) {
	cardPos := [71]pos.Card{0, 21, 22, 22, 0, 0, 0, 0, 0, 0, 0, 22, 21, 21, 22, 22, 21, 0, 0, 0, 0, 0, 0, 0, 21, 22, 0, 0, 0, 21, 0, 13, 0, 15, 15, 0, 0, 0, 0, 0, 15, 13, 0, 0, 5, 0, 0, 0, 21, 12, 0, 0, 0, 0, 4, 4, 4, 0, 14, 22, 0, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23}

	conePos := [10]pos.Cone{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	pos := NewPos()
	pos.CardPos = cardPos
	pos.ConePos = conePos
	pos.LastMoveType = MoveTypeAll.Deck
	pos.LastMoveIx = 24
	pos.LastMover = 1
	pos.PlayerReturned = NoPlayer
	move := pos.CalcMoves()[1]
	updMoves, failedClaimExs := removeFailedClaim(move.Moves, move.Mover, pos)
	for flagix, exs := range failedClaimExs {
		if flagix == move.Moves[0].Index-1 {
			if len(exs) == 0 {
				t.Errorf("Claim should fail for move: %v ", move)
			}
		}
	}
	if len(updMoves) == len(move.Moves) {
		t.Error("moves should be updated")
	}

}
func TestRemoveFailedClaimed1(t *testing.T) {
	cardPos := [71]pos.Card{0, 21, 22, 11, 8, 22, 21, 0, 2, 0, 22, 22, 3, 3, 22, 16, 1, 1, 22, 18, 21, 0, 9, 21, 5, 22, 19, 21, 0, 21, 0, 13, 9, 15, 15, 16, 21, 0, 0, 18, 15, 13, 0, 11, 5, 16, 19, 0, 2, 12, 17, 0, 9, 11, 4, 4, 4, 7, 14, 12, 0, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23}

	conePos := [10]pos.Cone{0, 0, 0, 0, 1, 0, 0, 0, 0, 0}
	pos := NewPos()
	pos.CardPos = cardPos
	pos.ConePos = conePos
	pos.LastMoveType = MoveTypeAll.Deck
	pos.LastMoveIx = 87
	pos.LastMover = 1
	pos.PlayerReturned = NoPlayer
	move := pos.CalcMoves()[1]
	updMoves, failedClaimExs := removeFailedClaim(move.Moves, move.Mover, pos)
	for flagix, exs := range failedClaimExs {
		if flagix == move.Moves[0].Index-1 {
			if len(exs) == 0 {
				t.Errorf("Claim should fail for move: %v ", move)
			}
			if len(exs) > 0 && len(exs) < 3 {
				t.Errorf("Example should contain 3 or 4 cards: %v ", exs)
			}
		}
	}
	if len(updMoves) == len(move.Moves) {
		t.Error("moves should be updated")
	}

}
func TestRemoveFailedClaimed2(t *testing.T) {
	cardPos := [71]pos.Card{0, 12, 1, 2, 19, 9, 18, 15, 21, 17, 4, 22, 11, 2, 0, 6, 6, 6, 0, 13, 4, 12, 21, 2, 22, 22, 8, 8, 8, 22, 4, 12, 11, 18, 21, 9, 18, 0, 0, 13, 0, 1, 1, 0, 19, 21, 16, 16, 0, 17, 21, 21, 11, 22, 0, 9, 3, 22, 0, 14, 14, 22, 23, 23, 23, 23, 23, 23, 21, 23, 23}

	conePos := [10]pos.Cone{0, 2, 1, 0, 0, 0, 0, 0, 1, 1}
	pos := NewPos()
	pos.CardPos = cardPos
	pos.ConePos = conePos
	pos.LastMoveType = MoveTypeAll.Deck
	pos.LastMoveIx = 90
	pos.LastMover = 1
	pos.PlayerReturned = NoPlayer
	move := pos.CalcMoves()[1]
	updMoves, failedClaimExs := removeFailedClaim(move.Moves, move.Mover, pos)
	for flagix, exs := range failedClaimExs {
		if flagix == 3 || flagix == 5 {
			if len(exs) == 0 {
				t.Errorf("Claim should fail for move: %v flagix: %v", move, flagix)
			}
			if len(exs) > 0 && len(exs) < 3 {
				t.Errorf("Example should contain 3 or 4 cards: %v ", exs)
			}
		}
	}
	if len(updMoves) == len(move.Moves) {
		t.Error("moves should be updated")
	}

}
func TestRemoveFailedClaimed3(t *testing.T) {
	cardPos := [71]pos.Card{0, 22, 15, 8, 12, 12, 0, 0, 0, 18, 0, 21, 0, 0, 1, 21, 0, 17, 21, 0, 14, 0, 11, 9, 11, 22, 0, 7, 0, 7, 0, 4, 15, 21, 21, 0, 0, 8, 0, 18, 6, 19, 19, 9, 0, 0, 2, 17, 5, 5, 22, 22, 15, 9, 22, 22, 2, 8, 21, 18, 0, 21, 23, 23, 23, 23, 23, 23, 23, 22, 23}
	conePos := [10]pos.Cone{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	pos := NewPos()
	pos.CardPos = cardPos
	pos.ConePos = conePos
	pos.LastMoveType = MoveTypeAll.Deck
	pos.LastMoveIx = 68
	pos.LastMover = 0
	pos.PlayerReturned = NoPlayer
	move := pos.CalcMoves()[1]
	updMoves, failedClaimExs := removeFailedClaim(move.Moves, move.Mover, pos)
	for flagix, exs := range failedClaimExs {
		if flagix == 4 {
			if len(exs) == 0 {
				t.Errorf("Claim should fail for move: %v flagix: %v", move, flagix)
			}
			if len(exs) > 0 && len(exs) < 3 {
				t.Errorf("Example should contain 3 or 4 cards: %v ", exs)
			}
		}
	}
	if len(updMoves) == len(move.Moves) {
		t.Error("moves should be updated")
	}

}
