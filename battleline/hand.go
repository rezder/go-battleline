package battleline

import (
	"fmt"
	"github.com/rezder/go-battleline/battleline/cards"
	slice "github.com/rezder/go-slice/int"
)

type TroopTac struct {
	Troops []int
	Tacs   []int
}

func (hand *TroopTac) Equal(other *TroopTac) (equal bool) {
	if other == nil && hand == nil {
		equal = true
	} else if other != nil && hand != nil {
		if hand == other {
			equal = true
		} else if slice.Equal(other.Troops, hand.Troops) && slice.Equal(other.Tacs, hand.Tacs) {
			equal = true
		}
	}

	return equal
}
func (hand *TroopTac) Copy() (c *TroopTac) {
	if hand != nil {
		c = new(TroopTac)
		c.Troops = make([]int, len(hand.Troops), cap(hand.Troops))
		if len(hand.Troops) != 0 {
			copy(c.Troops, hand.Troops)
		}
		c.Tacs = make([]int, len(hand.Tacs), cap(hand.Tacs))
		if len(hand.Tacs) != 0 {
			copy(c.Tacs, hand.Tacs)
		}

	}
	return c
}

//Hand a battleline hand.
type Hand TroopTac

func (hand *Hand) Copy() *Hand {
	tt := TroopTac(*hand)
	h := Hand(*tt.Copy())
	return &h
}
func (hand *Hand) Equal(other *Hand) bool {
	tt := TroopTac(*hand)
	ott := TroopTac(*other)
	ttp := &tt
	return ttp.Equal(&ott)
}

func NewHand() (hand *Hand) {
	hand = new(Hand)
	hand.Troops = initEmptyHand()
	hand.Tacs = initEmptyHand()
	return hand
}

//initEmptyHand returns a empty slice the capacity of a max hand.
func initEmptyHand() []int {
	return make([]int, 0, 9)
}

func (hand *Hand) String() (txt string) {
	if hand == nil {
		txt = "Hand{nil}"
	} else {
		txt = fmt.Sprintf("Hand{%v,%v}", hand.Troops, hand.Tacs)
	}
	return txt
}

//Play removes a card from the hand.
func (hand *Hand) Play(cardix int) {
	if cards.IsTac(cardix) {
		hand.Tacs = slice.Remove(hand.Tacs, cardix)
	} else if cards.IsTroop(cardix) {
		hand.Troops = slice.Remove(hand.Troops, cardix)
	} else {
		panic("Card index is not valid")
	}
}

//PlayMulti removes cards from hand.
func (hand *Hand) PlayMulti(cardixs []int) {
	for _, cardix := range cardixs {
		hand.Play(cardix)
	}
}

//Draw adds card to hand.
func (hand *Hand) Draw(cardix int) {
	if cards.IsTac(cardix) {
		hand.Tacs = append(hand.Tacs, cardix)
	} else if cards.IsTroop(cardix) {
		hand.Troops = append(hand.Troops, cardix)
	} else {
		panic("Card index is not valid")
	}
}

//Size the total number of cards in hand.
func (hand *Hand) Size() int {
	res := 0
	if hand != nil {
		res = len(hand.Troops) + len(hand.Tacs)
	}
	return res
}

//Dish a container for dished troops and tactics.
type Dish TroopTac

func NewDish() (dish *Dish) {
	dish = new(Dish)
	dish.Tacs = dishInitTacs()
	dish.Troops = dishInitTroops()
	return dish
}
func dishInitTacs() []int {
	return make([]int, 0, 6)
}
func dishInitTroops() []int {
	return make([]int, 0, 3)
}
func (dish *Dish) Copy() *Dish {
	tt := TroopTac(*dish)
	d := Dish(*tt.Copy())
	return &d
}
func (dish *Dish) Equal(other *Dish) bool {
	tt := TroopTac(*dish)
	ott := TroopTac(*other)
	ttp := &tt
	return ttp.Equal(&ott)
}
func (dish *Dish) String() (txt string) {
	if dish == nil {
		txt = "Dish{nil}"
	} else {
		txt = fmt.Sprintf("Dish{%v,%v}", dish.Troops, dish.Tacs)
	}
	return txt
}

//DishCard add card to the dish.
func (dish *Dish) DishCard(cardix int) {
	if cards.IsTac(cardix) {
		dish.Tacs = append(dish.Tacs, cardix)
	} else if cards.IsTroop(cardix) {
		dish.Troops = append(dish.Troops, cardix)
	} else {
		panic("Card should exist")
	}
}
