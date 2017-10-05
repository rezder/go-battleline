package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
	math "github.com/rezder/go-math/int"
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

//Ana analyze a combination.
func Ana(
	comb *Combination,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	handTroops []card.Troop,
	drawSet map[card.Troop]bool,
	drawNo int,
	isMud bool) (ana *Analysis) {
	if len(flagTroops) == 0 {
		panic("Flag must have troops")
	}
	formationNo := 3
	if isMud {
		formationNo = 4
	}
	ana = new(Analysis)
	ana.Comb = comb
	nAll := uint64(len(drawSet))
	dAll := uint64(drawNo)
	isValid, validFlagTroops, validFlagMorales, color := anaCombiFlagCards(comb.Troops, flagTroops, flagMorales, comb.Formation.Value)
	if !isValid || formationNo-len(flagTroops)-len(flagMorales) > drawNo {
		ana.Prop = 0
	} else {
		if len(validFlagTroops) != 0 {
			validHandTroops, validDrawTroops := anaCombiHandDraw(comb.Troops[color], handTroops, drawSet) //sorted after combi.Troops
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
							handStrTroops, drawStrTroops, flagStrTroops, formationNo)
						validHandTroops = strenghtTroopsToTroops(handStrTroops) //sorted after map rnd.
						validDrawTroops = strenghtTroopsToTroops(drawStrTroops)
					}
					if !isInvalidOnly {
						anaWedgePhalanx(ana, nAll, dAll, formationNo, validFlagTroops, validFlagMorales, validHandTroops, validDrawTroops)
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
				anaWedgePhalanx(ana, nAll, dAll, formationNo, validFlagTroops, validFlagMorales, validHandTroops, validDrawTroops)

			case card.FBattalion:
				anaBattalion(ana, nAll, dAll, formationNo, comb.Strength, validFlagMorales,
					validFlagTroops, validHandTroops, validDrawTroops)

			case card.FSkirmish:
				anaSkirmish(ana, nAll, dAll, formationNo, validFlagMorales, validFlagTroops, validHandTroops, validDrawTroops)

			} //end switch
			ana.Playables = playablesSort(ana.Playables)
		} else { //Only tacs
			//Should not bee needed
			panic("Illegal argument flag most have troops, this have only moral tactic troops")
		}
	}
	return ana
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

//anaBattalion analyze a Battalion combination same colors.
//Its is the worst because many combination of strenghts can make the
//same sum. So which card is played can influnce the probabilty of making the sum.
//to get around that problem we just assume one will allways play the higest card.
//this may not be the card that create the best probability for the sum. The lowest
//usually is best to play. The idea was that it was the best player strategi, but
//i am not sure, it may just be easier. It is also assumed all moral cards have
//its higest strenght.
//We transfer all cards to its strenght and hopfully keep track of which strenght belong to
//which card. We often bunk strenghts from hand and draw together as any of then can be
//used in the sum.
//#ana
func anaBattalion(ana *Analysis, nAll, dAll uint64, formationNo, strength int,
	flagMorales map[card.Morale]map[int]bool, flagTroops, handTroops, drawTroops []card.Troop) {
	sum := strength
	for _, strSet := range flagMorales {
		for s := range strSet {
			sum = sum - s
		}
	}
	for _, troop := range flagTroops {
		sum = sum - troop.Strenght()
	}
	elementNo := formationNo - (len(flagTroops) + len(flagMorales)) //must 1-3 as minimum one troop on flag.
	if elementNo == 0 {
		if sum == 0 {
			ana.Prop = 1
		} else {
			ana.Prop = 0
		}
	} else {
		_, factors := anaBattalionStrenghts(handTroops, drawTroops, sum, elementNo)

		if len(factors) != 0 {
			handTroops, drawTroops = anaBattalionReduce(factors, handTroops, drawTroops)
			var strenghts []int
			strenghts, factors = anaBattalionStrenghts(handTroops, drawTroops, sum, elementNo) //to keep index in strenghts and troop aligned.
			if len(handTroops)+len(drawTroops) < elementNo {
				ana.Prop = 0
			} else {
				var handMaxUsedStrix int
				//Playables rnd after map
				handMaxUsedStrix, ana.Prop, ana.Playables = anaBattalionHandCombi(factors, handTroops, strenghts, elementNo)
				if ana.Prop != 1 {
					if handMaxUsedStrix != -1 {
						elementNo, sum, handTroops, drawTroops, ana.Playables = anaBattalionHandMax(handMaxUsedStrix,
							elementNo, sum, factors, strenghts, handTroops, drawTroops) //Only one troop in Playables
					}
					ana.Valid = anaBattalionPerm(handTroops, drawTroops, sum, elementNo, nAll, dAll)
				}
			}
		} else {
			ana.Prop = 0
		}
	}

}

