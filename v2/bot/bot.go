package bot

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-battleline/v2/http/games"
	lg "github.com/rezder/go-battleline/v2/http/login"
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
	ws            *websocket.Conn
	doneCh        chan struct{}
	FinConnCh     chan struct{}
	inviteHandler InviteHandler
	mover         Mover
	name          string
	limitNoGame   int
}

// New create a battleline bot.
// remember to call cancel or stop to close all connections
// if created without errors.
func New(
	scheme, gameURL, name, pw string,
	limitNoGames int,
	inviteHandler InviteHandler,
	mover Mover,
) (bot *Bot, err error) {
	bot = new(Bot)
	bot.doneCh = make(chan struct{})
	bot.FinConnCh = make(chan struct{})
	bot.inviteHandler = inviteHandler
	bot.mover = mover
	bot.name = name
	bot.limitNoGame = limitNoGames
	cookies, err := login(scheme, gameURL, name, pw)
	if err != nil {
		return nil, err
	}
	ws, err := createWs(scheme, gameURL, cookies)
	if err != nil {
		return nil, err
	}
	bot.ws = ws
	return bot, err
}

//Cancel cancel a created bot.
func (bot *Bot) Cancel() {
	wsErr := bot.ws.Close()
	if wsErr != nil {
		log.PrintErr(wsErr)
	}
}

//Start starts a batttleline bot.
func (bot *Bot) Start() {
	go serve(bot.ws, bot.doneCh, bot.FinConnCh, bot.inviteHandler, bot.mover, bot.name, bot.limitNoGame)
}

//Stop stops a battleline bot.
func (bot *Bot) Stop() {
	close(bot.doneCh)
	log.Println(log.Verbose, "Bot stopping because of interrupt signal")
	<-bot.FinConnCh
	log.Println(log.Verbose, "Bot stopped")
}

