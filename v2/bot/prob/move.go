package prob

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	fa "github.com/rezder/go-battleline/v2/bot/prob/flag"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
)

var tacsPRI []card.Card

func init() {
	tacsPRI = []card.Card{card.TCTraitor, card.TCDeserter,
		card.TCAlexander, card.TCDarius,
		card.TCFog, card.TCMud,
		card.TC8, card.TC123,
		card.TCScout, card.TCRedeploy}
}

//MoveClaim makes a claim flag move.
func MoveClaim(viewPos *game.ViewPos) (moveix int) {
	posCards := game.NewPosCards(viewPos.CardPos)
	posCards = mudTrim(posCards, viewPos.CardPos[card.TCMud])
	botix := viewPos.Playerix()
	deckTroop := posCards.SimDeckTroops()
	flags := game.FlagsCreate(posCards, viewPos.ConePos)
	var coneixs []int
	for flagix, flag := range flags {
		isClaim, _ := flag.IsClaimable(botix, deckTroop)
		if isClaim {
			coneixs = append(coneixs, flagix+1)
		}
	}
	var moves Moves = viewPos.Moves
	moveix = moves.FindCone(coneixs)
	return moveix
}

//MoveDeck makes deck move.
func MoveDeck(viewPos *game.ViewPos) (moveix int) {
	posCards := game.NewPosCards(viewPos.CardPos)
	posCards = mudTrim(posCards, viewPos.CardPos[card.TCMud])
	deck := fa.NewDeck(viewPos, posCards)
	botix := viewPos.Playerix()
	var deckPos pos.Card
	switch {
	case deck.DeckTacNo() == 0:
		deckPos = pos.CardAll.DeckTroop
	case deck.DeckTroopNo() == 0:
		deckPos = pos.CardAll.DeckTac
	default:
		hands := createHands(posCards, botix)
		flags := game.FlagsCreate(posCards, viewPos.ConePos)
		tacAna := newPlayableTacAna(viewPos.CardPos, botix)
		if hands[botix].NoTacs() == 0 {
			deckPos = deckZeroTacMove(tacAna.botNo, tacAna.IsBotLeader, deck, hands, botix, flags[:], posCards, viewPos.ConePos)
		} else {
			deckPos = deckScoutMove(tacAna.botNo, deck, hands, botix, flags[:])
		}
	}
	var moves Moves = viewPos.Moves
	moveix = moves.FindDeck(deckPos)

	return moveix
}
func createHands(posCards game.PosCards, botix int) (hands [2]*card.Cards) {
	for i := 0; i < len(hands); i++ {
		hands[i] = posCards.SortedCards(pos.CardAll.Players[i].Hand)
	}
	return hands
}
func opp(ix int) (oppix int) {
	oppix = ix + 1
	if oppix > 1 {
		oppix = 0
	}
	return oppix
}

//MoveScoutReturn make the scout return move.
//Strategi return redeploy, return leader if not usable.
//return exces tactic cards priority traitor,deserter,leader,fog,mud,8,123
//return troops minimum not in keep.
func MoveScoutReturn(viewPos *game.ViewPos) (moveix int) {
	var tacs []card.Card
	var troops []card.Troop
	posCards := game.NewPosCards(viewPos.CardPos)
	posCards = mudTrim(posCards, viewPos.CardPos[card.TCMud])
	botix := viewPos.Playerix()
	deck := fa.NewDeck(viewPos, posCards)
	hands := createHands(posCards, botix)
	drawFirst := opp(botix)
	drawNos := deck.DrawNos(drawFirst)
	deckHandTroops := dht.NewCache(deck.Troops(), [2][]card.Troop{hands[0].Troops, hands[1].Troops}, drawNos)
	flags := game.FlagsCreate(posCards, viewPos.ConePos)
	noReturn := hands[botix].No() - game.NOHandInit

	flagsAna := analyzeFlags(flags[:], deckHandTroops, botix)
	analyzeFlagsAddFlagValue(flagsAna)
	keep := NewKeep(flagsAna, viewPos.Moves, deckHandTroops, botix)
	playTacAna := newPlayableTacAna(viewPos.CardPos, botix)
	tacs = scoutReturnTacs(hands[botix], playTacAna.IsBotLeader)
	if len(tacs) >= noReturn {
		tacs = tacs[0:noReturn]
		troops = make([]card.Troop, 0, 2)
	} else {
		noReturnTroops := noReturn - len(tacs)
		troops = keep.DemandScoutReturn(noReturnTroops, flagsAna)
	}
	var moves Moves = viewPos.Moves
	moveix = moves.FindScoutReturn(tacs, troops)
	return moveix
}

