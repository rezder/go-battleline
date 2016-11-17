package gamepos

import (
	"fmt"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battbot/flag"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-error/cerrors"
	"log"
	"sort"
	"strconv"
)

//lostFlagTacticDbFlagMove makes a traitor or redeploy move.
//Simulate all possible moves for a lost flag. Prioritize them
//and play the best if it is a win win move or if it is a prevent loss.
func lostFlagTacticDbFlagMove(
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int,
	tacix int,
	handMoves map[string][]bat.Move) (cardix int, move bat.Move) {
	dbMoves := handMoves[strconv.Itoa(tacix)]
	if len(dbMoves) > 0 {
		simHandTroopixs := make([]int, len(handTroopixs)+1)
		copy(simHandTroopixs, handTroopixs)
		for flagix, flagAna := range flagsAna {
			if flagAna.IsLost {
				dbFlagAnas := dbFlagSimulateMoves(dbMoves, tacix, flagix, flagsAna, handTroopixs, deck, deckMaxValues)
				if dbFlagAnas.Len() > 0 {
					sort.Sort(dbFlagAnas)
					if cerrors.LogLevel() == cerrors.LOG_Debug {
						tac, _ := cards.DrTactic(tacix)
						log.Printf("Double flag. Lost flag: %v, Tactic: %v, Analysis simulated Moves: %+v\n", flagix, tac.Name(), dbFlagAnas)
					}
					if dbFlagAnas[dbFlagAnas.Len()-1].isWinWin() || flagAna.IsLoosingGame {
						move = dbFlagAnas[dbFlagAnas.Len()-1].move
						cardix = cards.TCTraitor
					}
				}

			}

		}
	}
	return cardix, move
}
func dbFlagSimulateMoves(
	moves []bat.Move,
	tacix int,
	lostFlagix int,
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (ta dbFlagAnas) {

	ta = make([]*dbFlagAna, 0, 0)
	for _, move := range moves {
		var outFlagix, inFlagix, cardix int
		switch dbMove := move.(type) {
		case bat.MoveTraitor:
			outFlagix = dbMove.OutFlag
			inFlagix = dbMove.InFlag
			cardix = dbMove.OutCard
		case bat.MoveRedeploy:
			outFlagix = dbMove.OutFlag
			inFlagix = dbMove.InFlag
			cardix = dbMove.OutCard
		default:
			panic("Ilegal move")
		}
		if outFlagix == lostFlagix || inFlagix == lostFlagix {
			lostSimAna, collateralSimAna := dbFlagSimMove(outFlagix, inFlagix, lostFlagix, cardix, tacix, flagsAna, handTroopixs, deck, deckMaxValues)
			var oldCollateralSimAna *flag.Analysis
			if collateralSimAna != nil { //Redeploy to dish
				oldCollateralSimAna = flagsAna[collateralSimAna.Flagix]
			}
			ta = append(ta, newDbFlagAna(move, lostSimAna, flagsAna[lostFlagix], collateralSimAna, oldCollateralSimAna))
		}
	}
	return ta
}
func dbFlagSimMove(
	outFlagix, inFlagix, lostFlagix, cardix, tacix int,
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (lostFlagSimAna, collFlagSimAna *flag.Analysis) {

	lostFlagSim := flagsAna[lostFlagix].Flag.Copy()
	dbFlagSimFlagUpd(outFlagix == lostFlagix, cardix, tacix, lostFlagSim)
	lostFlagSimAna = flag.NewAnalysis(lostFlagSim, handTroopixs, deckMaxValues, deck, lostFlagix, true)

	if inFlagix != bat.REDeployDishix {
		collix := inFlagix
		if outFlagix != lostFlagix {
			collix = outFlagix
		}
		collFlagSim := flagsAna[collix].Flag.Copy()
		dbFlagSimFlagUpd(outFlagix != lostFlagix, cardix, tacix, collFlagSim)
		collFlagSimAna = flag.NewAnalysis(collFlagSim, handTroopixs, deckMaxValues, deck, collix, true)
	}
	return lostFlagSimAna, collFlagSimAna
}
func dbFlagSimFlagUpd(isOutFlag bool, cardix, tacix int, simFlag *flag.Flag) {
	if isOutFlag {
		if tacix == cards.TCRedeploy {
			simFlag.PlayRemoveCardix(cardix)
		} else {
			simFlag.OppRemoveCardix(cardix)
		}
	} else {
		simFlag.PlayAddCardix(cardix)
	}
}

type dbFlagAnas []*dbFlagAna

//Len returns length.
func (ta dbFlagAnas) Len() int {
	return len(ta)
}

//Less compare two double flag moves and returns true if first(i) is worse than second(j).
func (ta dbFlagAnas) Less(i, j int) bool {
	return ta[i].compare(ta[j]) == -1
}

//Swap swaps
func (ta dbFlagAnas) Swap(i, j int) {
	tj := ta[j]
	ta[j] = ta[i]
	ta[i] = tj
}

type dbFlagAna struct {
	move bat.Move
	lost *dbFlagAnaSim
	coll *dbFlagAnaSim
}

func (dba *dbFlagAna) String() string {
	if dba == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Move:%v lost:%v coll:%v}", dba.move, dba.lost, dba.coll)
}

type dbFlagAnaSim struct {
	isLosingGame bool
	isWin        bool
	winProb      float64
	phalanxProb  float64
	oldAna       *flag.Analysis
	ana          *flag.Analysis
	flagValue    int
	botTroopNo   int
}

func (dba *dbFlagAnaSim) String() string {
	if dba == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Win:%v LoseGame:%v WinProb:%v PhalanxWin:%v Ana: %v FlagValue:%v}",
		dba.isWin, dba.isLosingGame, dba.winProb, dba.ana, dba.oldAna, dba.flagValue)
}
func newDbFlagAnaSim(oldAna, ana *flag.Analysis) (dba *dbFlagAnaSim) {
	dba = new(dbFlagAnaSim)
	dba.oldAna = oldAna
	dba.ana = ana
	dba.isWin = ana.IsWin()
	dba.winProb, dba.phalanxProb = dbFlagCalcImprovProb(oldAna, ana)
	dba.botTroopNo = len(ana.Flag.PlayTroops)
	dba.flagValue = oldAna.FlagValue
	dba.isLosingGame = oldAna.IsLoosingGame
	return dba
}

