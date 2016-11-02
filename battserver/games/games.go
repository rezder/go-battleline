//Package games contains the game server
//the games server consist of the tables server and players server.
package games

import (
	"github.com/rezder/go-battleline/battserver/players"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-battleline/battserver/tables"
)

//Server the game server structur.
type Server struct {
	tables  *tables.Server
	players *players.Server
}

//New Create a game server.
func New(errCh chan<- error, save bool, saveDir string) (g *Server, err error) {
	g = new(Server)
	list := pub.New()
	tables, err := tables.New(list, errCh, save, saveDir)
	if err != nil {
		return g, err
	}
	g.tables = tables
	g.players = players.New(list, g.tables.StartGameChCl)
	return g, err
}

//Start starts the game server.
func (g *Server) Start() {
	g.tables.Start()
	g.players.Start()
}

//Stop stops the game server.
func (g *Server) Stop() {
	g.players.Stop()
	g.tables.Stop()
}

//PlayersJoinCh returns the channel player join the players server.
func (g *Server) PlayersJoinCh() chan<- *players.Player {
	return g.players.JoinCh
}

//PlayersDisableCh returns the channel to disable or enable player on
//the players server.
func (g *Server) PlayersDisableCh() chan<- *players.DisData {
	return g.players.DisableCh
}
