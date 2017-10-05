package game

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	math "github.com/rezder/go-math/int"
	slice "github.com/rezder/go-slice/int"
)

func calcMovesLoop(mover int,
	moveType MoveType,
	conePos [10]pos.Cone,
	cardPos [71]pos.Card,
	posCards PosCards) ([]*Move, PosCards) {
	var moves []*Move
	switch moveType {
	case MoveTypeAll.Hand:
		moves, posCards = createMovesHand(cardPos, conePos, posCards, mover)
	case MoveTypeAll.ScoutReturn:
		moves = createMovesScoutReturn(cardPos, mover)
	case MoveTypeAll.Cone:
		moves, posCards = createMovesCone(cardPos, conePos, posCards, mover)
	case MoveTypeAll.Scout2:
		fallthrough
	case MoveTypeAll.Scout3:
		moves = createMovesDeck(cardPos, mover, true)
		for _, move := range moves {
			move.MoveType = moveType
		}
	case MoveTypeAll.Deck:
		moves = createMovesDeck(cardPos, mover, false)
		for _, move := range moves {
			move.MoveType = moveType
		}

	default:
		panic(fmt.Sprintf("Move type: %v does not have a calulation", moveType))
	}
	return moves, posCards
}
func createMovesDeck(cardPos [71]pos.Card, mover int, isScout bool) (moves []*Move) {
	noTac := 0
	noTroop := 0
	noHand := 0
	handPos := pos.CardAll.Players[mover].Hand
	for i, oldPos := range cardPos {
		if i != 0 {
			if oldPos.IsInDeck() {
				if i > card.NOTroop {
					noTac++
				} else {
					noTroop++
				}
			} else if oldPos == handPos {
				noHand++
			}
		}
	}
	if noHand < 7 || isScout {
		if noTac > 0 {
			move := createMoveDeck(true, mover)
			moves = append(moves, move)
		}
		if noTroop > 0 {
			move := createMoveDeck(false, mover)
			moves = append(moves, move)
		}
	}
	return moves
}
func createMoveDeck(isTac bool, mover int) *Move {
	move := NewMove(mover, MoveTypeAll.Deck)
	backix := card.BACKTroop
	oldPos := pos.CardAll.DeckTroop
	if isTac {
		backix = card.BACKTac
		oldPos = pos.CardAll.DeckTac
	}
	move.Moves = append(move.Moves, &BoardPieceMove{
		BoardPiece: BoardPieceAll.Card,
		Index:      backix,
		OldPos:     uint8(oldPos),
		NewPos:     uint8(pos.CardAll.Players[mover].Hand),
	})
	return move
}
func createMovesCone(cardPos [71]pos.Card,
	conePos [10]pos.Cone,
	posCards PosCards,
	mover int) ([]*Move, PosCards) {

	if posCards == nil {
		posCards = NewPosCards(cardPos)
	}
	moves := make([]*Move, 0, 0)
	var flagixs []int
	for i := 0; i < 9; i++ {
		if NewFlag(i, posCards, conePos).HasFormation(mover) {
			flagixs = append(flagixs, i)
		}
	}
	if len(flagixs) > 0 {
		moves = coneCombiMoves(flagixs, mover)
	}

	return moves, posCards
}
func coneCombiMoves(flagixs []int, mover int) []*Move {
	noFlagixs := len(flagixs)
	m := make([]*Move, 0, coneCombiNo(noFlagixs))
	m = append(m, CreateMoveCone(nil, mover))        //no claims
	m = append(m, CreateMoveCone(flagixs[:], mover)) // all
	switch noFlagixs {
	case 2:
		m = coneCombiAddOne(m, false, flagixs, mover)
	case 3:
		m = coneCombiAddOne(m, true, flagixs, mover)
	case 4:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, false, flagixs, 2, mover)
	case 5:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, true, flagixs, 2, mover)
	case 6:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, true, flagixs, 2, mover)
		m = coneCombiAdd(m, false, flagixs, 3, mover)
	case 7:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, true, flagixs, 2, mover)
		m = coneCombiAdd(m, true, flagixs, 3, mover)
	case 8:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, true, flagixs, 2, mover)
		m = coneCombiAdd(m, true, flagixs, 3, mover)
		m = coneCombiAdd(m, false, flagixs, 4, mover)
	case 9:
		m = coneCombiAddOne(m, true, flagixs, mover)
		m = coneCombiAdd(m, true, flagixs, 2, mover)
		m = coneCombiAdd(m, true, flagixs, 3, mover)
		m = coneCombiAdd(m, true, flagixs, 4, mover)
	}
	return m
}
func coneCombiAddOne(m []*Move, reverse bool, flagixs []int, mover int) []*Move {
	for i := range flagixs {
		m = append(m, CreateMoveCone(flagixs[i:i+1], mover))
		if reverse {
			m = append(m, CreateMoveCone(slice.WithOutNew(flagixs, flagixs[i:i+1]), mover))
		}
	}
	return m
}

