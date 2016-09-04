package battleline

import (
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-battleline/battleline/flag"
	"github.com/rezder/go-card/deck"
	math "github.com/rezder/go-math/int"
	slice "github.com/rezder/go-slice/int"
)

const (
	TURN_FLAG = 0
	TURN_HAND = 1
	//TURN_SCOUT2 player picks second of tree scout cards.
	TURN_SCOUT2 = 2
	//TURN_SCOUT2 player picks last of tree scout cards.
	TURN_SCOUT1 = 3
	//TURN_SCOUTR player return 3 cards to decks.
	TURN_SCOUTR = 4
	TURN_DECK   = 5
	TURN_FINISH = 6
	TURN_QUIT   = 7

	DECK_TAC   = 1
	DECK_TROOP = 2
)

// Turn hold the information of a turn, whos turn is it, what kind of turn (State) and
// the possible moves.
type Turn struct {
	Player    int
	State     int
	Moves     []Move
	MovesHand map[int][]Move
	MovePass  bool
}

//start set up the first turn.
func (turn *Turn) start(starter int, hand *Hand, flags *[FLAGS]*flag.Flag, deckTac *deck.Deck,
	deckTroop *deck.Deck, dishs *[2]*Dish) {

	turn.Player = starter
	turn.State = TURN_HAND
	turn.MovesHand, _ = getMoveHand(starter, hand, flags, deckTac, deckTroop,
		dishs[starter].Tacs, dishs[opponent(starter)].Tacs)
}
func (turn *Turn) Equal(other *Turn) (equal bool) {
	if other == nil && turn == nil {
		equal = true
	} else if other != nil && turn != nil {
		if other == turn {
			equal = true
		} else if other.Player == turn.Player && other.State == turn.State && other.MovePass == turn.MovePass {
			mequal := false
			if len(other.Moves) == 0 && len(turn.Moves) == 0 {
				mequal = true
			} else if len(other.Moves) == len(turn.Moves) {
				mequal = true
				for i, v := range other.Moves {
					if !v.MoveEqual(turn.Moves[i]) {
						mequal = false
						break
					}
				}
			}
			if mequal {
				mhequal := false
				if len(other.MovesHand) == 0 && len(turn.MovesHand) == 0 {
					mhequal = true
				} else if len(other.MovesHand) == len(turn.MovesHand) {
					mhequal = true
				Card:
					for cardix, moves := range other.MovesHand {
						turnMoves, found := turn.MovesHand[cardix]
						if found && len(moves) == len(turnMoves) {
							for i, v := range moves {
								if !v.MoveEqual(turnMoves[i]) {
									mhequal = false
									break Card
								}
							}
						} else {
							mhequal = false
							break
						}
					}
				}
				if mhequal {
					equal = true
				}
			}

		}
	}
	return equal
}
func (turn *Turn) Copy() (c *Turn) {
	if turn != nil {
		c = new(Turn)
		c.Player = turn.Player
		c.State = turn.State
		if len(turn.Moves) != 0 {
			c.Moves = make([]Move, len(turn.Moves))
			for i, v := range turn.Moves {
				c.Moves[i] = v.Copy()
			}
		}
		if len(turn.MovesHand) != 0 {
			c.MovesHand = make(map[int][]Move)
			for cardix, moves := range turn.MovesHand {
				copyMoves := make([]Move, len(moves))
				for i, move := range moves {
					copyMoves[i] = move.Copy()
				}
				c.MovesHand[cardix] = copyMoves
			}
		}
		c.MovePass = turn.MovePass
	}
	return c
}

// GetMoveix find the move index.
func (turn *Turn) GetMoveix(handCardix int, move Move) (ix int) {
	ix = -1
	if len(turn.Moves) != 0 {
		for i, d := range turn.Moves {
			if d.MoveEqual(move) {
				ix = i
				break
			}
		}
	} else if len(turn.MovesHand) != 0 {
		moves := turn.MovesHand[handCardix]
		if len(moves) != 0 {
			for i, d := range moves {
				if d.MoveEqual(move) {
					ix = i
					break
				}
			}
		}
	}
	return ix
}

