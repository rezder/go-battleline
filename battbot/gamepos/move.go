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
	"strconv"
)

func makeMoveClaim(moves []bat.Move) (moveix int) {
	max := -1
	allix := -1
	for i, mv := range moves {
		mc, _ := mv.(bat.MoveClaim)
		if len(mc.Flags) > max {
			allix = i
			max = len(mc.Flags)
		}
	}
	return allix
}
func makeMoveDeck(pos *Pos) (moveix int) {
	move := *bat.NewMoveDeck(bat.DECK_TROOP)
	if len(pos.playHand.Tacs) == 0 {
		botNo := playableTacNoBot(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)
		if botNo > 0 || pos.deck.OppTacNo() > 0 {
			flagsAna := analyzeFlags(pos.flags, pos.playHand.Troops, pos.deck)
			handAna := flag.HandAnalyze(pos.playHand.Troops, pos.deck)
			analyzeFlagsAddKeep(flagsAna, handAna)
			if len(flagsAna[0].KeepFlagHandTroopixs) > 3 {
				move = *bat.NewMoveDeck(bat.DECK_TAC)
			}
		}

	}
	if move.Deck == bat.DECK_TROOP && pos.deck.DeckTroopNo() == 0 {
		move = *bat.NewMoveDeck(bat.DECK_TAC)
	}
	if move.Deck == bat.DECK_TAC && pos.deck.DeckTacNo() == 0 {
		move = *bat.NewMoveDeck(bat.DECK_TROOP)
	}

	moveix = findMoveIndex(pos.turn.Moves, move)

	return moveix
}
func makeMoveScoutReturn(pos *Pos) (moveix int) {
	var tacs []int
	troops := []int{1, 2}
	move := bat.NewMoveScoutReturn(tacs, troops)
	//TODO Scout return strategi
	pos.playHand.PlayMulti(move.Tac)
	pos.playHand.PlayMulti(move.Troop)
	pos.deck.PlayScoutReturn(move.Troop, move.Tac)
	moveix = findMoveIndex(pos.turn.Moves, move)

	return moveix
}
func makeMoveHand(pos *Pos) (moveixs [2]int) {
	if cerrors.LogLevel() == cerrors.LOG_Debug {
		log.Printf("Hand: %v\n", pos.playHand)
	}
	flagsAna := analyzeFlags(pos.flags, pos.playHand.Troops, pos.deck)
	handAna := flag.HandAnalyze(pos.playHand.Troops, pos.deck)
	analyzeFlagsAddKeep(flagsAna, handAna)

	if pos.turn.MovesPass {
		_, _, botLeader, _ := playableTac(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)
		cardix, move := lostFlagTacticMove(flagsAna, pos.playHand.Tacs, botLeader)
		if cardix != 0 {
			moveixs[1] = findMoveHandIndex(pos.turn.MovesHand, cardix, move)
			moveixs[0] = cardix
		} else {
			moveixs[0] = 0
			moveixs[1] = bat.SM_Pass

		}
	} else {
		cardix := 0
		var move bat.Move
		botTacNo, _, botLeader, _ := playableTac(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)
		if botTacNo > 0 && len(pos.playHand.Tacs) > 0 {
			cardix, move = lostFlagTacticMove(flagsAna, pos.playHand.Tacs, botLeader)
		}
		if cardix == 0 {
			cardix, move = lostFlagDumpMove(flagsAna, handAna, pos.playHand.Troops)
		}
		if cardix == 0 {
			cardix, move = prioritizedMove(flagsAna, handAna, pos.playHand)
		}
		moveixs[0] = cardix
		moveixs[1] = findMoveHandIndex(pos.turn.MovesHand, cardix, move)
	}
	return moveixs
}

func analyzeFlags(flags [bat.FLAGS]*flag.Flag, botHandTroops []int, deck *botdeck.Deck) (
	flagsAna map[int]*flag.Analysis) {
	flagsAna = make(map[int]*flag.Analysis)
	deckMaxValues := deck.MaxValues()
	for flagix, iflag := range flags {
		flagsAna[flagix] = flag.NewAnalysis(iflag, botHandTroops, deckMaxValues, deck, flagix)
	}
	return flagsAna
}

