package prob

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"sort"
)

//lostFlagTacDbFlag makes a traitor or redeploy move.
//Simulate all possible moves for a lost flag. Prioritize them
//and play the best if it is a win win move or if it is a prevent loss.
func lostFlagTacDbFlag(
	flagsAna map[int]*fa.Analysis,
	priLiveFlagixs []int,
	guile card.Guile,
	sim *Sim,
) (moveix int) {
	moveix = -1
	dbMoves, dbMoveixs := sim.FindDoubleFlags(guile)
	if len(dbMoves) > 0 {
		for _, flagix := range priLiveFlagixs {
			flagAna := flagsAna[flagix]
			if flagAna.IsLost {
				dbFlagAnas := dbFlagSimulateMoves(dbMoves, dbMoveixs, flagix, flagsAna, sim)
				if dbFlagAnas.Len() > 0 {
					bestFlagAna := dbFlagAnas.BestAna()
					logtxt := "Double flag. Lost flag: %v, Tactic: %v, Analysis simulated Moves: %+v\n"
					log.Printf(log.Debug, logtxt, flagix, guile, bestFlagAna)
					if bestFlagAna.isWinWin() || flagAna.IsLoosingGame {
						moveix = bestFlagAna.moveix
					}
				}
			}
		}
	}
	return moveix
}
func dbFlagSimulateMoves(
	moves []*game.Move,
	moveixs []int,
	lostFlagix int,
	flagsAna map[int]*fa.Analysis,
	sim *Sim,
) (ta dbFlagAnas) {

	ta = make([]*dbFlagAna, 0, 0)
	for ix, move := range moves {
		if pos.Card(move.Moves[1].NewPos).Flagix() == lostFlagix || pos.Card(move.Moves[1].OldPos).Flagix() == lostFlagix {
			lostSimAna, collateralSimAna := dbFlagSimMove(lostFlagix, move, sim)
			var oldCollateralSimAna *fa.Analysis
			if collateralSimAna != nil {
				oldCollateralSimAna = flagsAna[collateralSimAna.Flagix]
			}
			ta = append(ta, newDbFlagAna(moveixs[ix], lostSimAna, flagsAna[lostFlagix], collateralSimAna, oldCollateralSimAna))
		}
	}
	return ta
}
func dbFlagSimMove(lostFlagix int, move *game.Move, sim *Sim,
) (lostFlagSimAna, collFlagSimAna *fa.Analysis) {
	outFlagAna, inFlagAna := sim.Move(move)

	outFlagix := pos.Card(move.Moves[1].OldPos).Flagix()
	inFlagix := pos.Card(move.Moves[1].NewPos).Flagix()

	if outFlagix == inFlagix {
		lostFlagSimAna = outFlagAna
	} else {
		if lostFlagix == outFlagix {
			lostFlagSimAna = outFlagAna
			collFlagSimAna = inFlagAna
		} else {
			lostFlagSimAna = inFlagAna
			collFlagSimAna = outFlagAna
		}
	}

	return lostFlagSimAna, collFlagSimAna
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
func (ta dbFlagAnas) BestAna() (bestAna *dbFlagAna) {
	sort.Sort(ta)
	bestFlagAna := ta[ta.Len()-1]
	return bestFlagAna
}

type dbFlagAna struct {
	moveix int
	lost   *dbFlagAnaSim
	coll   *dbFlagAnaSim
}

func (dba *dbFlagAna) String() string {
	if dba == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Move:%v lost:%v coll:%v}", dba.moveix, dba.lost, dba.coll)
}

type dbFlagAnaSim struct {
	isLosingGame bool
	isWin        bool
	winProb      float64
	phalanxProb  float64
	oldAna       *fa.Analysis
	ana          *fa.Analysis
	flagValue    int
	botTroopNo   int
}

func (dba *dbFlagAnaSim) String() string {
	if dba == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Flag:%v Win:%v LoseGame:%v WinProb:%v PhalanxWin: %v FlagValue:%v, TroopsNo:%v}",
		dba.oldAna.Flagix, dba.isWin, dba.isLosingGame, dba.winProb, dba.phalanxProb, dba.flagValue, dba.botTroopNo)
}
func newDbFlagAnaSim(oldAna, ana *fa.Analysis) (dba *dbFlagAnaSim) {
	dba = new(dbFlagAnaSim)
	dba.oldAna = oldAna
	dba.ana = ana
	dba.isWin = ana.IsWin
	dba.winProb, dba.phalanxProb = dbFlagCalcImprovProb(oldAna, ana)
	dba.botTroopNo = ana.Flag.PlayerFormationSize(ana.Playix)
	dba.flagValue = oldAna.FlagValue
	dba.isLosingGame = oldAna.IsLoosingGame
	return dba
}
func (dba *dbFlagAnaSim) equal(other *dbFlagAnaSim) (equal bool) {
	if other == nil && dba == nil {
		equal = true
	} else if other != nil && dba != nil {
		if other == dba {
			equal = true
		} else if other.isLosingGame == dba.isLosingGame &&
			other.isWin == dba.isWin &&
			other.winProb == dba.winProb &&
			other.phalanxProb == dba.phalanxProb &&
			other.flagValue == dba.flagValue &&
			other.botTroopNo == dba.botTroopNo {
			equal = true

		}
	}
	return equal
}

