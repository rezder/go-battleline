package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battbot/gamepos"
	"github.com/rezder/go-battleline/battserver/players"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var addr string
	var name string
	var port string
	var scheme string // http or https
	var pw string
	var logLevel int
	var addrPort string
	var sendInvite bool
	flag.StringVar(&scheme, "scheme", "http", "Scheme http or https")
	flag.StringVar(&addr, "addr", "game.rezder.com", "The server address with out port")
	flag.StringVar(&port, "port", "8181", "The port")
	flag.StringVar(&name, "name", "Rene", "User name")
	flag.StringVar(&pw, "pw", "12345678", "User password")
	flag.IntVar(&logLevel, "loglevel", 0, "Log level 0 default lowest, 3 highest")
	flag.BoolVar(&sendInvite, "send", false, "If true send invites else accept invite")
	flag.Parse()
	if len(port) != 0 {
		addrPort = addr + ":" + port
	} else {
		addrPort = addr
	}
	log.InitLog(logLevel)
	cookies, err := login(scheme, addrPort, addr, name, pw)
	if err != nil {
		err = errors.Wrap(err, "Login failed")
		log.PrintErr(err)
		return
	}
	conn, err := createWs(scheme, addrPort, cookies)
	if err != nil {
		err = errors.Wrap(err, "Creating websocket failed")
		log.PrintErr(err)
		return
	}
	defer conn.Close()
	doneCh := make(chan struct{})
	finConnCh := make(chan struct{})
	go start(conn, doneCh, finConnCh, sendInvite, name)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Printf(log.Min, "Bot (%v) up and running. Close with ctrl+c\n", name)
Loop:
	for {
		select {
		case <-stop:
			close(doneCh)
			log.Println(log.Verbose, "Bot stopped with interrupt signal")
		case <-finConnCh:
			break Loop
		}
	}

}

//login logs in to the game server.
func login(scheme string, addrPort string, addr string, name string,
	pw string) (cookies []*http.Cookie, err error) {
	client := new(http.Client) //TODO can handle https maybe because of a nginx bug solved in 1.11 current docker version is 1.10
	jar, err := cookiejar.New(nil)
	if err != nil {
		return cookies, err
	}
	client.Jar = jar
	resp, err := client.PostForm(scheme+"://"+addrPort+"/form/login",
		url.Values{"txtUserName": {name}, "pwdPassword": {pw}})
	if err != nil {
		return cookies, err
	}
	url, err := url.Parse(scheme + "://" + addr + "/in/game")
	if err != nil {
		return cookies, err
	}
	cookies = client.Jar.Cookies(url) //Strips the path from cookie proberly not a problem

	defer resp.Body.Close()
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
		return cookies, err
	}
	return cookies, err
}

//createWs creates the websocket connection.
func createWs(scheme string, addrPort string, cookies []*http.Cookie) (conn *websocket.Conn,
	err error) {
	wsScheme := "ws://"
	if scheme == "https" {
		wsScheme = "wss://"
	}
	//Second argument is the orgin of the javascript that create the websocket.
	//It is only used in handshake my handshake is not using it .
	config, err := websocket.NewConfig(wsScheme+addrPort+"/in/gamews", "http://localhost/")
	if err != nil {
		return conn, err
	}
	addCookies(config, cookies)
	return websocket.DialConfig(config)

}

//addCookies adds cookies to the websocket header.
func addCookies(config *websocket.Config, cookies []*http.Cookie) {
	if len(cookies) != 0 {
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
	}
}

