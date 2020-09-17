package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"sync"
	"github.com/fsnotify/fsnotify"
)

func watchHostsDir(dirPath string, targetPath string, m *sync.Mutex) {
	// Create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Configure event loop
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					//log.Println("modified file:", event.Name)
					collectHosts(dirPath, targetPath, m)
				}
			}
		}
	}()

	// Configure fsnotify to watch dirpath
	err = watcher.Add(dirpath)
	if err != nil {
		panic(err)
	}

	// Do Initial Run (with mutex to avoid issues)
	go collectHosts(dirPath, targetPath, m)
	
	<-done
}

func collectHosts(dirPath string, targetPath string, m *sync.Mutex) {
	tmpfile, err := ioutil.TempFile("", "hostCollector")
	if err != nil {
		panic(err)
	}
	defer tmpfile.Close()

	m.Lock()

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fullPath := filepath.Join(dirPath, file.Name())
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				panic(err)
			}
			tmpfile.WriteString(mt.Sprintf("# %s\n", fullPath))
			tmpfile.Write(contents)
			tmpfile.Write([]byte("\n"))
		}
	}

	err = tmpfile.Chmod(0644)
	if err != nil {
		panic(err)
	}
	os.Rename(tmpfile.Name(), targetPath)

	m.Unlock()
}

func main() {
	dirPath := flag.String("dirpath", "/etc/hosts.d", "Path containing the individual hosts files")
	targetPath := flag.String("targetpath", "/etc/hosts", "Path to target /etc/hosts file to be written")
	//scanInterval := flag.Int("scaninterval", 5, "Number of seconds between each scan of dirPath")

	flag.Parse()

	var mutex = &sync.Mutex{}

	watchHostsDir(*dirPath, *targetPath, &mutex)
	
	// FIXME: Use inotify!
	//ticker := time.NewTicker(time.Second * time.Duration(*scanInterval))

	//for _ = range ticker.C {
	//	collectHosts(*dirPath, *targetPath, &mutex)
	//}
}
