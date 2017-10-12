package gamepos

import (
	"fmt"

	"github.com/rezder/go-battleline/battbot/deck"
	"github.com/rezder/go-battleline/battbot/flag"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-battleline/battleline/cards"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-battleline/battserver/tables"
	"github.com/rezder/go-error/log"
)

//Pos is a game position.
type Pos struct {
	PlayHand *bat.Hand
	Flags    [bat.NOFlags]*flag.Flag
	PlayDish *bat.Dish
	OppDish  *bat.Dish
	Deck     *deck.Deck
	Turn     *pub.Turn
}

//New create a game position.
func New() (pos *Pos) {
	pos = new(Pos)
	pos.PlayHand = bat.NewHand()
	for i := range pos.Flags {
		pos.Flags[i] = flag.New()
	}
	pos.PlayDish = bat.NewDish()
	pos.OppDish = bat.NewDish()
	pos.Deck = deck.NewDeck()
	return pos
}

//Reset resets the game position to before any move have been made.
func (pos *Pos) Reset() {
	pos.Turn = nil
	pos.Deck.Reset()
	pos.OppDish = bat.NewDish()
	pos.PlayDish = bat.NewDish()
	pos.PlayHand = bat.NewHand()
	for i := range pos.Flags {
		pos.Flags[i] = flag.New()
	}
}

