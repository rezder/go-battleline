package http

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// Pages is a html file cache with a read write log.
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

//addFile adds a file to the cache.
func (pages *Pages) addFile(file string) {
	pages.list[file] = nil
}

//addDir adds all html files from a directory
func (pages *Pages) addDir(dir string) (err error) {
	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		err = errors.Wrapf(err, "Error reading directory %v", dir)
		return err
	}
	for _, fInfo := range dirInfo {
		if filepath.Ext(fInfo.Name()) == ".html" {
			filepath.Join(dir, fInfo.Name())
			pages.list[filepath.Join(dir, fInfo.Name())] = nil
		}
	}
	return err
}

//load the files.
func (pages *Pages) load() (err error) {
	var b []byte
	for name := range pages.list {
		b, err = ioutil.ReadFile(name)
		if err != nil {
			err = errors.Wrapf(err, "Error reading file %v", name)
			return err
		}
		pages.list[name] = b

	}
	return err
}

//loadLock loads the file before loading activate the write lock.
func (pages *Pages) loadLock() error {
	pages.Lock()
	defer pages.Unlock()
	return pages.load()
}

//readPage reads a page using read lock.
func (pages *Pages) readPage(page string) (res []byte) {
	pages.RLock()
	res, found := pages.list[page]
	pages.RUnlock()
	if !found {
		panic(fmt.Sprintf("File %v do not exist", page))
	}
	return res
}
