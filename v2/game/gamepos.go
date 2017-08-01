package game

import (
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
)

const (
	//NoPlayer is the value a player index is set to when
	//no player exist.
	NoPlayer = 2
)

var (
	//ViewAll all views.
	ViewAll ViewAllST
)

func init() {
	ViewAll = newViewAllST()
}

// Pos is a battleline game position.
type Pos struct {
	CardPos [71]pos.Card
	ConePos [10]pos.Cone

	PlayerReturned int
	CardsReturned  [2]card.Move

	LastMoveType MoveType
	LastMover    int
	LastMoveIx   int
}

//IsEqual check if two postion is equal.
func (g *Pos) IsEqual(o *Pos) bool {
	if g == o {
		return true
	}
	if (g == nil && o != nil) || (g != nil && o == nil) {
		return false
	}
	if g.LastMoveIx == o.LastMoveIx &&
		g.LastMoveType == o.LastMoveType &&
		g.LastMover == o.LastMover &&
		g.CardPos == o.CardPos &&
		g.ConePos == o.ConePos &&
		g.PlayerReturned == o.PlayerReturned &&
		g.CardsReturned == o.CardsReturned {
		return true

	}

	return false
}

//NewPos creates a new postion.
func NewPos() (g *Pos) {
	g = new(Pos)
	for tacix := card.NOTroop + 1; tacix < card.NOTroop+card.NOTac+1; tacix++ {
		g.CardPos[tacix] = pos.CardAll.DeckTac
	}
	g.PlayerReturned = NoPlayer
	g.LastMover = NoPlayer
	g.LastMoveType = MoveTypeAll.None
	g.LastMoveIx = -1
	return g
}

//AddMove adds a move to the postion.
func (g *Pos) AddMove(gameMove *Move) (winner int) {
	winner = NoPlayer
	g.LastMoveIx = g.LastMoveIx + 1
	g.LastMoveType = gameMove.MoveType
	g.LastMover = gameMove.Mover
	if !gameMove.MoveType.IsPause() {
		for i, move := range gameMove.Moves {
			if move.IsCard() {
				g.CardPos[move.Index] = pos.Card(move.NewPos)
				if gameMove.MoveType == MoveTypeAll.ScoutReturn {
					g.CardsReturned[i] = card.Move(move.Index)
					g.PlayerReturned = gameMove.Mover
				}
			} else {
				g.ConePos[move.Index] = pos.Cone(move.NewPos)
			}
		}
		if gameMove.MoveType == MoveTypeAll.GiveUp {
			winner = opp(gameMove.Mover)
		} else if gameMove.MoveType == MoveTypeAll.Cone {
			winner = calcWinner(g.ConePos)
		}
	}
	return winner
}
func calcWinner(conePos [10]pos.Cone) int {
	winner := NoPlayer
	var playerTotal [2]int
	var inARow int
	lastWinner := -1
	for cone, posCone := range conePos {
		if cone > 0 {
			if posCone.IsWon() {
				playerTotal[posCone.Winner()] = playerTotal[posCone.Winner()] + 1
				if lastWinner == posCone.Winner() {
					inARow = inARow + 1
					if inARow == 3 {
						break
					}
				} else {
					inARow = 1
					lastWinner = posCone.Winner()
				}
			} else {
				inARow = 0
				lastWinner = -1
			}
		}
	}
	if inARow == 3 {
		winner = lastWinner
	} else if playerTotal[0] > 4 {
		winner = 0
	} else if playerTotal[1] > 4 {
		winner = 1
	}
	return winner
}

//RemovePause removes a pause move.
//Assume pause is always after initMove
func (g *Pos) RemovePause(lastMoveType MoveType) {
	g.LastMoveIx = g.LastMoveIx - 1
	g.LastMoveType = lastMoveType
}

