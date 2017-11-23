package combi

import (
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob/dht"
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

	if n != 29 {
		t.Errorf("Expected 29 got: %v", n)
	}

}
func TestCombi4(t *testing.T) {
	//for i, c := range createCombi(4) {
	//	fmt.Printf("rank: %v,combi: %v\n", i, *c)
	//}
	combis := createCombi(4)
	n := len(combis)
	t.Log(combis)
	if n != 27 {
		t.Errorf("Expected 27 got: %v", n)
	}
}
func TestAnaWedge(t *testing.T) {
	combinations := []card.Troop{1, 2, 3, 4}
	dummies := []card.Troop{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 7
	combiNo4 := 6
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetHostStr, targetBattStr, combinations, dummies)
	//t.Error("Forced error")
}

func TestAnaPhalanx(t *testing.T) {
	combinations := []card.Troop{8, 18, 28, 38}
	dummies := []card.Troop{5, 6, 7, 17, 22, 33, 52, 21, 33}
	combiNo3 := 10
	combiNo4 := 9
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetHostStr, targetBattStr, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaPhalanxDrawNo(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{51}
	dummies := make([]card.Troop, 0, 48)
	for i := 2; i <= 50; i++ {
		if i%10 != 1 {
			dummies = append(dummies, card.Troop(i))
		}
	}
	//dummies = append(dummies, []card.Troop{52}...)
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	prob := combiTestDrawNo(t, 17, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr, 1)
	if prob > 1 {
		t.Errorf("Prob: %v  bigger than 1", prob)
	}

}
func TestAnaPhalanxShortStack2(t *testing.T) {
	combi := Combinations(3)
	goodTroops := []card.Troop{56, 18, 43, 58, 52, 21, 57, 51, 59, 5, 22, 55, 54, 30}
	deckTroops := testSortTroops(goodTroops)
	targetHostStr := 0
	targetBattStr := 0
	targetRank := 1
	drawNo := 7 + ((len(deckTroops) - 7) / 2)
	drawNos := [2]int{drawNo, 0}
	playix := 0
	flagTroops := []card.Troop{53, 33}
	flagMorales := []card.Morale{}
	handTroops := [2][]card.Troop{[]card.Troop{}, []card.Troop{}}
	deckHandTroops := dht.NewCache(deckTroops, handTroops, drawNos)
	ana := Ana(combi[15], flagTroops, flagMorales, deckHandTroops, playix, 3, false, targetRank, targetHostStr, targetBattStr)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[15], flagTroops, handTroops, drawNo, ana)
	if ana.Valid != 715 {
		t.Errorf("Valid combination should be 715 but is %v", ana.Valid)
	}
}
func testSortTroops(troops []card.Troop) (sortTroops []card.Troop) {
	sortTroops = make([]card.Troop, 0, len(troops))
	for _, troop := range troops {
		sortTroops = troop.AppendStrSorted(sortTroops)
	}
	return sortTroops
}

func TestAnaPhalanxShortStack(t *testing.T) {
	combi := Combinations(3)
	goodTroops := []card.Troop{9, 1, 14, 37, 49, 38, 24}
	deckTroops := testSortTroops(goodTroops)
	targetHostStr := 0
	targetBattStr := 0
	targetRank := 1
	drawNo := 7 + ((len(deckTroops) - 7) / 2)
	drawNos := [2]int{drawNo, 0}
	playix := 0
	flagTroops := []card.Troop{34, 4}
	flagMorales := []card.Morale{}
	handTroops := [2][]card.Troop{[]card.Troop{}, []card.Troop{}}
	deckHandTroops := dht.NewCache(deckTroops, handTroops, drawNos)
	ana := Ana(combi[14], flagTroops, flagMorales, deckHandTroops, playix, 3, false, targetRank, targetHostStr, targetBattStr)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[14], flagTroops, handTroops, drawNo, ana)
	if ana.Valid != 1 {
		t.Errorf("Valid combination should be 1 but is %v", ana.Valid)
	}
}

