package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/battleline/cards"
	math "github.com/rezder/go-math/int"
	slice "github.com/rezder/go-slice/int"
)

//Analysis the result of a combination analysis.
type Analysis struct {
	Comb        *Combination
	Valid       uint64
	All         uint64
	HandCardixs []int
	Prop        float64
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	txt := fmt.Sprintf("{Rank:%v Prob:%v Formation:%v Strenght:%v Valid:%v All:%v Hand:%v}",
		ana.Comb.Rank, ana.Prop, ana.Comb.Formation.Name, ana.Comb.Strength, ana.Valid, ana.All, ana.HandCardixs)
	return txt
}

//Ana analyze a combination.
func Ana(comb *Combination, flagCards []int, handCards []int, drawSet map[int]bool,
	drawNo int, mud bool) (ana *Analysis) {
	formationNo := 3
	if mud {
		formationNo = 4
	}
	ana = new(Analysis)
	ana.Comb = comb
	nAll := uint64(len(drawSet))
	dAll := uint64(drawNo)
	valid, flagTroops, flagMorales, color := anaCombiFlagCards(comb.Troops, flagCards, comb.Formation.Value)
	if !valid {
		ana.Prop = 0
	} else {
		if len(flagTroops) != 0 {
			handTroops, drawTroops := anaCombiHandDraw(comb.Troops[color], handCards, drawSet)
			switch comb.Formation {
			case cards.FWedge:
				validMorale := moralesReduce(flagTroops, flagMorales)
				if !validMorale {
					ana.Prop = 0
				} else {
					reduceCalc := false
					invalidOnly := false
					if len(flagMorales) != 0 {
						flagValues := TroopsToValuexTroopixs(flagTroops)
						handValues := TroopsToValuexTroopixs(handTroops)
						drawValues := TroopsToValuexTroopixs(drawTroops)
						reduceCalc, invalidOnly = anaStraightMorales(flagMorales,
							handValues, drawValues, flagValues, formationNo)
						handTroops = valuexTroopixsToTroops(handValues)
						drawTroops = valuexTroopixsToTroops(drawValues)
					}
					if !invalidOnly {
						anaWedgePhalanx(ana, nAll, dAll, formationNo, flagTroops, flagMorales, handTroops, drawTroops)
						//This the rar case of mud,123 joker and troop with value 3.
						if reduceCalc { // 1 out 3 is bad when draw two cards
							if nAll > dAll {
								ana.Valid = ana.Valid - math.Comb(nAll-uint64(3), dAll-uint64(2))
							}
						}
					} else {
						ana.Prop = 0
					}
				}

			case cards.FPhalanx:
				anaWedgePhalanx(ana, nAll, dAll, formationNo, flagTroops, flagMorales, handTroops, drawTroops)

			case cards.FBattalion:
				anaBattalion(ana, nAll, dAll, formationNo, comb.Strength, flagMorales,
					flagTroops, handTroops, drawTroops)

			case cards.FSkirmish:
				anaSkirmish(ana, nAll, dAll, formationNo, flagMorales, flagTroops, handTroops, drawTroops)

			} //end switch
		} else { //Only tacs
			//Should not bee needed
			panic("Illegal argument flag most have troops, this have only moral tactic troops")
		}
	}
	return ana
}

