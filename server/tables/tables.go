package tables

import (
	bat "rezder.com/game/card/battleline"
	pub "rezder.com/game/card/battleline/server/publist"
	"strconv"
)

//Start tables server.
//doneCh closing this channel will close down the tables server.
func Start(startGameChCl *StartGameChCl, pubList *pub.List, doneCh chan struct{}) {
	finishTableCh := make(chan [2]int)
	startCh := startGameChCl.Channel
	var done bool
	games := make(map[int]*pub.GameData)
	oldGames := make(map[string]*bat.Game) //TODO load it from file
Loop:
	for {
		select {
		case fin := <-finishTableCh:
			delete(games, fin[0])
			delete(games, fin[1])
			if done && len(games) == 0 {
				break Loop
			}
			publish(games, pubList)
		case start := <-startCh:
			_, found := games[start.PlayerIds[0]]
			if found {
				close(start.PlayerChs[0])
				close(start.PlayerChs[1])
			} else {
				_, found = games[start.PlayerIds[1]]
				if found {
					close(start.PlayerChs[0])
					close(start.PlayerChs[1])
				} else {
					game, old := oldGames[gameId(start.PlayerIds)]
					if old {
						delete(oldGames, gameId(start.PlayerIds))
					}
					watch := pub.NewWatchChCl()
					go table(start.PlayerIds, start.PlayerChs, watch, game, finishTableCh)
					games[start.PlayerIds[0]] = NewGameData(start.PlayerIds[1], watch)
					games[start.PlayerIds[1]] = NewGameData(start.PlayerIds[0], watch)
					publish(games, pubList)
				}
			}
		case <-startGameChCl.Close:
			if len(games) == 0 {
				break Loop
			} else {
				startCh = nil
				done = true
			}
		}
	}
	//TODO save oldGames to file
	close(doneCh)
}

// publish the current games list.
func publish(games map[int]*pub.GameData, pubList *pub.List) {
	copy := make(map[int]*pub.GameData)
	for key, v := range games {
		copy[key] = v
	}
	go pubList.UpdateGames(copy)
}

//NewGameData create a new GameData pointer.
func NewGameData(opp int, watch *pub.WatchChCl) (g *pub.GameData) {
	g = new(pub.GameData)
	g.Opp = opp
	g.Watch = watch
	return g
}

// StartGameData is the information need to start a game.
type StartGameData struct {
	PlayerIds [2]int
	PlayerChs [2]chan<- *pub.MoveView
}

// StartGameChCl the start game channel.
type StartGameChCl struct {
	Channel chan *StartGameData
	Close   chan struct{}
}

func NewStartGameChCl() (sgc *StartGameChCl) {
	sgc = new(StartGameChCl)
	sgc.Channel = make(chan *StartGameData)
	sgc.Close = make(chan struct{})
	return sgc
}

// gameId makes a unique game id.
func gameId(players [2]int) (id string) {
	if players[0] > players[1] {
		id = strconv.Itoa(players[0]) + strconv.Itoa(players[1])
	} else {
		id = strconv.Itoa(players[1]) + strconv.Itoa(players[0])
	}
	return id
}
