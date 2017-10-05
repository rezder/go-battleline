package combi

import (
	"github.com/rezder/go-battleline/v2/game/card"
	"testing"
)

func TestCombi3(t *testing.T) {
	//	for i, c := range CreateCombi(3) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := createCombi(3)
	t.Log(combis)
	n := len(combis)

	if n != 47 {
		t.Errorf("Expected 47 got: %v", n)
	}

}
func TestCombi4(t *testing.T) {
	//for i, c := range createCombi(4) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := createCombi(4)
	n := len(combis)
	t.Log(combis)
	if n != 50 {
		t.Errorf("Expected 50 got: %v", n)
	}
}
func TestAnaWedge(t *testing.T) {
	combinations := []card.Troop{1, 2, 3, 4}
	dummies := []card.Troop{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 7
	combiNo4 := 6
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaPhalanx(t *testing.T) {
	combinations := []card.Troop{8, 18, 28, 38}
	dummies := []card.Troop{5, 6, 7, 17, 22, 33, 52, 21, 33}
	combiNo3 := 10
	combiNo4 := 9
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaPhalanxShortStack2(t *testing.T) {
	combi := Combinations(3)
	drawSet := make(map[card.Troop]bool)
	goodTroops := []card.Troop{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	for _, troop := range goodTroops {
		drawSet[troop] = true
	}
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagTroops := []card.Troop{53, 33}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	ana := Ana(combi[15], flagTroops, flagMorales, handTroops, drawSet, drawNo, false)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagTroops, handTroops, drawNo, ana)
	if ana.Valid != 715 {
		t.Errorf("Valid combination should be 715 but is %v", ana.Valid)
	}
}
func TestAnaPhalanxShortStack(t *testing.T) {
	combi := Combinations(3)
	drawSet := make(map[card.Troop]bool)
	goodTroops := []card.Troop{9, 1, 14, 37, 49, 38, 24}
	for _, troop := range goodTroops {
		drawSet[troop] = true
	}
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagTroops := []card.Troop{34, 4}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	ana := Ana(combi[14], flagTroops, flagMorales, handTroops, drawSet, drawNo, false)
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
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaSkirmishSimple(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	flagTroops := combinations[:1]
	flagMorales := []card.Morale{}
	handTroops := make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3)
	//t.Error("Forced error")
}
func TestAnaSkirmish123(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	combiNo4 := 49
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaSkirmish123GoodOdds(t *testing.T) {
	//combinations := []int{1, 32, 43}
	combi := Combinations(3)
	drawSet := make(map[card.Troop]bool)
	goodTroops := []card.Troop{2, 3, 12, 13, 22, 23, 32, 33, 42, 43, 52, 53, 27}
	for _, troop := range goodTroops {
		drawSet[troop] = true
	}
	drawNo := (len(drawSet) - 7) / 2
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	ana := Ana(combi[46], flagTroops, flagMorales, handTroops, drawSet, drawNo, false)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[46], flagTroops, handTroops, drawNo, ana)
	//t.Error("Forced error")
}
func TestAnaBad(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{card.TC123}
	handTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	dummies := []card.Troop{29, 36}
	combiTest(t, 48, flagTroops, handTroops, dummies, flagMorales, 4)
	handTroops[0] = card.Troop(2)
	combiTest(t, 48, flagTroops, handTroops, dummies, flagMorales, 4)
	//t.Error("Forced error")
}

func combiMultiTest(t *testing.T, combiNo3, combiNo4 int, combination, dummies []card.Troop) {
	flagTroops := []card.Troop{combination[0]}
	handTroops := make([]card.Troop, 7)
	flagMorales := []card.Morale{}
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3)
	combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4)

	prop := combiTest(t, combiNo3, combination[:3], handTroops, dummies[7:], flagMorales, 3)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:3], prop)
	}
	prop = combiTest(t, combiNo4, combination[:4], handTroops, dummies[7:], flagMorales, 4)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:4], prop)
	}
	for i := 1; i < 4; i++ {
		handTroops[i-1] = combination[i]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4)
	}
	handTroops = make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	jokers := []card.Morale{card.TC123, card.TCAlexander, card.TCDarius, card.TC8}
	for i := range jokers {
		morales := jokers[i : i+1]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4)
		for i := 1; i < 3; i++ {
			handTroops[i-1] = combination[i]
			combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3)
			combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4)
		}
		handTroops = make([]card.Troop, 7)
		copy(handTroops, dummies[:7])
	}
}
func combiTest(t *testing.T,
	combiNo int,
	flagTroops, handTroops, dummies []card.Troop,
	flagMorales []card.Morale,
	formationSize int) float64 {

	combi := Combinations(formationSize)
	drawSet := createDrawSet()
	updateDraws(flagTroops, handTroops, dummies, drawSet)
	drawNo := (len(drawSet) - 7) / 2
	ana := Ana(combi[combiNo], flagTroops, flagMorales, handTroops, drawSet, drawNo, formationSize == 4)
	t.Logf("Combi: %+v\nFlag: %v,%v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagTroops, flagMorales, handTroops, drawNo, ana)
	return ana.Prop
}

func createDrawSet() map[card.Troop]bool {
	drawSet := make(map[card.Troop]bool)
	for i := 1; i <= card.NOTroop; i++ {
		drawSet[card.Troop(i)] = true
	}
	return drawSet
}
func updateDraws(flagTroops, handTroops, dummies []card.Troop, drawSet map[card.Troop]bool) {
	for _, t := range flagTroops {
		delete(drawSet, t)
	}
	for _, t := range handTroops {
		delete(drawSet, t)
	}

	for _, t := range dummies {
		delete(drawSet, t)
	}
}
