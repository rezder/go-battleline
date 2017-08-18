package viewer

import (
	"fmt"
	"github.com/rezder/go-error/log"
	"testing"
	"time"
)

func TestModelCont(t *testing.T) {
	log.InitLog(log.Debug)
	modelDir := "/home/rho/Python/tensorflow/battleline/model/"
	port := 6555
	mc, err := NewModelCont(modelDir, port)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 10)
	fmt.Println("Reading +++++++++++++++++++++")
	txt, isClosed := mc.Read()
	//	t.Log(txt)
	if isClosed {
		t.Error("Pipes should not be closed")
	}
	if len(txt) == 0 {
		t.Error("Should read something")
	}
	fmt.Println("Reading +++++++++++++++++++++")
	txt2, isClosed2 := mc.Read()
	fmt.Println("+++++++++++Reading  End+++++++++++++++++++++")
	if txt != txt2 {
		t.Errorf("Should read the same %v,%v", txt, txt2)
	}
	//t.Error("forced")
	if isClosed2 {
		t.Error("Pipes should not be closed")
	}
	fmt.Println("+++++++++++Stoping+++++++++++++++++++++")
	mc.Stop()
	fmt.Println("+++++++++++End stoping+++++++++++++++++++++")
}
