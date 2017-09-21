package main

import (
	"flag"
	"fmt"
	"github.com/rezder/go-battleline/v2/game/conv"
	"github.com/rezder/go-error/log"
	"os"
	"strings"
)

const (
	suffixFile = "batt2"
	suffixDb   = "db2"
)

func main() {
	var isDb bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: src [dst]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.BoolVar(&isDb, "db", false, "Database file")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Println(log.Min, "You must specify the file to convert")
		return
	}
	if len(args) > 2 {
		log.Println(log.Min, "Max two arguments.")
		return
	}
	src := args[0]
	var dest string
	suffix := suffixFile
	if isDb {
		suffix = suffixDb
	}
	if len(args) == 2 {
		dest = args[1]
	} else {
		ix := strings.LastIndex(src, ".")
		if ix == -1 {
			dest = src + "." + suffix
		} else {
			dest = src[:ix+1] + suffix
		}
	}
	var err error
	if isDb {
		err = conv.DbFile(src, dest)
	} else {
		err = conv.GameFile(src, dest)
	}
	if err != nil {
		log.PrintErr(err)
	}
}
