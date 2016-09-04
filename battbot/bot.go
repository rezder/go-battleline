package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	bat "github.com/rezder/go-battleline/battleline"
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
)

func main() {
	var addr string
	var name string
	var port string
	var scheme string // http or https
	var pw string
	var logLevel int
	var addrPort string
	flag.StringVar(&scheme, "scheme", "http", "Scheme http or https")
	flag.StringVar(&addr, "addr", "game.rezder.com", "The server address with out port")
	flag.StringVar(&port, port, "8181", "The port")
	flag.StringVar(&name, "name", "Rene", "User name")
	flag.StringVar(&pw, "pw", "12345678", "User password")
	flag.IntVar(&logLevel, "loglevel", 2, "Log level 0 default lowest, 2 highest") //TODO change to 0
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
	//defer conn.Close()
	doneCh := make(chan struct{})
	finConnCh := make(chan struct{})
	go start(conn, doneCh, finConnCh)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Printf("Bot (%v) up and running. Close with ctrl+c\n", name)
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
	client := new(http.Client)
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Jar error")
	}
	client.Jar = jar
	resp, err := client.PostForm(scheme+"://"+addrPort+"/form/login",
		url.Values{"txtUserName": {name}, "pwdPassword": {pw}})
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
func start(conn *websocket.Conn, doneCh <-chan struct{}, finConnCh chan<- struct{}) {
	defer close(finConnCh)
	messCh := make(chan *JsonDataTemp)
	messDoneCh := make(chan struct{})
	go netRead(conn, messCh, messDoneCh)
	var gamePos *Pos
Loop:
	for {
		select {
		case <-doneCh:
			close(messDoneCh)
			break Loop
		case jsonDataTemp, open := <-messCh:
			if open {
				switch jsonDataTemp.JsonType {
				case players.JT_List:
				case players.JT_Invite:
					if gamePos == nil {
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
						act := players.NewAction()
						act.ActType = players.ACT_INVACCEPT
						act.Id = invite.InvitorId
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					}
				case players.JT_Mess:
				case players.JT_Move:
					if gamePos == nil {
						gamePos = NewPos()
					}
					mv, err := unmarshalMoveJSON(jsonDataTemp.Data)
					if err != nil {
						if cerrors.LogLevel() != cerrors.LOG_Debug {
							log.Printf("Invite unmarshal error: %v", err)
						} else {
							log.Printf("Invite unmarshal error: %+v", err)
						}
						close(messDoneCh)
						break Loop
					}

					fmt.Printf("Recived Move %+v\n\n", mv)
					if gamePos.Move(mv) {
						gamePos = nil
					} else {
						fmt.Printf("Updated pos: %+v\n\n", gamePos)
						if gamePos.turn.MyTurn {
							moveix := makeMove(gamePos)
							if gamePos.turn.Moves != nil {
								sm, ok := gamePos.turn.Moves[moveix[1]].(bat.MoveScoutReturn)
								if ok {
									gamePos.playHand.PlayMulti(sm.Tac)
									gamePos.playHand.PlayMulti(sm.Troop)
									gamePos.deck.PlayScoutReturn(sm.Troop, sm.Tac)
								}
							}
							act := players.NewAction()
							act.ActType = players.ACT_MOVE
							act.Move = moveix
							ok := netWrite(conn, act)
							if !ok {
								close(messDoneCh)
								break Loop
							}
						}
					}
				case players.JT_BenchMove:
				case players.JT_CloseCon:
					close(messDoneCh)
					break Loop
				case players.JT_ClearInvites:
				default:
					txt := fmt.Sprintf("Message not implemented yet for %v\n", jsonDataTemp.JsonType)
					panic(txt)
				}

			} else { //closed messCh
				break Loop
			}
		}
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
