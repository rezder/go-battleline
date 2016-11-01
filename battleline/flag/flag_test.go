package flag

import (
	//"errors"
	//"fmt"
	"bytes"
	"encoding/gob"
	"github.com/rezder/go-battleline/battleline/cards"
	"testing"
)

//TestFlagT1LeaderWedge testing wedge with one leader
func TestFlagT1LeaderWedge(t *testing.T) {
	flag := flagLeader(t, []int{2, 3}, []int{11, 13}, 9, 6, &cards.FWedge)

	//----------Buttom
	player1 := 0
	player2 := 1
	mud1, mud2, err := flag.Remove(2, player1)
	if mud1 != -1 || mud2 != -1 {
		t.Errorf("No exces card should be removed do to mud. Card %v,%v was moved", mud1, mud2)
	}
	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	mud1, mud2, err = flag.Remove(3, player1)
	if mud1 != -1 || mud2 != -1 {
		t.Errorf("No exces card should be removed do to mud. Card %v,%v was moved", mud1, mud2)
	}
	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	flag = flagLeader(t, []int{9, 10}, []int{11, 13}, 27, 6, &cards.FWedge)
	flag = flagLeader(t, []int{9, 10}, []int{11, 13, cards.TCFog}, 29, 14, &cards.FHost)

	//===========Mud
	flag = New()
	//-----------Top
	flag = flagLeader(t, []int{cards.TCMud, 1, 2, 3}, []int{20, 19, 18}, 10, 34, &cards.FWedge)

	mud1, mud2, err = flag.Remove(cards.TCMud, player1)
	ex := 1
	if mud1 != ex {
		t.Errorf("Expected mud 1 index: %v got: %v", ex, mud1)
	}
	ex = cards.TCDarius // this can only happen if you did not claim the flag as you hadd the best Formation
	if mud2 != ex {
		t.Errorf("Expected mud 2 index: %v got: %v", ex, mud2)
	}

	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.FWedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 9
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	ex = 27
	if flag.Players[player2].Formation != &cards.FWedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	//-----------Middel Top
	flagLeader(t, []int{cards.TCMud, 1, 2, 4}, []int{20, 19, 17}, 10, 34, &cards.FWedge)

	//-----------Miss two step
	flagLeader(t, []int{cards.TCMud, 1, 3, 5}, []int{20, 16, 17}, 19, 33, &cards.FBattalion)

}

