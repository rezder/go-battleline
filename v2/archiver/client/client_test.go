package client

import (
	"github.com/rezder/go-battleline/v2/archiver/arnet"
	bg "github.com/rezder/go-battleline/v2/game"
	"testing"
	"time"
)

func TestMq(t *testing.T) {

	client, err := New(7373, "")
	if err != nil {
		t.Fatalf("Create poke listener failed: %v", err)
	}
	client.Start()
	nzr, err := arnet.NewReceiver("7272")
	if err != nil {
		client.Stop()
		t.Fatalf("Creating net reciever failed %v", err)
	}
	nzr.Start()
	arnet.PokeClient("localhost:7373", "localhost:7272")
	t.Log("sleep 1")
	time.Sleep(time.Second)

	game := createGame(1, 2, 7)
	client.Archive(game.Hist)
	histBytes, open := <-nzr.HistCh
	if !open {
		t.Error("No Game send channel closed")
	} else {
		t.Log("recieved game")
	}
	client.Stop()
	err = nzr.Close()
	if err != nil {
		t.Errorf("Closing zmq receiver failed with %v", err)
	}
	if open {
		histSend, err := arnet.HistDecoder(histBytes)
		if err != nil {
			t.Errorf("Hist decode failed %v", err)
		}
		if !histSend.IsEqual(game.Hist) {
			t.Error("Histories do not compare")
		}
	}
}
func addMoves(no int, game *bg.Game) {
	for i := 0; i < no; i++ {
		moves := game.Pos.CalcMoves()
		game.Move(moves[len(moves)-1])
	}
}
func createGame(id1, id2, moveNo int) (game *bg.Game) {
	game = bg.NewGame()
	game.Start([2]int{id1, id2}, 0)
	addMoves(moveNo, game)
	return game
}