//MoveHand makes a move from the hand.
func MoveHand(viewPos *game.ViewPos) (moveix int) {
	posCards := game.NewPosCards(viewPos.CardPos)
	posCards = mudTrim(posCards, viewPos.CardPos[card.TCMud])
	botix := viewPos.Playerix()
	deck := fa.NewDeck(viewPos, posCards)
	hands := createHands(posCards, botix)
	drawFirst := botix
	deckDrawNos := deck.DrawNos(drawFirst)
	deckHandTroops := dht.NewCache(deck.Troops(), [2][]card.Troop{hands[0].Troops, hands[1].Troops}, deckDrawNos)
	flags := game.FlagsCreate(posCards, viewPos.ConePos)
	log.Printf(log.Debug, "Hand: %v\n", hands[botix])

	flagsAna := analyzeFlags(flags[:], deckHandTroops, botix)
	sim := &Sim{
		deckHandTroops: deckHandTroops,
		botix:          botix,
		posCards:       posCards,
		conePos:        viewPos.ConePos,
		Moves:          viewPos.Moves,
	}
	analyzeFlagsAddFlagValue(flagsAna)
	analyzeFlagsAddLooseGameFlags(flagsAna)
	keep := NewKeep(flagsAna, viewPos.Moves, deckHandTroops, botix)
	playTacAna := newPlayableTacAna(viewPos.CardPos, botix)
	passMoveix := sim.SearchPass()
	moveix = -1
	if passMoveix != -1 {
		moveix = lostFlagTacticMove(flagsAna, keep.PriLiveFlagixs, hands[botix], playTacAna, deck.DeckTroopNo(), sim)
		if moveix == -1 {
			moveix = passMoveix
		}
	} else {
		if playTacAna.botNo > 0 && hands[botix].NoTacs() > 0 {
			moveix = lostFlagTacticMove(flagsAna, keep.PriLiveFlagixs, hands[botix], playTacAna, deck.DeckTroopNo(), sim)
		}
		if moveix == -1 {
			moveix = lostFlagDumpMove(flagsAna, keep, viewPos.Moves)
		}
		if moveix == -1 {
			moveix = prioritizedMove(flagsAna, keep, playTacAna, hands[botix], deck.DeckTroopNo(), sim)
		}
		if moveix == -1 {
			panic(fmt.Sprintf("Failed to find move: %v", viewPos.Moves[moveix]))
		}
	}
	return moveix
}
func moveTf(probas []float64) (moveix int) {
	var maxProba float64
	for i, proba := range probas {
		if maxProba < proba {
			maxProba = proba
			moveix = i
		}
	}
	return moveix
}

//PlayableTac returns the numbers of playable tactic cards 0,1 or 2
//and if a leader is playable.
func newPlayableTacAna(cardPos [71]pos.Card, botix int) (ana *playTacAna) {
	ana = new(playTacAna)
	ana.IsBotLeader = true
	ana.isOppLeader = true
	botNo := 0
	oppNo := 0
	for cardix := len(cardPos) - 1; cardix > card.NOTroop; cardix-- {
		cardPos := cardPos[cardix]
		if cardPos.IsOnTable() {
			if cardPos.Player() == botix {
				botNo = botNo + 1
				if card.Card(cardix).IsMorale() && card.Morale(cardix).IsLeader() {
					ana.IsBotLeader = false
				}
			} else {
				oppNo = oppNo + 1
				if card.Card(cardix).IsMorale() && card.Morale(cardix).IsLeader() {
					ana.isOppLeader = false
				}
			}
		}

	}
	ana.botNo = oppNo - botNo + 1
	ana.oppNo = botNo - oppNo + 1
	return ana
}

type playTacAna struct {
	botNo       int
	oppNo       int
	IsBotLeader bool
	isOppLeader bool
}

