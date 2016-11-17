package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/battbot/combi"
	botdeck "github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battleline/cards"
	slice "github.com/rezder/go-slice/int"
)

type Analysis struct {
	Flagix         int
	TargetRank     int
	TargetSum      int
	OppTroopsNo    int
	BotTroopNo     int
	FormationSize  int
	FlagValue      int
	IsTargetRanked bool
	IsTargetMade   bool
	IsLost         bool
	IsNewFlag      bool
	IsPlayable     bool
	IsClaimed      bool
	IsFog          bool
	IsLoosingGame  bool
	Analysis       []*combi.Analysis
	SumCards       []int
	Flag           *Flag
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	combiAna := "<nil>"
	if ana.Analysis != nil {
		positive := make([]*combi.Analysis, 0, 0)
		for _, c := range ana.Analysis {
			if c.Prop > 0 {
				positive = append(positive, c)
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
	txt := fmt.Sprintf("{Flag:%v Target:%v FlagValue:%v TargetSum:%v Lost %v LooseGame:%v SumCards:%v Combination: %v", ana.Flagix+1, ana.TargetRank, ana.FlagValue, ana.TargetSum, ana.IsLost, ana.IsLoosingGame, ana.SumCards, combiAna)
	return txt
}

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
		fa.OppTroopsNo = len(flag.OppTroops)
		if fa.OppTroopsNo != 0 && !flag.IsFog() {
			fa.IsTargetRanked = true
			if len(flag.OppTroops) == flag.FormationSize() {
				fa.IsTargetMade = true
			}
			fa.TargetRank = calcMaxRank(flag.OppTroops, deck.OppHand(),
				deck.Troops(), deck.OppDrawNo(!isBotFirst), flag.IsMud())
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
					deck.Troops(), deck.BotDrawNo(isBotFirst), flag.IsMud())
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

func (ana *Analysis) IsWin() (isWin bool) {
	if !ana.IsLost {
		targetRank := ana.TargetRank
		if targetRank == 0 {
			targetRank = 100
		}
		if ana.IsTargetMade && ana.TargetRank != 0 {
			targetRank = targetRank - 1
		}
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
		if !isWin && ana.TargetRank == 0 {
			if len(ana.SumCards) != 0 {
				isWin = true //A win is not garantied if more than cards have to be played, but it is ok.
			}
		}

	}
	return isWin
}
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
func isNewFlagRankBeatable(
	deck *botdeck.Deck,
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
	combinations := combi.CombinationsMud(mud)
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
	if mud {
		formationNo = 4
	}
	combinations := combi.Combinations(formationNo)
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
