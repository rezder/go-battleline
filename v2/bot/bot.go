package bot

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battbot/gamepos"
	"github.com/rezder/go-battleline/battbot/tf"
	"github.com/rezder/go-battleline/battserver/players"
	pub "github.com/rezder/go-battleline/battserver/publist"
	"github.com/rezder/go-battleline/machine"
	"github.com/rezder/go-error/log"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

//Bot a battleline bot. The bot may finish on its own if the connection is lost
// in that case the finConnCh is closed.
type Bot struct {
	ws           *websocket.Conn
	doneCh       chan struct{}
	FinConnCh    chan struct{}
	isSendIvites bool
	name         string
	limitNoGame  int
	tfCon        *tf.Con
}

// New create a battleline bot.
// remember to call cancel or stop to close all connections
// if created without errors.
func New(
	scheme, gameUrl, tfUrl, name, pw string,
	limitNoGames int,
	isSendInvite bool,
) (bot *Bot, err error) {
	bot = new(Bot)
	bot.doneCh = make(chan struct{})
	bot.FinConnCh = make(chan struct{})
	bot.isSendIvites = isSendInvite
	bot.name = name
	bot.limitNoGame = limitNoGames
	cookies, err := login(scheme, gameUrl, name, pw)
	if err != nil {
		return nil, err
	}
	ws, err := createWs(scheme, gameUrl, cookies)
	if err != nil {
		return nil, err
	}
	bot.ws = ws
	if len(tfUrl) > 0 {
		tfCon, err := tf.New(tfUrl)
		if err != nil {
			err = errors.Wrap(err, "Creating zmq socket failed")
			cerr := ws.Close()
			if cerr != nil {
				log.PrintErr(cerr)
			}
			return nil, err
		}
		bot.tfCon = tfCon
	}
	return bot, err
}

//Cancel cancel a created bot.
func (bot *Bot) Cancel() {
	wsErr := bot.ws.Close()
	tfcErr := bot.tfCon.Close()
	if wsErr != nil {
		log.PrintErr(wsErr)
	}
	if tfcErr != nil {
		log.PrintErr(tfcErr)
	}
}

//Start starts a batttleline bot.
func (bot *Bot) Start() {
	go serve(bot.ws, bot.doneCh, bot.FinConnCh, bot.isSendIvites, bot.name, bot.limitNoGame, bot.tfCon)
}

//Stop stops a battleline bot.
func (bot *Bot) Stop() {
	close(bot.doneCh)
	log.Println(log.Verbose, "Bot stopping because of interrupt signal")
	<-bot.FinConnCh
	log.Println(log.Verbose, "Bot stopped")
}

//login logs in to the game server.
func login(scheme, gameUrl, name, pw string) (cookies []*http.Cookie, err error) {
	client := new(http.Client) //TODO can not handle https maybe because of a nginx bug solved in 1.11 current docker version is 1.10
	jar, err := cookiejar.New(nil)
	if err != nil {
		return cookies, err
	}
	client.Jar = jar
	resp, err := client.PostForm(scheme+"://"+gameUrl+"/form/login",
		url.Values{"txtUserName": {name}, "pwdPassword": {pw}})
	if err != nil {
		return cookies, err
	}
	cookiesUrl, err := url.Parse(scheme + "://" + gameUrl + "/in/game")
	if err != nil {
		return cookies, err
	}
	cookies = client.Jar.Cookies(cookiesUrl) //Strips the path from cookie proberly not a problem

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

//serve plays battleline games.
func serve(
	conn *websocket.Conn,
	doneCh <-chan struct{},
	finConnCh chan<- struct{},
	sendIvites bool,
	name string,
	limitNoGame int,
	tfcon *tf.Con) {

	noGame := 0
	defer close(finConnCh)
	messCh := make(chan *JsonDataTemp)
	messDoneCh := make(chan struct{})
	go netRead(conn, messCh, messDoneCh)
	var gamePos *gamepos.Pos
	pongTimer := time.NewTimer(7 * time.Minute)
Loop:
	for {
		select {
		case <-pongTimer.C:
			act := players.NewAction(players.ACTList)
			log.Println(log.DebugMsg, "Timer Request list")
			ok := netWrite(conn, act)
			if !ok {
				close(messDoneCh)
				break Loop
			}
			pongTimer.Reset(7 * time.Minute)
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
						if noGame == limitNoGame {
							close(messDoneCh)
							break Loop
						}
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
							var moveixs [2]int
							if tfcon != nil {
								if !gamePos.IsHandMove() {
									moveixs = gamePos.MakeMove()
								} else {
									byteData, moves := machine.CreatePosBot(gamePos)
									probas, err := tfcon.ReqProba(byteData, len(moves))
									if err != nil {
										err = errors.Wrap(err, "Failed to get probabilities from tensorflow model")
										log.PrintErr(err)
										moveixs = gamePos.MakeMove()
									} else {
										moveixs = gamePos.MakeTfMove(probas, moves)
									}
								}
							} else {
								moveixs = gamePos.MakeMove()
							}
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
	pongTimer.Stop()
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
