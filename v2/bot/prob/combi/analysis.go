package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	"github.com/rezder/go-battleline/v2/game/card"
	math "github.com/rezder/go-math/int"
	"sort"
)

//Analysis the result of a combination analysis.
type Analysis struct {
	Comb      *Combination
	Valid     uint64
	All       uint64
	Playables []card.Troop
	Prop      float64
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	txt := fmt.Sprintf("{Rank:%v Prob:%v Formation:%v Strenght:%v Valid:%v All:%v Hand:%v}",
		ana.Comb.Rank, ana.Prop, ana.Comb.Formation.Name, ana.Comb.Strength, ana.Valid, ana.All, ana.Playables)
	return txt
}

//SetAll updates the All combinations field and calculate
//the probabilties.
func (ana *Analysis) SetAll(all uint64) {
	ana.All = all
	if ana.Valid > 0 {
		ana.Prop = float64(ana.Valid) / float64(ana.All)
	}
}

//Ana analyze a combination.
func Ana(
	comb *Combination,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize int,
	isFog bool,
	targetRank, targetHostStr, targetBattStr int,
) (ana *Analysis) {
	nAll := uint64(len(deckHandTroops.SrcDeckTroops))
	dAll := uint64(deckHandTroops.SrcDrawNos[playix])
	if comb.TieBreaker.IsStrenght() {
		if comb.Formation.Value == card.FHost.Value {
			ana = anaHost(comb, flagTroops, flagMorales, deckHandTroops, playix, formationSize, targetRank, targetHostStr, nAll, dAll)
		} else {
			if !isFog {
				ana = anaBattalionStr(comb, flagTroops, flagMorales, deckHandTroops, playix, formationSize, targetBattStr, nAll, dAll)
			} else {
				ana = new(Analysis)
				ana.Comb = comb
			}
		}
	} else {
		ana = new(Analysis)
		ana.Comb = comb
		drawNo := deckHandTroops.SrcDrawNos[playix]
		if !isFog {
			isValid, validFlagTroops, validFlagMorales, color := anaCombiFlagCards(comb.Troops, flagTroops, flagMorales, comb.Formation.Value)
			if !isValid {
				ana.Prop = 0
			} else {
				if len(validFlagTroops) != 0 {
					validHandTroops, validDrawTroops := anaCombiHandDraw(comb.Troops[color], deckHandTroops.SrcHandTroops[playix], deckHandTroops.OnlyDeckSet()) //sorted after combi.Troops
					if formationSize-len(validHandTroops)-len(flagMorales)-len(flagTroops) > drawNo {
						ana.Prop = 0
					} else {
						switch comb.Formation {
						case card.FWedge:
							isValidMorale := moralesReduce(validFlagTroops, validFlagMorales)
							if !isValidMorale {
								ana.Prop = 0
							} else {
								isReduceCalc := false
								isInvalidOnly := false
								if len(validFlagMorales) != 0 {
									flagStrTroops := TroopsToStrenghtTroops(validFlagTroops)
									handStrTroops := TroopsToStrenghtTroops(validHandTroops)
									drawStrTroops := TroopsToStrenghtTroops(validDrawTroops)
									isReduceCalc, isInvalidOnly = anaStraightMorales(validFlagMorales,
										handStrTroops, drawStrTroops, flagStrTroops, formationSize)
									validHandTroops = strenghtTroopsToTroops(handStrTroops) //sorted after map rnd.
									validDrawTroops = strenghtTroopsToTroops(drawStrTroops)
								}
								if !isInvalidOnly {
									anaWedgePhalanx(ana, nAll, dAll, formationSize, validFlagTroops, validFlagMorales, validHandTroops, validDrawTroops)
									//This the rar case of mud,123 joker and troop with strenght 3.
									if isReduceCalc { // 1 out 3 is bad when draw two cards
										if nAll > dAll {
											ana.Valid = ana.Valid - math.Comb(nAll-uint64(3), dAll-uint64(2))
										}
									}
								} else {
									ana.Prop = 0
								}
							}

						case card.FPhalanx:
							anaWedgePhalanx(ana, nAll, dAll, formationSize, validFlagTroops, validFlagMorales, validHandTroops, validDrawTroops)

						case card.FBattalion:
							anaWedgePhalanx(ana, nAll, dAll, formationSize, validFlagTroops, validFlagMorales, validHandTroops, validDrawTroops)

						case card.FSkirmish:
							anaSkirmish(ana, nAll, dAll, formationSize, validFlagMorales, validFlagTroops, validHandTroops, validDrawTroops)

						} //end switch
						ana.Playables = playablesSort(ana.Playables)
					}
				} else { //Only tacs
					// we can handle only one special case, that come up during mud dish but dont call this function with only morales
					if len(flagMorales) == 3 && formationSize == 3 {
						battalion21 := combinations3[24]
						if ana.Comb != battalion21 {
							ana.Prop = 0
						} else {
							ana.Prop = 1
						}
					} else {
						panic(fmt.Sprintf("Illegal argument flag most have troops, this have only moral tactic cards: %v,formationNo: %v", flagMorales, formationSize))
					}
				}
			}
		}
	}
	return ana
}
func anaBattalionStr(
	comb *Combination,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize, targetSum int,
	nAll, dAll uint64,
) (ana *Analysis) {
	ana = new(Analysis)
	ana.Comb = comb
	if targetSum == 0 {
		ana.Prop = 0
	} else {
		flagColorix := FlagColor(flagTroops)
		if flagColorix == card.COLNone {
			ana.Prop = 0
		} else {
			flagStr := moraleTroopsSum(flagTroops, flagMorales)
			missNo := formationSize - len(flagTroops) - len(flagMorales)
			targetRes := deckHandTroops.TargetSum(playix, flagColorix, missNo, targetSum-flagStr)
			if !targetRes.IsPossibel {
				ana.Prop = 0
				sum, isOk := deckHandTroops.Sum(playix, flagColorix, missNo)
				if isOk {
					ana.Playables = battPlayables(sum, missNo, playix, flagColorix, deckHandTroops)
				}
			} else {
				if targetRes.IsMade {
					ana.Prop = 1
					ana.Playables = battPlayables(targetSum-flagStr, missNo, playix, flagColorix, deckHandTroops)
				} else {
					ana.Valid = hostValid(targetRes.ValidDeckTroops, targetRes.ValidHandTroops,
						targetRes.NewNo, targetRes.NewSum, nAll, dAll)
					ana.Playables = battPlayables(targetSum-flagStr, missNo, playix, flagColorix, deckHandTroops)
				}
			}
		}
	}
	return ana
}
func battPlayables(targetSum, missNo, botix, colorix int, deckHandTroops *dht.Cache) (playables []card.Troop) {
	if missNo == 1 {
		for _, handTroop := range deckHandTroops.SrcHandTroops[botix] {
			if handTroop.Color() == colorix && handTroop.Strenght() >= targetSum {
				playables = append(playables, handTroop)
			}
		}
	} else {
		maxStrs, _ := deckHandTroops.Sum(botix, colorix, missNo-1)
		for _, handTroop := range deckHandTroops.SrcHandTroops[botix] {
			if handTroop.Color() == colorix && handTroop.Strenght() >= targetSum-maxStrs {
				playables = append(playables, handTroop)
			}
		}
	}

	return playables
}
func FlagColor(flagTroops []card.Troop) (flagColorix int) {
	flagColorix = card.COLNone
	for _, troop := range flagTroops {
		if flagColorix == card.COLNone {
			flagColorix = troop.Color()
		} else if flagColorix != troop.Color() {
			flagColorix = card.COLNone
			break
		}
	}
	return flagColorix
}
func anaHost(
	comb *Combination,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	deckHandTroops *dht.Cache,
	playix int,
	formationSize, targetRank, targetSum int,
	nAll, dAll uint64,
) (ana *Analysis) {
	ana = new(Analysis)
	ana.Comb = comb
	missNo := formationSize - len(flagTroops) - len(flagMorales)
	var flagStr = moraleTroopsSum(flagTroops, flagMorales)
	targetRes := deckHandTroops.TargetSum(playix, card.COLNone, missNo, targetSum-flagStr)

	if !targetRes.IsPossibel {
		ana.Prop = 0
		sum, isOk := deckHandTroops.Sum(playix, card.COLNone, missNo)
		if isOk {
			ana.Playables = hostPlayables(sum, missNo, playix, deckHandTroops)
		}
	} else {
		if targetRes.IsMade {
			ana.Prop = 1
			ana.Playables = hostPlayables(targetSum-flagStr, missNo, playix, deckHandTroops)
		} else {
			if targetRank == RankHost(formationSize) {
				ana.Valid = hostValid(targetRes.ValidDeckTroops, targetRes.ValidHandTroops,
					targetRes.NewNo, targetRes.NewSum, nAll, dAll)
			} else {
				ana.Valid = math.Comb(nAll, dAll) / 2
			}
			ana.Playables = hostPlayables(targetSum-flagStr, missNo, playix, deckHandTroops)
		}

	}
	return ana
}
func hostValid(
	validDeckTroops, validHandTroops []card.Troop,
	missNo, targetSum int,
	nAll, dAll uint64,
) (valid uint64) {

	nValid := uint64(len(validDeckTroops))
	dValid := uint64(missNo - len(validHandTroops))
	if dValid < uint64(1) {
		dValid = uint64(1)
	}
	if nValid >= dValid {
		drawMaxValid := min(dAll, nValid) //Can't use cards that I can't draw.
		drawMinValue := dValid
		if nValid+dAll > nAll {
			drawMinValue = max(dValid, nValid+dAll-nAll) //Must use all drawn cards.
		}
		for d := drawMaxValid; d >= drawMinValue; d-- {
			if nAll != nValid {
				valid = valid + (hostProbPermCombi(validDeckTroops, validHandTroops, targetSum, missNo, int(d)) *
					math.Comb(nAll-nValid, dAll-d))
			} else {
				valid = valid + hostProbPermCombi(validDeckTroops, validHandTroops, targetSum, missNo, int(d))
			}
		}
	}
	return valid
}
func hostProbPermCombi(drawTroops, handTroops []card.Troop, sum, missNO, drawNo int) (validNo uint64) {
	if len(drawTroops) == 1 {
		validNo = 1
	} else {
		worstDraw := drawTroops[len(drawTroops)-drawNo:]
		if missNO > 1 && hostMaxStrenght(worstDraw, handTroops, missNO) < sum {
			math.Perm(len(drawTroops), drawNo, func(ixs []int) bool {
				troops := make([]card.Troop, len(ixs))
				sort.Ints(ixs) //ixs is allready sorted but I do not want that dependcy
				for i, ix := range ixs {
					troops[i] = drawTroops[ix]
					if i == missNO-1 {
						break
					}
				}
				if hostMaxStrenght(troops, handTroops, missNO) >= sum {
					validNo++
				}
				return false
			})
		} else {
			validNo = math.Comb(uint64(len(drawTroops)), uint64(drawNo))
		}
	}
	return validNo
}
func hostMaxStrenght(troops, handTroops []card.Troop, noTroops int) (maxStr int) {
	handTrs := handTroops[:]
	for i := 0; i < noTroops; i++ {
		if len(troops) > 0 && len(handTrs) > 0 {
			if handTrs[0].Strenght() >= troops[0].Strenght() {
				maxStr = maxStr + handTrs[0].Strenght()
				handTrs = handTrs[1:]
			} else {
				maxStr = maxStr + troops[0].Strenght()
				troops = troops[1:]
			}
		} else if len(handTrs) > 0 {
			maxStr = maxStr + handTrs[0].Strenght()
			handTrs = handTrs[1:]
		} else if len(troops) > 0 {
			maxStr = maxStr + troops[0].Strenght()
			troops = troops[1:]
		} else {
			panic("This should not happen there should be ennough troops")
		}
	}
	return maxStr
}
func hostPlayables(targetSum, missNo, botix int, deckHandTroops *dht.Cache) (playables []card.Troop) {
	if missNo == 1 {
		for _, handTroop := range deckHandTroops.SrcHandTroops[botix] {
			if handTroop.Strenght() >= targetSum {
				playables = append(playables, handTroop)
			} else {
				break
			}
		}
	} else {
		maxStrs, _ := deckHandTroops.Sum(botix, card.COLNone, missNo-1)
		for _, handTroop := range deckHandTroops.SrcHandTroops[botix] {
			if handTroop.Strenght() >= targetSum-maxStrs {
				playables = append(playables, handTroop)
			} else {
				break
			}
		}
	}

	return playables
}

