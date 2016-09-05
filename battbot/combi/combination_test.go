package combi

import (
	"github.com/rezder/go-battleline/battleline/cards"
	"testing"
)

func TestCombi3(t *testing.T) {
	//	for i, c := range CreateCombi(3) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := CreateCombi(3)
	t.Log(combis)
	n := len(combis)

	if n != 47 {
		t.Errorf("Expected 46 got: %v", n)
	}

}
func TestCombi4(t *testing.T) {
	//for i, c := range CreateCombi(4) {
	//fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := CreateCombi(4)
	n := len(combis)
	t.Log(combis)
	if n != 49 {
		t.Errorf("Expected 47 got: %v", n)
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

func TestAnaBattalion(t *testing.T) {
	combinations := []int{46, 45, 48, 43}
	dummies := []int{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 26
	combiNo4 := 30
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaSkirmish123(t *testing.T) {
	combinations := []int{1, 32, 43, 44}
	dummies := []int{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 46
	combiNo4 := 48
	combiMultiTest(t, combiNo3, combiNo4, combinations, dummies)
	//	t.Error("Forced error")
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
func combiTest(t *testing.T, combiNo int, flagCards, handCards, dummies []int, mud int) {
	combi := CreateCombi(mud)
	drawSet := createDrawSet()
	updateDraws(flagCards, handCards, dummies, drawSet)
	drawNo := (len(drawSet) - 7) / 2
	ana := Ana(combi[combiNo], flagCards, handCards, drawSet, drawNo, mud == 4)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagCards, handCards, drawNo, ana)
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
