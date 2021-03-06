package prob

import (
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
)

//Sim the general information need to simulate a move.
type Sim struct {
	deckHandTroops *dht.Cache
	botix          int
	posCards       [][]card.Card
	conePos        [10]pos.Cone
	Moves
}

//Move simulates a move and analyze the involved flags.
//a move may effect two flags a in-flag and a out-flag.
//The flags may be the same incase of a traitor move to the same flag,
//And one flag may be missing.
//Hand to flag move only in-flag.
//Desserter only out-flag.
//Traitor have both and the may be the same.
//Redeploy may have two or only a out-flag.
func (sim *Sim) Move(move *game.Move) (outFlagAna, inFlagAna *fa.Analysis) {
	//TODO maybe check Only works for tac moves as deckHandTroops is not changes when tac is used
	//or if dht.Sim() is made use it even if it have no effect just incase.
	_, simOutFlag, simInFlag, outFlagix, inFlagix := simHandMove(sim.posCards, sim.conePos, move)
	if inFlagix != -1 {
		inFlagAna = fa.NewAnalysis(sim.botix, simInFlag, sim.deckHandTroops, inFlagix)
		if inFlagix == outFlagix {
			outFlagAna = inFlagAna
		}
	}
	if outFlagix != -1 && outFlagAna == nil {
		outFlagAna = fa.NewAnalysis(sim.botix, simOutFlag, sim.deckHandTroops, outFlagix)
	}
	return outFlagAna, inFlagAna
}

func simHandMove(
	posCards [][]card.Card,
	conePos [10]pos.Cone,
	gameMove *game.Move) (hand *card.Cards, outFlag, inFlag *game.Flag, outFlagix, inFlagix int) {
	trimFlagix := -1
	outFlagix = -1
	inFlagix = -1
	if gameMove.MoveType != game.MoveTypeAll.Hand {
		panic("Only hand move is possible to simulate")
	} else {
		moves := gameMove.Moves
		simPosCards := make([][]card.Card, len(posCards))
		for posix, cards := range posCards {
			for _, move := range moves {
				if move.NewPos == uint8(posix) {
					simCards := simPosCards[posix]
					if simCards == nil {
						simCards = make([]card.Card, len(cards))
						copy(simCards, cards)
					}
					simCards = append(simCards, card.Card(move.Index))
					simPosCards[posix] = simCards
				} else if move.OldPos == uint8(posix) {
					simCards := make([]card.Card, 0, len(cards)-1)
					for _, oldCard := range posCards[posix] {
						if oldCard != card.Card(move.Index) {
							simCards = append(simCards, oldCard)
						} else {
							if oldCard == card.TCMud {
								trimFlagix = pos.Card(move.OldPos).Flagix()
							}
						}
					}
					simPosCards[posix] = simCards
				}
			}
			if simPosCards[posix] == nil {
				simPosCards[posix] = cards
			}
		}
		if trimFlagix != -1 {
			simPosCards, _ = mudTrimFlag(simPosCards, trimFlagix)
		}
		for _, move := range moves {
			if pos.Card(move.OldPos).IsOnHand() {
				hand = game.PosCards(simPosCards).SortedCards(pos.Card(move.OldPos))
				flagix := pos.Card(move.NewPos).Flagix()
				if flagix != -1 {
					inFlag = game.NewFlag(flagix, simPosCards, conePos)
					inFlagix = flagix
				}
			} else {
				outix := pos.Card(move.OldPos).Flagix()
				if outix != -1 {
					outFlagix = outix
					outFlag = game.NewFlag(outFlagix, simPosCards, conePos)
					inix := pos.Card(move.NewPos).Flagix()
					if inix != -1 {
						inFlagix = inix
						inFlag = game.NewFlag(inFlagix, simPosCards, conePos)
					}
				}
			}
		}
	}
	return hand, outFlag, inFlag, outFlagix, inFlagix
}