//moraleTroopsSum sums the strenght of troops and morales.
//Morales use the maxStrenght.
//TODO move to card and delete the same function in fa.
func moraleTroopsSum(flagTroops []card.Troop, flagMorales []card.Morale) (sum int) {
	for _, troop := range flagTroops {
		sum = sum + troop.Strenght()
	}
	for _, morale := range flagMorales {
		sum = sum + morale.MaxStrenght()
	}
	return sum
}
func playablesSort(troops []card.Troop) (sortTroops []card.Troop) {
	if len(troops) > 1 {
		sortTroops = make([]card.Troop, 0, len(troops))
		for _, troop := range troops {
			sortTroops = troop.AppendStrSorted(sortTroops)
		}
	} else {
		sortTroops = troops
	}
	return sortTroops
}

//anaSkirmish analyze a skirmish combination 1,2,3,4 in any colors.
//The wedge code for morales can be reused if colapsing the cards of the same strenght.
//The colapsing transformation of the data is very helpfull, as it makes it easy to
//treat cards of same strenght in the same way. As soon as a card(strenght) is used any other
//cards of the same strenghts is useless. The reduceCalc problem is a matter of checking if
//only cards of the jokers strenghts are drawn.
//#ana
func anaSkirmish(ana *Analysis, nAll, dAll uint64, formationNo int,
	flagMorales map[card.Morale]map[int]bool, flagTroops, handTroops, drawTroops []card.Troop) {
	validMorale := moralesReduce(flagTroops, flagMorales)
	if !validMorale {
		ana.Prop = 0
	} else {
		flagStrTroops := TroopsToStrenghtTroops(flagTroops)
		if len(flagStrTroops) != len(flagTroops) {
			ana.Prop = 0
		} else {
			handStrTroops := TroopsToStrenghtTroops(handTroops)
			drawStrTroops := TroopsToStrenghtTroops(drawTroops)
			for v := range flagStrTroops {
				delete(handStrTroops, v)
				delete(drawStrTroops, v)
			}
			reduceCalc := false
			invalidOnly := false

			if len(flagMorales) != 0 {
				reduceCalc, invalidOnly = anaStraightMorales(flagMorales,
					handStrTroops, drawStrTroops, flagStrTroops, formationNo)
			}
			if len(handStrTroops) != 0 {
				for v := range handStrTroops {
					delete(drawStrTroops, v)
				}
			}
			if !invalidOnly {
				missingNo := formationNo - (len(flagStrTroops) + len(flagMorales) + len(handStrTroops))
				ana.Playables = handTroops //not sorted (sorted after combi)
				if missingNo <= 0 {
					ana.Prop = 1
				} else {
					if len(drawStrTroops) >= missingNo {
						drawTroops = strenghtTroopsToTroops(drawStrTroops)
						nValid := uint64(len(drawTroops))
						dValid := uint64(missingNo)
						drawStrs := troopsToStrenghts([][]card.Troop{drawTroops})
						var reducedMoraleStrs map[int]bool
						if reduceCalc {
							reducedMoraleStrs = make(map[int]bool)
							for _, flagMoraleStrs := range flagMorales {
								if len(flagMoraleStrs) == 2 {
									for str := range flagMoraleStrs {
										reducedMoraleStrs[str] = true
									}
									break
								}
							}
						}
						drawMaxValid := min(dAll, nValid) //Can't use cards that I can't draw.
						drawMinValid := dValid
						if nValid+dAll > nAll {
							drawMinValid = max(dValid, nValid+dAll-nAll) //Must use all drawn cards.
						}
						for d := drawMaxValid; d >= drawMinValid; d-- {
							ana.Valid = ana.Valid + (anaSkirmishCombi(drawStrs, missingNo, int(d), reducedMoraleStrs) *
								math.Comb(nAll-nValid, dAll-d))
						}
					}
				}
			} else {
				ana.Prop = 0
			}
		}
	}
}
func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}
func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

