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
	"strconv"
)

var tacsPRI []int

func init() {
	tacsPRI = []int{cards.TCTraitor, cards.TCDeserter,
		cards.TCAlexander, cards.TCDarius,
		cards.TCFog, cards.TCMud,
		cards.TC8, cards.TC123,
		cards.TCScout, cards.TCRedeploy}
}
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
func deckZeroTacMove(
	playableTacNo int,
	playableLeader bool,
	deck *botdeck.Deck,
	handTroopixs []int,
	flags [bat.NOFlags]*flag.Flag) bat.MoveDeck {

	move := *bat.NewMoveDeck(bat.DECKTroop)
	if playableTacNo > 0 || deck.OppTacNo() > 0 {
		isBotFirst := true
		flagsAna, deckMaxValues := analyzeFlags(flags, handTroopixs, deck, isBotFirst)
		analyzeFlagsAddFlagValue(flagsAna)
		analyzeFlagsAddLooseGameFlags(flagsAna)
		keep := newKeep(flagsAna, handTroopixs, deck, isBotFirst)
		if keep.deckCalcPickTac(flagsAna, deck, playableTacNo, playableLeader, handTroopixs, deckMaxValues) {
			move = *bat.NewMoveDeck(bat.DECKTac)
		}
	}
	return move
}
func deckScoutMove(
	playableTacNo int,
	deck *botdeck.Deck,
	hand *bat.Hand,
	flags [bat.NOFlags]*flag.Flag) bat.MoveDeck {

	move := *bat.NewMoveDeck(bat.DECKTroop)
	handTacNo := len(hand.Tacs)
	if handTacNo > 1 && handTacNo < 4 &&
		!slice.Contain(hand.Tacs, cards.TCTraitor) &&
		slice.Contain(hand.Tacs, cards.TCScout) {
		move = *bat.NewMoveDeck(bat.DECKTac)
	} else if handTacNo == 1 && hand.Tacs[0] == cards.TCScout && playableTacNo > 0 {
		isBotFirst := true
		flagsAna, _ := analyzeFlags(flags, hand.Troops, deck, isBotFirst)
		analyzeFlagsAddFlagValue(flagsAna)
		analyzeFlagsAddLooseGameFlags(flagsAna)
		keep := newKeep(flagsAna, hand.Troops, deck, isBotFirst)
		if keep.calcIsHandGood(flagsAna, 2) {
			move = *bat.NewMoveDeck(bat.DECKTac)
		}
	}
	return move
}
func makeMoveDeck(pos *Pos) (moveix int) {
	var move bat.MoveDeck
	switch {
	case pos.deck.DeckTacNo() == 0:
		move = *bat.NewMoveDeck(bat.DECKTroop)
	case pos.deck.DeckTroopNo() == 0:
		move = *bat.NewMoveDeck(bat.DECKTac)
	default:
		handTacNo := len(pos.playHand.Tacs)
		tacAna := newPlayableTacAna(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)

		if handTacNo == 0 {
			move = deckZeroTacMove(tacAna.botNo, tacAna.botLeader, pos.deck, pos.playHand.Troops, pos.flags)
		} else {
			move = deckScoutMove(tacAna.botNo, pos.deck, pos.playHand, pos.flags)
		}
	}
	moveix = findMoveIndex(pos.turn.Moves, move)

	return moveix
}

//tacsPrioritize returns prioritized indices for a list of tactic cards
//Best first.
func tacsPrioritize(tacixs []int) (pixs []int) {
	pixs = make([]int, 0, len(tacixs))
	if len(tacixs) != 0 {
		for _, ptacix := range tacsPRI {
			for i, tactix := range tacixs {
				if tactix == ptacix {
					pixs = append(pixs, i)
					break
				}
			}
			if len(pixs) == len(tacixs) {
				break
			}
		}
	}
	return pixs
}

