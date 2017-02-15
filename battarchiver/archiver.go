package main

import (
	"flag"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/arnet"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	portFlag := flag.String("port", "7070", "Archiver port")
	backupPortFlag := flag.String("backupport", "", "Back http server port. No port no server")
	clientFlag := flag.String("client", "", "Url of client without protecol if specified client is poked when the server is ready")
	myAddrFlag := flag.String("addr", "arch", "Archiver addr only used if client is specified, port is added to the address")
	logFlag := flag.Int("loglevel", 0, "Log level 0 default lowest, 3 highest")
	dbFileFlag := flag.String("dbfile", "bdb.db", "The database file name")
	flag.Parse()
	log.InitLog(*logFlag)

	bd, err := bolt.Open(*dbFileFlag, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", *dbFileFlag)
		log.PrintErr(err)
		return
	}
	defer bd.Close()
	bdb := battdb.New(battdb.KeyPlayersTime, bd, 10000)
	err = bdb.Init()
	if err != nil {
		err = errors.Wrapf(err, "Init database %v failed", *dbFileFlag)
		log.PrintErr(err)
		return
	}
	nz, err := arnet.NewZmqReciver(*portFlag)
	if err != nil {
		err = errors.Wrapf(err, "Creating zmq game listener on port %v failed", *portFlag)
		log.PrintErr(err)
		return
	}
	defer nz.Close()

	if len(*backupPortFlag) != 0 {
		go arnet.StartBackUpServer(bdb, *backupPortFlag)
	}
	saveFinCh := make(chan struct{})
	go startSaveServer(bdb, nz.GameCh, saveFinCh)
	nz.Start()

	if len(*clientFlag) > 0 {
		addr := *myAddrFlag + ":" + *portFlag
		arnet.PokeClient(*clientFlag, addr)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Print(log.Min, "Server up and running. Close with ctrl+c")
	<-stop
	nz.Close()
	<-saveFinCh
}

func startSaveServer(bdb *battdb.Db, gameCh <-chan []byte, finCh chan<- struct{}) {
	noSaved := 0
Loop:
	for {
		no := cropNo(len(gameCh))
		games := make([]*bat.Game, 0, no)
		for i := 0; i < no; i++ {
			gameBytes, open := <-gameCh
			if open {
				game, err := arnet.ZmqDecoder(gameBytes)
				if err != nil {
					err = errors.Wrap(err, "Decoding game msg failed") //TODO do not need stack remove with new interface
					log.PrintErr(err)
				} else {
					games = append(games, game)
				}
			} else {
				break Loop
			}
		}
		_, err := bdb.Puts(games)
		if err != nil {
			log.PrintErr(err)
		} else {
			noSaved = noSaved + len(games)
			if noSaved == 1 || noSaved%50 == 0 {
				log.Printf(log.DebugMsg, "Number of saved games: %v\n", noSaved)
			}
		}

	}
	close(finCh)
}
func cropNo(no int) (updNo int) {
	noMax := 100
	updNo = 1
	if no > noMax {
		updNo = noMax
	} else if no > 0 {
		updNo = no
	}
	return updNo
}
