package main

import (
	"flag"
	"github.com/boltdb/bolt"
	"github.com/rezder/go-battleline/battarchiver/arnet"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/cerrors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	portFlag := flag.String("port", "7070", "Archiver port")
	backupPortFlag := flag.String("backupport", "", "Back http server port. No port no server")
	clientFlag := flag.String("client", "", "Url of client without protecol if specified client is poked when the server is ready")
	myAddrFlag := flag.String("addr", "arch", "Archiver addr only used if client is specified, port is added to the address")
	logFlag := flag.Int("loglevel", 2, "Log level 0 default lowest, 2 highest") //TODO change default 0
	dbFileFlag := flag.String("dbfile", "bdb.db", "The database file name")
	flag.Parse()
	cerrors.InitLog(*logFlag)

	bd, err := bolt.Open(*dbFileFlag, 0600, nil)
	if err != nil {
		log.Printf("Open data base file failed %v", err)
		return
	}
	defer bd.Close()
	bdb := battdb.New(battdb.KeyPlayersTime, bd, 10000)
	err = bdb.Init()
	if err != nil {
		log.Printf("Init database failed %v", err)
		return
	}
	nz, err := arnet.NewZmqReciver(*portFlag)
	if err != nil {
		log.Printf("Creating net listener failed %v", err)
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
	log.Println("Server up and running. Close with ctrl+c")
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
					log.Printf("Decoding ms falid %v", err)
				} else {
					games = append(games, game)
				}
			} else {
				break Loop
			}
		}
		_, err := bdb.Puts(games)
		if err != nil {
			log.Println("Saving games in database failded %v", err)
		} else {
			noSaved = noSaved + len(games)
			if noSaved == 1 || noSaved%50 == 0 {
				if cerrors.LogLevel() == cerrors.LOG_Debug {
					log.Printf("Number of saved games: %v\n", noSaved)
				}
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
