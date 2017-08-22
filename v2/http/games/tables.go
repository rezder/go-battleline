package games

import (
	"encoding/gob"
	"github.com/pkg/errors"
	arch "github.com/rezder/go-battleline/v2/archiver/client"
	bg "github.com/rezder/go-battleline/v2/game"
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
	StartGameChCl *StartGameChCl
	pubList       *PubList
	doneCh        chan struct{}
	savedGames    map[string]*bg.Game
	archiver      *arch.Client
}

//NewTablesServer creates a battleline tables server.
func NewTablesServer(
	pubList *PubList, archiverPort int) (s *TablesServer, err error) {

	s = new(TablesServer)
	s.pubList = pubList
	s.StartGameChCl = NewStartGameChCl()
	s.doneCh = make(chan struct{})
	s.savedGames, err = loadSavedGames(SAVEGamesFile)
	if err != nil {
		return s, err
	}
	s.archiver, err = arch.New(archiverPort, "")
	return s, err
}

//Start starts the tables server.
func (s *TablesServer) Start(errCh chan<- error) {
	go startTables(s.StartGameChCl, s.pubList, s.doneCh, errCh, s.savedGames, s.archiver)
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
	savedGames map[string]*bg.Game,
	archiver *arch.Client) {

	finishTableCh := make(chan *FinishTableData)
	startCh := startGameChCl.Channel
	var isDone bool
	games := make(map[int]*GameData)
	archiver.Start()
Loop:
	for {
		select {
		case fin := <-finishTableCh:
			delete(games, fin.ids[0])
			delete(games, fin.ids[1])
			if fin.game != nil {
				if fin.game.Pos.LastMoveType.IsPause() {
					savedGames[gameID(fin.ids)] = fin.game
				} else {
					archiver.Archive(fin.game.Hist)
				}
			}
			if isDone && len(games) == 0 {
				break Loop
			}
			publishTables(games, pubList)
		case start := <-startCh:

			if isPlaying(start.PlayerIds, games) {
				close(start.PlayerChs[0])
				close(start.PlayerChs[1])
			} else {
				var savedGame *bg.Game
				savedGame, start, savedGames = getOldGame(start, savedGames)
				joinWatchCh := NewJoinWatchChCl()
				go tableServe(start.PlayerIds, start.PlayerChs, joinWatchCh, savedGame, finishTableCh, errCh)
				games[start.PlayerIds[0]] = NewGameData(start.PlayerIds[1], joinWatchCh)
				games[start.PlayerIds[1]] = NewGameData(start.PlayerIds[0], joinWatchCh)
				publishTables(games, pubList)
			}

		case <-startGameChCl.Close:
			if len(games) == 0 {
				break Loop
			} else {
				startCh = nil
				isDone = true
			}
		}
	} //loop
	err := saveGames(savedGames)
	if err != nil {
		errCh <- err
	}
	archiver.Stop()
	close(doneCh)
}
func getOldGame(start *StartGameChData, games map[string]*bg.Game) (*bg.Game, *StartGameChData, map[string]*bg.Game) {
	game, old := games[gameID(start.PlayerIds)]
	if old {
		delete(games, gameID(start.PlayerIds))
		if game.Hist.PlayerIDs != start.PlayerIds {
			start.PlayerIds = [2]int{start.PlayerIds[1], start.PlayerIds[0]}
			start.PlayerChs = [2]chan<- *PlayingChData{start.PlayerChs[1],
				start.PlayerChs[0]}
		}
		_ = game.Resume() //Assume we do not save finsihed game
	}
	return game, start, games
}
func isPlaying(ids [2]int, games map[int]*GameData) bool {
	_, isFound := games[ids[0]]
	if !isFound {
		_, isFound = games[ids[1]]
	}

	return isFound
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
func NewGameData(opp int, watch *JoinWatchChCl) (g *GameData) {
	g = new(GameData)
	g.Opp = opp
	g.JoinWatchChCl = watch
	return g
}

//gameID makes a unique game id.
func gameID(playerIDs [2]int) (id string) {
	if playerIDs[0] > playerIDs[1] {
		id = strconv.Itoa(playerIDs[0]) + "," + strconv.Itoa(playerIDs[1])
	} else {
		id = strconv.Itoa(playerIDs[1]) + "," + strconv.Itoa(playerIDs[0])
	}
	return id
}

//FinishTableData the data structur send on the finish channel.
type FinishTableData struct {
	ids  [2]int
	game *bg.Game
}

func saveGames(games map[string]*bg.Game) (err error) {
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
	hists := make(map[string]*bg.Hist)
	for key, game := range games {
		hists[key] = game.Hist
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(hists)
	return err
}

// loadSaveGames loads the saved games.
func loadSavedGames(filePath string) (games map[string]*bg.Game, err error) {
	file, err := os.Open(filePath)
	if err == nil {
		defer func() {
			if cerr := file.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()
		decoder := gob.NewDecoder(file)
		hists := make(map[string]*bg.Hist)
		err = decoder.Decode(&hists)
		if err == nil {
			games = make(map[string]*bg.Game)
			for key, hist := range hists {
				game := bg.NewGame()
				game.LoadHist(hist)
				games[key] = game
			}
		}
	} else {
		if os.IsNotExist(err) {
			err = nil
			games = make(map[string]*bg.Game) //first start
		} else {
			err = errors.Wrap(err, log.ErrNo(16)+"Open saved games file")
			return games, err
		}
	}
	return games, err
}

// StartGameChData is the information need to start a game.
type StartGameChData struct {
	PlayerIds [2]int
	PlayerChs [2]chan<- *PlayingChData
}

// StartGameChCl the start game channel.
type StartGameChCl struct {
	Channel chan *StartGameChData
	Close   chan struct{}
}

//NewStartGameChCl creates a StartGameChCl.
func NewStartGameChCl() (sgc *StartGameChCl) {
	sgc = new(StartGameChCl)
	sgc.Channel = make(chan *StartGameChData)
	sgc.Close = make(chan struct{})
	return sgc
}
