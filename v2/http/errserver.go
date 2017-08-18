package http

import (
	"github.com/rezder/go-error/log"
)

// ErrServer handles while the servers are running.
// All errors are send here so the error server can deside
// what to to do. Errors during close down does not need
// to be send here.
type ErrServer struct {
	clients *Clients
	ch      chan error
	finCh   chan struct{}
}

//NewErrServer creates a new error server
func NewErrServer(clients *Clients) (e *ErrServer) {
	e = new(ErrServer)
	e.ch = make(chan error, 10)
	e.clients = clients
	e.finCh = make(chan struct{})
	return e
}

//Ch returns the channel errors should be send on.
func (e *ErrServer) Ch() chan<- error {
	return e.ch
}

//Start starts the server.
func (e *ErrServer) Start() {
	go errServe(e.ch, e.finCh)
}

//Stop stopd the server.
func (e *ErrServer) Stop() {
	close(e.ch)
	<-e.finCh
}

//errServe runs a error server.
//all errors should be send where the power to close down exist.
//Currently it does nothing but log the errors.
func errServe(errChan chan error, finCh chan struct{}) {
	// TODO Add error count on player id and auto disable player with to many errors.
	for {
		err, open := <-errChan
		if open {
			log.PrintErr(err)
		} else {
			close(finCh)
			break
		}
	}
}