//start starts the bot.
func start(
	conn *websocket.Conn,
	doneCh <-chan struct{},
	finConnCh chan<- struct{},
	sendIvites bool,
	name string) {
	noGame := 0
	defer close(finConnCh)
	messCh := make(chan *JsonDataTemp)
	messDoneCh := make(chan struct{})
	go netRead(conn, messCh, messDoneCh)
	var gamePos *gamepos.Pos
Loop:
	for {
		select {
		case <-doneCh:
			close(messDoneCh)
			break Loop
		case jsonDataTemp, open := <-messCh:
			if open {
				switch jsonDataTemp.JsonType {
				case players.JTList:
					if sendIvites {
						if gamePos == nil {
							var list map[string]*pub.Data
							err := json.Unmarshal(jsonDataTemp.Data, &list)
							if err != nil {
								err = errors.Wrap(err, "Unmarshal List failed")
								log.PrintErr(err)
								close(messDoneCh)
								break Loop
							}
							invite := 0
							for _, listData := range list {
								if name != listData.Name && listData.Opp == 0 {
									invite = listData.ID
									break
								}
							}
							if invite != 0 {
								act := players.NewAction(players.ACTInvite)
								act.ID = invite
								log.Printf(log.DebugMsg, "Sending invite to %v", invite)
								ok := netWrite(conn, act)
								if !ok {
									close(messDoneCh)
									break Loop
								}
							}
						}
					}
				case players.JTInvite:
					var invite pub.Invite
					err := json.Unmarshal(jsonDataTemp.Data, &invite)
					if err != nil {
						err = errors.Wrap(err, "Unmarshal Invite failed")
						log.PrintErr(err)
						close(messDoneCh)
						break Loop
					}
					if !sendIvites {
						if gamePos == nil {
							act := players.NewAction(players.ACTInvAccept)
							act.ID = invite.InvitorID
							log.Printf(log.DebugMsg, "Accepting invite from %v", invite.InvitorID)
							ok := netWrite(conn, act)
							if !ok {
								close(messDoneCh)
								break Loop
							}
						}
					} else {
						act := players.NewAction(players.ACTInvDecline)
						act.ID = invite.InvitorID
						log.Printf(log.DebugMsg, "Declining invite from %v", invite.InvitorID)
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					}
				case players.JTMess:
				case players.JTMove:
					if gamePos == nil {
						gamePos = gamepos.New()
						noGame = noGame + 1
					}
					mv, err := unmarshalMoveJSON(jsonDataTemp.Data)
					if err != nil {
						err = errors.Wrap(err, "Unmarshal Move failed")
						close(messDoneCh)
						break Loop
					}

					if gamePos.UpdMove(mv) {
						gamePos = nil
						if sendIvites {
							act := players.NewAction(players.ACTList)
							log.Println(log.DebugMsg, "Request list")
							ok := netWrite(conn, act)
							if !ok {
								close(messDoneCh)
								break Loop
							}
						}
					} else {
						if gamePos.IsBotTurn() {
							moveixs := gamePos.MakeMove()
							act := players.NewAction(players.ACTMove)
							act.Move = moveixs

							ok := netWrite(conn, act)
							if !ok {
								close(messDoneCh)
								break Loop
							}
						}
					}
				case players.JTBenchMove:
				case players.JTCloseCon:
					var closeCon players.CloseCon
					err := json.Unmarshal(jsonDataTemp.Data, &closeCon)
					if err == nil {
						log.Printf(log.DebugMsg, "Server closed connection: %v", closeCon.Reason)
					} else {
						err = errors.Wrap(err, "Unmarshal CloseCon failed")
						log.PrintErr(err)
					}
					close(messDoneCh)

					break Loop
				case players.JTClearInvites:
				default:
					txt := fmt.Sprintf("Message not implemented yet for %v\n", jsonDataTemp.JsonType)
					panic(txt)
				}

			} else { //closed messCh
				break Loop
			}
		}
	}
	log.Printf(log.Verbose, "Number of game played %v", noGame)
}

//netWrite writes to the websocket.
func netWrite(conn *websocket.Conn, act *players.Action) bool {
	err := websocket.JSON.Send(conn, act)
	if err != nil {
		err = errors.Wrapf(err, "Writing action %v to websocket failed", act)
		log.PrintErr(err)
		return false
	}
	return true

}

//netRead listent to the websocket and forward the message.
func netRead(conn *websocket.Conn, messCh chan<- *JsonDataTemp, messDoneCh <-chan struct{}) {
	defer close(messCh)
Loop:
	for {
		var jsonMess JsonDataTemp
		err := websocket.JSON.Receive(conn, &jsonMess)
		if err == nil {
			select {
			case messCh <- &jsonMess:
			case <-messDoneCh:
				break Loop
			}
		} else if err == io.EOF {
			break Loop
		} else {
			err = errors.Wrap(err, "Reading from websocket failed")
			log.PrintErr(err)
			break Loop
		}
	}
}

//JsonDataTemp is the JsonData type before we can umarshal the interface values.
type JsonDataTemp struct {
	JsonType int
	Data     json.RawMessage
}