//anaSkirmish analyze a skirmish combination 1,2,3,4 in any colors.
//The wedge code for morales can be reused if colapsing the cards of the same value.
//The colapsing transformation of the data is very helpfull, as it makes it easy to
//treat cards of same values in the same way. As soon as a card(value) is used any other
//cards of the same values is useless. The reduceCalc problem is a matter of checking if
//only cards of the jokers values is drawn.
//#ana
func anaSkirmish(ana *Analysis, nAll, dAll uint64, formationNo int,
	flagMorales map[int]map[int]bool, flagTroops, handTroops, drawTroops []int) {
	validMorale := moralesReduce(flagTroops, flagMorales)
	if !validMorale {
		ana.Prop = 0
	} else {
		flagValues := TroopsToValuexTroopixs(flagTroops)
		if len(flagValues) != len(flagTroops) {
			ana.Prop = 0
		} else {
			handValues := TroopsToValuexTroopixs(handTroops)
			drawValues := TroopsToValuexTroopixs(drawTroops)
			for v := range flagValues {
				delete(handValues, v)
				delete(drawValues, v)
			}
			reduceCalc := false
			invalidOnly := false

			if len(flagMorales) != 0 {
				reduceCalc, invalidOnly = anaStraightMorales(flagMorales,
					handValues, drawValues, flagValues, formationNo)
			}
			if len(handValues) != 0 {
				for v := range handValues {
					delete(drawValues, v)
				}
			}
			if !invalidOnly {
				missingNo := formationNo - (len(flagValues) + len(flagMorales) + len(handValues))
				ana.HandCardixs = handTroops
				if missingNo <= 0 {
					ana.Prop = 1
				} else {
					if len(drawValues) >= missingNo {
						drawTroops = valuexTroopixsToTroops(drawValues)
						nValid := uint64(len(drawTroops))
						dValid := uint64(missingNo)
						values := troopsToValues([][]int{drawTroops})
						var moralValues map[int]bool
						if reduceCalc {
							moralValues = make(map[int]bool)
							for _, moral := range flagMorales {
								if len(moral) == 2 {
									for value := range moral {
										moralValues[value] = true
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
							ana.Valid = ana.Valid + (anaSkirmishCombi(values, missingNo, int(d), moralValues) *
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
//drawn values. When the number of distinct values equal the number of cards we need to make the combination,
//the permutation of the good cards is valid unless it is the two joker values from the 123 morale.
func anaSkirmishCombi(values []int, missingNo, d int, moralValues map[int]bool) (validNo uint64) {
	validNo = 0
	math.Perm(len(values), d, func(perm []int) bool {
		var unique [11]bool //for speed maps are expensive
		uniqueNo := 0
		for _, ix := range perm {
			if !unique[values[ix]] {
				uniqueNo = uniqueNo + 1
				unique[values[ix]] = true
			}
			if uniqueNo >= missingNo {
				if len(moralValues) == missingNo {
					equal := true
					for v := range unique {
						if !moralValues[v] {
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
//Its is the worst because many combination of values can make the
//same sum. So which card is played can influnce the probabilty of making the sum.
//to get around that problem we just assume one will allways play the higest card.
//this may not be the card that create the best probability for the sum. The lowest
//usually is best to play. The idea was that it was the best player strategi, but
//i am not sure, it may just be easier. It is also assumed all moral cards have
//its higest value.
//We transfer all cards to its value and hopfully keep track of which value belong to
//which card. We often bunk values from hand and draw together as any of then can be
//used in the sum.
//#ana
func anaBattalion(ana *Analysis, nAll, dAll uint64, formationNo, strength int,
	flagMorales map[int]map[int]bool, flagTroops, handTroops, drawTroops []int) {
	sum := strength
	for _, valueMap := range flagMorales {
		for value := range valueMap {
			sum = sum - value
		}
	}
	for _, troopix := range flagTroops {
		troop, _ := cards.DrTroop(troopix)
		sum = sum - troop.Value()
	}
	elementNo := formationNo - (len(flagTroops) + len(flagMorales)) //must 1-3 as minimum one troop on flag.
	if elementNo == 0 {
		if sum == 0 {
			ana.Prop = 1
		} else {
			ana.Prop = 0
		}
	} else {
		values, factors := anaBattalionValues(handTroops, drawTroops, sum, elementNo)

		if len(factors) != 0 {
			handTroops, drawTroops = anaBattalionReduce(factors, handTroops, drawTroops)
			values, factors = anaBattalionValues(handTroops, drawTroops, sum, elementNo) //to keep index in values and troop aligned.
			if len(handTroops)+len(drawTroops) < elementNo {
				ana.Prop = 0
			} else {
				var handMaxUsedValueix int
				handMaxUsedValueix, ana.Prop, ana.HandCardixs = anaBattalionHandCombi(factors, handTroops, values, elementNo)
				if ana.Prop != 1 {
					if handMaxUsedValueix != -1 {
						elementNo, sum, handTroops, drawTroops, ana.HandCardixs = anaBattalionHandMax(handMaxUsedValueix,
							elementNo, sum, factors, values, handTroops, drawTroops)
					}
					ana.Valid = anaBattalionPerm(handTroops, drawTroops, sum, elementNo, nAll, dAll)
				}
			}
		} else {
			ana.Prop = 0
		}
	}

}

//anaBattalionHandMax reduce the problem with the higest hand value. We move the card to the "flag".
//Some combination of sum may no longer be possible.
func anaBattalionHandMax(maxix, elementNo, sum int, factors [][]int, values, handTroops,
	drawTroops []int) (updElementNo, updSum int, updHandTroops, updDrawTroops, handCardixs []int) {
	maxFactors := make([][]int, 0, 0)
	for _, factor := range factors {
		containMax := false
		for _, valueix := range factor {
			if valueix == maxix {
				containMax = true
				break
			}
		}
		if containMax {
			maxFactors = append(maxFactors, factor)
		}
	}
	handCardixs = []int{handTroops[maxix]}
	updElementNo = elementNo - 1
	updSum = sum - values[maxix]
	maxCardix := handTroops[maxix]
	updHandTroops, updDrawTroops = anaBattalionReduce(maxFactors, handTroops, drawTroops)
	updHandTroops = slice.Remove(handTroops, maxCardix)

	return updElementNo, updSum, updHandTroops, updDrawTroops, handCardixs
}

//anaBattalionHandCombi checks if cards on hand alone can make the combination (prop=1), it also
//find the index of the higest value card of the hand. This is only need if the hand can not make the
//combinations.
func anaBattalionHandCombi(factors [][]int, handTroops, values []int, elementNo int) (handMaxUsedValueix int, prop float64, handCardixs []int) {
	handCardixSet := make(map[int]bool)
	handMaxUsedValueix = -1
	for _, factor := range factors {
		factorHandCardixs := make([]int, 0, 0)
		for _, valueix := range factor {
			if valueix < len(handTroops) {
				factorHandCardixs = append(factorHandCardixs, handTroops[valueix])
				if handMaxUsedValueix == -1 || values[valueix] > values[handMaxUsedValueix] {
					handMaxUsedValueix = valueix
				}
			}
		}
		if len(factorHandCardixs) == elementNo {
			prop = 1
			for _, cardix := range factorHandCardixs {
				handCardixSet[cardix] = true
			}
		}
	}
	if prop == 1 {
		for cardix := range handCardixSet {
			handCardixs = append(handCardixs, cardix)
		}
	}
	return handMaxUsedValueix, prop, handCardixs
}

//anaBattalionValues factor the values of the hand and the draw.
//factor contain the index of the values and the index is also
//related to the hand and draw troops. It must not get out of
//synch.
func anaBattalionValues(handTroops, drawTroops []int, sum, elementNo int) (values []int, factors [][]int) {
	values = troopsToValues([][]int{handTroops, drawTroops})
	factors = math.FactorSum(values, sum, elementNo, true)
	return values, factors
}

//troopsToValues combins the values of troops from any lists of troops.
func troopsToValues(troops [][]int) (values []int) {
	values = make([]int, 0, 10)
	for _, ts := range troops {
		for _, t := range ts {
			troop, _ := cards.DrTroop(t)
			values = append(values, troop.Value())
		}
	}
	return values
}

//anaBattalionReduce reduces hand and draw to the valid cards
//according the factors. Warning it assumes the index match
//hand and draw troops.
func anaBattalionReduce(factors [][]int, handTroops, drawTroops []int) (updHandTroops, updDrawTroops []int) {
	updHandTroops = anaBattalionReduceList(handTroops, factors, 0)
	updDrawTroops = anaBattalionReduceList(drawTroops, factors, len(handTroops))
	return updHandTroops, updDrawTroops
}

//anaBattalionReduceList removes the troops that can no longer make a sum.
func anaBattalionReduceList(troopixs []int, factors [][]int, offset int) (updTroopixs []int) {
	for ix, troopix := range troopixs {
		found := false
		ix = ix + offset
	Factor:
		for _, factor := range factors {
			for _, valueix := range factor {
				if valueix == ix {
					found = true
					break Factor
				}
			}
		}
		if found {
			updTroopixs = append(updTroopixs, troopix)
		}
	}
	return updTroopixs
}

//anaStraightMorales handles the moral tactic cards on the flag.
//It tries to deduce the moral value if possible and reduce
//the problem accordingly.
//The problem is the limmited moral tactic cards 123 or 8.
//Example if the flag have 8 value the moral 8 makes the combination
//impossible. The 123 can be equal limmited depending on the combination.
//In the combination 3,4,5 it can only be one value.
//Depending on the hand values we may figure out the moral value.
//If the moral card can only be one value we can remove that value
//from the hand and draw.
//If the moral card can be more values than we need then it does
//not matter as moral card just takes the missing value.
//The only real problem a 4 card combination one of the 1,2,3 cards
//together with 123 and no cards on the hand in that case if we
//draw the two cards that the moral can be we do not have a combination.
//So is the draw only contain those two values invalidOnly is true.
//If all three remain we must correct for the invalid ones. For a wedge
//Combination it is easy 1 out three (3 over 2) is invalid.
//#flagMorales
//#handValues
//#drawValues
func anaStraightMorales(flagMorales map[int]map[int]bool, handValues map[int][]int, drawValues map[int][]int,
	flagValues map[int][]int, formationNo int) (reduceCalc bool, invalidOnly bool) {
	for _, moral := range flagMorales {
		if len(moral) == 2 { //This is only possible for 123 and Leaders
			missingNo := formationNo - (len(flagValues) + len(flagMorales))
			if missingNo >= len(moral) { //Only 2 = 2 is possible for now, because of one flag troop. This only possible for 123 morale card again because of one flag troop
				handMoralValues := make([]int, 0, 0)
				for v := range moral {
					_, found := handValues[v]
					if found {
						handMoralValues = append(handMoralValues, v)
					}
				}
				switch len(handMoralValues) {
				case 1: //if one we know the moral value.
					fallthrough
				case 2: //if two we just pick one and we know the moral value.
					delete(moral, handMoralValues[0])
				case 0: //if zero invalid draws is possible from deck
					invalidOnly = true
					if len(drawValues) > 1 {
						for v := range drawValues {
							_, found := moral[v] //if draw contain a none joker
							if !found {
								invalidOnly = false
								break
							}
						}
						if len(drawValues) == 3 { // one out 3 is bad when draw two cards
							reduceCalc = true
						}
					}
				}
			}
		}
		if !invalidOnly {
			for _, moral := range flagMorales {
				if len(moral) == 1 { //Here can be only moral 123 or 8
					for v := range moral {
						delete(handValues, v)
						delete(drawValues, v)
					}
				}
			}
		}
	}
	return reduceCalc, invalidOnly
}

//valuexTroopixsToTroops converts back to troopixs.
func valuexTroopixsToTroops(valuexTroopixs map[int][]int) (allTroopixs []int) {
	allTroopixs = make([]int, 0, len(valuexTroopixs))
	for _, troopixs := range valuexTroopixs {
		allTroopixs = append(allTroopixs, troopixs...)
	}
	return allTroopixs
}

//troopsToValues colapses troopixs on values.
func TroopsToValuexTroopixs(troopixs []int) (valuexTroopixs map[int][]int) {
	valuexTroopixs = make(map[int][]int)
	for _, troopix := range troopixs {
		troop, _ := cards.DrTroop(troopix)
		troops, found := valuexTroopixs[troop.Value()]
		if !found {
			troops = make([]int, 0, 1)
		}
		troops = append(troops, troopix)
		valuexTroopixs[troop.Value()] = troops
	}
	return valuexTroopixs
}

//moralesReduce removes morales with same values a troopixs.
//#morales
func moralesReduce(troops []int, morales map[int]map[int]bool) (valid bool) {
	valid = true
	if len(morales) != 0 {
		for _, troopix := range troops {
			for _, moral := range morales {
				troop, _ := cards.DrTroop(troopix)
				delete(moral, troop.Value())
				if len(moral) == 0 {
					valid = false
					break
				}
			}
		}
	}
	return valid
}

//anaBattalionPerm update analysis with the valid number of combinations and probability.
func anaBattalionPerm(handTroops, drawTroops []int,
	sum, elementNo int, nAll, dAll uint64) (valid uint64) {
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
func anaBattalionPermCombi(drawTroops, handTroops []int, sum, elementNo, d int) (validNo uint64) {
	validNo = 1
	if len(drawTroops) > 1 {
		validNo = 0
		handValues := troopsToValues([][]int{handTroops})
		drawValues := troopsToValues([][]int{drawTroops})
		math.Perm(len(drawTroops), d, func(perm []int) bool {
			values := make([]int, d+len(handValues))
			for i, drawix := range perm {
				values[i] = drawValues[drawix]
			}
			if len(handTroops) != 0 {
				copy(values[d:], handValues)
			}
			factors := math.FactorSum(values, sum, elementNo, false)
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
func anaWedgePhalanx(ana *Analysis, nAll uint64, dAll uint64, formationNo int, flagTroops []int,
	flagMorales map[int]map[int]bool, handTroops []int, drawTroops []int) {
	missingNo := formationNo - (len(flagTroops) + len(flagMorales) + len(handTroops))
	ana.HandCardixs = handTroops
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
func anaCombiHandDraw(troops []int, handCards []int, drawSet map[int]bool) (handTroops []int, drawTroops []int) {
	for _, troop := range troops {
		for _, handCard := range handCards {
			if troop == handCard {
				handTroops = append(handTroops, handCard)
				break
			}
		}
		if drawSet[troop] {
			drawTroops = append(drawTroops, troop)
		}
	}
	return handTroops, drawTroops
}

//anaCombiFlagCards finds the valid troops of the combination given the flag cards.
//valid: True if a valid combination may exit.
//validTroops: The valid troops.
//validMorales: The possible values for moral tactic cards.
//color: The color of the combination.
func anaCombiFlagCards(troops map[int][]int, flagCards []int, formation int) (valid bool,
	validTroops []int, validMorales map[int]map[int]bool, color int) {
	validMorales = make(map[int]map[int]bool)
	valid = true
	color = cards.COLNone
	colorFormation := false
	var valueTroops []int
	if len(troops[cards.COLNone]) == 0 {
		colorFormation = true
		valueTroops = troops[cards.COLGreen] //if only morales card we never switch to a color and the prop will be wrong
	} else {
		valueTroops = troops[cards.COLNone]
	}
	for _, cix := range flagCards {
		found := false
		if cards.IsMorale(cix) {
			if cix == cards.TC123 {
				if formation != cards.FBattalion.Value {
					found = moralesUpd(validMorales, cix, valueTroops, []int{1, 2, 3})
				} else {
					moralesSetValue(validMorales, cix, 3)
					found = true
				}
			} else if cix == cards.TC8 {
				if formation != cards.FBattalion.Value {
					found = moralesUpd(validMorales, cix, valueTroops, []int{8})
				} else {
					moralesSetValue(validMorales, cix, 8)
					found = true
				}
			} else {
				if formation != cards.FBattalion.Value {
					found = moralesUpd(validMorales, cix, valueTroops, nil)
				} else {
					moralesSetValue(validMorales, cix, 10)
					found = true
				}
			}
		} else {
			if colorFormation && color == cards.COLNone {
				troop, _ := cards.DrTroop(cix)
				color = troop.Color()
			}
			for _, v := range troops[color] {
				if v == cix {
					found = true
					validTroops = append(validTroops, cix)
					break
				}
			}
		}
		if !found {
			valid = false
			break
		}
	}
	return valid, validTroops, validMorales, color
}

//moralesUpd set a moral tactic card, if moral is a leader
//set moralValues to nil for all values.
//#moralValues
func moralesUpd(validMorales map[int]map[int]bool, moralix int, troops []int, moralValues []int) (found bool) {
	for _, troopix := range troops {
		troop, _ := cards.DrTroop(troopix)
		ok := false
		if len(moralValues) == 0 {
			ok = true
		} else {
			for _, v := range moralValues {
				if v == troop.Value() {
					ok = true
					break
				}
			}
		}
		if ok {
			found = true
			moralesSetValue(validMorales, moralix, troop.Value())
		}
	}
	return found
}

//moralesSetValue sets a value on moral.
//#moralValues
func moralesSetValue(m map[int]map[int]bool, moralix, value int) {
	valueSet, ok := m[moralix]
	if !ok {
		valueSet = make(map[int]bool)
		m[moralix] = valueSet
	}
	valueSet[value] = true
}
