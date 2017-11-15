package viewer

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-battleline/v2/game"
	dpos "github.com/rezder/go-battleline/v2/game/pos"
	"github.com/rezder/go-error/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Server a view battleline server.
type Server struct {
	port int
	*sync.RWMutex
	dbHistFile string
	dbHist     *dbhist.Db
	modelConts map[string]*ModelCont
}

//New creates a new battleline voew server.
func New(port int) (server *Server, err error) {
	server = new(Server)
	server.port = port
	server.RWMutex = new(sync.RWMutex)
	server.modelConts = make(map[string]*ModelCont)
	return server, err
}

//Start starts a server.
func (server *Server) Start() {
	http.Handle("/", http.FileServer(http.Dir("/home/rho/js/batt-app/build/")))
	http.Handle("/dir/", &dirHandler{server: server})
	http.Handle("/games/", &gamesHandler{server: server})
	http.Handle("/model/", &modelHandler{server: server})
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%v", server.port), nil)
		if err != nil {
			err = errors.Wrap(err, "Http server closed with a error")
			log.PrintErr(err)
		}
	}()
}

//Stop stops a server.
//Its is not a perfect break as some one could add a new db
//before we close but it is just a Gui not a real http server.
func (server *Server) Stop() {
	server.Lock()
	defer server.Unlock()
	if server.dbHist != nil {
		err := server.dbHist.Close()
		if err != nil {
			err = errors.Wrapf(err, "Error closing database %v, Error: %v", server.dbHistFile, err)
			log.PrintErr(err)
		}
	}
	for _, mc := range server.modelConts {
		mc.Stop()
	}
}

//UpdateDb updates the server with a new database.
func (server *Server) UpdateDb(dbFilePath string) (err error) {
	server.Lock()
	defer server.Unlock()
	if server.dbHistFile != dbFilePath {
		var hdb *dbhist.Db
		hdb, err = loadDb(dbFilePath)
		if err != nil {
			return err
		}
		if server.dbHist != nil {
			cerr := server.dbHist.Close()
			if cerr != nil {
				log.Printf(log.Min, "Error closing database: %v, Error: %v ", dbFilePath, cerr)
			}
		}
		server.dbHist = hdb
		server.dbHistFile = dbFilePath
	}
	return err
}

//UpdateModel updates the server with a new model.
func (server *Server) UpdateModel(modelDir string) (err error) {
	server.Lock()
	defer server.Unlock()
	_, ok := server.modelConts[modelDir]
	if !ok {
		port := len(server.modelConts) + 5555
		var mc *ModelCont
		mc, err = NewModelCont(modelDir, port)
		if err != nil {
			return err
		}
		log.Printf(log.Debug, "Model: %v is up and running", modelDir)
		server.modelConts[modelDir] = mc
	}
	return err
}

//UpdateStdOut reads the stdOut from a model.
func (server *Server) UpdateStdOut(modelDir string) (stdOut string) {
	server.Lock()
	defer server.Unlock()
	mc, ok := server.modelConts[modelDir]
	if ok {
		stdOut, _ = mc.Read()
	}
	log.Printf(log.DebugMsg, "Reading stdOut from: %v, read: %v, Found model :%v", modelDir, stdOut, ok)
	return stdOut
}

//ReqProbs returns the probabilities of moves.
func (server *Server) ReqProbs(model string, moverView *game.ViewPos) (probs []float64, err error) {
	server.RLock()
	defer server.RUnlock()
	mc, ok := server.modelConts[model]
	if !ok {
		return probs, errors.New("Model does not exist")
	}
	probs, err = mc.Request(moverView)
	return probs, err
}

//NoGames returns the number of games of the current database.
func (server *Server) NoGames() (no int) {
	server.RLock()
	defer server.RUnlock()
	if server.dbHist != nil {
		_, _, _ = server.dbHist.Search(func(hist *game.Hist) bool {
			no++
			return false
		}, nil)
	}
	return no
}
func createKey(playerIDs []int, ts time.Time, dbHist *dbhist.Db) (key []byte) {
	hist := &game.Hist{
		PlayerIDs: [2]int{playerIDs[0], playerIDs[1]},
		Time:      ts,
	}
	return dbHist.Key(hist)
}

//Hists returns game histories.
func (server *Server) Hists(file string, noGames int, playerIDs []int, ts time.Time) (hists []*game.Hist, err error) {
	server.RLock()
	defer server.RUnlock()
	if file != server.dbHistFile {
		return hists, fmt.Errorf("Database: %v no longer loaded", file)
	}
	var startKey []byte
	if playerIDs != nil {
		startKey = createKey(playerIDs, ts, server.dbHist)
	}
	hists, _, err = server.dbHist.Search(nil, startKey)
	if err != nil {
		return hists, err
	}
	if len(hists) > noGames {
		hists = hists[:noGames]
	}
	return hists, err
}

//Hist returns a game history.
func (server *Server) Hist(file string, noGames int, playerIDs []int, ts time.Time) (hist *game.Hist, err error) {
	server.RLock()
	defer server.RUnlock()
	if file != server.dbHistFile {
		return hist, fmt.Errorf("Database: %v no longer loaded", file)
	}
	key := createKey(playerIDs, ts, server.dbHist)
	hist, err = server.dbHist.Get(key)
	return hist, err
}