//scoutReturnTacs returns maximum of two tactic cards that can be returned without problem.
//Prioritised that least valued card first.
func scoutReturnTacs(handTacixs []int, playLeader bool) (tacixs []int) {
	tacixs = make([]int, 0, 2)
	if slice.Contain(handTacixs, cards.TCRedeploy) {
		tacixs = append(tacixs, cards.TCRedeploy)
	}
	if !playLeader {
		if slice.Contain(handTacixs, cards.TCAlexander) {
			tacixs = append(tacixs, cards.TCAlexander)
		}
		if slice.Contain(handTacixs, cards.TCDarius) {
			tacixs = append(tacixs, cards.TCDarius)
		}
	}
	if slice.Contain(handTacixs, cards.TCAlexander) && slice.Contain(handTacixs, cards.TCDarius) {
		tacixs = append(tacixs, cards.TCDarius)
	}
	if len(tacixs) < 2 {
		leftTacixs := slice.WithOutNew(handTacixs, tacixs)
		if len(leftTacixs) > 1 {
			pixs := tacsPrioritize(leftTacixs)
			for i := len(pixs) - 1; i > 0; i-- {
				tacixs = append(tacixs, leftTacixs[pixs[i]])
			}
		}
	}

	return tacixs
}

//makeMoveScoutReturn make the scout return move.
//Strategi return redeploy, return leader if not usable.
//return exces tactic cards priority traitor,deserter,leader,fog,mud,8,123
//return troops minimum not in keep.
func makeMoveScoutReturn(pos *Pos) (moveix int) {
	var tacixs, troopixs []int
	noReturn := pos.playHand.Size() - bat.NOHandInit
	isOppFirst := true
	flagsAna, _ := analyzeFlags(pos.flags, pos.playHand.Troops, pos.deck, isOppFirst)
	analyzeFlagsAddFlagValue(flagsAna)
	keep := newKeep(flagsAna, pos.playHand.Troops, pos.deck, isOppFirst)
	playTacAna := newPlayableTacAna(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)
	tacixs = scoutReturnTacs(pos.playHand.Tacs, playTacAna.botLeader)
	if len(tacixs) >= noReturn {
		tacixs = tacixs[0:noReturn]
		troopixs = make([]int, 0, 2)
	} else {
		noReturnTroops := noReturn - len(tacixs)
		troopixs = keep.demandScoutReturn(noReturnTroops, flagsAna, pos.deck)
	}
	move := *bat.NewMoveScoutReturn(tacixs, troopixs)
	pos.playHand.PlayMulti(move.Tac)
	pos.playHand.PlayMulti(move.Troop)
	pos.deck.PlayScoutReturn(move.Troop, move.Tac)
	moveix = findMoveIndex(pos.turn.Moves, move)
	return moveix
}

func makeMoveHand(pos *Pos) (moveixs [2]int) {
	log.Printf(log.Debug, "Hand: %v\n", pos.playHand)
	isBotFirst := true
	flagsAna, deckMaxValues := analyzeFlags(pos.flags, pos.playHand.Troops, pos.deck, isBotFirst)
	analyzeFlagsAddFlagValue(flagsAna)
	analyzeFlagsAddLooseGameFlags(flagsAna)
	keep := newKeep(flagsAna, pos.playHand.Troops, pos.deck, isBotFirst)
	playTacAna := newPlayableTacAna(pos.flags, pos.playDish.Tacs, pos.oppDish.Tacs)

	if pos.turn.MovesPass {
		cardix, move := lostFlagTacticMove(flagsAna, pos.playHand.Tacs, playTacAna,
			pos.playHand.Troops, pos.deck, deckMaxValues, pos.turn.MovesHand)
		if cardix != 0 {
			moveixs[1] = findMoveHandIndex(pos.turn.MovesHand, cardix, move)
			moveixs[0] = cardix
		} else {
			moveixs[0] = 0
			moveixs[1] = bat.SMPass

		}
	} else {
		cardix := 0
		var move bat.Move
		if playTacAna.botNo > 0 && len(pos.playHand.Tacs) > 0 {
			cardix, move = lostFlagTacticMove(flagsAna, pos.playHand.Tacs, playTacAna,
				pos.playHand.Troops, pos.deck, deckMaxValues, pos.turn.MovesHand)
		}
		if cardix == 0 {
			cardix, move = lostFlagDumpMove(flagsAna, keep)
		}
		if cardix == 0 {
			cardix, move = prioritizedMove(flagsAna, keep, playTacAna, pos.playHand, pos.deck, deckMaxValues)
		}
		moveixs[0] = cardix
		moveixs[1] = findMoveHandIndex(pos.turn.MovesHand, cardix, move)
	}
	return moveixs
}

