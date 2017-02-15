package gamepos

import (
	"fmt"
	"github.com/rezder/go-battleline/battbot/combi"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battbot/flag"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-error/log"
	slice "github.com/rezder/go-slice/int"
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

func (k *keep) String() string {
	txt := fmt.Sprintf("{Flag:%v FlagHand:%v NewFlagTroop:%v NewFlagMove:%v PriFlag:%v}",
		k.flag, k.flagHand, k.newFlagTroopix, k.newFlagMove, k.priFlagixs)
	return txt
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
		if flagAna.IsNewFlag && !isNewFlag {
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
	log.Printf(log.Debug, "Keep: flag,flagHand: %v,%v\n", k.flag, k.flagHand)
	return k
}
func (k *keep) flagSize() int {
	return len(k.flag)
}

//calcIsHandGood calculate if the hand is good.
//Calculate the number of good cards on the hand.
//Numbers of keep(flagHand) + good cards for new flags.
//More than two new flags must exist.
//Calculate suited connectors.
//Remove keep cards and loop over remaining n-2 and compare with
//cards to the right if connected add card to good set.
//Phalanx use troopsToValuexTroopixs
//Big bigger than 8.
func (k *keep) calcIsHandGood(flagsAna map[int]*flag.Analysis, capNo int) bool {
	noGood := len(k.flagHand)
	if !(noGood > capNo) {
		noNewFlags := 0
		for _, flagAna := range flagsAna {
			if flagAna.IsNewFlag {
				noNewFlags = noNewFlags + 1
			}
		}
		if noNewFlags > 1 {
			troopixs := make([]int, 0, 7)
			for _, ix := range k.handTroopixs {
				if !k.flagHand[ix] {
					troopixs = append(troopixs, ix)
				}
			}
			if len(troopixs) > 1 {
				goodSet := suitedConnecters(troopixs)
				if !(noGood+len(goodSet) > capNo) {
					goodSet = phalanxTroops(troopixs, goodSet)
					if !(noGood+len(goodSet) > capNo) {
						for _, troopix := range troopixs {
							troop, _ := cards.DrTroop(troopix)
							if troop.Value() > 7 {
								goodSet[troopix] = true
							}
						}
					}
				}
				noGood = noGood + len(goodSet)
			}
		}
	}
	return noGood > capNo
}

//phalanxTroops update map with troops with same value.
func phalanxTroops(troopixs []int, goodSet map[int]bool) map[int]bool {
	valueMap := combi.TroopsToValuexTroopixs(troopixs)
	for _, troops := range valueMap {
		if len(troops) > 1 {
			for _, troopix := range troops {
				goodSet[troopix] = true
			}
		}
	}
	return goodSet
}

//suitedConnecters find suited connecters.
func suitedConnecters(troopixs []int) (goodSet map[int]bool) {
	goodSet = make(map[int]bool)
	for i := 0; i < len(troopixs)-2; i++ {
		for j := i + 1; j < len(troopixs)-1; j++ {
			iTroop, _ := cards.DrTroop(troopixs[i])
			jTroop, _ := cards.DrTroop(troopixs[j])
			if iTroop.Color() == jTroop.Color() {
				if iTroop.Value() == jTroop.Value()+1 || iTroop.Value() == jTroop.Value()-1 {
					goodSet[troopixs[i]] = true
					goodSet[troopixs[j]] = true
				}
			}
		}
	}
	return goodSet
}

//calcPickTac evaluate if it is a good idea to pick tactic card.
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
//missing, redeploy could be included her but I do not think it
//is wort it.
//When opponent have formation or n-1 and higher rank mud,traitor and
//deserter is good if max sum higher fog is good.
//When n-1 and not win already simulate morale cards if win then cards is good.
func (k *keep) deckCalcPickTac(
	flagsAna map[int]*flag.Analysis,
	deck *botdeck.Deck,
	playableTacNo int,
	playableLeader bool,
	handTroopixs []int,
	deckMaxValues []int) bool {

	pickTac := false
	if deck.DeckTroopNo() == 0 {
		pickTac = true
	} else {
		offenceTacs, defenceTacs := offDefTacs(deck, playableLeader)
		offFlagSet, offenceTacSet := findOffenceFlags(offenceTacs, flagsAna, handTroopixs, deck, deckMaxValues)
		defFlagSet, defenceTacSet := findDefenceFlags(defenceTacs, flagsAna)
		if len(offenceTacSet) > 0 || len(defenceTacSet) > 0 {
			logTxt := ""
			if looseWinFlagExist(flagsAna) {
				logTxt = "Loose or win flag exist"
				pickTac = true
			} else if len(offFlagSet)+len(defFlagSet) > 4 && playableTacNo != 0 {
				logTxt = "Many flag need tactic cards"
				pickTac = true
			}
			if len(logTxt) != 0 {
				logTxt = logTxt + fmt.Sprintf("\nOffence Flags: %v Offence Tactics: %v\n", offFlagSet, offenceTacSet)
				logTxt = logTxt + fmt.Sprintf("Defence Flags: %v Defence Tactics: %v\n", defFlagSet, defenceTacSet)
				log.Print(log.Debug, logTxt)
			}
		}
	}
	return pickTac
}

//offDeffTacs find the relevante offencive and defencive tatic cards remaining in the deck.
//When we know the next tactic card only that card is used.
//if leader is not playable leader cards is removed.
func offDefTacs(deck *botdeck.Deck, playableLeader bool) (offenceTacs, defenceTacs []int) {
	defenceTacs = make([]int, 0, 4)
	offenceTacs = make([]int, 0, 4)
	scoutPeek := deck.ScoutReturnTacPeek()
	if scoutPeek != 0 {
		if isOffenceTac(scoutPeek) {
			offenceTacs = append(offenceTacs, scoutPeek)
		} else if isDefenceTac(scoutPeek) {
			defenceTacs = append(defenceTacs, scoutPeek)
		}
	} else {
		for tac := range deck.Tacs() {
			if isOffenceTac(tac) {
				offenceTacs = append(offenceTacs, tac)
			} else if isDefenceTac(tac) {
				defenceTacs = append(defenceTacs, tac)
			}
		}
	}
	if !playableLeader && len(offenceTacs) > 0 {
		copytac := make([]int, 0, 2)
		for _, tac := range offenceTacs {
			if !cards.IsLeader(tac) {
				copytac = append(copytac, tac)
			}
		}
		offenceTacs = copytac
	}
	return offenceTacs, defenceTacs
}
func looseWinFlagExist(flagsAna map[int]*flag.Analysis) (exist bool) {
	lostNo := countFlags(flagsAna, isFlagLostOrClaimed)
	if lostNo > 3 {
		exist = true
	} else {
		wonNo := countFlags(flagsAna, isFlagWonOrClaimed)
		if wonNo > 3 {
			exist = true
		} else {
			for _, flagAna := range flagsAna {
				if threeFlagsInRow(flagAna.Flagix, flagsAna, isFlagLostOrClaimed) ||
					threeFlagsInRow(flagAna.Flagix, flagsAna, isFlagWonOrClaimed) {
					exist = true
					break
				}
			}
		}
	}
	return exist
}
func countFlags(
	flagsAna map[int]*flag.Analysis,
	cond func(*flag.Analysis) bool) (no int) {
	for _, flagAna := range flagsAna {
		if cond(flagAna) {
			no = no + 1
		}
	}
	return no
}
func isFlagWonOrClaimed(flagAna *flag.Analysis) bool {
	return flagAna.IsWin() || flagAna.Flag.Claimed == flag.CLAIMPlay
}
func isOffenceTac(tac int) bool {
	return tac == cards.TC123 || tac == cards.TCDarius || tac == cards.TCAlexander || tac == cards.TC8
}
func isDefenceTac(tac int) bool {
	return tac == cards.TCMud || tac == cards.TCTraitor || tac == cards.TCDeserter || tac == cards.TCFog
}
func findDefenceFlags(
	defenceTacs []int,
	flagsAna map[int]*flag.Analysis) (flagixSet, tacixSet map[int]bool) {

	flagixSet = make(map[int]bool)

	tacixSet = make(map[int]bool)
	if len(defenceTacs) > 0 {
		for _, flagAna := range flagsAna {
			if len(flagAna.Flag.OppTroops) >= flagAna.FormationSize-1 &&
				!flagAna.IsClaimed && !flagAna.IsLost {
				for _, tacix := range defenceTacs {
					if tacix == cards.TCFog {
						if flagAna.TargetSum <= flagAna.BotMaxSum {
							tacixSet[tacix] = true
							flagixSet[flagAna.Flagix] = true
						}
					} else {
						tacixSet[tacix] = true
						flagixSet[flagAna.Flagix] = true
					}
				}
			}
		}
	}
	return flagixSet, tacixSet
}
func findOffenceFlags(
	offenceTacs []int,
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (flagixSet, tacixSet map[int]bool) {
	flagixSet = make(map[int]bool)
	tacixSet = make(map[int]bool)
	if len(offenceTacs) > 0 {
		for _, flagAna := range flagsAna {
			if len(flagAna.Flag.PlayTroops)+1 == flagAna.FormationSize && !flagAna.IsWin() {
				for _, tacix := range offenceTacs {
					isWin, _ := tacticMoveSim(flagAna.Flagix, flagAna.Flag, tacix, handTroopixs, deck, deckMaxValues)
					if isWin {
						flagixSet[flagAna.Flagix] = true
						tacixSet[tacix] = true
					}
				}
			}
		}
	}
	return flagixSet, tacixSet
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
		troopixs = troopixs[0:no]
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
		troopixs = keep2LowestValue(handTroopixs, keepFlag)
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
				if len(handTroopixs)-len(keepFlag) > 1 && missTroopsNo == 2 {
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
	logTxt = fmt.Sprintf("Dump move card,flag: %v,%v\n", troopix, flagAna.Flagix) + logTxt
	logTxt = logTxt + fmt.Sprintf("\nKeeps: %v", k)
	log.Print(log.Debug, logTxt)
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
func copyMap(m map[int]int) (cp map[int]int) {
	if m != nil {
		cp = make(map[int]int)
		for key, value := range m {
			cp[key] = value
		}
	}
	return cp
}
func (km *keepMap) calcHand(handTroopixs []int) (handKeep map[int]bool) {
	handKeep = make(map[int]bool)
	colorMap := copyMap(km.colorMap)
	valueMap := copyMap(km.valueMap)
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
			if colorMap[troop.Color()] != 0 {
				colorMap[troop.Color()] = colorMap[troop.Color()] - 1
				handKeep[troopix] = true
			} else if valueMap[troop.Value()] != 0 {
				valueMap[troop.Value()] = valueMap[troop.Value()] - 1
				handKeep[troopix] = true
			}
		}
	}
	return handKeep
}