func TestAnaBattalion(t *testing.T) {
	combinations := []card.Troop{46, 45, 48, 43}
	dummies := []card.Troop{5, 6, 7, 18, 22, 33, 52, 21, 33}
	combiNo3 := 19
	combiNo4 := 18
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetHostStr, targetBattStr, combinations, dummies)
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
	targetHostStr := 0
	targetBattStr := 0
	prob := combiTest(t, 19, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	ex := float64(0)
	if prob == ex {
		t.Error("Battalion probability failed expected bigger than zero")
	}
}
func TestAnaBattalionStr(t *testing.T) {
	flagTroops := []card.Troop{54}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{}
	dummies := make([]card.Troop, 0, 60)
	for i := 1; i <= 31; i++ {
		dummies = append(dummies, card.Troop(i))
	}
	dummies = append(dummies, card.Troop(60))
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 21
	prob := combiTest(t, 18, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	ex := float64(0)
	if prob == ex {
		t.Error("Battalion probability failed expected bigger than zero")
	}
}
func TestAnaSkirmishSimple(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 26
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	flagTroops := combinations[:1]
	flagMorales := []card.Morale{}
	handTroops := make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	//t.Error("Forced error")
}
func TestAnaSkirmish123(t *testing.T) {
	combinations := []card.Troop{1, 32, 43, 44}
	dummies := []card.Troop{5, 6, 7, 18, 25, 37, 58, 29, 36}
	combiNo3 := 28
	combiNo4 := 26
	targetRank := 1
	targetHostStr := 0
	targerBattStr := 0
	combiMultiTest(t, combiNo3, combiNo4, targetRank, targetHostStr, targerBattStr, combinations, dummies)
	//t.Error("Forced error")
}
func TestAnaSkirmish123GoodOdds(t *testing.T) {
	//combinations := []int{1, 32, 43}
	combi := Combinations(3)
	goodTroops := []card.Troop{2, 3, 12, 13, 22, 23, 32, 33, 42, 43, 52, 53, 27}
	deckTroops := testSortTroops(goodTroops)
	targetHostStr := 0
	targetBattStr := 0
	targetRank := 1
	drawNo := (len(deckTroops) - 7) / 2
	drawNos := [2]int{drawNo, 0}
	botix := 0
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	botHandTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	handTroops := [2][]card.Troop{botHandTroops, []card.Troop{}}
	deckHandTroops := dht.NewCache(deckTroops, handTroops, drawNos)
	ana := Ana(combi[26], flagTroops, flagMorales, deckHandTroops, botix, 3, false, targetRank, targetHostStr, targetBattStr)
	t.Logf("Combi: %+v\nFlag: %v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[26], flagTroops, handTroops, drawNo, ana)
	//t.Error("Forced error")
}
func TestAnaBad(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{card.TC123}
	handTroops := []card.Troop{5, 6, 7, 18, 25, 37, 58}
	dummies := []card.Troop{29, 36}
	targetRank := 1
	targetHostStr := 0
	targetBattStr := 0
	combiTest(t, 26, flagTroops, handTroops, dummies, flagMorales, 4, targetRank, targetHostStr, targetBattStr)
	handTroops[0] = card.Troop(2)
	combiTest(t, 26, flagTroops, handTroops, dummies, flagMorales, 4, targetRank, targetHostStr, targetBattStr)
	//t.Error("Forced error")
}
func TestAnaHostMade(t *testing.T) {
	flagTroops := []card.Troop{10}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{20, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := RankHost(3)
	targetHostStr := 21
	targetBattStr := 0
	prob := combiTest(t, RankHost(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	if prob != 1 {
		t.Errorf("Host probability failed exp:%v got %v", 1, prob)
	}

}
func TestAnaHostLost(t *testing.T) {
	flagTroops := []card.Troop{1}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{22, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := RankHost(3)
	targetHostStr := 22
	targerBattStr := 0
	prob := combiTest(t, RankHost(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targerBattStr)
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

	targetRank := RankHost(3)
	targetHostStr := 22
	targetBattStr := 0
	prob := combiTest(t, RankHost(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	ex := float64(0)
	if prob == ex {
		t.Error("Host probability failed expected positive probability")
	}
	dummies = append(dummies, 51)
	prob = combiTest(t, RankHost(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	if prob != ex {
		t.Errorf("Host probability failed expected %v probability got %v", ex, prob)
	}
}
func TestAnaHostProb(t *testing.T) {
	flagTroops := []card.Troop{10}
	flagMorales := []card.Morale{}
	handTroops := []card.Troop{22, 2, 1, 22, 23, 31, 33}
	dummies := []card.Troop{29, 36}
	targetRank := RankHost(3)
	targetHostStr := 29
	targetBattStr := 0
	prob := combiTest(t, RankHost(3)-1, flagTroops, handTroops, dummies, flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	ex := float64(0)
	if prob == ex {
		t.Error("Host probability failed expected bigger than zero")
	}
}

func combiMultiTest(t *testing.T, combiNo3, combiNo4, targetRank, targetHostStr, targetBattStr int, combination, dummies []card.Troop) {
	flagTroops := []card.Troop{combination[0]}
	handTroops := make([]card.Troop, 7)
	flagMorales := []card.Morale{}
	copy(handTroops, dummies[:7])
	combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4, targetRank, targetHostStr, targetBattStr)

	prop := combiTest(t, combiNo3, combination[:3], handTroops, dummies[7:], flagMorales, 3, targetRank, targetHostStr, targetBattStr)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:3], prop)
	}
	prop = combiTest(t, combiNo4, combination[:4], handTroops, dummies[7:], flagMorales, 4, targetRank, targetHostStr, targetBattStr)
	if prop != 1 {
		t.Errorf("Combination: %v was not found. Prop %v", combination[:4], prop)
	}
	for i := 1; i < 4; i++ {
		handTroops[i-1] = combination[i]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], flagMorales, 3, targetRank, targetHostStr, targetBattStr)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], flagMorales, 4, targetRank, targetHostStr, targetBattStr)
	}
	handTroops = make([]card.Troop, 7)
	copy(handTroops, dummies[:7])
	jokers := []card.Morale{card.TC123, card.TCAlexander, card.TCDarius, card.TC8}
	for i := range jokers {
		morales := jokers[i : i+1]
		combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3, targetRank, targetHostStr, targetBattStr)
		combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4, targetRank, targetHostStr, targetBattStr)
		for i := 1; i < 3; i++ {
			handTroops[i-1] = combination[i]
			combiTest(t, combiNo3, flagTroops, handTroops, dummies[7:], morales, 3, targetRank, targetHostStr, targetBattStr)
			combiTest(t, combiNo4, flagTroops, handTroops, dummies[7:], morales, 4, targetRank, targetHostStr, targetBattStr)
		}
		handTroops = make([]card.Troop, 7)
		copy(handTroops, dummies[:7])
	}
}
func combiTest(t *testing.T,
	combiNo int,
	flagTroops, botHandTroops, dummies []card.Troop,
	flagMorales []card.Morale,
	formationSize, targetRank, targetHostStr, targetBattStr int,
) float64 {

	combi := Combinations(formationSize)
	deckTroops, flagTroops, botHandTroops := updateDraws(flagTroops, botHandTroops, dummies)
	botix := 0
	drawNo := (len(deckTroops) - 7) / 2
	drawNos := [2]int{drawNo, 0}
	handTroops := [2][]card.Troop{botHandTroops, nil}
	deckHandTroops := dht.NewCache(deckTroops, handTroops, drawNos)
	ana := Ana(combi[combiNo], flagTroops, flagMorales, deckHandTroops, botix, formationSize, false, targetRank, targetHostStr, targetBattStr)
	t.Logf("Combi: %+v\nFlag: %v,%v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagTroops, flagMorales, handTroops, drawNo, ana)
	allCombi := math.Comb(uint64(len(deckTroops)), uint64(drawNo))
	ana.SetAll(allCombi)
	if ana.Prop > 1 {
		t.Errorf("Prob bigger than 1 got %v", ana.Prop)
	}
	return ana.Prop
}
func combiTestDrawNo(t *testing.T,
	combiNo int,
	flagTroops, botHandTroops, dummies []card.Troop,
	flagMorales []card.Morale,
	formationSize, targetRank, targetHostStr, targetBattStr, drawNo int,
) float64 {

	combi := Combinations(formationSize)
	deckTroops, flagTroops, botHandTroops := updateDraws(flagTroops, botHandTroops, dummies)
	botix := 0
	drawNos := [2]int{drawNo, 0}
	handTroops := [2][]card.Troop{botHandTroops, nil}
	deckHandTroops := dht.NewCache(deckTroops, handTroops, drawNos)
	fmt.Println(len(deckTroops), combi[combiNo].Formation.Name, combi[combiNo].Strength)
	ana := Ana(combi[combiNo], flagTroops, flagMorales, deckHandTroops, botix, formationSize, false, targetRank, targetHostStr, targetBattStr)
	t.Logf("Combi: %+v\nFlag: %v,%v\nHand: %v\nDraws: %v\nResult: %v\n",
		*combi[combiNo], flagTroops, flagMorales, handTroops, drawNo, ana)
	allCombi := math.Comb(uint64(len(deckTroops)), uint64(drawNo))
	ana.SetAll(allCombi)
	if ana.Prop > 1 {
		t.Errorf("Prob bigger than 1 got %v", ana.Prop)
	}
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
) (sortDeckTroops, sortFlagTroops, sortHandTroops []card.Troop) {
	drawSet := createDrawSet()
	for _, t := range flagTroops {
		delete(drawSet, t)
	}
	for _, t := range handTroops {
		delete(drawSet, t)
	}
	for _, t := range dummies {
		delete(drawSet, t)
	}
	sortDeckTroops = make([]card.Troop, 0, len(drawSet))
	for troop := range drawSet {
		sortDeckTroops = troop.AppendStrSorted(sortDeckTroops)
	}

	for _, troop := range flagTroops {
		sortFlagTroops = troop.AppendStrSorted(sortFlagTroops)
	}
	for _, troop := range handTroops {
		sortHandTroops = troop.AppendStrSorted(sortHandTroops)
	}
	return sortDeckTroops, sortFlagTroops, sortHandTroops
}