type gamesHandler struct {
	server *Server
}

func formValueToInt(key string, r *http.Request) (no int, err error) {
	keyTxt := r.FormValue(key)
	if len(keyTxt) > 0 {
		no, err = strconv.Atoi(keyTxt)
		return no, err
	}
	return 0, fmt.Errorf("Key: %v does not contain any integer", key)
}
func loadDb(dbFilePath string) (dbHist *dbhist.Db, err error) {
	boltDb, err := bolt.Open(dbFilePath, 0600, nil)
	if err != nil {
		return dbHist, err
	}
	hdb := dbhist.New(dbhist.KeyPlayersTime, boltDb, 500)
	err = hdb.Init()
	if err != nil {
		cerr := boltDb.Close()
		if cerr != nil {
			log.Printf(log.Min, "Error closing database: %v, Error: %v ", dbFilePath, cerr)
		}
	}
	return hdb, err
}

func parseGamesFormValues(r *http.Request) (dbFilePath string, noGames int, playerIDs []int, ts time.Time, err error) {
	dbFilePath = r.FormValue("file")
	if len(dbFilePath) == 0 {
		return dbFilePath, noGames, playerIDs, ts, errors.New("Empty file")
	}
	noTxt := r.FormValue("no-games")
	if len(noTxt) > 0 {
		var no int
		no, err = strconv.Atoi(noTxt)
		if err != nil {
			return dbFilePath, noGames, playerIDs, ts, err
		}
		noGames = no
	} else {
		noGames = 0
	}
	p0No, perr := formValueToInt("player0id", r)
	if perr == nil {
		p1No := 0
		p1No, perr = formValueToInt("player1id", r)
		if perr == nil {
			tsText := r.FormValue("ts")
			if len(tsText) > 0 {
				ts, err = time.Parse(time.RFC3339Nano, tsText)
				if err != nil {
					return dbFilePath, noGames, playerIDs, ts, err
				}
				playerIDs = []int{p0No, p1No}
			}
		}
	}
	log.Printf(log.DebugMsg, "Request Games: File: %v,Number of games: %v,PlayerIDs: %v,Time stamp: %v",
		dbFilePath, noGames, playerIDs, ts)
	return dbFilePath, noGames, playerIDs, ts, err
}

// ServeHTTP handles the games requests.
// there is 3 types:
//1) Load database file.
//2) Return game histories.
//3) Return a games god view positions.
func (g *gamesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbFilePath, noGames, playerIDs, ts, err := parseGamesFormValues(r)
	if err != nil {
		log.PrintErr(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if playerIDs == nil && noGames == 0 {
		err := g.server.UpdateDb(dbFilePath)
		if err != nil {
			log.PrintErr(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		no := g.server.NoGames()
		httpWrite(struct{ NoGames int }{NoGames: no}, w)

	} else if noGames > 0 {
		hists, err := g.server.Hists(dbFilePath, noGames, playerIDs, ts)
		if err != nil {
			log.PrintErr(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpWrite(hists, w)
	} else {
		hist, err := g.server.Hist(dbFilePath, noGames, playerIDs, ts)
		if err != nil {
			log.PrintErr(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gamePoss := make([]*game.ViewPos, 0, len(hist.Moves))
		g := game.NewGame()
		g.LoadHist(hist)
		winner := dpos.NoPlayer
		isNext := true
		for isNext {
			winner, isNext = g.ScrollForward()
			gamePos := game.NewViewPos(g.Pos, game.ViewAll.God, winner)
			gamePoss = append(gamePoss, gamePos)
		}
		httpWrite(gamePoss, w)
	}

}
func httpWrite(v interface{}, w http.ResponseWriter) {
	js, err := json.Marshal(v)
	if err != nil {
		log.PrintErr(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.PrintErr(err)
		return
	}
	cap := 0
	if len(js) > 100 {
		cap = len(js) - 100
	}
	log.Printf(log.DebugMsg, "Respond example: %v", string(js[cap:]))
}

type dirHandler struct {
	server *Server
}

func (g *dirHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir := r.FormValue("dir")
	suffix := r.FormValue("suffix")
	isDirTxt := r.FormValue("isdir")
	log.Printf(log.DebugMsg, "Dir: %v,Suffix: %v,isDirTxt: %v", dir, suffix, isDirTxt)
	isDir := false
	if isDirTxt == "true" {
		isDir = true
	}
	fullInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		log.PrintErr(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var infos []*fileInfo
	if len(fullInfos) > 0 {
		infos = make([]*fileInfo, 0, len(fullInfos))
		for _, fullinfo := range fullInfos {
			info := &fileInfo{Name: fullinfo.Name(), IsDir: fullinfo.IsDir()}
			if isDir {
				if !info.IsDir {
					continue
				}
				infos = append(infos, info)
			} else {
				if info.IsDir || (!info.IsDir && len(suffix) == 0) {
					infos = append(infos, info)
				} else {
					if strings.HasSuffix(info.Name, suffix) {
						infos = append(infos, info)
					}
				}
			}
		}
	}
	httpWrite(infos, w)
}

type fileInfo struct {
	Name  string
	IsDir bool
}