func deckZeroTacMove(
	playableTacNo int,
	playableLeader bool,
	deck *fa.Deck,
	hands [2]*card.Cards,
	botix int,
	flags []*game.Flag,
	posCards game.PosCards,
	conePos [10]pos.Cone,
) (deckPos pos.Card) {

	deckPos = pos.CardAll.DeckTroop
	if playableTacNo > 0 || deck.OppHandTacNo() > 0 {
		drawFirst := botix
		deckHandTroops := dht.NewCache(deck.Troops(), [2][]card.Troop{hands[0].Troops, hands[1].Troops}, deck.DrawNos(drawFirst))
		flagsAna := analyzeFlags(flags, deckHandTroops, botix)
		sim := &Sim{
			deckHandTroops: deckHandTroops,
			botix:          botix,
			posCards:       posCards,
			conePos:        conePos,
		}
		analyzeFlagsAddFlagValue(flagsAna)
		analyzeFlagsAddLooseGameFlags(flagsAna)
		if deckCalcPickTac(flagsAna, playableTacNo, playableLeader, deck.Tacs(), deck.ScoutReturnTacPeek(), sim) {
			deckPos = pos.CardAll.DeckTac
		}
	}
	return deckPos
}
func analyzeFlags(
	flags []*game.Flag,
	deckHandTroops *dht.Cache,
	botix int) (flagsAna map[int]*fa.Analysis) {
	flagsAna = make(map[int]*fa.Analysis)
	for flagix, flag := range flags {
		flagsAna[flagix] = fa.NewAnalysis(botix, flag, deckHandTroops, flagix)
	}
	return flagsAna
}
func analyzeFlagsAddFlagValue(flagsAna map[int]*fa.Analysis) {
	flagValues := make([]int, len(flagsAna))
	for _, ana := range flagsAna {
		if !ana.IsClaimed {
			flagValues[ana.Flagix] = flagValues[ana.Flagix] + ana.OppFormationSize
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
func analyzeFlagsAddLooseGameFlags(flagsAna map[int]*fa.Analysis) {
	lostixs := make([]int, 0, 9)
	oppClaimedNo := 0
	for _, flagAna := range flagsAna {
		if flagAna.IsLost {
			lostixs = append(lostixs, flagAna.Flagix)
		} else if flagAna.IsClaimed && flagAna.Claimer != flagAna.Playix {
			oppClaimedNo++
		}
	}
	if len(lostixs)+oppClaimedNo > 4 {
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
func isFlagLostOrClaimed(flagAna *fa.Analysis) bool {
	return flagAna.IsLost || (flagAna.IsClaimed && flagAna.Claimer != flagAna.Playix)
}
func threeFlagsInRow(
	lostFlagix int,
	flagsAna map[int]*fa.Analysis,
	cond func(*fa.Analysis) bool) (loose bool) {

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

//probMaxSumTroop findes the troop with the bigest probability if any. Tiebreaker
//troop strenght.
func probMaxSumTroop(cardSumProp map[card.Troop]float64) (maxCard card.Card) {
	maxSumProp := float64(0)
	for troop, sumProp := range cardSumProp {
		if sumProp > maxSumProp {
			maxCard = card.Card(troop)
			maxSumProp = sumProp
		} else if sumProp == maxSumProp {
			if maxCard.IsNone() || troop.Strenght() > card.Troop(maxCard).Strenght() {
				maxCard = card.Card(troop)
			}
		}
	}
	return maxCard
}

//deckCalcPickTac evaluate if it is a good idea to pick tactic card.
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
func deckCalcPickTac(
	flagsAna map[int]*fa.Analysis,
	playableTacNo int,
	playableLeader bool,
	deckTacs []card.Card,
	peekTac card.Card,
	sim *Sim) (isPickTac bool) {
	offenceTacs, defenceTacs := offDefTacs(playableLeader, deckTacs, peekTac)
	offFlagSet, offenceTacSet := findOffenceFlags(offenceTacs, flagsAna, sim)
	defFlagSet, defenceTacSet := findDefenceFlags(defenceTacs, flagsAna)
	if len(offenceTacSet) > 0 || len(defenceTacSet) > 0 {
		logTxt := ""
		if looseWinFlagExist(flagsAna) {
			logTxt = "Loose or win flag exist"
			isPickTac = true
		} else if len(offFlagSet)+len(defFlagSet) > 4 && playableTacNo != 0 {
			logTxt = "Many flag need tactic cards"
			isPickTac = true
		}
		if len(logTxt) != 0 {
			logTxt = logTxt + fmt.Sprintf("\nOffence Flags: %v Offence Tactics: %v\n", offFlagSet, offenceTacSet)
			logTxt = logTxt + fmt.Sprintf("Defence Flags: %v Defence Tactics: %v\n", defFlagSet, defenceTacSet)
			log.Print(log.Debug, logTxt)
		}

	}
	return isPickTac
}

//offDeffTacs find the relevante offensive and defensive tatic cards remaining in the deck.
//When we know the next tactic card only that card is used.
//if leader is not playable leader cards is removed.
func offDefTacs(
	playableLeader bool,
	deckTacs []card.Card,
	peekTac card.Card,
) (offenceTacs, defenceTacs []card.Card) {
	defenceTacs = make([]card.Card, 0, 4)
	offenceTacs = make([]card.Card, 0, 4)
	if !peekTac.IsNone() {
		if isOffenceTac(peekTac) {
			offenceTacs = append(offenceTacs, peekTac)
		} else if isDefenceTac(peekTac) {
			defenceTacs = append(defenceTacs, peekTac)
		}
	} else {
		for _, tac := range deckTacs {
			if isOffenceTac(tac) {
				offenceTacs = append(offenceTacs, tac)
			} else if isDefenceTac(tac) {
				defenceTacs = append(defenceTacs, tac)
			}
		}
	}
	if !playableLeader && len(offenceTacs) > 0 {
		copytac := make([]card.Card, 0, 2)
		for _, tac := range offenceTacs {
			if !tac.IsMorale() || !card.Morale(tac).IsLeader() {
				copytac = append(copytac, tac)
			}
		}
		offenceTacs = copytac
	}
	return offenceTacs, defenceTacs
}
func looseWinFlagExist(flagsAna map[int]*fa.Analysis) (exist bool) {
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
	flagsAna map[int]*fa.Analysis,
	cond func(*fa.Analysis) bool) (no int) {
	for _, flagAna := range flagsAna {
		if cond(flagAna) {
			no = no + 1
		}
	}
	return no
}
func isFlagWonOrClaimed(flagAna *fa.Analysis) bool {
	return (!flagAna.IsClaimed && flagAna.IsWin) || flagAna.Claimer == flagAna.Playix
}
func isOffenceTac(tac card.Card) bool {
	return tac.IsMorale()
}
func isDefenceTac(tac card.Card) (isDef bool) {
	if tac.IsEnv() {
		isDef = true
	} else if tac.IsGuile() {
		guile := card.Guile(tac)
		isDef = guile.IsDeserter() || guile.IsTraitor()
	}
	return isDef
}
func findDefenceFlags(defenceTacs []card.Card, flagsAna map[int]*fa.Analysis,
) (flagixSet map[int]bool, tacSet map[card.Card]bool) {
	flagixSet = make(map[int]bool)
	tacSet = make(map[card.Card]bool)
	if len(defenceTacs) > 0 {
		for _, flagAna := range flagsAna {
			if flagAna.Flag.PlayerFormationSize(opp(flagAna.Playix)) >= flagAna.FormationSize-1 &&
				!flagAna.IsClaimed && !flagAna.IsLost {
				for _, tac := range defenceTacs {
					if tac.IsEnv() && card.Env(tac).IsFog() {
						if flagAna.TargetSum <= flagAna.BotMaxSum {
							tacSet[tac] = true
							flagixSet[flagAna.Flagix] = true
						}
					} else {
						tacSet[tac] = true
						flagixSet[flagAna.Flagix] = true
					}
				}
			}
		}
	}
	return flagixSet, tacSet
}
func findOffenceFlags(
	offenceTacs []card.Card,
	flagsAna map[int]*fa.Analysis,
	sim *Sim) (flagixSet map[int]bool, tacSet map[card.Card]bool) {

	flagixSet = make(map[int]bool)
	tacSet = make(map[card.Card]bool)
	if len(offenceTacs) > 0 {
		for _, flagAna := range flagsAna {
			if flagAna.Flag.PlayerFormationSize(flagAna.Playix)+1 == flagAna.FormationSize && !flagAna.IsWin {
				for _, tac := range offenceTacs {
					simMove := createHandFlagMove(flagAna.Flagix, tac, flagAna.Playix)
					_, simFlagAna := sim.Move(simMove)
					if simFlagAna.IsWin {
						flagixSet[flagAna.Flagix] = true
						tacSet[tac] = true
					}
				}
			}
		}
	}
	return flagixSet, tacSet
}
func deckScoutMove(
	playableTacNo int,
	deck *fa.Deck,
	hands [2]*card.Cards,
	botix int,
	flags []*game.Flag,
) (deckPos pos.Card) {

	deckPos = pos.CardAll.DeckTroop
	botHand := hands[botix]
	handTacNo := botHand.NoTacs()
	if handTacNo > 1 && handTacNo < 4 {
		hasTraitor := false
		hasScout := false
		for _, guile := range botHand.Guiles {
			if guile.IsTraitor() {
				hasTraitor = true
				break
			}
			if guile.IsScout() {
				hasScout = true
			}
		}
		if hasScout && !hasTraitor {
			deckPos = pos.CardAll.DeckTac
		}
	} else if len(botHand.Guiles) == 1 && botHand.Guiles[0].IsScout() && playableTacNo > 0 {
		drawFirst := botix
		deckHandTroops := dht.NewCache(deck.Troops(), [2][]card.Troop{hands[0].Troops, hands[1].Troops}, deck.DrawNos(drawFirst))
		flagsAna := analyzeFlags(flags, deckHandTroops, botix)
		analyzeFlagsAddFlagValue(flagsAna)
		analyzeFlagsAddLooseGameFlags(flagsAna)
		keep := NewKeep(flagsAna, nil, deckHandTroops, botix)
		if keep.CalcIsHandGood(flagsAna, 2) {
			deckPos = pos.CardAll.DeckTac
		}
	}
	return deckPos
}

//scoutReturnTacs returns maximum of two tactic cards that can be returned without problem.
//Prioritised that least valued card first.
func scoutReturnTacs(hand *card.Cards, isPlayLeader bool) (tacs []card.Card) {
	tacs = make([]card.Card, 0, 2)
	if hand.Contain(card.TCRedeploy) {
		tacs = append(tacs, card.TCRedeploy)
	}
	var leaders []card.Morale
	for _, morale := range hand.Morales {
		if !isPlayLeader {
			if morale.IsLeader() {
				tacs = append(tacs, card.Card(morale))
				break
			}
		} else {
			if morale.IsLeader() {
				leaders = append(leaders, morale)
			}
		}
	}
	if len(leaders) == 2 {
		tacs = append(tacs, card.Card(leaders[0]))
	}
	if len(tacs) < 2 {
		noLeft := hand.NoTacs() - len(tacs)
		if noLeft > 1 {
			leftTacixs := make([]card.Card, 0, noLeft)
			for _, handTac := range hand.Tacs() {
				addTac := true
				for _, tac := range tacs {
					if handTac == tac {
						addTac = false
					}
				}
				if addTac {
					leftTacixs = append(leftTacixs, handTac)
				}
			}
			pixs := tacsPrioritize(leftTacixs)
			for i := len(pixs) - 1; i > 0; i-- {
				tacs = append(tacs, leftTacixs[pixs[i]])
			}
		}
	}

	return tacs
}

//tacsPrioritize returns prioritized indices for a list of tactic cards
//Best first.
func tacsPrioritize(tacs []card.Card) (pixs []int) {
	pixs = make([]int, 0, len(tacs))
	if len(tacs) != 0 {
		for _, ptac := range tacsPRI {
			for i, tact := range tacs {
				if tact == ptac {
					pixs = append(pixs, i)
					break
				}
			}
			if len(pixs) == len(tacs) {
				break
			}
		}
	}
	return pixs
}

//lostFlagTacticMove handle the strategi of preventing a lost flag to become lost.
//Deserter strategi just remove the best troop if it gives a probability of a win.
//Best troop is morale or mid card. for straight and higist for battalion and sum.
func lostFlagTacticMove(
	flagsAna map[int]*fa.Analysis,
	priLiveFlagixs []int,
	botHand *card.Cards,
	playTacAna *playTacAna,
	deckTroopNo int,
	sim *Sim,
) (moveix int) {
	moveix = -1
	if playTacAna.botNo > 0 {
		noBotHandTacs := botHand.NoTacs()
		if noBotHandTacs > 2 && deckTroopNo > 1 && botHand.Contain(card.TCScout) {
			if botHand.Contain(card.TCTraitor) || noBotHandTacs == 4 {
				moveix = sim.FindScoutMove(pos.CardAll.DeckTroop)
			}
		} else {
			botHandTacs := botHand.Tacs()
			priHandTacs := tacsPrioritize(botHandTacs)
			for _, pTac := range priHandTacs {
				botHandTac := botHandTacs[pTac]
				switch botHandTac {
				case card.TCRedeploy:
					fallthrough
				case card.TCTraitor:
					moveix = lostFlagTacDbFlag(flagsAna, priLiveFlagixs, card.Guile(botHandTac), sim)
				case card.TCDeserter:
					moveix = lostFlagTacDeserter(flagsAna, priLiveFlagixs, botHand, sim)
				case card.TCScout:

				default:
					moveix = lostFlagTacHandFlag(flagsAna, priLiveFlagixs, botHandTac, playTacAna, sim)
				}
				if moveix != -1 {
					break
				}
			}
			if moveix == -1 && noBotHandTacs == 2 && deckTroopNo > 1 &&
				botHand.Contain(card.TCScout) && botHand.Contain(card.TCTraitor) {
				moveix = sim.FindScoutMove(pos.CardAll.DeckTroop)
			}
		}
	}
	return moveix
}

//lostFlagDumpMove handle the strategi of dumbing card on a lost flag.
//Dump the smallest card if it exist of card not usable in the best formation or in case
//of the best formation being wedge also the phalanx.
func lostFlagDumpMove(
	flagsAna map[int]*fa.Analysis,
	keep *Keep,
	moves Moves) (moveix int) {
	moveix = -1
	lostFlagix := -1
	for _, flagix := range keep.PriPlayFlagixs {
		ana := flagsAna[flagix]
		if ana.IsLost {
			lostFlagix = ana.Flagix
			break
		}
	}
	if lostFlagix != -1 {
		troop := keep.RequestFlagHandLowestStrenght()
		if !troop.IsNone() {
			log.Printf(log.Debug, "Made a Lost Flag Dump move flagix: %v, troop: %v", lostFlagix, troop)
			_, moveix = moves.FindHandFlag(lostFlagix, troop)
		}
	}
	return moveix
}
func prioritizedMove(
	flagsAna map[int]*fa.Analysis,
	keep *Keep,
	playTacAna *playTacAna,
	botHand *card.Cards,
	deckTroopNo int,
	sim *Sim,
) (moveix int) {
	moveix = -1
	troop, flagix := pri3CardsMove(keep.PriPlayFlagixs, flagsAna)
	if !troop.IsNone() {
		_, moveix = sim.FindHandFlag(flagix, troop)
		return moveix
	}
	troop, flagix = pri2CardsMove(keep.PriPlayFlagixs, flagsAna)
	if !troop.IsNone() {
		_, moveix = sim.FindHandFlag(flagix, troop)
		return moveix
	}
	troop, flagix = pri1CardMove(keep.PriPlayFlagixs, flagsAna)
	if !troop.IsNone() {
		_, moveix = sim.FindHandFlag(flagix, troop)
		return moveix
	}

	moveix = keep.NewFlagMoveix
	if moveix != -1 {
		log.Printf(log.Debug, "Makes a new flag move ix: %v", moveix)
		return moveix
	}
	moveix = priTacticMove(flagsAna, keep, playTacAna, botHand, deckTroopNo, sim)
	if moveix != -1 {
		return moveix
	}

	troop, flagix = priFlagNLoop(keep.PriPlayFlagixs, flagsAna, 1, pfnf2Pick1Card)
	if !troop.IsNone() {
		_, moveix = sim.FindHandFlag(flagix, troop)
		return moveix
	}

	troop, flagix = priDumpMove(flagsAna, keep)
	if !troop.IsNone() {
		_, moveix = sim.FindHandFlag(flagix, troop)
		return moveix
	}

	return moveix

}
func pri3CardsMove(flagixs []int, flagsAna map[int]*fa.Analysis) (troop card.Card, flagix int) {
	troop, flagix = priNCardsMove(3, flagixs, flagsAna)
	return troop, flagix
}
func priNCardsMove(n int, flagixs []int, flagsAna map[int]*fa.Analysis) (troop card.Card, flagix int) {
	troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfBestWinCard)
	if troop.IsNone() {
		troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfTopRankCard)
		if troop.IsNone() {
			troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfButtomWedge)
		}
	}
	return troop, flagix
}
func priFlagLoop(
	flagixs []int,
	flagsAna map[int]*fa.Analysis,
	pff func(flagAna *fa.Analysis) (troop card.Card, logTxt string)) (moveTroop card.Card, moveFlagix int) {
	moveFlagix = -1
	for _, flagix := range flagixs {
		flagAna := flagsAna[flagix]
		troop, logTxt := pff(flagAna)
		if !troop.IsNone() {
			logTxt = logTxt + fmt.Sprintf("\nFlagix: %v, Troop: %v", flagix, card.Troop(troop))
			log.Print(log.Debug, logTxt)
			moveFlagix = flagix
			moveTroop = troop
			break
		}
	}
	return moveTroop, moveFlagix
}
func priFlagNLoop(
	flagixs []int,
	flagsAna map[int]*fa.Analysis,
	n int,
	pfnf func(flagAna *fa.Analysis) (troop card.Card, logTxt string)) (moveTroop card.Card, moveFlagix int) {

	moveTroop, moveFlagix = priFlagLoop(flagixs, flagsAna, func(flagAna *fa.Analysis) (fTroop card.Card, logTxt string) {
		if flagAna.BotFormationSize == n {
			fTroop, logTxt = pfnf(flagAna)
		}
		return fTroop, logTxt
	})
	return moveTroop, moveFlagix
}
func pri2CardsMove(flagixs []int, flagsAna map[int]*fa.Analysis) (troop card.Card, flagix int) {
	troop, flagix = priNCardsMove(2, flagixs, flagsAna)
	return troop, flagix
}
func pri1CardMove(flagixs []int, flagsAna map[int]*fa.Analysis) (troop card.Card, flagix int) {
	n := 1
	troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfBestWinCard)
	if troop.IsNone() {
		troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfTopRankCard)
		if troop.IsNone() {
			troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfWedgeConnector)
			if troop.IsNone() {
				troop, flagix = priFlagNLoop(flagixs, flagsAna, n, pfnfMadePhalanx)
			}
		}
	}
	return troop, flagix
}

