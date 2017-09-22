package main

import (
	"flag"
	"github.com/rezder/go-battleline/v2/bot"
	"github.com/rezder/go-error/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var gameUrl string
	var name string
	var tfUrl string
	var scheme string // http or https
	var pw string
	var logLevel int
	var limitNoGame int
	var isSendInvite bool
	flag.StringVar(&scheme, "scheme", "http", "Scheme http or https")
	flag.StringVar(&gameUrl, "gameurl", "game.rezder.com:8181", "The server url example: game.rezder.com:8181")
	flag.StringVar(&tfUrl, "tfurl", "", "The tensorflow server url example: localhost:5555")
	flag.StringVar(&name, "name", "Rene", "User name")
	flag.StringVar(&pw, "pw", "12345678", "User password")
	flag.IntVar(&logLevel, "loglevel", 0, "Log level 0 default lowest, 3 highest")
	flag.BoolVar(&isSendInvite, "send", false, "If true send invites else accept invite")
	flag.IntVar(&limitNoGame, "limit", 0, "When the number of game played reach the limit the bot closes down")
	flag.Parse()

	log.InitLog(logLevel)
	battBot, err := bot.New(scheme, gameUrl, tfUrl, name, pw, limitNoGame, isSendInvite)
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