func coneCombiAdd(m []*Move, reverse bool, allFlagixs []int, d, mover int) []*Move {
	n := len(allFlagixs)
	_ = math.Perm(n, d, func(per []int) bool {
		var moveFlagixs = make([]int, d)
		for i, ix := range per {
			moveFlagixs[i] = allFlagixs[ix]
		}
		m = append(m, CreateMoveCone(moveFlagixs, mover))
		if reverse {
			m = append(m, CreateMoveCone(slice.WithOutNew(allFlagixs, moveFlagixs), mover))
		}
		return false
	})
	return m
}

// coneCombiNo calculate the number of combination you
// can make with x number of claimable flags.
func coneCombiNo(flagsNo int) (no int) {
	switch flagsNo {
	case 1:
		no = 2
	case 2:
		no = 2 + flagsNo
	case 3:
		no = 2 + (flagsNo * 2)
	case 4:
		no = 2 + (flagsNo * 2) + int(math.Comb(uint64(flagsNo), uint64(2)))
	case 5:
		no = 2 + (flagsNo * 2) + 2*int(math.Comb(uint64(flagsNo), uint64(2)))
	case 6:
		no = 2 + (flagsNo * 2) + 2*int(math.Comb(uint64(flagsNo), uint64(2))) + int(math.Comb(uint64(flagsNo), uint64(3)))
	case 7:
		no = 2 + (flagsNo * 2) + 2*int(math.Comb(uint64(flagsNo), uint64(2))+math.Comb(uint64(flagsNo), uint64(3)))
	case 8:
		no = 2 + (flagsNo * 2) + 2*int(math.Comb(uint64(flagsNo), uint64(2))+math.Comb(uint64(flagsNo), uint64(3))) + int(math.Comb(uint64(flagsNo), uint64(4)))
	case 9:
		no = 2 + (flagsNo * 2) + 2*int(math.Comb(uint64(flagsNo), uint64(2))+math.Comb(uint64(flagsNo), uint64(3))+math.Comb(uint64(flagsNo), uint64(4)))
	}
	return no
}

//CreateMoveGivUp creates a give up move
func CreateMoveGivUp(conePos [10]pos.Cone, mover int) *Move {
	move := NewMove(mover, MoveTypeAll.GiveUp)
	for ix, cp := range conePos {
		if !cp.IsWon() {
			move.Moves = append(move.Moves, &BoardPieceMove{
				BoardPiece: BoardPieceAll.Cone,
				Index:      ix,
				OldPos:     uint8(pos.ConeAll.None),
				NewPos:     uint8(pos.ConeAll.Players[opp(mover)]),
			})
		}
	}
	return move
}

