//Package battleline contains the card game battleline.
package battleline

import (
	"encoding/gob"
	"github.com/rezder/go-battleline/battleline/cards"
	"github.com/rezder/go-battleline/battleline/flag"
	"github.com/rezder/go-card/deck"
	"os"
)

const (
	//NOFlags is the numbers of flags
	NOFlags = 9
	//NOHandInit number of cards in init hand.
	NOHandInit = 7
	//SMPass is the move pass
	SMPass = -1
	//SMGiveUp is the move give up
	SMGiveUp = -2
)

//Game The struct that holds all information about a game
type Game struct {
	PlayerIds     [2]int
	Pos           *GamePos
	InitDeckTac   deck.Deck
	InitDeckTroop deck.Deck
	Starter       int
	Moves         [][2]int
}

func New(playerIds [2]int) (game *Game) {
	game = new(Game)
	game.PlayerIds = playerIds
	//game.Pos = NewGamePos()
	return game
}
func (game *Game) Equal(other *Game) (equal bool) {
	if other == nil && game == nil {
		equal = true
	} else if other != nil && game != nil {
		if game == other {
			equal = true
		} else if other.PlayerIds == game.PlayerIds &&
			other.Starter == game.Starter &&
			other.InitDeckTroop.Equal(&game.InitDeckTroop) &&
			other.InitDeckTac.Equal(&game.InitDeckTac) &&
			other.Pos.Equal(game.Pos) {
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

// ResetGame reset the game to the first position and returns
// and clear the moves.
func (game *Game) ResetGame() (moves [][2]int) {
	pos := NewGamePos()
	pos.DeckTroop = *game.InitDeckTroop.Copy()
	pos.DeckTac = *game.InitDeckTac.Copy()
	deal(&pos.Hands, &pos.DeckTroop)
	pos.Turn.start(game.Starter, pos.Hands[game.Starter], &pos.Flags,
		&pos.DeckTac, &pos.DeckTroop, &pos.Dishs)
	game.Pos = pos

	moves = make([][2]int, len(game.Moves))
	copy(moves, game.Moves)
	game.Moves = make([][2]int, 0, len(moves))
	return moves
}

//histMove update game with stored move.
//#game
func histMove(moveixs [2]int, game *Game) (move Move, moveCardix, dealtix int, isPass, isGiveUp bool, claimFailExs [9][]int, mudDishixs []int) {
	switch {
	case moveixs[1] == SMGiveUp:
		game.Quit(game.Pos.Player)
		isGiveUp = true
	case moveixs[1] == SMPass:
		game.Pass()
		isPass = true
	case moveixs[1] >= 0:
		if moveixs[0] > 0 {
			move = game.Pos.MovesHand[moveixs[0]][moveixs[1]]
			dealtix, mudDishixs = game.MoveHand(moveixs[0], moveixs[1])
			moveCardix = moveixs[0]
		} else {
			move = game.Pos.Moves[moveixs[1]]
			dealtix, claimFailExs = game.Move(moveixs[1])
		}
	default:
		panic("This should not happen. Move data is corrupt")
	}
	return move, moveCardix, dealtix, isPass, isGiveUp, claimFailExs, mudDishixs
}

//CalcPos calculate the current posistion from the initial position and
//the moves. The new position replace the old position.
func (game *Game) CalcPos() {
	moves := game.ResetGame()
	for _, move := range moves {
		histMove(move, game)
	}
}
func (game *Game) Start(starter int) {
	pos := NewGamePos()
	game.Starter = starter
	game.InitDeckTroop = *pos.DeckTroop.Copy()
	game.InitDeckTac = *pos.DeckTac.Copy()
	deal(&pos.Hands, &pos.DeckTroop)
	pos.Turn.start(starter, pos.Hands[starter], &pos.Flags, &pos.DeckTac, &pos.DeckTroop, &pos.Dishs)
	game.Pos = pos
	game.Moves = make([][2]int, 0)
}

func (game *Game) addMove(cardix int, moveix int) {
	game.Moves = append(game.Moves, [2]int{cardix, moveix})
}

//Quit handle player giving up.
func (game *Game) Quit(playerix int) {
	game.Pos.quit()
	game.Pos.Info = ""
	game.addMove(0, SMGiveUp)
}

//Pass player choose not to make a move.
func (game *Game) Pass() {
	if game.Pos.MovePass {
		game.Pos.Info = ""
		game.addMove(0, SMPass)
		game.Pos.next(false, &game.Pos.Hands, &game.Pos.Flags, &game.Pos.DeckTac, &game.Pos.DeckTroop, &game.Pos.Dishs)
	} else {
		panic("Calling pass when not possible")
	}
}

//Move makes a none card move. Claim flags, Getting or returning cards to deck.
//dealtix the card in deal move.
//claimFailMap the failed claim map in a claim flag move. Is never nil.
func (game *Game) Move(move int) (dealtix int, claimsFailExs [9][]int) {
	game.addMove(0, move)
	pos := game.Pos //Update
	pos.Info = ""
	switch pos.State {

	case TURNFlag:
		moveC, ok := pos.Moves[move].(MoveClaim)
		if ok {
			claimsFailExs = moveClaimFlag(pos.Player, moveC.Flags, &pos.Flags, &pos.Hands, &pos.DeckTroop)
		} else {
			panic("There should be only claim moves")
		}
	case TURNScout1, TURNScout2, TURNDeck:
		moveD, ok := pos.Moves[move].(MoveDeck)
		if ok {
			dealtix = moveDeck(moveD, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
		} else {
			panic("There should be only pick deck moves ")
		}
	case TURNScoutR:
		moveSctR, ok := pos.Moves[move].(MoveScoutReturn)
		if ok {
			moveScoutRet(&moveSctR, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
		} else {
			panic("There should not only scout return deck moves ")
		}
	case TURNHand:
		panic(" There should be now hand move here, pass hand move is not handle her")
	case TURNFinish:
		panic("Calling move when the game is finish is point less")
	default:
		panic("Unexpected turn state")
	}
	pos.next(false, &pos.Hands, &pos.Flags, &pos.DeckTac, &pos.DeckTroop, &pos.Dishs)
	return dealtix, claimsFailExs
}

//moveScoutReturn make the return from scout.
//#deckTack
//#deckTroop
//#hand
func moveScoutRet(move *MoveScoutReturn, deckTack *deck.Deck, deckTroop *deck.Deck, hand *Hand) {
	if len(move.Tac) != 0 {
		reTac := make([]int, len(move.Tac))
		for i, v := range move.Tac {
			reTac[i] = deckFromTactic(v)
		}
		deckTack.Return(reTac)
		hand.PlayMulti(move.Tac)
	}
	if len(move.Troop) != 0 {
		reTroop := make([]int, len(move.Troop))
		for i, v := range move.Troop {
			reTroop[i] = deckFromTroop(v)
		}
		deckTroop.Return(reTroop)
		hand.PlayMulti(move.Troop)
	}
}

//moveClaimFlag make a claim flag move.
//claimixs is the flag indexs that should be claimed if possible.
//claimsFailExs jsoned so nil is change to empty array empty and nil is not different, we must always have a example if
//claim failed.
//#flags
func moveClaimFlag(playerix int, claimixs []int, flags *[NOFlags]*flag.Flag, hands *[2]*Hand,
	deckTroop *deck.Deck) (claimsFailExs [9][]int) {
	unPlayCards := simTroops(deckTroop, hands[0].Troops, hands[1].Troops)
	for _, claim := range claimixs {
		ok, ex := flags[claim].ClaimFlag(playerix, unPlayCards) //ex contain 0
		if !ok {
			claimsFailExs[claim] = ex
		}
	}
	return claimsFailExs
}

//moveDeck make select deck move.
//#tacDeck
//#troopDeck
//#hand
func moveDeck(deck MoveDeck, tacDeck *deck.Deck, troopDeck *deck.Deck, hand *Hand) (dealt int) {
	switch deck.Deck {
	case DECKTac:
		dealt = deckDealTactic(tacDeck)
		hand.Draw(dealt)
	case DECKTroop:
		dealt = deckDealTroop(troopDeck)
		hand.Draw(dealt)
	}
	return dealt
}

//MoveHand play a card from hand.
//dealtix the delt cardix when the scout card is played.
//dishixs the dished cards witch may results of a redeploy or desert of the mud card.
//in case of redeploy it also holds the redeploy card if it is dished.
func (game *Game) MoveHand(cardix int, moveix int) (dealtix int, dishixs []int) {
	game.addMove(cardix, moveix)
	pos := game.Pos //Update
	pos.Info = ""
	scout := false
	var err error
	if pos.State == TURNHand {
		pos.Hands[pos.Player].Play(cardix)
		switch move := pos.MovesHand[cardix][moveix].(type) {
		case MoveCardFlag:
			err = pos.Flags[move.Flagix].Set(cardix, pos.Player)
		case MoveDeck: //scout
			dealtix = moveDeck(move, &pos.DeckTac, &pos.DeckTroop, pos.Hands[pos.Player])
			pos.Dishs[pos.Player].DishCard(cardix)
			scout = true
		case MoveDeserter:
			dishixs, err = moveDeserter(&move, &pos.Flags, pos.Opp(), &pos.Dishs)
			pos.Dishs[pos.Player].DishCard(cardix)
		case MoveTraitor:
			err = moveTraitor(&move, &pos.Flags, pos.Player)
			pos.Dishs[pos.Player].DishCard(cardix)
		case MoveRedeploy:
			dishixs, err = moveRedeploy(&move, &pos.Flags, pos.Player, &pos.Dishs)
			pos.Dishs[pos.Player].DishCard(cardix)
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
	return dealtix, dishixs
}

//moveRedeploy handle the reploy move.
//In case of a redeploying
//the mud card two extra dish cards is possible.
//#flags
//#dish
func moveRedeploy(move *MoveRedeploy, flags *[NOFlags]*flag.Flag, playerix int,
	dishs *[2]*Dish) (dishixs []int, err error) {
	outFlag := flags[move.OutFlag]
	m0ix, m1ix, err := outFlag.Remove(move.OutCard, playerix)
	if err != nil {
		return dishixs, err
	}
	if m0ix != -1 {
		dishs[0].DishCard(m0ix)
		dishixs = append(dishixs, m0ix)
	}
	if m1ix != -1 {
		dishs[1].DishCard(m1ix)
		dishixs = append(dishixs, m1ix)
	}
	if move.InFlag != -1 {
		inFlag := flags[move.InFlag]
		err = inFlag.Set(move.OutCard, playerix)
	} else {
		dishs[playerix].DishCard(move.OutCard)
	}

	return dishixs, err
}

//moveTraitor handle the traitor move.
//#flags
func moveTraitor(move *MoveTraitor, flags *[NOFlags]*flag.Flag, playerix int) (err error) {
	outFlag := flags[move.OutFlag]
	_, _, err = outFlag.Remove(move.OutCard, opponent(playerix)) //Only troop can be a traitor so no mudix
	if err != nil {
		return err
	}
	inFlag := flags[move.InFlag]
	err = inFlag.Set(move.OutCard, playerix)
	return err
}

//moveDeserter handle the deserter move.
//#flags
//#dishs
func moveDeserter(move *MoveDeserter, flags *[NOFlags]*flag.Flag, oppix int,
	dishs *[2]*Dish) (dishixs []int, err error) {
	flag := flags[move.Flag]
	m0ix, m1ix, err := flag.Remove(move.Card, oppix)
	if err != nil {
		return dishixs, err
	}
	dishs[oppix].DishCard(move.Card)
	if m0ix != -1 {
		dishs[0].DishCard(m0ix)
		dishixs = append(dishixs, m0ix)
	}
	if m1ix != -1 {
		dishs[1].DishCard(m1ix)
		dishixs = append(dishixs, m1ix)
	}
	return dishixs, err
}

//GamePos a game position.
type GamePos struct {
	Flags     [NOFlags]*flag.Flag
	Dishs     [2]*Dish
	Hands     [2]*Hand
	DeckTac   deck.Deck
	DeckTroop deck.Deck
	Turn
	Info string
}

func NewGamePos() *GamePos {
	pos := new(GamePos)
	for i := range pos.Flags {
		pos.Flags[i] = flag.New()
	}
	pos.DeckTac = *deck.New(cards.NOTac)
	pos.DeckTroop = *deck.New(cards.NOTroop)
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
			other.DeckTroop.Equal(&pos.DeckTroop) && other.Info == pos.Info {
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
	c.Dishs[0] = pos.Dishs[0].Copy()
	c.Dishs[1] = pos.Dishs[1].Copy()
	c.Hands[0] = pos.Hands[0].Copy()
	c.Hands[1] = pos.Hands[1].Copy()
	c.DeckTac = *pos.DeckTac.Copy()
	c.DeckTroop = *pos.DeckTroop.Copy()
	c.Turn = *pos.Turn.Copy()
	c.Info = pos.Info
	return c
}

//simTroops return the troops need for simulation.
func simTroops(deck *deck.Deck, troops1 []int, troops2 []int) (troops []int) {
	dr := deck.Remaining()
	troops = make([]int, len(dr), len(dr)+len(troops1)+len(troops2))
	if len(dr) > 0 {
		for i, v := range dr {
			troops[i] = deckToTroop(v)
		}
	}
	if len(troops1) > 0 {
		troops = append(troops, troops1...)
	}
	if len(troops2) > 0 {
		troops = append(troops, troops2...)
	}
	return troops
}

//opponent return the opponent.
func opponent(playerix int) int {
	if playerix == 0 {
		return 1
	}
	return 0

}

// deal deals the initial hands.
//#players.
//#deck.
func deal(hands *[2]*Hand, deck *deck.Deck) {
	for _, hand := range hands {
		for i := 0; i < NOHandInit; i++ {
			hand.Draw(deckDealTroop(deck))
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

//deckToTroop calculate the troop card index from deck index.
func deckToTroop(deckix int) int {
	return deckix + 1
}

//deckFromTroop calculate the deck index from troop card index.
func deckFromTroop(cardix int) int {
	return cardix - 1
}

//deckToTactic calculate the tactic card index from deck index.
func deckToTactic(deckix int) int {
	return deckix + 1 + cards.NOTroop
}

//deckFromTactic calculate the deck index from tactic card index.
func deckFromTactic(cardix int) int {
	return cardix - 1 - cards.NOTroop
}

//GobRegistor register all move interfaces.
func GobRegistor() {
	gob.Register(MoveCardFlag{})
	gob.Register(MoveClaim{})
	gob.Register(MoveDeck{})
	gob.Register(MoveDeserter{})
	gob.Register(MoveRedeploy{})
	gob.Register(MoveScoutReturn{})
	gob.Register(MoveTraitor{})
}

//Save save a game.
//Warning it set gamePos to nil before saving with savePos false,
//the pos is return after save.
func Save(game *Game, file *os.File, savePos bool) (err error) {
	if !savePos {
		pos := game.Pos
		game.Pos = nil
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(game)
		game.Pos = pos
	} else {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(game)
	}
	return err
}

//Load load a game.
func Load(file *os.File) (game *Game, err error) {
	decoder := gob.NewDecoder(file)
	g := *new(Game)

	err = decoder.Decode(&g)
	if err != nil {
		return game, err
	}
	game = &g
	if game.Pos == nil {
		game.CalcPos()
	}

	return game, err
}
func (game *Game) GameMoveLoop(posFunc func(
	gameMoveix int,
	prePos, postPos *GamePos,
	moveCardix, dealtix, moveix int,
	move Move,
	isGiveUp, isPass bool,
	claimFailExs [9][]int,
	mudDishixs []int,
)) {
	moves := game.ResetGame()
	for gameMoveix, moveixs := range moves {
		prePos := game.Pos.Copy()
		move, moveCardix, dealtix, isPass, isGiveUp, claimFailexs, mudDishixs := histMove(moveixs, game)
		posFunc(gameMoveix, prePos, game.Pos, moveCardix, dealtix, moveixs[1], move, isGiveUp, isPass, claimFailexs, mudDishixs)
	}
}