// TestFlagT1NWedge testing wedge with one number joker.
// Player 1 have 8 player 2 have 123.
func TestFlagT1NWedge(t *testing.T) {
	//------Top
	flag8x123(t, []int{7, 6}, []int{11, 12}, 21, 6, &cards.FWedge)

	//----------Buttom
	flag := flag8x123(t, []int{10, 9}, []int{12, 13}, 27, 6, &cards.FWedge)

	//----------Fog
	flag8x123(t, []int{10, 9, cards.TCFog}, []int{12, 13}, 27, 8, &cards.FHost)
	//----------Middel
	flag8x123(t, []int{9, 7}, []int{11, 13}, 24, 6, &cards.FWedge)

	//-----------Top Mud
	flag = flag8x123(t, []int{cards.TCMud, 5, 7, 6}, []int{11, 12, 14}, 26, 10, &cards.FWedge)
	player1 := 0
	player2 := 1
	t.Logf("Remove mud 8 and 123")
	mud1, mud2, err := flag.Remove(cards.TCMud, player1)
	ex := 5
	if mud1 != ex {
		t.Errorf("Expected mud 1 index: %v got: %v", ex, mud1)
	}
	ex = 11
	if mud2 != ex {
		t.Errorf("Expected mud 2 index: %v got: %v", ex, mud2)
	}
	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.FWedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 21
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	ex = 9
	if flag.Players[player2].Formation != &cards.FWedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	//-----------Middel Top
	flag8x123(t, []int{cards.TCMud, 6, 7, 9}, []int{11, 12, 14}, 30, 10, &cards.FWedge)

	//-----------Middel button
	flag8x123(t, []int{cards.TCMud, 7, 9, 10}, []int{11, 13, 14}, 34, 10, &cards.FWedge)

	troops1 := []int{cards.TCMud, 10, 9, 8, cards.TC8} // Miss
	troops2 := []int{1, 2, 3, cards.TC123}             //Miss big hole
	flagStd(t, troops1, troops2, 35, 9, &cards.FBattalion)
}
func flagStd(t *testing.T, troops1, troops2 []int, ex1, ex2 int, formation *cards.Formation) (flag *Flag) {
	flag = New()
	player1 := 0
	player2 := 1
	for _, cix := range troops1 {
		flag.Set(cix, player1)
	}
	for _, cix := range troops2 {
		flag.Set(cix, player2)
	}
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != formation {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex1 {
		t.Errorf("Strenght wrong expect %v got %v", ex1, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != formation {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != ex2 {
		t.Errorf("Strenght wrong expect %v got %v", ex2, flag.Players[player2].Strenght)
	}
	return flag
}

//TestFlagT1Phalanx testing Phalax
func TestFlagT1Phalanx(t *testing.T) {
	flagLeader(t, []int{7, 17}, []int{11, 21}, 21, 3, &cards.FPhalanx)
	flag8x123(t, []int{8, 18}, []int{12, 22}, 24, 6, &cards.FPhalanx)
	flag8x123(t, []int{8, 18, 28}, []int{12, 22, 32, cards.TCMud}, 32, 8, &cards.FPhalanx)
}
func flag8x123(t *testing.T, troops1, troops2 []int, ex1, ex2 int,
	formation *cards.Formation) (flag *Flag) {
	troops1 = append(troops1, cards.TC8)
	troops2 = append(troops2, cards.TC123)
	return flagStd(t, troops1, troops2, ex1, ex2, formation)
}

//TestFlagLeader test one leader.
func flagLeader(t *testing.T, troops1, troops2 []int, ex1, ex2 int, formation *cards.Formation) (flag *Flag) {
	troops1 = append(troops1, cards.TCAlexander)
	troops2 = append(troops2, cards.TCDarius)
	return flagStd(t, troops1, troops2, ex1, ex2, formation)
}

//TestFlagT1Battalion testing Battalion
func TestFlagT1Battalion(t *testing.T) {
	flagLeader(t, []int{7, 1}, []int{11, 20}, 18, 21, &cards.FBattalion)
	flag8x123(t, []int{7, 1}, []int{11, 20}, 16, 14, &cards.FBattalion)
}

//TestFlagT1Line testing a line Formation
func TestFlagT1Line(t *testing.T) {
	flagLeader(t, []int{7, 18}, []int{3, 15}, 24, 12, &cards.FSkirmish)
	flag8x123(t, []int{7, 19}, []int{11, 22}, 24, 6, &cards.FSkirmish)
}

//TestFlagT1Host testing a no Formation
func TestFlagT1Host(t *testing.T) {
	flagLeader(t, []int{7, 20}, []int{3, 16}, 27, 19, &cards.FHost)
	flag8x123(t, []int{7, 30}, []int{11, 24}, 25, 8, &cards.FHost)
}

//TestFlagT2Wedge testing wedge with two jokers
func TestFlagT2Wedge(t *testing.T) {
	//---------Low
	flagLeader(t, []int{cards.TC8, 6}, []int{cards.TC123, 1}, 21, 6, &cards.FWedge)

	//---------Middel
	flagLeader(t, []int{cards.TC8, 7}, []int{cards.TC123, 2}, 24, 9, &cards.FWedge)
	flagLeader(t, []int{cards.TC8, 9}, []int{cards.TC123, 3}, 27, 9, &cards.FWedge)
	flagLeader(t, []int{cards.TC123, 5}, []int{cards.TC123, 4}, 12, 12, &cards.FWedge)
	flagLeader(t, []int{cards.TC8, 10}, []int{cards.TC123, 5}, 27, 12, &cards.FWedge)

	//============Mud
	flagLeader(t, []int{cards.TCMud, cards.TC8, 7, 9}, []int{cards.TC123, 2, 3}, 34, 10, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 7, 10}, []int{cards.TC123, 1, 4}, 34, 10, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 7, 6}, []int{cards.TC123, 1, 3}, 30, 10, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 10, 9}, []int{cards.TC123, 5, 3}, 34, 14, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 6, 9}, []int{cards.TC123, 4, 3}, 30, 14, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 6, 5}, []int{cards.TC123, 1, 2}, 26, 10, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 7, 5}, []int{cards.TC123, 5, 6}, 26, 18, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC123, 2, 4}, []int{cards.TC123, 4, 6}, 14, 18, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC123, 1, 4}, []int{cards.TC123, 4, 5}, 10, 18, &cards.FWedge)
	flagLeader(t, []int{cards.TCMud, cards.TC123, 1, 5}, []int{cards.TC123, 11, 15}, 19, 19, &cards.FBattalion)
}

