package flag

import (
	"github.com/rezder/go-battleline/battbot/combi"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battleline/cards"
	slice "github.com/rezder/go-slice/int"
)

type Analysis struct {
	Flagix               int
	TargetRank           int
	TargetSum            int
	OppTroopsNo          int
	BotTroopNo           int
	FormationSize        int
	IsTargetRanked       bool
	IsTargetMade         bool
	IsLost               bool
	IsNewFlag            bool
	IsPlayable           bool
	IsClaimed            bool
	IsFog                bool
	Analysis             []*combi.Analysis
	SumCards             []int
	KeepFlagTroopixs     map[int]bool
	KeepFlagHandTroopixs map[int]bool
}

func NewAnalysis(flag *Flag, botHandTroops []int, deckMaxValues []int, deck *botdeck.Deck, flagix int) (fa *Analysis) {
	fa = new(Analysis)
	fa.Flagix = flagix
	if flag.IsClaimed() {
		fa.IsClaimed = true
	} else {
		fa.OppTroopsNo = len(flag.OppTroops)
		if fa.OppTroopsNo != 0 && !flag.IsFog() {
			fa.IsTargetRanked = true
			if len(flag.OppTroops) == flag.FormationSize() {
				fa.IsTargetMade = true
			}
			fa.TargetRank = calcMaxRank(flag.OppTroops, deck.OppHand(),
				deck.Troops(), deck.OppDrawNo(), flag.IsMud())
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
		if fa.BotTroopNo == 0 {
			fa.IsNewFlag = true
		} else {
			if !flag.IsFog() {
				fa.Analysis = rankAnalyze(flag.PlayTroops, botHandTroops,
					deck.Troops(), deck.BotDrawNo(), flag.IsMud())
			}
		}
		botDeckValues := make([]int, len(deckMaxValues))
		copy(botDeckValues, deckMaxValues)

		cards, botMaxSum := sumCards(botDeckValues, flag.PlayTroops, botHandTroops,
			fa.TargetSum, flag.FormationSize())
		fa.SumCards = cards
		if fa.IsTargetMade {
			if fa.BotTroopNo != 0 {
				fa.IsLost = lost(fa.Analysis, fa.TargetRank, fa.TargetSum, botMaxSum)
			} else {
				fa.IsLost = lostNewFlag(fa.TargetRank, fa.TargetSum, botMaxSum, deck, botHandTroops, flag.IsMud())
			}
		}

	}

	return fa
}
func (ana *Analysis) AddKeepTroops(keepFlagTroopixs, keepFlagHandTroopixs map[int]bool) {
	ana.KeepFlagHandTroopixs = keepFlagHandTroopixs
	ana.KeepFlagTroopixs = keepFlagTroopixs
}
func HandAnalyze(handTroops []int, deck *botdeck.Deck) (ana map[int][]*combi.Analysis) {
	ana = make(map[int][]*combi.Analysis)
	for _, troopix := range handTroops {
		handCards := slice.Remove(handTroops, troopix)
		ana[troopix] = rankAnalyze([]int{troopix}, handCards, deck.Troops(), deck.BotDrawNo(), false)
	}
	return ana
}
func lost(analysis []*combi.Analysis, targetRank, targetSum, botMaxSum int) bool {
	isLost := false
	botRank := 0
	for _, ana := range analysis {
		if ana.Prop > 0 {
			botRank = ana.Comb.Rank
			break
		}
	}
	if targetRank > 0 {
		if botRank >= targetRank {
			isLost = true
		}
	} else { //skirmish line
		if botRank == 0 {
			if botMaxSum < targetSum {
				isLost = true
			}
		}
	}
	return isLost

}
func lostNewFlag(targetRank, targetSum, botMaxSum int, deck *botdeck.Deck, handTroopixs []int, isMud bool) (isLost bool) {
	if targetRank == 0 {
		if botMaxSum < targetSum {
			isLost = true
		} else {
			if !isNewFlagRankBeatable(deck, handTroopixs, 0, isMud) {
				isLost = true
			}
		}
	} else {
		if !isNewFlagRankBeatable(deck, handTroopixs, targetRank, isMud) {
			isLost = true
		}
	}
	return isLost
}
func isNewFlagRankBeatable(deck *botdeck.Deck,
	handTroopixs []int,
	targetRank int,
	isMud bool) (beatable bool) {
	if targetRank != 1 {
		flagCards := make([]int, 1)
		drawTroopixs := make([]int, 0, len(deck.Troops())+len(handTroopixs))
		for troopix := range deck.Troops() {
			drawTroopixs = append(drawTroopixs, troopix)
		}
		drawTroopixs = append(drawTroopixs, handTroopixs...)
		for _, troopix := range drawTroopixs {
			handixs := slice.Remove(drawTroopixs, troopix)
			flagCards[0] = troopix
			maxRank := calcMaxRank(flagCards, handixs, nil, 0, isMud)
			if maxRank < targetRank {
				beatable = true
				break
			}
		}
	}
	return beatable
}

//sumCards the troops playable in a skirmish line formation (sum).
//Its cards that give enough to reach the target or if target is not possible
//the higest cards that still exist.
//#deckValues
func sumCards(deckValues, flagTroops, botHandTroops []int, targetSum, formationSize int) (playableCards []int, botMaxSum int) {
	if len(flagTroops) == formationSize {
		botMaxSum = moralesTroopsSum(flagTroops)
	} else {
		botMaxSum, deckValues = maxSum(flagTroops, botHandTroops, deckValues, formationSize == 4)
		if targetSum >= botMaxSum {
			playableCards = make([]int, 0, len(botHandTroops))
			for _, cardix := range botHandTroops {
				troop, err := cards.DrTroop(cardix)
				if err == nil {
					if troop.Value() >= deckValues[formationSize-len(flagTroops)-1] {
						playableCards = append(playableCards, cardix)
					}
				}
			}
		} else {
			flagSum := moralesTroopsSum(flagTroops)
			playableCards = make([]int, 0, len(botHandTroops))
			needValue := targetSum - flagSum
			avgNeedValue := float32(needValue) / float32(formationSize-len(flagTroops))
			for _, cardIx := range botHandTroops {
				troop, err := cards.DrTroop(cardIx)
				if err == nil {
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

func calcMaxRank(flagCards []int, handCards []int, drawSet map[int]bool,
	drawNo int, mud bool) (rank int) {
	combinations := combi.Combinations3
	if mud {
		combinations = combi.Combinations4
	}
	for _, comb := range combinations {
		ana := combi.Ana(comb, flagCards, handCards, drawSet, drawNo, mud)
		if ana.Prop > 0 {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}
func moralesTroopsSum(flagTroops []int) (sum int) {
	for _, cardix := range flagTroops {
		troop, err := cards.DrTroop(cardix)
		if err == nil {
			sum = sum + troop.Value()
		} else {
			sum = sum + cards.MoraleMaxValue(cardix)
		}
	}
	return sum
}

//maxSum calculate the maximum sum a flag posistion can reach without using morale cards from the hand.
//#deckValues
func maxSum(flagTroops, handCards, deckValues []int, mud bool) (sum int, updDeckValues []int) {
	cardsNo := 3
	if mud {
		cardsNo = 4
	}
	sum = moralesTroopsSum(flagTroops)
	if len(flagTroops) < cardsNo {
		max := false
		for _, cardix := range handCards {
			if cards.IsTroop(cardix) {
				deckValues, max = botdeck.MaxValuesUpd(cardix, deckValues)
				if max {
					break
				}
			}
		}
		for i := 0; i < cardsNo-len(flagTroops); i++ {
			sum = sum + deckValues[i]
		}
	}
	updDeckValues = deckValues
	return sum, updDeckValues
}
func rankAnalyze(flagCards []int, handCards []int, drawSet map[int]bool,
	drawNo int, mud bool) (ranks []*combi.Analysis) {
	formationNo := 3
	combinations := combi.Combinations3
	if mud {
		combinations = combi.Combinations4
		formationNo = 4
	}
	formationMade := len(flagCards) == formationNo
	ranks = make([]*combi.Analysis, len(combinations))
	for i, comb := range combinations {
		ana := combi.Ana(comb, flagCards, handCards, drawSet, drawNo, mud)
		ranks[i] = ana
		if formationMade && ana.Prop == 1 {
			break
		}
	}
	return ranks
}