//Opp the opponent to the player that have the turn.
func (turn *Turn) Opp() int {
	return opponent(turn.Player)
}

//quit set the state to quit.
func (turn *Turn) quit() {
	turn.State = TURN_QUIT
}

// next role the turn over to the next turn that need player action.
func (turn *Turn) next(handScout bool, hands *[2]*Hand, flags *[FLAGS]*flag.Flag, deckTac *deck.Deck,
	deckTroop *deck.Deck, dishs *[2]*Dish) {
	turn.Player, turn.State, turn.MovePass, turn.Moves, turn.MovesHand = updateTurn(turn.Player,
		turn.State, hands, flags, deckTac, deckTroop, dishs, handScout)
}

//updateTurn role the turn over to the next turn that need player action.
func updateTurn(oldPlayer int, oldState int, hands *[2]*Hand, flags *[FLAGS]*flag.Flag,
	deckTac *deck.Deck, deckTroop *deck.Deck, dishs *[2]*Dish, handScout bool) (player int, state int,
	movePass bool, moves []Move, movesHand map[int][]Move) {
	player = oldPlayer
	switch oldState {
	case TURN_FLAG:
		if !win(flags, oldPlayer) {
			handMap, pass := getMoveHand(player, hands[player], flags, deckTac, deckTroop,
				dishs[player].Tacs, dishs[opponent(player)].Tacs)
			state = TURN_HAND
			if len(handMap) != 0 {
				movesHand = handMap
				movePass = pass
			} else {
				player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags,
					deckTac, deckTroop, dishs, false) //I think there always is a move right!!!
			}
		} else {
			state = TURN_FINISH
		}
	case TURN_HAND:
		if handScout {
			state = TURN_SCOUT2
		} else { //deck
			state = TURN_DECK
		}
		moves = getMoveDeck(deckTac, deckTroop, hands[player], state)
		if len(moves) == 0 {
			player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags, deckTac,
				deckTroop, dishs, false)
		}

	case TURN_DECK:
		state = TURN_FLAG
		player = opponent(oldPlayer)
		moves = getMoveClaim(player, flags)
		if len(moves) == 0 {
			player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags,
				deckTac, deckTroop, dishs, false)
		}

	case TURN_FINISH:
		panic("There is no turn after finish")

	case TURN_SCOUT1:
		state = TURN_SCOUTR
		moves = getMoveScoutReturn(hands[player])
		if len(moves) == 0 {
			player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags,
				deckTac, deckTroop, dishs, false)
		}
	case TURN_SCOUT2:
		state = TURN_SCOUT1
		moves = getMoveDeck(deckTac, deckTroop, hands[player], state)
		if len(moves) == 0 {
			player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags,
				deckTac, deckTroop, dishs, false)

		}
	case TURN_SCOUTR:
		state = TURN_FLAG
		player = opponent(oldPlayer)
		moves = getMoveClaim(player, flags)
		if len(moves) == 0 {
			player, state, movePass, moves, movesHand = updateTurn(player, state, hands, flags,
				deckTac, deckTroop, dishs, false)
		}
	}
	return player, state, movePass, moves, movesHand
}

//win check if a player have met the criteria for wining a game.
func win(flags *[FLAGS]*flag.Flag, playerix int) (w bool) {
	total := 0
	row := 0
	for _, fla := range flags {
		if fla.Won()[playerix] {
			total++
			row++
		} else {
			row = 0
		}
		if row == 3 {
			w = true
			break
		} else if total == 5 {
			w = true
			break
		}
	}
	return w
}

