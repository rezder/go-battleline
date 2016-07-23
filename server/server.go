// Run a battleline game server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"rezder.com/cerrors"
	"rezder.com/game/card/battleline/server/html"
	"strconv"
)

// Start the battleline server.
// The server needs the html/pages and the html/static files
// and will create a directory data with two gob files.
func main() {
	portFlag := flag.Int("port", 8181, "tcp port")
	saveFlag := flag.Bool("save", false, "Save games.")
	saveDirFlag := flag.String("savedir", "temp", "Save game directory")
	logFlag := flag.Int("loglevel", 0, "Log level 0 default lowest, 2 highest")
	flag.Parse()
	cerrors.InitLog(*logFlag)
	var port string
	if *portFlag == 80 { //Add https port if use https
		port = ""
	} else {
		port = ":" + strconv.Itoa(*portFlag)
	}
	errCh := make(chan error, 10)
	httpServer, err := html.New(errCh, port, *saveFlag, *saveDirFlag)
	if err != nil {
		if cerrors.LogLevel() != cerrors.LOG_Debug {
			log.Printf("Create server fail. Error: %v\n", err)
		} else {
			log.Printf("Create server fail. Error: %+v\n", err)
		}
		return
	}
	finErrCh := make(chan struct{})
	go errServer(errCh, finErrCh)
	httpServer.Start()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Server up and running. Close with ctrl+c")
	_ = <-stop
	if cerrors.IsVerbose() {
		log.Println("Server closed with interrupt signal")
	}
	httpServer.Stop()
	close(errCh)
	_ = <-finErrCh

}

//errServer start a error server.
//all errors should be send here where the power to close down exist.
//Currently it does nothing but log the errors.
func errServer(errChan chan error, finCh chan struct{}) {
	for {
		err, open := <-errChan
		if open {
			if cerrors.LogLevel() != cerrors.LOG_Debug {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Error: %+v", err)
			}
		} else {
			close(finCh)
			break
		}
	}
}
