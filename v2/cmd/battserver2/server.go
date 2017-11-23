// Run a battleline game server.
package main

import (
	"flag"
	//"github.com/pkg/profile"
	"github.com/rezder/go-battleline/v2/http"
	"github.com/rezder/go-error/log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Start the battleline server.
// The server needs the html/pages and the html/static files
// and will create a directory data with two gob files.
func main() {
	rand.Seed(time.Now().UnixNano())
	portFlag := flag.Int("port", 8282, "tcp port")
	archPokePortFlag := flag.Int("archpokeport", 7373, "Arciver poke tcp port, the archivers poke this port when ready")
	archAddrFlag := flag.String("archaddr", "", "Archiver address, if the archiver is allready running and ready")
	logFlag := flag.Int("loglevel", 0, "Log level 0 default lowest, 3 highest")
	//TODO change rootDirFlag default to ./server/htmlroot
	rootDirFlag := flag.String("rootdir", "/home/rho/js/batt-game-app/build/", "The server files root directory")
	//TODO make backup server to databases
	flag.Parse()
	log.InitLog(*logFlag)
	//defer profile.Start(profile.MemProfile, profile.NoShutdownHook).Stop()
	var port string
	if *portFlag == 80 { //Add https port if use https
		port = ""
	} else {
		port = ":" + strconv.Itoa(*portFlag)
	}
	httpServer, err := http.New(port, *archPokePortFlag, *archAddrFlag, *rootDirFlag)
	if err != nil {
		log.PrintErr(err)
		return
	}
	httpServer.Start()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Print(log.Min, "Server up and running. Close with ctrl+c")
	<-stop
	log.Print(log.Verbose, "Server closed with interrupt signal")
	httpServer.Stop()
}