//priTacticMove makes a tactic card move because no good troop move exist.
//Only morale cards, scout
//Fog,Mud,deserter,traitor,redeploy is for defence.
//Morale strategi: cards out if morale cards leads to a win play it.
//Scout strategi: have many in keep try to get traitor. Got traitor or 3 tactics + scout play it.
//Just play it if not playing try to get traitor(scout + 1 or more tactics)
func priTacticMove(
	flagsAna map[int]*fa.Analysis,
	keep *Keep,
	playTacAna *playTacAna,
	botHand *card.Cards,
	deckTroopNo int,
	sim *Sim,
) (moveix int) {
	moveix = -1
	noHandTacs := botHand.NoTacs()
	if playTacAna.botNo > 0 && noHandTacs > 0 {
		if noHandTacs > 1 && deckTroopNo > 1 && botHand.Contain(card.TCScout) {
			if botHand.Contain(card.TCTraitor) || noHandTacs == 4 {
				moveix = sim.FindScoutMove(pos.CardAll.DeckTroop)
			}
		} else if noHandTacs == 1 && botHand.Contain(card.TCScout) && deckTroopNo > 1 {
			if !keep.CalcIsHandGood(flagsAna, 2) {
				moveix = sim.FindScoutMove(pos.CardAll.DeckTroop)
			}
		} else {
			moveix = priMoraleMoveSim(flagsAna, keep.PriPlayFlagixs, playTacAna, botHand.Morales, sim)
		}
	}
	return moveix
}

