package archiver

import (
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/archiver/arnet"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
	"net/http"
)

//Server is archiving battleline histories.
type Server struct {
	histDb       *dbhist.Db
	archReceiver *arnet.Receiver
	finishedCh   chan struct{}
	port         string
}

//NewServer create archiver server.
func NewServer(db *bolt.DB, port string) (server *Server, err error) {
	server = new(Server)
	dbh := dbhist.New(dbhist.KeyPlayers, db, 500)
	err = dbh.Init()
	if err != nil {
		return server, err
	}
	server.histDb = dbh
	zmqReceiver, err := arnet.NewReceiver(port)
	if err != nil {
		return server, err
	}
	server.archReceiver = zmqReceiver
	server.finishedCh = make(chan struct{})
	return server, err
}

//Start starts the server.
func (server *Server) Start(backupPort, clientURL, addr string) {
	if len(backupPort) != 0 {
		go backUpServe(server.histDb, backupPort)
	}
	go startSaveServe(server.histDb, server.archReceiver.HistCh, server.finishedCh)
	server.archReceiver.Start()

	if len(clientURL) > 0 {
		archURL := addr + ":" + server.port
		arnet.PokeClient(clientURL, archURL)
	}
}

//Stop stops the server.
func (server *Server) Stop() {
	err := server.archReceiver.Close()
	if err != nil {
		err = errors.Wrap(err, "Closing archiver failed.")
		log.PrintErr(err)
	}
	<-server.finishedCh
}

func startSaveServe(hdb *dbhist.Db, histCh <-chan []byte, finCh chan<- struct{}) {
	noSaved := 0
Loop:
	for {
		no := cropNo(len(histCh))
		hists := make([]*bg.Hist, 0, no)
		for i := 0; i < no; i++ {
			histBytes, open := <-histCh
			if open {
				hist, err := arnet.HistDecoder(histBytes)
				if err != nil {
					err = errors.WithMessage(err, "Decoding hist msg failed")
					log.PrintErr(err)
				} else {
					hists = append(hists, hist)
				}
			} else {
				break Loop
			}
		}
		err := hdb.Puts(hists)
		if err != nil {
			log.PrintErr(err)
		} else {
			noSaved = noSaved + len(hists)
			if noSaved == 1 || noSaved%50 == 0 {
				log.Printf(log.DebugMsg, "Number of saved game histories: %v\n", noSaved)
			}
		}

	}
	log.Printf(log.DebugMsg, "Final number of saved game histories: %v\n", noSaved)
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

//backUpServe serves http backups
func backUpServe(hdb *dbhist.Db, port string) {
	http.HandleFunc("/backup",
		func(resp http.ResponseWriter, req *http.Request) {
			hdb.BackupHandleFunc(resp, req)
		})
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		err = errors.Wrap(err, "Backup http server failed")
		log.PrintErr(err)
	}
}
