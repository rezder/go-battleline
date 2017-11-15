package game

import (
	"github.com/rezder/go-battleline/v2/game/card"
	"github.com/rezder/go-battleline/v2/game/pos"
	"math/rand"
	"time"
)

//Game a battleline game.
type Game struct {
	Pos  *Pos
	Hist *Hist
}

//NewGame creates a new battleline game.
func NewGame() (g *Game) {
	g = new(Game)
	g.Pos = NewPos()
	return g
}

//Start starts a game.
func (g *Game) Start(playerIDs [2]int, dealer int) {
	moves := make([]*Move, 0, 75)
	moves = append(moves, moveCreateInit(dealer))
	g.Hist = &Hist{
		Moves:     moves,
		PlayerIDs: playerIDs,
		Time:      time.Now(),
	}
	g.Pos.AddMove(moves[0])
}

//Move makes a move.
func (g *Game) Move(move *Move) (winner int, failedClaimsExs [9][]card.Card) {
	if move.MoveType == MoveTypeAll.Deck || move.MoveType.IsScout() {
		for _, bpMove := range move.Moves {
			cardMove := card.Card(bpMove.Index)
			if cardMove.IsBack() {
				cardBack := card.Back(cardMove)
				if cardBack.IsTac() {
					dealCardix := 0
					if g.Pos.PlayerReturned != pos.NoPlayer {
						for i := len(g.Pos.CardsReturned) - 1; i >= 0; i-- {
							card := g.Pos.CardsReturned[i]
							if card.IsTac() && g.Pos.CardPos[int(card)].IsInDeck() {
								dealCardix = int(card)
								break
							}
						}
					}
					if dealCardix != 0 {
						bpMove.Index = dealCardix
					} else {
						deck := make([]int, 0, 10)
						for tacix := card.NOTroop + 1; tacix < len(g.Pos.CardPos); tacix++ {
							cardPos := g.Pos.CardPos[tacix]
							if cardPos == pos.CardAll.DeckTac {
								deck = append(deck, tacix)
							}
						}
						bpMove.Index = deck[rand.Intn(len(deck))]
					}
				} else {
					dealCardix := 0
					if g.Pos.PlayerReturned != pos.NoPlayer {
						for i := len(g.Pos.CardsReturned) - 1; i >= 0; i-- {
							card := g.Pos.CardsReturned[i]
							if card.IsTroop() && g.Pos.CardPos[int(card)].IsInDeck() {
								dealCardix = int(card)
								break
							}
						}
					}
					if dealCardix != 0 {
						bpMove.Index = dealCardix
					} else {
						deck := make([]int, 0, 46)
						for troopix := 1; troopix <= card.NOTroop; troopix++ {
							cardPos := g.Pos.CardPos[troopix]
							if cardPos == pos.CardAll.DeckTroop {
								deck = append(deck, troopix)
							}
						}
						bpMove.Index = deck[rand.Intn(len(deck))]
					}
				}
			}
		}
	} else if move.MoveType == MoveTypeAll.Cone {
		if len(move.Moves) > 0 {
			move.Moves, failedClaimsExs = removeFailedClaim(move.Moves, move.Mover, g.Pos)
		}
	}
	g.Hist.AddMove(move)
	winner = g.Pos.AddMove(move)
	return winner, failedClaimsExs
}
func removeFailedClaim(
	moves []*BoardPieceMove,
	mover int,
	oldPos *Pos) (updMoves []*BoardPieceMove, failedClaimsExs [9][]card.Card) {
	var deleteBPIx []int
	posCards := NewPosCards(oldPos.CardPos)
	deckTroops := posCards.SimDeckTroops()
	for i, bpMove := range moves {
		flagix := bpMove.Index - 1
		flag := NewFlag(flagix, posCards, oldPos.ConePos)
		isClaim, exCardixs := flag.IsClaimable(mover, deckTroops)
		if !isClaim {
			deleteBPIx = append(deleteBPIx, i)
			failedClaimsExs[flagix] = exCardixs
		}
	}
	if len(deleteBPIx) > 0 {
		noUpdMoves := len(moves) - len(deleteBPIx)
		if noUpdMoves > 0 {
			updMoves = make([]*BoardPieceMove, 0, noUpdMoves)
			for i, bpMove := range moves {
				if len(deleteBPIx) == 0 || i != deleteBPIx[0] {
					updMoves = append(updMoves, bpMove)
				} else {
					deleteBPIx = deleteBPIx[1:]
				}
			}
		}
	} else {
		updMoves = moves
	}
	return updMoves, failedClaimsExs
}