//priMoraleMoveSim simulate morale tactic cards and create move
//if win exist.
func priMoraleMoveSim(
	flagsAna map[int]*fa.Analysis,
	priFlagixs []int,
	playTacAna *playTacAna,
	botHandMorales []card.Morale,
	sim *Sim,
) (moveix int) {

	moveix = -1
Loop:
	for _, morale := range botHandMorales {
		isLeader := morale.IsLeader()
		if (isLeader && playTacAna.IsBotLeader) || !isLeader {
			for _, flagix := range priFlagixs {
				flagAna := flagsAna[flagix]
				if !flagAna.IsNewFlag {
					simMove, simMoveix := sim.FindHandFlag(flagAna.Flagix, card.Card(morale))
					_, simFlagAna := sim.Move(simMove)
					if simFlagAna.IsWin {
						moveix = simMoveix
						break Loop
					}
				}
			}
		}

	}
	return moveix
}
func priDumpMove(
	flagsAna map[int]*fa.Analysis,
	keep *Keep,
) (troop card.Card, flagix int) {

	flagix = keep.PriPlayFlagixs[len(keep.PriPlayFlagixs)-1]
	troop = card.Card(keep.DemandDump(flagsAna, flagix))
	return troop, flagix
}

func pfnf2Pick1Card(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made 2 pick on one card flag"
	for _, combiAna := range flagAna.RankAnas {
		if combiAna.Comb.Formation.Value >= card.FPhalanx.Value && len(combiAna.Playables) > 0 {
			troop = card.Card(maxPlayables(combiAna.Playables))
			break
		}
	}
	return troop, logTxt
}

