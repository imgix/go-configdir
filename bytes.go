/*
Package configdir monitors a config directory for a supplied file suffix. Channel updates
return bytes.Buffer containing all matching files concatenated together.

Updates return
only if the update contains a unique byte array from previous runs by calculating md5 checksums.
*/
package configdir

import (
	"gopkg.in/fsnotify.v1"

	"bytes"
	"crypto/md5"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//Printer interface just to make a passable Logger
//
//
type Printer interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

var logger Printer

//DirectoryUpdates starts a goroutine and returns bytes.Buffer receive channel
//
//
func DirectoryUpdates(dir, suffix string, p Printer) (chan *bytes.Buffer, error) {
	if p != nil {
		logger = p
	} else {
		logger = log.New(os.Stderr, "[dircfg]",
			log.Ldate|log.Ltime|log.Lshortfile,
		)

	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(dir)
	if err != nil {
		return nil, err
	}

	dirBytesCh := make(chan *bytes.Buffer)

	go func() {
		var (
			dirBytesB          *bytes.Buffer
			currentMD5, tmpMD5 []byte
		)
		dirBytesB, tmpMD5 = bytesFromDir(dir, suffix)
		currentMD5 = make([]byte, len(tmpMD5))
		copy(currentMD5, tmpMD5)
		dirBytesCh <- dirBytesB

		evI := func(o fsnotify.Op) bool {
			switch {
			case o&fsnotify.Write == fsnotify.Write:
				return true
			case o&fsnotify.Rename == fsnotify.Rename:
				return true

			}

			return false
		}
		updates := 0

		for {
			select {
			case ev := <-watcher.Events:

				if strings.HasSuffix(ev.Name, suffix) && evI(ev.Op) {
					dirBytesB, tmpMD5 = bytesFromDir(dir, suffix)
				}
			case err := <-watcher.Errors:
				logger.Println("error:", err)
			}
			if !bytes.Equal(tmpMD5, currentMD5) {
				copy(currentMD5, tmpMD5)
				dirBytesCh <- dirBytesB
				updates++

			} else {
				logger.Println("ignoring no-op byte changes")
			}
		}
	}()
	return dirBytesCh, nil
}

func bytesFromDir(dir, suffix string) (*bytes.Buffer, []byte) {

	var buf *bytes.Buffer = &bytes.Buffer{}
	md5 := md5.New()

	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), suffix) {

			if b, err := ioutil.ReadFile(filepath.Join(dir, f.Name())); err != nil {
				logger.Println(err)
				logger.Printf("unable tor read file: %s, '%s', skipping", f.Name(), err)
			} else {
				buf.Write(b)
				md5.Write(b)
			}

		}
	}
	return buf, md5.Sum(nil)

}
