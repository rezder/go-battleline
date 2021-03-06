package arnet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battarchiver/battdb"
	bat "github.com/rezder/go-battleline/battleline"
	"github.com/rezder/go-error/log"
	"io"
	"net"
	"net/http"
)

//PokeListener listen for pokes and send them on.
//Create listner with ln, err := net.Listen("tcp", clientAddr)
//stop with ln.Close()
//should be called in a goroutine.
func PokeListener(ln net.Listener, addCh chan<- string) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			err = errors.Wrap(err, "Poke Accept failed")
			log.PrintErr(err)
			break
		}
		buf := new(bytes.Buffer)
		io.Copy(buf, conn)
		addCh <- buf.String()
	}
}

//PokeClient poke the client with address to signal it is ready.
func PokeClient(clientAdr, myAddr string) {
	//adr  "golang.org:80"
	conn, err := net.Dial("tcp", clientAdr)
	if err != nil {
		err = errors.Wrap(err, "Poke client dial failed")
		log.PrintErr(err)
		return
	}
	defer conn.Close()
	n, err := fmt.Fprintf(conn, myAddr)
	if err != nil {
		err = errors.Wrapf(err, "Poke client write failed sending %v bytes", n)
		log.PrintErr(err)
	}
}

//StartBackUpServer starts backup server.
func StartBackUpServer(bdb *battdb.Db, port string) {
	http.HandleFunc("/backup",
		func(resp http.ResponseWriter, req *http.Request) {
			bdb.BackupHandleFunc(resp, req)
		})
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		err = errors.Wrap(err, "Backup http server failed")
		log.PrintErr(err)
	}
}

//NetZmq zero messages queue.
type NetZmq struct {
	context *zmq4.Context
	soc     *zmq4.Socket
}

type NetZmqReciver struct {
	NetZmq
	GameCh chan []byte
}

func NewZmqReciver(port string) (nz *NetZmqReciver, err error) {
	nz = new(NetZmqReciver)
	nz.context, err = zmq4.NewContext()
	if err != nil {
		return nz, err
	}
	nz.soc, err = nz.context.NewSocket(zmq4.PULL)
	if err != nil {
		nz.context.Term() // ignore error
		nz.context = nil
		return nz, err
	}
	err = nz.soc.Bind("tcp://*:" + port)
	if err != nil {
		nz.soc.Close()
		nz.context.Term()
		nz.soc = nil
		nz.context = nil
		return nz, err
	}
	nz.GameCh = make(chan []byte, 400)
	return nz, err
}

func (nz *NetZmqReciver) Start() {
	go zmqListen(nz.soc, nz.GameCh)
}
func (nz *NetZmqReciver) Close() (err error) {

	if nz.context != nil {
		log.Print(log.DebugMsg, "Terminate context")
		err = nz.context.Term()
		log.Print(log.DebugMsg, "Context terminated")
		nz.context = nil
	}
	return err
}

//zmqServe listent for incoming games.
//The recv call is blocking only killing context will free it if there is no messages.
//
//Setting it to not blocking zmq4.DONTWAIT and using zmq4.Errno==11(EAGAIN)
//to check for continue would allow for stoping the loop. Close socket would then
//be able to close the loop.
//We could set and check a atomic/sync/channel but it is alot to avoid the error
//message when closing.
//
//If there is a error on the net work it closes down it may be to harsh
//and we need to have atomic/sync/channel variable to check anyway.
func zmqListen(receiver *zmq4.Socket, gameCh chan<- []byte) {
	var err error
	var msBytes []byte
	for {
		msBytes, err = receiver.RecvBytes(0)
		if err != nil {
			err = errors.Wrap(err, "Closing receiver. Net work receiver zmq failed")
			log.PrintErr(err)
			receiver.Close()
			log.Print(log.DebugMsg, "Socket closed")
			break
		} else {
			gameCh <- msBytes
		}
	}
	close(gameCh)
}

type NetZmqSender struct {
	NetZmq
	FinCh    chan struct{}
	GameCh   chan *bat.Game
	BrokenCh chan *bat.Game
}

func NewZmqSender(addr string) (nz *NetZmqSender, err error) {
	nz = new(NetZmqSender)
	nz.context, err = zmq4.NewContext()
	if err != nil {
		return nz, err
	}
	nz.soc, err = nz.context.NewSocket(zmq4.PUSH)
	if err != nil {
		nz.context.Term() // ignore error
		nz.context = nil
		return nz, err
	}
	nz.soc.SetLinger(1)                   //see http://api.zeromq.org/4-1:zmq-ctx-term
	err = nz.soc.Connect("tcp://" + addr) //"tcp://localhost:5558"
	if err != nil {
		nz.soc.Close()
		nz.context.Term()
		nz.soc = nil
		nz.context = nil
		return nz, err
	}
	nz.FinCh = make(chan struct{})
	nz.BrokenCh = make(chan *bat.Game)
	nz.GameCh = make(chan *bat.Game)
	return nz, err
}

func (nz *NetZmqSender) Start() {
	go zmqSend(nz.soc, nz.GameCh, nz.FinCh, nz.BrokenCh)
}
func (nz *NetZmqSender) Close() (err error) {
	if nz.soc != nil {
		log.Print(log.DebugMsg, "Close socket")
		err = nz.soc.Close()
		log.Print(log.DebugMsg, "Socket closed")
		nz.soc = nil
	}
	if nz.context != nil {
		log.Print(log.DebugMsg, "Terminate context")
		err = nz.context.Term()
		log.Print(log.DebugMsg, "Context terminated")
		nz.context = nil
	}
	return err
}
func zmqSend(
	sender *zmq4.Socket,
	gameCh <-chan *bat.Game,
	finCh chan<- struct{},
	brokenCh chan<- *bat.Game) {

	var err error
	var msBytes []byte
	var game *bat.Game
	var open bool
	var n int
	for {
		game, open = <-gameCh
		if open {
			msBytes, err = ZmqEncode(game)
			if err != nil {
				err = errors.Wrap(err, "Game encoding for zmq failed")
				log.PrintErr(err)
			} else {
				log.Print(log.DebugMsg, "Zmq sends game")
				n, err = sender.SendBytes(msBytes, 0)
				log.Printf(log.DebugMsg, "Zmq send %v bytes", n)
				if err != nil { //TODO MAYBE resend.
					brokenCh <- game
				}
			}
		} else {
			break
		}
	}
	close(finCh)
}
func ZmqDecoder(gameBytes []byte) (game *bat.Game, err error) {
	buf := bytes.NewBuffer(gameBytes)
	decoder := gob.NewDecoder(buf)
	g := *new(bat.Game)
	err = decoder.Decode(&g)
	if err != nil {
		return game, err
	}
	game = &g
	return game, err
}
func ZmqEncode(game *bat.Game) (gameBytes []byte, err error) {
	pos := game.Pos
	game.Pos = nil
	var gameBuf bytes.Buffer
	encoder := gob.NewEncoder(&gameBuf)
	err = encoder.Encode(game)
	game.Pos = pos
	if err != nil {
		return gameBytes, err
	}
	gameBytes = gameBuf.Bytes()
	return gameBytes, err
}