// anaFlagsAddKeep adds keep troops maps to all flags analysis
//#flagsAna
func analyzeFlagsAddKeep(flagsAna map[int]*flag.Analysis, handAna map[int][]*combi.Analysis) {
	keepFlag := make(map[int]bool)
	keepFlagHand := make(map[int]bool)
	for _, flagAna := range flagsAna {
		if flagAna.Analysis != nil {
			keepTroops(flagAna.Analysis, keepFlag)
			keepTroops(flagAna.Analysis, keepFlagHand)
		}
	}

	for _, combiAna := range handAna {
		keepTroops(combiAna, keepFlagHand)
	}
	for _, flagAna := range flagsAna {
		flagAna.AddKeepTroops(keepFlag, keepFlagHand)
	}
}
func prioritizedMove(
	flagsAna map[int]*flag.Analysis,
	handAna map[int][]*combi.Analysis,
	hand *bat.Hand) (cardix int, move bat.Move) {

	prioritizedFlagixs := prioritizeFlags(flagsAna)

	cardix, move = pri3CardsMove(prioritizedFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = pri2CardsMove(prioritizedFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = pri1CardMove(prioritizedFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priFlagLoop(prioritizedFlagixs, flagsAna, pffFog)
	if cardix != 0 {
		return cardix, move
	}
	if cardix != 0 {
		cardix, move = priFlagHandLoop(prioritizedFlagixs, flagsAna, handAna, pfhfNewFlagRanked)
	}
	cardix, move = newFlagMove(prioritizedFlagixs, flagsAna, handAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priTacticMove(prioritizedFlagixs, flagsAna, hand.Tacs)
	if cardix != 0 {
		return cardix, move
	}

	cardix, move = priFlagNLoop(prioritizedFlagixs, flagsAna, 1, pfnf2Pick1Card)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priFlagLoop(prioritizedFlagixs, flagsAna, pffSum)

	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priDumpMove(prioritizedFlagixs, flagsAna, hand.Troops)
	return cardix, move

}
func lostFlagTacticMove(
	flagsAna map[int]*flag.Analysis,
	handTacs []int,
	botLeader bool) (cardix int, move bat.Move) {
	//TODO Tactic
	return cardix, move
}

//lostFlagDumpMove handle the strategi of dumbing card on lost flag.
//Dump the smallest card if it exist of card not usable in the best formation or in case
//of the best formation being wedge also the phalanx.
func lostFlagDumpMove(
	flagsAna map[int]*flag.Analysis,
	handAna map[int][]*combi.Analysis,
	handTroops []int) (cardix int, move bat.Move) {
	lostFlagix := -1
	for _, ana := range flagsAna {
		if ana.IsPlayable && ana.IsLost {
			lostFlagix = ana.Flagix
			break
		}
	}
	if lostFlagix != -1 {
		cardix = keepLowestValue(handTroops, flagsAna[0].KeepFlagHandTroopixs)
		if cardix != 0 {
			if cerrors.LogLevel() == cerrors.LOG_Debug {
				log.Println("Made a Lost Flag Dump move")
			}
			move = *bat.NewMoveCardFlag(lostFlagix)
		}
	}
	return cardix, move
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

//keepTroops keeps all cards from the top formation, incase of wedge
//formation the phalanx formations is also included.
func keepTroops(combiAnas []*combi.Analysis, keepTroops map[int]bool) {
	cutOffFormationValue := 0
	for _, combiAna := range combiAnas {
		if combiAna != nil {
			if cutOffFormationValue == 0 {
				if combiAna.Prop > 0 {
					cutOffFormationValue = combiAna.Comb.Formation.Value
					if cards.FWedge.Value == cutOffFormationValue {
						cutOffFormationValue = cards.FPhalanx.Value
					}
				}
			}
			if combiAna.Comb.Formation.Value < cutOffFormationValue {
				break
			} else {
				for _, troopix := range combiAna.HandCardixs {
					keepTroops[troopix] = true
				}
			}
		} else {
			break
		}

	}
}
func prioritizeFlags(flagsAna map[int]*flag.Analysis) (flagixs []int) {
	flagValues := make([]int, bat.FLAGS)
	for _, ana := range flagsAna {
		if ana.IsPlayable {
			flagValues[ana.Flagix] = flagValues[ana.Flagix] + ana.OppTroopsNo
			if ana.Flagix == 0 || ana.Flagix == 8 {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 10
			} else if ana.Flagix == 1 || ana.Flagix == 7 {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 20
			} else {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 30
			}
			if ana.IsClaimed {
				if ana.Flagix+1 < 9 {
					flagValues[ana.Flagix+1] = flagValues[ana.Flagix+1] + 40
				}
				if ana.Flagix-1 >= 0 {
					flagValues[ana.Flagix+1] = flagValues[ana.Flagix+1] + 40
				}
				if ana.Flagix+2 < 9 {
					flagValues[ana.Flagix+1] = flagValues[ana.Flagix+1] + 30
				}
				if ana.Flagix-2 >= 0 {
					flagValues[ana.Flagix+1] = flagValues[ana.Flagix+1] + 30
				}
			}
		}
	}
	ixs := slice.SortWithIx(flagValues)
	flagixs = make([]int, 0, len(ixs))
	for i := len(ixs) - 1; i >= 0; i-- {
		if flagValues[i] > 0 {
			flagixs = append(flagixs, ixs[i])
		} else {
			break
		}
	}
	return flagixs
}
func priNCardsMove(n int, flagixs []int, flagsAna map[int]*flag.Analysis) (troopix int, move bat.Move) {
	troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfBestWinCard)
	if troopix == 0 {
		troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfTopRankCard)
	}
	if troopix == 0 {
		troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfButtomWedge)
	}
	return troopix, move
}

func pfnfBestWinCard(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a n flag cards move: Best wining card."
	if !flagAna.IsFog && flagAna.IsTargetRanked {
		targetRank := flagAna.TargetRank
		if flagAna.IsTargetMade {
			targetRank = targetRank - 1
		}
		cardSumProp := make(map[int]float64)
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Comb.Rank <= targetRank && combiAna.Prop > 0 {
				for _, troopix := range combiAna.HandCardixs {
					cardSumProp[troopix] = cardSumProp[troopix] + combiAna.Prop
				}
			}
		}
		troopix = findMaxSum(cardSumProp)
	}
	return troopix, logTxt
}

func findMaxSum(cardSumProp map[int]float64) (troopix int) {
	maxSumProp := float64(0)
	for cix, sumProp := range cardSumProp {
		if sumProp > maxSumProp {
			troopix = cix
			maxSumProp = sumProp
		}
	}
	return troopix
}
func pfnfTopRankCard(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a n flag cards move: Top ranked card"
	if !flagAna.IsFog {
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Prop > 0 {
				troopix = maxTroop(combiAna.HandCardixs)
				break
			}
			if combiAna.Comb.Formation.Value < cards.FPhalanx.Value {
				break
			}
		}
	}
	return troopix, logTxt
}