func newDbFlagAna(moveix int, lost, oldLost, collateral, oldCollateral *fa.Analysis) (ta *dbFlagAna) {
	ta = new(dbFlagAna)
	ta.moveix = moveix
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
//3 Win nil no tiebreaker just pick one.
//4 Win worse collateral tiebreaker best collateral
//5 Best lost.
//WARNING all flags may be new flag.
//Bets equal to most improved wining probability as we are simulating
//and impossible move. if best equal then compare coll.
//LIMIT Probabilities is only use full when not dealing with sum.
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
	//No win of lost flag
	if !dba.lost.equal(other.lost) {
		taIsNotLost := !dba.lost.ana.IsLost && !dba.isCollLoseGame()
		otIsNotLost := !other.lost.ana.IsLost && !other.isCollLoseGame()
		comp = dbFlagLessCompare(taIsNotLost, otIsNotLost, func() bool {
			return dbFlagLessCompareProb(dba.lost, other.lost)
		})
		if comp != 0 {
			return comp
		}
		comp = dbFlagLessCompareLost(dba.lost, other.lost)

	} else {
		taCollWin := dba.is2Flag() && dba.coll.isWin
		otCollWin := other.is2Flag() && other.coll.isWin
		comp = dbFlagLessCompare(taCollWin, otCollWin, func() bool {
			return dba.coll.ana.FlagValue < other.coll.ana.FlagValue
		})
		if comp != 0 {
			return comp
		}
		taCollImprove := dba.is2Flag() &&
			(dba.coll.winProb > 0 || dba.coll.phalanxProb > 0 && dba.coll.winProb >= 0)
		otCollImprove := other.is2Flag() &&
			(other.coll.winProb > 0 || other.coll.phalanxProb > 0 && other.coll.winProb >= 0)
		comp = dbFlagLessCompare(taCollImprove, otCollImprove, func() bool {
			return dbFlagLessCompareProb(dba.coll, other.coll)
		})
		if comp != 0 {
			return comp
		}
		comp = dbFlagLessCompare(!dba.is2Flag(), !other.is2Flag(), func() bool {
			return true
		})
		if comp != 0 {
			return comp
		}
		comp = dbFlagLessCompare(!dba.coll.ana.IsLost, !other.coll.ana.IsLost, func() bool {
			return dbFlagLessCompareProb(dba.coll, other.coll)
		})
		if comp != 0 {
			return comp
		}
		comp = dbFlagLessCompareLost(dba.coll, other.coll)
	}
	return comp
}
func dbFlagLessCompareLost(dba, other *dbFlagAnaSim) int {
	if dba.botTroopNo > other.botTroopNo {
		return -1
	} else if dba.botTroopNo < other.botTroopNo {
		return 1
	}
	return 0
}

// dbFlagLessCompareProb compare first using probabilities then troop numbers
// and then flag values. WARNING if all equal less is false.
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

func dbFlagCalcImprovProb(oldAna, ana *fa.Analysis) (diffWinProb, diffPhalanxProb float64) {
	oldWinProb, oldPhalanxprob := flagAnaProb(oldAna)
	winProb, phalanxprob := flagAnaProb(ana)
	diffWinProb = winProb - oldWinProb
	diffPhalanxProb = phalanxprob - oldPhalanxprob
	return diffWinProb, diffPhalanxProb
}
func flagAnaProb(flagAna *fa.Analysis) (win, phalanx float64) {
	if flagAna.RankAnas != nil && !flagAna.IsFog && !flagAna.IsNewFlag {
		targetRank := flagAna.TargetRank
		if flagAna.IsTargetMade && targetRank != combi.HostRank(flagAna.FormationSize) {
			targetRank = targetRank - 1
		}
		for _, combiAna := range flagAna.RankAnas {
			if combiAna == nil {
				break
			}
			if combiAna.Comb.Rank <= targetRank && combiAna.Prop > 0 {
				win = win + combiAna.Prop
			}
		}
		for _, combiAna := range flagAna.RankAnas {
			if combiAna == nil {
				break
			}
			if combiAna.Comb.Formation.Value >= card.FPhalanx.Value && combiAna.Prop > 0 {
				phalanx = phalanx + combiAna.Prop
			}
		}
	}
	return win, phalanx
}
