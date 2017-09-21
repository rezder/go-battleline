package viewer

import (
	"bufio"
	"encoding/json"
	//"fmt"
	"github.com/pkg/errors"
	"github.com/rezder/go-battleline/battbot/tf"
	"github.com/rezder/go-battleline/v2/game"
	"github.com/rezder/go-error/log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type modelHandler struct {
	server *Server
}

// ServeHTTP handles the model requests.
// there is 3 types:
//1) Load model file.
//2) Return probabilities.
//3) Check StdOut and StdErr
func (m *modelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	model, moverView, isStdOut, err := parseModelFormValues(r)
	if err != nil {
		log.PrintErr(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if moverView != nil {
		probs, err := m.server.ReqProbs(model, moverView)
		if err != nil {
			log.PrintErr(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpWrite(probs, w)
	} else if isStdOut {
		stdOut := m.server.UpdateStdOut(model)
		httpWrite(struct{ StdOut string }{StdOut: stdOut}, w)
	} else {
		err := m.server.UpdateModel(model)
		if err != nil {
			log.PrintErr(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpWrite(struct{ Ok bool }{Ok: true}, w)
	}
}
func parseModelFormValues(r *http.Request) (model string, moverView *game.ViewPos, isStdOut bool, err error) {
	model = r.FormValue("model")
	if len(model) == 0 {
		return model, moverView, isStdOut, errors.New("Empty model")
	}
	moverViewJSON := r.FormValue("mover-view")
	if len(moverViewJSON) > 0 {
		var mv game.ViewPos
		err = json.Unmarshal([]byte(moverViewJSON), &mv)
		if err != nil {
			return model, moverView, isStdOut, err
		}
		moverView = &mv
	} else {
		isStdOutTxt := r.FormValue("std-out")
		if len(isStdOutTxt) > 0 {
			isStdOut = true
		}
	}
	log.Printf(log.DebugMsg, "Request Model: model: %v,moverView: %v, isStdOut: %v", model, moverView, isStdOut)
	return model, moverView, isStdOut, err
}

//ModelCont model controler contains the data to control tensorflow
//models.
type ModelCont struct {
	cmd      *exec.Cmd
	outData  string
	port     int
	modelDir string
	con      *tf.Con
	stdOutCh <-chan string
	stdErrCh <-chan string
}

//NewModelCont creates a model controler.
func NewModelCont(modelDir string, port int) (mc *ModelCont, err error) {
	cpFile := filepath.Join(modelDir, "checkpoint")
	_, err = os.Stat(cpFile)
	if err != nil {
		err = errors.Wrapf(err, "Can not access file: %v in directory: %v", modelDir, cpFile)
		return mc, err
	}

	mc = new(ModelCont)
	mc.modelDir = modelDir
	mc.port = port
	mc.cmd = exec.Command("python", "/home/rho/Python/tensorflow/battleline/botserver.py", "--model_dir="+modelDir, "--port="+strconv.Itoa(port))

	stdErrCh := make(chan string, 100)
	stdOutCh := make(chan string, 100)
	stdOutPipe, err := mc.cmd.StdoutPipe()
	if err != nil {
		err = errors.Wrapf(err, "Connecting to stdOut failed for command: %v", mc.cmd.Args)
		return mc, err
	}
	stdErrPipe, err := mc.cmd.StderrPipe()
	if err != nil {
		err = errors.Wrapf(err, "Connecting to stdErr failed for command: %v", mc.cmd.Args)
		return mc, err
	}
	err = mc.cmd.Start()
	if err != nil {
		err = errors.Wrapf(err, "Starting command:%v failed", mc.cmd.Args)
		return mc, err
	}
	log.Printf(log.DebugMsg, "Command %v started, sleep 2 seconds", mc.cmd.Args)
	time.Sleep(time.Second * 2)
	mc.con, err = tf.New("localhost:" + strconv.Itoa(mc.port))
	if err != nil {
		mc.con = nil
		mc.Stop()
		err = errors.Wrapf(err, "Creating Zmq connection port %v failed", mc.port)
	}
	mc.stdOutCh = stdOutCh
	mc.stdErrCh = stdErrCh
	go readFromPipe(stdOutCh, stdOutPipe)
	go readFromPipe(stdErrCh, stdErrPipe)
	return mc, err
}
func capLog(bs []byte, n int) (txt string) {
	logCap := 50
	if n < logCap {
		logCap = n
	}
	return string(bs[:logCap])
}
func capLogTxt(logTxt string) (txt string) {
	logCap := 50
	if len(logTxt) < logCap {
		return logTxt
	}
	return logTxt[:logCap]
}
func readFromPipe(ch chan<- string, pipe io.ReadCloser) {
	buf := bufio.NewReader(pipe)
	bufSize := 1024
	for {
		bs := make([]byte, bufSize)
		n, err := buf.Read(bs)
		log.Printf(log.DebugMsg, "Read from pipe sample: %v, Error %v", capLog(bs, n), err)
		if err != nil {
			close(ch)
			if n > 0 {
				ch <- string(bs[:n])
			}
			if err != io.EOF {
				log.PrintErr(err)
			} else {
				log.Println(log.DebugMsg, "Closing pipe EOF reached")
			}
			break
		}
		ch <- string(bs[:n])
		if n != bufSize {
			time.Sleep(time.Second)
		}
	}
}

//Stop stops the model server.
func (mc *ModelCont) Stop() {
	if mc.con != nil {
		err := mc.con.Close()
		if err != nil {
			log.PrintErr(err)
		}
	}
	err := mc.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		log.PrintErr(err)
	}
	for isClosed := false; !isClosed; _, isClosed = mc.Read() {
		time.Sleep(time.Millisecond * 100)
	}
	err = mc.cmd.Wait()
	if err != nil {
		log.PrintErr(errors.Wrapf(err, "The command: %v exits code", mc.cmd.Args))
	}
}

//Request requests the probabilities for moves.
func (mc *ModelCont) Request(moverView *game.ViewPos) (probs []float64, err error) {

	noMoves := len(moverView.Moves)
	//bs := make([]byte, 0, 100)//TODO convert new game ViewPos to model data. When machine is change or not
	//probs, err = mc.con.ReqProba(bs, noMoves)
	probs = make([]float64, noMoves)
	for i := range probs {
		probs[i] = float64(i) / float64(noMoves)
	}
	return probs, err
}
func (mc *ModelCont) Read() (outData string, isClosed bool) {
	isOutClosed := true
	isErrClosed := true
	if mc.stdOutCh != nil {
		var outTxt string
		outTxt, isOutClosed = readFromCh(mc.stdOutCh)
		mc.outData = mc.outData + outTxt
		if isOutClosed {
			mc.stdOutCh = nil
		}
	}
	if mc.stdErrCh != nil {
		var errTxt string
		errTxt, isErrClosed = readFromCh(mc.stdErrCh)
		mc.outData = mc.outData + errTxt
		if isErrClosed {
			mc.stdErrCh = nil
		}
	}
	isClosed = isErrClosed && isOutClosed
	return mc.outData, isClosed
}
func readFromCh(ch <-chan string) (data string, isClosed bool) {
	isClosed = true
	if ch != nil {
	Loop:
		for {
			select {
			case txt, isOpen := <-ch:
				if isOpen {
					data = data + txt
					log.Printf(log.DebugMsg, "Read from channel sample: %v", capLogTxt(txt))
					isClosed = false
				} else {
					break Loop
				}
			default:
				isClosed = false
				break Loop
			}
		}
	}
	return data, isClosed
}
