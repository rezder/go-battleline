package gamepos

import (
	"github.com/rezder/go-battleline/battbot/combi"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battbot/flag"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-error/cerrors"
	slice "github.com/rezder/go-slice/int"
	"log"
)

type keep struct {
	flag           map[int]bool
	flagHand       map[int]bool
	handAna        map[int][]*combi.Analysis
	newFlagTroopix int
	newFlagMove    bat.Move
	priFlagixs     []int
	handTroopixs   []int
}

func newKeep(
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	isBotFirst bool) (k *keep) {

	k = new(keep)
	k.handTroopixs = handTroopixs
	keepMap := newKeepMap()
	isNewFlag := false
	k.priFlagixs = prioritizePlayableFlags(flagsAna)

	for _, flagAna := range flagsAna {
		if flagAna.IsNewFlag {
			isNewFlag = true
		}
		if keepMapFlagAnaKeepTroops(flagAna) {
			missingNo := flagAna.FormationSize - flagAna.BotTroopNo
			keepMap.insert(flagAna.Analysis, missingNo)
		}
	}

	k.flag = keepMap.calcHand(k.handTroopixs)
	if isNewFlag {
		k.handAna = flag.HandAnalyze(k.handTroopixs, deck, isBotFirst)
		k.newFlagTroopix, k.newFlagMove = priNewFlagMove(k.priFlagixs, flagsAna, k.handAna, k.handTroopixs, k.flag)
		if k.newFlagTroopix != 0 {
			keepMap.cardsMap[k.newFlagTroopix] = true
			keepMap.insert(k.handAna[k.newFlagTroopix], 2)
			k.flagHand = keepMap.calcHand(k.handTroopixs)
		}
	}
	if cerrors.LogLevel() == cerrors.LOG_Debug {
		log.Printf("Keep: flag,flagHand: %v,%v\n", k.flag, k.flagHand)
	}
	return k
}
func (k *keep) flagSize() int {
	return len(k.flag)
}
func (k *keep) calcIsHandGood() bool {
	//TODO len(k.flagHand)+higher than 7 and suited connecters and phalanx > 3
	return len(k.flagHand) > 2
}

//calcPickTac evaluate if it is a good idea to pick as tactic card.
//Tactic cards can be used planed offensive, opportunistic offensively
//and opportunistic defensively. Currently only scout is used planed.
//-
//We know when we are going to use a tactic card for scout it is when
//hand is good for the rest it is when they can prevent a lost
//or create a win. Find the cards in the deck that can be used.
//-
//The cost of having a tactic card is higher when keep is low.
//-
//The opponent move next that can change one flag status to lost
//going to a formation but only if it is ranked higher. When
//that flag is a losing game a destroy card may be wort going for.
//-
//The remaining troop cards also play a role as the alternative
//to chose a tactic card. We could check for good cards in the
//deck by evaluate all cards in the deck.
//-
//The destroy cards fog, mud, traitor or deserter is usefull when
//the opponent have a made formation or losing game flag exist.
//Check the sum on fog to evaluate it the rest is alwas good.
//The morale cards must be simulated for a win when one card is
//missing redeploy could be included her but i do not think it
//is wort it.
func (k *keep) deckCalcPickTac(deck *botdeck.Deck) bool {
	//TODO calc pick tac deck
	return k.calcIsHandGood()
}
func keep2LowestValue(handTroopixs []int, keepTroopixs map[int]bool) (troopixs []int) {
	troopixs = make([]int, 2)
	copyTropixs := make([]int, len(handTroopixs))
	copy(copyTropixs, handTroopixs)
	ltix := keepLowestValue(copyTropixs, keepTroopixs)
	troopixs[0] = ltix
	copyTropixs = slice.Remove(copyTropixs, ltix)
	ltix = keepLowestValue(copyTropixs, keepTroopixs)
	troopixs[1] = ltix
	return troopixs
}
func keepLowestValue(handTroops []int, keepTroops map[int]bool) (troopix int) {
	value := 0
	for _, handTroopix := range handTroops {
		if !keepTroops[handTroopix] {
			handTroop, _ := cards.DrTroop(handTroopix)
			if troopix != 0 {
				if handTroop.Value() < value {
					troopix = handTroopix
					value = handTroop.Value()
				}
			} else {
				troopix = handTroopix
				value = handTroop.Value()
			}
		}
	}
	return troopix
}
func keepFirstCard(keepTroops map[int]bool, troopixs []int) (firstTroopix int) {
	for _, troopix := range troopixs {
		if !keepTroops[troopix] {
			firstTroopix = troopix
			break
		}
	}
	return firstTroopix
}
func (k *keep) requestFlagHandLowestValue(requestTroopixs []int) (troopix int) {
	return keepLowestValue(requestTroopixs, k.flagHand)
}
func (k *keep) requestFirst(reqTroopixs []int) (troopix int) {
	troopix = keepFirstCard(k.flagHand, reqTroopixs)
	if troopix == 0 {
		troopix = keepFirstCard(k.flag, reqTroopixs)
	}
	return troopix
}
func (k *keep) requestFirstHand(reqTroopixs []int) (troopix int) {
	troopix = keepFirstCard(k.flagHand, reqTroopixs)
	return troopix
}

