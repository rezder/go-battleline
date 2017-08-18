package games

import (
	"encoding/gob"
	"github.com/pkg/errors"
	arch "github.com/rezder/go-battleline/battarchiver/client"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"os"
	"strconv"
)

const (
	//SAVEGamesFile the file where unfinished games are saved.
	SAVEGamesFile = "server/data/savegames.gob"
)

//TablesServer the battleline tables server
//Keeps track of games being played. Who is playing, who is watching
//and saved unfinished games. It here you ask to start a game.
type TablesServer struct {
	save          bool
	saveDir       string
	StartGameChCl *StartGameChCl
	pubList       *PubList
	doneCh        chan struct{}
	saveGames     *SaveGames
	archiver      *arch.Client
}

//NewTablesServer creates a battleline tables server.
func NewTablesServer(
	pubList *PubList, archiverPort int) (s *TablesServer, err error) {

	s = new(TablesServer)
	s.pubList = pubList
	s.StartGameChCl = NewStartGameChCl()
	s.doneCh = make(chan struct{})
	//bat.GobRegistor()
	s.saveGames, err = LoadSaveGames(SAVEGamesFile)
	if err != nil {
		return s, err
	}
	s.archiver, err = arch.New(archiverPort, "")
	return s, err
}

//Start starts the tables server.
func (s *TablesServer) Start(errCh chan<- error) {
	go startTables(s.StartGameChCl, s.pubList, s.doneCh, errCh, s.saveGames, s.archiver)
}

//Stop stops the tables server.
func (s *TablesServer) Stop() {
	log.Print(log.DebugMsg, "Closing start game channel on tables")
	close(s.StartGameChCl.Close)
	<-s.doneCh
	log.Print(log.DebugMsg, "Receiving done from tables")
}

//Start tables server.
//doneCh closing this channel will close down the tables server.
func startTables(
	startGameChCl *StartGameChCl,
	pubList *PubList, doneCh chan struct{},
	errCh chan<- error,
	saveGames *SaveGames,
	archiver *arch.Client) {

	finishTableCh := make(chan *FinishTableData)
	startCh := startGameChCl.Channel
	var done bool
	games := make(map[int]*GameData)
	archiver.Start()
Loop:
	for {
		select {
		case fin := <-finishTableCh:
			delete(games, fin.ids[0])
			delete(games, fin.ids[1])
			if fin.game != nil {
				if fin.game.Pos.Turn.State == bat.TURNQuit || fin.game.Pos.Turn.State == bat.TURNFinish {
					archiver.Archive(fin.game)
				} else {
					saveGames.Games[gameID(fin.ids)] = fin.game
				}
			}
			if done && len(games) == 0 {
				break Loop
			}
			publishTables(games, pubList)
		case start := <-startCh:

			if isPlaying(start.PlayerIds, games) {
				close(start.PlayerChs[0])
				close(start.PlayerChs[1])
			} else {
				var game *bat.Game
				game, start, saveGames = getOldGame(start, saveGames)
				watch := NewWatchChCl()
				go tableServe(start.PlayerIds, start.PlayerChs, watch, game, finishTableCh, errCh)
				games[start.PlayerIds[0]] = NewGameData(start.PlayerIds[1], watch)
				games[start.PlayerIds[1]] = NewGameData(start.PlayerIds[0], watch)
				publishTables(games, pubList)
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
	archiver.Stop()
	archiver.WaitToFinish()
	close(doneCh)
}
func getOldGame(start *StartGameData, saveGames *SaveGames) (*bat.Game, *StartGameData, *SaveGames) {
	game, old := saveGames.Games[gameID(start.PlayerIds)]
	if old {
		delete(saveGames.Games, gameID(start.PlayerIds))
		if game.PlayerIds != start.PlayerIds {
			start.PlayerIds = [2]int{start.PlayerIds[1], start.PlayerIds[0]}
			start.PlayerChs = [2]chan<- *MoveView{start.PlayerChs[1],
				start.PlayerChs[0]}
		}
		if game.Pos == nil {
			game.CalcPos()
		}
	}
	return game, start, saveGames
}
func isPlaying(ids [2]int, games map[int]*GameData) bool {
	_, found := games[ids[0]]
	if !found {
		_, found = games[ids[1]]
	}

	return found
}

// publish the current games list.
func publishTables(games map[int]*GameData, pubList *PubList) {
	copy := make(map[int]*GameData)
	for key, v := range games {
		copy[key] = v
	}
	go pubList.UpdateGames(copy)
}

//NewGameData create a new GameData pointer.
func NewGameData(opp int, watch *WatchChCl) (g *GameData) {
	g = new(GameData)
	g.Opp = opp
	g.Watch = watch
	return g
}

//gameID makes a unique game id.
func gameID(players [2]int) (id string) {
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

//NewSaveGames creates a SaveGames structur.
//To use Gob we single structur tow same multible games.
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
	file, err := os.Create(SAVEGamesFile)
	if err != nil {
		err = errors.Wrap(err, log.ErrNo(15)+"Create games file")
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(games)
	return err
}

// LoadSaveGames loads the save games.
func LoadSaveGames(filePath string) (games *SaveGames, err error) {
	file, err := os.Open(filePath)
	if err == nil {
		defer func() {
			if cerr := file.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()
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
			err = errors.Wrap(err, log.ErrNo(16)+"Open saved games file")
			return games, err
		}
	}
	return games, err
}

// StartGameData is the information need to start a game.
type StartGameData struct {
	PlayerIds [2]int
	PlayerChs [2]chan<- *MoveView
}

// StartGameChCl the start game channel.
type StartGameChCl struct {
	Channel chan *StartGameData
	Close   chan struct{}
}

//NewStartGameChCl creates a StartGameChCl.
func NewStartGameChCl() (sgc *StartGameChCl) {
	sgc = new(StartGameChCl)
	sgc.Channel = make(chan *StartGameData)
	sgc.Close = make(chan struct{})
	return sgc
}
