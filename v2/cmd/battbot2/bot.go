package main

import (
	"flag"
	//"github.com/rezder/go-battleline/machine"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battbot/tf"
	"github.com/rezder/go-battleline/v2/bot"
	"github.com/rezder/go-battleline/v2/bot/prob"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/http/games"
	"github.com/rezder/go-error/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var gameURL string
	var name string
	var tfURL string
	var scheme string // http or https
	var pw string
	var logLevel int
	var limitNoGame int
	var isSendInvite bool
	flag.StringVar(&scheme, "scheme", "http", "Scheme http or https")
	flag.StringVar(&gameURL, "gameurl", "game.rezder.com:8181", "The server url example: game.rezder.com:8181")
	flag.StringVar(&tfURL, "tfurl", "", "The tensorflow server url example: localhost:5555")
	flag.StringVar(&name, "name", "Rene", "User name")
	flag.StringVar(&pw, "pw", "12345678", "User password")
	flag.IntVar(&logLevel, "loglevel", 0, "Log level 0 default lowest, 3 highest")
	flag.BoolVar(&isSendInvite, "send", false, "If true send invites else accept invite")
	flag.IntVar(&limitNoGame, "limit", 0, "When the number of game played reach the limit the bot closes down")
	flag.Parse()

	log.InitLog(logLevel)
	inviteHandler := bot.NewStdInviteHandler(isSendInvite)
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
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Printf(log.Min, "Bot (%v) up and running. Close with ctrl+c\n", name)
	select {
	case <-stop:
		battBot.Stop()
	case <-battBot.FinConnCh:
	}
}

type mover struct {
	tfCon *tf.Con
}

func newMover(tfURL string) (m *mover, err error) {
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
	case game.MoveTypeAll.ScoutReturn:
		moveix = prob.MoveScoutReturn(viewPos)
	default:
		if m.tfCon != nil {
			//TODO make tf move moves is no longer need to keep track of order
			/*byteData, moves := machine.CreatePosBot(viewPos)
			probas, err := tfcon.ReqProba(byteData)
			if err != nil {
				err = errors.Wrap(err, "Failed to get probabilities from tensorflow model")
				log.PrintErr(err)
				moveix = prob.MoveHand(viewPos)
			} else {
				moveix = tf.Move(probas)
			}*/
			//	}
		} else {
			moveix = prob.MoveHand(viewPos)
		}
	}
	log.Printf(log.Debug, "Move: %v\n\n", viewPos.Moves[moveix])
	return moveix
}

func (m *mover) close() (err error) {
	if m.tfCon != nil {
		err = m.tfCon.Close()
	}
	return err
}