//TestFlagT2Phalanx testing Phalanx with two jokers
func TestFlagT2Phalanx(t *testing.T) {
	//---------Low
	flagLeader(t, []int{cards.TC8, 8}, []int{cards.TC8, 18}, 24, 24, &cards.FPhalanx)
	flagLeader(t, []int{cards.TC123, 1}, []int{cards.TC123, 11}, 6, 6, &cards.FWedge)

	flagLeader(t, []int{cards.TCMud, cards.TC8, 8, 18}, []int{cards.TC123, 1, 11}, 32, 4, &cards.FPhalanx)

	flagLeader(t, []int{cards.TCMud, cards.TC8, 8, 19}, []int{cards.TC8, 3, 19}, 35, 30, &cards.FHost)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 8, 18}, []int{cards.TC123, 2, 22}, 32, 8, &cards.FPhalanx)

	flagLeader(t, []int{cards.TCMud, cards.TC8, 8, 28}, []int{cards.TC123, 3, 33}, 32, 12, &cards.FPhalanx)
}

//TestSort tests sort.
func TestSort(t *testing.T) {
	v := []int{4, 3, 2, 7}
	exv := []int{7, 4, 3, 2}
	sortInt(v)
	for i := range v {
		if v[i] != exv[i] {
			t.Errorf("sort big first %v ", v)
		}
	}
}

//TestFlagT2Line testing Line with two jokers
func TestFlagT2Line(t *testing.T) {
	flagLeader(t, []int{cards.TCMud, cards.TC8, 9, 27}, []int{cards.TC123, 3, 14}, 34, 14, &cards.FSkirmish)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 9, 20}, []int{cards.TC123, 21, 33}, 34, 10, &cards.FSkirmish)
	flagLeader(t, []int{cards.TCMud, cards.TC8, 1, 19}, []int{cards.TC123, 1, 36}, 28, 20, &cards.FHost)
}

//TestFlagWedge testing wedge no jokers
func TestFlagWedge(t *testing.T) {
	flagStd(t, []int{7, 9, 8}, []int{31, 32, 33}, 24, 6, &cards.FWedge)
	flagStd(t, []int{cards.TCMud, 8, 10, 9, 7}, []int{4, 5, 6, 7}, 34, 22, &cards.FWedge)
}

//TestFlagPhalanx testing Phalanx no jokers
func TestFlagPhalanx(t *testing.T) {
	flagStd(t, []int{7, 17, 27}, []int{8, 38, 58}, 21, 24, &cards.FPhalanx)
	flagStd(t, []int{cards.TCMud, 8, 18, 28, 38}, []int{31, 11, 1, 41}, 32, 4, &cards.FPhalanx)
	flagStd(t, []int{cards.TCMud, 8, 18, 28, 38}, []int{cards.TCFog, 31, 11, 1, 41}, 32, 4, &cards.FHost)
}

//TestFlagBattalion testing line no jokers
func TestFlagBattalion(t *testing.T) {
	flagStd(t, []int{7, 10, 8}, []int{31, 39, 32}, 25, 12, &cards.FBattalion)
	flagStd(t, []int{cards.TCMud, 8, 1, 9, 7}, []int{1, 5, 6, 7}, 25, 19, &cards.FBattalion)
}

//TestFlagLine testing wedge no jokers
func TestFlagLine(t *testing.T) {
	flagStd(t, []int{7, 19, 8}, []int{31, 43, 32}, 24, 6, &cards.FSkirmish)
	flagStd(t, []int{cards.TCMud, 8, 10, 39, 7}, []int{4, 15, 6, 7}, 34, 22, &cards.FSkirmish)
}

