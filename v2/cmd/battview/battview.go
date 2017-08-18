package main

import (
	"flag"
	"github.com/rezder/go-battleline/v2/viewer"
	"github.com/rezder/go-error/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logLevelFlag := flag.Int("loglevel", 3, "Log level 0 default lowest, 3 highest")
	portFlag := flag.Int("port", 9021, "The http server port")

	flag.Parse()
	log.InitLog(*logLevelFlag)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Print(log.Min, "Server running. Close with ctrl+c\n")
	server, err := viewer.New(*portFlag)
	if err != nil {
		log.PrintErr(err)
		return
	}
	server.Start()
	<-stop
	server.Stop()
}