//anaSkirmishCombi calculate the valid n over d combinations of valid skirmish cards buy looking at the distinct
//drawn strenghts. When the number of distinct strenghts equal the number of cards we need to make the combination,
//the permutation of the good cards is valid unless it is the two joker strenght from the 123 morale.
func anaSkirmishCombi(drawStrenghts []int, missingNo, d int, moralStrenghts map[int]bool) (validNo uint64) {
	validNo = 0
	math.Perm(len(drawStrenghts), d, func(perm []int) bool {
		var unique [11]bool //for speed maps are expensive
		uniqueNo := 0
		for _, ix := range perm {
			if !unique[drawStrenghts[ix]] {
				uniqueNo = uniqueNo + 1
				unique[drawStrenghts[ix]] = true
			}
			if uniqueNo >= missingNo {
				if len(moralStrenghts) == missingNo {
					equal := true
					for v := range unique {
						if !moralStrenghts[v] {
							equal = false
							break
						}
					}
					if !equal {
						validNo++
						break
					}
				} else {
					validNo++
					break
				}
			}
		}

		return false
	})
	return validNo
}

//troopsToStrenghts combins the strenghts of troops from any lists of troops.
func troopsToStrenghts(troops [][]card.Troop) (strenghts []int) {
	strenghts = make([]int, 0, 10)
	for _, ts := range troops {
		for _, t := range ts {
			strenghts = append(strenghts, t.Strenght())
		}
	}
	return strenghts
}

