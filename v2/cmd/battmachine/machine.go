package main

import (
	"flag"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/bot/tf"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-error/log"
	"os"
)

const (
	dbMaxFetch = 1000
)

func main() {
	dbFlag := flag.String("dbfile", "bdb.db2", "The database file to read game hists from.")
	outFileFlag := flag.String("outfile", "out.cvs", "Out put file, truncated if exist")
	gameLimitFlag := flag.Int("limit", 200, "The max number of games to write.")
	flag.Parse()
	db, err := bolt.Open(*dbFlag, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", *dbFlag)
		log.PrintErr(err)
		return
	}
	defer func() {
		cerr := db.Close()
		if cerr != nil {
			log.PrintErr(cerr)
		}
	}()
	bdb := dbhist.New(dbhist.KeyPlayersTime, db, dbMaxFetch)
	err = bdb.Init()
	if err != nil {
		err = errors.Wrapf(err, "Init database %v failed", *dbFlag)
		log.PrintErr(err)
		return
	}
	file, err := os.Create(*outFileFlag)
	if err != nil {
		err = errors.Wrapf(err, "Create File: %v failed.", *outFileFlag)
		log.PrintErr(err)
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			err = errors.Wrapf(err, "Close File: %v failed.", *outFileFlag)
			log.PrintErr(err)
		}
	}()
	tf.PrintGames(bdb, file, *gameLimitFlag)
}
