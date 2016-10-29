package gamepos

import (
	"fmt"
	"github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battbot/flag"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-battleline/battserver/tables"
	"log"
)

//Pos is a game position.
type Pos struct {
	playHand *bat.Hand
	flags    [bat.FLAGS]*flag.Flag
	playDish *bat.Dish
	oppDish  *bat.Dish
	deck     *deck.Deck
	turn     *pub.Turn
}

//New create a game position.
func New() (pos *Pos) {
	pos = new(Pos)
	pos.playHand = bat.NewHand()
	for i := range pos.flags {
		pos.flags[i] = flag.New()
	}
	pos.playDish = bat.NewDish()
	pos.oppDish = bat.NewDish()
	pos.deck = deck.NewDeck()
	return pos
}

//Reset resets the game position to before any move have been made.
func (pos *Pos) Reset() {
	pos.turn = nil
	pos.deck.Reset()
	pos.oppDish = bat.NewDish()
	pos.playDish = bat.NewDish()
	pos.playHand = bat.NewHand()
	for i := range pos.flags {
		pos.flags[i] = flag.New()
	}
}

//UpdMove update position with a move.
//Return true if game is done.
func (pos *Pos) UpdMove(moveView *pub.MoveView) (done bool) {
	if moveView.State == bat.TURN_FINISH || moveView.State == bat.TURN_QUIT {
		done = true
	} else {
		pos.turn = moveView.Turn
		switch move := moveView.Move.(type) {
		case tables.MoveInit:
			for _, cardix := range move.Hand {
				pos.playHand.Draw(cardix)
				pos.deck.PlayDraw(cardix)
			}
		case tables.MoveInitPos:
			for i := range pos.flags {
				pos.flags[i] = flag.TransferTableFlag(move.Pos.Flags[i])
				pos.deck.InitRemoveCards(pos.flags[i].OppEnvs)
				pos.deck.InitRemoveCards(pos.flags[i].OppTroops)
				pos.deck.InitRemoveCards(pos.flags[i].PlayEnvs)
				pos.deck.InitRemoveCards(pos.flags[i].PlayTroops)

			}
			pos.deck.InitRemoveCards(move.Pos.DishTacs)
			for _, cardix := range move.Pos.DishTacs {
				pos.playDish.DishCard(cardix)
			}
			pos.deck.InitRemoveCards(move.Pos.DishTroops)
			for _, cardix := range move.Pos.DishTroops {
				pos.playDish.DishCard(cardix)
			}
			pos.deck.InitRemoveCards(move.Pos.OppDishTacs)
			for _, cardix := range move.Pos.OppDishTacs {
				pos.oppDish.DishCard(cardix)
			}
			pos.deck.InitRemoveCards(move.Pos.OppDishTroops)
			for _, cardix := range move.Pos.OppDishTroops {
				pos.oppDish.DishCard(cardix)
			}
			oppTroops := 0
			oppTacs := 0
			for _, troop := range move.Pos.OppHand {
				if troop {
					oppTroops = oppTroops + 1
				} else {
					oppTacs = oppTacs + 1
				}
			}
			//oppTroops init to 7 so must set incase opponent have less
			pos.deck.OppSetInitHand(oppTroops, oppTacs)

			for _, cardix := range move.Pos.Hand {
				pos.deck.PlayDraw(cardix)
				pos.playHand.Draw(cardix)
			}
		case bat.MoveCardFlag:
			if moveView.Mover {
				pos.playHand.Play(moveView.MoveCardix)
				pos.flags[move.Flagix].PlayAddCardix(moveView.MoveCardix)
			} else {
				pos.deck.OppPlay(moveView.MoveCardix)
				pos.playHand.Play(moveView.MoveCardix)
				pos.flags[move.Flagix].OppAddCardix(moveView.MoveCardix)
			}
		case bat.MoveDeck:
			if moveView.Mover {
				pos.playHand.Draw(moveView.DeltCardix)
				pos.deck.PlayDraw(moveView.DeltCardix)
			} else { //Opponent
				pos.deck.OppDraw(move.Deck == bat.DECK_TROOP)
			}
		case tables.MoveClaimView:
			if len(move.Claimed) > 0 {
				var claimed int
				if moveView.Mover {
					claimed = flag.CLAIMPlay
				} else {
					claimed = flag.CLAIMOpp
				}
				for _, v := range move.Claimed {
					pos.flags[v].Claimed = claimed
				}
			}
		case tables.MoveDeserterView:
			flag := pos.flags[move.Move.Flag]
			if moveView.Mover {
				pos.playHand.Play(moveView.MoveCardix)     //Deserter card
				pos.playDish.DishCard(moveView.MoveCardix) //Deserter card

				pos.oppDish.DishCard(move.Move.Card) //Target card
				flag.OppRemoveCardix(move.Move.Card)
			} else { //Opp move
				pos.deck.OppPlay(moveView.MoveCardix)     //Deserter card
				pos.oppDish.DishCard(moveView.MoveCardix) //Deserter card

				pos.playDish.DishCard(move.Move.Card) //Target card
				flag.PlayRemoveCardix(move.Move.Card)
			}
			updateMudDishixs(flag, move.Dishixs, pos.oppDish, pos.playDish)
		case tables.MoveScoutReturnView:
			if moveView.Mover {
				pos.playHand.Play(cards.TCScout)
				pos.playDish.DishCard(cards.TCScout)
			} else {
				pos.deck.OppPlay(cards.TCScout)
				pos.playDish.DishCard(cards.TCScout)
				pos.deck.OppScoutReturn(move.Troop, move.Tac)
			}
		case bat.MoveTraitor:
			outFlag := pos.flags[move.OutFlag]
			inFlag := pos.flags[move.InFlag]

			if moveView.Mover {
				pos.playHand.Play(moveView.MoveCardix)     //Traitor card
				pos.playDish.DishCard(moveView.MoveCardix) //Traitor card

				outFlag.OppRemoveCardix(move.OutCard)
				inFlag.PlayAddCardix(move.OutCard)
			} else { //Opp move
				pos.deck.OppPlay(moveView.MoveCardix)     //Traitor card
				pos.oppDish.DishCard(moveView.MoveCardix) //Traitor card

				outFlag.PlayRemoveCardix(move.OutCard)
				inFlag.OppAddCardix(move.OutCard)
			}
		case tables.MoveRedeployView:
			outFlag := pos.flags[move.Move.OutFlag]
			var inFlag *flag.Flag
			if move.Move.InFlag >= 0 {
				inFlag = pos.flags[move.Move.InFlag]
			}
			if moveView.Mover {
				pos.playHand.Play(moveView.MoveCardix)     //Redeploy card
				pos.playDish.DishCard(moveView.MoveCardix) //Redeploy card
				outFlag.PlayRemoveCardix(move.Move.OutCard)
				if inFlag != nil {
					inFlag.PlayAddCardix(move.Move.OutCard)
				} else {
					pos.playDish.DishCard(move.Move.OutCard)
				}

			} else {
				pos.deck.OppPlay(moveView.MoveCardix)     //Redeploy card
				pos.oppDish.DishCard(moveView.MoveCardix) //Redeploy card
				outFlag.OppRemoveCardix(move.Move.OutCard)
				if inFlag != nil {
					inFlag.OppAddCardix(move.Move.OutCard)
				} else {
					pos.oppDish.DishCard(move.Move.OutCard)
				}
			}
			updateMudDishixs(outFlag, move.RedeployDishixs, pos.oppDish, pos.playDish)

		case tables.MovePass:
		case tables.MoveQuit:
			done = true
		case tables.MoveSave:
			done = true
		default:
			panic("Missing type implementation for move:")
		}
	}
	return done
}

