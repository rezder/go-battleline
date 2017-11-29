package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/bot"
	"github.com/rezder/go-battleline/v2/bot/prob"
	"github.com/rezder/go-battleline/v2/bot/tf"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/http/games"
	"github.com/rezder/go-error/log"
)

func main() {
	var gameURL string
	var name string
	var tfURL string
	var tfServerFile string
	var tfModelDir string
	var scheme string // http or https
	var pw string
	var logLevel int
	var limitNoGame int
	var isSendInvite bool
	flag.StringVar(&scheme, "scheme", "http", "Scheme http or https")
	flag.StringVar(&gameURL, "gameurl", "game.rezder.com:8282", "The server url example: game.rezder.com:8181")
	flag.StringVar(&tfURL, "tfurl", "", "The tensorflow server url example: localhost:5555")
	flag.StringVar(&tfServerFile, "tfserver", "", "The python server file ex.: /home/rho/Python/tensorflow/battleline/botserver.py")
	flag.StringVar(&tfModelDir, "tfmodeldir", "", "The tensorflow model dir ex.: /home/rho/Python/tensorflow/battleline/model1")
	flag.StringVar(&name, "name", "Rene", "User name")
	flag.StringVar(&pw, "pw", "12345678", "User password")
	flag.IntVar(&logLevel, "loglevel", 0, "Log level 0 default lowest, 3 highest")
	flag.BoolVar(&isSendInvite, "send", false, "If true send invites else accept invite")
	flag.IntVar(&limitNoGame, "limit", 0, "When the number of game played reach the limit the bot closes down")
	flag.Parse()

	log.InitLog(logLevel)
	inviteHandler := bot.NewStdInviteHandler(isSendInvite)
	if len(tfServerFile) > 0 && len(tfModelDir) > 0 && len(tfURL) == 0 {
		log.Print(log.Min, "-tfurl was not included as option, it must as tfserver and tfmodeldir start as tensorflow server")
		return
	}
	serverCmd, err := cmdStart(tfServerFile, tfModelDir, tfURL)
	if err != nil {
		log.PrintErr(err)
		return
	}
	if serverCmd != nil {
		time.Sleep(time.Second * 3) //It should also take some time to log on
		defer cmdStop(serverCmd)
	}
	battMover, err := newMover(tfURL)
	if err != nil {
		log.PrintErr(err)
		return
	}
	defer func() {
		if cerr := battMover.close(); cerr != nil {
			cerr = errors.Wrap(cerr, "Closing tensorflow connection failed")
			log.PrintErr(cerr)
		}
	}()
	battBot, err := bot.New(scheme, gameURL, name, pw, limitNoGame, inviteHandler, battMover)
	if err != nil {
		log.PrintErr(err)
		return
	}
	battBot.Start()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Printf(log.Min, "Bot (%v) up and running. Close with ctrl+c\n", name)
	select {
	case <-stop:
		log.Printf(log.DebugMsg, "Bot (%v) receive stop signal.", name)
		battBot.Stop()
		log.Printf(log.DebugMsg, "Bot (%v) send stop.", name)
	case <-battBot.FinConnCh:
		log.Printf(log.DebugMsg, "Bot (%v) done.", name)
	}
}

type mover struct {
	tfCon *tf.Con
}

func newMover(tfURL string) (m *mover, err error) {
	m = new(mover)
	if len(tfURL) > 0 {
		var tfCon *tf.Con
		tfCon, err = tf.New(tfURL)
		if err != nil {
			err = errors.Wrap(err, "Creating zmq socket failed")
			return nil, err
		}
		m.tfCon = tfCon
	}
	return m, err
}
func (m *mover) GameStart(playingData *games.PlayingChData) {
}
func (m *mover) GameRestart(playingData *games.PlayingChData) {
}
func (m *mover) GameStop(playingData *games.PlayingChData) {
}
func (m *mover) GameFinish(playingData *games.PlayingChData) {

}
func (m *mover) Move(viewPos *game.ViewPos) (moveix int) {
	firstMove := viewPos.Moves[0]
	switch firstMove.MoveType {
	case game.MoveTypeAll.Cone:
		moveix = prob.MoveClaim(viewPos)
	case game.MoveTypeAll.Scout2:
		fallthrough
	case game.MoveTypeAll.Scout3:
		fallthrough
	case game.MoveTypeAll.Deck:
		moveix = prob.MoveDeck(viewPos)
	case game.MoveTypeAll.ScoutReturn: //TODO add tf support for scoutreturn change here and when create data
		moveix = prob.MoveScoutReturn(viewPos)
	default:
		if m.tfCon != nil {
			moveix = tf.MoveHand(viewPos, m.tfCon)
		} else {
			moveix = prob.MoveHand(viewPos)
		}
	}
	log.Printf(log.Debug, "Moveix: %v,Move: %v\n\n", moveix, viewPos.Moves[moveix])
	return moveix
}

func (m *mover) close() (err error) {
	if m.tfCon != nil {
		err = m.tfCon.Close()
	}
	return err
}
func cmdStop(cmd *exec.Cmd) {
	if cmd != nil {
		err := cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			log.PrintErr(err)
			return
		}
		_ = cmd.Wait() //TODO change to inspect error when close down works zmq4 python
	}
	return
}
func cmdStart(serverFile, modelDir, tfURLTxt string) (cmd *exec.Cmd, err error) {
	if len(serverFile) > 0 && len(modelDir) > 0 {
		port := findPort(tfURLTxt)
		if len(port) == 0 {
			err = errors.New(fmt.Sprintf("-tfurl: %v does not include a port", tfURLTxt))
			return cmd, err
		}
		log.Printf(log.DebugMsg, "Start python server: %v dir: %v", serverFile, modelDir)
		cmd = exec.Command("python", serverFile, "--model_dir="+modelDir, "--port="+port)
		log.Printf(log.Debug, "The command", cmd)
		err = cmd.Start()
		if err != nil {
			return nil, err
		}
	}
	return cmd, err
}

//findPort finds the port from the txt.
//It expect the txt to be host:port
func findPort(txt string) (port string) {
	ix := strings.Index(txt, ":")
	if ix != -1 && ix != len(txt)-1 {
		port = txt[ix+1 : len(txt)]
	}
	return port
}
