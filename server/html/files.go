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
}

func NewPages() (pages *Pages) {
	pages = new(Pages)
	pages.RWMutex = new(sync.RWMutex)
	pages.list = make(map[string][]byte)
	return pages
}
func (pages *Pages) addFile(file string) {
	pages.list[file] = nil
}
func (pages *Pages) addDir(dir string) {
	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err.Error())
	}
	for _, fInfo := range dirInfo {
		if filepath.Ext(fInfo.Name()) == ".html" {
			filepath.Join(dir, fInfo.Name())
			pages.list[filepath.Join(dir, fInfo.Name())] = nil
		}
	}
}
func (pages *Pages) load() {
	for name, _ := range pages.list {
		b, err := ioutil.ReadFile(name)
		if err == nil {
			pages.list[name] = b
		} else {
			panic(err.Error())
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