//updateMudDishixs update the dishes with extra cards that were removed from the
//flag do to mud card no long exist on the flag.
func updateMudDishixs(flag *flag.Flag, dishixs []int, oppDish *bat.Dish, playDish *bat.Dish) {
	for _, cardix := range dishixs {
		if flag.OppRemoveCardix(cardix) {
			oppDish.DishCard(cardix)
		}
		if flag.PlayRemoveCardix(cardix) {
			playDish.DishCard(cardix)
		}
	}
}
func (pos *Pos) String() string {
	res := "&Pos{"
	res = res + fmt.Sprintf("playHand: %v\n", *pos.playHand)
	res = res + fmt.Sprintf("playDish: %v\n", *pos.playDish)
	res = res + "flags: ["
	for _, v := range pos.flags {
		res = res + fmt.Sprint(*v)
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("oppDish: %v\n", *pos.oppDish)
	res = res + fmt.Sprintf("deck: %v\n", *pos.deck)
	res = res + "}"
	return res
}

//MakeMove calculate the next bot move.
func (pos *Pos) MakeMove() (moveix [2]int) {
	if pos.turn.Moves != nil {
		moveix[1] = 0
		switch move := pos.turn.Moves[0].(type) {
		case bat.MoveDeck:
			moveix[1] = makeMoveDeck(pos)

		case bat.MoveClaim:
			moveix[1] = makeMoveClaim(pos.turn.Moves)
		case bat.MoveScoutReturn:
			moveix[1] = makeMoveScoutReturn(pos)
		default:
			log.Panicf("All cases should be covered %+v", move)

		}

	} else {
		moveix = makeMoveHand(pos)
	}
	return moveix
}

//IsBotTurn returns if it bot time to move.
func (pos *Pos) IsBotTurn() bool {
	return pos.turn.MyTurn
}