//anaStraightMorales handles the moral tactic cards on the flag.
//It tries to deduce the moral strenght if possible and reduce
//the problem accordingly.
//The problem is the limmited moral tactic cards 123 or 8.
//Example if the flag have 8 strenght the moral 8 makes the combination
//impossible. The 123 can be equal limmited depending on the combination.
//In the combination 3,4,5 it can only be one strenght.
//Depending on the hand strenghts we may figure out the moral strenght.
//If the moral card can only be one strenght we can remove that strenght
//from the hand and draw.
//If the moral card can be more strenghts than we need then it does
//not matter as moral card just takes the missing strenght.
//The only real problem a 4 card combination one of the 1,2,3 cards
//together with 123 and no cards on the hand in that case if we
//draw the two cards that the moral can be we do not have a combination.
//So is the draw only contain those two strenghts invalidOnly is true.
//If all three remain we must correct for the invalid ones. For a wedge
//Combination it is easy 1 out three (3 over 2) is invalid.
//#flagMorales
//#handStrTroops
//#drawStrTroops
func anaStraightMorales(
	flagMorales map[card.Morale]map[int]bool,
	handStrTroops, drawStrTroops, flagStrTroops map[int][]card.Troop,
	formationNo int) (isReduceCalc bool, isInvalidOnly bool) {
	for _, moraleStrs := range flagMorales {
		if len(moraleStrs) == 2 { //This is only possible for 123 and Leaders
			missingNo := formationNo - (len(flagStrTroops) + len(flagMorales))
			if missingNo >= len(moraleStrs) { //Only 2 = 2 is possible for now, because of one flag troop. This only possible for 123 morale card again because of one flag troop
				handMoraleStrs := make([]int, 0, 0)
				for s := range moraleStrs {
					_, found := handStrTroops[s]
					if found {
						handMoraleStrs = append(handMoraleStrs, s)
					}
				}
				switch len(handMoraleStrs) {
				case 1: //if one we know the moral strenght.
					fallthrough
				case 2: //if two we just pick one and we know the moral strenght.
					delete(moraleStrs, handMoraleStrs[0])
				case 0: //if zero invalid draws is possible from deck
					isInvalidOnly = true
					if len(drawStrTroops) > 1 {
						for s := range drawStrTroops {
							_, found := moraleStrs[s] //if draw contain a none joker
							if !found {
								isInvalidOnly = false
								break
							}
						}
						if len(drawStrTroops) == 3 { // one out 3 is bad when draw two cards
							isReduceCalc = true
						}
					}
				}
			}
		}
		if len(moraleStrs) == 1 { //Here can be only moral 123 or 8
			for s := range moraleStrs {
				delete(handStrTroops, s)
				delete(drawStrTroops, s)
			}
		}
	}
	return isReduceCalc, isInvalidOnly
}