func maxTroop(troopixs []int) (troopix int) {
	troopix, _ = maxMinTroop(troopixs)
	return troopix
}
func minTroop(troopixs []int) (troopix int) {
	_, troopix = maxMinTroop(troopixs)
	return troopix
}
func maxMinTroop(troopixs []int) (maxTroopix, minTroopix int) {
	maxValue := 0
	minValue := 0
	for _, troopix := range troopixs {
		troop, _ := cards.DrTroop(troopix)
		if troop.Value() > maxValue {
			maxValue = troop.Value()
			maxTroopix = troopix
		}
		if troop.Value() < minValue {
			minValue = troop.Value()
			minTroopix = troopix
		}
	}
	return maxTroopix, minTroopix
}
func pfnfButtomWedge(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a n flag cards move: Buttom wedge card"
	if !flagAna.IsTargetMade && !flagAna.IsFog {
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Comb.Formation.Value == cards.FWedge.Value {
				if combiAna.Prop == 1 {
					troopix = maxTroop(combiAna.HandCardixs)
					break
				}
			} else {
				break
			}

		}
	}
	return troopix, logTxt
}

func pri3CardsMove(flagixs []int, flagsAna map[int]*flag.Analysis) (troopix int, move bat.Move) {
	troopix, move = priNCardsMove(3, flagixs, flagsAna)
	return troopix, move
}
func pri2CardsMove(flagixs []int, flagsAna map[int]*flag.Analysis) (troopix int, move bat.Move) {
	troopix, move = priNCardsMove(2, flagixs, flagsAna)
	return troopix, move
}
func pri1CardMove(flagixs []int, flagsAna map[int]*flag.Analysis) (troopix int, move bat.Move) {
	n := 1
	troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfBestWinCard)
	if troopix == 0 {
		troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfTopRankCard)
		if troopix == 0 {
			troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfWedgeConnector)
			if troopix == 0 {
				troopix, move = priFlagNLoop(flagixs, flagsAna, n, pfnfMadePhalanx)
			}
		}
	}
	return troopix, move
}
func pfnfWedgeConnector(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a wedge connector move"
	if !flagAna.IsFog {
		cardWedges := make(map[int]int)
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Comb.Formation.Value == cards.FWedge.Value {
				if len(combiAna.HandCardixs) > 0 {
					for _, combiTroopix := range combiAna.HandCardixs {
						cardWedges[combiTroopix] = cardWedges[combiTroopix] + 1
					}
				}
			} else {
				break
			}
		}
		for combiTroopix, wedgeNo := range cardWedges {
			if wedgeNo == flagAna.FormationSize-1 {
				troopix = combiTroopix
				break
			}
		}
	}
	return troopix, logTxt
}