//scoutReturnTroops returns maximum of two troop cards that can be returned without problem.
//Prioritised that least valued card first.
func (k *keep) demandScoutReturn(
	no int,
	flagsAna map[int]*flag.Analysis,
	deck *botdeck.Deck) (troopixs []int) {

	troopixs = make([]int, 0, 2)
	troopixs = keepScoutReturnRequest(k.handTroopixs, k.flag, k.flagHand)
	if len(troopixs) >= no {
		troopixs = troopixs[0:]
	} else {
		missTroopsNo := no - len(troopixs)
		leftTroopixs := slice.WithOutNew(k.handTroopixs, troopixs)

		missixs := keepScoutReturnDemand(flagsAna, deck, leftTroopixs, missTroopsNo, k.priFlagixs)
		troopixs = append(troopixs, missixs...)
	}
	return troopixs
}

//keepScoutReturnRequest findes 2 troops witout problem if possible.
func keepScoutReturnRequest(handTroopixs []int, keepFlag, keepFlagHand map[int]bool) (troopixs []int) {
	n := 2
	troopixs = make([]int, 0, n)
	if len(handTroopixs)-len(keepFlagHand) > 1 {
		troopixs = keep2LowestValue(handTroopixs, keepFlagHand)
	} else if len(handTroopixs)-len(keepFlag) > 1 {
		troopixs = keep2LowestValue(handTroopixs, keepFlagHand)
	} else {
		if len(handTroopixs)-len(keepFlagHand) == 1 {
			troopixs = append(troopixs, keepLowestValue(handTroopixs, keepFlagHand))
		} else if len(handTroopixs)-len(keepFlag) > 1 {
			troopixs = append(troopixs, keepLowestValue(handTroopixs, keepFlag))
		}
	}
	return troopixs
}

//keepScoutReturnDemand finds the troops to return when all is in keep.
//Strategi removes the least valued flag from keepFlag untill ennough
//troops exist.
func keepScoutReturnDemand(
	flagsAna map[int]*flag.Analysis,
	deck *botdeck.Deck,
	handTroopixs []int,
	missTroopsNo int,
	priFlagixs []int) (troopixs []int) {

	troopixs = make([]int, 0, 2)
	copyFlagsAna := make(map[int]*flag.Analysis)
	for key, value := range flagsAna {
		copyFlagsAna[key] = value
	}
	for i := len(priFlagixs) - 1; i > 0; i-- {
		delFlagAna := flagsAna[i]
		delete(copyFlagsAna, delFlagAna.Flagix)
		if keepMapFlagAnaKeepTroops(delFlagAna) {
			keepMap := newKeepMap()
			for _, flagAna := range copyFlagsAna {
				if keepMapFlagAnaKeepTroops(flagAna) {
					missingNo := flagAna.FormationSize - flagAna.BotTroopNo
					keepMap.insert(flagAna.Analysis, missingNo)
				}
			}
			keepFlag := keepMap.calcHand(handTroopixs)
			if len(handTroopixs)-len(keepFlag) >= missTroopsNo {
				if len(handTroopixs)-len(keepFlag) == 2 {
					troopixs = keep2LowestValue(handTroopixs, keepFlag)
				} else {
					troopix := keepLowestValue(handTroopixs, keepFlag)
					troopixs = append(troopixs, troopix)
				}
				break
			}
		}

	}
	if len(troopixs) == 0 {
		troopixs = min2Troop(handTroopixs)
		troopixs = troopixs[:missTroopsNo]
	}
	return troopixs
}

