package flag

import (
	"github.com/rezder/go-battleline/v2/bot/prob/combi"
	"github.com/rezder/go-battleline/v2/game/card"
	"testing"
)

func TestWedgeOneMorale(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{10, 9, 22, 1, 11}
	formationSize := 3
	expRank := 1
	rank := calcMaxRankNewFlag(flagMorales, handTroops, createSetFrom(deck), formationSize)
	if rank != expRank {
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanx(t *testing.T) {
	flagTroops := []card.Troop{53, 33}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	deck := []card.Troop{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	formationSize := 3
	expRank := 16
	rank := CalcMaxRank(flagTroops, flagMorales, handTroops, createSetFrom(deck), 10, 4)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[expRank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestPhalanxNewFlag(t *testing.T) {
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 41, 52, 59, 49}
	formationSize := 3
	expRank := 10
	rank := calcMaxRankNewFlag(flagMorales, handTroops, createSetFrom(deck), formationSize)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalionOneMorale(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{4, 10, 31, 41, 52}
	formationSize := 3
	expRank := 24
	rank := calcMaxRankNewFlag(flagMorales, handTroops, createSetFrom(deck), formationSize)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestBattalion3Morale(t *testing.T) {
	flagMorales := []card.Morale{69, 68, 67}
	handTroops := []card.Troop{}
	deck := []card.Troop{4, 10, 31, 41, 52}
	formationSize := 3
	expRank := 25
	rank := calcMaxRankNewFlag(flagMorales, handTroops, createSetFrom(deck), formationSize)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func TestLineOneMorale(t *testing.T) {
	flagMorales := []card.Morale{68}
	handTroops := []card.Troop{}
	deck := []card.Troop{19, 10, 31, 41, 52}
	formationSize := 3
	expRank := 40
	rank := calcMaxRankNewFlag(flagMorales, handTroops, createSetFrom(deck), formationSize)
	if rank != expRank {
		t.Logf("Combination %v", combi.Combinations(formationSize)[rank-1])
		t.Errorf("Expect rank %v got %v\n", expRank, rank)
	}
}
func createSetFrom(troops []card.Troop) map[card.Troop]bool {
	drawSet := make(map[card.Troop]bool)
	for _, troop := range troops {
		drawSet[troop] = true
	}
	return drawSet
}
func createDrawSet() map[card.Troop]bool {
	drawSet := make(map[card.Troop]bool)
	for i := 1; i <= card.NOTroop; i++ {
		drawSet[card.Troop(i)] = true
	}
	return drawSet
}
func removeFromDeck(troops []card.Troop, deckSet map[card.Troop]bool) {
	for _, troop := range troops {
		delete(deckSet, troop)
	}
}
