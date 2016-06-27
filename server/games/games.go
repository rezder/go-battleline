package games

import (
	"rezder.com/game/card/battleline/server/players"
	pub "rezder.com/game/card/battleline/server/publist"
	"rezder.com/game/card/battleline/server/tables"
)

type Server struct {
	list       *pub.List
	tableChCl  *tables.StartGameChCl
	finTables  chan struct{}
	playerCh   chan *players.Player
	finPlayers chan struct{}
	errCh      chan<- error
	save       bool
	saveDir    string
}

func New(errCh chan<- error) (g *Server) {
	g = new(Server)
	g.list = pub.New()
	g.tableChCl = tables.NewStartGameChCl()
	g.finTables = make(chan struct{})
	g.playerCh = make(chan *players.Player)
	g.finPlayers = make(chan struct{})
	g.errCh = errCh
	return g
}
func (g *Server) Start(save bool, saveDir string) {
	go tables.Start(g.tableChCl, g.list, g.finTables, save, saveDir, g.errCh)
	go players.Start(g.playerCh, g.list, g.tableChCl, g.finPlayers)
}
func (g *Server) Stop() {
	close(g.playerCh)
	_ = g.finPlayers
	close(g.tableChCl.Close)
	_ = <-g.finTables
}
func (g *Server) PlayerCh() chan<- *players.Player {
	return g.playerCh
}
