package http

import (
	"github.com/boltdb/bolt"
	"os"
	"testing"
)

func TestCDb(t *testing.T) {
	filePath := "_test/testdb.go"
	cdb, err := NewCdb(filePath)
	if err != nil {
		t.Fatalf("Failed creating database with error:%v", err)
	}
	clients := testInsert(cdb, t)
	testGet(cdb, clients, t)
	testDisable(cdb, clients[1], t)
	err = cdb.Close()
	if err != nil {
		t.Errorf("Failed closed with error :%v", err)
	}
	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Failed remove file with error :%v", err)
	}
}
func testLogDb(t *testing.T, cdb *CDb) {
	err := cdb.db.View(func(tx *bolt.Tx) error {
		cb := tx.Bucket(cdb.clientBck)
		_ = cb.ForEach(func(k []byte, v []byte) error {
			client, _ := decode(v)
			t.Log(Btoi(k), client)
			return nil
		})
		ib := tx.Bucket(cdb.ixBck)
		_ = ib.ForEach(func(k []byte, v []byte) error {
			t.Log(string(k), Btoi(v))
			return nil
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
func testIndex(cdb *CDb, clients []*Client, t *testing.T) {
	for _, client := range clients {
		_, isUpd, err := cdb.UpdInsert(client)
		if err != nil {
			t.Errorf("Insert failed with Error: %v", err)
		}
		if isUpd {
			t.Error("Should not have updated")
		}
	}
}
func testGet(cdb *CDb, clients []*Client, t *testing.T) {
	for _, client := range clients {
		nameClient, isFound, err := cdb.GetName(client.Name)
		if err != nil {
			t.Errorf("GetName failed with Error: %v", err)
		}
		if !isFound {
			t.Errorf("Should found: %v", client.Name)
		}
		if !nameClient.IsEqual(client) {
			t.Errorf("Saved client deviates from get id. %v,%v ", nameClient, client)
		}
		idClient, isFound, err := cdb.GetID(client.ID)
		if err != nil {
			t.Errorf("GetID failed with Error: %v", err)
		}
		if !isFound {
			t.Errorf("Should found: %v", client.ID)
		}
		if !idClient.IsEqual(client) {
			t.Errorf("Saved client deviates from get id. %v,%v ", idClient, client)
		}
	}
}
func testInsert(cdb *CDb, t *testing.T) []*Client {
	clients := make([]*Client, 0, 2)
	client, _ := NewClient("Rene", "12345678")
	inClient, isUpd, err := cdb.UpdInsert(client)
	if err != nil {
		_ = cdb.Close()
		t.Fatalf("Insert failed with Error: %v", err)
	}
	if !isUpd {
		_ = cdb.Close()
		t.Fatal("Should have updated")
	}
	t.Logf("saved player name,id: %v,%v", inClient.Name, inClient.ID)
	clients = append(clients, inClient)
	client2, _ := NewClient("Peter", "12345678")
	var inClient2 *Client
	inClient2, isUpd, err = cdb.UpdInsert(client2)
	if err != nil {
		_ = cdb.Close()
		t.Fatalf("Insert failed with Error: %v", err)
	}
	if !isUpd {
		_ = cdb.Close()
		t.Fatal("Should have updated")
	}
	t.Logf("saved player name,id: %v,%v", inClient2.Name, inClient2.ID)
	testLogDb(t, cdb)
	clients = append(clients, inClient2)
	testIndex(cdb, clients, t)
	return clients
}

func testDisable(cdb *CDb, inClient2 *Client, t *testing.T) {
	name, isUpd, err := cdb.UpdDisable(inClient2.ID, true)
	if err != nil {
		t.Errorf("Upd disable failed with Error: %v", err)
	}
	if name != inClient2.Name {
		t.Errorf("Name deviates %v,%v", name, inClient2.Name)
	}
	if !isUpd {
		t.Error("Update disable failed")
	}
	sClient, isFound, err := cdb.GetName(inClient2.Name)
	if err != nil {
		t.Errorf("GetName failed with Error: %v", err)
	}
	if !isFound {
		t.Errorf("Should found: %v", inClient2.Name)
	}
	if sClient.IsEqual(inClient2) {
		t.Error("Saved client should")
	}
	if !sClient.IsDisable {
		t.Error("Saved client should be disable")
	}
}
