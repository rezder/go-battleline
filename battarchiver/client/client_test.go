package client

import (
	"github.com/rezder/go-battleline/battarchiver/arnet"
	bat "github.com/rezder/go-battleline/battleline"
	"testing"
	"time"
)

func TestMq(t *testing.T) {

	client, err := New(7373, "")
	if err != nil {
		t.Fatalf("Create poke listener failed: %v", err)
	}
	client.Start()
	nzr, err := arnet.NewZmqReciver("7272")
	if err != nil {
		client.Stop()
		client.WaitToFinish()
		t.Fatalf("Creating net reciever failed %v", err)
	}
	nzr.Start()
	arnet.PokeClient("localhost:7373", "localhost:7272")
	t.Log("sleep 1")
	time.Sleep(time.Second)

	game := createGame(1, 2, 7)
	client.Archive(game)
	gameBytes, open := <-nzr.GameCh
	if !open {
		t.Error("No Game send channel closed")
	} else {
		t.Log("recieved game")
	}
	client.Stop()
	t.Log("client stop")
	client.WaitToFinish()
	t.Log("finished")
	nzr.Close()
	if open {
		gameSend, err := arnet.ZmqDecoder(gameBytes)
		if err != nil {
			t.Errorf("Game decode failed %v", err)
		}
		gameSend.CalcPos()
		if !game.Equal(gameSend) {
			t.Error("Games do not compare")
		}
	}
}
func addMoves(no int, game *bat.Game) {
	game.Start(0)
	for i := 0; i < no; i++ {
		if len(game.Pos.MovesHand) > 0 {
			cardix := 0
			for ix := range game.Pos.MovesHand {
				cardix = ix
				break
			}
			game.MoveHand(cardix, len(game.Pos.MovesHand[cardix])-1)
		} else {
			game.Move(len(game.Pos.Moves) - 1)
		}
	}
}
func createGame(id1, id2, moveNo int) (game *bat.Game) {
	game = bat.New([2]int{id1, id2})
	addMoves(moveNo, game)
	return game
}
