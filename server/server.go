// Run a battleline game server.
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"rezder.com/game/card/battleline/server/games"
	"rezder.com/game/card/battleline/server/html"
)

func main() {
	errCh := make(chan error, 10)
	finErrCh := make(chan struct{})
	go errServer(errCh, finErrCh)
	laddr, err := net.ResolveTCPAddr("tcp", ":8181")
	if err != nil {
		fmt.Println(err)
		return
	}
	netListener, err := net.ListenTCP("tcp", laddr)
	//netListener, err := net.ListenTCP("tcp", ":8181")
	if err != nil {
		fmt.Println(err)
		return
	}
	newGameServer := games.New(errCh)

	newGameServer.Start()
	clients := html.NewClients(newGameServer) //TODO load from file
	finHtmlCh := make(chan struct{})
	go html.Start(errCh, netListener, clients, finHtmlCh)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Server up and running. Close with ctrl+c")
	_ = <-stop
	log.Println("Server closed with interrupt signal")
	curGameServer := clients.SetGameServer(nil)
	netListener.Close()
	_ = <-finHtmlCh
	if curGameServer != nil {
		curGameServer.Stop()
	}
	close(errCh)
	_ = <-finErrCh
}
func errServer(errChan chan error, finCh chan struct{}) {
	for {
		err, open := <-errChan
		if open {
			log.Println("Error: ", err.Error())
		} else {
			close(finCh)
			break
		}
	}
}
