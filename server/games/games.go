package games

import (
	"rezder.com/game/card/battleline/server/players"
	pub "rezder.com/game/card/battleline/server/publist"
	"rezder.com/game/card/battleline/server/tables"
)

type Server struct {
	tables  *tables.Server
	players *players.Server
}

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
func (g *Server) Start() {
	g.tables.Start()
	g.players.Start()
}
func (g *Server) Stop() {
	g.players.Stop()
	g.tables.Stop()
}
func (g *Server) PlayersJoinCh() chan<- *players.Player {
	return g.players.JoinCh
}
