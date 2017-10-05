package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/battbot/combi"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-error/log"
	math "github.com/rezder/go-math/int"
	slice "github.com/rezder/go-slice/int"
	"sort"
)

// Analysis a analysis of the flag.
type Analysis struct {
	Flagix        int
	TargetRank    int
	TargetSum     int
	BotMaxRank    int
	BotMaxSum     int
	OppTroopsNo   int
	BotTroopNo    int
	FormationSize int
	FlagValue     int
	IsTargetMade  bool
	IsLost        bool
	IsNewFlag     bool
	IsPlayable    bool
	IsClaimed     bool
	IsFog         bool
	IsLoosingGame bool
	Analysis      []*combi.Analysis
	SumCards      []int
	Flag          *Flag
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	combiAna := "<nil>"
	if ana.Analysis != nil {
		positive := make([]*combi.Analysis, 0, 0)
		for _, c := range ana.Analysis {
			if c != nil {
				if c.Prop > 0 {
					positive = append(positive, c)
				}
			}
		}
		if len(positive) == 0 {
			combiAna = "[]"
		} else {
			combiAna = "["
			for i, c := range positive {
				if i == 4 {
					break
				}
				if i != 0 {
					combiAna = combiAna + " "
				}
				combiAna = combiAna + c.String()
			}
			combiAna = combiAna + "]"
		}
	}
	txt := fmt.Sprintf("{Flagix:%v TargetRank:%v TargetSum:%v BotMaxRank:%v BotMaxSum: %v SumCards:%v FlagValue:%v IsLost:%v LooseGame:%v Combination:%v", ana.Flagix+1, ana.TargetRank, ana.TargetSum, ana.BotMaxRank, ana.BotMaxSum, ana.SumCards, ana.FlagValue, ana.IsLost, ana.IsLoosingGame, combiAna)
	return txt
}

// NewAnalysis create a flag analysis.
func NewAnalysis(
	flag *Flag,
	botHandTroops []int,
	deckMaxValues []int,
	deck *botdeck.Deck,
	flagix int,
	isBotFirst bool) (fa *Analysis) {

	fa = new(Analysis)
	fa.Flagix = flagix
	fa.Flag = flag
	if flag.IsClaimed() {
		fa.IsClaimed = true
	} else {
		oppMoraleNo := countMoraleCards(flag.OppTroops)
		fa.OppTroopsNo = len(flag.OppTroops)
		if len(flag.OppTroops) == flag.FormationSize() {
			fa.IsTargetMade = true
		}
		if !flag.IsFog() {
			if fa.OppTroopsNo != oppMoraleNo {
				fa.TargetRank = CalcMaxRank(flag.OppTroops, deck.OppHand(),
					deck.Troops(), deck.OppDrawNo(!isBotFirst), flag.IsMud())
			} else {
				fa.TargetRank = calcMaxRankNewFlag(flag.OppTroops, deck.OppHand(),
					deck.Troops(), flag.FormationSize())
			}
		}
		oppDeckValues := make([]int, len(deckMaxValues))
		copy(oppDeckValues, deckMaxValues)
		fa.TargetSum, _ = maxSum(flag.OppTroops, deck.OppHand(), oppDeckValues, flag.IsMud())
		if fa.IsTargetMade {
			fa.TargetSum = fa.TargetSum + 1
		}

		fa.IsFog = flag.IsFog()
		fa.FormationSize = flag.FormationSize()
		fa.BotTroopNo = len(flag.PlayTroops)
		if fa.BotTroopNo < flag.FormationSize() {
			fa.IsPlayable = true
		}
		botMoraleNo := countMoraleCards(flag.PlayTroops)
		if fa.BotTroopNo == botMoraleNo {
			fa.IsNewFlag = true //TODO what if moral 8 or 1,2,3 is alone then handAna is of no use.
			// It could only happen if moral card was played as second card and traitor need the
			// troop.
			if !flag.IsFog() {
				fa.BotMaxRank = calcMaxRankNewFlag(flag.PlayTroops, botHandTroops, deck.Troops(), flag.FormationSize())
			}
		} else {
			if !flag.IsFog() {
				fa.Analysis = rankAnalyze(flag.PlayTroops, botHandTroops,
					deck.Troops(), deck.BotDrawNo(isBotFirst), flag.IsMud())
				fa.BotMaxRank = calcBotMaxRank(fa.Analysis)
			}
		}
		botDeckValues := make([]int, len(deckMaxValues))
		copy(botDeckValues, deckMaxValues)

		fa.SumCards, fa.BotMaxSum = sumCards(botDeckValues, flag.PlayTroops, botHandTroops,
			fa.TargetSum, flag.FormationSize())

		if fa.IsTargetMade {
			fa.IsLost = lost(fa.TargetRank, fa.TargetSum, fa.BotMaxSum, fa.BotMaxRank)
		}
		log.Printf(log.Debug, "Flag:%v Lost:%v BotRank:%v BotSum:%v TargetRank:%v TargetSum:%v\n", fa.Flagix, fa.IsLost, fa.BotMaxRank, fa.BotMaxSum, fa.TargetRank, fa.TargetSum)
	}

	return fa
}

