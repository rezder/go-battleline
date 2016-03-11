package flag

import (
	//"errors"
	//"fmt"
	"rezder.com/game/card/battleline/cards"
	"testing"
)

//TestFlagT1LeaderWedge testing wedge with one leader
func TestFlagT1LeaderWedge(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != 9 {
		t.Errorf("Strenght wrong expect %v got %v", 9, flag.players[player1].strenght)
	}
	//----------Middel
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	if flag.players[player2].strenght != 6 {
		t.Errorf("Strenght wrong expect %v got %v", 6, flag.players[player2].strenght)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != 27 {
		t.Errorf("Strenght wrong expect %v got %v", 27, flag.players[player1].strenght)
	}
	//----------Fog
	flag.Set(cards.TC_Fog, player1)
	t.Logf("Flag %+v", flag)
	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player2].strenght != 14 {
		t.Errorf("Strenght wrong expect %v got %v", 14, flag.players[player1].strenght)
	}

	//===========Mud
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != 10 {
		t.Errorf("Strenght wrong expect %v got %v", 10, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	if flag.players[player2].strenght != 34 {
		t.Errorf("Strenght wrong expect %v got %v", 34, flag.players[player2].strenght)
	}
	mud1, mud2, err = flag.Remove(cards.TC_Mud, player1)
	ex := 1
	if mud1 != ex {
		t.Errorf("Expected mud 1 index: %v got: %v", ex, mud1)
	}
	ex = cards.TC_Darius // this can only happen if you did not claim the flag as you hadd the best formation
	if mud2 != ex {
		t.Errorf("Expected mud 2 index: %v got: %v", ex, mud2)
	}
	if err != nil {
		t.Errorf("We do not expect a error: %v", err)
	}
	t.Logf("Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 9
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	ex = 27
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != 10 {
		t.Errorf("Strenght wrong expect %v got %v", 10, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	if flag.players[player2].strenght != 34 {
		t.Errorf("Strenght wrong expect %v got %v", 34, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 19
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 33
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

// TestFlagT1NWedge testing wedge with one number joker.
// Player 1 have 8 player 2 have 123.
func TestFlagT1NWedge(t *testing.T) {
	//------Top
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Top Formation 8 wrong : %v", flag.players[player1].formation)
	}
	ex := 21
	if flag.players[player1].strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Top Formation 123  wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	//----------Buttom
	flag = new(Flag)
	flag.Set(cards.TC_8, player1)
	flag.Set(10, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(12, player2)
	flag.Set(13, player2)
	t.Logf("Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Button Formation 8 wrong : %v", flag.players[player1].formation)
	}
	ex = 27
	if flag.players[player1].strenght != ex {
		t.Errorf("Button  Strenght 8 wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Button Formation 123  wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Button Strenght 123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}

	//----------Fog
	flag.Set(cards.TC_Fog, player1)
	t.Logf("Flag %+v", flag)
	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Fog Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 8
	if flag.players[player2].strenght != ex {
		t.Errorf("Fog Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	ex = 27
	if flag.players[player1].strenght != ex {
		t.Errorf("Fog Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	//----------Middel
	flag = new(Flag)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(13, player2)
	t.Logf("Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Button Formation 8 wrong : %v", flag.players[player1].formation)
	}
	ex = 24
	if flag.players[player1].strenght != ex {
		t.Errorf("Button  Strenght 8 wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Button Formation 123  wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Button Strenght 123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}

	//===========Mud
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Mud top 8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 26
	if flag.players[player1].strenght != ex {
		t.Errorf("Mud top 8 strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Mud top 123 formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("Mud top 123 strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 21
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	ex = 9
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 30
	if flag.players[player1].strenght != ex {
		t.Errorf("Mud Middel Top 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("Mud Middel Top 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Mud Middel Top 8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("Mud Middel button 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Mud Middel button 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("Mud Middel button 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}

	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 35
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 9
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}

}

//TestFlagT1Phalanx testing Phalax
func TestFlagT1Phalanx(t *testing.T) {
	flag := new(Flag)
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(17, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(11, player2)
	flag.Set(21, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Alex formation wrong : %v", flag.players[player1].formation)
	}
	ex := 21
	if flag.players[player1].strenght != ex {
		t.Errorf("Alex formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("Darius  wrong : %v", flag.players[player2].formation)
	}
	ex = 3
	if flag.players[player2].strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)

	flag.Set(cards.TC_8, player1)
	flag.Set(8, player1)
	flag.Set(18, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(12, player2)
	flag.Set(22, player2)

	t.Logf("N Flag %+v", flag)

	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 8 wrong : %v", flag.players[player1].formation)
	}
	ex = 24
	if flag.players[player1].strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 123  wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag.Set(cards.TC_Mud, player2)
	flag.Set(28, player1)
	flag.Set(32, player2)
	t.Logf("N Mud Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 8 wrong : %v", flag.players[player1].formation)
	}
	ex = 32
	if flag.players[player1].strenght != ex {
		t.Errorf("Top Strenght 8 wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("Top Formation 123  wrong : %v", flag.players[player2].formation)
	}
	ex = 8
	if flag.players[player2].strenght != ex {
		t.Errorf("Top Strenght 123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagT1Battalion testing Battalion
func TestFlagT1Battalion(t *testing.T) {
	flag := new(Flag)
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)

	flag.Set(1, player1)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(11, player2)
	flag.Set(20, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("Alex formation wrong : %v", flag.players[player1].formation)
	}
	ex := 18
	if flag.players[player1].strenght != ex {
		t.Errorf("Alex formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("Darius  wrong : %v", flag.players[player2].formation)
	}
	ex = 21
	if flag.players[player2].strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(1, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(20, player2)

	t.Logf("N Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 16
	if flag.players[player1].strenght != ex {
		t.Errorf("8 formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("123 wrong : %v", flag.players[player2].formation)
	}
	ex = 14
	if flag.players[player2].strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagT1Line testing a line formation
func TestFlagT1Line(t *testing.T) {
	flag := new(Flag)
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(18, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(15, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("Alex formation wrong : %v", flag.players[player1].formation)
	}
	ex := 24
	if flag.players[player1].strenght != ex {
		t.Errorf("Alex formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("Darius  wrong : %v", flag.players[player2].formation)
	}
	ex = 12
	if flag.players[player2].strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(19, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(22, player2)

	t.Logf("N Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 24
	if flag.players[player1].strenght != ex {
		t.Errorf("8 formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("123 wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}

}

//TestFlagT1Host testing a no formation
func TestFlagT1Host(t *testing.T) {
	flag := new(Flag)
	player1 := 0
	player2 := 1
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(7, player1)
	flag.Set(20, player1)

	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	flag.Set(16, player2)

	t.Logf("Leader Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Alex formation wrong : %v", flag.players[player1].formation)
	}
	ex := 27
	if flag.players[player1].strenght != ex {
		t.Errorf("Alex formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Darius  wrong : %v", flag.players[player2].formation)
	}
	ex = 19
	if flag.players[player2].strenght != ex {
		t.Errorf("Darius  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)
	flag.Set(30, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(11, player2)
	flag.Set(24, player2)

	t.Logf("N Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("8 formation wrong : %v", flag.players[player1].formation)
	}
	ex = 25
	if flag.players[player1].strenght != ex {
		t.Errorf("8 formation wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("123 wrong : %v", flag.players[player2].formation)
	}
	ex = 8
	if flag.players[player2].strenght != ex {
		t.Errorf("123  wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagT2Wedge testing wedge with two jokers
func TestFlagT2Wedge(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Low 8 Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Low 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Low 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Low 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	//---------Middel
	flag = new(Flag)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(7, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(2, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Middel 8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 24
	if flag.players[player1].strenght != ex {
		t.Errorf("Middel 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 9
	if flag.players[player2].strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(9, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(3, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Middel 8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 27
	if flag.players[player1].strenght != ex {
		t.Errorf("Middel 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 9
	if flag.players[player2].strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_123, player1)
	flag.Set(5, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(4, player2)
	t.Logf("Middel Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 12
	if flag.players[player1].strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Middel 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 12
	if flag.players[player2].strenght != ex {
		t.Errorf("Middel 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
	flag.Set(cards.TC_Alexander, player1)
	flag.Set(cards.TC_8, player1)
	flag.Set(10, player1)

	flag.Set(cards.TC_123, player2)
	flag.Set(cards.TC_Darius, player2)
	flag.Set(5, player2)
	t.Logf("High Flag %+v", flag)
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 27
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf(" 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 12
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	//============Mud
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 30
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 14
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 30
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 14
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 26
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 26
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 18
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 14
	if flag.players[player1].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 18
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("123 Miss Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 19
	if flag.players[player1].strenght != ex {
		t.Errorf("123 miss Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 18
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagT2Phalanx testing Phalanx with two jokers
func TestFlagT2Phalanx(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Low 8 Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Low 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Miss 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Miss 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 32
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 4
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 35
	if flag.players[player1].strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 8
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 30
	if flag.players[player1].strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 12
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagT2Line testing Line with two jokers
func TestFlagT2Line(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("Mud  8 Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Mud 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("Mud 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 14
	if flag.players[player2].strenght != ex {
		t.Errorf("Mud 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 10
	if flag.players[player2].strenght != ex {
		t.Errorf("123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	v := []int{4, 3, 2, 7}
	exv := []int{7, 4, 3, 2}
	sortInt(v)
	for i := range v {
		if v[i] != exv[i] {
			t.Errorf("sort big first %v ", v)
		}
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Miss 8 Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 28
	if flag.players[player1].strenght != ex {
		t.Errorf("Miss 8 Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Miss 123 Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 20
	if flag.players[player2].strenght != ex {
		t.Errorf("Miss 123 Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagWedge testing wedge no jokers
func TestFlagWedge(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Wedge {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 22
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagPhalanx testing Phalanx no jokers
func TestFlagPhalanx(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 24
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 32
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Phalanx {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 4
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 32
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 4
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagBattalion testing line no jokers
func TestFlagBattalion(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 12
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 25
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_BattalionOrder {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 19
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagLine testing wedge no jokers
func TestFlagLine(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 6
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 34
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_SkirmishLine {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 22
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
}

//TestFlagHost testing Host no jokers
func TestFlagHost(t *testing.T) {
	flag := new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}
	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 5
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
	}
	flag = new(Flag)
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
	if flag.players[player1].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player1].formation)
	}
	ex = 25
	if flag.players[player1].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player1].strenght)
	}

	if flag.players[player2].formation != &cards.F_Host {
		t.Errorf("Formation wrong : %v", flag.players[player2].formation)
	}
	ex = 31
	if flag.players[player2].strenght != ex {
		t.Errorf("Strenght wrong expect %v got %v", ex, flag.players[player2].strenght)
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
	flag := new(Flag)
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
	flag = new(Flag)
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
	flag = new(Flag)
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(18, player1)
	flag.Set(6, player1)
	flag.Set(7, player1)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.players[player1].won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = new(Flag)
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
	if !flag.players[player1].won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = new(Flag)
	flag.Set(cards.TC_Mud, player1)
	flag.Set(9, player1)
	flag.Set(8, player1)
	flag.Set(5, player1)
	flag.Set(7, player1)

	flag.Set(9, player2)
	flag.Set(18, player2)
	t.Logf("Pree battalion sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.players[player1].won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = new(Flag)
	flag.Set(cards.TC_Mud, player1)
	flag.Set(17, player1)
	flag.Set(27, player1)
	flag.Set(37, player1)
	flag.Set(7, player1)

	flag.Set(9, player2)
	flag.Set(18, player2)
	t.Logf("Pree phalanx sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.players[player1].won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 11, 22, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = new(Flag)
	flag.Set(cards.TC_Mud, player1)
	flag.Set(6, player1)
	flag.Set(15, player1)
	flag.Set(4, player1)
	flag.Set(3, player1)

	flag.Set(9, player2)
	flag.Set(11, player2)
	t.Logf("Pree Line sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if !flag.players[player1].won { //ok
		t.Errorf("Should have succed. Ok: %v, res: %v\n", ok, res)
	}
	unUsed = []int{1, 18, 16, 46, 55, 56}
	t.Logf("UnUsed", unUsed)
	flag = new(Flag)
	flag.Set(cards.TC_Mud, player1)
	flag.Set(6, player1)
	flag.Set(15, player1)
	flag.Set(4, player1)
	flag.Set(3, player1)

	flag.Set(17, player2)
	flag.Set(9, player2)
	t.Logf("Fail pree Line sim exit Flag %+v", flag)
	ok, res = flag.ClaimFlag(player1, unUsed)
	if flag.players[player1].won { //ok
		t.Errorf("Should have failed. Ok: %v, res: %v\n", ok, res)
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
