package prob

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-error/log"
	slice "github.com/rezder/go-slice/int"
)

//Keep keeps track of higher prioritized flags reserved cards.
type Keep struct {
	flag          map[card.Troop]bool
	flagHand      map[card.Troop]bool
	handAna       map[card.Troop][]*combi.Analysis
	NewFlagMoveix int
	PriFlagixs    []int
	botHand       *card.Cards
}

func (k *Keep) String() string {
	txt := fmt.Sprintf("{Flag:%v FlagHand:%v NewFlagMove:%v PriFlag:%v}",
		k.flag, k.flagHand, k.NewFlagMoveix, k.PriFlagixs)
	return txt
}

//NewKeep creates a keep.
func NewKeep(
	flagsAna map[int]*fa.Analysis,
	botHand, oppHand *card.Cards,
	moves Moves,
	deck *fa.Deck,
	isBotFirst bool) (k *Keep) {
	k = new(Keep)
	k.botHand = botHand
	k.NewFlagMoveix = -1
	keepMap := newKeepMap()
	isNewFlag := false
	k.PriFlagixs = prioritizePlayableFlags(flagsAna)

	for _, flagAna := range flagsAna {
		if flagAna.IsNewFlag && !isNewFlag {
			isNewFlag = true
		}
		if keepMapFlagAnaKeepTroops(flagAna) {
			missingNo := flagAna.FormationSize - flagAna.BotFormationSize
			keepMap.insert(flagAna.Analysis, missingNo)
		}
	}

	k.flag = keepMap.calcHand(k.botHand.Troops)
	if isNewFlag {
		k.handAna = fa.HandAnalyze(k.botHand.Troops, deck, isBotFirst)
		newCardMove, newFlagix := priNewFlagMove(k.PriFlagixs, flagsAna, k.handAna, k.botHand.Troops, k.flag)
		if !newCardMove.IsNone() {
			_, k.NewFlagMoveix = moves.FindHandFlag(newFlagix, newCardMove)
			if newCardMove.IsTroop() {
				troop := card.Troop(newCardMove)
				keepMap.cardsMap[troop] = true
				keepMap.insert(k.handAna[troop], 2)
				k.flagHand = keepMap.calcHand(k.botHand.Troops)
			} else {
				panic("New flag move should be a troop")
			}
		}
	}
	log.Printf(log.Debug, "Keep: flag,flagHand: %v,%v\n", k.flag, k.flagHand)
	return k
}

