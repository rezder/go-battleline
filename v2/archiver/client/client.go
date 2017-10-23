package client

import (
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/v2/archiver/arnet"
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
	"net"
	"strconv"
)

//Client handle connection to game history archiver.
type Client struct {
	archiverAdr  string
	histCh       chan *bg.Hist
	finishCh     chan struct{}
	pokeListener net.Listener
}

//New create client that handles connection to an archiver.
func New(pokePort int, achiverAdr string) (c *Client, err error) {
	c = new(Client)
	c.archiverAdr = achiverAdr
	c.histCh = make(chan *bg.Hist, 25)
	c.finishCh = make(chan struct{})
	c.pokeListener, err = net.Listen("tcp", ":"+strconv.Itoa(pokePort))
	return c, err
}

//Start stars the connection to archiver.
func (c *Client) Start() {
	addServerCh := make(chan string)
	go arnet.PokeListener(c.pokeListener, addServerCh)
	go start(c.archiverAdr, c.histCh, addServerCh, c.finishCh)
}
func startZmqSender(archiverAdrs []string) (*arnet.Sender, []string) {
	var conn *arnet.Sender
	var err error
	for len(archiverAdrs) > 0 {
		log.Printf(log.DebugMsg, "Starting zmq sender %v", archiverAdrs[0])
		conn, err = arnet.NewSender(archiverAdrs[0])
		if err != nil {
			err = errors.Wrapf(err, "Creating connection to %v failed", archiverAdrs[0])
			log.PrintErr(err)
			archiverAdrs = archiverAdrs[1:]
		} else {
			break
		}
	}
	if conn != nil {
		conn.Start()
	}
	return conn, archiverAdrs
}
func start(
	serverAddr string,
	histCh <-chan *bg.Hist,
	addServerCh <-chan string,
	finishCh chan<- struct{}) {

	var archConn *arnet.Sender

	archiverAdrs := make([]string, 0, 2)
	if len(serverAddr) != 0 {
		archiverAdrs = append(archiverAdrs, serverAddr)
		archConn, archiverAdrs = startZmqSender(archiverAdrs)
	}
Loop:
	for {
		select {
		case hist, open := <-histCh: //TODO what if 25 game histories buffer is to small and it blocks because zmq is to slow. Should it close down archiver and try a new one. Hint check len(histCh and close if full)
			if !open {
				break Loop
			} else {
				if archConn != nil {
					//Assume no buffer on histCh and brokenCh as it will block on broken or miss game histories
					select {
					case archConn.HistCh <- hist:
					case resendHist := <-archConn.BrokenCh:
						log.Printf(log.DebugMsg, "Zmq broken closing history channel on %v", archiverAdrs[0])
						archConn.Stop()
						archConn = nil
						archiverAdrs = archiverAdrs[1:]
						archConn, archiverAdrs = startZmqSender(archiverAdrs)
						if archConn != nil {
							archConn.HistCh <- resendHist
						}
					}

				}
			}
		case server := <-addServerCh:
			archiverAdrs = append(archiverAdrs, server)
			if len(archiverAdrs) == 1 {
				archConn, archiverAdrs = startZmqSender(archiverAdrs)
			}
		}
	}
	if archConn != nil {
		archConn.Stop()
	}
	close(finishCh)

}

//Archive send a game history to the archiver.
func (c *Client) Archive(hist *bg.Hist) {
	c.histCh <- hist
}

//Stop stops the connection to the archiver.
func (c *Client) Stop() {
	err := c.pokeListener.Close()
	if err != nil {
		err = errors.Wrap(err, "Closing server, join listner failed.")
		log.PrintErr(err)
	}
	log.Print(log.DebugMsg, "Closing game history channel.")
	close(c.histCh)
	<-c.finishCh
	log.Print(log.DebugMsg, "Archiver client finished.")

}
