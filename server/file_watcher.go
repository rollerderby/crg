// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-fsnotify/fsnotify"
	"github.com/rollerderby/crg/statemanager"
)

func addFileWatcher(mediaType, prefix, path string) (*fsnotify.Watcher, error) {
	fullpath := filepath.Join(statemanager.BaseFilePath(), prefix, path)
	os.MkdirAll(fullpath, 0775)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = watcher.Add(fullpath)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	f, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	for _, name := range names {
		short := filepath.Base(name)
		full := filepath.Join(path, short)

		statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, full), short)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				short := filepath.Base(event.Name)
				full := filepath.Join(path, short)

				if event.Op&fsnotify.Create == fsnotify.Create {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, full), short)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, full), nil)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					statemanager.StateUpdate(fmt.Sprintf("Media.Type(%v).File(%v)", mediaType, full), nil)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	return watcher, nil
}