// getMoveHand returns the possible hand moves.
func getMoveHand(playerix int, hand *Hand, flags *[FLAGS]*flag.Flag, tacDeck *deck.Deck,
	troopDeck *deck.Deck, dishTac []int, oppDishTac []int) (m map[int][]Move, pass bool) {
	m = make(map[int][]Move)
	troopSpace := make([]int, 0, FLAGS)

	used := make([]int, 0, 5)
	oppUsed := make([]int, 0, 5)
	var usedv [2][]int
	notClaimed := make([]int, 0, FLAGS)
	for i, flag := range flags {
		if !flag.Claimed() {
			notClaimed = append(notClaimed, i)
			if flag.Free()[playerix] {
				troopSpace = append(troopSpace, i)
			}
		}
		usedv = flag.UsedTac()
		used = append(used, usedv[playerix]...)
		oppUsed = append(oppUsed, usedv[opponent(playerix)]...)

	}
	for _, tac := range oppDishTac {
		oppUsed = append(oppUsed, tac)
	}
	for _, tac := range dishTac {
		used = append(used, tac)
	}

	playedLeader := slice.Contain(used, cards.TC_Alexander)
	if (!playedLeader) && slice.Contain(used, cards.TC_Darius) {
		playedLeader = true
	}

	playedTac := len(used)
	oppPlayedTac := len(oppUsed)

	playTac := playedTac <= oppPlayedTac
	var moves []Move
	for _, v := range hand.Tacs {
		if playTac {
			switch v {
			case cards.TC_123, cards.TC_8:
				moves = getCardFlagMoves(troopSpace)
			case cards.TC_Alexander, cards.TC_Darius:
				if !playedLeader {
					moves = getCardFlagMoves(troopSpace)
				} else {
					moves = nil
				}
			case cards.TC_Fog, cards.TC_Mud:
				moves = getCardFlagMoves(notClaimed)
			case cards.TC_Deserter:
				moves = getDeserterMoves(flags, opponent(playerix))
			case cards.TC_Redeploy:
				moves = getRedeployMoves(flags, playerix)
			case cards.TC_Scout:
				moves = getScoutMoves(tacDeck, troopDeck)
			case cards.TC_Traitor:
				moves = getTraitorMoves(flags, playerix)
			}
			if len(moves) != 0 {
				m[v] = moves
			}
		}
	}
	if len(troopSpace) > 0 && len(hand.Troops) > 0 {
		for _, troop := range hand.Troops {
			moves = getCardFlagMoves(troopSpace)
			if len(moves) != 0 {
				m[troop] = moves
			}
		}
	} else {

		if len(m) != 0 {
			pass = true
		} else {
			m = nil
		}
	}

	return m, pass
}

// getScoutMoves returns the possible scout moves.
func getScoutMoves(tac *deck.Deck, troop *deck.Deck) (moves []Move) {
	moves = make([]Move, 0, 2)
	if !tac.Empty() {
		moves = append(moves, *NewMoveDeck(DECK_TAC))
	}
	if !troop.Empty() {
		moves = append(moves, *NewMoveDeck(DECK_TROOP))
	}
	return moves
}

// getTraitorMoves returns the possible traiter moves.
func getTraitorMoves(flags *[FLAGS]*flag.Flag, playerix int) (moves []Move) {
	moves = make([]Move, 0, (FLAGS*3+3)*FLAGS) //270
	for oppFlagix, oppFlag := range flags {
		if !oppFlag.Claimed() {
			for flagix, flag := range flags {
				if !flag.Claimed() && flag.Free()[playerix] {
					for _, troop := range oppFlag.Troops(opponent(playerix)) {
						if troop <= cards.TROOP_NO {
							moves = append(moves, *NewMoveTraitor(oppFlagix, troop, flagix))
						}
					}
				}
			}
		}
	}
	return moves
}

