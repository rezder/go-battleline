package battleline

import (
	"encoding/gob"
	"fmt"
	"os"
	"rezder.com/game/card/battleline/cards"
	"rezder.com/game/card/battleline/flag"
	"rezder.com/game/card/deck"
)

const (
	FLAGS = 9
	HAND  = 7
)

type Game struct {
	Id        int
	PlayerIds [2]int
	Pos       GamePos
	PosInit   GamePos
	Moves     [][2]int
}

func New(id int, playerIds [2]int) (game *Game) {
	game = new(Game)
	game.Id = id
	game.PlayerIds = playerIds
	game.Pos = *NewGamePos()

	return game
}
func (game *Game) Equal(other *Game) (equal bool) {
	if other == nil && game == nil {
		equal = true
	} else if other != nil && game != nil {
		if game == other {
			equal = true
		} else if other.Id == game.Id && other.PlayerIds == game.PlayerIds &&
			other.PosInit.Equal(&game.PosInit) && other.Pos.Equal(&game.Pos) {
			equal = true
			for i, v := range other.Moves {
				if v != game.Moves[i] {
					equal = false
					break
				}
			}
		}
	}
	return equal
}
func (game *Game) Start(starter int) {
	pos := &game.Pos
	deal(&pos.Hands, &pos.DeckTroop)
	pos.Turn.start(starter, pos.Hands[starter], &pos.Flags, &pos.DeckTac, &pos.DeckTroop, &pos.Dishs)
	game.PosInit = *game.Pos.Copy()
	game.Moves = make([][2]int, 0)
}

func (game *Game) addMove(cardix int, moveix int) {
	game.Moves = append(game.Moves, [2]int{cardix, moveix})
}