//CreateMoveCone creates a cone move.
func CreateMoveCone(flagixs []int, mover int) *Move {
	move := NewMove(mover, MoveTypeAll.Cone)
	for _, flagix := range flagixs {
		move.Moves = append(move.Moves, &BoardPieceMove{
			BoardPiece: BoardPieceAll.Cone,
			Index:      flagix + 1,
			OldPos:     uint8(pos.ConeAll.None),
			NewPos:     uint8(pos.ConeAll.Players[mover]),
		})
	}
	return move
}
func FlagsCreate(posCards PosCards, conePos [10]pos.Cone) (flags [9]*Flag) {
	for i := 0; i < 9; i++ {
		flags[i] = NewFlag(i, posCards, conePos)
	}
	return flags
}
func createMovesHandFlagTroops(flagix int, flag *Flag, handTroops []card.Troop, mover int) (moves []*Move) {
	moves = make([]*Move, 0, len(handTroops))
	for _, troop := range handTroops {
		if flag.IsTroopPlayable(mover) {
			move := CreateMoveHand(int(troop), flagix, mover)
			moves = append(moves, move)
		}
	}
	return moves
}
func createMovesHandFlagMorales(flagix int, flag *Flag, morales []card.Morale, mover int, isLeaderPlayable bool) (moves []*Move) {
	moves = make([]*Move, 0, len(morales))
	for _, morale := range morales {
		if flag.IsMoralePlayable(mover) {
			if !morale.IsLeader() || (morale.IsLeader() && isLeaderPlayable) {
				move := CreateMoveHand(int(morale), flagix, mover)
				moves = append(moves, move)
			}
		}
	}
	return moves
}
func createMovesHandFlagEnvs(flagix int, flag *Flag, envs []card.Env, mover int) (moves []*Move) {
	moves = make([]*Move, 0, len(envs))
	for _, env := range envs {
		if flag.IsEnvPlayable(mover) {
			move := CreateMoveHand(int(env), flagix, mover)
			moves = append(moves, move)
		}
	}
	return moves
}
func createMovesHand(cardPos [71]pos.Card, conePos [10]pos.Cone, posCards PosCards, mover int) ([]*Move, PosCards) {
	moves := make([]*Move, 0, 0)
	if posCards == nil {
		posCards = NewPosCards(cardPos)
	}
	moves = make([]*Move, 0, 0)
	flags := FlagsCreate(posCards, conePos)

	hand := posCards.SortedCards(pos.CardAll.Players[mover].Hand)
	isTacPlayable, isLeaderPlayable := anaPlayTacs(cardPos[card.NOTroop+1:], mover)
	noTroopMoves := 0
	for flagix, flag := range flags {
		troopMoves := createMovesHandFlagTroops(flagix, flag, hand.Troops, mover)
		noTroopMoves = noTroopMoves + len(troopMoves)
		moves = append(moves, troopMoves...)
		if isTacPlayable {
			moves = append(moves, createMovesHandFlagMorales(flagix, flag, hand.Morales, mover, isLeaderPlayable)...)
			moves = append(moves, createMovesHandFlagEnvs(flagix, flag, hand.Envs, mover)...)
		}
	}
	if len(hand.Guiles) > 0 && isTacPlayable {
		moves = append(moves, createMovesGuile(hand.Guiles, cardPos, flags, mover)...)
	}
	if len(moves) > 0 && noTroopMoves == 0 { //Pass move
		moves = append(moves, NewMove(mover, MoveTypeAll.Hand))
	}
	return moves, posCards
}

func createMovesDeserter(dishGuileMove *BoardPieceMove, flags [9]*Flag, mover int) (moves []*Move) {
	opponent := opp(mover)
	for _, flag := range flags {
		if !flag.IsWon {
			for _, troop := range flag.Players[opponent].Troops {
				moves = append(moves, CreateMoveDeserter(flag.Positions[opponent], int(troop), mover, dishGuileMove))
			}
			for _, morale := range flag.Players[opponent].Morales {
				moves = append(moves, CreateMoveDeserter(flag.Positions[opponent], int(morale), mover, dishGuileMove))
			}
			for _, env := range flag.Players[opponent].Envs {
				moves = append(moves, CreateMoveDeserter(flag.Positions[opponent], int(env), mover, dishGuileMove))
			}
		}
	}
	return moves
}

//CreateMoveDeserter creates a deserter move.
func CreateMoveDeserter(oldPos pos.Card, killedCardix, mover int, dishGuileMove *BoardPieceMove) (move *Move) {
	move = NewMove(mover, MoveTypeAll.Hand)
	move.Moves = append(move.Moves, dishGuileMove)
	move.Moves = append(move.Moves, CreateBPMoveDish(killedCardix, opp(mover), oldPos))
	return move
}
func createMovesRedeployInFlag(outFlag, inFlag *Flag, mover int, dishGuileMove *BoardPieceMove) (moves []*Move) {
	if outFlag != inFlag {
		if inFlag.IsTroopPlayable(mover) {
			for _, troop := range outFlag.Players[mover].Troops {
				moves = append(moves,
					CreateMoveDouble(
						outFlag.Positions[mover],
						inFlag.Positions[mover],
						int(troop),
						mover,
						dishGuileMove))
			}
		}
		if inFlag.IsMoralePlayable(mover) {
			for _, morale := range outFlag.Players[mover].Morales {
				moves = append(moves,
					CreateMoveDouble(
						outFlag.Positions[mover],
						inFlag.Positions[mover],
						int(morale),
						mover,
						dishGuileMove))
			}
		}
		if inFlag.IsEnvPlayable(mover) {
			for _, env := range outFlag.Players[mover].Envs {
				moves = append(moves,
					CreateMoveDouble(
						outFlag.Positions[mover],
						inFlag.Positions[mover],
						int(env),
						mover,
						dishGuileMove))
			}
		}
	}
	return moves
}
func createMovesRedeployDish(outFlag *Flag, mover int, dishGuileMove *BoardPieceMove) (moves []*Move) {
	dishPos := pos.CardAll.Players[mover].Dish
	for _, troop := range outFlag.Players[mover].Troops {
		moves = append(moves,
			CreateMoveDouble(
				outFlag.Positions[mover],
				dishPos,
				int(troop),
				mover,
				dishGuileMove))
	}

	for _, morale := range outFlag.Players[mover].Morales {
		moves = append(moves,
			CreateMoveDouble(
				outFlag.Positions[mover],
				dishPos,
				int(morale),
				mover,
				dishGuileMove))
	}

	for _, env := range outFlag.Players[mover].Envs {
		moves = append(moves,
			CreateMoveDouble(
				outFlag.Positions[mover],
				dishPos,
				int(env),
				mover,
				dishGuileMove))
	}
	return moves
}
func createMovesRedeploy(dishGuileMove *BoardPieceMove, flags [9]*Flag, mover int) (moves []*Move) {
	for _, outFlag := range flags {
		if !outFlag.IsWon {
			for _, inFlag := range flags {
				moves = append(moves, createMovesRedeployInFlag(outFlag, inFlag, mover, dishGuileMove)...)
			}
			moves = append(moves, createMovesRedeployDish(outFlag, mover, dishGuileMove)...)
		}
	}
	return moves
}