func newDbFlagAna(move bat.Move, lost, oldLost, collateral, oldCollateral *flag.Analysis) (ta *dbFlagAna) {
	ta = new(dbFlagAna)
	ta.move = move
	ta.lost = newDbFlagAnaSim(oldLost, lost)
	if collateral != nil {
		ta.coll = newDbFlagAnaSim(oldCollateral, collateral)
	}
	return ta
}

//compare compares two double flag moves and returns true if this is worse than the other.
//Wining both the lost flag and the collateral flag is of course the best outcome a Win Win.
//Next comes wining the lost flag and improving the collateral flag a Win Improve and so on.
//1 Win win tiebreaker sum of flags value.
//2 Win improve collateral tiebreaker best collateral.
//3 Win lost flag is also in flag no tiebreaker just pick one.
//4 Win worse collateral tiebreaker best collateral
//5 Best lost.
//WARNING all flags may be new flag.
//Bets equal to most improved wining probability as we are simulating
//and impossible move.
func (dba *dbFlagAna) compare(other *dbFlagAna) (comp int) {
	comp = dbFlagLessCompare(dba.isWinWin(), other.isWinWin(), func() bool {
		return dba.coll.ana.FlagValue < other.coll.ana.FlagValue
	})
	if comp != 0 {
		return comp
	}

	taIsWinImprove := dba.is2Flag() && dba.lost.isWin &&
		(dba.coll.winProb > 0 || dba.coll.phalanxProb > 0 && dba.coll.winProb >= 0)
	otIsWinImprove := other.is2Flag() && other.lost.isWin &&
		(other.coll.winProb > 0 || other.coll.phalanxProb > 0 && other.coll.winProb >= 0)
	comp = dbFlagLessCompare(taIsWinImprove, otIsWinImprove, func() bool {
		return dbFlagLessCompareProb(dba.coll, other.coll)
	})
	if comp != 0 {
		return comp
	}

	taIsWin := !dba.is2Flag() && dba.lost.isWin
	otIswin := !dba.is2Flag() && other.lost.isWin
	comp = dbFlagLessCompare(taIsWin, otIswin, func() bool {
		return true
	})
	if comp != 0 {
		return comp
	}

	taIsWinWorse := dba.is2Flag() && dba.lost.isWin && !dba.isCollLoseGame()
	otIsWinWorse := other.is2Flag() && other.lost.isWin && !other.isCollLoseGame()
	comp = dbFlagLessCompare(taIsWinWorse, otIsWinWorse, func() bool {
		return dbFlagLessCompareProb(dba.coll, other.coll)
	})
	if comp != 0 {
		return comp

	}

	taIsNotLost := !dba.lost.ana.IsLost && !dba.isCollLoseGame()
	otIsNotLost := !other.lost.ana.IsLost && !other.isCollLoseGame()
	comp = dbFlagLessCompare(!taIsNotLost, otIsNotLost, func() bool {
		return dbFlagLessCompareProb(dba.lost, other.lost)
	})

	return comp
}

