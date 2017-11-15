package tf

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"io"
)

//PrintGames print games for machine learning.
func PrintGames(bdb *dbhist.Db, writer io.Writer, gameLimit int) {
	no := 0
	var nextKey []byte
	for {
		var dbErr error
		_, nextKey, dbErr = bdb.Search(func(hist *game.Hist) bool {
			if gameLimit > no {
				winner := hist.Winner()
				if winner != pos.NoPlayer {
					battGame := game.NewGame()
					battGame.LoadHist(hist)
					moveWinner, hasNext := battGame.ScrollForward() //initMove
					for ; hasNext; moveWinner, hasNext = battGame.ScrollForward() {
						if hasNext {
							move := battGame.Hist.Moves[battGame.Pos.LastMoveIx+1]
							if move.Mover == winner && move.MoveType == game.MoveTypeAll.Hand {
								viewPos := game.NewViewPos(battGame.Pos, game.ViewAll.Players[winner], moveWinner)
								tfAnas, moveix := CalcTfAnas(viewPos, move)
								if moveix == -1 {
									log.PrintErr(fmt.Errorf("Move: %v was not found in moves :\n%v", move, viewPos.Moves))
								} else {
									err := printMoves(tfAnas, moveix, writer)
									if err != nil {
										log.PrintErr(err)
									}
								}
							}
						}
					}
				}
			}
			no++
			return false
		}, nextKey)

		if dbErr != nil {
			log.PrintErr(dbErr)
			break
		} else if len(nextKey) == 0 || gameLimit > no {
			break
		}
	}
}

//CalcTfAnas calculates machine learning flag analysis for every moves and
// finds the move ix
func CalcTfAnas(viewPos *game.ViewPos, move *game.Move) (tfAnas [][]*fa.TfAna, moveix int) {
	moveix = -1

	for i, simMove := range viewPos.Moves {
		if moveix == -1 && simMove.IsEqual(move) {
			moveix = i
		}
		if simMove.IsScout() {
			tfAnas = append(tfAnas, nil)
		} else {
			simViewPos := simulate(simMove, viewPos)
			posCards := game.NewPosCards(simViewPos.CardPos)
			deck := fa.NewDeck(simViewPos, posCards)
			hands := createHands(posCards, move.Mover)
			drawFirst := move.Mover
			deckDrawNos := deck.DrawNos(drawFirst)
			deckHandTroops := dht.NewCache(deck.Troops(), hands, deckDrawNos) //TODO be faster with dht.CopyWithOutHand
			flags := game.FlagsCreate(posCards, simViewPos.ConePos)
			tfFlagsAnas := make([]*fa.TfAna, 9)
			for flagix, flag := range flags {
				tfFlagsAnas[flagix] = fa.NewTfAnalysis(move.Mover, flag, deckHandTroops, flagix)
			}
			tfAnas = append(tfAnas, tfFlagsAnas)
		}
	}
	return tfAnas, moveix
}
func createHands(posCards game.PosCards, botix int) (hands [2][]card.Troop) {
	for i := 0; i < len(hands); i++ {
		hands[i] = posCards.SortedCards(pos.CardAll.Players[i].Hand).Troops
	}
	return hands
}
func simulate(gameMove *game.Move, viewPos *game.ViewPos) (simViewPos *game.ViewPos) {
	simViewPos = viewPos.Copy()
	for _, move := range gameMove.Moves {
		if move.IsCard() {
			simViewPos.CardPos[move.Index] = pos.Card(move.NewPos)
		}
	}

	return simViewPos
}
func printMoves(tfMoveAnas [][]*fa.TfAna, moveix int, writer io.Writer) (err error) {
	lnDelimiter := []byte("\n")
	delimiter := []byte(",")
	one := []byte("1")
	zero := []byte("0")
	for ix, tfFlagsAna := range tfMoveAnas {
		if moveix != ix && tfFlagsAna != nil {
			printFlagsAna(tfMoveAnas[moveix], delimiter, writer)
			printFlagsAna(tfFlagsAna, delimiter, writer)
			_, _ = writer.Write(one)
			_, _ = writer.Write(lnDelimiter)
			printFlagsAna(tfFlagsAna, delimiter, writer)
			printFlagsAna(tfMoveAnas[moveix], delimiter, writer)
			_, _ = writer.Write(zero)
			_, _ = writer.Write(lnDelimiter)
		}
	}
	return err
}
func printFlagsAna(tfFlagsAna []*fa.TfAna, delimiter []byte, writer io.Writer) {
	for _, tfFlagAna := range tfFlagsAna {
		_, _ = writer.Write([]byte(tfFlagAna.MachineFormat()))
		_, _ = writer.Write(delimiter)
	}
}
