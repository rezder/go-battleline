package flag

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
	"github.com/rezder/go-battleline/v2/game/card"
	"testing"
)

func TestWedgeLeader(t *testing.T) {
	flagMorales := []card.Morale{69}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 9, 22, 1, 11}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 1
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestWedge8(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 9, 22, 1, 11}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 1
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestWedge123(t *testing.T) {
	flagMorales := []card.Morale{67}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 8, 9, 22, 1, 3, 11}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 8
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func testCreateDHT(deckTroops, botHandTroops []card.Troop, botDrawNo int) (dhTroops *dht.Cache, botix int) {
	drawNos := [2]int{botDrawNo, 0}
	handTroops := [2][]card.Troop{botHandTroops, nil}
	sortDeckTroops := make([]card.Troop, 0, len(deckTroops))
	for _, troop := range deckTroops {
		sortDeckTroops = troop.AppendStrSorted(sortDeckTroops)
	}
	dhTroops = dht.NewCache(sortDeckTroops, handTroops, drawNos)
	return dhTroops, botix
}
func TestPhalanx(t *testing.T) {
	flagTroops := []card.Troop{53, 33}
	flagMorales := []card.Morale{}
	botHandTroops := []card.Troop{}
	deckTroops := []card.Troop{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	formationSize := 3
	expRank := 16
	drawNo := 10
	targetSum := 0
	targetRank := 1
	isFog := false
	dhTroops, botix := testCreateDHT(deckTroops, botHandTroops, drawNo)
	rank := CalcMaxRank(flagTroops, flagMorales, dhTroops, botix, formationSize, isFog, targetRank, targetSum)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[expRank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanxNewFlag(t *testing.T) {
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 41, 52, 59, 49}
	dhTroops, botix := testCreateDHT(deck, handTroops, 7)
	formationSize := 3
	isFog := false
	expRank := 10
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanx123(t *testing.T) {
	flagMorales := []card.Morale{67}
	handTroops := []card.Troop{}
	deck := []card.Troop{4, 10, 31, 41, 52}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 18
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalion8(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{4, 10, 31, 41, 52}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 24
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalion123(t *testing.T) {
	flagMorales := []card.Morale{67}
	handTroops := []card.Troop{}
	deck := []card.Troop{9, 10, 8, 27, 21, 52}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 24
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalion3Morale(t *testing.T) {
	flagMorales := []card.Morale{69, 68, 67}
	handTroops := []card.Troop{}
	deck := []card.Troop{4, 10, 9, 8, 7, 31, 41, 52}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 25
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
	formationSize = 4
	expRank = 23
	rank = calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestLine8(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 41, 52, 23}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 40
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestLine123(t *testing.T) {
	flagMorales := []card.Morale{67}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 45, 52, 23}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 47
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalion1238(t *testing.T) {
	flagMorales := []card.Morale{67, 68}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 45, 52, 23}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 25
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalionL8(t *testing.T) {
	flagMorales := []card.Morale{69, 68}
	handTroops := []card.Troop{}
	deck := []card.Troop{15, 1, 31, 45, 52, 23}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 23
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanxL8(t *testing.T) {
	flagMorales := []card.Morale{69, 68}
	handTroops := []card.Troop{}
	deck := []card.Troop{18, 1, 31, 45, 52, 23}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 3
	isFog := false
	expRank := 11
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanxL123(t *testing.T) {
	flagMorales := []card.Morale{69, 67}
	handTroops := []card.Troop{}
	deck := []card.Troop{18, 1, 31, 48, 57, 26}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 4
	isFog := false
	expRank := 17
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestLineL8(t *testing.T) {
	flagMorales := []card.Morale{69, 68}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 12, 31, 45, 52, 27}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 4
	isFog := false
	expRank := 44
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	fmt.Println(rank, expRank)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestWedgeL123(t *testing.T) {
	flagMorales := []card.Morale{69, 67}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 12, 31, 45, 52, 54, 27}
	dhTroops, botix := testCreateDHT(deck, handTroops, 5)
	formationSize := 4
	isFog := false
	expRank := 6
	rank := calcMaxRankNewFlag(flagMorales, dhTroops, botix, formationSize, isFog)
	fmt.Println(rank, expRank)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func removeFromDeck(troops []card.Troop, deckSet map[card.Troop]bool) {
	for _, troop := range troops {
		delete(deckSet, troop)
	}
}
