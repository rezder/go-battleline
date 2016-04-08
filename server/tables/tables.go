package tables

import (
	bat "rezder.com/game/card/battleline"
	pub "rezder.com/game/card/battleline/server/publist"
	"strconv"
)

//Start tables server.
//doneGame should be close when the server should stop.
//doneServer will be closed when the server is done.
func Start(startGameChan *StartGameChan, pubList *pub.List, doneServer chan struct{}) {
	finishTable := make(chan [2]int)
	startChan := startGameChan.Channel
	var done bool
	games := make(map[int]*pub.GameData)
	oldGames := make(map[string]*bat.Game) //TODO load it from file
Loop:
	for {
		select {
		case fin := <-finishTable:
			delete(games, fin[0])
			delete(games, fin[1])
			if done && len(games) == 0 {
				break Loop
			}
			publish(games, pubList)
		case start := <-startChan:
			_, found := games[start.PlayerIds[0]]
			if found {
				close(start.PlayerChans[0])
				close(start.PlayerChans[1])
			} else {
				_, found = games[start.PlayerIds[1]]
				if found {
					close(start.PlayerChans[0])
					close(start.PlayerChans[1])
				} else {
					game, old := oldGames[gameId(start.PlayerIds)]
					if old {
						delete(oldGames, gameId(start.PlayerIds))
					}
					watch := pub.NewWatchChan()
					go table(start.PlayerIds, start.PlayerChans, watch, game, finishTable)
					games[start.PlayerIds[0]] = NewGameData(start.PlayerIds[1], watch)
					games[start.PlayerIds[1]] = NewGameData(start.PlayerIds[0], watch)
					publish(games, pubList)
				}
			}
		case <-startGameChan.Close:
			if len(games) == 0 {
				break Loop
			} else {
				startChan = nil
				done = true
			}
		}
	}
	//TODO save oldGames to file
	close(doneServer)
}
func publish(games map[int]*pub.GameData, pubList *pub.List) {
	copy := make(map[int]*pub.GameData)
	for key, v := range games {
		copy[key] = v
	}
	go pubList.UpdateGames(copy)
}

func NewGameData(opp int, watch *pub.WatchChan) (g *pub.GameData) {
	g = new(pub.GameData)
	g.Opp = opp
	g.Watch = watch
	return g
}

type StartGameData struct {
	PlayerIds   [2]int
	PlayerChans [2]chan<- *pub.MoveView
}
type StartGameChan struct {
	Channel chan *StartGameData
	Close   chan struct{}
}

func NewStartGameChan() (sgc *StartGameChan) {
	sgc = new(StartGameChan)
	sgc.Channel = make(chan *StartGameData)
	sgc.Close = make(chan struct{})
	return sgc
}

func gameId(players [2]int) (id string) {
	if players[0] > players[1] {
		id = strconv.Itoa(players[0]) + strconv.Itoa(players[1])
	} else {
		id = strconv.Itoa(players[1]) + strconv.Itoa(players[0])
	}
	return id
}
