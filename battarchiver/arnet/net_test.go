package arnet

import (
	"bytes"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"io"
	"net"
	"testing"
)

func TestPoke(t *testing.T) {
	clientAdr := ":7171"
	myAddr := "peter:8080"
	go pokeListner(clientAdr, myAddr, t)
	PokeClient("localhost"+clientAdr, myAddr)
}

func pokeListner(clientAddr, myAddr string, t *testing.T) {
	ln, err := net.Listen("tcp", clientAddr)
	if err != nil {
		t.Errorf("Create listener failed: %v", err)
		return
	}
	defer ln.Close()
	conn, err := ln.Accept()
	if err != nil {
		t.Errorf("Accept failed: %v", err)
		return
	}
	defer conn.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, conn)
	if buf.String() != myAddr {
		t.Errorf("Addresses do not match expected: %v,got: %v\n", myAddr, buf.String())
	}
}
func TestSender(t *testing.T) {
	log.InitLog(3)
	rec, err := NewZmqReciver("7575")
	if err != nil {
		t.Fatalf("Receiver faled with %v", err)
	}
	rec.Start()
	sender, err := NewZmqSender("localhost:7575")
	if err != nil {
		cerr := rec.Close()
		t.Errorf("close receiver failed with: %v", cerr)
		t.Fatalf("Receiver faled with %v", err)
	}
	sender.Start()
	sender.GameCh <- bat.New([2]int{1, 2})
	game := <-rec.GameCh
	t.Log(game)
}
