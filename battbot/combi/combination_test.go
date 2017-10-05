package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/battleline/cards"
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
	combinations := []int{1, 2, 3, 4}
	dummies := []int{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 7
	combiNo4 := 6
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaPhalanx(t *testing.T) {
	combinations := []int{8, 18, 28, 38}
	dummies := []int{5, 6, 7, 17, 22, 33, 52, 21, 33}
	combiNo3 := 10
	combiNo4 := 9
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaPhalanxShortStack2(t *testing.T) {
	combi := Combinations(3)
	drawSet := make(map[int]bool)
	goodCards := []int{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	for _, troop := range goodCards {
		drawSet[troop] = true
	}
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagCards := []int{53, 33}
	handCards := []int{}
	ana := Ana(combi[15], flagCards, handCards, drawSet, drawNo, false)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagCards, handCards, drawNo, ana)
	if ana.Valid != 715 {
		t.Errorf("Valid combination should be 715 but is %v", ana.Valid)
	}
}
func TestAnaPhalanxShortStack(t *testing.T) {
	combi := Combinations(3)
	drawSet := make(map[int]bool)
	goodCards := []int{9, 1, 14, 37, 49, 38, 24}
	for _, troop := range goodCards {
		drawSet[troop] = true
	}
	drawNo := 7 + ((len(drawSet) - 7) / 2)
	flagCards := []int{34, 4}
	handCards := []int{}
	ana := Ana(combi[14], flagCards, handCards, drawSet, drawNo, false)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagCards, handCards, drawNo, ana)
	if ana.Valid != 1 {
		t.Errorf("Valid combination should be 1 but is %v", ana.Valid)
	}
}

func TestAnaBattalion(t *testing.T) {
	combinations := []int{46, 45, 48, 43}
	dummies := []int{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 26
	combiNo4 := 31
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaSkirmishSimple(t *testing.T) {
	combinations := []int{1, 32, 43, 44}
	dummies := []int{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	flagCards := combinations[:1]
	handCards := make([]int, 7)
	copy(handCards, dummies[:7])
	combiTest(t, combiNo3, flagCards, handCards, dummies[7:], 3)
	//t.Error("Forced error")
}
func TestAnaSkirmish123(t *testing.T) {
	combinations := []int{1, 32, 43, 44}
	dummies := []int{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	combiNo4 := 49
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaSkirmish123GoodOdds(t *testing.T) {
	//combinations := []int{1, 32, 43}
	combi := Combinations(3)
	drawSet := make(map[int]bool)
	goodCards := []int{2, 3, 12, 13, 22, 23, 32, 33, 42, 43, 52, 53, 27}
	for _, troop := range goodCards {
		drawSet[troop] = true
	}
	drawNo := (len(drawSet) - 7) / 2
	flagCards := []int{1}
	handCards := []int{5, 6, 7, 18, 25, 37, 58}
	ana := Ana(combi[46], flagCards, handCards, drawSet, drawNo, false)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[46], flagCards, handCards, drawNo, ana)
	//t.Error("Forced error")
}
func TestAnaBad(t *testing.T) {
	flagCards := []int{1, 67}
	handCards := []int{5, 6, 7, 18, 25, 37, 58}
	dummies := []int{29, 36}
	combiTest(t, 48, flagCards, handCards, dummies, 4)
	handCards[0] = 2
	combiTest(t, 48, flagCards, handCards, dummies, 4)
	//t.Error("Forced error")
}

func combiMultiTest(t *testing.T, combiNo3, combiNo4 int, combination, dummies []int) {
	flagCards := []int{combination[0]}
	handCards := make([]int, 7)
	copy(handCards, dummies[:7])
	combiTest(t, combiNo3, flagCards, handCards, dummies[7:], 3)
	combiTest(t, combiNo4, flagCards, handCards, dummies[7:], 4)

	prop := combiTest(t, combiNo3, combination[:3], handCards, dummies[7:], 3)
	if prop != 1 {
		t.Errorf("Combination: %v was not found", combination[:3])
	}
	prop = combiTest(t, combiNo4, combination[:4], handCards, dummies[7:], 4)
	if prop != 1 {
		t.Errorf("Combination: %v was not found", combination[:4])
	}
	for i := 1; i < 4; i++ {
		handCards[i-1] = combination[i]
		combiTest(t, combiNo3, flagCards, handCards, dummies[7:], 3)
		combiTest(t, combiNo4, flagCards, handCards, dummies[7:], 4)
	}
	handCards = make([]int, 7)
	copy(handCards, dummies[:7])
	jokers := []int{cards.TC123, cards.TCAlexander, cards.TCDarius, cards.TC8}
	flagCards = append(flagCards, 0)
	for _, j := range jokers {
		flagCards[1] = j
		combiTest(t, combiNo3, flagCards, handCards, dummies[7:], 3)
		combiTest(t, combiNo4, flagCards, handCards, dummies[7:], 4)
		for i := 1; i < 3; i++ {
			handCards[i-1] = combination[i]
			combiTest(t, combiNo3, flagCards, handCards, dummies[7:], 3)
			combiTest(t, combiNo4, flagCards, handCards, dummies[7:], 4)
		}
		handCards = make([]int, 7)
		copy(handCards, dummies[:7])

	}
}
func combiTest(t *testing.T, combiNo int, flagCards, handCards, dummies []int, formationSize int) float64 {
	combi := Combinations(formationSize)
	drawSet := createDrawSet()
	updateDraws(flagCards, handCards, dummies, drawSet)
	drawNo := (len(drawSet) - 7) / 2
	ana := Ana(combi[combiNo], flagCards, handCards, drawSet, drawNo, formationSize == 4)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagCards, handCards, drawNo, ana)
	fmt.Println(ana.Valid, ana.Prop, flagCards)
	return ana.Prop
}

func createDrawSet() map[int]bool {
	drawSet := make(map[int]bool)
	for i := 1; i <= cards.NOTroop; i++ {
		drawSet[i] = true
	}
	return drawSet
}
func updateDraws(flagCards, handCards, dummies []int, drawSet map[int]bool) {
	for _, v := range flagCards {
		delete(drawSet, v)
	}
	for _, v := range handCards {
		delete(drawSet, v)
	}

	for _, v := range dummies {
		delete(drawSet, v)
	}
}