func maxPlayables(troops []card.Troop) (troop card.Troop) {
	if len(troops) > 0 {
		return troops[0]
	}
	return troop
}

func pfnfButtomWedge(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made a n flag cards move: Buttom wedge card"
	if !flagAna.IsTargetMade {
		for _, combiAna := range flagAna.RankAnas {
			if combiAna.Comb.Formation.Value == card.FWedge.Value {
				if combiAna.Prop == 1 {
					troop = card.Card(maxPlayables(combiAna.Playables))
					break
				}
			} else {
				break
			}

		}
	}
	return troop, logTxt
}

func pfnfWedgeConnector(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made a wedge connector move"
	if flagAna.BotMaxRank <= combi.LastFormationRank(card.FWedge, flagAna.FormationSize) {
		cardWedges := make(map[card.Troop]int)
		for _, combiAna := range flagAna.RankAnas {
			if combiAna.Comb.Formation.Value == card.FWedge.Value {
				if len(combiAna.Playables) > 0 {
					for _, combiTroop := range combiAna.Playables {
						cardWedges[combiTroop] = cardWedges[combiTroop] + 1
					}
				}
			} else {
				break
			}
		}
		for combiTroop, wedgeNo := range cardWedges {
			if wedgeNo == flagAna.FormationSize-1 {
				troop = card.Card(combiTroop)
				break
			}
		}
	}
	return troop, logTxt
}