// CreateMoveDouble creates a redeploy or traitor move.
func CreateMoveDouble(oldPos, newPos pos.Card, moveCardix, mover int, dishGuileMove *BoardPieceMove) (move *Move) {
	move = NewMove(mover, MoveTypeAll.Hand)
	move.Moves = append(move.Moves, dishGuileMove)
	move.Moves = append(move.Moves, &BoardPieceMove{
		BoardPiece: BoardPieceAll.Card,
		Index:      moveCardix,
		OldPos:     uint8(oldPos),
		NewPos:     uint8(newPos),
	})
	return move
}
func createMovesTraitor(dishGuileMove *BoardPieceMove, flags [9]*Flag, mover int) (moves []*Move) {
	for _, outFlag := range flags {
		if !outFlag.IsWon {
			for _, inFlag := range flags {
				if inFlag.IsTroopPlayable(mover) {
					for _, troop := range outFlag.Players[opp(mover)].Troops {
						moves = append(moves,
							CreateMoveDouble(
								outFlag.Positions[opp(mover)],
								inFlag.Positions[mover],
								int(troop),
								mover,
								dishGuileMove))
					}
				}
			}
		}
	}
	return moves
}

//CreateBPMoveDish create board piece move that moves a card to the dish.
func CreateBPMoveDish(cardix int, player int, oldPos pos.Card) *BoardPieceMove {
	return &BoardPieceMove{
		BoardPiece: BoardPieceAll.Card,
		Index:      cardix,
		OldPos:     uint8(oldPos),
		NewPos:     uint8(pos.CardAll.Players[player].Dish),
	}
}
func createMovesGuile(handGuiles []card.Guile, cardPos [71]pos.Card, flags [9]*Flag, mover int) (moves []*Move) {
	oldPos := pos.CardAll.Players[mover].Hand
	for _, guile := range handGuiles {
		dishMove := CreateBPMoveDish(int(guile), mover, oldPos)
		switch {
		case guile.IsDeserter():
			moves = append(moves, createMovesDeserter(dishMove, flags, mover)...)
		case guile.IsRedeploy():
			moves = append(moves, createMovesRedeploy(dishMove, flags, mover)...)
		case guile.IsScout():
			moves = append(moves, createMovesScout(cardPos, dishMove, mover)...)
		case guile.IsTraitor():
			moves = append(moves, createMovesTraitor(dishMove, flags, mover)...)
		}
	}
	return moves
}
func createMovesScout(cardPos [71]pos.Card, dishGuileMove *BoardPieceMove, mover int) (moves []*Move) {
	deckMoves := createMovesDeck(cardPos, mover, true)
	if len(deckMoves) > 0 {
		for _, deckMove := range deckMoves {
			move := CreateMoveScout(dishGuileMove, deckMove.Moves[0], mover)
			moves = append(moves, move)
		}
	}
	return moves
}

//CreateMoveScout creates a scout move.
func CreateMoveScout(dishMove, deckMove *BoardPieceMove, mover int) (move *Move) {
	move = NewMove(mover, MoveTypeAll.Scout1)
	move.Moves = append(move.Moves, dishMove)
	move.Moves = append(move.Moves, deckMove)
	return move
}

