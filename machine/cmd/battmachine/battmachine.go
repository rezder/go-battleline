package main

import (
	"bufio"
	"flag"
	"os"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	mach "github.com/rezder/go-battleline/machine"
	"github.com/rezder/go-error/log"
)

const (
	gameDbMaxFetch = 10000
	mposDbMaxFetch = 10000
)

func main() {
	vFlag := flag.Bool("v", false, "Verbose")
	gameDbFileFlag := flag.String("dbgamefile", "", "The database file to read game from.")
	outTypeFlag := flag.String("out", "Card", "The output type Card, Deck, Claim.")
	gameLimitFlag := flag.Int("limit", 200, "The max number of games to write.")
	sparseFlag := flag.Bool("sparse", true, "Sparse cathegories features.")
	flag.Parse()
	if *vFlag {
		log.InitLog(log.Debug)
	}
	args := flag.Args()
	posDbFile := "mposFile.db"
	if len(args) > 0 {
		posDbFile = args[0]
	}
	ok, outType := validateArgs(*outTypeFlag)
	if !ok {
		return
	}
	var bdb *battdb.Db
	if len(*gameDbFileFlag) > 0 {
		gameDb, err := bolt.Open(*gameDbFileFlag, 0600, nil)
		if err != nil {
			err = errors.Wrapf(err, "Open data base file %v failed", *gameDbFileFlag)
			log.PrintErr(err)
			return
		}
		defer gameDb.Close()
		bdb = battdb.New(battdb.KeyPlayersTime, gameDb, gameDbMaxFetch)
		err = bdb.Init()
		if err != nil {
			err = errors.Wrapf(err, "Init database %v failed", *gameDbFileFlag)
			log.PrintErr(err)
			return
		}
	}

	mposDb, err := bolt.Open(posDbFile, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", posDbFile)
		log.PrintErr(err)
		return
	}
	defer mposDb.Close()
	mdb := mach.NewDbPos(mposDb, mposDbMaxFetch)
	err = mdb.Init()
	if err != nil {
		err = errors.Wrapf(err, "Init database %v failed", posDbFile)
		log.PrintErr(err)
		return
	}

	if bdb != nil {
		mach.AddGames(bdb, mdb)
		if err != nil {
			log.PrintErr(err)
			return
		}
	}
	stdWriter := bufio.NewWriter(os.Stdout)
	mach.PrintMachineData(outType, *sparseFlag, stdWriter, mdb, *gameLimitFlag)
	stdWriter.Flush()
}

func validateArgs(outTypeArg string) (ok bool, outType int) {
	ok = true
	switch outTypeArg {
	case "Card":
		outType = mach.OutTypeMoveCard
	case "Deck":
		outType = mach.OutTypeDeck
	case "Claim":
		outType = mach.OutTypeClaim
	default:
		ok = false
		log.Printf(log.Min, "Illegal out argument only Std, Special, Deck, Claim is allowed.")
	}
	return ok, outType
}