func pfnfMadePhalanx(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made a maded phalanx or higher move"
	if !flagAna.IsTargetMade {
		for _, combiAna := range flagAna.RankAnas {
			if combiAna.Comb.Formation.Value == card.FPhalanx.Value {
				if combiAna.Prop == 1 && len(combiAna.Playables) != 0 {
					troop = card.Card(combiAna.Playables[0])
					break
				}

			} else if combiAna.Comb.Formation.Value < card.FPhalanx.Value {
				break
			}
		}
	}
	return troop, logTxt
}
func pfnfBestWinCard(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made a n flag cards move: Best wining card."

	targetRank := flagAna.TargetRank
	if flagAna.IsTargetMade && combi.RankTieBreaker(targetRank, flagAna.FormationSize).IsRank() {
		targetRank = targetRank - 1
	}
	cardSumProp := make(map[card.Troop]float64)
	for _, combiAna := range flagAna.RankAnas {
		if combiAna.Comb.Rank <= targetRank && combiAna.Prop > 0 {
			for _, troop := range combiAna.Playables {
				cardSumProp[troop] = cardSumProp[troop] + combiAna.Prop
			}
		}
	}
	troop = probMaxSumTroop(cardSumProp)
	log.Printf(log.Debug, "card sum prob: %v", cardSumProp)

	return troop, logTxt
}

func pfnfTopRankCard(flagAna *fa.Analysis) (troop card.Card, logTxt string) {
	logTxt = "Made a n flag cards move: Top ranked card"
	if !flagAna.IsFog {
		for _, combiAna := range flagAna.RankAnas {
			if combiAna.Prop > 0 {
				troop = card.Card(maxPlayables(combiAna.Playables))
				break
			}
			if combiAna.Comb.Formation.Value < card.FPhalanx.Value {
				break
			}
		}
	}
	return troop, logTxt
}
func lostFlagTacHandFlag(
	flagsAna map[int]*fa.Analysis,
	priLiveFlagixs []int,
	tac card.Card,
	playTacAna *playTacAna,
	sim *Sim,
) (moveix int) {
	moveix = -1
	for _, flagix := range priLiveFlagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsLost {
			isSim := false
			if tac.IsMorale() && flagAna.IsPlayable && !flagAna.IsNewFlag {
				morale := card.Morale(tac)
				if !morale.IsLeader() {
					isSim = true
				} else {
					isSim = playTacAna.IsBotLeader
				}
			} else if tac.IsEnv() {
				isSim = true
			}
			if isSim {
				simMove, simMoveix := sim.FindHandFlag(flagix, tac)
				_, simFlagAna := sim.Move(simMove)
				if flagAna.IsLoosingGame && !simFlagAna.IsLost || simFlagAna.IsWin {
					log.Printf(log.Debug, "Lost flag tactic to flag move. IsLost: %v, IsWin: %v", simFlagAna.IsLost, simFlagAna.IsWin)
					moveix = simMoveix
					break
				}
			}
		}
	}
	return moveix
}

//lostFlagTacDeserter makes a deserter move.
//Deserter strategi just remove the best troop if it gives a win or prevent losing game.
//Best troop is morale or mid card. for straight and highest for battalion and sum where morale is morale max strenght.
//For mud or fog we simulate.
func lostFlagTacDeserter(
	flagsAna map[int]*fa.Analysis,
	priLiveFlagixs []int,
	botHand *card.Cards,
	sim *Sim,
) (moveix int) {
	moveix = -1
	oppix := opp(flagsAna[0].Playix)
Loop:
	for _, flagix := range priLiveFlagixs {
		flagAna := flagsAna[flagix]
		if flagAna.IsLost {
			oppEnvs := flagAna.Flag.Players[oppix].Envs
			killCards := make([]card.Card, 0, 3)
			for _, e := range oppEnvs {
				killCards = append(killCards, card.Card(e))
			}
			opp := flagAna.Flag.Players[oppix]
			killFormationCard := finDeserterKillCard(opp.Troops, opp.Morales, flagAna.TargetRank, flagAna.FormationSize)
			killCards = append(killCards, killFormationCard)
			preventLossMoveix := -1
			for _, killCard := range killCards {
				killMove, killMoveix := sim.FindDeserterMove(killCard)
				simAna, _ := sim.Move(killMove)
				if simAna.IsWin {
					log.Printf(log.Debug, "Deserter win kill move card: %v", killCard)
					moveix = killMoveix
					break Loop
				}
				if !simAna.IsLost {
					preventLossMoveix = killMoveix
				}
			}
			if preventLossMoveix != -1 && flagAna.IsLoosingGame {
				log.Printf(log.Debug, "Deserter prevent kill move card: %v", preventLossMoveix)
				moveix = preventLossMoveix
				break Loop
			}
		}
	}
	return moveix
}
func finDeserterKillCard(
	oppTroops []card.Troop,
	oppMorales []card.Morale,
	targetRank,
	formationSize int) (desertCard card.Card) {

	combination := combi.Combinations(formationSize)[targetRank-1]
	switch combination.Formation.Value {
	case card.FWedge.Value:
		fallthrough
	case card.FSkirmish.Value:
		if len(oppMorales) > 0 {
			desertCard = card.Card(findMaxStrenghtMorale(oppMorales))
		} else {
			desertCard = card.Card(oppTroops[1])
		}
	case card.FPhalanx.Value:
		if len(oppMorales) > 0 {
			desertCard = card.Card(findMaxStrenghtMorale(oppMorales))
		} else {
			desertCard = card.Card(oppTroops[0])
		}
	case card.FHost.Value:
		fallthrough
	case card.FBattalion.Value:
		desertCard = deserterKillStrenght(oppMorales, oppTroops[0])
	}
	return desertCard
}
func findMaxStrenghtMorale(morales []card.Morale) (max card.Morale) {
	for _, morale := range morales {
		if max != 0 {
			if morale.MaxStrenght() > max.MaxStrenght() {
				max = morale
			}
		} else {
			max = morale
		}
	}
	return max
}
func deserterKillStrenght(morales []card.Morale, troop card.Troop) (desert card.Card) {
	if len(morales) > 0 {
		morale := findMaxStrenghtMorale(morales)
		if morale.MaxStrenght() >= troop.Strenght() {
			desert = card.Card(morale)
		} else {
			desert = card.Card(troop)
		}
	} else {
		desert = card.Card(troop)
	}
	return desert
}

