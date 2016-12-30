// Run a battleline game server.
package main

import (
	"flag"
	"github.com/rezder/go-battleline/battserver/html"
	"github.com/rezder/go-error/cerrors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Server up and running. Close with ctrl+c")
	<-stop
	if cerrors.IsVerbose() {
		log.Println("Server closed with interrupt signal")
	}
	httpServer.Stop()
	close(errCh)
	<-finErrCh

}

//errServer start a error server.
//all errors should be send where the power to close down exist.
//Currently it does nothing but log the errors.
func errServer(errChan chan error, finCh chan struct{}) {
	//TODO MAYBE move the error server to html
	// Add error count on player id and auto disable player with to many errors.
	// disable must not be possible during save and error log should be active to the
	// end. This leaves two possiblities tell error server to stop disable players during close
	// or return fail when call disable during save/close, because error channel is buffered
	// it is not enough to stop the servers that produce the errors we need acttive stop of the
	// error server for disable.
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
