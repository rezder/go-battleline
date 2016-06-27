package flag

import (
	//"errors"
	//"fmt"
	"bytes"
	"encoding/gob"
	"rezder.com/game/card/battleline/cards"
	"testing"
)

//TestFlagT1LeaderWedge testing wedge with one leader
func TestFlagT1LeaderWedge(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(2, player1)
	flag.Set(3, player1)

	flag.Set(11, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(13, player2)
	t.Logf("Flag %+v", flag)
	//---------Top
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != 9 {
		t.Errorf("Strenght wrong expect %v got %v", 9, flag.Players[player1].Strenght)
	}
	//----------Middel
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != 6 {
		t.Errorf("Strenght wrong expect %v got %v", 6, flag.Players[player2].Strenght)
	}
	//----------Buttom

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

	flag.Set(9, player1)
	flag.Set(10, player1)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != 27 {
		t.Errorf("Strenght wrong expect %v got %v", 27, flag.Players[player1].Strenght)
	}
	//----------Fog
	flag.Set(cards.TC_Fog, player1)
	t.Logf("Flag %+v", flag)
	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player2].Strenght != 14 {
		t.Errorf("Strenght wrong expect %v got %v", 14, flag.Players[player1].Strenght)
	}

	//===========Mud
	flag = New()
	//-----------Top
	flag.Set(cards.TC_Mud, player1)
	flag.Set(1, player1)
	flag.Set(2, player1)
	flag.Set(3, player1)
	flag.Set(cards.TC_Alexander, player1)
	//---------Buttom
	flag.Set(20, player2)
	flag.Set(19, player2)
	flag.Set(18, player2)
	flag.Set(cards.TC_Darius, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != 10 {
		t.Errorf("Strenght wrong expect %v got %v", 10, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != 34 {
		t.Errorf("Strenght wrong expect %v got %v", 34, flag.Players[player2].Strenght)
	}
	mud1, mud2, err = flag.Remove(cards.TC_Mud, player1)
	ex := 1
	if mud1 != ex {
		t.Errorf("Expected mud 1 index: %v got: %v", ex, mud1)
	}
	ex = cards.TC_Darius // this can only happen if you did not claim the flag as you hadd the best Formation
	if mud2 != ex {
		t.Errorf("Expected mud 2 index: %v got: %v", ex, mud2)
	}
	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 9
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	ex = 27
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	//-----------Middel Top
	flag.Set(cards.TC_Mud, player1)
	flag.Set(1, player1)
	flag.Set(2, player1)
	flag.Set(4, player1)
	flag.Set(cards.TC_Alexander, player1)
	//---------Middel Buttom
	flag.Set(20, player2)
	flag.Set(19, player2)
	flag.Set(17, player2)
	flag.Set(cards.TC_Darius, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != 10 {
		t.Errorf("Strenght wrong expect %v got %v", 10, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != 34 {
		t.Errorf("Strenght wrong expect %v got %v", 34, flag.Players[player2].Strenght)
	}
	flag = New()
	//-----------Miss two step
	flag.Set(cards.TC_Mud, player1)
	flag.Set(1, player1)
	flag.Set(3, player1)
	flag.Set(5, player1)
	flag.Set(cards.TC_Alexander, player1)
	//--------- Miss big hole
	flag.Set(20, player2)
	flag.Set(16, player2)
	flag.Set(17, player2)
	flag.Set(cards.TC_Darius, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 19
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 33
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

// TestFlagT1NWedge testing wedge with one number joker.
// Player 1 have 8 player 2 have 123.
func TestFlagT1NWedge(t *testing.T) {
	//------Top
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(6, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(12, player2)

	t.Logf("Flag %+v", flag)
	//---------Top
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Top Formation 8 wrong : %v", flag.Players[player1].Formation)
	}
	ex := 21
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Top Formation 123  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	//----------Buttom
	flag = New()
	flag.Set(cards.TC_8, player1)
	flag.Set(10, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(12, player2)
	flag.Set(13, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Button Formation 8 wrong : %v", flag.Players[player1].Formation)
	}
	ex = 27
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Button  Strenght 8 wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Button Formation 123  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Button Strenght 123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

	//----------Fog
	flag.Set(cards.TC_Fog, player1)
	t.Logf("Flag %+v", flag)
	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Fog Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 8
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Fog Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	ex = 27
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Fog Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	//----------Middel
	flag = New()
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(13, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Button Formation 8 wrong : %v", flag.Players[player1].Formation)
	}
	ex = 24
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Button  Strenght 8 wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Button Formation 123  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Button Strenght 123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

	//===========Mud
	flag = New()
	//-----------Top
	flag.Set(cards.TC_Mud, player1)
	flag.Set(5, player1)
	flag.Set(7, player1)
	flag.Set(6, player1)
	flag.Set(cards.TC_8, player1)

	flag.Set(11, player2)
	flag.Set(12, player2)
	flag.Set(14, player2)
	flag.Set(cards.TC_123, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Mud top 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 26
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Mud top 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Mud top 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Mud top 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	t.Logf("Remove mud 8 and 123")
	mud1, mud2, err := flag.Remove(cards.TC_Mud, player1)
	ex = 5
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
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 21
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	ex = 9
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	//-----------Middel Top
	flag.Set(cards.TC_Mud, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	flag.Set(9, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(11, player2)
	flag.Set(12, player2)
	flag.Set(14, player2)
	flag.Set(cards.TC_123, player2)
	t.Logf("Mud Middel Top Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 30
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Mud Middel Top 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Mud Middel Top 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	//-----------Middel button
	flag.Set(cards.TC_Mud, player1)
	flag.Set(7, player1)
	flag.Set(9, player1)
	flag.Set(10, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(11, player2)
	flag.Set(13, player2)
	flag.Set(14, player2)
	flag.Set(cards.TC_123, player2)
	t.Logf("Mud Middel button Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Mud Middel button 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Mud Middel button 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Mud Middel button 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

	flag = New()
	//-----------Miss
	flag.Set(cards.TC_Mud, player1)
	flag.Set(10, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)
	flag.Set(cards.TC_8, player1)
	//---------Miss big hole
	flag.Set(1, player2)
	flag.Set(2, player2)
	flag.Set(3, player2)
	flag.Set(cards.TC_123, player2)
	t.Logf("Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 35
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 9
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

}

//TestFlagT1Phalanx testing Phalax
func TestFlagT1Phalanx(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(17, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(11, player2)
	flag.Set(21, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Alex Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex := 21
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Alex Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("Darius  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 3
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()

	flag.Set(cards.TC_8, player1)
	flag.Set(8, player1)
	flag.Set(18, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(12, player2)
	flag.Set(22, player2)

	t.Logf("N Flag %+v", flag)

	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 8 wrong : %v", flag.Players[player1].Formation)
	}
	ex = 24
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 123  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag.Set(cards.TC_Mud, player2)
	flag.Set(28, player1)
	flag.Set(32, player2)
	t.Logf("N Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 8 wrong : %v", flag.Players[player1].Formation)
	}
	ex = 32
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 123  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 8
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagT1Battalion testing Battalion
func TestFlagT1Battalion(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)

	flag.Set(1, player1)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(11, player2)
	flag.Set(20, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("Alex Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex := 18
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Alex Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("Darius  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 21
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(1, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(20, player2)

	t.Logf("N Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 16
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("123 wrong : %v", flag.Players[player2].Formation)
	}
	ex = 14
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagT1Line testing a line Formation
func TestFlagT1Line(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(18, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(15, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("Alex Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex := 24
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Alex Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("Darius  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 12
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(19, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(22, player2)

	t.Logf("N Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 24
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("123 wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

}

//TestFlagT1Host testing a no Formation
func TestFlagT1Host(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(20, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(16, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Alex Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex := 27
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Alex Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Darius  wrong : %v", flag.Players[player2].Formation)
	}
	ex = 19
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(30, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(24, player2)

	t.Logf("N Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 25
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Formation wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("123 wrong : %v", flag.Players[player2].Formation)
	}
	ex = 8
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagT2Wedge testing wedge with two jokers
func TestFlagT2Wedge(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	//---------Low
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(6, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	t.Logf("Low Flag %+v", flag)
	ex := 21
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Low 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Low 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Low 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Low 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	//---------Middel
	flag = New()
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(2, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Middel 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 24
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Middel 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 9
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Middel 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 27
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Middel 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 9
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_123, player1)
	flag.Set(5, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(4, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 12
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 12
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(10, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(5, player2)
	t.Logf("High Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 27
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf(" 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 12
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	//============Mud
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(2, player2)
	flag.Set(3, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(10, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	flag.Set(4, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(6, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	flag.Set(3, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 30
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(10, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(5, player2)
	flag.Set(3, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 14
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(6, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(4, player2)
	flag.Set(3, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 30
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 14
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(6, player1)
	flag.Set(5, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	flag.Set(2, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 26
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(5, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(5, player2)
	flag.Set(6, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 26
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 18
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_123, player1)
	flag.Set(2, player1)
	flag.Set(4, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(4, player2)
	flag.Set(6, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 14
	if flag.Players[player1].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 18
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_123, player1)
	flag.Set(1, player1)
	flag.Set(5, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(4, player2)
	flag.Set(5, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("123 Miss Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 19
	if flag.Players[player1].Strenght != ex {
		t.Errorf("123 miss Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 18
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagT2Phalanx testing Phalanx with two jokers
func TestFlagT2Phalanx(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	//---------Low
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(8, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	t.Logf(" Flag %+v", flag)
	ex := 24
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Low 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Low 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Miss 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Miss 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(8, player1)
	flag.Set(18, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	flag.Set(1, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 32
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 4
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(8, player1)
	flag.Set(19, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(2, player2)
	flag.Set(2, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 35
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 8
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(3, player1)
	flag.Set(19, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(3, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 30
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 12
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagT2Line testing Line with two jokers
func TestFlagT2Line(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Mud, player2)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(9, player1)
	flag.Set(27, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(14, player2)
	t.Logf(" Mud Flag %+v", flag)
	ex := 34
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("Mud  8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Mud 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("Mud 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 14
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Mud 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(9, player1)
	flag.Set(20, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(21, player2)
	flag.Set(33, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 10
	if flag.Players[player2].Strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	v := []int{4, 3, 2, 7}
	exv := []int{7, 4, 3, 2}
	sortInt(v)
	for i := range v {
		if v[i] != exv[i] {
			t.Errorf("sort big first %v ", v)
		}
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(1, player1)
	flag.Set(19, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(1, player2)
	flag.Set(36, player2)

	t.Logf("Miss Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 28
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Miss 123 Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 20
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Miss 123 Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagWedge testing wedge no jokers
func TestFlagWedge(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(7, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)

	flag.Set(31, player2)
	flag.Set(33, player2)
	flag.Set(32, player2)
	t.Logf("Wedge Flag %+v", flag)
	ex := 24
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(8, player1)
	flag.Set(10, player1)
	flag.Set(9, player1)
	flag.Set(7, player1)

	flag.Set(4, player2)
	flag.Set(5, player2)
	flag.Set(6, player2)
	flag.Set(7, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 22
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}

}

//TestFlagPhalanx testing Phalanx no jokers
func TestFlagPhalanx(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(7, player1)
	flag.Set(17, player1)
	flag.Set(27, player1)

	flag.Set(8, player2)
	flag.Set(38, player2)
	flag.Set(58, player2)
	t.Logf("Phalanx Flag %+v", flag)
	ex := 21
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 24
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(8, player1)
	flag.Set(18, player1)
	flag.Set(28, player1)
	flag.Set(38, player1)

	flag.Set(31, player2)
	flag.Set(41, player2)
	flag.Set(1, player2)
	flag.Set(51, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 32
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 4
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(8, player1)
	flag.Set(18, player1)
	flag.Set(28, player1)
	flag.Set(38, player1)

	flag.Set(cards.TC_Fog, player2)
	flag.Set(31, player2)
	flag.Set(41, player2)
	flag.Set(1, player2)
	flag.Set(51, player2)

	t.Logf("Mud/Fog Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 32
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 4
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagBattalion testing line no jokers
func TestFlagBattalion(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(7, player1)
	flag.Set(10, player1)
	flag.Set(8, player1)

	flag.Set(31, player2)
	flag.Set(39, player2)
	flag.Set(32, player2)
	t.Logf("Battalion Flag %+v", flag)
	ex := 25
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 12
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(8, player1)
	flag.Set(1, player1)
	flag.Set(9, player1)
	flag.Set(7, player1)

	flag.Set(1, player2)
	flag.Set(5, player2)
	flag.Set(6, player2)
	flag.Set(7, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 25
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 19
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagLine testing wedge no jokers
func TestFlagLine(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(7, player1)
	flag.Set(19, player1)
	flag.Set(8, player1)

	flag.Set(31, player2)
	flag.Set(43, player2)
	flag.Set(32, player2)
	t.Logf("Line Flag %+v", flag)
	ex := 24
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 6
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(8, player1)
	flag.Set(10, player1)
	flag.Set(39, player1)
	flag.Set(7, player1)

	flag.Set(4, player2)
	flag.Set(15, player2)
	flag.Set(6, player2)
	flag.Set(7, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 34
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 22
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}

//TestFlagHost testing Host no jokers
func TestFlagHost(t *testing.T) {
	flag := New()
	player1 := 0
	player2 := 1
	flag.Set(6, player1)
	flag.Set(19, player1)
	flag.Set(9, player1)

	flag.Set(41, player2)
	flag.Set(32, player2)
	flag.Set(52, player2)
	t.Logf("Host Flag %+v", flag)
	ex := 24
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}
	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 5
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(18, player1)
	flag.Set(1, player1)
	flag.Set(9, player1)
	flag.Set(7, player1)

	flag.Set(8, player2)
	flag.Set(18, player2)
	flag.Set(28, player2)
	flag.Set(7, player2)

	t.Logf("Mud Flag %+v", flag)
	if flag.Players[player1].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player1].Formation)
	}
	ex = 25
	if flag.Players[player1].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player1].Strenght)
	}

	if flag.Players[player2].Formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.Players[player2].Formation)
	}
	ex = 31
	if flag.Players[player2].Strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.Players[player2].Strenght)
	}
}
func TestFlagClaims(t *testing.T) {
	player1 := 0
	all := make([]int, 60)
	for i := 0; i < 60; i++ {
		all[i] = i + 1
	}
	del := []int{9, 8, 7, 6, 20, 30, 50, 40}
	unUsed := deleteCards(all, del)
	t.Logf("UnUsed", unUsed)
	flag := New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	ok, res := flag.ClaimFlag(player1, unUsed)
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
	} else { //ok
		t.Errorf("Should fail")
	}
	del = []int{9, 18, 7, 6, 20, 30, 50, 40}
	unUsed = deleteCards(all, del)
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(18, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if ok { //ok
		t.Errorf("Should have fail")
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(18, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.Players[player1].Won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	player2 := 1
	flag.Set(9, player2)
	flag.Set(18, player2)
	t.Logf("Pree wedge sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.Players[player1].Won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)
	flag.Set(5, player1)
	flag.Set(7, player1)

	flag.Set(9, player2)
	flag.Set(18, player2)
	t.Logf("Pree battalion sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.Players[player1].Won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(17, player1)
	flag.Set(27, player1)
	flag.Set(37, player1)
	flag.Set(7, player1)

	flag.Set(9, player2)
	flag.Set(18, player2)
	t.Logf("Pree phalanx sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.Players[player1].Won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(6, player1)
	flag.Set(15, player1)
	flag.Set(4, player1)
	flag.Set(3, player1)

	flag.Set(9, player2)
	flag.Set(11, player2)
	t.Logf("Pree Line sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.Players[player1].Won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = New()
	flag.Set(cards.TC_Mud, player1)
	flag.Set(6, player1)
	flag.Set(15, player1)
	flag.Set(4, player1)
	flag.Set(3, player1)

	flag.Set(17, player2)
	flag.Set(9, player2)
	t.Logf("Fail pree Line sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if flag.Players[player1].Won { //ok
		t.Errorf("Should have failed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	t.Logf("UnUsed: %v\n", unUsed)
	flag = New()
	flag.Set(40, player1)
	flag.Set(39, player1)

	flag.Set(59, player2)
	flag.Set(60, player2)
	flag.Set(58, player2)
	t.Logf("Max formation %+v", flag)
	ok, res = flag.ClaimFlag(player2, unUsed)
	if !flag.Players[player2].Won { //ok
		t.Errorf("Should have succeded. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	t.Logf("UnUsed: %v\n", unUsed)
	flag = New()
	flag.Set(40, player1)
	flag.Set(38, player1)
	flag.Set(39, player1)

	flag.Set(59, player2)
	flag.Set(60, player2)
	flag.Set(58, player2)
	t.Logf("Same formation %+v", flag)
	ok, res = flag.ClaimFlag(player2, unUsed)
	if !flag.Players[player2].Won { //ok
		t.Errorf("Should have succeded. Ok: %v, res: %v\n", ok, res)
	}
}
func deleteCards(source []int, del []int) (res []int) {
	res = make([]int, len(source)-len(del))
	r := 0
	var delete bool
	for _, v := range source {
		delete = false
		for j, d := range del {
			if d == v {
				delete = true
				del = append(del[:j], del[j+1:]...)
				break
			}
		}
		if !delete {
			res[r] = v
			r++
		}
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
