package http

import (
	"github.com/rezder/go-battleline/v2/http/games"
	"os"
	"testing"
)

const (
	testPW = "12345678"
)

func TestClients(t *testing.T) {
	dbClientsFILE = "_test/clients.db"
	gameServer, err := games.New(7373)
	if err != nil {
		t.Fatalf("Init games server failed, with error: %v", err)
	}
	clients, err := NewClients(gameServer)
	if err != nil {
		_ = gameServer.Cancel()
		t.Fatalf("Init clients failed, with error: %v", err)
	}
	testCreateClientValid(clients, "Hans", testPW, t)
	testCreateClientValid(clients, "Peter", testPW, t)
	logIns := make([]*Client, 0, len(clients.logIns))
	for _, client := range clients.logIns {
		logIns = append(logIns, client)
	}
	testCreateClientInValid(clients, logIns, t)
	testVerifySid(clients, logIns, t)
	testLogInOut(clients, logIns, t)
	//clients.LogOut(client.Name)

	_ = clients.SetGameServer(nil)
	testDown(clients, logIns, t)
	err = gameServer.Cancel()
	if err != nil {
		t.Errorf("Cancel game server failed, with error: %v", err)
	}
	err = clients.Close()
	if err != nil {
		t.Errorf("Closed clients failed, with error: %v", err)
	}
	err = os.Remove(dbClientsFILE)
	if err != nil {
		t.Errorf("Remove file %v failed, with error: %v", dbClientsFILE, err)
	}
}
func testClientDis(clients *Clients, logIns []*Client, t *testing.T) {
	if len(logIns) > 0 {
		client := logIns[0]
		clients.LogOut(client.Name)
		err := clients.UpdateDisable(client.ID, true)
		if err != nil {
			t.Errorf("Disable client failed, with error: %v", err)
			return
		}
		status, sid, err := clients.LogIn(client.Name, testPW)
		t.Log(status, sid, err)
		if !status.IsDisable() {
			t.Error("Login should have failed with disable")
			return
		}
		err = clients.UpdateDisable(client.ID, false)
		if err != nil {
			t.Errorf("Disable set to false failed, with error: %v", err)
			return
		}
		status, sid, err = clients.LogIn(client.Name, testPW)
		t.Log(status, sid, err)
		if !status.IsOk() {
			t.Error("Login should have succed")
		}
	}
}
func testDown(clients *Clients, logIns []*Client, t *testing.T) {
	for _, client := range logIns {
		ok, isDown := clients.VerifySid(client.Name, client.sid)
		t.Log(ok, isDown)
		if !isDown {
			t.Error("Server should be down")
		}
	}
}
func testLogInOut(clients *Clients, logIns []*Client, t *testing.T) {
	for _, client := range logIns {
		status, _, err := clients.LogIn(client.Name, testPW)
		if err != nil {
			t.Errorf("login client failed, with error: %v", err)
		}
		if !status.IsExist() {
			t.Error("Loging should have failed with allready login")
		}
		clients.LogOut(client.Name)
	}
	for _, client := range logIns {
		status, sid, err := clients.LogIn(client.Name, "123456789")
		t.Log(status, sid, err)
		if !status.IsInValid() {
			t.Error("Loging should have failed with invalid pasword")
		}
		if err != nil {
			t.Errorf("login client failed, with error: %v", err)
		}
	}
	for _, client := range logIns {
		status, sid, err := clients.LogIn(client.Name, testPW)
		t.Log(status, sid, err)
		if !status.IsOk() {
			t.Error("Loging failed after logout")
		}
		if err != nil {
			t.Errorf("login client failed, with error: %v", err)
		}
	}
}
func testVerifySid(clients *Clients, logIns []*Client, t *testing.T) {
	for _, client := range logIns {
		ok, isDown := clients.VerifySid(client.Name, client.sid)
		t.Log(ok, isDown)
		if !ok {
			t.Error("Verify should not fail")
		}
		ok, _ = clients.VerifySid(client.Name, "arg")
		if ok {
			t.Error("Verify should fail")
		}
	}
}
func testCreateClientInValid(clients *Clients, logIns []*Client, t *testing.T) {
	for _, client := range logIns {
		status, sid, err := clients.AddNew(client.Name, testPW)
		t.Log(status, sid, err)
		if !status.IsExist() {
			t.Error("Creating existing client should have failed")
		}
	}
}
func testCreateClientValid(clients *Clients, name, pw string, t *testing.T) {
	status, sid, err := clients.AddNew(name, pw)
	t.Log(status, sid, err)
	if err != nil {
		t.Errorf("Creating client: %v failed with error: %v", name, err)
	}
	if !status.IsOk() {
		t.Errorf("Failed to create client %v", name)
	}
}
