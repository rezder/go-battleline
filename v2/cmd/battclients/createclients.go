package main

import (
	"flag"
	"fmt"
	"github.com/rezder/go-battleline/v2/http"
	"github.com/rezder/go-error/log"
	"os"
)

func main() {
	var dbfile string
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-dbfile] name password\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&dbfile, "dbfile", "clients.db", "The clients database file")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Print(log.Min, " Two arguments is need name and password")
		return
	}
	cdb, err := http.NewCdb(dbfile)
	if err != nil {
		log.PrintErr(err)
		return
	}
	defer func() {
		err = cdb.Close()
		if err != nil {
			log.PrintErr(err)
		}
	}()
	client, err := http.NewClient(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.PrintErr(err)
		return
	}
	client, isUpd, err := cdb.UpdInsert(client)
	if err != nil {
		log.PrintErr(err)
		return
	}
	if isUpd {
		log.Print(log.Min, "Client added")
	} else {
		log.Print(log.Min, "Client was not added as client allready exist")
	}
}
