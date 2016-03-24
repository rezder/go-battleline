package battleline

import (
	"fmt"
	"rezder.com/game/card/battleline/cards"
	slice "rezder.com/slice/int"
)

const (
	DISH_TACS   = 6
	DISH_TROOPS = 3
)

type Hand struct {
	Troops []int
	Tacs   []int
}

func NewHand() (hand *Hand) {
	hand = new(Hand)
	hand.Troops = initEmptyHand()
	hand.Tacs = initEmptyHand()
	return hand
}
func initEmptyHand() []int {
	return make([]int, 0, HAND+2)
}
func (hand *Hand) Equal(other *Hand) (equal bool) {
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
func (hand *Hand) Copy() (c *Hand) {
	if hand != nil {
		c = new(Hand)
		if len(hand.Troops) != 0 {
			c.Troops = make([]int, len(hand.Troops), HAND+2)
			copy(c.Troops, hand.Troops)
		} else {
			c.Troops = initEmptyHand()
		}
		if len(hand.Tacs) != 0 {
			c.Tacs = make([]int, len(hand.Tacs), HAND+2)
			copy(c.Tacs, hand.Tacs)
		} else {
			c.Tacs = initEmptyHand()
		}

	}
	return c
}
func (hand *Hand) String() (txt string) {
	if hand == nil {
		txt = "Hand{nil}"
	} else {
		txt = fmt.Sprintf("Hand{%v,%v}", hand.Troops, hand.Tacs)
	}
	return txt
}
func (hand *Hand) play(cardix int) {
	cardTac, _ := cards.DrTactic(cardix)
	if cardTac != nil {
		hand.Tacs = slice.Remove(hand.Tacs, cardix)
	} else if cardTroop, _ := cards.DrTroop(cardix); cardTroop != nil {
		hand.Troops = slice.Remove(hand.Troops, cardix)
	} else {
		panic("Card index is not valid")
	}
}
func (hand *Hand) playMulti(cardixs []int) {
	for _, cardix := range cardixs {
		hand.play(cardix)
	}
}
func (hand *Hand) draw(cardix int) {
	cardTac, _ := cards.DrTactic(cardix)
	if cardTac != nil {
		hand.Tacs = append(hand.Tacs, cardix)
	} else if cardTroop, _ := cards.DrTroop(cardix); cardTroop != nil {
		hand.Troops = append(hand.Troops, cardix)
	} else {
		panic("Card index is not valid")
	}
}

type Dish struct {
	Tacs   []int
	Troops []int
}

func NewDish() (dish *Dish) {
	dish = new(Dish)
	dish.Tacs = dishInitTacs()
	dish.Troops = dishInitTroops()
	return dish
}
func dishInitTacs() []int {
	return make([]int, 0, DISH_TACS)
}
func dishInitTroops() []int {
	return make([]int, 0, DISH_TROOPS)
}
func (dish *Dish) Equal(other *Dish) (equal bool) {
	if other == nil && dish == nil {
		equal = true
	} else if other != nil && dish != nil {
		if dish == other {
			equal = true
		} else if slice.Equal(other.Troops, dish.Troops) && slice.Equal(other.Tacs, dish.Tacs) {
			equal = true
		}
	}

	return equal
}

func (dish *Dish) Copy() (c *Dish) {
	if dish != nil {
		c = new(Dish)
		if len(dish.Troops) != 0 {
			c.Troops = make([]int, len(dish.Troops), DISH_TROOPS)
			copy(c.Troops, dish.Troops)
		} else {
			c.Troops = dishInitTroops()
		}
		if len(dish.Tacs) != 0 {
			c.Tacs = make([]int, len(dish.Tacs), DISH_TACS)
			copy(c.Tacs, dish.Tacs)
		} else {
			c.Tacs = dishInitTacs()
		}
	}
	return c
}
func (dish *Dish) String() (txt string) {
	if dish == nil {
		txt = "Dish{nil}"
	} else {
		txt = fmt.Sprintf("Dish{%v,%v}", dish.Troops, dish.Tacs)
	}
	return txt
}
func (dish *Dish) dishCard(cardix int) {
	if tac, _ := cards.DrTactic(cardix); tac != nil {
		dish.Tacs = append(dish.Tacs, cardix)
	} else if troop, _ := cards.DrTroop(cardix); troop != nil {
		dish.Troops = append(dish.Troops, cardix)
	} else {
		panic("Card should exist")
	}
}
