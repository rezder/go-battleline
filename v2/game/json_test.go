package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSON(t *testing.T) {
	t.SkipNow()
	game := NewGame()
	game.Start([2]int{1, 2}, 0)
	moves := game.Pos.CalcMoves()
	game.Move(moves[4])
	hist := game.Hist
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(hist)
	if err != nil {
		t.Errorf("Error encoding json: %v", err)
		return
	}
	printView(game.Pos, ViewAll.Players[0], t)
	printView(game.Pos, ViewAll.Players[1], t)
	printView(game.Pos, ViewAll.Spectator, t)
	printView(game.Pos, ViewAll.God, t)
}
func printView(gamePos *Pos, view View, t *testing.T) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(NewViewPos(gamePos, view, NoPlayer))
	if err != nil {
		t.Errorf("Error encoding json: %v", err)
	}
}