// getRedeployMoves returns the possible redeploy moves.
func getRedeployMoves(flags *[FLAGS]*flag.Flag, playerix int) (moves []Move) {
	moves = make([]Move, 0, (FLAGS*3+3)*(FLAGS+1)) //300
	for outFlagix, outFlag := range flags {
		if !outFlag.Claimed() {
			for inFlagix, inFlag := range flags {
				if !inFlag.Claimed() && inFlag.Free()[playerix] && outFlagix != inFlagix {
					for _, troop := range outFlag.Troops(playerix) {
						moves = append(moves, *NewMoveRedeploy(outFlagix, troop, inFlagix))
					}
					for _, tac := range outFlag.Env(playerix) {
						moves = append(moves, *NewMoveRedeploy(outFlagix, tac, inFlagix))
					}
				}
			}
			for _, troop := range outFlag.Troops(playerix) {
				moves = append(moves, *NewMoveRedeploy(outFlagix, troop, -1))
			}
			for _, tac := range outFlag.Env(playerix) {
				moves = append(moves, *NewMoveRedeploy(outFlagix, tac, -1))
			}
		}
	}
	return moves
}

// getDeserterMoves retunrs the possible deserter moves.
func getDeserterMoves(flags *[FLAGS]*flag.Flag, opp int) (moves []Move) {
	moves = make([]Move, 0, FLAGS*3+3)
	for flagix, flag := range flags {
		if !flag.Claimed() {
			for _, troop := range flag.Troops(opp) {
				moves = append(moves, *NewMoveDeserter(flagix, troop))
			}
			for _, tac := range flag.Env(opp) {
				moves = append(moves, *NewMoveDeserter(flagix, tac))
			}
		}
	}
	return moves
}

// getCardFlagMoves create all the card to flag moves.
// flags with space.
func getCardFlagMoves(flags []int) (moves []Move) {
	moves = make([]Move, len(flags))
	for i, v := range flags {
		moves[i] = *NewMoveCardFlag(v)
	}
	return moves
}

// setMoveScout returns the scout moves.
// The numbers of moves should be len(tac)=nta and len(troop)=nto
// (nta*nta-1) + nto*nto-1+nta*nto if there is enough cards and
// there is two cards to return.
func getMoveScoutReturn(hand *Hand) (m []Move) {
	nta := len(hand.Tacs)
	nto := len(hand.Troops)
	ret := nta + nto - HAND
	if ret == 2 {
		m = make([]Move, 0, 72)
	} else {
		m = make([]Move, 0, 18)
	}
	if ret == 2 {
		if nta > 1 {
			for i, vi := range hand.Tacs {
				for j, vj := range hand.Tacs {
					if i != j {
						m = append(m, *NewMoveScoutReturn([]int{vi, vj}, nil))
					}
				}
			}
		}
		if nto > 1 {
			for i, vi := range hand.Troops {
				for j, vj := range hand.Troops {
					if i != j {
						m = append(m, *NewMoveScoutReturn(nil, []int{vi, vj}))
					}
				}
			}
		}
		if nta > 0 && nto > 0 {
			for _, tac := range hand.Tacs {
				for _, troop := range hand.Troops {
					m = append(m, *NewMoveScoutReturn([]int{tac}, []int{troop}))
				}
			}
		}
	}
	if ret == 1 {
		if nta > 0 {
			for _, tac := range hand.Tacs {
				m = append(m, *NewMoveScoutReturn([]int{tac}, nil))
			}
		}
		if nto > 0 {
			for _, troop := range hand.Troops {
				m = append(m, *NewMoveScoutReturn(nil, []int{troop}))
			}
		}
	}

	return m
}