func pfnfMadePhalanx(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a maded phalanx or higher move"
	if !flagAna.IsTargetMade && !flagAna.IsFog {
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Comb.Formation.Value == cards.FPhalanx.Value {
				if combiAna.Prop == 1 && len(combiAna.HandCardixs) != 0 {
					troopix = combiAna.HandCardixs[0]
					break
				}

			} else if combiAna.Comb.Formation.Value < cards.FPhalanx.Value {
				break
			}
		}
	}
	return troopix, logTxt
}
func pffFog(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a fog move"
	if flagAna.IsFog && len(flagAna.SumCards) > 0 {
		troopix = minTroop(flagAna.SumCards)
	}
	return troopix, logTxt
}

func pfhfNewFlagRanked(flagAna *flag.Analysis, handAna map[int][]*combi.Analysis) (troopix int, logTxt string) {
	logTxt = "Made a new flag rank move"
	if flagAna.IsNewFlag && flagAna.IsTargetRanked {
		targetRank := flagAna.TargetRank
		if flagAna.IsTargetMade {
			targetRank = targetRank - 1
		}
		troopix = newFlagTargetMove(handAna, targetRank)
	}
	return troopix, logTxt
}
func newFlagMove(flagixs []int,
	flagsAna map[int]*flag.Analysis,
	handAna map[int][]*combi.Analysis) (troopix int, move bat.Move) {

	nfTroopix, logTxt := newFlagMadeCombiMove(handAna)
	if nfTroopix == 0 {
		nfTroopix, logTxt = newFlagPhalanxMove(handAna)
	}
	if nfTroopix == 0 {
		nfTroopix, logTxt = newFlagHigestRankMove(handAna, flagsAna[0].KeepFlagTroopixs)
	}

	if nfTroopix != 0 {
		flagix := newFlagSelectFlag(nfTroopix, flagixs, flagsAna)
		if flagix != -1 {
			troopix = nfTroopix
			if cerrors.LogLevel() == cerrors.LOG_Debug {
				log.Println(logTxt)
				log.Printf("Cardix: %v", troopix)
				log.Printf("Prioritised flags: %v\n", flagixs)
			}
			move = *bat.NewMoveCardFlag(flagix)
		}
	}

	return troopix, move

}
func newFlagSelectFlag(troopix int, flagixs []int, flagsAna map[int]*flag.Analysis) (flagix int) {
	flagix = -1
	newFlagixs := make([]int, 0, len(flagixs))
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && !flagAna.IsTargetRanked {
			newFlagixs = append(newFlagixs, flagix)
		}
	}
	if len(newFlagixs) != 0 {
		if len(newFlagixs) == 1 {
			flagix = newFlagixs[0]
		} else {
			troop, _ := cards.DrTroop(troopix)
			troopValue := troop.Value()
			switch {
			case troopValue > 7:
				flagix = newFlagixs[0]
			case troopValue < 8 && troopValue > 5:
				flagix = newFlagixs[len(newFlagixs)/2]
			default:
				flagix = newFlagixs[len(newFlagixs)-1]
			}
		}
	}
	return flagix
}
func newFlagPhalanxMove(handAna map[int][]*combi.Analysis) (troopix int, logTxt string) {

	logTxt = "Made a new flag best phalanx or higher move"
	targetRank := combi.LastFormationRank(cards.FBattalion, 3)
	troopix = newFlagTargetMove(handAna, targetRank)
	return troopix, logTxt
}
func newFlagTargetMove(handAna map[int][]*combi.Analysis, targetRank int) (troopix int) {

	troopSumProp := make(map[int]float64)
	for handTroopix, combiAnas := range handAna {
		for _, combiAna := range combiAnas {
			if combiAna.Prop > 0 && combiAna.Comb.Rank <= targetRank {
				troopSumProp[handTroopix] = troopSumProp[handTroopix] + combiAna.Prop
			}
		}
	}
	troopix = findMaxSum(troopSumProp)
	return troopix
}