func calcBotMaxRank(combisAna []*combi.Analysis) (rank int) {
	for _, ana := range combisAna {
		if ana.Prop > 0 {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}

func countMoraleCards(cardixs []int) (no int) {
	for _, cardix := range cardixs {
		if cards.IsMorale(cardix) {
			no = no + 1
		}
	}
	return no
}

//IsWin check for a won flag, only works if target is ranked and
//Bot have played a troop on the flag.
func (ana *Analysis) IsWin() (isWin bool) {
	if !ana.IsLost {
		targetRank := ana.TargetRank
		if targetRank == 0 && !ana.IsFog {
			targetRank = 100
		}
		if ana.IsTargetMade && ana.TargetRank != 0 {
			targetRank = targetRank - 1
		}
		if targetRank != 0 {
			for _, combiAna := range ana.Analysis {
				if combiAna.Prop == 1 {
					if combiAna.Comb.Rank <= targetRank {
						isWin = true
					}
					break
				}
				if combiAna.Comb.Rank > targetRank {
					break
				}
			}
		}
		if !isWin {
			if targetRank == 100 || targetRank == 0 {
				curSum := MoraleTroopsSum(ana.Flag.PlayTroops)
				if ana.BotTroopNo < ana.FormationSize {
					curSum = curSum + ana.FormationSize - ana.BotTroopNo
				}
				if ana.TargetSum < curSum {
					isWin = true
				}
			}
		}

	}
	return isWin
}

// HandAnalyze create a rank analysis of each troop on the hand.
func HandAnalyze(
	handTroops []int,
	deck *botdeck.Deck,
	isBotFirst bool) (ana map[int][]*combi.Analysis) {
	ana = make(map[int][]*combi.Analysis)
	for _, troopix := range handTroops {
		handCards := slice.Remove(handTroops, troopix)
		ana[troopix] = rankAnalyze([]int{troopix}, handCards, deck.Troops(), deck.BotDrawNo(isBotFirst), false)
	}
	return ana
}

//lost calculate if a flag is lost it assume opponent moves first and win
//when rank or sum is equal
func lost(targetRank, targetSum, botMaxSum, botRank int) bool {
	isLost := false
	if targetRank > 0 {
		if botRank >= targetRank {
			isLost = true
		}
	} else { //skirmish line
		if botRank == 0 {
			if botMaxSum < targetSum { //target Sum has been increased with one because it was made
				isLost = true
			}
		}
	}
	return isLost

}

//sumCards the troops playable in a skirmish line formation (sum).
//Its cards that give enough to reach the target or if target is not possible
//the higest cards that still exist.
//#deckValues
func sumCards(deckValues, flagTroops, botHandTroops []int, targetSum, formationSize int) (playableCards []int, botMaxSum int) {
	if len(flagTroops) == formationSize {
		botMaxSum = MoraleTroopsSum(flagTroops)
	} else {
		botMaxSum, deckValues = maxSum(flagTroops, botHandTroops, deckValues, formationSize == 4)
		if targetSum >= botMaxSum {
			playableCards = make([]int, 0, len(botHandTroops))
			for _, cardix := range botHandTroops {
				troop, ok := cards.DrTroop(cardix)
				if ok {
					if troop.Value() >= deckValues[formationSize-len(flagTroops)-1] {
						playableCards = append(playableCards, cardix)
					}
				}
			}
		} else {
			flagSum := MoraleTroopsSum(flagTroops)
			playableCards = make([]int, 0, len(botHandTroops))
			needValue := targetSum - flagSum
			avgNeedValue := float32(needValue) / float32(formationSize-len(flagTroops))
			for _, cardIx := range botHandTroops {
				troop, ok := cards.DrTroop(cardIx)
				if ok {
					if needValue > 0 {
						if float32(troop.Value()) >= avgNeedValue {
							playableCards = append(playableCards, cardIx)
						}
					} else {
						playableCards = append(playableCards, cardIx)
					}
				}
			}
		}
	}
	return playableCards, botMaxSum
}

func CalcMaxRank(flagCards []int, handCards []int, drawSet map[int]bool,
	drawNo int, mud bool) (rank int) {
	combinations := combi.CombinationsMud(mud)
	allCombi := math.Comb(uint64(len(drawSet)), uint64(drawNo))
	for _, comb := range combinations {
		ana := combi.Ana(comb, flagCards, handCards, drawSet, drawNo, mud)
		ana.All = allCombi
		if ana.Valid > 0 {
			ana.Prop = float64(ana.Valid) / float64(allCombi)
		}
		if ana.Prop > 0 {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}
func calcMaxRankNewFlag(flagCards, handCards []int, drawSet map[int]bool,
	formationSize int) (rank int) {
	combinations := combi.Combinations(formationSize)
	moraleNo := countMoraleCards(flagCards)
	flagTroopValue := 0
	for _, cardix := range flagCards {
		flagTroopValue = flagTroopValue + cards.MoraleMaxValue(cardix)
	}
	if moraleNo == formationSize {
		rank = findBattalionRank(flagTroopValue, combinations)
	} else {
		allDrawSet := make(map[int]bool)
		for i, v := range drawSet {
			allDrawSet[i] = v
		}
		for _, v := range handCards {
			if cards.IsTroop(v) {
				allDrawSet[v] = true
			}
		}
		if moraleNo == 3 || moraleNo == 2 && formationSize == 3 {
			maxValue := maxTroopValue(allDrawSet)
			rank = findBattalionRank(flagTroopValue+maxValue, combinations)
		} else if moraleNo == 2 {
			mv1, mv2 := max2TroopValues(allDrawSet)
			rank = findBattalionRank(mv1+mv2+flagTroopValue, combinations)
		} else if moraleNo == 1 {
			rank = calcMaxRankNewFlagOneMoral(flagCards[0], allDrawSet, formationSize, combinations)
		} else {
			rank = calcMaxRankNewFlagZeroMoral(allDrawSet, formationSize, combinations)
		}
	}
	return rank
}
func calcMaxRankNewFlagZeroMoral(
	allDrawSet map[int]bool,
	formationSize int,
	combinations []*combi.Combination) (rank int) {
	calcBatt := false
Loop:
	for _, comb := range combinations {
		switch comb.Formation.Value {
		case cards.FWedge.Value:
			for _, troopixs := range comb.Troops {
				made := true
				for _, troopix := range troopixs {
					if !allDrawSet[troopix] {
						made = false
						break
					}
				}
				if made {
					rank = comb.Rank
					break Loop
				}
			}
		case cards.FPhalanx.Value:
			if findMissingTroops(formationSize, comb.Troops[cards.COLNone], allDrawSet) {
				rank = comb.Rank
				break Loop
			}
		case cards.FBattalion.Value:
			if !calcBatt {
				calcBatt = true
				missingNo := formationSize
				moraleValue := 0
				rank = calcBestBattalionRank(missingNo, moraleValue, allDrawSet, combinations)
				if rank != 0 {
					break Loop
				}
			}

		case cards.FSkirmish.Value:
			valueTroops := sortTroopValue(comb.Troops[cards.COLNone])
			failed := false
			for _, troopixs := range valueTroops {
				found := false
				for _, troopix := range troopixs {
					if allDrawSet[troopix] {
						found = true
						break
					}
				}
				if !found {
					failed = true
					break
				}
			}
			if !failed {
				rank = comb.Rank
				break Loop
			}
		}
	}
	return rank
}
func sortTroopValue(troopixs []int) (m map[int][]int) {
	return sortTroopsValueOrColor(troopixs, false)
}
func sortTroopsValueOrColor(troopixs []int, isColor bool) (valueTroops map[int][]int) {
	valueTroops = make(map[int][]int)
	for _, troopix := range troopixs {
		troop, _ := cards.DrTroop(troopix)
		value := 0
		if isColor {
			value = troop.Color()
		} else {
			value = troop.Value()
		}
		m, ok := valueTroops[value]
		if ok {
			m = append(m, troopix)
		} else {
			m = make([]int, 0, 6)
			m = append(m, troopix)
		}
		valueTroops[value] = m
	}
	return valueTroops
}
func calcMaxRankNewFlagOneMoral(
	moralix int,
	allDrawSet map[int]bool,
	formationSize int,
	combinations []*combi.Combination) (rank int) {
	calcBatt := false
Loop:
	for _, comb := range combinations {
		switch comb.Formation.Value {
		case cards.FWedge.Value:
			for _, troopixs := range comb.Troops {
				made := true
				usedJoker := false
				for _, troopix := range troopixs {
					if !allDrawSet[troopix] {
						if (!usedJoker) && validJoker(troopix, moralix) {
							usedJoker = true
						} else {
							made = false
							break
						}
					}
				}
				if made {
					rank = comb.Rank
					break Loop
				}
			}
		case cards.FPhalanx.Value:
			if validJoker(comb.Troops[cards.COLNone][0], moralix) {
				missingNo := formationSize - 1
				if findMissingTroops(missingNo, comb.Troops[cards.COLNone], allDrawSet) {
					rank = comb.Rank
					break Loop
				}
			}

		case cards.FBattalion.Value:
			if !calcBatt {
				calcBatt = true
				missingNo := formationSize - 1
				moraleValue := cards.MoraleMaxValue(moralix)
				rank = calcBestBattalionRank(missingNo, moraleValue, allDrawSet, combinations)
				if rank != 0 {
					break Loop
				}
			}

		case cards.FSkirmish.Value:
			valueTroops := sortTroopValue(comb.Troops[cards.COLNone])
			failed := false
			usedJoker := false
			for _, troopixs := range valueTroops {
				found := false
				for _, troopix := range troopixs {
					if allDrawSet[troopix] {
						found = true
						break
					}
				}
				if !found {
					if (!usedJoker) && validJoker(troopixs[0], moralix) {
						usedJoker = true
					} else {
						failed = true
						break
					}
				}
			}
			if !failed {
				rank = comb.Rank
				break Loop
			}
		}
	}
	return rank
}
func calcBestBattalionRank(
	missingNo, moraleValue int,
	deckSet map[int]bool,
	combinations []*combi.Combination) (rank int) {

	colorTroops := make(map[int][]int)
	for troopix := range deckSet {
		troop, _ := cards.DrTroop(troopix)
		m, ok := colorTroops[troop.Color()]
		if ok {
			m = append(m, troopix)
		} else {
			m = make([]int, 0, 10)
			m = append(m, troopix)
		}
		colorTroops[troop.Color()] = m
	}
	maxSum := 0
	for _, troopixs := range colorTroops {
		if len(troopixs) >= missingNo {
			sort.Sort(sort.Reverse(sort.IntSlice(troopixs)))
			sum := 0
			for _, troopix := range troopixs[:missingNo] {
				troop, _ := cards.DrTroop(troopix)
				sum = sum + troop.Value()
			}
			if sum > maxSum {
				maxSum = sum
			}
		}
	}
	if maxSum > 0 {
		rank = findBattalionRank(maxSum+moraleValue, combinations)
	}
	return rank
}
func findMissingTroops(missingNo int, troops []int, deckSet map[int]bool) (found bool) {
	count := 0
	for _, troopix := range troops {
		if deckSet[troopix] {
			count = count + 1
			if count == missingNo {
				found = true
				break
			}
		}
	}
	return found
}
func validJoker(troopix, moralix int) (valid bool) {
	troop, ok := cards.DrTroop(troopix)
	if ok {
		if moralix == cards.TC123 {
			if troop.Value() > 0 && troop.Value() < 4 {
				valid = true
			}
		} else if moralix == cards.TC8 {
			if troop.Value() == 8 {
				valid = true
			}
		} else if cards.IsLeader(moralix) {
			valid = true
		}
	}
	return valid
}
func findBattalionRank(strenght int, combinations []*combi.Combination) (rank int) {
	for _, v := range combinations {
		if v.Strength == strenght && v.Formation.Value == cards.FBattalion.Value {
			rank = v.Rank
			break
		}
	}
	return rank
}
func max2TroopValues(troops map[int]bool) (maxValue1, maxValue2 int) {
	for troopix := range troops {
		troop, _ := cards.DrTroop(troopix)
		if troop.Value() > maxValue1 {
			maxValue1 = troop.Value()
			maxValue2 = maxValue1
		} else if troop.Value() > maxValue2 {
			maxValue2 = troop.Value()
		}
		if maxValue2 == 10 {
			break
		}
	}
	return maxValue1, maxValue2
}
func maxTroopValue(troops map[int]bool) int {
	maxValue := 0
	for troopix := range troops {
		troop, _ := cards.DrTroop(troopix)
		if troop.Value() > maxValue {
			maxValue = troop.Value()
		}
		if maxValue == 10 {
			break
		}
	}
	return maxValue
}
func MoraleTroopsSum(flagTroops []int) (sum int) {
	for _, cardix := range flagTroops {
		troop, ok := cards.DrTroop(cardix)
		if ok {
			sum = sum + troop.Value()
		} else {
			sum = sum + cards.MoraleMaxValue(cardix)
		}
	}
	return sum
}

//maxSum calculate the maximum sum a flag posistion can reach without
//using morale cards from the hand.
//#deckValues
func maxSum(flagTroops, handCards, deckValues []int, mud bool) (sum int, updDeckValues []int) {
	formationSize := 3
	if mud {
		formationSize = 4
	}
	sum = MoraleTroopsSum(flagTroops)
	if len(flagTroops) < formationSize {
		max := false
		for _, cardix := range handCards {
			if cards.IsTroop(cardix) {
				deckValues, max = botdeck.MaxValuesUpd(cardix, deckValues)
				if max {
					break
				}
			}
		}

		for i := 0; i < formationSize-len(flagTroops); i++ {
			sum = sum + deckValues[i]
		}
	}
	updDeckValues = deckValues
	return sum, updDeckValues
}
func rankAnalyze(flagCards []int, handCards []int, drawSet map[int]bool,
	drawNo int, mud bool) (ranks []*combi.Analysis) {
	formationNo := 3
	if mud {
		formationNo = 4
	}
	combinations := combi.Combinations(formationNo)
	formationMade := len(flagCards) == formationNo
	ranks = make([]*combi.Analysis, len(combinations))
	allCombi := math.Comb(uint64(len(drawSet)), uint64(drawNo))
	for i, comb := range combinations {
		ana := combi.Ana(comb, flagCards, handCards, drawSet, drawNo, mud)
		ranks[i] = ana
		ana.All = allCombi
		if ana.Valid > 0 {
			ana.Prop = float64(ana.Valid) / float64(allCombi)
		}
		if formationMade && ana.Prop == 1 {
			break
		}
	}
	return ranks
}
