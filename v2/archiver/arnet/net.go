package arnet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	bg "github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
	"io"
	"net"
)

//PokeListener listen for pokes and send them on.
//Create listner with ln, err := net.Listen("tcp", clientAddr)
//stop with ln.Close()
//should be called in a goroutine.
func PokeListener(ln net.Listener, addCh chan<- string) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			err = errors.Wrap(err, "Poke Accept failed.")
			log.PrintErr(err)
			break
		}
		buf := new(bytes.Buffer)
		no, err := io.Copy(buf, conn)
		if err != nil {
			err = errors.Wrap(err, "PokeListener failed reading.")
			log.PrintErr(err)
			if no == 0 {
				return
			}
		}
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
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			cerr = errors.Wrap(cerr, "Failed closing poke connection.")
			log.PrintErr(cerr)
		}
	}()
	n, err := fmt.Fprintf(conn, myAddr)
	if err != nil {
		err = errors.Wrapf(err, "Poke client write failed sending %v bytes", n)
		log.PrintErr(err)
	}
}

//netZmq zero messages queue.
type netZmq struct {
	context *zmq4.Context
	soc     *zmq4.Socket
}

// Receiver a zmq receiver of game histories.
type Receiver struct {
	*netZmq
	HistCh chan []byte
}

// NewReceiver creates a zmq receiver of game histories.
func NewReceiver(port string) (nz *Receiver, err error) {
	nz = new(Receiver)
	nz.netZmq = new(netZmq)
	nz.context, err = zmq4.NewContext()
	if err != nil {
		return nz, err
	}
	nz.soc, err = nz.context.NewSocket(zmq4.PULL)
	if err != nil {
		closeZmq(nz.netZmq)
		return nz, err
	}
	err = nz.soc.Bind("tcp://*:" + port)
	if err != nil {
		closeZmq(nz.netZmq)
		return nz, err
	}
	nz.HistCh = make(chan []byte, 400)
	return nz, err
}

// Start starts zmg reciver of histories.
func (nz *Receiver) Start() {
	go zmqListen(nz.soc, nz.HistCh)
}

// Close closes the zmq receiver.
func (nz *Receiver) Close() (err error) {

	if nz.context != nil {
		log.Print(log.DebugMsg, "Terminate context")
		err = nz.context.Term()
		log.Print(log.DebugMsg, "Context terminated")
		nz.context = nil
	}
	return err
}

//zmqServe listent for incoming game histories.
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
func zmqListen(receiver *zmq4.Socket, histCh chan<- []byte) {
	var err error
	var msBytes []byte
	for {
		msBytes, err = receiver.RecvBytes(0)
		if err != nil {
			err = errors.Wrap(err, "Closing receiver. Net work receiver zmq failed")
			log.PrintErr(err)
			err = receiver.Close()
			if err != nil {
				log.PrintErr(err)
			}
			log.Print(log.DebugMsg, "Socket closed")
			break
		} else {
			histCh <- msBytes
		}
	}
	close(histCh)
}

// Sender a zmq sender of game history
type Sender struct {
	*netZmq
	finCh    chan struct{}
	HistCh   chan *bg.Hist
	BrokenCh chan *bg.Hist
}

//NewSender creates a zmq sender of game histories.
func NewSender(addr string) (nz *Sender, err error) {
	nz = new(Sender)
	nz.netZmq = new(netZmq)
	nz.context, err = zmq4.NewContext()
	if err != nil {
		return nz, err
	}
	nz.soc, err = nz.context.NewSocket(zmq4.PUSH)
	if err != nil {
		closeZmq(nz.netZmq)
		return nz, err
	}
	err = nz.soc.SetLinger(1) //see http://api.zeromq.org/4-1:zmq-ctx-term
	if err != nil {
		closeZmq(nz.netZmq)
		return nz, err
	}
	err = nz.soc.Connect("tcp://" + addr) //"tcp://localhost:5558"
	if err != nil {
		closeZmq(nz.netZmq)
		return nz, err
	}

	nz.finCh = make(chan struct{})
	nz.BrokenCh = make(chan *bg.Hist)
	nz.HistCh = make(chan *bg.Hist)
	return nz, err
}

// Start starts a zmq sender.
func (nz *Sender) Start() {
	go zmqSend(nz.soc, nz.HistCh, nz.finCh, nz.BrokenCh)
}

// Stop stops a zmq sender.
func (nz *Sender) Stop() {
	log.Print(log.DebugMsg, "Closing zmq sender hist channel")
	close(nz.HistCh)
	select {
	case <-nz.finCh:
	case <-nz.BrokenCh:
		<-nz.finCh
	}
	log.Print(log.DebugMsg, "Closing zmq sender connection")
	closeZmq(nz.netZmq)
	log.Print(log.DebugMsg, "Closed zmq sender connection")
}
func closeZmq(zmq *netZmq) {
	var err error
	if zmq.soc != nil {
		log.Print(log.DebugMsg, "Close socket")
		err = zmq.soc.Close()
		log.Print(log.DebugMsg, "Socket closed")
		zmq.soc = nil
	}
	if err != nil {
		err = errors.Wrap(err, "Closing zmq socket failed.")
		log.PrintErr(err)
	}
	if zmq.context != nil {
		log.Print(log.DebugMsg, "Terminate context")
		err = zmq.context.Term()
		log.Print(log.DebugMsg, "Context terminated")
		zmq.context = nil
	}
	if err != nil {
		err = errors.Wrap(err, "Terminate zmq context failed.")
		log.PrintErr(err)
	}
}
func zmqSend(
	sender *zmq4.Socket,
	histCh <-chan *bg.Hist,
	finCh chan<- struct{},
	brokenCh chan<- *bg.Hist) {

	var err error
	var msBytes []byte
	var hist *bg.Hist
	var open bool
	var n int
	for {
		hist, open = <-histCh
		if open {
			msBytes, err = HistEncode(hist)
			if err != nil {
				err = errors.Wrap(err, "Game history encoding for zmq failed")
				log.PrintErr(err)
			} else {
				log.Print(log.DebugMsg, "Zmq sends game history")
				n, err = sender.SendBytes(msBytes, 0)
				log.Printf(log.DebugMsg, "Zmq send %v bytes", n)
				if err != nil { //TODO MAYBE resend.
					brokenCh <- hist
				}
			}
		} else {
			break
		}
	}
	close(finCh)
}

// HistDecoder decodes game history.
func HistDecoder(histBytes []byte) (hist *bg.Hist, err error) {
	buf := bytes.NewBuffer(histBytes)
	decoder := gob.NewDecoder(buf)
	h := *new(bg.Hist)
	err = decoder.Decode(&h)
	if err != nil {
		return hist, err
	}
	hist = &h
	return hist, err
}

// HistEncode encodes game history.
func HistEncode(hist *bg.Hist) (histBytes []byte, err error) {
	var histBuf bytes.Buffer
	encoder := gob.NewEncoder(&histBuf)
	err = encoder.Encode(hist)
	if err != nil {
		return histBytes, err
	}
	histBytes = histBuf.Bytes()
	return histBytes, err
}