//strenghtTroopsToTroops converts back to troopis.
func strenghtTroopsToTroops(strTroops map[int][]card.Troop) (allTroops []card.Troop) {
	allTroops = make([]card.Troop, 0, len(strTroops))
	for _, troops := range strTroops {
		allTroops = append(allTroops, troops...)
	}
	return allTroops
}

//TroopsToStrenghtTroops colapses troopixs on strenghts.
func TroopsToStrenghtTroops(troops []card.Troop) (strTroops map[int][]card.Troop) {
	strTroops = make(map[int][]card.Troop)
	for _, troop := range troops {
		mapTroops, found := strTroops[troop.Strenght()]
		if !found {
			mapTroops = make([]card.Troop, 0, 1)
		}
		mapTroops = append(mapTroops, troop)
		strTroops[troop.Strenght()] = mapTroops
	}
	return strTroops
}

//moralesReduce removes morales with same strenghts a troopixs.
//#morales
func moralesReduce(troops []card.Troop, morales map[card.Morale]map[int]bool) (isValid bool) {
	isValid = true
	if len(morales) != 0 {
		for _, troop := range troops {
			for _, moral := range morales {
				delete(moral, troop.Strenght())
				if len(moral) == 0 {
					isValid = false
					break
				}
			}
		}
	}
	return isValid
}