func dbFlagLessCompareProb(dba, other *dbFlagAnaSim) bool {
	if dba.winProb < other.winProb {
		return true
	} else if dba.winProb == other.winProb && dba.phalanxProb < other.phalanxProb {
		return true
	} else if dba.winProb == other.winProb && dba.phalanxProb == other.phalanxProb &&
		dba.botTroopNo < other.botTroopNo {
		return true
	} else if dba.winProb == other.winProb && dba.phalanxProb == other.phalanxProb &&
		dba.botTroopNo == other.botTroopNo && dba.flagValue < other.flagValue {
		return true
	}

	return false
}
func dbFlagLessCompare(lessCon, bigCon bool, tieBreak func() bool) int {
	if lessCon || bigCon {
		if !lessCon && bigCon {
			return -1
		}
		if lessCon && bigCon {
			if tieBreak() {
				return -1
			}
		}
		return 1
	}
	return 0
}

func (dba *dbFlagAna) isWinWin() bool {
	return dba.is2Flag() && dba.coll.isWin && dba.lost.isWin
}
func (dba *dbFlagAna) is2Flag() bool {
	return dba.coll != nil
}
func (dba *dbFlagAna) isCollLoseGame() bool {
	return dba.coll != nil && dba.coll.ana.IsLost && dba.coll.isLosingGame
}

func dbFlagCalcImprovProb(oldAna, ana *flag.Analysis) (diffWinProb, diffPhalanxProb float64) {
	oldWinProb, oldPhalanxprob := flagAnaProb(oldAna)
	winProb, phalanxprob := flagAnaProb(ana)
	diffWinProb = winProb - oldWinProb
	diffPhalanxProb = phalanxprob - oldPhalanxprob
	return diffWinProb, diffPhalanxProb
}
func flagAnaProb(flagAna *flag.Analysis) (win, phalanx float64) {
	//fmt.Printf("Db flagAna %+v", flagAna)
	if flagAna.Analysis != nil && !flagAna.IsFog && !flagAna.IsNewFlag {
		if flagAna.IsTargetRanked {
			targetRank := flagAna.TargetRank
			if flagAna.IsTargetMade {
				targetRank = targetRank - 1
			}
			for _, combiAna := range flagAna.Analysis {
				if combiAna == nil {
					break
				}
				if combiAna.Comb.Rank <= targetRank && combiAna.Prop > 0 {
					win = win + combiAna.Prop
				}
			}
		}

		for _, combiAna := range flagAna.Analysis {
			if combiAna == nil {
				break
			}
			if combiAna.Comb.Formation.Value >= cards.FPhalanx.Value && combiAna.Prop > 0 {
				phalanx = phalanx + combiAna.Prop
			}
		}
	}
	return win, phalanx
}
