package tables

import (
	bat "github.com/rezder/go-battleline/battleline"
	pub "github.com/rezder/go-battleline/battserver/publist"
	slice "github.com/rezder/go-slice/int"
)

// The bench server for a table.
// Handle all things related with watching the game. Adding and removing watchers and
// relaying the game information.
func bench(watchChCl *pub.WatchChCl, game <-chan *MoveBenchView) {
	watchers := make(map[int]chan<- *pub.MoveBench)
	var initMove *pub.MoveBench
Loop:
	for {
		select {
		case p := <-watchChCl.Channel:
			_, found := watchers[p.ID]
			del := p.Send == nil
			if found && del {
				delete(watchers, p.ID)
			} else if !found && !del {
				watchers[p.ID] = p.Send
				if initMove != nil {
					p.Send <- initMove
				}
			} else if found && !del {
				close(p.Send)
			}

		case moveView, open := <-game:
			if !open {
				close(watchChCl.Close) //stope join and leave
				if len(watchers) > 0 {
					for _, ch := range watchers {
						close(ch)
					}
				}
				break Loop
			} else {
				initMove = NewMoveBench(moveView, true)
				if len(watchers) > 0 {
					move := NewMoveBench(moveView, false)
					for _, ch := range watchers {
						ch <- move
					}
				}
			}
		} //select
	} //for
}

// MoveBenchView the information send from the table to the bench.
// The informationn is then relayed to all watchers. As watchers
// can be new it must always contain a init move and a move unless it is the
// first move in a new game. This is because we can always have new watchers.
type MoveBenchView struct {
	pub.MoveBench
	MoveInit bat.Move
}

func NewMoveBench(view *MoveBenchView, ini bool) (move *pub.MoveBench) {
	move = new(pub.MoveBench)
	move.Mover = view.Mover
	if ini {
		if view.MoveInit != nil {
			move.Move = view.MoveInit
		} else { //new game first move
			move.Move = view.Move
		}

	} else {
		move.Move = view.Move
		move.MoveCardix = view.MoveCardix
	}
	move.NextMover = view.NextMover
	return move
}

// BenchPos the complete data that describe a game position for
// a watcher.
type BenchPos struct {
	Flags      [bat.NOFlags]*Flag //Player 0 flags
	DishTroops [2][]int
	DishTacs   [2][]int
	Hands      [2][]bool
	DeckTacs   int
	DeckTroops int
	Ids        [2]int
}

func NewBenchPos(pos *bat.GamePos, ids [2]int) (bench *BenchPos) {
	bench = new(BenchPos)
	for i, v := range pos.Flags {
		bench.Flags[i] = NewFlag(v, 0)
	}
	for i, dish := range pos.Dishs {
		bench.DishTroops[i] = make([]int, len(dish.Troops))
		copy(bench.DishTroops[i], dish.Troops)

		bench.DishTacs[i] = make([]int, len(dish.Tacs))
		copy(bench.DishTacs[i], dish.Tacs)
	}

	for i, hand := range pos.Hands {
		troops := len(hand.Troops)
		bench.Hands[i] = make([]bool, troops+len(hand.Tacs))
		for j := range hand.Troops {
			bench.Hands[i][j] = true
		}
		for j := range hand.Tacs {
			bench.Hands[i][j+troops] = false
		}
	}
	bench.Ids = ids
	bench.DeckTroops = len(pos.DeckTroop.Remaining())
	bench.DeckTacs = len(pos.DeckTac.Remaining())
	return bench
}

func (b *BenchPos) Equal(other *BenchPos) (equal bool) {
	if other == nil && b == nil {
		equal = true
	} else if other != nil && b != nil {
		if b == other {
			equal = true
		} else if b.DeckTacs == other.DeckTacs && b.DeckTroops == other.DeckTroops && b.Ids == other.Ids {
			equal = true
			for i := 0; i < 2; i++ {
				if !slice.Equal(b.DishTroops[i], other.DishTroops[i]) {
					equal = false
					break
				}
				if !slice.Equal(b.DishTacs[i], other.DishTacs[i]) {
					equal = false
					break
				}
				if len(b.Hands[i]) != len(other.Hands[i]) {
					equal = false
					break
				} else {
					if len(b.Hands[i]) > 0 {
						for j, v := range b.Hands[i] {
							if v != other.Hands[i][j] {
								equal = false
								break
							}
						}

					}
				}
			}
			if equal {
				for i, v := range other.Flags {
					if !v.Equal(b.Flags[i]) {
						equal = false
						break
					}
				}
			}
		}
	}
	return equal
}

// MoveBenchPos the first move when restarting a game or when
// joining a game.
type MoveBenchPos struct {
	Pos      *BenchPos
	JsonType string
}

func NewMoveBenchPos(pos *BenchPos) *MoveBenchPos {
	res := new(MoveBenchPos)
	res.Pos = pos
	res.JsonType = "MoveBenchPos"
	return res
}
func (m MoveBenchPos) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveBenchPos)
		if ok {
			equal = m.Pos.Equal(om.Pos)
		}
	}
	return equal
}
func (m MoveBenchPos) Copy() (c bat.Move) {
	c = m //no deep copy
	return c
}

// MoveBenchInit the first move in new game.
type MoveBenchInit struct {
	Ids      [2]int
	JsonType string
}

func NewMoveBenchInit(ids [2]int) *MoveBenchInit {
	res := new(MoveBenchInit)
	res.Ids = ids
	res.JsonType = "MoveBenchInit"
	return res
}
func (m MoveBenchInit) MoveEqual(other bat.Move) (equal bool) {
	if other != nil {
		om, ok := other.(MoveBenchInit)
		if ok {
			equal = m.Ids == om.Ids
		}
	}
	return equal
}
func (m MoveBenchInit) Copy() (c bat.Move) {
	c = m //no deep copy
	return c
}
