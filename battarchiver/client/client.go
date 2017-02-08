package client

import (
	"github.com/rezder/go-battleline/battarchiver/arnet"
	bat "github.com/rezder/go-battleline/battleline"
	"log"
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
			log.Printf("Creating connection to %v failed %v", serverAddr, err)
		} else {
			serverAddrs = append(serverAddrs, serverAddr)
			conn.Start()
		}
	}
Loop:
	for {
		select {
		case game, open := <-gameCh:
			if !open {
				break Loop
			} else {
				if conn != nil {
					//Assume no buffer on gameCh and brokenCh as it will block on broken or miss games
					select {
					case conn.GameCh <- game:
					case resendGame := <-conn.BrokenCh:
						close(conn.GameCh)
						<-conn.FinCh
						conn.Close()
						conn = nil
						serverAddrs = serverAddrs[1:len(serverAddrs)]
						for len(serverAddrs) > 0 {
							conn, err = arnet.NewZmqSender(serverAddrs[0])
							if err != nil {
								log.Printf("Creating connection to %v failed %v", serverAddr, err)
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
					log.Printf("Creating connection to %v failed %v", serverAddrs[0], err)
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
		conn.Close()
	}
	close(finishCh)

}
func (c *Client) Archive(game *bat.Game) {
	c.gameCh <- game
}

func (c *Client) Stop() {
	err := c.ln.Close()
	if err != nil {
		log.Printf("Closing server join listner failed %v", err)
	}
	close(c.gameCh)
}

func (c *Client) WaitToFinish() {
	<-c.finishCh
}
