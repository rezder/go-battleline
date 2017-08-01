package pos

import (
	"encoding/json"
)

const (
	NoPlayer = 2
)

var (
	//CardAll all the card postions.
	CardAll CardAllST
	//ConeAll all the cone postion.
	ConeAll ConeAllST
)

func init() {
	CardAll = newCardAllST()
	ConeAll = newConeAllST()
}

//CardAllST card postion all singleton.
type CardAllST struct {
	Size      int
	DeckTroop Card
	DeckTac   Card
	Players   [2]CardAllPlayerST
}

//newCardAllST create a new all card postion singleton
func newCardAllST() (c CardAllST) {
	c.DeckTroop = 0
	for i := 0; i < 9; i++ {
		c.Players[0].Flags[i] = Card(i + 1)
		c.Players[1].Flags[i] = Card(i + 11)
	}
	c.Players[0].Dish = 10
	c.Players[1].Dish = 20
	c.Players[0].Hand = 21
	c.Players[1].Hand = 22
	c.DeckTac = 23
	c.Size = 24
	return c
}

//CardAllPlayerST player of card postion singleton.
type CardAllPlayerST struct {
	Flags [9]Card
	Dish  Card
	Hand  Card
}

// Card postion.
type Card uint8

//UnmarshalJSON unmarshalls a json number to
//a card postion. This is need because a of a slice of uint8 is confusing
// json unmarchal single value number slice is string.
func (c *Card) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*c = Card(i)
	return err
}

//MarshalJSON marshalls a card postion to json number to make
// a uint8 (byte) readable as a number.
func (c Card) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(c))
}

//Player return the player of the card postion
// if none it returns -1
func (c Card) Player() int {
	if c > 0 && c < 11 {
		return 0
	} else if c > 10 && c < 21 {
		return 1
	} else if c == CardAll.Players[0].Hand {
		return 0
	} else if c == CardAll.Players[1].Hand {
		return 1
	}
	return NoPlayer
}

//IsOnHand returns true is card postion is on either players hand.
func (c Card) IsOnHand() bool {
	return c == CardAll.Players[0].Hand || c == CardAll.Players[1].Hand
}

//IsInDeck returns true if card postion is in deck.
func (c Card) IsInDeck() bool {
	return c == CardAll.DeckTac || c == CardAll.DeckTroop
}

//IsOnTable returns true if card postion is on the table.
func (c Card) IsOnTable() bool {
	return !c.IsOnHand() && !c.IsInDeck()
}

//Cone the cone position.
type Cone uint8

//UnmarshalJSON unmarshalls json number to
//to a cone postion.
func (c *Cone) UnmarshalJSON(b []byte) (err error) {
	var i int
	if err = json.Unmarshal(b, &i); err != nil {
		return err
	}
	*c = Cone(i)
	return err
}

//MarshalJSON marshalls cone to json int.
func (c Cone) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(int(c))
}

//IsWon reurns true if the cone is in a win postion.
func (c Cone) IsWon() bool {
	return c != ConeAll.None
}

//Winner return the winner if any else 2
func (c Cone) Winner() int {
	if c == ConeAll.Players[0] {
		return 0
	} else if c == ConeAll.Players[1] {
		return 1
	}
	return 2
}

// ConeAllST All cone postions singleton.
type ConeAllST struct {
	None    Cone
	Players [2]Cone
}

//newConeAllST returns all cone postion singleton.
func newConeAllST() (c ConeAllST) {
	c.None = 0
	c.Players[0] = 1
	c.Players[1] = 2
	return c
}
