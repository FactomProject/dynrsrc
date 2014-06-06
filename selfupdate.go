package dynrsrc

import (
	"io/ioutil"
	"path/filepath"
	
	"github.com/howeyc/fsnotify"
)

var done = false
var watcher *fsnotify.Watcher
var dynrsrcs = make(map[string]func([]byte))

func Start(watchEH, readEH func(error)) (err error) {
	watcher, err = fsnotify.NewWatcher()
	
	if err == nil {
		go process(watchEH, readEH)
	}
	
	return
}

func process(watchEH, readEH func(error)) {
	for done {
		select {
		case event := <-watcher.Event:
			if handler, ok := dynrsrcs[event.Name]; ok {
				file, err := ioutil.ReadFile(event.Name)
				if err != nil {
					readEH(err)
					continue
				}
				handler(file)
			}
			
		case err := <-watcher.Error:
			watchEH(err)
		}
	}
}

func CreateDynamicResource(filename string, handler func([]byte)) error {
	dynrsrcs[filename] = handler
	
	file, err := ioutil.ReadFile(filename)
	if err != nil { return err }
	handler(file)
	
	return watcher.Watch(filepath.Dir(filename))
}

func DestroyDynamicResource(filename string) {
	delete(dynrsrcs, filename)
}

func Stop() {
	done = true
	watcher.Close()
}