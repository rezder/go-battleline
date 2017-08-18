//Package games contains the game server
//the games server consist of the tables server and players server.
package games

//Server the game server structur.
type Server struct {
	tables  *TablesServer
	players *PlayersServer
}

//New Create a game server.
func New(archiverPort int) (g *Server, err error) {
	g = new(Server)
	list := NewList()
	tables, err := NewTablesServer(list, archiverPort)
	if err != nil {
		return g, err
	}
	g.tables = tables
	g.players = NewPlayersServer(list, g.tables.StartGameChCl)
	return g, err
}

//Start starts the game server.
func (g *Server) Start(errCh chan<- error) {
	g.tables.Start(errCh)
	g.players.Start()
}

//Stop stops the game server.
func (g *Server) Stop() {
	g.players.Stop()
	g.tables.Stop()
}

//PlayersJoinCh returns the channel player join the players server.
func (g *Server) PlayersJoinCh() chan<- *Player {
	return g.players.JoinCh
}

//PlayersDisableCh returns the channel to disable or enable player on
//the players server.
func (g *Server) PlayersDisableCh() chan<- *PlayersDisData {
	return g.players.DisableCh
}