//RemoveMove removes a move from postion.
func (g *Pos) RemoveMove(gameMove, beforeGameMove *Move) {
	g.LastMoveIx = g.LastMoveIx - 1
	g.LastMover, g.LastMoveType = beforeGameMove.GetMoverAndType()
	if !gameMove.MoveType.IsPause() {
		for i, move := range gameMove.Moves {
			if move.IsCard() {
				g.CardPos[move.Index] = pos.Card(move.OldPos)
				if gameMove.MoveType == MoveTypeAll.ScoutReturn {
					g.CardsReturned[i] = 0
					g.PlayerReturned = NoPlayer
				}
			} else {
				g.ConePos[move.Index] = pos.Cone(move.OldPos)
			}
		}
	}
}

// CalcMoves calulates all the possible moves for the postion.
func (g *Pos) CalcMoves() (moves []*Move) {
	if g.LastMoveType.HasNext() {
		mover := g.LastMover
		moveType := g.LastMoveType
		var posCards PosCards
		for len(moves) == 0 {
			moveType, mover = moveType.Next(mover)
			moves, posCards = calcMovesLoop(mover, moveType, g.ConePos, g.CardPos, posCards)
		}
	}
	return moves
}

//opp returns the opponent to a player.
func opp(player int) (opponent int) {
	if player == NoPlayer {
		opponent = NoPlayer
	} else {
		opponent = player + 1
		if opponent > 1 {
			opponent = 0
		}
	}
	return opponent
}

// ViewPos a view of a game postion.
type ViewPos struct {
	*Pos
	View
	Winner   int
	NoTacs   [2]int
	NoTroops [2]int
	Moves    []*Move
}

// IsEqual checks if two views are equal.
func (v *ViewPos) IsEqual(o *ViewPos) bool {
	if o == v {
		return true
	}
	if (o == nil && v != nil) || (v == nil && o != nil) {
		return false
	}
	isEqual := false
	if v.View == o.View &&
		v.NoTroops == o.NoTroops &&
		v.NoTacs == o.NoTacs &&
		v.Winner == o.Winner &&
		v.Pos.IsEqual(o.Pos) {
		isEqual = true
		for i, move := range v.Moves {
			if !move.IsEqual(o.Moves[i]) {
				isEqual = false
				break
			}
		}
	}
	return isEqual
}

// NewViewPos create new view of a game postion.
func NewViewPos(gamePos *Pos, view View, winner int) (v *ViewPos) {
	v = new(ViewPos)
	dr := *gamePos
	v.Pos = &dr
	v.View = view
	v.Winner = winner

	if winner == NoPlayer {
		moves := gamePos.CalcMoves()
		if view.IsViewSeePlayer(moves[0].Mover) {
			v.Moves = moves
		}
	}
	for cardix := 1; cardix < len(v.Pos.CardPos); cardix++ {
		cardMove := card.Move(cardix)
		cardMoveReturnedix := -1
		if cardMove == v.CardsReturned[0] {
			cardMoveReturnedix = 0
		} else if cardMove == v.CardsReturned[1] {
			cardMoveReturnedix = 1
		}
		isCardMoveReturned := cardMoveReturnedix != -1
		if view.DontSeePos(v.Pos.CardPos[cardix]) { //Opponent hand
			if !isCardMoveReturned || (isCardMoveReturned && !view.IsViewSeePlayer(v.PlayerReturned)) {
				posPlayer := v.Pos.CardPos[cardix].Player()
				if cardMove.IsTac() {
					v.NoTacs[posPlayer] = v.NoTacs[posPlayer] + 1
					v.CardPos[cardix] = pos.CardAll.DeckTac
				} else {
					v.CardPos[cardix] = pos.CardAll.DeckTroop
					v.NoTroops[posPlayer] = v.NoTroops[posPlayer] + 1
				}
				if isCardMoveReturned {
					v.CardsReturned[cardMoveReturnedix] = 0
				}
			}
		} else {
			if isCardMoveReturned && !view.IsViewSeePlayer(v.PlayerReturned) {
				if v.CardPos[cardix].IsInDeck() {
					if cardMove.IsTac() {
						v.CardsReturned[cardMoveReturnedix] = card.BACKTac
					} else {
						v.CardsReturned[cardMoveReturnedix] = card.BACKTroop
					}
				} else if view.IsPlayer() && v.CardPos[cardix] == pos.CardAll.Players[view.Playerix()].Hand {
					v.CardsReturned[cardMoveReturnedix] = cardMove
				} else {
					v.CardsReturned[cardMoveReturnedix] = 0
				}
			} else if isCardMoveReturned {
				v.CardsReturned[cardMoveReturnedix] = cardMove
			}
		}
	}
	return v
}

