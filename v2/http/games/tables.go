package games

import (
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	arch "github.com/rezder/go-battleline/v2/archiver/client"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
)

const (
	//histsDbFile  unfinished games hitory data base file.
	histDbFILE = "server/data/savegames.db"
)

//TablesServer the battleline tables server
//Keeps track of games being played. Who is playing, who is watching
//and saved unfinished games. It here you ask to start a game.
type TablesServer struct {
	StartGameChCl *StartGameChCl
	pubList       *PubList
	doneCh        chan struct{}
	savedGamesDb  *dbhist.Db
	archiver      *arch.Client
}

//NewTablesServer creates a battleline tables server.
//archAddr maybe empty or nil
func NewTablesServer(
	pubList *PubList,
	archPokePort int,
	archAddr string,
) (s *TablesServer, err error) {

	s = new(TablesServer)
	s.pubList = pubList
	s.StartGameChCl = NewStartGameChCl()
	s.doneCh = make(chan struct{})
	db, err := bolt.Open(histDbFILE, 0600, nil)
	if err != nil {
		err = errors.Wrapf(err, "Open data base file %v failed", histDbFILE)
		return s, err
	}
	s.savedGamesDb = dbhist.New(dbhist.KeyPlayers, db, 500)
	err = s.savedGamesDb.Init()
	if err != nil {
		_ = db.Close()
		return s, err
	}
	if err != nil {
		_ = db.Close()
		return s, err
	}
	s.archiver, err = arch.New(archPokePort, archAddr)
	return s, err
}

//CloseDb closes the database, stop close the database but
// if not starting the server after init the database must be closed.
func (s *TablesServer) CloseDb() error {
	return s.savedGamesDb.Close()
}

//Start starts the tables server.
func (s *TablesServer) Start(errCh chan<- error) {
	go startTables(s.StartGameChCl, s.pubList, s.doneCh, errCh, s.savedGamesDb, s.archiver)
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
	savedGamesDb *dbhist.Db,
	archiver *arch.Client) {

	finishTableCh := make(chan *bg.Game)
	startCh := startGameChCl.Channel
	var isDone bool
	games := make(map[int]*GameData)
	archiver.Start()
Loop:
	for {
		select {
		case game := <-finishTableCh:
			delete(games, game.Hist.PlayerIDs[0])
			delete(games, game.Hist.PlayerIDs[1])
			if game.Pos.LastMoveType.IsPause() {
				log.Printf(log.DebugMsg, "Saving stopped game: %v,%v", game.Hist.PlayerIDs, game.Hist.Time)
				err := savedGamesDb.Put(game.Hist)
				if err != nil {
					errTxt := "Save game player ids: %v failed."
					err = errors.Wrapf(err, errTxt, game.Hist.PlayerIDs)
					errCh <- err
				}
			} else {
				log.Printf(log.DebugMsg, "Archiving game: %v,%v", game.Hist.PlayerIDs, game.Hist.Time)
				archiver.Archive(game.Hist)
			}

			if isDone && len(games) == 0 {
				break Loop
			}
			publishTables(games, pubList)
		case start := <-startCh:
			if isPlaying(start.PlayerIds, games) {
				log.Printf(log.DebugMsg, "Close requested start game: %v", start)
				close(start.PlayerChs[0])
				close(start.PlayerChs[1])
			} else {
				log.Printf(log.DebugMsg, "Tables starts game: %v", start)
				var savedGame *bg.Game
				savedGame, start = getOldGame(start, savedGamesDb, errCh)
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
	err := savedGamesDb.Close()
	if err != nil {
		errCh <- err
	}
	archiver.Stop()
	close(doneCh)
}
func getOldGame(
	start *StartGameChData,
	hdb *dbhist.Db,
	errCh chan<- error) (*bg.Game, *StartGameChData) {

	var game *bg.Game
	key := dbhist.KeyPlayerIDs(start.PlayerIds)
	hist, err := hdb.Get(key)
	if err != nil {
		err = errors.Wrap(err, "Loading game from data base failed.")
		errCh <- err
	}
	if hist != nil {
		err = hdb.Delete(key)
		if err != nil {
			err = errors.Wrapf(err, "Failed deleting history for %v", start.PlayerIds)
			errCh <- err
		}
		game = bg.NewGame()
		game.LoadHist(hist)
		_ = game.Resume() //Assumes we do not save finsihed game
		if game.Hist.PlayerIDs != start.PlayerIds {
			start.PlayerIds = [2]int{start.PlayerIds[1], start.PlayerIds[0]}
			start.PlayerChs = [2]chan<- *PlayingChData{start.PlayerChs[1],
				start.PlayerChs[0]}
		}
	}
	return game, start
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
