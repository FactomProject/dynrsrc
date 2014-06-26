package dynrsrc

import (
	"errors"
	"os"
	
	"io/ioutil"
	"path/filepath"
	
	"github.com/howeyc/fsnotify"
)

var state = 0
var watcher *fsnotify.Watcher
var dynfiles = make(map[string]func([]byte))
var dyndirs = make(map[string]func([]byte))

func Start(watchEH, readEH func(error)) (err error) {
	if state != 0 { return }
	
	watcher, err = fsnotify.NewWatcher()
	
	if err == nil {
		state = 1
		go process(watchEH, readEH)
	}
	
	return
}

func process(watchEH, readEH func(error)) {
	for state == 1 {
		select {
		case event := <-watcher.Event:
			if (!event.IsModify()) {
				continue
			}
			
			if handler, ok := dynfiles[event.Name]; ok {
				file, err := ioutil.ReadFile(event.Name)
				if err != nil {
					readEH(err)
					continue
				}
				handler(file)
			}
			if handler, ok := dyndirs[filepath.Dir(event.Name)]; ok {
				handler(nil)
			}
			
		case err := <-watcher.Error:
			watchEH(err)
		}
	}
}

func CreateDynamicResource(filename string, handler func([]byte)) error {
	if state != 1 {
		return errors.New("Dynrsrc not started")
	}
	
	finfo, err := os.Stat(filename)
	if err != nil { return err }
	
	err = watcher.Watch(filename)
	if err != nil { return err }
	
	if finfo.IsDir() {
		dyndirs[filename] = handler
		handler(nil)
		return nil
	}
	
	dynfiles[filename] = handler
	
	file, err := ioutil.ReadFile(filename)
	if err != nil { return err }
	handler(file)
	
	return nil
}

func DestroyDynamicResource(filename string) {
	delete(dynfiles, filename)
	delete(dyndirs, filename)
}

func Stop() {
	state = 2
	watcher.Close()
}