package main

import (
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	"strconv"
)

func makeMove(pos *Pos) (move [2]int) {
	if pos.turn.Moves != nil {
		move[1] = 0
		for i, v := range pos.turn.Moves {
			mv, ok := v.(bat.MoveDeck)
			if ok {
				deck := bat.DECK_TROOP
				if len(pos.playHand.Tacs) == 0 {
					deck = bat.DECK_TAC
				}
				if mv.Deck == deck {
					move[1] = i
					break
				}
			} else {
				break
			}
		}
	} else {
		for is := range pos.turn.MovesHand {
			cardix, _ := strconv.ParseInt(is, 10, 0)
			if cardix > cards.TROOP_NO {
				move[0] = int(cardix)
				break
			}
			move[0] = int(cardix)
		}
	}
	return move
}