//UpdMove update position with a move.
//Return true if game is done.
func (pos *Pos) UpdMove(moveView *pub.MoveView) (done bool) {
	if moveView.State == bat.TURNFinish || moveView.State == bat.TURNQuit {
		done = true
	} else {
		pos.Turn = moveView.Turn
		switch move := moveView.Move.(type) {
		case tables.MoveInit:
			for _, cardix := range move.Hand {
				pos.PlayHand.Draw(cardix)
				pos.Deck.PlayDraw(cardix)
			}
		case tables.MoveInitPos:
			for i := range pos.Flags {
				pos.Flags[i] = flag.TransferTableFlag(move.Pos.Flags[i])
				pos.Deck.InitRemoveCards(pos.Flags[i].OppEnvs)
				pos.Deck.InitRemoveCards(pos.Flags[i].OppTroops)
				pos.Deck.InitRemoveCards(pos.Flags[i].PlayEnvs)
				pos.Deck.InitRemoveCards(pos.Flags[i].PlayTroops)

			}
			pos.Deck.InitRemoveCards(move.Pos.DishTacs)
			for _, cardix := range move.Pos.DishTacs {
				pos.PlayDish.DishCard(cardix)
			}
			pos.Deck.InitRemoveCards(move.Pos.DishTroops)
			for _, cardix := range move.Pos.DishTroops {
				pos.PlayDish.DishCard(cardix)
			}
			pos.Deck.InitRemoveCards(move.Pos.OppDishTacs)
			for _, cardix := range move.Pos.OppDishTacs {
				pos.OppDish.DishCard(cardix)
			}
			pos.Deck.InitRemoveCards(move.Pos.OppDishTroops)
			for _, cardix := range move.Pos.OppDishTroops {
				pos.OppDish.DishCard(cardix)
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
			pos.Deck.OppSetInitHand(oppTroops, oppTacs)

			for _, cardix := range move.Pos.Hand {
				pos.Deck.PlayDraw(cardix)
				pos.PlayHand.Draw(cardix)
			}
		case bat.MoveCardFlag:
			if moveView.Mover {
				pos.PlayHand.Play(moveView.MoveCardix)
				pos.Flags[move.Flagix].PlayAddCardix(moveView.MoveCardix)
			} else {
				pos.Deck.OppPlay(moveView.MoveCardix)
				pos.PlayHand.Play(moveView.MoveCardix)
				pos.Flags[move.Flagix].OppAddCardix(moveView.MoveCardix)
			}
		case bat.MoveDeck:
			if moveView.Mover {
				pos.PlayHand.Draw(moveView.DeltCardix)
				pos.Deck.PlayDraw(moveView.DeltCardix)
				if moveView.MoveCardix == cards.TCScout {
					pos.PlayHand.Play(cards.TCScout)
					pos.PlayDish.DishCard(cards.TCScout)
				}
			} else { //Opponent
				pos.Deck.OppDraw(move.Deck == bat.DECKTroop)
				if moveView.MoveCardix == cards.TCScout {
					pos.Deck.OppPlay(cards.TCScout)
					pos.OppDish.DishCard(cards.TCScout)
				}
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
					pos.Flags[v].Claimed = claimed
				}
			}
		case tables.MoveDeserterView:
			flag := pos.Flags[move.Move.Flag]
			if moveView.Mover {
				pos.PlayHand.Play(moveView.MoveCardix)     //Deserter card
				pos.PlayDish.DishCard(moveView.MoveCardix) //Deserter card

				pos.OppDish.DishCard(move.Move.Card) //Target card
				flag.OppRemoveCardix(move.Move.Card)
			} else { //Opp move
				pos.Deck.OppPlay(moveView.MoveCardix)     //Deserter card
				pos.OppDish.DishCard(moveView.MoveCardix) //Deserter card

				pos.PlayDish.DishCard(move.Move.Card) //Target card
				flag.PlayRemoveCardix(move.Move.Card)
			}
			updateMudDishixs(flag, move.Dishixs, pos.OppDish, pos.PlayDish)
		case tables.MoveScoutReturnView:
			if moveView.Mover {
				//TODO Add opp know card in deck and on hand.
			} else {
				pos.Deck.OppScoutReturn(move.Troop, move.Tac)
			}
		case bat.MoveTraitor:
			outFlag := pos.Flags[move.OutFlag]
			inFlag := pos.Flags[move.InFlag]

			if moveView.Mover {
				pos.PlayHand.Play(moveView.MoveCardix)     //Traitor card
				pos.PlayDish.DishCard(moveView.MoveCardix) //Traitor card

				outFlag.OppRemoveCardix(move.OutCard)
				inFlag.PlayAddCardix(move.OutCard)
			} else { //Opp move
				pos.Deck.OppPlay(moveView.MoveCardix)     //Traitor card
				pos.OppDish.DishCard(moveView.MoveCardix) //Traitor card

				outFlag.PlayRemoveCardix(move.OutCard)
				inFlag.OppAddCardix(move.OutCard)
			}
		case tables.MoveRedeployView:
			outFlag := pos.Flags[move.Move.OutFlag]
			var inFlag *flag.Flag
			if move.Move.InFlag >= 0 {
				inFlag = pos.Flags[move.Move.InFlag]
			}
			if moveView.Mover {
				pos.PlayHand.Play(moveView.MoveCardix)     //Redeploy card
				pos.PlayDish.DishCard(moveView.MoveCardix) //Redeploy card
				outFlag.PlayRemoveCardix(move.Move.OutCard)
				if inFlag != nil {
					inFlag.PlayAddCardix(move.Move.OutCard)
				} else {
					pos.PlayDish.DishCard(move.Move.OutCard)
				}

			} else {
				pos.Deck.OppPlay(moveView.MoveCardix)     //Redeploy card
				pos.OppDish.DishCard(moveView.MoveCardix) //Redeploy card
				outFlag.OppRemoveCardix(move.Move.OutCard)
				if inFlag != nil {
					inFlag.OppAddCardix(move.Move.OutCard)
				} else {
					pos.OppDish.DishCard(move.Move.OutCard)
				}
			}
			updateMudDishixs(outFlag, move.RedeployDishixs, pos.OppDish, pos.PlayDish)

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
func updateMudDishixs(flag *flag.Flag, dishixs []int, OppDish *bat.Dish, PlayDish *bat.Dish) {
	for _, cardix := range dishixs {
		if flag.OppRemoveCardix(cardix) {
			OppDish.DishCard(cardix)
		}
		if flag.PlayRemoveCardix(cardix) {
			PlayDish.DishCard(cardix)
		}
	}
}
func (pos *Pos) String() string {
	res := "&Pos{"
	res = res + fmt.Sprintf("PlayHand: %v\n", *pos.PlayHand)
	res = res + fmt.Sprintf("PlayDish: %v\n", *pos.PlayDish)
	res = res + "flags: ["
	for _, v := range pos.Flags {
		res = res + fmt.Sprint(*v)
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("OppDish: %v\n", *pos.OppDish)
	res = res + fmt.Sprintf("deck: %v\n", *pos.Deck)
	res = res + "}"
	return res
}

//MakeTfMove calculate the next bot move using tensorflow data.
func (pos *Pos) MakeTfMove(probas []float64, moveixs [][2]int) (moveix [2]int) {
	var maxProba float64
	for i, proba := range probas {
		if maxProba < proba {
			maxProba = proba
			moveix = moveixs[i]
		}
	}
	return moveix
}

//MakeMove calculate the next bot move.
func (pos *Pos) MakeMove() (moveix [2]int) {
	if pos.Turn.Moves != nil {
		moveix[1] = 0
		switch move := pos.Turn.Moves[0].(type) {
		case bat.MoveDeck:
			moveix[1] = makeMoveDeck(pos)

		case bat.MoveClaim:
			moveix[1] = makeMoveClaim(pos.Turn.Moves)
		case bat.MoveScoutReturn:
			moveix[1] = makeMoveScoutReturn(pos)
		default:
			logTxt := fmt.Sprintf("All cases should be covered %+v", move)
			log.Print(log.Min, logTxt)
			panic(logTxt)

		}

	} else {
		moveix = makeMoveHand(pos)
	}
	return moveix
}

//IsBotTurn returns if it bot time to move.
func (pos *Pos) IsBotTurn() bool {
	return pos.Turn.MyTurn
}

//IsHandMove returns if hand move.
func (pos *Pos) IsHandMove() bool {
	return pos.Turn.Moves == nil
}
