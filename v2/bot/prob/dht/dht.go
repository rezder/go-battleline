package dht

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
)

const (
	maxNo = 4
)

//Cache the hand deck troops cache.
type Cache struct {
	SrcDeckTroops []card.Troop
	SrcHandTroops [2][]card.Troop
	SrcDrawNos    [2]int
	onlyDeckSet   map[card.Troop]bool //TODO could be faster with out map
	players       [2]player
}

//NewCache create cache.
func NewCache(deckTroops []card.Troop, handTroops [2][]card.Troop, drawNos [2]int) (c *Cache) {
	_ = fmt.Sprintln("") //TODO remove
	c = new(Cache)
	c.SrcDeckTroops = deckTroops
	c.SrcHandTroops = handTroops
	c.SrcDrawNos = drawNos
	return c
}

//OnlyDeckSet returns the deck set  with out any hand troos.
func (c *Cache) OnlyDeckSet() (set map[card.Troop]bool) {
	set = c.onlyDeckSet
	if set == nil {
		c.onlyDeckSet = make(map[card.Troop]bool)
		for _, troop := range c.SrcDeckTroops {
			c.onlyDeckSet[troop] = true
		}
		set = c.onlyDeckSet
	}
	return set
}
func (c *Cache) getColorData(playerix, color int) (d *data) {
	d = c.players[playerix].max[color]
	if d == nil {
		if color == card.COLNone {
			c.players[playerix].max[color] = newData(c.SrcDeckTroops[:c.SrcDrawNos[playerix]], c.SrcHandTroops[playerix])
		} else {
			var colorTroops [card.NOColors + 1][]card.Troop
			for _, troop := range c.SrcDeckTroops {
				colorTroops[troop.Color()] = append(colorTroops[troop.Color()], troop)
			}
			var colorHandTroops [card.NOColors + 1][]card.Troop
			for _, troop := range c.SrcHandTroops[playerix] {
				colorHandTroops[troop.Color()] = append(colorHandTroops[troop.Color()], troop)
			}
			for colorix := 1; colorix < card.NOColors+1; colorix++ {
				troops := colorTroops[colorix]
				if len(troops) > c.SrcDrawNos[playerix] {
					troops = troops[:c.SrcDrawNos[playerix]]
				}
				c.players[playerix].max[colorix] = newData(troops, colorHandTroops[colorix])
			}
		}
		d = c.players[playerix].max[color]
	}
	return d
}

//Sum returns the highest sum you can make with both deck and hand troops.
func (c *Cache) Sum(playerix, color, no int) (sum int, isOk bool) {
	if no > 0 {
		data := c.getColorData(playerix, color)
		if len(data.rolStr) >= no {
			sum = data.rolStr[no-1]
			isOk = true
		}
	} else if no == 0 {
		isOk = true
	}
	return sum, isOk
}

//Troops returns the sorted troops from the hand and deck.
func (c *Cache) Troops(playerix, color, no int) (troops []card.Troop, isOk bool) {
	if no > 0 {
		data := c.getColorData(playerix, color)
		if len(data.troops) >= no {
			troops = data.troops[:no]
			isOk = true
		}
	}
	return troops, isOk
}

//TargetSum returns the result of the target sum calculation.
func (c *Cache) TargetSum(playerix, color, no, targetSum int) (res *TargetResult) {
	res = new(TargetResult)
	if no == 0 {
		if targetSum <= 0 {
			res.IsPossibel = true
			res.IsMade = true
		}
	} else {
		data := c.getColorData(playerix, color)
		if len(data.rolStr) >= no {
			maxSum := data.rolStr[no-1]
			if maxSum >= targetSum {
				res.IsPossibel = true
				if len(data.rolHandOnlyStr) >= no {
					maxHandSum := data.rolHandOnlyStr[no-1]
					if maxHandSum >= targetSum {
						res.IsMade = true
					}
				}
				if !res.IsMade {
					res.NewSum = targetSum - data.rolHandStr[no-1]
					res.NewNo = no - data.rolHandNo[no-1]
					newHandTroops := data.handTroops[data.rolHandNo[no-1]:]
					var trimStr int
					if res.NewNo > 1 {
						maxSum, _ := c.Sum(playerix, color, res.NewNo-1)
						trimStr = res.NewSum - maxSum
					} else {
						trimStr = res.NewSum
					}
					res.ValidDeckTroops = make([]card.Troop, 0, len(data.deckTroops))
					for _, troop := range data.deckTroops {
						if troop.Strenght() >= trimStr {
							res.ValidDeckTroops = append(res.ValidDeckTroops, troop)
						}
					}
					res.ValidHandTroops = make([]card.Troop, 0, len(newHandTroops))
					for _, troop := range newHandTroops {
						if troop.Strenght() >= trimStr {
							res.ValidHandTroops = append(res.ValidHandTroops, troop)
						}
					}

				}

			}
		}
	}

	return res
}

