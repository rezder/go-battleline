package html

import (
	"bytes"
	"encoding/gob"
	"golang.org/x/net/html"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseHtml(t *testing.T) {
	fileName := "pages/client.html"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("Load file error: %v\n", err.Error())
	}
	reader := bytes.NewReader(data)
	startNode, err := html.Parse(reader)
	if err != nil {
		t.Errorf("Load file error: %v\n", err.Error())
	}
	addNode := createPTextNode("Test adding text")
	t.Logf("Start node: %+v\nAddNode: %+v\n", startNode, addNode)
	found := addPTextNode(startNode, addNode)
	if !found {
		t.Errorf("Body was not found in file")
	}
	file, err := os.Create("test/client.html")
	if err == nil {
		defer file.Close()
		err = html.Render(file, startNode)
		if err != nil {
			t.Errorf("Error rending html file: %v", err.Error())
		}
	} else {
		t.Errorf("Error creating file", err.Error())
	}
}
func TestGobClient(t *testing.T) {
	c := createClient("Rene", 1, nil)
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	// Encoding the map
	err := e.Encode(c)
	if err != nil {
		t.Errorf("Error encoding: %v", err)
	}

	var loadC Client
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	err = d.Decode(&loadC)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	}

}