func analyzeFlags(
	flags [bat.NOFlags]*flag.Flag,
	botHandTroops []int,
	deck *botdeck.Deck,
	isBotFirst bool) (flagsAna map[int]*flag.Analysis, deckMaxValues []int) {
	flagsAna = make(map[int]*flag.Analysis)
	deckMaxValues = deck.MaxValues()
	for flagix, iflag := range flags {
		flagsAna[flagix] = flag.NewAnalysis(iflag, botHandTroops, deckMaxValues, deck, flagix, isBotFirst)
	}
	return flagsAna, deckMaxValues
}
func analyzeFlagsAddFlagValue(flagsAna map[int]*flag.Analysis) {
	flagValues := make([]int, bat.NOFlags)
	for _, ana := range flagsAna {
		if !ana.IsClaimed {
			flagValues[ana.Flagix] = flagValues[ana.Flagix] + ana.OppTroopsNo
			if ana.Flagix == 0 || ana.Flagix == 8 {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 10
			} else if ana.Flagix == 1 || ana.Flagix == 7 {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 20
			} else {
				flagValues[ana.Flagix] = flagValues[ana.Flagix] + 30
			}
		} else {
			if ana.Flagix+1 < 9 && !flagsAna[ana.Flagix+1].IsClaimed {
				flagValues[ana.Flagix+1] = flagValues[ana.Flagix+1] + 40
			}
			if ana.Flagix-1 >= 0 && !flagsAna[ana.Flagix-1].IsClaimed {
				flagValues[ana.Flagix-1] = flagValues[ana.Flagix-1] + 40
			}

			if ana.Flagix+2 < 9 && !flagsAna[ana.Flagix+2].IsClaimed {
				flagValues[ana.Flagix+2] = flagValues[ana.Flagix+2] + 30
			}
			if ana.Flagix-2 >= 0 && !flagsAna[ana.Flagix-2].IsClaimed {
				flagValues[ana.Flagix-2] = flagValues[ana.Flagix-2] + 30
			}
		}
	}
	for i, ana := range flagsAna {
		ana.FlagValue = flagValues[i]
	}
}

//anaFlagsAddLooseGameFlags add if game is lost if lost flag is lost.
//#flagsAna
func analyzeFlagsAddLooseGameFlags(flagsAna map[int]*flag.Analysis) {
	lostixs := make([]int, 0, 9)
	claimedNo := 0
	for _, flagAna := range flagsAna {
		if flagAna.IsLost {
			lostixs = append(lostixs, flagAna.Flagix)
		} else if flagAna.Flag.Claimed == flag.CLAIMOpp {
			claimedNo++
		}
	}
	if len(lostixs)+claimedNo > 4 {
		for i := 0; i < len(lostixs); i++ {
			flagsAna[lostixs[i]].IsLoosingGame = true
		}
	} else {
		for i := 0; i < len(lostixs); i++ {
			flagix := lostixs[i]
			if threeFlagsInRow(flagix, flagsAna, isFlagLostOrClaimed) {
				flagsAna[flagix].IsLoosingGame = true
			}

		}
	}
}

func threeFlagsInRow(lostFlagix int, flagsAna map[int]*flag.Analysis, cond func(*flag.Analysis) bool) (loose bool) {
	if lostFlagix > 1 &&
		cond(flagsAna[lostFlagix-1]) &&
		cond(flagsAna[lostFlagix-2]) {
		loose = true
	} else if lostFlagix < 7 &&
		cond(flagsAna[lostFlagix+1]) &&
		cond(flagsAna[lostFlagix+2]) {
		loose = true
	} else if lostFlagix < 8 && lostFlagix > 0 &&
		cond(flagsAna[lostFlagix+1]) &&
		cond(flagsAna[lostFlagix-1]) {
		loose = true
	}

	return loose
}

func isFlagLostOrClaimed(flagAna *flag.Analysis) bool {
	return flagAna.IsLost || flagAna.Flag.Claimed == flag.CLAIMOpp
}