//login logs in to the game server.
func login(scheme, gameURL, name, pw string) (cookies []*http.Cookie, err error) {
	client := new(http.Client) //TODO can not handle https maybe because of a nginx bug solved in 1.11 current docker version is 1.10
	jar, err := cookiejar.New(nil)
	if err != nil {
		return cookies, err
	}
	client.Jar = jar
	resp, err := client.PostForm(scheme+"://"+gameURL+"/post/login",
		url.Values{"txtUserName": {name}, "pwdPassword": {pw}})
	if err != nil {
		return cookies, err
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
		return cookies, err
	}
	loginStatus := tmp.LogInStatus
	if !loginStatus.IsOk() {
		err = fmt.Errorf("Login failed: %v", loginStatus)
		return cookies, err
	}

	cookiesURL, err := url.Parse(scheme + "://" + gameURL + "/in/gamews")
	if err != nil {
		return cookies, err
	}
	cookies = client.Jar.Cookies(cookiesURL) //Strips the path from cookie proberly not a problem
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
	inviteHandler InviteHandler,
	mover Mover,
	name string,
	limitNoGame int) {

	noGame := 0
	noWins := 0
	defer close(finConnCh)
	messCh := make(chan *JsonDataTemp)
	messDoneCh := make(chan struct{})
	go netRead(conn, messCh, messDoneCh)
	var playingData *games.PlayingChData
	pongTimer := time.NewTimer(7 * time.Minute)
Loop:
	for {
		select {
		case <-pongTimer.C:
			act := games.NewAction(games.ACTIDList)
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
				case games.JTList:
					if playingData == nil {
						var list map[string]*games.PubData
						err := json.Unmarshal(jsonDataTemp.Data, &list)
						if err != nil {
							err = errors.Wrap(err, "Unmarshal List failed")
							log.PrintErr(err)
							close(messDoneCh)
							break Loop
						}
						readyOpps := make([]*games.PubData, 0, len(list)-1)
						for _, player := range list {
							if player.Opp == 0 && player.Name != name {
								readyOpps = append(readyOpps, player)
							}
						}
						if len(readyOpps) > 0 {
							invite := inviteHandler.SendInvite(readyOpps)
							if invite != 0 {
								act := games.NewAction(games.ACTIDInvite)
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

				case games.JTInvite:
					var invite games.Invite
					err := json.Unmarshal(jsonDataTemp.Data, &invite)
					if err != nil {
						err = errors.Wrap(err, "Unmarshal Invite failed")
						log.PrintErr(err)
						close(messDoneCh)
						break Loop
					}
					if playingData == nil && inviteHandler.AcceptInvite(&invite) {
						act := games.NewAction(games.ACTIDInvAccept)
						act.ID = invite.InvitorID
						log.Printf(log.DebugMsg, "Accepting invite from %v", invite.InvitorID)
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					} else {
						act := games.NewAction(games.ACTIDInvDecline)
						act.ID = invite.InvitorID
						log.Printf(log.DebugMsg, "Declining invite from %v", invite.InvitorID)
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					}
				case games.JTMess:
				case games.JTPlaying:
					var tmpPlayingData games.PlayingChData
					err := json.Unmarshal(jsonDataTemp.Data, tmpPlayingData)
					if err != nil {
						err = errors.Wrap(err, "Unmarshal playing data failed")
						log.PrintErr(err)
						close(messDoneCh)
						break Loop
					}
					playingData = &tmpPlayingData
					viewPos := playingData.ViewPos
					if playingData == nil {
						noGame = noGame + 1
						if viewPos.LastMoveType == game.MoveTypeAll.Init {
							mover.GameStart(playingData)
						} else {
							mover.GameRestart(playingData)
						}
					}
					if viewPos.Winner < 2 || !viewPos.LastMoveType.HasNext() {
						if viewPos.Winner == viewPos.View.Playerix() {
							noWins = noWins + 1
						}
						if viewPos.LastMoveType.IsPause() {
							mover.GameStop(playingData)
						} else {
							mover.GameFinish(playingData)
						}
						if noGame == limitNoGame {
							close(messDoneCh)
							break Loop
						}
						act := games.NewAction(games.ACTIDList)
						log.Println(log.DebugMsg, "Request list")
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					} else if len(viewPos.Moves) > 0 {
						moveix := mover.Move(viewPos)
						act := games.NewAction(games.ACTIDMove)
						act.Moveix = moveix
						ok := netWrite(conn, act)
						if !ok {
							close(messDoneCh)
							break Loop
						}
					}
				case games.JTWatching:
				case games.JTCloseCon:
					var closeCon games.CloseCon
					err := json.Unmarshal(jsonDataTemp.Data, &closeCon)
					if err == nil {
						log.Printf(log.DebugMsg, "Server closed connection: %v", closeCon.Reason)
					} else {
						err = errors.Wrap(err, "Unmarshal CloseCon failed")
						log.PrintErr(err)
					}
					close(messDoneCh)

					break Loop
				case games.JTClearInvites:
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
	log.Printf(log.Verbose, "Number of game played: %v, number of wins: %v", noGame, noWins)
}

//netWrite writes to the websocket.
func netWrite(conn *websocket.Conn, act *games.Action) bool {
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
type InviteHandler interface {
	AcceptInvite(invite *games.Invite) bool
	SendInvite(readyOpps []*games.PubData) (playerId int)
}
type StdInviteHandler struct {
	isInviter bool
}

func NewStdInviteHandler(isInviter bool) *StdInviteHandler {
	s := new(StdInviteHandler)
	s.isInviter = isInviter
	return s
}
func (s *StdInviteHandler) AcceptInvite(invite *games.Invite) bool {
	return !s.isInviter
}
func (s *StdInviteHandler) SendInvite(readyOpps []*games.PubData) (playerId int) {
	if s.isInviter {
		playerId = readyOpps[0].ID
	}
	return playerId
}

type Mover interface {
	GameStart(playingData *games.PlayingChData)
	GameRestart(playingData *games.PlayingChData)
	GameStop(playingData *games.PlayingChData)
	GameFinish(playingData *games.PlayingChData)
	Move(viewPos *game.ViewPos) (moveis int)
}
