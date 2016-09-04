package main

import (
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	slice "github.com/rezder/go-slice/int"
)

//Deck contain all information of unknown cards.
//That includes decks and the hand of the opponent.
//It also tracs cards returned to deck when scout is played.
type Deck struct {
	troops            map[int]bool
	tacs              map[int]bool
	scoutReturnTroops []int
	scoutReturnTacs   []int
	oppHand           []int //contains opponent scout return cards until played.
	oppTroops         int
	oppTacs           int
}

//NewDeck creates a new deck.
func NewDeck() (deck *Deck) {
	deck = new(Deck)
	deck.oppTroops = bat.HAND
	deck.troops = make(map[int]bool)
	deck.tacs = make(map[int]bool)
	initDecks(deck.troops, deck.tacs)
	deck.scoutReturnTacs = make([]int, 0, 2)
	deck.scoutReturnTroops = make([]int, 0, 2)
	deck.oppHand = make([]int, 0, 2)
	return deck
}

//initDecks initialize the deck content to all cards.
func initDecks(troops map[int]bool, tacs map[int]bool) {
	for i := 1; i <= cards.TROOP_NO; i++ {
		troops[i] = true
	}
	for i := cards.TROOP_NO + 1; i <= cards.TAC_NO+cards.TROOP_NO; i++ {
		tacs[i] = true
	}
	return
}

//Reset reset the deck to its initial state.
func (d *Deck) Reset() {
	d.oppTroops = bat.HAND
	d.oppTacs = 0
	initDecks(d.troops, d.tacs)
	d.scoutReturnTacs = d.scoutReturnTacs[:0]
	d.scoutReturnTroops = d.scoutReturnTroops[:0]
	d.oppHand = d.oppHand[:0]
}

//PlayDraw updates the deck with a card drawn by the bot.
func (d *Deck) PlayDraw(cardix int) {
	if cardix <= cards.TROOP_NO {
		nscout := len(d.scoutReturnTroops)
		if nscout == 0 {
			delete(d.troops, cardix)
		} else {
			d.scoutReturnTroops = d.scoutReturnTroops[:nscout-1]
		}
	} else {
		nscout := len(d.scoutReturnTacs)
		if nscout == 0 {
			delete(d.tacs, cardix)
		} else {
			d.scoutReturnTacs = d.scoutReturnTacs[:nscout-1]
		}
	}
}

//OppPlay update the deck with a card played by the opponent of the bot.
func (d *Deck) OppPlay(cardix int) {
	nHand := len(d.oppHand)
	if nHand != 0 {
		d.oppHand = slice.Remove(d.oppHand, cardix)
	}
	if nHand == len(d.oppHand) {
		if cardix <= cards.TROOP_NO {
			delete(d.troops, cardix)
			d.oppTroops = d.oppTroops - 1
		} else {
			delete(d.tacs, cardix)
			d.oppTacs = d.oppTacs - 1
		}
	}
}

//OppDraw update the deck with a card drawn by the opponent of the bot.
func (d *Deck) OppDraw(troop bool) {
	if troop {
		nscout := len(d.scoutReturnTroops)
		if nscout == 0 {
			d.oppTroops = d.oppTroops + 1
		} else {
			d.oppHand = append(d.oppHand, d.scoutReturnTroops[nscout-1])
			d.scoutReturnTroops = d.scoutReturnTroops[:nscout-1]
		}
	} else {
		nscout := len(d.scoutReturnTacs)
		if nscout == 0 {
			d.oppTacs = d.oppTacs + 1
		} else {
			d.oppHand = append(d.oppHand, d.scoutReturnTacs[nscout-1])
			d.scoutReturnTacs = d.scoutReturnTacs[:nscout-1]
		}
	}
}

//OppSetInitHand sets the opponents initial hand.
//Only used when restarting a old game.
//The deck initialize with 7 troops card.
func (d *Deck) OppSetInitHand(troops int, tacs int) {
	d.oppTacs = tacs
	d.oppTroops = troops
}

//OppScoutReturn update the deck with the opponent scout return information.
func (d *Deck) OppScoutReturn(troops int, tacs int) {
	d.oppTroops = d.oppTroops - troops
	d.oppTacs = d.oppTacs - tacs
}

//PlayScoutReturn registor a scout return move. Cards are delt from the back.
func (d *Deck) PlayScoutReturn(troops []int, tacs []int) {
	d.scoutReturnTroops = troops
	d.scoutReturnTacs = tacs
}

//DeckTacNo return the deck size but does not include known cards from scout return
func (d *Deck) DeckTacNo() int {
	return len(d.tacs) - d.oppTacs
}

//DeckTroopNo return the deck size but does not include known cards from scout return
func (d *Deck) DeckTroopNo() int {
	return len(d.troops) - d.oppTroops
}