//TestFlagHost testing Host no jokers
func TestFlagHost(t *testing.T) {
	flagStd(t, []int{6, 19, 9}, []int{41, 32, 52}, 24, 5, &cards.FHost)
	flagStd(t, []int{cards.TCMud, 18, 1, 9, 7}, []int{8, 18, 28, 7}, 25, 31, &cards.FHost)
}
func flagClaimsTest(t *testing.T, formation, unUsed, oppCards, env []int, playerix int) (ok bool, eks []int) {
	t.Logf("UnUsed: %v", unUsed)
	flag := New()
	for _, cardix := range formation {
		flag.Set(cardix, playerix)
	}
	for _, cardix := range oppCards {
		flag.Set(cardix, opponent(playerix))
	}
	for _, cardix := range env {
		flag.Set(cardix, playerix)
	}
	t.Logf("Flag: %+v", *flag)
	ok, eks = flag.ClaimFlag(playerix, unUsed)
	return ok, eks
}
func flagClaimsTestUsed(t *testing.T, formation, all, used, oppCards, env []int, playerix int) (ok bool, eks []int) {
	unUsed := deleteCards(all, used)
	unUsed = deleteCards(unUsed, formation)
	unUsed = deleteCards(unUsed, oppCards)
	ok, eks = flagClaimsTest(t, formation, unUsed, oppCards, env, playerix)
	return ok, eks
}
func flagClaimExpectFail(t *testing.T, formation, unUsed, oppCards, env []int, playerix int) {
	ok, _ := flagClaimsTest(t, formation, unUsed, oppCards, env, playerix)
	if ok {
		t.Errorf("Claim should have fail")
	}
}
func flagClaimExpectSucces(t *testing.T, formation, unUsed, oppCards, env []int, playerix int) {
	ok, eks := flagClaimsTest(t, formation, unUsed, oppCards, env, playerix)
	if !ok { //ok
		t.Errorf("Claim should have succed. but example exist: %v", eks)
	}
}
func TestFlagClaims(t *testing.T) {
	player1 := 0
	player2 := opponent(player1)
	all := make([]int, 60)
	for i := 0; i < 60; i++ {
		all[i] = i + 1
	}
	formation := []int{9, 8, 7, 6}
	used := []int{20, 30, 50, 40}
	env := []int{cards.TCMud}
	var oppCards []int
	ok, res := flagClaimsTestUsed(t, formation, all, used, oppCards, env, player1)
	if !ok {
		if res != nil {
			exp := []int{60, 59, 58, 57}
			for i, v := range exp {
				if v != res[i] {
					t.Errorf("Expected %v got %v", exp, res)
					break
				}
			}
		} else {
			t.Errorf("Should have a result")
		}
	}
	formation = []int{9, 18, 6, 7}
	used = []int{20, 30, 50, 40}
	env = []int{cards.TCMud}
	oppCards = nil
	ok, _ = flagClaimsTestUsed(t, formation, all, used, oppCards, env, player1)
	if ok {
		t.Errorf("Claim should have fail")
	}

	formation = []int{9, 18, 6, 7}
	unUsed := []int{1, 11, 22, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = nil
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player1)

	formation = []int{9, 8, 6, 7}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = []int{29, 18}
	t.Logf("Pree wedge sim exit player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player1)

	formation = []int{9, 8, 5, 7}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = []int{29, 18}
	t.Logf("Pree battalion sim exit player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player1)

	formation = []int{17, 27, 37, 7}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = []int{29, 18}
	t.Logf("Pree phalanx sim exit player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player1)

	formation = []int{6, 15, 4, 3}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = []int{9, 21}
	t.Logf("Pree Line sim exit player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player1)

	formation = []int{6, 15, 4, 3}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	env = []int{cards.TCMud}
	oppCards = []int{17, 9}
	t.Logf("Fail pree Line sim exit player,opponent %v,%v", formation, oppCards)
	flagClaimExpectFail(t, formation, unUsed, oppCards, env, player1)

	formation = []int{59, 60, 58}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	env = nil
	oppCards = []int{40, 39}
	t.Logf("Max formation player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player2)

	formation = []int{59, 60, 58}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	env = nil
	oppCards = []int{40, 38, 39}
	t.Logf("Same formation player,opponent %v,%v", formation, oppCards)
	flagClaimExpectSucces(t, formation, unUsed, oppCards, env, player2)

}
func deleteCards(source []int, del []int) (res []int) {
	if len(del) != 0 {
		res = make([]int, len(source)-len(del))
		r := 0
		var delete bool
		copyDel := make([]int, len(del))
		copy(copyDel, del)
		for _, v := range source {
			delete = false
			for j, d := range copyDel {
				if d == v {
					delete = true
					copyDel = append(copyDel[:j], copyDel[j+1:]...)
					break
				}
			}
			if !delete {
				res[r] = v
				r++
			}
		}
	} else {
		res = source
	}
	return res
}
func TestDecoder(t *testing.T) {
	flag := New()
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	// Encoding the map
	err := e.Encode(flag)
	if err != nil {
		t.Errorf("Error encoding")
	}

	var loadFlag Flag
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err = d.Decode(&loadFlag)
	if err != nil {
		t.Errorf("Error decoding")
	} else {
		if !flag.Equal(&loadFlag) {
			t.Logf("Deck :%v\nLoad :%v", flag, loadFlag)
			t.Error("Save and load deck not equal")
		}
	}
}
func TestCopy(t *testing.T) {
	flag := New()
	flag.Set(40, 0)
	flag.Set(20, 1)
	flag2 := flag.Copy()
	flag.Set(10, 1)
	if flag.Equal(flag2) {
		t.Error("should be differnt")
	}
}
