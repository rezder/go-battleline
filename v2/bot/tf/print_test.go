package tf

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"io"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	filePathDb := "_test/testdb.db2"
	bdb := testOpenDb(filePathDb, t)
	defer func() {
		err := bdb.Close()
		if err != nil {
			t.Errorf("Close db failded")
		}
	}()
	var b []byte
	buf := bytes.NewBuffer(b)
	PrintGames(bdb, buf, 25)
	filePath := "_test/out.txt"
	var file *os.File
	var err error
	if _, existErr := os.Stat(filePath); os.IsNotExist(existErr) {
		file, err = os.Create(filePath)
		if err != nil {
			t.Errorf("Create File: %v failed: %v", filePath, err)
			return
		}
		_, err = buf.WriteTo(file)
		if err != nil {
			t.Errorf("Write to File: %v failed: %v", filePath, err)
		}
	} else {
		if existErr != nil {
			t.Errorf("File: %v failed: %v", filePath, existErr)
			return
		}
		file, err = os.Open(filePath)
		if err != nil {
			t.Errorf("Open File: %v falied: %v", filePath, err)
			return
		}
	}
	defer func() {
		err := file.Close()
		if err != nil {
			t.Errorf("Close File: %v falied: %v", filePath, err)
		}
	}()
	comparePrint(buf, file, t)
}
func comparePrint(buf io.Reader, file io.Reader, t *testing.T) {
	bufScanner := bufio.NewScanner(buf)
	fileScanner := bufio.NewScanner(file)
	isFileMore := true
	lineNo := 0
	for bufScanner.Scan() {
		lineNo++
		isFileMore = fileScanner.Scan()
		if !isFileMore {
			t.Error("New data is too long")
			break
		}
		oldTxt := fileScanner.Text()
		newTxt := bufScanner.Text()
		if oldTxt != newTxt {
			t.Errorf("Old and new output deviates on line %v\nOld: %v\nNew: %v", lineNo, oldTxt, newTxt)
		}
	}
	if isFileMore && fileScanner.Scan() {
		t.Error("New data is too short")
	}
	fmt.Printf("Out contain %v lines\n", lineNo)
}
func testOpenDb(filePath string, t *testing.T) (bdb *dbhist.Db) {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		t.Fatalf("Open database file: %v failed: %v", filePath, err)
	}
	bdb = dbhist.New(dbhist.KeyPlayersTime, db, 25)
	err = bdb.Init()
	if err != nil {
		_ = bdb.Close()
		t.Fatalf("Init database failed: %v", err)
	}
	return bdb
}