func newFlagMadeCombiMove(handAna map[int][]*combi.Analysis) (troopix int, logTxt string) {

	logTxt = "Made a new flag made combi move"
	targetRank := combi.LastFormationRank(cards.FPhalanx, 3)
HandLoop:
	for handTroopix, combiAnas := range handAna {
		for _, combiAna := range combiAnas {
			if combiAna.Prop == 1 && combiAna.Comb.Rank <= targetRank {
				troopix = handTroopix
				break HandLoop
			}
		}
	}

	return troopix, logTxt
}

func newFlagHigestRankMove(handAna map[int][]*combi.Analysis, keepFlagTroopixs map[int]bool) (troopix int, logTxt string) {
	logTxt = "Made a new flag highest rank move"
	topRank := 10000
	for handTroopix, combiAnas := range handAna {
		if !keepFlagTroopixs[handTroopix] {
			for _, combiAna := range combiAnas {
				if combiAna.Prop > 0 {
					if combiAna.Comb.Rank < topRank {
						troopix = handTroopix
						topRank = combiAna.Comb.Rank
					}
					break
				}
			}
		}
	}

	return troopix, logTxt
}

func priTacticMove(flagixs []int, flagsAna map[int]*flag.Analysis, handTacs []int) (cardix int, move bat.Move) {
	//TODO Tactic
	return cardix, move
}

func pfnf2Pick1Card(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Made 2 pick on one card flag"
	for _, combiAna := range flagAna.Analysis {
		if combiAna.Comb.Formation.Value >= cards.FPhalanx.Value && len(combiAna.HandCardixs) > 0 {
			troopix = maxTroop(combiAna.HandCardixs)
			break
		}
	}
	return troopix, logTxt
}

//ssfSum if we can only make a sum or the opponent can only make a sum we may be
//able to play a sum card.
func pffSum(flagAna *flag.Analysis) (troopix int, logTxt string) {
	logTxt = "Sum move"

	if flagAna.IsTargetRanked && flagAna.TargetRank == 0 {
		troopix = keepFirstCard(flagAna.KeepFlagHandTroopixs, flagAna.SumCards)
	} else {
		if flagAna.BotTroopNo > 1 {
			botRank := 0
			for _, combiAna := range flagAna.Analysis {
				if combiAna.Prop > 0 {
					botRank = combiAna.Comb.Rank
					break
				}
			}
			if botRank == 0 {
				troopix = keepFirstCard(flagAna.KeepFlagTroopixs, flagAna.SumCards)
			}
		}
	}
	return troopix, logTxt
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
func priDumpMove(flagixs []int, flagsAna map[int]*flag.Analysis, handTroopixs []int) (troopix int, move bat.Move) {
	flagix := flagixs[len(flagixs)-1]
	flagAna := flagsAna[flagix]
	keepTroops := flagsAna[0].KeepFlagTroopixs
	if len(keepTroops) < len(handTroopixs) {
		for _, combiFlag := range flagAna.Analysis {
			if len(combiFlag.HandCardixs) > 0 {
				troopix = keepFirstCard(keepTroops, combiFlag.HandCardixs)
				if cerrors.LogLevel() == cerrors.LOG_Debug {
					log.Printf("Dump move card,flag: %v,%v\n", troopix, flagix)
					log.Printf("Keeps: %v", keepTroops)
				}
				break
			}
		}
		if troopix == 0 {
			troopix = keepFirstCard(keepTroops, handTroopixs)
		}
	} else {
		troopix = minTroop(handTroopixs)
	}
	move = *bat.NewMoveCardFlag(flagix)
	return troopix, move
}
func priFlagLoop(
	flagixs []int,
	flagsAna map[int]*flag.Analysis,
	pff func(flagAna *flag.Analysis) (troopix int, logTxt string)) (troopix int, move bat.Move) {

	var logTxt string
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		troopix, logTxt = pff(flagAna)
		if troopix != 0 {
			if cerrors.LogLevel() == cerrors.LOG_Debug {
				log.Println(logTxt)
				log.Printf("Cardix: %v", troopix)
			}
			move = *bat.NewMoveCardFlag(flagix)
			break
		}
	}
	return troopix, move
}
func priFlagNLoop(
	flagixs []int,
	flagsAna map[int]*flag.Analysis,
	n int,
	pfnf func(flagAna *flag.Analysis) (troopix int, logTxt string)) (troopix int, move bat.Move) {

	troopix, move = priFlagLoop(flagixs, flagsAna, func(flagAna *flag.Analysis) (fTroopix int, logTxt string) {
		if flagAna.BotTroopNo == n {
			fTroopix, logTxt = pfnf(flagAna)
		}
		return fTroopix, logTxt
	})
	return troopix, move
}
func priFlagHandLoop(
	flagixs []int,
	flagsAna map[int]*flag.Analysis,
	handAna map[int][]*combi.Analysis,
	pfhf func(flagAna *flag.Analysis, handAna map[int][]*combi.Analysis) (fTroopix int, logTxt string)) (troopix int, move bat.Move) {

	troopix, move = priFlagLoop(flagixs, flagsAna, func(flagAna *flag.Analysis) (fTroopix int, logTxt string) {
		fTroopix, logTxt = pfhf(flagAna, handAna)
		return fTroopix, logTxt
	})
	return troopix, move
}