func mudTrimDish(
	troops []card.Troop,
	morales []card.Morale,
	isFog bool) (dishCard card.Card) {
	formationSize := 3
	lowestRank := 1000
	higestSum := 0
	var handTroops [2][]card.Troop
	drawNos := [2]int{0, 0}
	deckHandTroops := dht.NewCache(nil, handTroops, drawNos)
	playix := 0
	targetSum := 41 //Do not think it matter too high or too low
	targetRank := 1
	moraleTroopSum := fa.MoraleTroopsSum(troops, morales)
	for _, outTroop := range troops {
		simTroops := make([]card.Troop, 0, len(troops)-1)
		for _, simTroop := range troops {
			if simTroop != outTroop {
				simTroops = append(simTroops, simTroop)
			}
		}
		rank := fa.CalcMaxRank(simTroops, morales, deckHandTroops, playix, formationSize, isFog, targetRank, targetSum)
		sum := 0
		rankTieBreaker := combi.RankTieBreaker(rank, formationSize)
		if !rankTieBreaker.IsRank() {
			sum = moraleTroopSum - outTroop.Strenght()
		}
		if rank < lowestRank {
			lowestRank = rank
			dishCard = card.Card(outTroop)
			higestSum = sum
		} else if lowestRank == rank && !rankTieBreaker.IsRank() {
			if sum > higestSum {
				dishCard = card.Card(outTroop)
				higestSum = sum
			}
		}
	}
	for _, outMorale := range morales {
		simMorales := make([]card.Morale, 0, len(morales)-1)
		for _, simMorale := range morales {
			if simMorale != outMorale {
				simMorales = append(simMorales, simMorale)
			}
		}
		rank := fa.CalcMaxRank(troops, simMorales, deckHandTroops, playix, formationSize, isFog, targetRank, targetSum)
		rankTieBreaker := combi.RankTieBreaker(rank, formationSize)
		sum := 0
		if !rankTieBreaker.IsRank() {
			sum = moraleTroopSum - outMorale.MaxStrenght()
		}
		if rank < lowestRank {
			lowestRank = rank
			dishCard = card.Card(outMorale)
			higestSum = sum
		} else if lowestRank == rank && !rankTieBreaker.IsRank() {
			if sum > higestSum {
				dishCard = card.Card(outMorale)
				higestSum = sum
			}
		}
	}

	return dishCard
}
func mudTrim(posCards [][]card.Card, mudPos pos.Card) [][]card.Card {
	if mudPos.IsOnTable() {
		mudFlagix := mudPos.Flagix()
		isTrimed := false
		for flagix := range pos.CardAll.Players[0].Flags {
			if flagix != mudFlagix {
				posCards, isTrimed = mudTrimFlag(posCards, flagix)
				if isTrimed {
					break
				}
			}
		}
	}
	return posCards
}
func mudTrimFlag(posCards [][]card.Card, flagix int) ([][]card.Card, bool) {
	isTrimed := false
	for _, playerPos := range pos.CardAll.Players {
		cards := game.PosCards(posCards).SortedCards(playerPos.Flags[flagix])
		if cards.NoFormation() > 3 {
			isTrimed = true
			dishCard := mudTrimDish(cards.Troops, cards.Morales, cards.Contain(card.TCFog))
			posix := int(playerPos.Flags[flagix])
			updCards := make([]card.Card, 0, len(posCards[posix]))
			for _, posCard := range posCards[posix] {
				if posCard != dishCard {
					updCards = append(updCards, posCard)
				}
			}
			posCards[posix] = updCards
			dishPosix := int(playerPos.Dish)
			posCards[dishPosix] = append(posCards[dishPosix], dishCard)
		}
	}
	return posCards, isTrimed
}

func createHandFlagMove(flagix int, cardMove card.Card, mover int) *game.Move {
	return game.CreateMoveHand(int(cardMove), flagix, mover)
}
