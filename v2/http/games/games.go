//Package games contains the game server
//the games server consist of the tables server and players server.
package games

import (
	"golang.org/x/net/websocket"
)

//Server the game server structur.
type Server struct {
	tables  *TablesServer
	players *PlayersServer
}

//New Create a game server.
func New(archPokePort int, archAddr string) (g *Server, err error) {
	g = new(Server)
	list := NewList()
	tables, err := NewTablesServer(list, archPokePort, archAddr)
	if err != nil {
		return g, err
	}
	g.tables = tables
	g.players = NewPlayersServer(list, g.tables.StartGameChCl)
	return g, err
}

//Cancel the server, must be called if not starting the server.
func (g *Server) Cancel() error {
	return g.tables.CloseDb()
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

//JoinClient asks the server to add player to the game server.
func (g *Server) JoinClient(
	id int,
	name string,
	ws *websocket.Conn,
	errCh chan<- error,
	joinedCh chan<- *Player) {
	player := NewPlayer(id, name, ws, errCh, joinedCh)
	g.players.JoinCh <- player
}

//BootPlayer asks the server to boot a player
func (g *Server) BootPlayer(playerID int) {
	g.players.DisableCh <- &PlayersDisData{Disable: true, PlayerID: playerID}
}

//BootPlayerStop asks the server to stop booting a player
func (g *Server) BootPlayerStop(playerID int) {
	g.players.DisableCh <- &PlayersDisData{Disable: false, PlayerID: playerID}
}
