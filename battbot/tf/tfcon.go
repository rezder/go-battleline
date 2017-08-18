package tf

import "github.com/pebbe/zmq4"
import "github.com/rezder/go-error/log"
import "encoding/binary"
import "bytes"
import "github.com/pkg/errors"
import "time"

// Con a Zmq client connection for tensorflow.
type Con struct {
	context *zmq4.Context
	socket  *zmq4.Socket
}

// New create connection.
func New(tfaddr string) (tfc *Con, err error) {
	tfc = new(Con)
	tfc.context, err = zmq4.NewContext()
	if err != nil {
		err = errors.Wrap(err, "Error creating zmq context")
		return tfc, err
	}
	tfc.socket, err = tfc.context.NewSocket(zmq4.REQ)
	tfc.socket.SetRcvtimeo(time.Second * 10) //return error with ErrNo EAGAIN
	if err != nil {
		_ = tfc.context.Term()
		tfc.context = nil
		err = errors.Wrap(err, "Error creating zmq socket")
		return tfc, err
	}
	err = tfc.socket.SetLinger(1)
	if err != nil {
		_ = tfc.socket.Close()
		_ = tfc.context.Term()
		tfc.socket = nil
		tfc.context = nil
		err = errors.Wrap(err, "Error setting linger on zmq socket")
	}
	addrtxt := "tcp://" + tfaddr
	err = tfc.socket.Connect(addrtxt) //"tcp://localhost:5558"
	if err != nil {
		_ = tfc.socket.Close()
		_ = tfc.context.Term()
		tfc.socket = nil
		tfc.context = nil
		err = errors.Wrapf(err, "Error binding zmq socket to %v", addrtxt)
		return tfc, err
	}
	return tfc, err
}

// Close close connection.
func (tfc *Con) Close() (err error) {
	if tfc.socket != nil {
		log.Print(log.DebugMsg, "Zmq close socket")
		err = tfc.socket.Close()
		log.Print(log.DebugMsg, "Zmq socket closed")
		tfc.socket = nil
	}
	if tfc.context != nil {
		log.Print(log.DebugMsg, "Zmq terminate context")
		err = tfc.context.Term()
		log.Print(log.DebugMsg, "Zmq context terminated")
		tfc.context = nil
	}
	return err
}

//ReqProba request probabilities for every possible move.
func (tfc *Con) ReqProba(data []byte, noMoves int) (proba []float64, err error) {
	log.Print(log.DebugMsg, "Zmq sends moves")
	n, err := tfc.socket.SendBytes(data, 0)
	log.Printf(log.DebugMsg, "Zmq send %v bytes", n)
	if err != nil {
		err = errors.Wrap(err, "Zmq error sending bytes")
		_ = tfc.Close()
		return proba, err
	}
	proba = make([]float64, noMoves)
	b, err := tfc.socket.RecvBytes(0)
	log.Printf(log.DebugMsg, "Zmq recieved %v bytes", len(b))
	if err != nil {
		err = errors.Wrap(err, "Zmq error receiving bytes")
		_ = tfc.Close()
		return proba, err
	}
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.LittleEndian, proba)
	if err != nil {
		err = errors.Wrap(err, "Error converting bytes to probabilities")
	}
	log.Printf(log.DebugMsg, "Zmq received %v", proba)
	return proba, err
}