//CalcIsHandGood calculate if the hand is good.
//Calculate the number of good cards on the hand.
//Numbers of keep(flagHand) + good cards for new flags.
//More than two new flags must exist.
//Calculate suited connectors.
//Remove keep cards and loop over remaining n-2 and compare with
//cards to the right if connected add card to good set.
//Phalanx use troopsToStrxTroops
//Big bigger than 8.
func (k *Keep) CalcIsHandGood(flagsAna map[int]*fa.Analysis, capNo int) bool {
	noGood := len(k.flagHand)
	if !(noGood > capNo) {
		noNewFlags := 0
		for _, flagAna := range flagsAna {
			if flagAna.IsNewFlag {
				noNewFlags = noNewFlags + 1
			}
		}
		if noNewFlags > 1 {
			troops := make([]card.Troop, 0, 7)
			for _, troop := range k.botHand.Troops {
				if !k.flagHand[troop] {
					troops = append(troops, troop)
				}
			}
			if len(troops) > 1 {
				goodSet := suitedConnecters(troops)
				if !(noGood+len(goodSet) > capNo) {
					goodSet = phalanxTroops(troops, goodSet)
					if !(noGood+len(goodSet) > capNo) {
						for _, troop := range troops {
							if troop.Strenght() > 7 {
								goodSet[troop] = true
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
func phalanxTroops(troops []card.Troop, goodSet map[card.Troop]bool) map[card.Troop]bool {
	strMap := combi.TroopsToStrenghtTroops(troops)
	for _, strTroops := range strMap {
		if len(strTroops) > 1 {
			for _, troop := range strTroops {
				goodSet[troop] = true
			}
		}
	}
	return goodSet
}

//suitedConnecters find suited connecters.
//WARNING assume troops are sorted decreasing
func suitedConnecters(troops []card.Troop) (goodSet map[card.Troop]bool) {
	goodSet = make(map[card.Troop]bool)
	for i := 0; i < len(troops)-2; i++ {
		if troops[i].Color() == troops[i+1].Color() {
			if troops[i].Strenght() == troops[i+1].Strenght()-1 {
				goodSet[troops[i]] = true
				goodSet[troops[i+1]] = true
			}
		}
	}
	return goodSet
}

func keepLowestStrenghts(
	handTroops []card.Troop,
	keepTroops map[card.Troop]bool,
	keepNo int) (troops []card.Troop) {

	troops = make([]card.Troop, 0, keepNo)
	for i := len(handTroops) - 1; i > -1; i-- {
		handTroop := handTroops[i]
		if !keepTroops[handTroop] {
			troops = append(troops, handTroop)
			if len(troops) == keepNo {
				break
			}
		}
	}
	return troops
}
func keepFirstCard(keepTroops map[card.Troop]bool, troops []card.Troop) (firstTroop card.Card) {
	for _, troop := range troops {
		if !keepTroops[troop] {
			firstTroop = card.Card(troop)
			break
		}
	}
	return firstTroop
}

//RequestFlagHandLowestStrenght returns a troop from the hand,
// that is not reserved if possibel.
func (k *Keep) RequestFlagHandLowestStrenght() (cardMove card.Card) {
	troops := keepLowestStrenghts(k.botHand.Troops, k.flagHand, 1)
	if len(troops) > 0 {
		cardMove = card.Card(troops[0])
	}
	return cardMove
}

// RequestFirst returns a troop that is not reserved by a flag
//formation if possible.
func (k *Keep) RequestFirst(reqTroops []card.Troop) (troop card.Card) {
	troop = keepFirstCard(k.flagHand, reqTroops)
	if troop.IsNone() {
		troop = keepFirstCard(k.flag, reqTroops)
	}
	return troop
}

//RequestFirstHand returns a troop that is not reserved by a flag or
//a hand(empty flag move ) formation if possible.
func (k *Keep) RequestFirstHand(reqTroops []card.Troop) (troop card.Card) {
	troop = keepFirstCard(k.flagHand, reqTroops)
	return troop
}

//DemandScoutReturn returns maximum of two troop cards that can be returned without problem.
//Prioritised that least valued card first.
func (k *Keep) DemandScoutReturn(
	no int,
	flagsAna map[int]*fa.Analysis) (troops []card.Troop) {

	troops = make([]card.Troop, 0, 2)
	troops = keepScoutReturnRequest(k.botHand.Troops, k.flag, k.flagHand)
	if len(troops) >= no {
		troops = troops[0:no]
	} else {
		missTroopsNo := no - len(troops)
		leftTroops := make([]card.Troop, 0, len(k.botHand.Troops)-1)
		for _, troop := range k.botHand.Troops {
			isKeep := true
			for _, removeTroop := range troops {
				if removeTroop == troop {
					isKeep = false
					break
				}
			}
			if isKeep {
				leftTroops = append(leftTroops, troop)
			}
		}
		missixs := keepScoutReturnDemand(flagsAna, leftTroops, missTroopsNo, k.PriFlagixs)
		troops = append(troops, missixs...)
	}
	return troops
}

//keepScoutReturnRequest findes 2 troops witout problem if possible.
func keepScoutReturnRequest(handTroops []card.Troop, keepFlag, keepFlagHand map[card.Troop]bool) (troops []card.Troop) {
	n := 2
	troops = make([]card.Troop, 0, n)
	if len(handTroops)-len(keepFlagHand) > 1 {
		troops = keepLowestStrenghts(handTroops, keepFlagHand, 2)
	} else if len(handTroops)-len(keepFlag) > 1 {
		troops = keepLowestStrenghts(handTroops, keepFlag, 2)
	} else {
		if len(handTroops)-len(keepFlagHand) == 1 {
			troops = append(troops, keepLowestStrenghts(handTroops, keepFlagHand, 1)[0])
		} else if len(handTroops)-len(keepFlag) > 1 {
			troops = append(troops, keepLowestStrenghts(handTroops, keepFlag, 1)[0])
		}
	}
	return troops
}

//keepScoutReturnDemand finds the troops to return when all is in keep.
//Strategi removes the least valued flag from keepFlag untill ennough
//troops exist.
func keepScoutReturnDemand(
	flagsAna map[int]*fa.Analysis,
	handTroops []card.Troop,
	missTroopsNo int,
	priFlagixs []int) (troops []card.Troop) {

	troops = make([]card.Troop, 0, 2)
	copyFlagsAna := make(map[int]*fa.Analysis)
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
					missingNo := flagAna.FormationSize - flagAna.BotFormationSize
					keepMap.insert(flagAna.Analysis, missingNo)
				}
			}
			keepFlag := keepMap.calcHand(handTroops)
			if len(handTroops)-len(keepFlag) >= missTroopsNo {
				troops = keepLowestStrenghts(handTroops, keepFlag, missTroopsNo)
				break
			}
		}

	}
	if len(troops) == 0 {
		no := len(handTroops)
		if missTroopsNo == 2 {
			troops = []card.Troop{handTroops[no-1]}
		} else {
			troops = []card.Troop{handTroops[no-1], handTroops[no-2]}
		}
	}
	return troops
}

// DemandDump returns the least valued troop.
func (k *Keep) DemandDump(
	flagsAna map[int]*fa.Analysis,
	flagix int) (troop card.Troop) {
	logTxt := ""
	if len(k.flag) < len(k.botHand.Troops) {
		for _, combiFlag := range flagsAna[flagix].Analysis {
			if len(combiFlag.Playables) > 0 {
				cardMove := keepFirstCard(k.flag, combiFlag.Playables)
				if cardMove.IsTroop() {
					logTxt = "A combination card"
					troop = card.Troop(cardMove)
					break
				}
			}
		}
		if len(logTxt) == 0 {
			troop = k.DemandScoutReturn(1, flagsAna)[0]
			logTxt = "lowest strenght card not in keep"
		}
	} else {
		troop = k.botHand.Troops[len(k.botHand.Troops)-1]
		logTxt = "Min troop"
	}
	logTxt = fmt.Sprintf("Dump move card,flag: %v,%v\n", troop, flagix) + logTxt
	logTxt = logTxt + fmt.Sprintf("\nKeeps: %v", k)
	log.Print(log.Debug, logTxt)
	return troop
}

type keepMap struct {
	cardsMap map[card.Troop]bool
	colorMap map[int]int
	valueMap map[int]int
}

func newKeepMap() (km *keepMap) {
	km = new(keepMap)
	km.cardsMap = make(map[card.Troop]bool)
	km.colorMap = make(map[int]int)
	km.valueMap = make(map[int]int)
	return km
}
func keepMapFlagAnaKeepTroops(flagAna *fa.Analysis) bool {
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
				if card.FWedge.Value == cutOffFormationValue {
					cutOffFormationValue = card.FPhalanx.Value
				}
			}
		}
		if combiAna.Comb.Formation.Value < cutOffFormationValue {
			break
		} else {
		Loop:
			for _, troop := range combiAna.Playables {
				switch formationValue {
				case card.FWedge.Value:
					km.cardsMap[troop] = true
				case card.FPhalanx.Value:
					km.valueMap[troop.Strenght()] = km.valueMap[troop.Strenght()] + missingNo
					break Loop
				case card.FBattalion.Value:
					km.valueMap[troop.Color()] = km.valueMap[troop.Color()] + missingNo
					break Loop
				case card.FSkirmish.Value:
					km.valueMap[troop.Strenght()] = km.valueMap[troop.Strenght()] + 1
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

//calcHand creates hand keep.
//WARNING assume hand troops is sorted
func (km *keepMap) calcHand(handTroops []card.Troop) (handKeep map[card.Troop]bool) {
	handKeep = make(map[card.Troop]bool)
	colorMap := copyMap(km.colorMap)
	valueMap := copyMap(km.valueMap)

	for _, troop := range handTroops {
		if km.cardsMap[troop] {
			handKeep[troop] = true
		} else {
			if colorMap[troop.Color()] != 0 {
				colorMap[troop.Color()] = colorMap[troop.Color()] - 1
				handKeep[troop] = true
			} else if valueMap[troop.Strenght()] != 0 {
				valueMap[troop.Strenght()] = valueMap[troop.Strenght()] - 1
				handKeep[troop] = true
			}
		}
	}
	return handKeep
}
func priNewFlagMove(flagixs []int, //TODO The new flag move should take in to accout the conflicting target like keep but for deck. the problem flag(10,_,8) should count against newflag 8,8,8.
	flagsAna map[int]*fa.Analysis,
	handAna map[card.Troop][]*combi.Analysis,
	handTroops []card.Troop,
	keepFlag map[card.Troop]bool) (moveCard card.Card, moveFlagix int) {

	var logTxt string
	logTxt = "New flag is rank move"
	moveFlagix = -1
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppFormationSize > 0 {
			targetRank := flagAna.TargetRank
			if flagAna.IsTargetMade {
				targetRank = targetRank - 1
			}
			moveCard = newFlagTargetMove(handAna, targetRank)
		}
		if moveCard.IsTroop() {
			moveFlagix = flagix
			break
		}
	}
	if moveCard.IsNone() {
		var nfCard card.Card
		nfCard, logTxt = newFlagMadeCombiMove(handAna)
		if nfCard.IsNone() {
			nfCard, logTxt = newFlagPhalanxMove(handAna)
		}
		if nfCard.IsNone() {
			nfCard, logTxt = newFlagHigestRankMove(handAna, keepFlag)
		}

		if nfCard.IsTroop() {
			flagix := newFlagSelectFlag(card.Troop(nfCard), flagixs, flagsAna)
			if flagix != -1 {
				moveCard = nfCard
				moveFlagix = flagix
			}
		}
	}
	if moveCard.IsTroop() {
		log.Printf(log.Debug, "%v Troop: %v", logTxt, card.Troop(moveCard))
	}

	return moveCard, moveFlagix
}
func prioritizePlayableFlags(flagsAna map[int]*fa.Analysis) (flagixs []int) {
	flagValues := make([]int, len(flagsAna))
	for i, ana := range flagsAna {
		flagValues[i] = ana.FlagValue
	}
	sortixs := slice.SortWithIx(flagValues)
	flagixs = make([]int, 0, len(sortixs))
	for i := len(sortixs) - 1; i >= 0; i-- {
		if flagsAna[sortixs[i]].IsPlayable {
			flagixs = append(flagixs, sortixs[i])
		}
	}
	log.Printf(log.Debug, "Prioritized flags: %v\n", flagixs)
	return flagixs
}

//newFlagSelectFlag selects the new flag to play on.
func newFlagSelectFlag(troop card.Troop, flagixs []int, flagsAna map[int]*fa.Analysis) (flagix int) {
	flagix = -1
	newFlagixs := make([]int, 0, len(flagixs))
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppFormationSize == 0 {
			newFlagixs = append(newFlagixs, flagix)
		}
	}
	if len(newFlagixs) != 0 {
		if len(newFlagixs) == 1 {
			flagix = newFlagixs[0]
		} else {
			troopStr := troop.Strenght()
			switch {
			case troopStr > 7: //TODO this too simple we need to look rank compare it to target and bot rank,and if flag is lose game
				flagix = newFlagixs[0]
			case troopStr < 8 && troopStr > 5:
				flagix = newFlagixs[len(newFlagixs)/2]
			default:
				flagix = newFlagixs[len(newFlagixs)-1]
			}
		}
	}
	return flagix
}
func newFlagPhalanxMove(handAna map[card.Troop][]*combi.Analysis) (troop card.Card, logTxt string) {

	logTxt = "New flag is best phalanx or higher move"
	targetRank := combi.LastFormationRank(card.FPhalanx, 3)
	troop = newFlagTargetMove(handAna, targetRank)
	return troop, logTxt
}
func newFlagTargetMove(handAna map[card.Troop][]*combi.Analysis, targetRank int) (troop card.Card) {

	troopSumProp := make(map[card.Troop]float64)
	for handTroop, combiAnas := range handAna {
		for _, combiAna := range combiAnas {
			if combiAna.Prop > 0 && combiAna.Comb.Rank <= targetRank {
				troopSumProp[handTroop] = troopSumProp[handTroop] + combiAna.Prop
			}
		}
	}
	troop = probMaxSumTroop(troopSumProp)
	return troop
}

func newFlagMadeCombiMove(handAna map[card.Troop][]*combi.Analysis) (troop card.Card, logTxt string) {

	logTxt = "New flag is a combi move"
	targetRank := combi.LastFormationRank(card.FPhalanx, 3)
HandLoop:
	for handTroop, combiAnas := range handAna {
		for _, combiAna := range combiAnas {
			if combiAna.Prop == 1 && combiAna.Comb.Rank <= targetRank {
				troop = card.Card(handTroop)
				break HandLoop
			}
		}
	}

	return troop, logTxt
}

func newFlagHigestRankMove(
	handAna map[card.Troop][]*combi.Analysis,
	keepFlagTroops map[card.Troop]bool) (troop card.Card, logTxt string) {
	logTxt = "New flag is highest rank move"
	topRank := 10000
	for handTroop, combiAnas := range handAna {
		if !keepFlagTroops[handTroop] {
			for _, combiAna := range combiAnas {
				if combiAna.Prop > 0 {
					if combiAna.Comb.Rank < topRank {
						troop = card.Card(handTroop)
						topRank = combiAna.Comb.Rank
					}
					break
				}
			}
		}
	}

	return troop, logTxt
}
