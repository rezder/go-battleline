package flag

import (
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
)

//Deck contain all information of unknown cards.
//That includes decks and the hand of the opponent.
//It also tracks cards returned to deck when scout is played.
type Deck struct {
	cards           *card.Cards
	returnerix      int
	returnedCards   [2]card.Card
	returnedPos     [2]pos.Card
	oppHandTroopsNo int
	oppHandTacsNo   int
	playerix        int
	oppix           int
}

//NewDeck creates a new deck.
func NewDeck(viewPos *game.ViewPos, posCards game.PosCards) (deck *Deck) {
	deck = new(Deck)
	deck.playerix = viewPos.View.Playerix()
	deck.oppix = deck.playerix + 1
	if deck.oppix > 1 {
		deck.oppix = 0
	}
	deckTroops := posCards.SortedCards(pos.CardAll.DeckTroop).Troops
	deck.cards = posCards.SortedCards(pos.CardAll.DeckTac)

	deck.cards.Troops = deckTroops
	deck.returnerix = viewPos.PlayerReturned
	deck.returnedCards = viewPos.CardsReturned
	for i, card := range deck.returnedCards {
		if !card.IsNone() && !card.IsBack() {
			deck.returnedPos[i] = viewPos.CardPos[int(card)]
		}
	}
	deck.oppHandTroopsNo = viewPos.NoTroops[deck.oppix]
	deck.oppHandTacsNo = viewPos.NoTacs[deck.oppix]

	return deck
}

//Troops returns a all troops in the deck.
func (deck *Deck) Troops() map[card.Troop]bool {
	troops := make(map[card.Troop]bool)
	for _, troop := range deck.cards.Troops {
		troops[troop] = true
	}
	return troops
}

//Guiles returns a all guiles cards in the deck.
func (deck *Deck) Guiles() []card.Guile {
	return deck.cards.Guiles
}

//Morales returns a all morales cards in the deck.
func (deck *Deck) Morales() []card.Morale {
	return deck.cards.Morales
}

//Envs returns a all enviroment cards in the deck.
func (deck *Deck) Envs() []card.Env {
	return deck.cards.Envs
}

//Tacs returns the tactic cards in the deck.
//WARNING adds the tactic card together.
func (deck *Deck) Tacs() []card.Card {
	return deck.cards.Tacs()
}

//OppDrawNo calculate the opponent number of unknown cards.
func (deck *Deck) OppDrawNo(isFirst bool) (no int) {
	no = deck.DeckTroopNo()
	if isFirst {
		no = no + 1
	}
	no = (no / 2) + deck.oppHandTroopsNo
	return no
}

//BotDrawNo calculate the bots number of the unknown cards.
func (deck *Deck) BotDrawNo(isFirst bool) (no int) {
	no = deck.DeckTroopNo()
	if isFirst {
		no = no + 1
	}
	no = no / 2
	return no
}

//DeckTacNo returns current the tactic card deck size.
func (deck *Deck) DeckTacNo() int {
	return deck.cards.NoTacs() - deck.oppHandTacsNo
}

//DeckTroopNo returns the current troop deck size.
func (deck *Deck) DeckTroopNo() int {
	return len(deck.cards.Troops) - deck.oppHandTroopsNo
}

//MaxStrenghts returns the 4 max strenghts a deck contain.
func (deck *Deck) MaxStrenghts() []int {
	return maxStrenghts(deck.cards.Troops)
}
func maxStrenghts(troops []card.Troop) (strenghts []int) {
	strenghts = make([]int, 0, 4)
	for i, troop := range troops {
		strenghts = append(strenghts, troop.Strenght())
		if i == 3 {
			break
		}
	}
	return strenghts
}

//OppHandTacNo the opponets number of hand tactic cards.
func (deck *Deck) OppHandTacNo() int {
	return deck.oppHandTacsNo
}

//OppHandTroopsNo the opponets number of hand troops cards.
func (deck *Deck) OppHandTroopsNo() int {
	return deck.oppHandTroopsNo
}

//ScoutReturnTacPeek returns the next tactic card in the deck
//if it is known.
func (deck *Deck) ScoutReturnTacPeek() card.Card {
	if deck.returnerix == deck.playerix {
		for i, cardPos := range deck.returnedPos {
			if cardPos == pos.CardAll.DeckTac {
				return deck.returnedCards[i]
			}
		}
	}
	return 0
}

//ScoutReturnTroopPeek returns the next troop in the deck
//if its known.
func (deck *Deck) ScoutReturnTroopPeek() card.Card {
	if deck.returnerix == deck.playerix {
		for i, cardPos := range deck.returnedPos {
			if cardPos == pos.CardAll.DeckTroop {
				return deck.returnedCards[i]
			}
		}
	}
	return 0
}