//anaBattalionPerm update analysis with the valid number of combinations and probability.
func anaBattalionPerm(
	handTroops, drawTroops []card.Troop,
	sum, elementNo int,
	nAll, dAll uint64) (valid uint64) {
	nValid := uint64(len(drawTroops))
	dValid := uint64(elementNo - len(handTroops))
	if dValid < uint64(1) {
		dValid = uint64(1)
	}
	if nValid >= dValid {
		drawMaxValid := min(dAll, nValid)
		drawMinValid := dValid
		if nValid+dAll > nAll {
			//This very rare: All morale on opponent hand and only good cards in deck
			drawMinValid = max(dValid, nValid+dAll-nAll) //Must use all drawn cards.
		}
		for d := drawMaxValid; d >= drawMinValid; d-- {
			valid = valid + (anaBattalionPermCombi(drawTroops, handTroops, sum, elementNo, int(d)) *
				math.Comb(nAll-nValid, dAll-d))
		}
	}
	return valid
}

//anaBattalionPermCombi calculates the valid combinations for a draw.
func anaBattalionPermCombi(drawTroops, handTroops []card.Troop,
	sum, elementNo, d int) (validNo uint64) {
	validNo = 1
	if len(drawTroops) > 1 {
		validNo = 0
		handStrs := troopsToStrenghts([][]card.Troop{handTroops})
		drawStrs := troopsToStrenghts([][]card.Troop{drawTroops})
		math.Perm(len(drawTroops), d, func(perm []int) bool {
			strenghts := make([]int, d+len(handStrs))
			for i, drawix := range perm {
				strenghts[i] = drawStrs[drawix]
			}
			if len(handTroops) != 0 {
				copy(strenghts[d:], handStrs)
			}
			factors := math.FactorSum(strenghts, sum, elementNo, false)
			if len(factors) != 0 {
				validNo++
			}
			return false
		})
	}
	return validNo
}

