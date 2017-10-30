package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-error/log"
	math "github.com/rezder/go-math/int"
)

// Analysis a analysis of the flag.
type Analysis struct {
	Flagix           int
	TargetRank       int
	TargetSum        int
	BotMaxRank       int
	BotMaxSum        int
	OppFormationSize int
	BotFormationSize int
	FormationSize    int
	FlagValue        int
	IsTargetMade     bool
	IsLost           bool
	IsWin            bool
	IsNewFlag        bool
	IsPlayable       bool
	IsClaimed        bool
	Claimer          int
	IsFog            bool
	IsLoosingGame    bool
	RankAnas         []*combi.Analysis
	Flag             *game.Flag
	Playix           int
	//TODO maybe add HostRank
}

func (ana *Analysis) String() string {
	if ana == nil {
		return "<nil>"
	}
	combiAna := "<nil>"
	if ana.RankAnas != nil {
		positive := make([]*combi.Analysis, 0, 0)
		for _, c := range ana.RankAnas {
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
	txt := fmt.Sprintf("{Flagix:%v TargetRank:%v BotMaxRank:%v FlagValue:%v IsLost:%v LooseGame:%v Combination:%v",
		ana.Flagix+1, ana.TargetRank, ana.BotMaxRank, ana.FlagValue, ana.IsLost, ana.IsLoosingGame, combiAna)
	return txt
}

// NewAnalysis create a flag analysis.
func NewAnalysis(
	playix int,
	flag *game.Flag,
	botHand *card.Cards,
	oppHand *card.Cards,
	deckMaxStrenghts []int,
	deck *Deck,
	flagix int,
	isBotFirst bool) (fa *Analysis) {
	fa = new(Analysis)
	oppix := opp(playix)
	fa.Playix = playix
	fa.Flagix = flagix
	fa.Flag = flag
	fa.Claimer = flag.ConePos.Winner()
	if flag.IsWon {
		fa.IsClaimed = true
	} else {
		fa.OppFormationSize = flag.PlayerFormationSize(oppix)
		if flag.PlayerFormationSize(oppix) == flag.FormationSize() {
			fa.IsTargetMade = true
		}
		oppMissNo := flag.FormationSize() - flag.PlayerFormationSize(oppix)
		oppFlagStr := MoraleTroopsSum(flag.Players[oppix].Troops, flag.Players[oppix].Morales)
		var isOppOK bool
		fa.TargetSum, _, isOppOK = combi.HostMaxHandDeck(oppMissNo, oppFlagStr, deckMaxStrenghts, oppHand.Troops)
		if !isOppOK {
			panic("did not expect this") //TODO remove when confirmed
		}
		if fa.IsTargetMade {
			fa.TargetSum = fa.TargetSum + 1
		}
		if len(flag.Players[oppix].Troops) > 0 {
			fa.TargetRank = CalcMaxRank(flag.Players[oppix].Troops,
				flag.Players[oppix].Morales, oppHand.Troops,
				deck.Troops(), deckMaxStrenghts, deck.OppDrawNo(!isBotFirst),
				flag.FormationSize(), flag.IsFog, fa.TargetSum, combi.HostRank(flag.FormationSize()))
		} else {
			fa.TargetRank = calcMaxRankNewFlag(
				flag.Players[oppix].Morales,
				oppHand.Troops,
				deck.Troops(),
				flag.FormationSize(), flag.IsFog)
		}

		fa.IsFog = flag.IsFog
		fa.FormationSize = flag.FormationSize()
		fa.BotFormationSize = flag.PlayerFormationSize(playix)
		if fa.BotFormationSize < flag.FormationSize() {
			fa.IsPlayable = true // this is only morale or troop
		}
		botMoraleNo := len(flag.Players[playix].Morales)
		if fa.BotFormationSize == botMoraleNo {
			fa.IsNewFlag = true //TODO what if moral 8 or 1,2,3 is alone then handAna is of no use.
			// It could only happen if moral card was played as second card and traitor need the
			// troop.
			fa.BotMaxRank = calcMaxRankNewFlag(flag.Players[playix].Morales,
				botHand.Troops, deck.Troops(), flag.FormationSize(), flag.IsFog)
		} else {
			fa.RankAnas = rankAnalyze(flag.Players[playix].Troops, flag.Players[playix].Morales,
				botHand.Troops, deck.Troops(), deckMaxStrenghts, deck.BotDrawNo(isBotFirst), flag.FormationSize(), flag.IsFog, fa.TargetRank, fa.TargetSum)
			fa.BotMaxRank = calcBotMaxRank(fa.RankAnas)
		}
		botFlagStr := MoraleTroopsSum(flag.Players[playix].Troops, flag.Players[playix].Morales)
		botMissNo := flag.FormationSize() - flag.PlayerFormationSize(playix)
		var isBotOK bool
		fa.BotMaxSum, _, isBotOK = combi.HostMaxHandDeck(botMissNo, botFlagStr, deckMaxStrenghts, botHand.Troops)
		if !isBotOK {
			panic("did not expect this") //TODO remove when confirmed
		}

		if fa.IsTargetMade {
			fa.IsLost = lost(fa.TargetRank, fa.TargetSum, fa.BotMaxSum, fa.BotMaxRank, fa.FormationSize)
		}
		if !fa.IsLost {
			fa.IsWin = isWin(fa.TargetRank, fa.TargetSum, flag.IsFog, fa.IsTargetMade, fa.RankAnas, flag.Players[playix].Troops,
				flag.Players[playix].Morales, fa.BotFormationSize, fa.FormationSize)
		}
		log.Printf(log.Debug, "Flagix:%v Flag:%v Lost:%v BotRank:%v BotSum:%v TargetRank:%v TargetSum:%v\n", fa.Flagix, fa.Flag, fa.IsLost, fa.BotMaxRank, fa.BotMaxSum, fa.TargetRank, fa.TargetSum)
	}

	return fa
}

func calcBotMaxRank(combisAna []*combi.Analysis) (rank int) {
	for _, ana := range combisAna {
		if ana.Prop > 0 || ana.Comb.Formation.Value == card.FHost.Value {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}

func isWin(
	anaTargetRank, targetSum int,
	isFog, isTargetMade bool,
	combAnas []*combi.Analysis,
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	botFormationSize, formationSize int) (isWin bool) {

	for _, combiAna := range combAnas {
		if combiAna.Prop == 1 {
			if (!isTargetMade && combiAna.Comb.Rank <= anaTargetRank) ||
				(isTargetMade && combiAna.Comb.Rank < anaTargetRank) ||
				(isTargetMade && combiAna.Comb.Rank == anaTargetRank && combiAna.Comb.Rank == combi.HostRank(formationSize)) {
				isWin = true
			}
			break
		}
		if combiAna.Comb.Rank > anaTargetRank {
			break
		}
	}
	return isWin
}
func opp(ix int) (oppix int) {
	oppix = ix + 1
	if oppix > 1 {
		oppix = 0
	}
	return oppix
}

// HandAnalyze create a rank analysis of each troop on the hand.
// it assume no/ ignore enviroment cards,
// and targetSum = 0 this means can use Prob or Playables on combination Host
func HandAnalyze(
	handTroops []card.Troop,
	deck *Deck,
	deckMaxStrenghts []int,
	isBotFirst bool) (ana map[card.Troop][]*combi.Analysis) {
	ana = make(map[card.Troop][]*combi.Analysis)
	targetRank := 1
	targetSum := 0
	for _, troop := range handTroops {
		simHandTroops := make([]card.Troop, 0, len(handTroops))
		for _, simTroop := range handTroops {
			if troop != simTroop {
				simHandTroops = simTroop.AppendStrSorted(simHandTroops)
			}
		}
		ana[troop] = rankAnalyze([]card.Troop{troop}, nil, simHandTroops, deck.Troops(), deckMaxStrenghts, deck.BotDrawNo(isBotFirst), 3, false, targetRank, targetSum)
	}
	return ana
}

//lost calculate if a flag is lost it assume opponent moves first and win
//when rank or sum is equal
func lost(targetRank, targetSum, botMaxSum, botRank, formationSize int) bool {
	isLost := false

	if targetRank < botRank || (targetRank == botRank && targetRank != combi.HostRank(formationSize)) {
		isLost = true
	} else if targetRank == botRank &&
		targetRank == combi.HostRank(formationSize) &&
		botMaxSum < targetSum { //target Sum has been increased with one because it was made
		isLost = true
	}

	return isLost

}

// CalcMaxRank calculate the maximum rank a
// play may reach on a flag.
func CalcMaxRank(
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	handTroops []card.Troop,
	drawSet map[card.Troop]bool,
	deckStrenghts []int,
	drawNo int,
	formationSize int,
	isFog bool,
	targetRank, targetSum int,
) (rank int) {

	combinations := combi.Combinations(formationSize)
	allCombi := math.Comb(uint64(len(drawSet)), uint64(drawNo))
	for _, comb := range combinations {
		ana := combi.Ana(comb, flagTroops, flagMorales, handTroops, drawSet, deckStrenghts, drawNo, formationSize, isFog, targetRank, targetSum)
		ana.SetAll(allCombi)
		if ana.Prop > 0 || comb.Formation.Value == card.FHost.Value {
			rank = ana.Comb.Rank
			break
		}
	}
	return rank
}
func calcMaxRankNewFlag(
	flagMorales []card.Morale,
	handTroops []card.Troop,
	drawSet map[card.Troop]bool,
	formationSize int,
	isFog bool,
) (rank int) {
	if isFog {
		rank = combi.HostRank(formationSize)
	} else {
		combinations := combi.Combinations(formationSize)
		moraleNo := len(flagMorales)
		flagTroopStrenght := 0
		for _, morale := range flagMorales {
			flagTroopStrenght = flagTroopStrenght + morale.MaxStrenght()
		}
		if moraleNo == formationSize {
			rank = findBattalionRank(flagTroopStrenght, combinations)
		} else {
			allDrawSet := make(map[card.Troop]bool)
			for i, v := range drawSet {
				allDrawSet[i] = v
			}
			for _, v := range handTroops {
				allDrawSet[v] = true
			}
			//TODO this does not work for two morales as they can make a wedge and max2TroopStrenghts
			//should be per color not all cards.
			if moraleNo == 3 || moraleNo == 2 && formationSize == 3 {
				maxValue := maxTroopStrenght(allDrawSet)
				rank = findBattalionRank(flagTroopStrenght+maxValue, combinations)
			} else if moraleNo == 2 {
				mv1, mv2 := max2TroopStrenghts(allDrawSet)
				rank = findBattalionRank(mv1+mv2+flagTroopStrenght, combinations)
			} else if moraleNo == 1 {
				rank = calcMaxRankNewFlagOneMoral(flagMorales[0], allDrawSet, formationSize, combinations)
			} else {
				rank = calcMaxRankNewFlagZeroMoral(allDrawSet, formationSize, combinations)
			}
		}
	}
	return rank
}
func calcMaxRankNewFlagZeroMoral(
	allDrawSet map[card.Troop]bool,
	formationSize int,
	combinations []*combi.Combination) (rank int) {
	calcBatt := false
Loop:
	for _, comb := range combinations {
		switch comb.Formation.Value {
		case card.FWedge.Value:
			for _, troops := range comb.Troops {
				made := true
				for _, troop := range troops {
					if !allDrawSet[troop] {
						made = false
						break
					}
				}
				if made {
					rank = comb.Rank
					break Loop
				}
			}
		case card.FPhalanx.Value:
			if findMissingTroops(formationSize, comb.Troops[combi.COLNone], allDrawSet) {
				rank = comb.Rank
				break Loop
			}
		case card.FBattalion.Value:
			if !calcBatt {
				calcBatt = true
				missingNo := formationSize
				moraleValue := 0
				rank = calcBestBattalionRank(missingNo, moraleValue, allDrawSet, combinations)
				if rank != 0 {
					break Loop
				}
			}

		case card.FSkirmish.Value:
			strTroops := sortTroopStr(comb.Troops[combi.COLNone])
			failed := false
			for _, troops := range strTroops {
				found := false
				for _, troop := range troops {
					if allDrawSet[troop] {
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
		case card.FHost.Value:
			rank = comb.Rank
		}
	}
	return rank
}
func sortTroopStr(troops []card.Troop) (m map[int][]card.Troop) {
	return sortTroopsStrengthOrColor(troops, false)
}
func sortTroopsStrengthOrColor(troops []card.Troop, isColor bool) (valueTroops map[int][]card.Troop) {
	valueTroops = make(map[int][]card.Troop)
	for _, troop := range troops {
		value := 0
		if isColor {
			value = troop.Color()
		} else {
			value = troop.Strenght()
		}
		m, ok := valueTroops[value]
		if ok {
			m = append(m, troop)
		} else {
			m = make([]card.Troop, 0, 6)
			m = append(m, troop)
		}
		valueTroops[value] = m
	}
	return valueTroops
}
func calcMaxRankNewFlagOneMoral(
	morale card.Morale,
	allDrawSet map[card.Troop]bool,
	formationSize int,
	combinations []*combi.Combination) (rank int) {
	calcBatt := false
Loop:
	for _, comb := range combinations {
		switch comb.Formation.Value {
		case card.FWedge.Value:
			for _, troops := range comb.Troops {
				made := true
				usedJoker := false
				for _, troop := range troops {
					if !allDrawSet[troop] {
						if (!usedJoker) && morale.ValidStrenght(troop.Strenght()) {
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
		case card.FPhalanx.Value:
			if morale.ValidStrenght(comb.Troops[combi.COLNone][0].Strenght()) {
				missingNo := formationSize - 1
				if findMissingTroops(missingNo, comb.Troops[combi.COLNone], allDrawSet) {
					rank = comb.Rank
					break Loop
				}
			}

		case card.FBattalion.Value:
			if !calcBatt {
				calcBatt = true
				missingNo := formationSize - 1
				rank = calcBestBattalionRank(missingNo, morale.MaxStrenght(), allDrawSet, combinations)
				if rank != 0 {
					break Loop
				}
			}

		case card.FSkirmish.Value: //TODO may be faster if sort deck after strenght map[str][]card.Troop and do like battalion
			valueTroops := sortTroopStr(comb.Troops[combi.COLNone])
			failed := false
			usedJoker := false
			for _, troops := range valueTroops {
				found := false
				for _, troop := range troops {
					if allDrawSet[troop] {
						found = true
						break
					}
				}
				if !found {
					if (!usedJoker) && morale.ValidStrenght(troops[0].Strenght()) {
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
		case card.FHost.Value:
			rank = comb.Rank
		}
	}
	return rank
}
func calcBestBattalionRank(
	missingNo, moraleValue int,
	deckSet map[card.Troop]bool,
	combinations []*combi.Combination) (rank int) {

	colorTroops := make(map[int][]card.Troop)
	for troop := range deckSet {
		troops, ok := colorTroops[troop.Color()]
		if ok {
			troops = troop.AppendStrSorted(troops)
		} else {
			troops = make([]card.Troop, 0, 10)
			troops = append(troops, troop)
		}
		colorTroops[troop.Color()] = troops
	}
	maxSum := 0
	for _, troops := range colorTroops {
		if len(troops) >= missingNo {
			sum := 0
			for _, troop := range troops[:missingNo] {
				sum = sum + troop.Strenght()
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
func findMissingTroops(missingNo int, troops []card.Troop, deckSet map[card.Troop]bool) (found bool) {
	count := 0
	for _, troop := range troops {
		if deckSet[troop] {
			count = count + 1
			if count == missingNo {
				found = true
				break
			}
		}
	}
	return found
}

func findBattalionRank(strenght int, combinations []*combi.Combination) (rank int) {
	for _, v := range combinations {
		if v.Strength == strenght && v.Formation.Value == card.FBattalion.Value {
			rank = v.Rank
			break
		}
	}
	return rank
}
func max2TroopStrenghts(troops map[card.Troop]bool) (maxStr1, maxStr2 int) {
	for troop := range troops {
		if troop.Strenght() > maxStr1 {
			maxStr1 = troop.Strenght()
			maxStr2 = maxStr1
		} else if troop.Strenght() > maxStr2 {
			maxStr2 = troop.Strenght()
		}
		if maxStr2 == 10 {
			break
		}
	}
	return maxStr1, maxStr2
}
func maxTroopStrenght(troops map[card.Troop]bool) int {
	maxStrenght := 0
	for troop := range troops {
		if troop.Strenght() > maxStrenght {
			maxStrenght = troop.Strenght()
		}
		if maxStrenght == 10 {
			break
		}
	}
	return maxStrenght
}

//MoraleTroopsSum sums the strenght of troops and morales.
//Morales use the maxStrenght.
func MoraleTroopsSum(flagTroops []card.Troop, flagMorales []card.Morale) (sum int) {
	for _, troop := range flagTroops {
		sum = sum + troop.Strenght()
	}
	for _, morale := range flagMorales {
		sum = sum + morale.MaxStrenght()
	}
	return sum
}

func rankAnalyze(
	flagTroops []card.Troop,
	flagMorales []card.Morale,
	handTroops []card.Troop,
	drawSet map[card.Troop]bool,
	deckStrenghts []int,
	drawNo, formationSize int,
	isFog bool,
	targetRank, targetSum int,
) (ranks []*combi.Analysis) {
	combinations := combi.Combinations(formationSize)
	formationMade := len(flagTroops)+len(flagMorales) == formationSize
	ranks = make([]*combi.Analysis, len(combinations))
	allCombi := math.Comb(uint64(len(drawSet)), uint64(drawNo))
	for i, comb := range combinations {
		ana := combi.Ana(comb, flagTroops, flagMorales, handTroops, drawSet, deckStrenghts, drawNo, formationSize, isFog, targetRank, targetSum)
		ana.SetAll(allCombi)
		ranks[i] = ana
		if formationMade && ana.Prop == 1 {
			break
		}
		//TODO make prob and playables for host, and include isFog in rankAnalyze.
		//Slight change to isWin(): == for host possible. and the isLost can reuse the rankAnalyze but
		//it must also work without.First make without prob calc just set .5 for undesided.
	}
	return ranks
}
