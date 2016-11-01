//The tables server.
//Keeping track of the games being played.
package tables

import (
	"encoding/gob"
	bat "github.com/rezder/go-battleline/battleline"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-error/cerrors"
	"log"
	"os"
	"strconv"
)

const (
	SAVE_GamesFile = "data/savegames.gob"
)

type Server struct {
	save          bool
	saveDir       string
	StartGameChCl *pub.StartGameChCl
	pubList       *pub.List
	doneCh        chan struct{}
	errCh         chan<- error
	saveGames     *SaveGames
}

func New(pubList *pub.List, errCh chan<- error, save bool, saveDir string) (s *Server, err error) {
	s = new(Server)
	s.pubList = pubList
	s.errCh = errCh
	s.save = save
	s.saveDir = saveDir
	s.StartGameChCl = pub.NewStartGameChCl()
	s.doneCh = make(chan struct{})
	bat.GobRegistor()
	s.saveGames, err = loadSaveGames()
	return s, err
}
func (s *Server) Start() {
	go start(s.StartGameChCl, s.pubList, s.doneCh, s.save, s.saveDir, s.errCh, s.saveGames)
}
func (s *Server) Stop() {
	if cerrors.IsVerbose() {
		log.Println("Closing start game channel on tables")
	}
	close(s.StartGameChCl.Close)
	<-s.doneCh
	if cerrors.IsVerbose() {
		log.Println("Receiving done from tables")
	}
}

//Start tables server.
//doneCh closing this channel will close down the tables server.
func start(startGameChCl *pub.StartGameChCl, pubList *pub.List, doneCh chan struct{},
	save bool, saveDir string, errCh chan<- error, saveGames *SaveGames) {
	finishTableCh := make(chan *FinishTableData)
	startCh := startGameChCl.Channel
	var done bool
	games := make(map[int]*pub.GameData)

Loop:
	for {
		select {
		case fin := <-finishTableCh:
			delete(games, fin.ids[0])
			delete(games, fin.ids[1])
			if fin.game != nil {
				saveGames.Games[gameId(fin.ids)] = fin.game
			}
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
					game, old := saveGames.Games[gameId(start.PlayerIds)]
					if old {
						delete(saveGames.Games, gameId(start.PlayerIds))
						if game.PlayerIds != start.PlayerIds {
							start.PlayerIds = [2]int{start.PlayerIds[1], start.PlayerIds[0]}
							start.PlayerChs = [2]chan<- *pub.MoveView{start.PlayerChs[1],
								start.PlayerChs[0]}
						}
						if game.Pos == nil {
							game.CalcPos()
						}
					}
					watch := pub.NewWatchChCl()
					go table(start.PlayerIds, start.PlayerChs, watch, game, finishTableCh, save,
						saveDir, errCh)
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
	} //loop
	noPosGames := saveGames.copyClearPos()
	err := noPosGames.save()
	if err != nil {
		errCh <- err
	}
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

// gameId makes a unique game id.
func gameId(players [2]int) (id string) {
	if players[0] > players[1] {
		id = strconv.Itoa(players[0]) + strconv.Itoa(players[1])
	} else {
		id = strconv.Itoa(players[1]) + strconv.Itoa(players[0])
	}
	return id
}

//FinishTableData the data structur send on the finish channel.
type FinishTableData struct {
	ids  [2]int
	game *bat.Game
}

//SaveGames the data structur for saved games used to save as Gob.
type SaveGames struct {
	Games map[string]*bat.Game
}

func NewSaveGames() (sg *SaveGames) {
	sg = new(SaveGames)
	sg.Games = make(map[string]*bat.Game)
	return sg
}

//copyClearPos makes a copy of the SaveGames with out the game position
//The game is not deep copied which mean the Moves slice array is
//still connected.
func (games *SaveGames) copyClearPos() (c *SaveGames) {
	c = NewSaveGames()
	if len(games.Games) > 0 {
		for k, v := range games.Games {
			g := *v
			g.Pos = nil
			c.Games[k] = &g
		}
	}
	return c
}
func (games *SaveGames) save() (err error) {
	file, err := os.Create(SAVE_GamesFile)
	if err != nil {
		err = cerrors.Wrap(err, 15, "Create games file")
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(games)
	return err
}
func loadSaveGames() (games *SaveGames, err error) {
	file, err := os.Open(SAVE_GamesFile)
	if err == nil {
		defer file.Close()
		decoder := gob.NewDecoder(file)
		lg := *NewSaveGames()
		err = decoder.Decode(&lg)
		if err == nil {
			games = &lg
		}
	} else {
		if os.IsNotExist(err) {
			err = nil
			games = NewSaveGames() //first start
		} else {
			err = cerrors.Wrap(err, 16, "Open saved games file")
			return games, err
		}
	}
	return games, err
}