//TargetResult the struct holds the result of a target sum calculation.
type TargetResult struct {
	IsPossibel      bool
	IsMade          bool
	ValidDeckTroops []card.Troop
	ValidHandTroops []card.Troop
	NewSum          int
	NewNo           int
}

//Equal compares to target results.
func (t *TargetResult) Equal(o *TargetResult) (isEqual bool) {
	if t == o {
		return true
	}
	if t.IsPossibel == o.IsPossibel &&
		t.IsMade == o.IsMade &&
		t.NewNo == o.NewNo &&
		t.NewSum == o.NewSum &&
		len(t.ValidDeckTroops) == len(o.ValidDeckTroops) &&
		len(t.ValidHandTroops) == len(o.ValidHandTroops) {
		isEqual = true
		for i, troop := range t.ValidDeckTroops {
			if troop != o.ValidDeckTroops[i] {
				isEqual = false
				break
			}
		}
		if isEqual {
			for i, troop := range t.ValidHandTroops {
				if troop != o.ValidHandTroops[i] {
					isEqual = false
					break
				}
			}
		}
	}
	return isEqual
}

//Set returns the deck hand set all troops from both hand and deck.
func (c *Cache) Set(playerix int) (set map[card.Troop]bool) {
	set = c.players[playerix].set
	if set == nil {
		set = make(map[card.Troop]bool)
		for _, troop := range c.SrcDeckTroops {
			set[troop] = true
		}
		for _, troop := range c.SrcHandTroops[playerix] {
			set[troop] = true
		}
		c.players[playerix].set = set
	}
	return set
}

//SortStrs the deck hand troops sorted after strenght.
func (c *Cache) SortStrs(playix int) (strs [][]card.Troop) {
	strs = c.players[playix].sortStrs
	if strs == nil {
		strs = make([][]card.Troop, card.MAXStr+1)
		for _, troop := range c.SrcDeckTroops {
			strs[troop.Strenght()] = append(strs[troop.Strenght()], troop)
		}
		for _, troop := range c.SrcHandTroops[playix] {
			strs[troop.Strenght()] = append(strs[troop.Strenght()], troop)
		}
		c.players[playix].sortStrs = strs
	}
	return strs
}

//CopyWithOutHand makes a copy with out one troop on the hand.
func (c *Cache) CopyWithOutHand(outTroop card.Troop, playix int) (copyCache *Cache) {
	copyCache = new(Cache)
	copyCache.SrcDeckTroops = c.SrcDeckTroops
	var removeix int
	copyCache.SrcHandTroops[playix], removeix = removeTroop(c.SrcHandTroops[playix], outTroop)
	if removeix == -1 {
		panic("Troop is not on hand")
	}
	copyCache.SrcDrawNos = c.SrcDrawNos
	copyCache.onlyDeckSet = c.onlyDeckSet
	for ix, p := range c.players {
		if playix == ix {
			if p.sortStrs != nil {
				copyCache.players[ix].sortStrs = make([][]card.Troop, card.MAXStr+1)
				for str, troops := range p.sortStrs {
					if str == outTroop.Strenght() {
						copyCache.players[ix].sortStrs[str], _ = removeTroop(troops, outTroop)
					} else {
						copyCache.players[ix].sortStrs[str] = troops
					}
				}
			}
			for color, data := range p.max {
				if data != nil {
					if color == card.COLNone && removeix < 4 {
						copyCache.players[ix].max[color] = nil
					} else if color == outTroop.Color() && len(p.max[color].handTroops) > 1 {
						handTroops := make([]card.Troop, 0, len(p.max[color].handTroops)-1)
						for _, troop := range copyCache.SrcHandTroops[playix] {
							if troop.Color() == outTroop.Color() {
								handTroops = append(handTroops, troop)
							}
						}
						copyCache.players[ix].max[color] = newData(p.max[color].deckTroops, handTroops)
					} else {
						copyCache.players[ix].max[color] = p.max[color]
					}
				}
			}
		} else {
			copyCache.players[ix] = p
		}
	}
	return copyCache
}