//anaWedgePhalanx update the analysis with the valid numbers of combination and the
//probability for a wedge or a phalanx combination.
//The idea is calculate all possible combinations with the valid cards "nValid"
//and multiply with all the possible combination of the none valid cards.
//Every number of valid cards is calulted seperately if 5 valid card exist and we need
//two cards to make a combination then we calulate 5,4,3,2 and there coresponding
//non valid combinations if draw 10 cards dAll that will be 5,6,7 and 8.
//The restriction min() and max() make sure we always make a combination that adds up
//to all our drawn cards.
//#ana
func anaWedgePhalanx(
	ana *Analysis,
	nAll, dAll uint64,
	formationNo int,
	flagTroops []card.Troop,
	flagMorales map[card.Morale]map[int]bool,
	handTroops, drawTroops []card.Troop) {

	missingNo := formationNo - (len(flagTroops) + len(flagMorales) + len(handTroops))
	ana.Playables = handTroops // sorted after map or combi
	if missingNo <= 0 {
		ana.Prop = 1
	} else {
		nValid := uint64(len(drawTroops))
		dValid := uint64(missingNo)
		if nValid >= dValid {
			drawMaxValid := min(dAll, nValid) //Can't use cards that I can't draw.
			drawMinValid := dValid
			if nValid+dAll > nAll {
				drawMinValid = max(dValid, nValid+dAll-nAll) //Must use all drawn cards.
			}
			for d := drawMaxValid; d >= drawMinValid; d-- {
				ana.Valid = ana.Valid + (math.Comb(nValid, d) * math.Comb(nAll-nValid, dAll-d))
			}
		}
	}
}

//anaCombiHandDraw reduces handTroopixs and drawTroopixs to only valid troopixs.
func anaCombiHandDraw(
	combiTroops, handTroops []card.Troop,
	drawSet map[card.Troop]bool) (validHandTroops, drawTroops []card.Troop) {
	for _, troop := range combiTroops {
		for _, handTroop := range handTroops {
			if troop == handTroop {
				validHandTroops = append(validHandTroops, handTroop)
				break
			}
		}
		if drawSet[troop] {
			drawTroops = append(drawTroops, troop)
		}
	}
	return validHandTroops, drawTroops
}

//anaCombiFlagCards finds the valid troops of the combination given the flag cards.
//if the flag have at least one troop.
//valid: True if a valid combination may exit.
//validTroops: The valid troops.
//validMorales: The possible strenght for moral tactic cards.
//color: The color of the combination.
func anaCombiFlagCards(
	combiTroops map[int][]card.Troop,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	formation int) (
	isValid bool,
	validTroops []card.Troop,
	validMorales map[card.Morale]map[int]bool,
	color int) {

	validMorales = make(map[card.Morale]map[int]bool)
	isValid = true
	color = card.COLNone
	colorFormation := false
	if len(combiTroops[card.COLNone]) == 0 {
		colorFormation = true
	}
	for _, troop := range flagTroops {
		isFound := false
		if colorFormation && color == card.COLNone {
			color = troop.Color()
		}
		for _, v := range combiTroops[color] {
			if v == troop {
				isFound = true
				validTroops = append(validTroops, troop)
				break
			}
		}
		if !isFound {
			isValid = false
			break
		}
	}
	if isValid {
		for _, moral := range flagMorales {
			isFound := false
			if formation != card.FBattalion.Value {
				isFound = moralesUpd(validMorales, moral, combiTroops[color])
			} else {
				moralesSetStrenght(validMorales, moral, moral.MaxStrenght())
				isFound = true
			}
			if !isFound {
				isValid = false
				break
			}
		}
	}

	return isValid, validTroops, validMorales, color
}

//moralesUpd set a moral tactic card
//#validMorales
func moralesUpd(validMorales map[card.Morale]map[int]bool, moral card.Morale, combiTroops []card.Troop) (isFound bool) {
	for _, troop := range combiTroops {
		ok := false
		if moral.IsLeader() {
			ok = true
		} else {
			ok = moral.ValidStrenght(troop.Strenght())
		}
		if ok {
			isFound = true
			moralesSetStrenght(validMorales, moral, troop.Strenght())
		}
	}
	return isFound
}

//moralesSetStrenght sets a strenght.
//#m
func moralesSetStrenght(m map[card.Morale]map[int]bool, morale card.Morale, troopStrenght int) {
	strenghtSet, ok := m[morale]
	if !ok {
		strenghtSet = make(map[int]bool)
		m[morale] = strenghtSet
	}
	strenghtSet[troopStrenght] = true
}
