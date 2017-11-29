package http

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-error/log"
	"net/http"
	"strconv"
)

//CDb a client database.
type CDb struct {
	clientBck []byte
	ixBck     []byte
	db        *bolt.DB
}

//NewCdb creates a new client database.
func NewCdb(file string) (cdb *CDb, err error) {
	cdb = new(CDb)
	cdb.ixBck = []byte("ClientsIxBucket")
	cdb.clientBck = []byte("ClientsBucket")
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", file)
		return cdb, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, txErr := tx.CreateBucketIfNotExists(cdb.clientBck)
		if txErr != nil {
			return errors.Wrapf(err, log.ErrNo(1)+"Creating bucket %v", string(cdb.clientBck))
		}
		_, txErr = tx.CreateBucketIfNotExists(cdb.ixBck)
		if txErr != nil {
			return errors.Wrapf(err, log.ErrNo(1)+"Creating bucket %v", string(cdb.ixBck))
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
	}
	cdb.db = db
	return cdb, err
}

//Close closes the database.
func (cdb *CDb) Close() error {
	return cdb.db.Close()
}

//GetName returns a client.
func (cdb *CDb) GetName(name string) (client *Client, isFound bool, err error) {
	var clientBs []byte
	err = cdb.db.View(func(tx *bolt.Tx) error {
		ixBck := tx.Bucket(cdb.ixBck)
		keyBs := ixBck.Get([]byte(name))
		if keyBs != nil {
			clientBck := tx.Bucket(cdb.clientBck)
			clientBs = clientBck.Get(keyBs)
		}
		return nil
	})
	if err != nil {
		return client, isFound, err
	}
	if clientBs != nil {
		isFound = true
		client, err = decode(clientBs)
	}
	return client, isFound, err
}

//GetID returns a client.
func (cdb *CDb) GetID(id int) (client *Client, isFound bool, err error) {
	var clientBs []byte
	err = cdb.db.View(func(tx *bolt.Tx) error {
		clientBck := tx.Bucket(cdb.clientBck)
		clientBs = clientBck.Get(itob(id))
		return nil
	})
	if err != nil {
		return client, isFound, err
	}
	if clientBs != nil {
		isFound = true
		client, err = decode(clientBs)
	}
	return client, isFound, err
}

//UpdDisable update a clients disable field.
func (cdb *CDb) UpdDisable(id int, isDisable bool) (name string, isUpd bool, err error) {

	err = cdb.db.Update(func(tx *bolt.Tx) (updErr error) {
		clientBck := tx.Bucket(cdb.clientBck)
		clientBs := clientBck.Get(itob(id))
		if clientBs == nil {
			updErr = errors.WithStack(fmt.Errorf("Client with id %v does not exist", id))
			return updErr
		}
		client, updErr := decode(clientBs)
		if updErr != nil {
			return updErr
		}
		if client.IsDisable != isDisable {
			client.IsDisable = isDisable
			clientBs, updErr = encode(client)
			if err != nil {
				return updErr
			}
			updErr = clientBck.Put(itob(client.ID), clientBs)
			if updErr != nil {
				return updErr
			}
			isUpd = true
			name = client.Name
		}

		return updErr
	})

	return name, isUpd, err
}

//BackupHandleFunc handles http back up requests.
func (cdb *CDb) BackupHandleFunc(w http.ResponseWriter, req *http.Request) {
	err := cdb.db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//UpdInsert inserts a client if it does not allready exist.
func (cdb *CDb) UpdInsert(inClient *Client) (updClient *Client, isUpd bool, err error) {
	client := inClient.Copy()
	err = cdb.db.Update(func(tx *bolt.Tx) (updErr error) {
		ixBck := tx.Bucket(cdb.ixBck)
		keyBs := ixBck.Get([]byte(client.Name))
		if keyBs == nil {
			clientBck := tx.Bucket(cdb.clientBck)
			id, _ := clientBck.NextSequence()
			client.ID = int(id)
			var clientBs []byte
			clientBs, updErr = encode(client)
			if updErr != nil {
				return updErr
			}
			keyBs = itob(client.ID)
			updErr = clientBck.Put(keyBs, clientBs)
			if updErr != nil {
				return updErr
			}
			updErr = ixBck.Put([]byte(client.Name), keyBs)
			if updErr != nil {
				return updErr
			}
			isUpd = true
			updClient = client
		}
		return updErr
	})
	return updClient, isUpd, err
}
func encode(client *Client) (bs []byte, err error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err = encoder.Encode(client)
	if err != nil {
		return bs, err
	}
	bs = buf.Bytes()
	return bs, err
}
func decode(bs []byte) (client *Client, err error) {
	buf := bytes.NewBuffer(bs)
	decoder := gob.NewDecoder(buf)
	c := *new(Client)
	err = decoder.Decode(&c)
	if err != nil {
		return client, err
	}
	client = &c
	return client, err
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Btoi returns an 8-byte big endian representation of v.
func Btoi(bs []byte) int {
	ui := binary.BigEndian.Uint64(bs)
	return int(ui)
}