// getMoveClaim returns all the possible claim flag moves.
// There is no validation, that is it is not checked if a claim will
// succede only that it is possible to make.
func getMoveClaim(playerix int, flags *[FLAGS]*flag.Flag) (m []Move) {
	posFlags := make([]int, 0, FLAGS)
	for i, flag := range flags {
		if !flag.Claimed() {
			if flag.Formations()[playerix] != nil {
				posFlags = append(posFlags, i)
			}
		}
	}
	m = claimCombi(posFlags)
	return m
}
func claimCombi(posFlags []int) (m []Move) {
	n := len(posFlags)
	if n != 0 {
		m = make([]Move, 0, claimCombiNo(n))
		m = append(m, *NewMoveClaim(make([]int, 0))) //no claims
		m = append(m, *NewMoveClaim(posFlags[:]))    // all
		switch n {
		case 2:
			m = claimAddOne(m, false, posFlags)
		case 3:
			m = claimAddOne(m, true, posFlags)
		case 4:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, false, posFlags, 2)
		case 5:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, true, posFlags, 2)
		case 6:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, true, posFlags, 2)
			m = claimAdd(m, false, posFlags, 3)
		case 7:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, true, posFlags, 2)
			m = claimAdd(m, true, posFlags, 3)
		case 8:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, true, posFlags, 2)
			m = claimAdd(m, true, posFlags, 3)
			m = claimAdd(m, false, posFlags, 4)
		case 9:
			m = claimAddOne(m, true, posFlags)
			m = claimAdd(m, true, posFlags, 2)
			m = claimAdd(m, true, posFlags, 3)
			m = claimAdd(m, true, posFlags, 4)
		}
	}
	return m
}

func claimAddOne(m []Move, reverse bool, posFlags []int) []Move {
	for i := range posFlags {
		m = append(m, *NewMoveClaim(posFlags[i : i+1]))
		if reverse {
			m = append(m, *NewMoveClaim(slice.WithOutNew(posFlags, posFlags[i:i+1])))
		}
	}
	return m
}

func claimAdd(m []Move, reverse bool, posFlags []int, d int) []Move {
	n := len(posFlags)
	_ = math.Perm(n, d, func(per []int) bool {
		var flagixs = make([]int, d)
		for i, ix := range per {
			flagixs[i] = posFlags[ix]
		}
		m = append(m, *NewMoveClaim(flagixs))
		if reverse {
			m = append(m, *NewMoveClaim(slice.WithOutNew(posFlags, flagixs)))
		}
		return false
	})
	return m
}

func claimCombiNo(flagsNo int) (no int) {
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

// getMoveDeck returns all the possible move deck.
func getMoveDeck(tacDeck *deck.Deck, troopDeck *deck.Deck, hand *Hand, turnState int) (m []Move) {
	m = make([]Move, 0, 2)
	if turnState != TURN_DECK || (turnState == TURN_DECK && hand.Size() < 7) {
		if !tacDeck.Empty() {
			m = append(m, *NewMoveDeck(DECK_TAC))
		}
		if !troopDeck.Empty() {
			m = append(m, *NewMoveDeck(DECK_TROOP))
		}
	}
	return m
}

// Move a interface for moves.
type Move interface {
	MoveEqual(Move) bool
	Copy() Move
}

// MoveCardFlag the place a card on a flag move.
// Its is just int for the flag index.
type MoveCardFlag struct {
	Flagix   int
	JsonType string
}

func NewMoveCardFlag(flagix int) *MoveCardFlag {
	res := new(MoveCardFlag)
	res.Flagix = flagix
	res.JsonType = "MoveCardFlag"
	return res
}
func (m MoveCardFlag) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveCardFlag)
		if ok && om == m {
			equal = true
		}
	}
	return equal
}
func (m MoveCardFlag) Copy() (c Move) {
	c = m
	return c
}

// MoveDeserter the deserter move. The flag and the index of the card to kill.
type MoveDeserter struct {
	Flag     int
	Card     int
	JsonType string
}

func NewMoveDeserter(flagix int, cardix int) *MoveDeserter {
	res := new(MoveDeserter)
	res.Flag = flagix
	res.Card = cardix
	res.JsonType = "MoveDeserter"
	return res
}

func (m MoveDeserter) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveDeserter)
		if ok && om == m {
			equal = true
		}
	}
	return equal
}
func (m MoveDeserter) Copy() (c Move) {
	c = m
	return c
}

// MoveTraitor the traitor move, the flag and card index of the card to move
// and destination flag.
type MoveTraitor struct {
	OutFlag  int
	OutCard  int
	InFlag   int
	JsonType string
}