func (k *keep) demandDump(flagAna *flag.Analysis) (troopix int) {
	logTxt := ""
	if k.flagSize() < len(k.handTroopixs) {
		for _, combiFlag := range flagAna.Analysis {
			if len(combiFlag.HandCardixs) > 0 {
				troopix = keepFirstCard(k.flag, combiFlag.HandCardixs)
				logTxt = "A combination card"
				break
			}
		}
		if troopix == 0 {
			troopix = keepFirstCard(k.flag, k.handTroopixs)
			logTxt = "First card not in keep"
		}
	} else {
		troopix = minTroop(k.handTroopixs)
		logTxt = "Min troop"
	}
	if cerrors.LogLevel() == cerrors.LOG_Debug {
		log.Printf("Dump move card,flag: %v,%v\n", troopix, flagAna.Flagix)
		log.Println(logTxt)
		log.Printf("Keeps: %v", k)
	}
	return troopix
}

type keepMap struct {
	cardsMap map[int]bool
	colorMap map[int]int
	valueMap map[int]int
}

func newKeepMap() (km *keepMap) {
	km = new(keepMap)
	km.cardsMap = make(map[int]bool)
	km.colorMap = make(map[int]int)
	km.valueMap = make(map[int]int)
	return km
}
func keepMapFlagAnaKeepTroops(flagAna *flag.Analysis) bool {
	return !flagAna.IsFog && !flagAna.IsLost && !flagAna.IsNewFlag && flagAna.IsPlayable
}

//keepTroops keeps all cards from the top formation, in case of wedge
//formation the phalanx formations is also included.
func (km *keepMap) insert(combiAnas []*combi.Analysis, missingNo int) {
	cutOffFormationValue := 0
	formationValue := 0
	for _, combiAna := range combiAnas {
		if cutOffFormationValue == 0 {
			if combiAna.Prop > 0 {
				cutOffFormationValue = combiAna.Comb.Formation.Value
				formationValue = cutOffFormationValue
				if cards.FWedge.Value == cutOffFormationValue {
					cutOffFormationValue = cards.FPhalanx.Value
				}
			}
		}
		if combiAna.Comb.Formation.Value < cutOffFormationValue {
			break
		} else {
		Loop:
			for _, troopix := range combiAna.HandCardixs {
				switch formationValue {
				case cards.FWedge.Value:
					km.cardsMap[troopix] = true
				case cards.FPhalanx.Value:
					troop, _ := cards.DrTroop(troopix)
					km.valueMap[troop.Value()] = km.valueMap[troop.Value()] + missingNo
					break Loop
				case cards.FBattalion.Value:
					troop, _ := cards.DrTroop(troopix)
					km.valueMap[troop.Color()] = km.valueMap[troop.Color()] + missingNo
					break Loop
				case cards.FSkirmish.Value:
					troop, _ := cards.DrTroop(troopix)
					km.valueMap[troop.Value()] = km.valueMap[troop.Value()] + 1
					break Loop
				}
			}
		}

	}
}
func (km *keepMap) calcHand(handTroopixs []int) (handKeep map[int]bool) {

	handKeep = make(map[int]bool)
	sortedHandTroopixs := make([]int, 0, 7)
	for _, handTroopix := range handTroopixs {
		sortedHandTroopixs = addSortedCards(sortedHandTroopixs, handTroopix, func(troopix int) int {
			troop, _ := cards.DrTroop(troopix)
			return troop.Value()
		})
	}

	for _, troopix := range sortedHandTroopixs {
		if km.cardsMap[troopix] {
			handKeep[troopix] = true
		} else {
			troop, _ := cards.DrTroop(troopix)
			if km.colorMap[troop.Color()] != 0 {
				km.colorMap[troop.Color()] = km.colorMap[troop.Color()] - 1
				handKeep[troopix] = true
			} else if km.valueMap[troop.Value()] != 0 {
				km.valueMap[troop.Value()] = km.valueMap[troop.Value()] - 1
				handKeep[troopix] = true
			}
		}
	}
	return handKeep
}