//CreateMoveHand creates a card move from the hand to a flag.
func CreateMoveHand(cardix, flagix, mover int) (move *Move) {
	move = NewMove(mover, MoveTypeAll.Hand)
	oldPos := pos.CardAll.Players[mover].Hand
	move.Moves = append(move.Moves, &BoardPieceMove{
		BoardPiece: BoardPieceAll.Card,
		Index:      cardix,
		OldPos:     uint8(oldPos),
		NewPos:     uint8(pos.CardAll.Players[mover].Flags[flagix]),
	})
	return move
}

// anaPlayTacs analyze to find out if a player can play tactic cards.
func anaPlayTacs(tacPos []pos.Card, player int) (isTacPlayable, isLeaderPlayable bool) {
	tacix := card.NOTroop + 1
	var played [3]int // 3 fo NoPlayer
	var playedLeader [3]bool
	for i, oldPos := range tacPos {
		if oldPos.IsOnTable() {
			morale := card.Morale(tacix + i)
			cardMove := card.Card(tacix + i)
			p := oldPos.Player()
			played[p] = played[p] + 1
			if cardMove.IsMorale() && morale.IsLeader() {
				playedLeader[p] = true
			}
		}
	}
	if !playedLeader[player] {
		isLeaderPlayable = true
	}
	if played[player] <= played[opp(player)] {
		isTacPlayable = true
	}
	return isTacPlayable, isLeaderPlayable
}
func createHand(player int, cardPos [71]pos.Card) (handTroops, handTacs []int) {
	handTroops = make([]int, 0, 9)
	handTacs = make([]int, 0, 9)
	for cardix, oldPos := range cardPos {
		if cardix > 0 {
			if oldPos == pos.CardAll.Players[player].Hand {
				if card.Card(cardix).IsTroop() {
					handTroops = append(handTroops, cardix)
				} else {
					handTacs = append(handTacs, cardix)
				}
			}
		}
	}
	return handTroops, handTacs
}

//createMovesScoutReturn returns the scout moves.
// The numbers of moves should be len(tac)=nta and len(troop)=nto
// (nta*nta-1) + nto*nto-1+nta*nto if there is enough cards and
// there is two cards to return.
func createMovesScoutReturn(cardPos [71]pos.Card, mover int) (moves []*Move) {
	handTroops, handTacs := createHand(mover, cardPos)
	noTacs := len(handTacs)
	noTroops := len(handTroops)
	noReturn := noTacs + noTroops - NOHandInit
	if noReturn == 2 {
		moves = make([]*Move, 0, 72)
	} else {
		moves = make([]*Move, 0, 8)
	}
	if noReturn == 2 {
		if noTacs > 1 {
			doubleLoop(handTacs, handTacs, func(a, b int, d bool) {
				if !d {
					moves = append(moves, CreateMoveScoutReturn([]int{a, b}, nil, mover))
				}
			})
		}
		if noTroops > 1 {
			doubleLoop(handTroops, handTroops, func(a, b int, d bool) {
				if !d {
					moves = append(moves, CreateMoveScoutReturn(nil, []int{a, b}, mover))
				}
			})
		}
		if noTacs > 0 && noTroops > 0 {
			doubleLoop(handTacs, handTroops, func(a, b int, _ bool) {
				moves = append(moves, CreateMoveScoutReturn([]int{a}, []int{b}, mover))
			})
		}
	}
	if noReturn == 1 {
		if noTacs > 0 {
			for _, tac := range handTacs {
				moves = append(moves, CreateMoveScoutReturn([]int{tac}, nil, mover))
			}
		}
		if noTroops > 0 {
			for _, troop := range handTroops {
				moves = append(moves, CreateMoveScoutReturn(nil, []int{troop}, mover))
			}
		}
	}

	return moves
}
func doubleLoop(av, bv []int, f func(a, b int, d bool)) {
	for i, a := range av {
		for j, b := range bv {
			f(a, b, i == j)
		}
	}
}

//CreateMoveScoutReturn creates a scout return move.
func CreateMoveScoutReturn(tacixs, troopixs []int, mover int) *Move {
	move := NewMove(mover, MoveTypeAll.ScoutReturn)
	oldPos := pos.CardAll.Players[mover].Hand
	for _, tacix := range tacixs {
		move.Moves = append(move.Moves, &BoardPieceMove{
			BoardPiece: BoardPieceAll.Card,
			Index:      tacix,
			OldPos:     uint8(oldPos),
			NewPos:     uint8(pos.CardAll.DeckTac),
		})
	}
	for _, troopix := range troopixs {
		move.Moves = append(move.Moves, &BoardPieceMove{
			BoardPiece: BoardPieceAll.Card,
			Index:      troopix,
			OldPos:     uint8(oldPos),
			NewPos:     uint8(pos.CardAll.DeckTroop),
		})
	}
	return move
}
