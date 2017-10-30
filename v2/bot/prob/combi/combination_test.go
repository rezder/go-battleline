package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/game/card"
	math "github.com/rezder/go-math/int"
	"testing"
)

func TestCombi3(t *testing.T) {
	_ = fmt.Sprintln("Dummy")
	//	for i, c := range CreateCombi(3) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := createCombi(3)
	t.Log(combis)
	n := len(combis)

	if n != 48 {
		t.Errorf("Expected 48 got: %v", n)
	}

}
func TestCombi4(t *testing.T) {
	//for i, c := range createCombi(4) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := createCombi(4)
	n := len(combis)
	t.Log(combis)
	if n != 51 {
		t.Errorf("Expected 51 got: %v", n)
	}
}
func TestAnaWedge(t *testing.T) {
	combinations := []card.Troop{1, 2, 3, 4}
	dummies := []card.Troop{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 7
	combiNo4 := 6
	targetRank := 1
	targetSum := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetSum, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaPhalanx(t *testing.T) {
	combinations := []card.Troop{8, 18, 28, 38}
	dummies := []card.Troop{5, 6, 7, 17, 22, 33, 52, 21, 33}
	combiNo3 := 10
	combiNo4 := 9
	targetRank := 1
	targetSum := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetSum, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaPhalanxShortStack2(t *testing.T) {
	combi := Combinations(3)

	goodTroops := []card.Troop{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	drawSet, deckMaxStrs := testCreateDrawSet(goodTroops)
	targetSum := 0
	targetRank := 1
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagTroops := []card.Troop{53, 33}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	ana := Ana(combi[15], flagTroops, flagMorales, handTroops, drawSet, deckMaxStrs, drawNo, 3, false, targetRank, targetSum)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagTroops, handTroops, drawNo, ana)
	if ana.Valid != 715 {
		t.Errorf("Valid combination should be 715 but is %v", ana.Valid)
	}
}
func testCreateDrawSet(troops []card.Troop) (drawSet map[card.Troop]bool, deckMaxStrs []int) {
	drawSet = make(map[card.Troop]bool)
	sortTroops := make([]card.Troop, 0, len(troops))
	for _, troop := range troops {
		drawSet[troop] = true
		troop.AppendStrSorted(sortTroops)
	}
	deckMaxStrs = make([]int, 0, 4)
	for _, troop := range sortTroops {
		deckMaxStrs = append(deckMaxStrs, troop.Strenght())
	}
	return drawSet, deckMaxStrs
}

func TestAnaPhalanxShortStack(t *testing.T) {
	combi := Combinations(3)
	goodTroops := []card.Troop{9, 1, 14, 37, 49, 38, 24}
	drawSet, deckMaxStrs := testCreateDrawSet(goodTroops)
	targetSum := 0
	targetRank := 1
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagTroops := []card.Troop{34, 4}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	ana := Ana(combi[14], flagTroops, flagMorales, handTroops, drawSet, deckMaxStrs, drawNo, 3, false, targetRank, targetSum)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagTroops, handTroops, drawNo, ana)
	if ana.Valid != 1 {
		t.Errorf("Valid combination should be 1 but is %v", ana.Valid)
	}
}

func TestAnaBattalion(t *testing.T) {
	combinations := []card.Troop{46, 45, 48, 43}
	dummies := []card.Troop{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 26
	combiNo4 := 31
	targetRank := 1
	targetSum := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetSum, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaBattalionGoodOdds(t *testing.T) {
	flagTroops := []card.Troop{57}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	dummies := make([]card.Troop, 0, 60)
	for i := 1; i <= 47; i++ {
		dummies = append(dummies, card.Troop(i))
	}
	targetRank := 1
	targetSum := 0
	prob := combiTest(t, 27, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	ex := float64(0)
	if prob == ex {
		t.Error("Host probability failed expected bigger than zero")
	}
}

func TestAnaSkirmishSimple(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	targetRank := 1
	targetSum := 0
	flagTroops := combinations[:1]
	flagMorales := []card.Morale{}
	handTroops := make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetSum)
	//t.Error("Forced error")
}
func TestAnaSkirmish123(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	combiNo4 := 49
	targetRank := 1
	targetSum := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetSum, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaSkirmish123GoodOdds(t *testing.T) {
	//combinations := []int{1, 32, 43}
	combi := Combinations(3)
	goodTroops := []card.Troop{2, 3, 12, 13, 22, 23, 32, 33, 42, 43, 52, 53, 27}
	drawSet, deckMaxStrs := testCreateDrawSet(goodTroops)
	targetSum := 0
	targetRank := 1
	drawNo := (len(drawSet) - 7) / 2
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	ana := Ana(combi[46], flagTroops, flagMorales, handTroops, drawSet, deckMaxStrs, drawNo, 3, false, targetRank, targetSum)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[46], flagTroops, handTroops, drawNo, ana)
	//t.Error("Forced error")
}
func TestAnaBad(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{card.TC123}
	handTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	dummies := []card.Troop{29, 36}
	targetRank := 1
	targetSum := 0
	combiTest(t, 48, flagTroops, handTroops, dummies, flagMorales, 4, targetRank, targetSum)
	handTroops[0] = card.Troop(2)
	combiTest(t, 48, flagTroops, handTroops, dummies, flagMorales, 4, targetRank, targetSum)
	//t.Error("Forced error")
}
func TestAnaHostMade(t *testing.T) {
	flagTroops := []card.Troop{10}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{20, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := HostRank(3)
	targetSum := 21
	prob := combiTest(t, HostRank(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	if prob != 1 {
		t.Errorf("Host probability failed exp:%v got %v", 1, prob)
	}

}
func TestAnaHostLost(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{22, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := HostRank(3)
	targetSum := 22
	prob := combiTest(t, HostRank(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	ex := float64(0)
	if prob != ex {
		t.Errorf("Host probability failed exp:%v got %v", ex, prob)
	}
}
func TestAnaHostLostShortStack(t *testing.T) {
	flagTroops := []card.Troop{2}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{22, 2, 1, 22, 23, 31, 33}
	dummies := make([]card.Troop, 0, 60)
	for i := 1; i <= 49; i++ {
		dummies = append(dummies, card.Troop(i))
	}

	targetRank := HostRank(3)
	targetSum := 22
	prob := combiTest(t, HostRank(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	ex := float64(0)
	if prob == ex {
		t.Error("Host probability failed expected positive probability")
	}
	dummies = append(dummies, 51)
	prob = combiTest(t, HostRank(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	if prob != ex {
		t.Errorf("Host probability failed expected %v probability got %v", ex, prob)
	}
}
func TestAnaHostProb(t *testing.T) {
	flagTroops := []card.Troop{10}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{22, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := HostRank(3)
	targetSum := 29
	prob := combiTest(t, HostRank(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetSum)
	ex := float64(0)
	if prob == ex {
		t.Error("Host probability failed expected bigger than zero")
	}
}

func combiMultiTest(t *testing.T, combiNo3, combiNo4, targetRank, targetSum int, combination, dummies []card.Troop) {
	flagTroops := []card.Troop{combination[0]}
	handTroops := make([]card.Troop, 7)
	flagMorales := []card.Morale{}
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetSum)
	combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4, targetRank, targetSum)

	prop := combiTest(t, combiNo3, combination[:3], handTroops, dummies[7:], flagMorales, 3, targetRank, targetSum)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:3], prop)
	}
	prop = combiTest(t, combiNo4, combination[:4], handTroops, dummies[7:], flagMorales, 4, targetRank, targetSum)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:4], prop)
	}
	for i := 1; i < 4; i++ {
		handTroops[i-1] = combination[i]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetSum)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4, targetRank, targetSum)
	}
	handTroops = make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	jokers := []card.Morale{card.TC123, card.TCAlexander, card.TCDarius, card.TC8}
	for i := range jokers {
		morales := jokers[i : i+1]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3, targetRank, targetSum)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4, targetRank, targetSum)
		for i := 1; i < 3; i++ {
			handTroops[i-1] = combination[i]
			combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3, targetRank, targetSum)
			combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4, targetRank, targetSum)
		}
		handTroops = make([]card.Troop, 7)
		copy(handTroops, dummies[:7])
	}
}
func combiTest(t *testing.T,
	combiNo int,
	flagTroops, handTroops, dummies []card.Troop,
	flagMorales []card.Morale,
	formationSize, targetRank, targetSum int,
) float64 {

	combi := Combinations(formationSize)
	drawSet := createDrawSet()
	deckMaxStrs, flagTroops, handTroops := updateDraws(flagTroops, handTroops, dummies, drawSet)
	drawNo := (len(drawSet) - 7) / 2
	ana := Ana(combi[combiNo], flagTroops, flagMorales, handTroops, drawSet, deckMaxStrs, drawNo, formationSize, false, targetRank, targetSum)
	t.Logf("Combi: %+v\nFlag: %v,%v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagTroops, flagMorales, handTroops, drawNo, ana)
	allCombi := math.Comb(uint64(len(drawSet)), uint64(drawNo))
	ana.SetAll(allCombi)
	return ana.Prop
}

func createDrawSet() map[card.Troop]bool {
	drawSet := make(map[card.Troop]bool)
	for i := 1; i <= card.NOTroop; i++ {
		drawSet[card.Troop(i)] = true
	}
	return drawSet
}
func updateDraws(
	flagTroops, handTroops, dummies []card.Troop,
	drawSet map[card.Troop]bool,
) (deckMaxStrs []int, sortFlagTroops, sortHandTroops []card.Troop) {
	for _, t := range flagTroops {
		delete(drawSet, t)
	}
	for _, t := range handTroops {
		delete(drawSet, t)
	}

	for _, t := range dummies {
		delete(drawSet, t)
	}
	sortTroops := make([]card.Troop, 0, len(drawSet))
	for troop := range drawSet {
		sortTroops = troop.AppendStrSorted(sortTroops)
	}
	deckMaxStrs = make([]int, 0, 4)
	for _, troop := range sortTroops {
		deckMaxStrs = append(deckMaxStrs, troop.Strenght())
		if len(deckMaxStrs) == cap(deckMaxStrs) {
			break
		}
	}
	for _, troop := range flagTroops {
		sortFlagTroops = troop.AppendStrSorted(sortFlagTroops)
	}
	for _, troop := range handTroops {
		sortHandTroops = troop.AppendStrSorted(sortHandTroops)
	}
	return deckMaxStrs, sortFlagTroops, sortHandTroops
}
