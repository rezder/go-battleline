package prob

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-error/log"
	slice "github.com/rezder/go-slice/int"
)

//Keep keeps track of higher prioritized flags reserved cards.
type Keep struct {
	flag           map[card.Troop]bool
	flagHand       map[card.Troop]bool
	handAna        map[card.Troop][]*combi.Analysis
	NewFlagMoveix  int
	PriPlayFlagixs []int
	PriLiveFlagixs []int
	botHandTroops  []card.Troop
}

func (k *Keep) String() string {
	txt := fmt.Sprintf("{Flag:%v FlagHand:%v NewFlagMove:%v PriFlag:%v}",
		k.flag, k.flagHand, k.NewFlagMoveix, k.PriPlayFlagixs)
	return txt
}

//NewKeep creates a keep.
func NewKeep(
	flagsAna map[int]*fa.Analysis,
	moves Moves,
	deckHandTroops *dht.Cache,
	botix int,
) (k *Keep) {

	k = new(Keep)
	k.botHandTroops = deckHandTroops.SrcHandTroops[botix]
	k.NewFlagMoveix = -1
	keepMap := newKeepMap()
	isNewFlag := false
	k.PriPlayFlagixs, k.PriLiveFlagixs = prioritizePlayableFlags(flagsAna)

	for _, flagAna := range flagsAna {
		if flagAna.IsNewFlag && !isNewFlag {
			isNewFlag = true
		}
		if keepMapFlagAnaKeepTroops(flagAna) {
			missingNo := flagAna.FormationSize - flagAna.BotFormationSize
			keepMap.insert(flagAna.RankAnas, missingNo)
		}
	}

	k.flag = keepMap.calcHand(k.botHandTroops)
	if isNewFlag {
		k.handAna = fa.HandAnalyze(deckHandTroops, botix)
		newCardMove, newFlagix := priNewFlagMove(k.PriPlayFlagixs, flagsAna, k.handAna, k.botHandTroops, k.flag)
		if !newCardMove.IsNone() {
			if len(moves) > 0 && moves[0].MoveType.IsHand() {
				_, k.NewFlagMoveix = moves.FindHandFlag(newFlagix, newCardMove)
			}
			if newCardMove.IsTroop() {
				troop := card.Troop(newCardMove)
				keepMap.cardsMap[troop] = true //This is random as any phalanx card may have more connected cards
				keepMap.insert(k.handAna[troop], 2)
				k.flagHand = keepMap.calcHand(k.botHandTroops)
			} else {
				panic("New flag move should be a troop")
			}
		}
	}
	log.Printf(log.Debug, "Keep: flag,flagHand: %v,%v,keepMap:%v\n", k.flag, k.flagHand, keepMap)
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
			for _, troop := range k.botHandTroops {
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

func keepLastCard(keepTroops map[card.Troop]bool, troops []card.Troop) (lastTroop card.Card) {
	for i := len(troops) - 1; i >= 0; i-- {
		troop := troops[i]
		if !keepTroops[troop] {
			lastTroop = card.Card(troop)
			break
		}
	}
	return lastTroop
}

//RequestFlagHandLowestStrenght returns a troop from the hand,
// that is not reserved if possibel.
func (k *Keep) RequestFlagHandLowestStrenght() (cardMove card.Card) {
	troops := keepLowestStrenghts(k.botHandTroops, k.flagHand, 1)
	if len(troops) > 0 {
		cardMove = card.Card(troops[0])
	}
	return cardMove
}

// RequestLast returns a troop that is not reserved by a flag
//formation if possible.
func (k *Keep) RequestLast(reqTroops []card.Troop) (troop card.Card) {
	troop = keepLastCard(k.flagHand, reqTroops)
	if troop.IsNone() {
		troop = keepLastCard(k.flag, reqTroops)
	}
	return troop
}

//RequestLastHand returns a troop that is not reserved by a flag or
//a hand(empty flag move ) formation if possible.
func (k *Keep) RequestLastHand(reqTroops []card.Troop) (troop card.Card) {
	troop = keepLastCard(k.flagHand, reqTroops)
	return troop
}

//DemandScoutReturn returns maximum of two troop cards that can be returned without problem.
//Prioritised that least valued card first.
func (k *Keep) DemandScoutReturn(
	no int,
	flagsAna map[int]*fa.Analysis) (troops []card.Troop) {

	troops = make([]card.Troop, 0, 2)
	troops = keepScoutReturnRequest(k.botHandTroops, k.flag, k.flagHand)
	if len(troops) >= no {
		troops = troops[0:no]
	} else {
		missTroopsNo := no - len(troops)
		leftTroops := make([]card.Troop, 0, len(k.botHandTroops)-1)
		for _, troop := range k.botHandTroops {
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
		missixs := keepScoutReturnDemand(flagsAna, leftTroops, missTroopsNo, k.PriPlayFlagixs)
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
					keepMap.insert(flagAna.RankAnas, missingNo)
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
	if len(k.flag) < len(k.botHandTroops) {
		for _, combiFlag := range flagsAna[flagix].RankAnas {
			if len(combiFlag.Playables) > 0 {
				cardMove := keepLastCard(k.flag, combiFlag.Playables)
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
		troop = k.botHandTroops[len(k.botHandTroops)-1]
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
func (km *keepMap) Len() (no int) {
	no = len(km.cardsMap)
	if len(km.colorMap) > 0 {
		for _, colNo := range km.colorMap {
			no = no + colNo
		}
	}
	if len(km.valueMap) > 0 {
		for _, strNo := range km.valueMap {
			no = no + strNo
		}
	}
	return no
}
func keepMapFlagAnaKeepTroops(flagAna *fa.Analysis) bool {
	return !flagAna.IsLost && !flagAna.IsNewFlag && flagAna.IsPlayable
}

//keepTroops keeps all cards from the top formation, in case of wedge
//formation the phalanx formations is also included.
func (km *keepMap) insert(combiAnas []*combi.Analysis, missingNo int) {
	cutOffFormationValue := 0
	for _, combiAna := range combiAnas {
		if cutOffFormationValue == 0 {
			if combiAna.Prop > 0 {
				cutOffFormationValue = combiAna.Comb.Formation.Value
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
				switch combiAna.Comb.Formation.Value {
				case card.FWedge.Value:
					km.cardsMap[troop] = true
				case card.FPhalanx.Value:
					km.valueMap[troop.Strenght()] = km.valueMap[troop.Strenght()] + missingNo
					break Loop
				case card.FBattalion.Value:
					km.colorMap[troop.Color()] = km.colorMap[troop.Color()] + missingNo
					break Loop
				case card.FHost.Value:
					fallthrough //TODO this is right for host only on card and allways the first last maybe ok if one card only missing.
					//I guess host and skirmish dont need two resever often.
				case card.FSkirmish.Value:
					//TODO this is right for Skirmish only one card ??? or maybe it is
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
func priNewFlagMove(flagixs []int,
	flagsAna map[int]*fa.Analysis,
	handAna map[card.Troop][]*combi.Analysis,
	handTroops []card.Troop,
	keepFlag map[card.Troop]bool) (moveCard card.Card, moveFlagix int) {

	var logTxt string

	moveCard, moveFlagix, logTxt = newFlagPriTargetMove(flagixs, flagsAna, handAna, keepFlag, 1)
	if moveCard.IsNone() {
		emptyFlagixs := emptyPriFlagixs(flagixs, flagsAna)
		if len(emptyFlagixs) > 0 {
			moveCard, moveFlagix, logTxt = newFlagMadeCombiMove(handAna, keepFlag, emptyFlagixs)
			if moveCard.IsNone() {
				moveCard, moveFlagix, logTxt = newFlagPhalanxMove(handAna, keepFlag, emptyFlagixs) //what about keep
			}
		}
		if moveCard.IsNone() {
			moveCard, moveFlagix, logTxt = newFlagPriTargetMove(flagixs, flagsAna, handAna, keepFlag, 0)
		}
	}
	if moveCard.IsTroop() {
		log.Printf(log.Debug, "%v Troop: %v Flagix: %v", logTxt, card.Troop(moveCard), moveFlagix)
	}

	return moveCard, moveFlagix
}
func newFlagPriTargetMove(
	priFlagixs []int,
	flagsAna map[int]*fa.Analysis,
	handAna map[card.Troop][]*combi.Analysis,
	keepFlagTroops map[card.Troop]bool,
	minOppFormSize int,
) (moveCard card.Card, moveFlagix int, logTxt string) {

	logTxt = fmt.Sprintf("New flag is target rank move opp. size: %v ", minOppFormSize)
	moveFlagix = -1
	for _, flagix := range priFlagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppFormationSize >= minOppFormSize {
			targetRank := flagAna.TargetRank
			if flagAna.IsTargetMade && combi.RankTieBreaker(targetRank, flagAna.FormationSize).IsRank() {
				targetRank = targetRank - 1 //rank 1 is not possible because the flag is lost
			}
			moveCard = newFlagTargetMove(handAna, targetRank, keepFlagTroops)
		}
		if moveCard.IsTroop() {
			moveFlagix = flagix
			break
		}
	}
	return moveCard, moveFlagix, logTxt
}
func prioritizePlayableFlags(flagsAna map[int]*fa.Analysis) (playFlagixs, liveFlagixs []int) {
	flagValues := make([]int, len(flagsAna))
	for i, ana := range flagsAna {
		flagValues[i] = ana.FlagValue
	}
	sortixs := slice.SortWithIx(flagValues)
	playFlagixs = make([]int, 0, len(sortixs))
	liveFlagixs = make([]int, 0, len(sortixs))
	for i := len(sortixs) - 1; i >= 0; i-- {
		if flagsAna[sortixs[i]].IsPlayable {
			playFlagixs = append(playFlagixs, sortixs[i])
		}
		if !flagsAna[sortixs[i]].IsClaimed {
			liveFlagixs = append(liveFlagixs, sortixs[i])
		}
	}
	log.Printf(log.Debug, "Prioritized playable flags: %v, live flags: %v\n", playFlagixs, liveFlagixs)
	return playFlagixs, liveFlagixs
}

func emptyPriFlagixs(priFlagixs []int, flagsAna map[int]*fa.Analysis) []int {
	newFlagixs := make([]int, 0, len(priFlagixs))
	for _, flagix := range priFlagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppFormationSize == 0 {
			newFlagixs = append(newFlagixs, flagix)
		}
	}
	return newFlagixs
}

//newFlagSelectFlag selects the new flag to play on.
//TODO this too simple we need to look rank compare it to target and bot rank,and if flag is lose game.
func newFlagSelectFlag(troop card.Troop, priFlagixs []int) (flagix int) {
	if len(priFlagixs) == 1 {
		flagix = priFlagixs[0]
	} else {
		troopStr := troop.Strenght()
		switch {
		case troopStr > 7:
			flagix = priFlagixs[0]
		case troopStr < 8 && troopStr > 5:
			flagix = priFlagixs[len(priFlagixs)/2]
		default:
			flagix = priFlagixs[len(priFlagixs)-1]
		}
	}

	return flagix
}
func newFlagPhalanxMove(
	handAna map[card.Troop][]*combi.Analysis,
	keepFlag map[card.Troop]bool,
	priFlagixs []int,
) (troop card.Card, flagix int, logTxt string) {
	flagix = -1
	logTxt = "New flag is best phalanx or higher move"
	targetRank := combi.LastFormationRank(card.FPhalanx, 3)
	troop = newFlagTargetMove(handAna, targetRank, keepFlag)
	if troop.IsTroop() {
		flagix = newFlagSelectFlag(card.Troop(troop), priFlagixs)
	}
	return troop, flagix, logTxt
}

//newFlagTargetMove find troop to play on new flag with a target.
//TODO could be better if playables not in keep is condsiddered.
func newFlagTargetMove(
	handAna map[card.Troop][]*combi.Analysis,
	targetRank int, //Made formation must be one lower
	keepFlagTroops map[card.Troop]bool,
) (troop card.Card) {

	troopSumProp := make(map[card.Troop]float64)
	for handTroop, combiAnas := range handAna {
		if !keepFlagTroops[handTroop] {
			for _, combiAna := range combiAnas {
				if combiAna.Prop > 0 && combiAna.Comb.Rank <= targetRank {
					troopSumProp[handTroop] = troopSumProp[handTroop] + combiAna.Prop
				}
			}
		}
	}
	log.Printf(log.Debug, "New flag target move prob: %v", troopSumProp)
	troops := probMaxSumTroops(troopSumProp)
	if len(troops) > 0 {
		if len(troops) == 1 {
			troop = card.Card(troops[0])
		} else {
			minKeepPhalanxNo := 100
			var minKeepMap *keepMap
			for _, t := range troops {
				keepMap := newKeepMap()
				keepMap.insert(handAna[t], 2)
				keepPhalanxNo := keepMap.Len()
				if keepPhalanxNo < minKeepPhalanxNo {
					troop = card.Card(t)
					minKeepPhalanxNo = keepPhalanxNo
					minKeepMap = keepMap
				} else if keepPhalanxNo == minKeepPhalanxNo {
					if len(minKeepMap.cardsMap) == 1 && len(keepMap.cardsMap) == 1 {
						for min := range minKeepMap.cardsMap {
							for cur := range keepMap.cardsMap {
								if min.Strenght() < cur.Strenght() {
									troop = card.Card(t)
									minKeepMap = keepMap
								}
							}
						}
					}
				}
			}
		}
	}
	return troop
}

//probMaxSumTroop findes the troops with the bigest probability if any. Tiebreaker
//troop strenght.
func probMaxSumTroops(cardSumProp map[card.Troop]float64) (maxTroops []card.Troop) {
	var maxTroop card.Troop
	maxSumProp := float64(0)
	for troop, sumProp := range cardSumProp {
		if sumProp > maxSumProp {
			maxTroop = troop
			maxSumProp = sumProp
		} else if sumProp == maxSumProp {
			if maxTroop == 0 || troop.Strenght() > maxTroop.Strenght() {
				maxTroop = troop
			}
		}
	}
	if maxTroop != 0 {
		for troop, sumProp := range cardSumProp {
			if sumProp == maxSumProp && maxTroop.Strenght() == troop.Strenght() {
				maxTroops = append(maxTroops, troop)
			}
		}
	}
	return maxTroops
}

//newFlagMadeCombiMove finds the best made combi wedge or phalanx.
//TODO CHECK this could be better if more combies exist but it does not happen often.
func newFlagMadeCombiMove(
	handAna map[card.Troop][]*combi.Analysis,
	keepFlagTroops map[card.Troop]bool,
	emptyFlagixs []int,
) (troop card.Card, flagix int, logTxt string) {

	logTxt = "New flag is a made combi move"
	flagix = -1
	targetRank := combi.LastFormationRank(card.FPhalanx, 3) + 1
	keepPhalanxNo := 100
	for handTroop, combiAnas := range handAna {
		if !keepFlagTroops[handTroop] {
			for _, combiAna := range combiAnas {
				if combiAna.Comb.Rank > targetRank {
					break
				}
				if combiAna.Prop == 1 && newFlagMadeCombiPlayable(combiAna.Playables, handTroop, keepFlagTroops) {
					if combiAna.Comb.Rank < targetRank {
						targetRank = combiAna.Comb.Rank
						if combiAna.Comb.Formation == card.FPhalanx {
							kepMap := newKeepMap()
							kepMap.insert(combiAnas, 2)
							keepPhalanxNo = kepMap.Len()
						}
						flagix = newFlagSelectFlag(handTroop, emptyFlagixs)
						troop = card.Card(handTroop)
						break
					} else if combiAna.Comb.Rank == targetRank && combiAna.Comb.Formation == card.FPhalanx {
						kepMap := newKeepMap()
						kepMap.insert(combiAnas, 2)
						if kepMap.Len() < keepPhalanxNo { //Smallest to dump card not need in case of 1b,1r,1g,2g
							flagix = newFlagSelectFlag(handTroop, emptyFlagixs)
							troop = card.Card(handTroop)
						}
						break
					}
				}
			}
		}
	}
	return troop, flagix, logTxt
}
func newFlagMadeCombiPlayable(playables []card.Troop, handTroop card.Troop, keepFlagTroops map[card.Troop]bool) bool {
	if !keepFlagTroops[handTroop] &&
		len(playables) > 1 &&
		playables[0].Strenght() <= handTroop.Strenght() &&
		len(keepLowestStrenghts(playables, keepFlagTroops, 2)) == 2 {
		return true
	}
	return false
}