func NewMoveTraitor(outFlagix int, outCardix int, inFlagix int) *MoveTraitor {
	res := new(MoveTraitor)
	res.OutFlag = outFlagix
	res.OutCard = outCardix
	res.InFlag = inFlagix
	res.JsonType = "MoveTraitor"
	return res
}

func (m MoveTraitor) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveTraitor)
		if ok && om == m {
			equal = true
		}
	}
	return equal
}
func (m MoveTraitor) Copy() (c Move) {
	c = m
	return c
}

// MoveRedeploy the redeploy move, the flag and card index of the card to move and the
// destination flag.
type MoveRedeploy struct {
	OutFlag  int
	OutCard  int
	InFlag   int //may be -1 no flag goes to dish.
	JsonType string
}

func NewMoveRedeploy(outFlagix int, outCardix int, inFlagix int) *MoveRedeploy {
	res := new(MoveRedeploy)
	res.OutFlag = outFlagix
	res.OutCard = outCardix
	res.InFlag = inFlagix
	res.JsonType = "MoveRedeploy"
	return res
}
func (m MoveRedeploy) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveRedeploy)
		if ok && om == m {
			equal = true
		}
	}
	return equal
}
func (m MoveRedeploy) Copy() (c Move) {
	c = m
	return c
}

// MoveScoutReturn the scout return move. The tactic cards and the troop cards.
// It is first in last out. The first card of the slice will be delt last.
type MoveScoutReturn struct {
	Tac      []int
	Troop    []int
	JsonType string
}

func NewMoveScoutReturn(tac []int, troop []int) *MoveScoutReturn {
	res := new(MoveScoutReturn)
	res.Tac = tac
	res.Troop = troop
	res.JsonType = "MoveScoutReturn"
	return res
}

func (m MoveScoutReturn) Equal(other MoveScoutReturn) (equal bool) {
	if slice.Equal(other.Tac, m.Tac) {
		if slice.Equal(other.Troop, m.Troop) {
			equal = true
		}
	}
	return equal
}
func (m MoveScoutReturn) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveScoutReturn)
		if ok {
			equal = m.Equal(om)
		}
	}
	return equal
}
func (m MoveScoutReturn) Copy() Move {
	scout := *new(MoveScoutReturn)
	if m.Tac != nil {
		scout.Tac = make([]int, len(m.Tac))
		copy(scout.Tac, m.Tac)
	}
	if m.Troop != nil {
		scout.Troop = make([]int, len(m.Troop))
		copy(scout.Troop, m.Troop)
	}
	return scout
}

// MoveDeck the deck move. DECK_TAC or DECK_TROOP
type MoveDeck struct {
	Deck     int
	JsonType string
}

func NewMoveDeck(deck int) *MoveDeck {
	res := new(MoveDeck)
	res.Deck = deck
	res.JsonType = "MoveDeck"
	return res
}

func (m MoveDeck) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveDeck)
		if ok && om == m {
			equal = true
		}
	}
	return equal
}
func (m MoveDeck) Copy() (c Move) {
	c = m
	return c
}

// MoveClaim the claim flags move. The slice contain the list of flags
// to claim.
type MoveClaim struct {
	Flags    []int
	JsonType string
}

func NewMoveClaim(flags []int) *MoveClaim {
	res := new(MoveClaim)
	res.Flags = flags
	res.JsonType = "MoveClaim"
	return res
}
func (m MoveClaim) Equal(other MoveClaim) (equal bool) {
	if slice.Equal(other.Flags, m.Flags) {
		equal = true
	}
	return equal
}
func (m MoveClaim) MoveEqual(other Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveClaim)
		if ok {
			equal = m.Equal(om)
		}
	}
	return equal
}
func (m MoveClaim) Copy() (c Move) {
	var v []int
	if m.Flags != nil {
		v = make([]int, len(m.Flags))
		copy(v, m.Flags)
	}

	c = *NewMoveClaim(v)
	return c
}
