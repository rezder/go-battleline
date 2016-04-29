package html

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type Pages struct {
	*sync.RWMutex
	list map[string][]byte
	dir  string
}

func NewPages(dir string) (pages *Pages) {
	pages = new(Pages)
	pages.dir = dir + "/"
	pages.RWMutex = new(sync.RWMutex)
	return pages
}
func (pages *Pages) addFile(file string) {
	pages.list[file] = nil
}
func (pages *Pages) load() {
	pages.list = make(map[string][]byte)
	files, err := ioutil.ReadDir(pages.dir)
	if err != nil {
		panic(err.Error())
	}
	for _, fInfo := range files {
		if filepath.Ext(fInfo.Name()) == ".html" {
			b, err := ioutil.ReadFile(pages.dir + fInfo.Name())
			if err == nil {
				pages.list[fInfo.Name()] = b
			} else {
				panic(err.Error())
			}
		}
	}

}
func (pages *Pages) loadLock() {
	pages.Lock()
	defer pages.Unlock()
	pages.load()
}
func (pages *Pages) readPage(page string) (res []byte) {
	pages.RLock()
	res, found := pages.list[page]
	pages.RUnlock()
	if !found {
		panic(fmt.Sprintf("File %v do not exist", page))
	}
	return res
}
