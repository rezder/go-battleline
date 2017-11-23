package tf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/rezder/go-battleline/v2/bot/prob"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestFloats(t *testing.T) {
	log.InitLog(log.Min)
	gamePos := game.NewPos()
	gamePos.CardPos = [71]pos.Card{0, 22, 18, 14, 9, 13, 12, 6, 11, 7, 7, 21, 22, 0, 22, 3, 2, 4, 15, 8, 13, 0, 18, 0, 16, 0, 15, 0, 17, 8, 0, 21, 15, 17, 0, 21, 1, 5, 6, 12, 0, 4, 5, 12, 0, 22, 3, 11, 17, 9, 11, 22, 0, 21, 22, 21, 18, 4, 6, 8, 22, 23, 23, 21, 23, 23, 23, 23, 23, 23, 21}
	gamePos.ConePos = [10]pos.Cone{0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	gamePos.ConePos = [10]pos.Cone{0, 2, 2, 1, 0, 0, 2, 2, 0, 1}
	gamePos.LastMoveType = game.MoveTypeAll.Cone
	gamePos.LastMoveIx = 92
	gamePos.LastMover = 0
	viewPos := game.NewViewPos(gamePos, game.ViewAll.Players[0], pos.NoPlayer)
	tfAnas, _ := CalcTfAnas(viewPos, nil)
	b, _ := movesToBytes(tfAnas)
	start := 19*3*4 + 4*4
	reader := bytes.NewReader(b[start : start+12])
	var fs [3]float32
	err := binary.Read(reader, binary.LittleEndian, &fs)
	if err != nil {
		t.Error("Failed to read floats")
	}
	exp := [3]float32{2, 16, 0.4852941176470588}
	if fs != exp {
		t.Errorf("Failed MissNo float deviates exp: %v, got: %v", exp, fs)
	}
	start = 0
	reader = bytes.NewReader(b[start : start+12])
	err = binary.Read(reader, binary.LittleEndian, &fs)
	if err != nil {
		t.Error("Failed to read floats")
	}
	exp = [3]float32{0, 0, 1}
	if fs != exp {
		t.Errorf("Failed MissNo float deviates exp: %v, got: %v", exp, fs)
	}
	fmt.Println("Start python server")
	serverCmd := exec.Command("python", "/home/rho/Python/tensorflow/battleline/botserver.py",
		"--model_dir=/home/rho/Python/tensorflow/battleline/model1")
	err = serverCmd.Start()
	if err != nil {
		t.Fatalf("Command failed with error: %v", err)
	}
	fmt.Println("Sleep 5 second:")
	time.Sleep(time.Second * 5)
	fmt.Println("Up")
	tfCon, err := New("localhost:5555")
	if err != nil {
		stopCmd(serverCmd, t)
		t.Fatalf("Connecting to server failed with error: %v", err)
	}
	moveix := MoveHand(viewPos, tfCon)
	expMoveix := prob.MoveHand(viewPos)
	if moveix != expMoveix {
		t.Errorf("Move: %v deviate from expMove: %v", viewPos.Moves[moveix], viewPos.Moves[expMoveix])
	}
	stopCmd(serverCmd, t)
}

func stopCmd(cmd *exec.Cmd, t *testing.T) {
	sigErr := cmd.Process.Signal(syscall.SIGINT)
	if sigErr != nil {
		t.Errorf("Sending interupt signal failed with error: %v", sigErr)
	}
	_ = cmd.Wait() //TODO change to inspect error when close down works zmq4 python
	return
}
