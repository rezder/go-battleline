package dht

import (
	"github.com/rezder/go-battleline/v2/game/card"
	"testing"
)

func TestDHT(t *testing.T) {
	deckTroops := []card.Troop{10, 20, 19, 18, 6, 1, 11, 21, 31, 41, 51}
	handTroops := [2][]card.Troop{[]card.Troop{9, 5, 35, 33, 12}, []card.Troop{16}}
	cache := NewCache(deckTroops, handTroops, [2]int{7, 7})
	onlySet := cache.OnlyDeckSet()
	if len(onlySet) != len(deckTroops) {
		t.Error("OnlyDeckSet is incomplete")
	}
	strs := cache.SortStrs(0)
	if len(strs[10]) != 2 {
		t.Error("Sorted after strengh failed")
	}
	color := 0
	playix := 0
	sum, isOk := cache.Sum(playix, color, 4)
	if !isOk {
		t.Error("Sum did not make a sum")
	} else {
		exp := 38
		if sum != exp {
			t.Errorf("Sum failed exepected: %v got %v", exp, sum)
		}
	}
	sum, isOk = cache.Sum(0, 0, 3)
	if !isOk {
		t.Error("Sum did not make a sum")
	} else {
		exp := 29
		if sum != exp {
			t.Errorf("Sum failed exepected: %v got %v", exp, sum)
		}
	}

	targetRes := cache.TargetSum(playix, color, 3, 29)
	expTarget := &TargetResult{
		IsPossibel:      true,
		IsMade:          false,
		NewSum:          20,
		NewNo:           2,
		ValidDeckTroops: []card.Troop{10, 20},
		ValidHandTroops: []card.Troop{},
	}
	if !expTarget.Equal(targetRes) {
		t.Errorf("Target sum failed expected: %v  got%v", expTarget, targetRes)
	}
	color = 1
	targetRes = cache.TargetSum(playix, color, 2, 15)
	expTarget = &TargetResult{
		IsPossibel:      true,
		IsMade:          false,
		NewSum:          6,
		NewNo:           1,
		ValidDeckTroops: []card.Troop{10, 6},
		ValidHandTroops: []card.Troop{},
	}
	if !expTarget.Equal(targetRes) {
		t.Errorf("Target sum failed expected: %v  got%v", expTarget, targetRes)
	}
	color = 0
	playix = 1
	targetRes = cache.TargetSum(playix, color, 3, 25)
	expTarget = &TargetResult{
		IsPossibel:      true,
		IsMade:          false,
		NewSum:          25,
		NewNo:           3,
		ValidDeckTroops: []card.Troop{10, 20, 19, 18, 6},
		ValidHandTroops: []card.Troop{16},
	}
	if !expTarget.Equal(targetRes) {
		t.Errorf("Target sum failed expected: %v  got%v", expTarget, targetRes)
	}
	if len(cache.OnlyDeckSet()) != len(deckTroops) {
		t.Error("only deck set failed")
	}
	testSet(deckTroops, handTroops, cache, t)
	testSortStrs(deckTroops, handTroops, cache, t)
	outixs := [2]int{2, 0}
	testCopyWithOut(outixs, deckTroops, handTroops, cache, t)
	outixs = [2]int{3, 0}
	testCopyWithOut(outixs, deckTroops, handTroops, cache, t)
}
func outColorix(handTroops []card.Troop, outix int) (outColorix int) {
	outColorix = -1
	for _, troop := range handTroops {
		if troop.Color() == handTroops[outix].Color() {
			outColorix = outColorix + 1
		}
		if troop == handTroops[outix] {
			break
		}
	}
	return outColorix
}
func testCopyWithOut(outixs [2]int, deckTroops []card.Troop, handTroops [2][]card.Troop, cache *Cache, t *testing.T) {
	for p, handTrs := range handTroops {
		outTroop := handTrs[outixs[p]]
		outix := outColorix(handTrs, outixs[p])
		var outHandTroops [2][]card.Troop
		outHandTroops[p], _ = removeTroop(handTrs, outTroop)
		outHandTroops[opp(p)] = handTroops[opp(p)]
		copyCache := cache.CopyWithOutHand(outTroop, p)
		testSet(deckTroops, outHandTroops, copyCache, t)
		testSortStrs(deckTroops, outHandTroops, copyCache, t)
		copyData := copyCache.players[p].max[outTroop.Color()]
		data := cache.players[p].max[outTroop.Color()]
		if copyData != nil {
			for i, s := range copyData.rolHandOnlyStr {
				if i < outix {
					if data.rolHandOnlyStr[i] != s {
						t.Errorf("CopyWithOutHand falied sum should be the same. Old: %v New: %v,outix: %v", data, copyData, outix)
					}
				} else {
					if data.rolHandOnlyStr[i] == s {
						t.Errorf("CopyWithOutHand falied sum should be difference. Old: %v New: %v", data, copyData)
					}
				}
			}
		} else {
			if data != nil && len(data.rolHandOnlyStr) > 1 {
				t.Errorf("CopyWithOutHand falied for color %v, old data: %v, new data:%v", card.COLNone, data, copyData)
			}
		}
	}

}
func opp(pix int) int {
	ix := pix + 1
	if ix > 1 {
		ix = 0
	}
	return ix
}

func testSet(deckTroops []card.Troop, handTroops [2][]card.Troop, cache *Cache, t *testing.T) {
	for p, handTrs := range handTroops {
		expNo := len(deckTroops) + len(handTrs)
		gotNo := len(cache.Set(p))
		if expNo != gotNo {
			t.Errorf("Deck hand troop set failed exp %v got %v set: %v ", expNo, gotNo, cache.Set(p))
		}
	}
}
func testSortStrs(deckTroops []card.Troop, handTroops [2][]card.Troop, cache *Cache, t *testing.T) {
	for p, handTrs := range handTroops {
		no := 0
		for str, troops := range cache.SortStrs(p) {
			no = no + len(troops)
			t.Logf("Strenght: %v,troops: %v", str, troops)
		}
		expNo := len(handTrs) + len(deckTroops)
		if expNo != no {
			t.Errorf("Sort after strenght failed exp %v got %v set: %v ", expNo, no, cache.SortStrs(p))
		}
	}
}
