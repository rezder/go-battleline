package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/rezder/go-battleline/battbot/gamepos"
	"github.com/rezder/go-battleline/battserver/players"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-error/cerrors"
	"golang.org/x/net/websocket"
	"io"
	"log"
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
	flag.IntVar(&logLevel, "loglevel", 0, "Log level 0 default lowest, 2 highest")
	flag.BoolVar(&sendInvite, "send", false, "If true send invites else accept invite")
	flag.Parse()
	if len(port) != 0 {
		addrPort = addr + ":" + port
	} else {
		addrPort = addr
	}
	cerrors.InitLog(logLevel)
	cookies, err := login(scheme, addrPort, addr, name, pw)
	if err != nil {
		if cerrors.LogLevel() != cerrors.LOG_Debug {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Error: %+v", err)
		}
		return
	}
	conn, err := createWs(scheme, addrPort, cookies)
	if err != nil {
		if cerrors.LogLevel() != cerrors.LOG_Debug {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Error: %+v", err)
		}
		return
	}
	defer conn.Close()
	doneCh := make(chan struct{})
	finConnCh := make(chan struct{})
	go start(conn, doneCh, finConnCh, sendInvite, name)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Printf("Bot (%v) up and running. Close with ctrl+c\n", name)
Loop:
	for {
		select {
		case <-stop:
			close(doneCh)
			if cerrors.IsVerbose() {
				log.Println("Bot stopped with interrupt signal")
			}
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
		panic("Jar error")
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
								if cerrors.LogLevel() != cerrors.LOG_Debug {
									log.Printf("List unmarshal error: %v", err)
								} else {
									log.Printf("List unmarshal error: %+v", err)
								}
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
								if cerrors.LogLevel() == cerrors.LOG_Debug {
									log.Printf("Sending invite to %v", invite)
								}
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
						if cerrors.LogLevel() != cerrors.LOG_Debug {
							log.Printf("Invite unmarshal error: %v", err)
						} else {
							log.Printf("Invite unmarshal error: %+v", err)
						}
						close(messDoneCh)
						break Loop
					}
					if !sendIvites {
						if gamePos == nil {
							act := players.NewAction(players.ACTInvAccept)
							act.ID = invite.InvitorID
							if cerrors.LogLevel() == cerrors.LOG_Debug {
								log.Printf("Accepting invite from %v", invite.InvitorID)
							}
							ok := netWrite(conn, act)
							if !ok {
								close(messDoneCh)
								break Loop
							}
						}
					} else {
						act := players.NewAction(players.ACTInvDecline)
						act.ID = invite.InvitorID
						if cerrors.LogLevel() == cerrors.LOG_Debug {
							log.Printf("Declining invite from %v", invite.InvitorID)
						}
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
						if cerrors.LogLevel() != cerrors.LOG_Debug {
							log.Printf("Move unmarshal error: %v", err)
						} else {
							log.Printf("Move unmarshal error: %+v", err)
						}
						close(messDoneCh)
						break Loop
					}

					if gamePos.UpdMove(mv) {
						gamePos = nil
						if sendIvites {
							act := players.NewAction(players.ACTList)
							if cerrors.LogLevel() == cerrors.LOG_Debug {
								log.Println("Request list")
							}
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
						log.Printf("Server closed connection: %v", closeCon.Reason)
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
	if cerrors.LogLevel() != cerrors.LOG_Default {
		log.Printf("Number of game played %v", noGame)
	}
}

//netWrite writes to the websocket.
func netWrite(conn *websocket.Conn, act *players.Action) bool {
	err := websocket.JSON.Send(conn, act)
	if err != nil {
		if cerrors.LogLevel() != cerrors.LOG_Debug {
			log.Printf("Send action: %v\n error: %v", act, err)
		} else {
			log.Printf("Send action: %v\n error: %+v", act, err)
		}
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
			if cerrors.LogLevel() != cerrors.LOG_Debug {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("Error: %+v", err)
			}
			break Loop
		}
	}
}

//JsonDataTemp is the JsonData type before we can umarshal the interface values.
type JsonDataTemp struct {
	JsonType int
	Data     json.RawMessage
}
