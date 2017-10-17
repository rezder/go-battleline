package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	arch "github.com/rezder/go-battleline/v2/archiver"
	"github.com/rezder/go-error/log"
)

func main() {
	portFlag := flag.String("port", "7272", "Archiver port")
	backupPortFlag := flag.String("backupport", "", "Back http server port. No port no server")
	clientFlag := flag.String("client", "", "Url of client without protecol if specified client is poked when the server is ready")
	myAddrFlag := flag.String("addr", "arch", "Archiver addr only used if client is specified, port is added to the address")
	logFlag := flag.Int("loglevel", 0, "Log level 0 default lowest, 3 highest")
	dbFileFlag := flag.String("dbfile", "bdb.db2", "The database file name")
	flag.Parse()
	log.InitLog(*logFlag)

	db, err := bolt.Open(*dbFileFlag, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", *dbFileFlag)
		log.PrintErr(err)
		return
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.PrintErr(cerr)
		}
	}()
	server, err := arch.NewServer(db, *portFlag)
	if err != nil {
		err = errors.Wrapf(err, "Creating archiver server failed on port %v ", *portFlag)
		log.PrintErr(err)
		return
	}
	server.Start(*backupPortFlag, *clientFlag, *myAddrFlag)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Print(log.Min, "Server up and running. Close with ctrl+c")
	<-stop
	log.Print(log.DebugMsg, "Recieved interupt signal or terminate closing down")
	server.Stop()
	log.Print(log.DebugMsg, "Closed down")
}
