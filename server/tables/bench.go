package tables

import (
	bat "rezder.com/game/card/battleline"
	pub "rezder.com/game/card/battleline/server/publist"
	slice "rezder.com/slice/int"
)

func bench(watchChan *pub.WatchChan, game <-chan *MoveBenchView) {
	initWatchers := make(map[chan<- pub.MoveBench]bool)
	watchers := make(map[chan<- pub.MoveBench]bool)
Loop:
	for {
		select {
		case p := <-watchChan.Channel: //If allready there it is close else join
			exist := initWatchers[p.Send]
			if exist {
				close(p.Send)
				delete(initWatchers, p.Send)
			} else {
				exist = watchers[p.Send]
				if exist {
					close(p.Send)
					delete(watchers, p.Send)
				}
			}
			if !exist {
				initWatchers[p.Send] = true
			}
		case moveView, open := <-game:
			if !open {
				close(watchChan.Close) //stope join and leave
				if len(initWatchers) > 0 {
					for key, _ := range initWatchers {
						close(key)
					}
				}
				if len(watchers) > 0 {
					for key, _ := range watchers {
						close(key)
					}
				}
				break Loop
			} else {
				if len(initWatchers) > 0 {
					move := *NewMoveBench(moveView, true)
					for key, _ := range initWatchers {
						key <- move
					}
				}
				if len(watchers) > 0 {
					move := *NewMoveBench(moveView, false)
					for key, _ := range watchers {
						key <- move
					}
				}
			}
		} //select
	} //for
}

type MoveBenchView struct {
	Mover     int
	Move      bat.Move
	NextMover int
	MoveInit  bat.Move
}

func NewMoveBench(view *MoveBenchView, ini bool) (move *pub.MoveBench) {
	move = new(pub.MoveBench)
	move.Mover = view.Mover
	if ini {
		move.Move = view.MoveInit
	} else {
		move.Move = view.Move
	}
	move.NextMover = view.NextMover
	return move
}

type BenchPos struct {
	Flags      [bat.FLAGS]*Flag //Player 0 flags
	DishTroops [2][]int
	DishTacs   [2][]int
	Hands      [2][]bool
	DeckTacs   int
	DeckTroops int
}

func NewBenchPos(pos *bat.GamePos) (bench *BenchPos) {
	bench = new(BenchPos)
	for i, v := range pos.Flags {
		bench.Flags[i] = NewFlag(v, 0)
	}
	for i, dish := range pos.Dishs {
		bench.DishTroops[i] = make([]int, len(dish.Troops))
		for j, v := range dish.Troops {
			bench.DishTroops[i][j] = v
		}
		bench.DishTacs[i] = make([]int, len(dish.Tacs))
		for j, v := range dish.Tacs {
			bench.DishTacs[i][j] = v
		}
	}

	for i, hand := range pos.Hands {
		troops := len(hand.Troops)
		bench.Hands[i] = make([]bool, troops+len(hand.Tacs))
		for j, _ := range hand.Troops {
			bench.Hands[i][j] = true
		}
		for j, _ := range hand.Tacs {
			bench.Hands[i][j+troops] = false
		}
	}
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
		} else if b.DeckTacs == other.DeckTacs && b.DeckTroops == other.DeckTroops {
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

type MoveBenchPos struct {
	Pos *BenchPos
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