type player struct {
	set      map[card.Troop]bool //TODO could be faster with out map
	sortStrs [][]card.Troop
	max      [card.NOColors + 1]*data
}
type data struct {
	deckTroops     []card.Troop
	handTroops     []card.Troop
	troops         []card.Troop
	rolStr         []int
	rolHandNo      []int
	rolHandStr     []int
	rolHandOnlyStr []int
}

func newData(deckTroops, handTroops []card.Troop) (d *data) {
	d = new(data)
	d.deckTroops = deckTroops
	d.handTroops = handTroops
	if len(handTroops) > maxNo {
		d.handTroops = d.handTroops[:maxNo]
	}

	deckTrs := d.deckTroops
	handTrs := d.handTroops
	no := len(d.deckTroops) + len(d.handTroops)
	if no > maxNo {
		no = maxNo
	}
	d.troops = make([]card.Troop, no)
	d.rolStr = make([]int, no)
	d.rolHandNo = make([]int, no)
	d.rolHandStr = make([]int, no)
	for i := 0; i < no; i++ {
		prei := i - 1
		if i == 0 {
			prei = 0
		}
		if len(deckTrs) > 0 && len(handTrs) > 0 {
			if handTrs[0].Strenght() >= deckTrs[0].Strenght() {
				d.troops[i] = handTrs[0]
				d.rolStr[i] = d.rolStr[prei] + handTrs[0].Strenght()
				d.rolHandNo[i] = d.rolHandNo[prei] + 1
				d.rolHandStr[i] = d.rolHandStr[prei] + handTrs[0].Strenght()
				handTrs = handTrs[1:]
			} else {
				d.troops[i] = deckTrs[0]
				d.rolStr[i] = d.rolStr[prei] + deckTrs[0].Strenght()
				d.rolHandNo[i] = d.rolHandNo[prei]
				d.rolHandStr[i] = d.rolHandStr[prei]
				deckTrs = deckTrs[1:]
			}
		} else if len(handTrs) > 0 {
			d.troops[i] = handTrs[0]
			d.rolStr[i] = d.rolStr[prei] + handTrs[0].Strenght()
			d.rolHandNo[i] = d.rolHandNo[prei] + 1
			d.rolHandStr[i] = d.rolHandStr[prei] + handTrs[0].Strenght()
			handTrs = handTrs[1:]
		} else if len(deckTrs) > 0 {
			d.troops[i] = deckTrs[0]
			d.rolStr[i] = d.rolStr[prei] + deckTrs[0].Strenght()
			d.rolHandNo[i] = d.rolHandNo[prei]
			d.rolHandStr[i] = d.rolHandStr[prei]
			deckTrs = deckTrs[1:]
		}
	}
	d.rolHandOnlyStr = make([]int, len(d.handTroops))
	for i, troop := range d.handTroops {
		prei := i - 1
		if i == 0 {
			prei = 0
		}
		d.rolHandOnlyStr[i] = d.rolHandOnlyStr[prei] + troop.Strenght()
	}

	return d
}

func removeTroop(troops []card.Troop, out card.Troop) (res []card.Troop, ix int) {
	ix = -1
	res = make([]card.Troop, 0, len(troops)-1)
	for i, troop := range troops {
		if troop == out {
			ix = i
		} else {
			res = append(res, troop)
		}
	}
	return res, ix
}