func findMoveHandIndex(movesHand map[string][]bat.Move, cardix int, move bat.Move) (moveix int) {
	log.Printf("Hand move: %v,%v\n\n", cardix, move)
	moveix = findMoveIndex(movesHand[strconv.Itoa(cardix)], move)
	return moveix
}
func findMoveIndex(moves []bat.Move, move bat.Move) (moveix int) {
	moveix = -1
	for i, mv := range moves {
		if mv.MoveEqual(move) {
			moveix = i
			break
		}
	}
	if moveix < 0 {
		log.Printf("Fail to find legal move.\nMoves: %v\nMove: %v\n ", moves, move)
		moveix = 0
		panic("Failed to find move")
	}
	return moveix
}

//PlayableTac returns the numbers of playable tactic cards 0,1 or 2
//and if a leader is playable.
func playableTac(
	flags [bat.FLAGS]*flag.Flag,
	botDishTac []int,
	oppDishTac []int) (botNo, oppNo int, botLeader, oppLeader bool) {
	botTacs := make([]int, 0, 5)
	oppTacs := make([]int, 0, 5)
	for _, flag := range flags {
		oppTacs = playedTacFlag(oppTacs, flag.OppEnvs, flag.OppTroops)
		botTacs = playedTacFlag(botTacs, flag.PlayEnvs, flag.PlayTroops)
	}
	botTacs = append(botTacs, botDishTac...)
	botTacs = append(oppTacs, oppDishTac...)

	botNo = len(oppTacs) - len(botTacs) + 1
	oppNo = len(botTacs) - len(oppTacs) + 1
	botLeader = !leaderSearch(botTacs)
	oppLeader = !leaderSearch(oppTacs)
	return botNo, oppNo, botLeader, oppLeader
}
func playableTacNoBot(
	flags [bat.FLAGS]*flag.Flag,
	botDishTac []int,
	oppDishTac []int) (botTacNo int) {
	botTacNo, _, _, _ = playableTac(flags, botDishTac, oppDishTac)
	return botTacNo
}

func leaderSearch(tacs []int) (found bool) {
	for _, tacix := range tacs {
		if tacix == cards.TCAlexander || tacix == cards.TCDarius {
			found = true
			break
		}
	}
	return found
}
func playedTacFlag(tacs []int, envs []int, troops []int) []int {
	tacs = append(tacs, envs...)
	for _, cardix := range troops {
		if cards.IsMorale(cardix) {
			tacs = append(tacs, cardix)
		}
	}
	return tacs
}
