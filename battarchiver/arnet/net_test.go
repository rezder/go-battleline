package arnet

import (
	"bytes"
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