//View a view of the gane position.
type View uint8

//IsPlayer return true if the view is a player.
func (v View) IsPlayer() bool {
	return v < 2
}

//Playerix returns the player index if the
//view is a player else NoPlayer.
func (v View) Playerix() int {
	if v < 2 {
		return int(v)
	}
	return NoPlayer
}

//IsViewSeePlayer return true if the viewer can see the players card.
func (v View) IsViewSeePlayer(player int) (dont bool) {
	if v == ViewAll.God {
		dont = true
	} else if v == ViewAll.Spectator {
		dont = false
	} else if player == NoPlayer {
		dont = false
	} else {
		dont = v.Playerix() == player
	}
	return dont
}

//DontSeePos return true if the position is not possible to see,
//by the viewer.
func (v View) DontSeePos(position pos.Card) (dont bool) {
	dont = false
	if position.IsOnHand() {
		dont = v.Playerix() != position.Player()
		if v == ViewAll.God {
			dont = false
		}
	}
	return dont
}

// ViewAllST All the Views singleton
type ViewAllST struct {
	Players   [2]View
	Spectator View
	God       View
}

// All returns all views.
func (v ViewAllST) All() []View {
	return []View{0, 1, 2, 3}
}
func newViewAllST() (v ViewAllST) {
	v.God = 3
	v.Spectator = 2
	v.Players[0] = 0
	v.Players[1] = 1
	return v
}

//PosCards cards sorted according to their current position
type PosCards [][]card.Move

//NewPosCards creates a list of cards for every postion.
func NewPosCards(cardPos [71]pos.Card) (posCards PosCards) {
	posCards = make([][]card.Move, pos.CardAll.Size)
	for i := range posCards {
		if i == 0 {
			posCards[i] = make([]card.Move, 0, 60)
		} else {
			posCards[i] = make([]card.Move, 0, 10)
		}
	}
	for cardix, cardPos := range cardPos {
		if cardix > 0 {
			posCards[int(cardPos)] = append(posCards[int(cardPos)], card.Move(cardix))
		}
	}
	return posCards
}

//Cards return the cards belong to a position.
func (posCards PosCards) Cards(posCard pos.Card) []card.Move {
	return posCards[int(posCard)]
}

func appendSortedTroop(troops []card.Troop, troop card.Troop) []card.Troop {
	no := len(troops)
	troops = append(troops, 0)
	for i, t := range troops {
		if i == no {
			troops[i] = troop
		} else {
			if !(troop.Strenght() < t.Strenght()) {
				copy(troops[i+1:], troops[i:])
				troops[i] = troop
				break
			}
		}
	}
	return troops
}

//SortedCards returns the cards belonging to a postion
// sorted after type and the troops is sorted after strenght
// strongest first.
func (posCards PosCards) SortedCards(posCard pos.Card) (
	troops []card.Troop,
	morales []card.Morale,
	guiles []card.Guile,
	envs []card.Env) {

	cards := posCards[int(posCard)]
	for _, cardMove := range cards {
		switch {
		case cardMove.IsTroop():
			troops = appendSortedTroop(troops, card.Troop(cardMove))
		case cardMove.IsEnv():
			envs = append(envs, card.Env(cardMove))
		case cardMove.IsGuile():
			guiles = append(guiles, card.Guile(cardMove))
		case cardMove.IsMorale():
			morales = append(morales, card.Morale(cardMove))
		}
	}
	return troops, morales, guiles, envs
}