// LoadHist loads a old history in to a game.
// the game postion is not upddate, use resume or scroll
// to use the history.
func (g *Game) LoadHist(hist *Hist) {
	g.Hist = hist
}

//Resume moves a game to the last postion of it history
// and wait for new moves. Returns ok if
//the game is not finished.
func (g *Game) Resume() (ok bool) {
	winner, okForward := g.ScrollForward()
	for okForward {
		winner, okForward = g.ScrollForward()
	}

	if g.Hist.RemovePause() {
		lastMove := g.Hist.LastMove()
		g.Pos.RemovePause(lastMove.MoveType, lastMove.Mover)
	}
	return okForward && winner == pos.NoPlayer
}

//IsAtBeginningOfHist checks is a game position is at
//the begining of history.
func (g *Game) IsAtBeginningOfHist() bool {
	return g.Pos.LastMoveIx == -1
}

//ScrollForward scrolls the game postion one move forward
//using history
func (g *Game) ScrollForward() (winner int, ok bool) {
	winner = pos.NoPlayer
	if g.Pos.LastMoveIx != len(g.Hist.Moves)-1 {
		move := g.Hist.Moves[g.Pos.LastMoveIx+1]
		winner = g.Pos.AddMove(move)
		if winner == pos.NoPlayer {
			ok = true
		}
	}
	return winner, ok
}

//ScrollBackward scrolls the game position one move back.
//returns false if at the end.
func (g *Game) ScrollBackward() (ok bool) {
	if !g.IsAtBeginningOfHist() {
		move := g.Hist.Moves[g.Pos.LastMoveIx]
		var before *Move
		if g.Pos.LastMoveIx-1 >= 0 {
			before = g.Hist.Moves[g.Pos.LastMoveIx-1]
		}
		g.Pos.RemoveMove(move, before)
		ok = true
	}
	return ok
}

//GiveUp gives up the game.
func (g *Game) GiveUp(moves []*Move) (winner int) {
	mover := moves[0].Mover
	move := CreateMoveGivUp(g.Pos.ConePos, mover)
	winner, _ = g.Move(move)
	return winner
}

//Pause pauses the game.
func (g *Game) Pause(moves []*Move) (winner int) {
	mover := moves[0].Mover
	move := NewMove(mover, MoveTypeAll.Pause)
	winner, _ = g.Move(move)
	return winner
}

// Hist the history of a battleline game, every move made.
type Hist struct {
	Moves     []*Move
	PlayerIDs [2]int
	Time      time.Time
}

// Copy makes a copy of game history
func (h *Hist) Copy() (copy *Hist) {
	if h != nil {
		copy = new(Hist)
		copy.Time = h.Time
		copy.PlayerIDs = h.PlayerIDs
		if h.Moves != nil {
			copy.Moves = make([]*Move, len(h.Moves))
			for i, refMove := range h.Moves {
				copy.Moves[i] = refMove.Copy()
			}
		}
	}
	return copy
}

// IsEqual checks if two moves are equal.
func (h *Hist) IsEqual(o *Hist) bool {
	if o == h {
		return true
	}
	if (o == nil && h != nil) || (o != nil && h == nil) {
		return false
	}
	isEqual := false
	if h.Time.Equal(o.Time) && h.PlayerIDs == o.PlayerIDs {
		if len(h.Moves) == len(o.Moves) {
			isEqual = true
			for i, move := range h.Moves {
				if !move.IsEqual(o.Moves[i]) {
					isEqual = false
					break
				}
			}
		}
	}
	return isEqual
}

// AddMove adds a move to history.
func (h *Hist) AddMove(move *Move) {
	h.Moves = append(h.Moves, move)
}

// RemovePause remove the last move if it is pause.
func (h *Hist) RemovePause() bool {
	if len(h.Moves) > 0 && h.LastMove().IsPause() {
		h.Moves = h.Moves[:len(h.Moves)-1]
		return true
	}
	return false
}

//LastMove return last move maybe nil
func (h *Hist) LastMove() (move *Move) {
	if len(h.Moves) > 0 {
		move = h.Moves[len(h.Moves)-1]
	}
	return move
}

//Winner returns winner may be NoPlayer if
//game is paused.
func (h *Hist) Winner() (winner int) {
	winner = pos.NoPlayer
	lastMove := h.LastMove()
	if !lastMove.IsPause() {
		if lastMove.MoveType == MoveTypeAll.GiveUp {
			winner = opp(lastMove.Mover)
		} else {
			winner = lastMove.Mover
		}
	}

	return winner
}
