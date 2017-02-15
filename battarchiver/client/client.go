package client

import (
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/arnet"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"net"
	"strconv"
)

type Client struct {
	serverAdrr string
	gameCh     chan *bat.Game
	finishCh   chan struct{}
	ln         net.Listener
}

func New(port int, serverAddr string) (c *Client, err error) {
	c = new(Client)
	c.serverAdrr = serverAddr
	c.gameCh = make(chan *bat.Game, 25)
	c.finishCh = make(chan struct{})
	c.ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return c, err
	}
	return c, err
}
func (c *Client) Start() {
	addServerCh := make(chan string)
	go arnet.PokeListener(c.ln, addServerCh)
	go start(c.serverAdrr, c.gameCh, addServerCh, c.finishCh)
}

func start(
	serverAddr string,
	gameCh <-chan *bat.Game,
	addServerCh <-chan string,
	finishCh chan<- struct{}) {

	var conn *arnet.NetZmqSender

	serverAddrs := make([]string, 0, 2)
	var err error
	if len(serverAddr) != 0 {
		conn, err = arnet.NewZmqSender(serverAddr)
		if err != nil {
			err = errors.Wrapf(err, "Creating connection to %v failed", serverAddr)
			log.PrintErr(err)
		} else {
			serverAddrs = append(serverAddrs, serverAddr)
			conn.Start()
		}
	}
Loop:
	for {
		select {
		case game, open := <-gameCh: //TODO what if 25 game buffer is to small and it blocks because zmq is to slow. Should it close down archiver and try a new one. Hint check len(gameCh and close if full)
			if !open {
				break Loop
			} else {
				if conn != nil {
					//Assume no buffer on gameCh and brokenCh as it will block on broken or miss games
					select {
					case conn.GameCh <- game:
					case resendGame := <-conn.BrokenCh:
						log.Printf(log.DebugMsg, "Zmq broken closing game channel on %v", serverAddrs[0])
						close(conn.GameCh)
						<-conn.FinCh
						log.Printf(log.DebugMsg, "Zmq broken %v return done close connection.", serverAddrs[0])
						conn.Close()
						conn = nil
						serverAddrs = serverAddrs[1:len(serverAddrs)]
						for len(serverAddrs) > 0 {
							log.Printf(log.DebugMsg, "Old zmq broken starting %v", serverAddrs[0])
							conn, err = arnet.NewZmqSender(serverAddrs[0])
							if err != nil {
								err = errors.Wrapf(err, "Creating connection to %v failed", serverAddr)
								log.PrintErr(err)
								serverAddrs = serverAddrs[1:len(serverAddrs)]
							} else {
								break
							}
						}
						if conn != nil {
							conn.Start()
							conn.GameCh <- resendGame
						}
					}

				}
			}
		case server := <-addServerCh:
			serverAddrs = append(serverAddrs, server)
			if len(serverAddrs) == 1 {
				conn, err = arnet.NewZmqSender(serverAddrs[0])
				if err != nil {
					err = errors.Wrapf(err, "Creating connection to %v failed", serverAddrs[0])
					log.PrintErr(err)
					serverAddrs = serverAddrs[1:len(serverAddrs)]
				} else {
					conn.Start()
				}
			}
		}
	}
	if conn != nil {
		close(conn.GameCh)
		select {
		case <-conn.FinCh:
		case <-conn.BrokenCh:
			<-conn.FinCh
		}
		log.Print(log.DebugMsg, "Closing zmq connection")
		conn.Close()
		log.Print(log.DebugMsg, "Closed zmq connection")
	}
	close(finishCh)

}
func (c *Client) Archive(game *bat.Game) {
	c.gameCh <- game
}

func (c *Client) Stop() {
	err := c.ln.Close()
	if err != nil {
		err = errors.Wrap(err, "Closing server, join listner failed")
		log.PrintErr(err)
	}
	close(c.gameCh)
}

func (c *Client) WaitToFinish() {
	<-c.finishCh
}