//anaBattalionHandMax reduce the problem with the higest card strenght from the hand. We move the card to the "flag".
//Some combination of sum may no longer be possible.
func anaBattalionHandMax(
	maxix, elementNo, sum int,
	factors [][]int,
	strenghts []int,
	handTroops, drawTroops []card.Troop) (updElementNo, updSum int,
	updHandTroops, updDrawTroops, handPlayableTroops []card.Troop) {

	maxFactors := make([][]int, 0, 0)
	for _, factor := range factors {
		containMax := false
		for _, strix := range factor {
			if strix == maxix {
				containMax = true
				break
			}
		}
		if containMax {
			maxFactors = append(maxFactors, factor)
		}
	}
	handPlayableTroops = []card.Troop{handTroops[maxix]}
	updElementNo = elementNo - 1
	updSum = sum - strenghts[maxix]
	var reduceHandTroops []card.Troop
	reduceHandTroops, updDrawTroops = anaBattalionReduce(maxFactors, handTroops, drawTroops)
	updHandTroops = make([]card.Troop, 0, len(handTroops))
	for _, ht := range reduceHandTroops {
		if ht != handPlayableTroops[0] {
			updHandTroops = append(updHandTroops, ht)
		}
	}

	return updElementNo, updSum, updHandTroops, updDrawTroops, handPlayableTroops
}

//anaBattalionHandCombi checks if cards on hand alone can make the combination (prop=1), it also
//find the index of the higest strenght card of the hand. This is only need if the hand can not make the
//combinations.
func anaBattalionHandCombi(
	factors [][]int,
	handTroops []card.Troop,
	strenghts []int,
	elementNo int) (handMaxUsedStrix int, prop float64, handPlayableTroops []card.Troop) {
	handTroopSet := make(map[card.Troop]bool)
	handMaxUsedStrix = -1
	for _, factor := range factors {
		factorHandTroops := make([]card.Troop, 0, 0)
		for _, strix := range factor {
			if strix < len(handTroops) {
				factorHandTroops = append(factorHandTroops, handTroops[strix])
				if handMaxUsedStrix == -1 || strenghts[strix] > strenghts[handMaxUsedStrix] {
					handMaxUsedStrix = strix
				}
			}
		}
		if len(factorHandTroops) == elementNo {
			prop = 1
			for _, troop := range factorHandTroops {
				handTroopSet[troop] = true
			}
		}
	}
	if prop == 1 {
		for troop := range handTroopSet {
			handPlayableTroops = append(handPlayableTroops, troop)
		}
	}
	return handMaxUsedStrix, prop, handPlayableTroops
}

//anaBattalionStrenghts factor the strenghts of the hand and the draw.
//factor contain the index of the strenghts and the index is also
//related to the hand and draw troops. It must not get out of
//synch.
func anaBattalionStrenghts(handTroops, drawTroops []card.Troop, sum, elementNo int) (strenghts []int, factors [][]int) {
	strenghts = troopsToStrenghts([][]card.Troop{handTroops, drawTroops})
	factors = math.FactorSum(strenghts, sum, elementNo, true)
	return strenghts, factors
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

//anaBattalionReduce reduces hand and draw to the valid cards
//according the factors. Warning it assumes the index match
//hand and draw troops.
func anaBattalionReduce(factors [][]int, handTroops, drawTroops []card.Troop) (updHandTroops, updDrawTroops []card.Troop) {
	updHandTroops = anaBattalionReduceList(handTroops, factors, 0)
	updDrawTroops = anaBattalionReduceList(drawTroops, factors, len(handTroops))
	return updHandTroops, updDrawTroops
}

//anaBattalionReduceList removes the troops that can no longer make a sum.
func anaBattalionReduceList(troops []card.Troop, factors [][]int, offset int) (updTroops []card.Troop) {
	for ix, troop := range troops {
		found := false
		ix = ix + offset
	Factor:
		for _, factor := range factors {
			for _, strix := range factor {
				if strix == ix {
					found = true
					break Factor
				}
			}
		}
		if found {
			updTroops = append(updTroops, troop)
		}
	}
	return updTroops
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
		for d := drawMaxValid; d >= dValid; d-- {
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
	color = COLNone
	colorFormation := false
	if len(combiTroops[COLNone]) == 0 {
		colorFormation = true
	}
	for _, troop := range flagTroops {
		isFound := false
		if colorFormation && color == COLNone {
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
