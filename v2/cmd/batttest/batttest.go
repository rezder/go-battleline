package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/db/dbhist"
	"github.com/rezder/go-battleline/v2/game"
	lg "github.com/rezder/go-battleline/v2/http/login"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func main() {
	rootDirFlag := flag.String("rootdir", "/home/rho/js/batt-game-app/build/", "The http server files root directory")
	noFlag := flag.Int("no", 1000, "The numbers of games to play")
	keepDbFlag := flag.Bool("keepdb", false, "Keep the game database")
	//tfAddrFlag := flag.String("tfaddr", "localhost:5555", "The tensorflow move server") TODO
	log.InitLog(log.Debug)
	flag.Parse()
	exitCode := 1
	defer func() { os.Exit(exitCode) }()
	dirPath := "server"
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		log.Printf(log.Min, "Directory: %v allready exist", dirPath)
		return
	}
	bdbFileName := "bdb.db2"
	if _, err := os.Stat(bdbFileName); !os.IsNotExist(err) {
		log.Printf(log.Min, "Data base file: %v allready exist", bdbFileName)
		return
	}
	dirDataPath := filepath.Join(dirPath, "data")
	if err := os.MkdirAll(dirDataPath, os.ModePerm); err != nil {
		err = errors.Wrapf(err, "Create directory: %v failed", dirDataPath)
		log.PrintErr(err)
		return
	}
	runCmds := make([]*exec.Cmd, 0, 4)
	archCmd := exec.Command("battarchiver2", "-dbfile="+bdbFileName, "-loglevel=3")
	archCmd.Stderr = os.Stderr
	archCmd.Stdout = os.Stdout
	err := archCmd.Start()
	if err != nil {
		err = errors.Wrap(err, "Batt Archiver2 cmd failed")
		log.PrintErr(err)
		intCmds(runCmds)
		return
	}
	runCmds = append(runCmds, archCmd)
	time.Sleep(time.Second * 1)
	httpCmd := exec.Command("battserver2", "-rootdir="+*rootDirFlag, "-archaddr=localhost:7272")
	httpCmd.Stderr = os.Stderr
	httpCmd.Stdout = os.Stdout
	err = httpCmd.Start()
	if err != nil {
		err = errors.Wrap(err, "Batt Server2 cmd failed")
		log.PrintErr(err)
		return
	}
	runCmds = append(runCmds, httpCmd)
	time.Sleep(time.Second * 1)

	err = createUsrs()
	if err != nil {
		err = errors.Wrap(err, "Create users failed")
		log.PrintErr(err)
		intCmds(runCmds)
		return
	}
	time.Sleep(time.Millisecond * 500)
	botCmd := exec.Command("battbot2", "-name=Peter")
	err = botCmd.Start()
	if err != nil {
		err = errors.Wrap(err, "Batt Bot2 cmd failed")
		log.PrintErr(err)
		intCmds(runCmds)
		return
	}
	runCmds = append(runCmds, botCmd)
	time.Sleep(time.Second * 1)
	botSendCmd := exec.Command("battbot2", "-send", "-limit="+strconv.Itoa(*noFlag))
	st := time.Now()
	log.Printf(log.Min, "Start playing %v games: %v", *noFlag, st.Format(time.Stamp))
	err = botSendCmd.Run()
	if err != nil {
		err = errors.Wrap(err, "Batt Bot2 send cmd failed")
		log.PrintErr(err)
		intCmds(runCmds)
		return
	}
	d := time.Since(st)
	log.Printf(log.Min, "Finsihed playing games %v", d)
	intCmds(runCmds)
	err = os.RemoveAll(dirPath)
	if err != nil {
		err = errors.Wrap(err, "Remove directories failed")
		log.PrintErr(err)
		return
	}
	err = checkDb(bdbFileName, *noFlag)
	if err != nil {
		log.PrintErr(err)
		return
	}
	if !*keepDbFlag {
		err = os.Remove(bdbFileName)
		if err != nil {
			log.PrintErr(err)
			return
		}
	}
	exitCode = 0
}
func checkDb(bdbFileName string, noGames int) (err error) {
	db, err := bolt.Open(bdbFileName, 0600, nil)
	if err != nil {
		err = errors.Wrap(err, "Open database failed")
		return err
	}
	defer func() {
		if cErr := db.Close(); cErr != nil && err != nil {
			err = cErr
		}
	}()
	dbHist := dbhist.New(dbhist.KeyPlayersTime, db, 1000)
	err = dbHist.Init()
	if err != nil {
		err = errors.Wrap(err, "Init database failed")
		return err
	}
	var wins [2]int
	_, _, err = dbHist.Search(func(hist *game.Hist) bool {
		wins[hist.LastMove().Mover]++
		return false
	}, nil)
	if err != nil {
		err = errors.Wrap(err, "Database search failed")
		return err
	}
	no := wins[0] + wins[1]
	if no != noGames {
		err = errors.New(fmt.Sprintf("Only %v games was completted", no))
		return err
	}
	log.Printf(log.Min, "Wins: %v", wins)
	return err
}

func intCmds(cmds []*exec.Cmd) {
	for i := len(cmds) - 1; i >= 0; i-- {
		intCmd(cmds[i])
	}
}
func intCmd(cmd *exec.Cmd) {
	sigErr := cmd.Process.Signal(syscall.SIGINT)
	if sigErr != nil {
		sigErr = errors.Wrap(sigErr, "Sending interupt signal failed")
		log.PrintErr(sigErr)
	}
	err := cmd.Wait()
	if err != nil {
		err = errors.Wrap(err, "Command wait failed")
		log.PrintErr(err)
	}
	return
}
func createUsrs() (err error) {
	gameURL := "localhost:8282"
	err = createUsr("Rene", "12345678", gameURL)
	if err != nil {
		return err
	}
	err = createUsr("Peter", "12345678", gameURL)
	return err
}
func createUsr(name, pw, gameURL string) (err error) {
	client := new(http.Client)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client.Jar = jar
	resp, err := client.PostForm("http://"+gameURL+"/post/client",
		url.Values{"txtUserName": {name}, "pwdPassword": {pw}})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	decoder := json.NewDecoder(resp.Body)
	var tmp struct{ LogInStatus lg.Status }
	err = decoder.Decode(&tmp)
	if err != nil {
		err = errors.Wrap(err, "Decoding of login status failed.")
		return err
	}
	loginStatus := tmp.LogInStatus
	if !loginStatus.IsOk() {
		err = fmt.Errorf("Login failed: %v", loginStatus)
		return err
	}

	cookiesURL, err := url.Parse("http://" + gameURL + "/in/gamews")
	if err != nil {
		return err
	}
	cookies := client.Jar.Cookies(cookiesURL) //Strips the path from cookie proberly not a problem
	okCookies := false
	if len(cookies) != 0 {
		for _, cookie := range cookies {
			if cookie.Name == "sid" {
				okCookies = true
				break
			}
		}
	}
	if !okCookies {
		err = errors.New("Invalid cookies")
		return err
	}
	wsScheme := "ws://"
	config, err := websocket.NewConfig(wsScheme+"localhost:8282/in/gamews", "http://localhost/")
	value := ""
	for _, cookie := range cookies {
		ctxt := fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		if len(value) == 0 {
			value = ctxt
		} else {
			value = value + "; " + ctxt
		}

	}
	config.Header.Set("Cookie", value)
	if err != nil {
		return err
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	err = ws.Close()
	return err
}
