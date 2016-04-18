package main

import (
	"log"
	"os"
	"os/signal"
	"rezder.com/game/card/battleline/server/players"
	pub "rezder.com/game/card/battleline/server/publist"
	"rezder.com/game/card/battleline/server/tables"
)

func main() {
	errChan := make(chan error, 10)
	go errors(errChan)
	list := pub.New()

	startGameChan := tables.NewStartGameChCl()
	finTables := make(chan struct{})
	go tables.Start(startGameChan, list, finTables)

	playerCh := make(chan *players.Player)
	finPlayers := make(chan struct{})
	go players.Start(playerCh, list, startGameChan, finPlayers)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	_ = <-stop
	close(playerCh)
	close(startGameChan.Close)
	_ = <-finPlayers
	_ = <-finTables
	close(errChan)
}
func errors(errChan chan error) {
	for {
		err, open := <-errChan
		if open {
			log.Println("Error: ", err.Error())
		} else {
			break
		}
	}
}