func (game *Game) Quit(playerix int) {
	game.Pos.quit()
	game.Pos.info = ""
	game.addMove(-1, -1)
}
func (game *Game) Pass() {
	if game.Pos.MovePass {
		game.Pos.info = ""
		game.addMove(-1, -1)
		game.Pos.next(false, &game.Pos.Hands, &game.Pos.Flags, &game.Pos.DeckTac, &game.Pos.DeckTroop, &game.Pos.Dishs)
	} else {
		panic("Calling pass when not possible")
	}
}
func (game *Game) Move(move int) {
	game.addMove(-1, move)
	pos := &game.Pos //Update
	pos.info = ""
	switch pos.State {

	case TURN_FLAG:
		moveC, ok := pos.Moves[move].(MoveClaim)
		if ok {
			pos.info = moveClaimFlag(pos.Player, moveC, &pos.Flags, &pos.Hands, &pos.DeckTroop)
		} else {
			panic("There should not be only claim moves")
		}
	case TURN_SCOUT1, TURN_SCOUT2, TURN_DECK:
		moveD, ok := pos.Moves[move].(MoveDeck)
		if ok {
			moveDeck(moveD, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
		} else {
			panic("There should not be only scout pick deck moves ")
		}
	case TURN_SCOUTR:
		moveSctR, ok := pos.Moves[move].(MoveScoutReturn)
		if ok {
			moveScoutRet(&moveSctR, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
		} else {
			panic("There should not be only scout return deck moves ")
		}
	case TURN_HAND:
		panic(" There should be now hand move here, pass hand move is not handle her")
	case TURN_FINISH:
		panic("Calling move when the game is finish is point less")
	default:
		panic("Unexpected turn state")
	}
	pos.next(false, &pos.Hands, &pos.Flags, &pos.DeckTac, &pos.DeckTroop, &pos.Dishs)

}

//moveScoutReturn make the return from scout.
//#deckTack
//#deckTroop
//#hand
func moveScoutRet(move *MoveScoutReturn, deckTack *deck.Deck, deckTroop *deck.Deck, hand *Hand) {
	if len(move.Tac) != 0 {
		reTac := make([]int, len(move.Tac))
		copy(reTac, move.Tac) // this may no be necessary
		deckTack.Return(reTac)
		hand.playMulti(move.Tac)
	}
	if len(move.Troop) != 0 {
		re := make([]int, len(move.Troop))
		copy(re, move.Troop)
		deckTroop.Return(re)
		hand.playMulti(move.Troop)
	}
}

//moveClaimFlag make a claim flag move.
//claims is the flag indexs that should be claimed if possible.
//#flags
func moveClaimFlag(playerix int, claimixs []int, flags *[FLAGS]*flag.Flag, hands *[2]*Hand,
	deckTroop *deck.Deck) (info string) {
	unPlayCards := simTroops(deckTroop, hands[0].Troops, hands[1].Troops)
	for _, claim := range claimixs {
		ok, ex := flags[claim].ClaimFlag(playerix, unPlayCards)
		if !ok {
			info = fmt.Sprintf("Claiming flag %v failed. Example: %v", claim, ex) // TODO make  with card information nicer
		}
	}
	return info
}

//moveDeck make select deck move.
//#tacDeck
//#troopDeck
//#hand
func moveDeck(deck MoveDeck, tacDeck *deck.Deck, troopDeck *deck.Deck, hand *Hand) {
	switch int(deck) {
	case DECK_TAC:
		hand.draw(deckDealTactic(tacDeck))
	case DECK_TROOP:
		hand.draw(deckDealTroop(troopDeck))
	}
}

func (game *Game) MoveHand(cardix int, moveix int) {
	game.addMove(cardix, moveix)
	pos := &game.Pos //Update
	pos.info = ""
	scout := false
	var err error
	if pos.State == TURN_HAND {
		pos.Hands[pos.Player].play(cardix)
		switch move := pos.MovesHand[cardix][moveix].(type) {
		case MoveCardFlag:
			err = pos.Flags[int(move)].Set(cardix, pos.Player)
		case MoveDeck: //scout
			moveDeck(move, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
			pos.Dishs[pos.Player].dishCard(cardix)
			scout = true
		case MoveDeserter:
			err = moveDeserter(&move, &pos.Flags, pos.Opp(), &pos.Dishs)
			pos.Dishs[pos.Player].dishCard(cardix)
		case MoveTraitor:
			err = moveTraitor(&move, &pos.Flags, pos.Player)
			pos.Dishs[pos.Player].dishCard(cardix)
		case MoveRedeploy:
			err = moveRedeploy(&move, &pos.Flags, pos.Player, &pos.Dishs)
			pos.Dishs[pos.Player].dishCard(cardix)
		default:
			panic("Illegal move type")
		}
		if err != nil {
			panic("This should not be possible only valid move should exist")
		} else {
			pos.next(scout, &pos.Hands, &pos.Flags, &pos.DeckTac, &pos.DeckTroop, &pos.Dishs)
		}
	} else {
		panic("Wrong move function turn is not play card")
	}
}

func moveRedeploy(move *MoveRedeploy, flags *[FLAGS]*flag.Flag, playerix int, dishs *[2]*Dish) (err error) {
	var outFlag *flag.Flag = flags[move.OutFlag]
	var inFlag *flag.Flag = flags[move.InFlag]
	m0ix, m1ix, err := outFlag.Remove(move.OutCard, playerix)
	if err == nil {
		if m0ix != -1 {
			dishs[0].dishCard(m0ix)
		}
		if m1ix != -1 {
			dishs[1].dishCard(m1ix)
		}
		if move.OutCard != -1 {
			err = inFlag.Set(move.OutCard, playerix)
		} else {
			dishs[playerix].dishCard(move.OutCard)
		}
	}
	return err
}
func moveTraitor(move *MoveTraitor, flags *[FLAGS]*flag.Flag, playerix int) (err error) {
	var outFlag *flag.Flag = flags[move.OutFlag]
	var inFlag *flag.Flag = flags[move.InFlag]
	_, _, err = outFlag.Remove(move.OutCard, opponent(playerix)) //Only troop can be a traitor so no mudix
	if err == nil {
		err = inFlag.Set(move.OutCard, playerix)
	}
	return err
}
func moveDeserter(move *MoveDeserter, flags *[FLAGS]*flag.Flag, oppix int, dishs *[2]*Dish) (err error) {
	var flag *flag.Flag = flags[move.Flag]
	m0ix, m1ix, err := flag.Remove(move.Card, oppix)
	if err == nil {
		dishs[oppix].dishCard(move.Card)
		if m0ix != -1 {
			dishs[0].dishCard(m0ix)
		}
		if m1ix != -1 {
			dishs[1].dishCard(m1ix)
		}
	}
	return err
}

type GamePos struct { //TODO make deep copy all map slice, interface or pointers need special treatment
	Flags     [FLAGS]*flag.Flag
	Dishs     [2]*Dish
	Hands     [2]*Hand
	DeckTac   deck.Deck
	DeckTroop deck.Deck
	Turn
	info string
}

func NewGamePos() *GamePos {
	pos := new(GamePos)
	for i := range pos.Flags {
		pos.Flags[i] = flag.New()
	}
	pos.DeckTac = *deck.New(cards.TAC_NO)
	pos.DeckTroop = *deck.New(cards.TROOP_NO)
	pos.Hands[0] = NewHand()
	pos.Hands[1] = NewHand()
	pos.Dishs[0] = NewDish()
	pos.Dishs[1] = NewDish()

	return pos
}
func (pos *GamePos) Equal(other *GamePos) (equal bool) {
	if other == nil && pos == nil {
		equal = true
	} else if other != nil && pos != nil {
		if pos == other {
			equal = true
		} else if other.Turn.Equal(&pos.Turn) && other.DeckTac.Equal(&pos.DeckTac) &&
			other.DeckTroop.Equal(&pos.DeckTroop) && other.info == pos.info {
			equalList := true
			for i, v := range other.Flags {
				if !v.Equal(pos.Flags[i]) {
					equalList = false
					break
				}
			}
			if equalList {
				for i, v := range other.Hands {
					if !v.Equal(pos.Hands[i]) {
						equalList = false
						break
					}
				}
				if equalList {
					for i, v := range other.Dishs {
						if !v.Equal(pos.Dishs[i]) {
							equalList = false
							break
						}
					}
				}
			}
			if equalList {
				equal = true
			}
		}
	}
	return equal
}
func (pos *GamePos) Copy() (c *GamePos) {
	c = new(GamePos)
	for i := range pos.Flags {
		c.Flags[i] = pos.Flags[i].Copy()
	}
	c.Flags = pos.Flags
	c.Dishs[0] = pos.Dishs[0].Copy()
	c.Dishs[1] = pos.Dishs[1].Copy()
	c.Hands[0] = pos.Hands[0].Copy()
	c.Hands[1] = pos.Hands[1].Copy()
	c.DeckTac = *pos.DeckTac.Copy()
	c.DeckTroop = *pos.DeckTroop.Copy()
	c.Turn = *pos.Turn.Copy()
	c.info = pos.info
	return c
}

func simTroops(deck *deck.Deck, troops1 []int, troops2 []int) (troops []int) {
	dr := deck.Remaining()
	troops = make([]int, len(dr), len(dr)+(2*HAND))
	copy(troops, dr)
	for i := 0; i < HAND; i++ {
		c, _ := cards.DrTroop(troops1[i])
		if c != nil {
			troops = append(troops, troops1[i])
		}
		c, _ = cards.DrTroop(troops2[i])
		if c != nil {
			troops = append(troops, troops2[i])
		}
	}
	return troops
}
func opponent(playerix int) int {
	if playerix == 0 {
		return 1
	} else {
		return 0
	}
}

// deal deals the initial hands.
//#players.
//#deck.
func deal(hands *[2]*Hand, deck *deck.Deck) {
	for _, hand := range hands {
		for i := 0; i < HAND; i++ {
			hand.draw(deckDealTroop(deck))
		}
	}
}

// deckDealTroop deals a troop from the deck
// #deck
func deckDealTroop(deckTroop *deck.Deck) (troop int) {
	c, err := deckTroop.Deal()
	if err != nil {
		panic("You should not deal a card if deck is empty")
	}
	return deckToTroop(c)
}

//deckDealTactic deals a tactic card from deck
//#deck
func deckDealTactic(deckTac *deck.Deck) (tac int) {
	c, err := deckTac.Deal()
	if err != nil {
		panic("You should not deal a card if deck is empty")
	}
	return deckToTactic(c)
}
func deckToTroop(deckix int) int {
	return deckix + 1
}
func deckToTactic(deckix int) int {
	return deckix + 1 + cards.TROOP_NO
}
func GobRegistor() {
	gob.Register(MoveCardFlag(0))
	gob.Register(MoveClaim([]int{}))
	gob.Register(MoveDeck(0))
	gob.Register(MoveDeserter{})
	gob.Register(MoveRedeploy{})
	gob.Register(MoveScoutReturn{})
	gob.Register(MoveTraitor{})
}
func Save(game *Game, file *os.File) (err error) {
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(game)
	return err
}
func Load(file *os.File) (game *Game, err error) {
	gob.Register(MoveCardFlag(0))
	decoder := gob.NewDecoder(file)
	var g Game = *New(1, [2]int{1, 2})

	err = decoder.Decode(&g)
	game = &g
	return game, err
}