func prioritizedMove(
	flagsAna map[int]*flag.Analysis,
	keep *keep,
	playTacAna *playTacAna,
	hand *bat.Hand,
	deck *botdeck.Deck,
	deckMaxValues []int) (cardix int, move bat.Move) {

	cardix, move = pri3CardsMove(keep.priFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = pri2CardsMove(keep.priFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = pri1CardMove(keep.priFlagixs, flagsAna)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priFlagKeepLoop(flagsAna, keep, pfkfFog)
	if cardix != 0 {
		return cardix, move
	}

	cardix = keep.newFlagTroopix
	if cardix != 0 {
		return cardix, keep.newFlagMove
	}
	cardix, move = priTacticMove(flagsAna, keep, playTacAna, hand, deck, deckMaxValues)
	if cardix != 0 {
		return cardix, move
	}

	cardix, move = priFlagNLoop(keep.priFlagixs, flagsAna, 1, pfnf2Pick1Card)
	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priFlagKeepLoop(flagsAna, keep, pfkfSum)

	if cardix != 0 {
		return cardix, move
	}
	cardix, move = priDumpMove(flagsAna, keep)
	return cardix, move

}

//lostFlagTacticMove handle the strategi of preventing a lost flag to become lost.

//Deserter strategi just remove the best troop if it gives a probability of a win.
//Best troop is morale or mid card. for straight and higist for battalion and sum.
func lostFlagTacticMove(
	flagsAna map[int]*flag.Analysis,
	handTacixs []int,
	playTacAna *playTacAna,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int,
	handMoves map[string][]bat.Move) (cardix int, move bat.Move) {

	if playTacAna.botNo > 0 {
		if len(handTacixs) > 2 && deck.DeckTroopNo() > 1 && slice.Contain(handTacixs, cards.TCScout) {
			if slice.Contain(handTacixs, cards.TCTraitor) || len(handTacixs) == 4 {
				cardix = cards.TCScout
				move = *bat.NewMoveDeck(bat.DECKTroop)
			}
		} else {
			priHandTacixs := tacsPrioritize(handTacixs)
			for _, pix := range priHandTacixs {
				handTacix := handTacixs[pix]
				switch handTacix {
				case cards.TCRedeploy:
					fallthrough
				case cards.TCTraitor:
					cardix, move = lostFlagTacticDbFlagMove(flagsAna, handTroopixs, deck, deckMaxValues, handTacix, handMoves)
				case cards.TCDeserter:
					cardix, move = lostFlagTacticDeserterMove(flagsAna, handTroopixs, deck, deckMaxValues)
				case cards.TCScout:

				default:
					cardix, move = lostFlagTacticSimMove(flagsAna, handTacix, playTacAna, handTroopixs, deck, deckMaxValues)
				}
				if cardix != 0 {
					break
				}
			}
			if cardix != 0 && len(handTacixs) == 2 && deck.DeckTroopNo() > 1 &&
				slice.Contain(handTacixs, cards.TCScout) &&
				slice.Contain(handTacixs, cards.TCTraitor) {
				cardix = cards.TCScout
				move = *bat.NewMoveDeck(bat.DECKTroop)
			}
		}
	}
	return cardix, move
}

//lostFlagTacticDeserterMove makes a deserter move.
//Deserter strategi just remove the best troop if it gives a win or prevent losing game.
//Best troop is morale or mid card. for straight and highest for battalion and sum where morale is morale max value.
//For mud or fog we simulate.
func lostFlagTacticDeserterMove(
	flagsAna map[int]*flag.Analysis,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (cardix int, move bat.Move) {

	for flagix, flagAna := range flagsAna {
		if flagAna.IsLost {
			isLost := true
			isWin := false
			desertix := 0
			envIsWin, envIsLost, envix := deserterKillEnvSim(flagAna.Flag, flagix, handTroopixs, deck, deckMaxValues)
			if envIsWin {
				isLost = envIsLost
				isWin = envIsWin
				desertix = envix
			} else {
				troopMoralIx := deserterKillTroopMoral(flagAna.Flag.OppTroops, flagAna.TargetRank, flagAna.FormationSize)
				simFlag := flagAna.Flag.Copy()
				simFlag.OppRemoveCardix(troopMoralIx)
				troopMoralIsWin, troopMoralIsLost := moveFlagHandSim(simFlag, flagix, handTroopixs, deck, deckMaxValues)
				if troopMoralIsWin || !troopMoralIsLost {
					isWin = troopMoralIsWin
					isLost = troopMoralIsLost
					desertix = troopMoralIx
				}
			}
			if flagAna.IsLoosingGame && !isLost || isWin {
				move = *bat.NewMoveDeserter(flagix, desertix)
				cardix = cards.TCDeserter
				break
			}
		}
	}
	return cardix, move
}
func simMudTrimFlag(simFlag *flag.Flag, cardix int) *flag.Flag {
	if cards.TCMud == cardix {
		if len(simFlag.OppTroops) > 3 {
			simFlag.OppRemoveCardix(mudAutoDish(simFlag.OppTroops))
		}
		if len(simFlag.PlayTroops) > 3 {
			simFlag.PlayRemoveCardix(mudAutoDish(simFlag.PlayTroops))
		}
	}
	return simFlag
}
func mudAutoDish(cardixs []int) (ix int) {
	lowestRank := 1
	lowestSum := 0
	handCards := make([]int, 0, 0)
	drawSet := make(map[int]bool)
	drawNo := 0
	mud := false
	sumRank := 200
	for _, outix := range cardixs {
		simCardixs := slice.WithOutNew(cardixs, []int{outix})
		rank := flag.CalcMaxRank(simCardixs, handCards, drawSet, drawNo, mud)
		sum := 0
		if rank == 0 {
			rank = sumRank
			sum = flag.MoraleTroopsSum(simCardixs)
		}
		if rank > lowestRank {
			lowestRank = rank
			ix = outix
			lowestSum = sum
		} else if rank == sumRank && lowestRank == sumRank {
			if sum < lowestSum {
				ix = outix
				lowestSum = sum
			}
		}

	}
	return ix
}
func deserterKillEnvSim(
	flag *flag.Flag,
	flagix int,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (isWin, isLost bool, desertEnvix int) {
	isLost = true
	if len(flag.OppEnvs) != 0 {
		for _, envix := range flag.OppEnvs {
			simFlag := flag.Copy()
			simFlag.OppRemoveCardix(envix)
			simFlag = simMudTrimFlag(simFlag, envix)
			log.Printf(log.Debug, "Deserter kill enviroment move\nSim Flag %+v\nOld Flag %+v", simFlag, flag)
			simIsWin, simIsLost := moveFlagHandSim(simFlag, flagix, handTroopixs, deck, deckMaxValues)
			if desertEnvix == 0 {
				isWin = simIsWin
				isLost = simIsLost
				desertEnvix = envix
				if isWin {
					break
				}
			} else {
				if simIsWin || !simIsLost {
					isWin = simIsWin
					isLost = simIsLost
					desertEnvix = envix
				}
			}
		}
	}
	return isWin, isLost, desertEnvix
}

func deserterKillTroopMoral(oppTroopixs []int, targetRank, formationSize int) (desertix int) {
	tacixs, troopixs := sortFlagCards(oppTroopixs)
	if targetRank != 0 {
		combination := combi.Combinations(formationSize)[targetRank]
		switch combination.Formation.Value {
		case cards.FWedge.Value:
			fallthrough
		case cards.FSkirmish.Value:
			if len(tacixs) > 0 {
				desertix = tacixs[0]
			} else {
				desertix = troopixs[1]
			}
		case cards.FPhalanx.Value:
			desertix = oppTroopixs[0]
		case cards.FBattalion.Value:
			desertix = deserterKillStrenght(tacixs, troopixs)
		}

	} else {
		desertix = deserterKillStrenght(tacixs, troopixs)
	}

	return desertix
}
func deserterKillStrenght(tacixs, troopixs []int) (desertix int) {
	tacValue := 0
	troopValue := 0
	if len(tacixs) > 0 {
		tacValue = cards.MoraleMaxValue(tacixs[0])
	}
	if len(troopixs) > 0 {
		troop, _ := cards.DrTroop(troopixs[0])
		troopValue = troop.Value()
	}
	if troopValue > tacValue {
		desertix = troopixs[0]
	} else {
		desertix = tacixs[0]
	}
	return desertix
}
func sortFlagCards(flagTroopixs []int) (tacixs, troopixs []int) {
	tacixs = make([]int, 0, 4)
	troopixs = make([]int, 0, 4)
	for _, flagTroopix := range flagTroopixs {
		if cards.IsTac(flagTroopix) {
			tacixs = addSortedCards(tacixs, flagTroopix, cards.MoraleMaxValue)
		} else {
			troopixs = addSortedCards(troopixs, flagTroopix, func(troopix int) int {
				troop, _ := cards.DrTroop(troopix)
				return troop.Value()
			})
		}
	}
	return tacixs, troopixs
}
func addSortedCards(list []int, cardix int, valuef func(int) int) []int {
	listSize := len(list)
	list = append(list, 0)

	for i, v := range list {
		if i == listSize {
			list[listSize] = cardix
		} else {
			if !(valuef(cardix) < valuef(v)) {
				copy(list[i+1:], list[i:])
				list[i] = cardix
				break
			}
		}
	}
	return list
}

func lostFlagTacticSimMove(
	flagsAna map[int]*flag.Analysis,
	handTacix int,
	playTacAna *playTacAna,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (cardix int, move bat.Move) {
	for flagix, flagAna := range flagsAna {
		if flagAna.IsLost {
			if flagAna.IsPlayable && cards.IsMorale(handTacix) && !cards.IsLeader(handTacix) && !flagAna.IsNewFlag ||
				cards.IsEnv(handTacix) ||
				cards.IsLeader(handTacix) && flagAna.IsPlayable && playTacAna.botLeader && !flagAna.IsNewFlag {
				isWin, isLost := tacticMoveSim(flagix, flagAna.Flag, handTacix, handTroopixs, deck, deckMaxValues)
				if flagAna.IsLoosingGame && !isLost || isWin {
					cardix = handTacix
					move = *bat.NewMoveCardFlag(flagix)
					break
				}
			}
		}
	}
	return cardix, move
}

//lostFlagDumpMove handle the strategi of dumbing card on lost flag.
//Dump the smallest card if it exist of card not usable in the best formation or in case
//of the best formation being wedge also the phalanx.
func lostFlagDumpMove(
	flagsAna map[int]*flag.Analysis,
	keep *keep) (cardix int, move bat.Move) {
	lostFlagix := -1
	for _, ana := range flagsAna {
		if ana.IsPlayable && ana.IsLost {
			lostFlagix = ana.Flagix
			break
		}
	}
	if lostFlagix != -1 {
		cardix = keep.requestFlagHandLowestValue(keep.handTroopixs)
		if cardix != 0 {
			log.Println(log.Debug, "Made a Lost Flag Dump move")
			move = *bat.NewMoveCardFlag(lostFlagix)
		}
	}
	return cardix, move
}

func prioritizePlayableFlags(flagsAna map[int]*flag.Analysis) (flagixs []int) {
	flagValues := make([]int, bat.NOFlags)
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
	if !flagAna.IsFog && flagAna.OppTroopsNo > 0 {
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
	troopValue := 0
	for cix, sumProp := range cardSumProp {
		if sumProp > maxSumProp {
			troopix = cix
			maxSumProp = sumProp
			troop, _ := cards.DrTroop(troopix)
			troopValue = troop.Value()
		} else if sumProp == maxSumProp {
			troop, _ := cards.DrTroop(troopix)
			if troop.Value() > troopValue {
				troopix = cix
				troopValue = troop.Value()
			}
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
	minValue := cards.NOTroop + 1
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
func min2Troop(troopixs []int) (minTroopixs []int) {
	minTroopixs = make([]int, 2)
	minValue1 := 0
	minValue2 := 0
	for _, troopix := range troopixs {
		troop, _ := cards.DrTroop(troopix)
		if troop.Value() < minValue1 {
			minValue1 = troop.Value()
			minTroopixs[1] = minTroopixs[0]
			minTroopixs[0] = troopix
		} else if troop.Value() < minValue2 {
			minValue2 = troop.Value()
			minTroopixs[1] = troopix
		}
	}
	return minTroopixs
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
func pfkfFog(flagAna *flag.Analysis, keep *keep) (troopix int, logTxt string) {
	logTxt = "Made a fog move"
	if flagAna.IsFog {
		if len(flagAna.SumCards) > 0 {
			troopix = keep.requestFirst(flagAna.SumCards)
		}
	} else if flagAna.TargetRank == 0 {
		if len(flagAna.SumCards) > 0 && flagAna.BotMaxSum >= flagAna.TargetSum {
			troopix = keep.requestFirst(flagAna.SumCards)
		}
	}
	return troopix, logTxt
}

func priNewFlagMove(flagixs []int,
	flagsAna map[int]*flag.Analysis,
	handAna map[int][]*combi.Analysis,
	handTroopixs []int,
	keepFlag map[int]bool) (troopix int, move bat.Move) {

	var logTxt string
	logTxt = "New flag is rank move"
	moveFlagix := -1
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppTroopsNo > 0 {
			targetRank := flagAna.TargetRank
			if flagAna.IsTargetMade {
				targetRank = targetRank - 1
			}
			troopix = newFlagTargetMove(handAna, targetRank)
		}
		if troopix != 0 {
			moveFlagix = flagix
			break
		}
	}
	if troopix == 0 {
		nfTroopix := 0
		nfTroopix, logTxt = newFlagMadeCombiMove(handAna)
		if nfTroopix == 0 {
			nfTroopix, logTxt = newFlagPhalanxMove(handAna)
		}
		if nfTroopix == 0 {
			nfTroopix, logTxt = newFlagHigestRankMove(handAna, keepFlag)
		}

		if nfTroopix != 0 {
			flagix := newFlagSelectFlag(nfTroopix, flagixs, flagsAna)
			if flagix != -1 {
				troopix = nfTroopix
				moveFlagix = flagix
			}

		}
	}
	if troopix != 0 {
		log.Printf(log.Debug, "%v Cardix: %v", logTxt, troopix)
		move = *bat.NewMoveCardFlag(moveFlagix)
	}

	return troopix, move

}

func newFlagSelectFlag(troopix int, flagixs []int, flagsAna map[int]*flag.Analysis) (flagix int) {
	flagix = -1
	newFlagixs := make([]int, 0, len(flagixs))
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsNewFlag && flagAna.OppTroopsNo == 0 {
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

	logTxt = "New flag is best phalanx or higher move"
	targetRank := combi.LastFormationRank(cards.FPhalanx, 3)
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

	logTxt = "New flag is a combi move"
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
	logTxt = "New flag is highest rank move"
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

//priTacticMove makes a tactic card move because no good troop move exist.
//Only morale cards, scout
//Fog,Mud,deserter,traitor,redeploy is for defence.
//Morale strategi: cards out if morale cards leads to a win play it.
//Scout strategi: have many in keep try to get traitor. Got traitor or 3 tactics + scout play it.
//Just play it if not playing try to get traitor(scout + 1 or more tactics)
func priTacticMove(
	flagsAna map[int]*flag.Analysis,
	keep *keep,
	playTacAna *playTacAna,
	hand *bat.Hand,
	deck *botdeck.Deck,
	deckMaxValues []int) (tacix int, move bat.Move) {

	noHandTacs := len(hand.Tacs)
	if playTacAna.botNo > 0 && noHandTacs > 0 {
		if noHandTacs > 1 && deck.DeckTroopNo() > 1 && slice.Contain(hand.Tacs, cards.TCScout) {
			if slice.Contain(hand.Tacs, cards.TCTraitor) || noHandTacs == 4 {
				tacix = cards.TCScout
				move = *bat.NewMoveDeck(bat.DECKTroop)
			}
		} else if hand.Tacs[0] == cards.TCScout && noHandTacs == 1 && deck.DeckTroopNo() > 1 {
			if !keep.calcIsHandGood(flagsAna, 2) {
				tacix = cards.TCScout
				move = *bat.NewMoveDeck(bat.DECKTroop)
			}
		} else {
			tacix, move = priTacticMoveSim(flagsAna, keep.priFlagixs, playTacAna, hand, deck, deckMaxValues)
		}
	}
	return tacix, move
}

//priTacticMoveSim simulate morale tactic cards and create move
//if win exist.
func priTacticMoveSim(
	flagsAna map[int]*flag.Analysis,
	priFlagixs []int,
	playTacAna *playTacAna,
	hand *bat.Hand,
	deck *botdeck.Deck,
	deckMaxValues []int) (tacix int, move bat.Move) {
Loop:
	for _, tac := range hand.Tacs {
		if cards.IsMorale(tac) {
			isLeader := cards.IsLeader(tac)
			if (isLeader && playTacAna.botLeader) || !isLeader {
				for i := 0; i < len(priFlagixs); i++ {
					if !flagsAna[i].IsNewFlag && flagsAna[i].IsPlayable {
						isWin, _ := tacticMoveSim(flagsAna[i].Flagix, flagsAna[i].Flag, tac, hand.Troops, deck, deckMaxValues)
						if isWin {
							tacix = tac
							move = *bat.NewMoveCardFlag(flagsAna[i].Flagix)
							break Loop
						}
					}
				}
			}
		}
	}
	return tacix, move
}
func moveFlagHandSim(
	simFlag *flag.Flag,
	simFlagix int,
	simHandTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (isWin bool, isLost bool) {

	fa := flag.NewAnalysis(simFlag, simHandTroopixs, deckMaxValues, deck, simFlagix, true)
	isLost = fa.IsLost
	isWin = fa.IsWin()
	log.Printf(log.Debug, "Simulated flag: %v, win,loss: %v,%v\n Analysis: %+v\n", simFlagix, isWin, isLost, fa)
	return isWin, isLost
}
func tacticMoveSim(flagix int, flag *flag.Flag,
	tacix int,
	handTroopixs []int,
	deck *botdeck.Deck,
	deckMaxValues []int) (isWin bool, isLost bool) {

	simFlag := flag.Copy()
	simFlag.PlayAddCardix(tacix)

	tac, _ := cards.DrTactic(tacix)
	log.Printf(log.Debug, "Tactic move %v\nSim Flag %+v\nOld Flag %+v", tac.Name(), simFlag, flag)
	isWin, isLost = moveFlagHandSim(simFlag, flagix, handTroopixs, deck, deckMaxValues)

	return isWin, isLost
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

//ssfSum if we can only make a sum we may be able to play a sum card.
func pfkfSum(flagAna *flag.Analysis, keep *keep) (troopix int, logTxt string) {
	logTxt = "Sum move"
	if flagAna.BotTroopNo > 1 {
		botRank := 0
		for _, combiAna := range flagAna.Analysis {
			if combiAna.Prop > 0 {
				botRank = combiAna.Comb.Rank
				break
			}
		}
		if botRank == 0 {
			troopix = keep.requestFirstHand(flagAna.SumCards)
		}
	}
	return troopix, logTxt
}

func priDumpMove(
	flagsAna map[int]*flag.Analysis,
	keep *keep) (troopix int, move bat.Move) {
	flagix := keep.priFlagixs[len(keep.priFlagixs)-1]
	flagAna := flagsAna[flagix]
	troopix = keep.demandDump(flagAna)
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
			logTxt = logTxt + fmt.Sprintf("\nCardix: %v", troopix)
			log.Print(log.Debug, logTxt)
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
func priFlagKeepLoop(
	flagsAna map[int]*flag.Analysis,
	keep *keep,
	pfkf func(flagAna *flag.Analysis, keep *keep) (fTroopix int, logTxt string)) (troopix int, move bat.Move) {

	troopix, move = priFlagLoop(keep.priFlagixs, flagsAna, func(flagAna *flag.Analysis) (fTroopix int, logTxt string) {
		fTroopix, logTxt = pfkf(flagAna, keep)
		return fTroopix, logTxt
	})
	return troopix, move
}

func findMoveHandIndex(movesHand map[string][]bat.Move, cardix int, move bat.Move) (moveix int) {
	log.Printf(log.Debug, "Hand move: %v,%v\n\n", cardix, move)
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
		logTxt := fmt.Sprintf("Fail to find legal move.\nMoves: %v\nMove: %v\n ", moves, move)
		log.Printf(log.Min, logTxt)
		moveix = 0
		panic(logTxt)
	}
	return moveix
}

//PlayableTac returns the numbers of playable tactic cards 0,1 or 2
//and if a leader is playable.
func newPlayableTacAna(
	flags [bat.NOFlags]*flag.Flag,
	botDishTac []int,
	oppDishTac []int) (ana *playTacAna) {

	ana = new(playTacAna)
	botTacs := make([]int, 0, 5)
	oppTacs := make([]int, 0, 5)
	for _, flag := range flags {
		oppTacs = playedTacFlag(oppTacs, flag.OppEnvs, flag.OppTroops)
		botTacs = playedTacFlag(botTacs, flag.PlayEnvs, flag.PlayTroops)
	}
	botTacs = append(botTacs, botDishTac...)
	oppTacs = append(oppTacs, oppDishTac...)

	ana.botNo = len(oppTacs) - len(botTacs) + 1
	ana.oppNo = len(botTacs) - len(oppTacs) + 1

	ana.botLeader = !leaderSearch(botTacs)
	ana.oppLeader = !leaderSearch(oppTacs)
	return ana
}

type playTacAna struct {
	botNo     int
	oppNo     int
	botLeader bool
	oppLeader bool
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
