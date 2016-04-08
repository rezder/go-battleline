package main

import (
	"log"
	"os"
	"os/signal"
	pub "rezder.com/game/card/battleline/server/publist"
	"rezder.com/game/card/battleline/server/tables"
)

func main() {
	errChan := make(chan error, 10)
	go errors(errChan)
	list := pub.New()
	startGameChan := tables.NewStartGameChan()
	finishTables := make(chan struct{})

	go tables.Start(startGameChan, list, finishTables)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	_ = <-stop
	close(startGameChan.Close)
	_ = <-finishTables
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
