package statemanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// Saver handles saving part (or all) of the state to a file
type Saver struct {
	sync.Mutex
	state       map[string]*string
	filename    string
	interval    time.Duration
	version     bool
	listener    *Listener
	saveTrigger chan bool
	lastSaved   time.Time
}

// NewSaver creates a new saver.
// filename: name of the file
// base: pattern to match (see PatternMatch for examples of matching)
// interval: time between saves.  Zero if you want/need a save on every change, will only
//           save if something has actually changed
// version: save older versions of the file (move file to file.1, file.1 to file.2, etc) NOT IMPLEMENTED!
func NewSaver(filename, base string, interval time.Duration, version bool) (*Saver, map[string]string) {
	log.Printf("Saver(%v): Opening", filename)
	savedState := loadState(filename)

	s := &Saver{
		state:       make(map[string]*string),
		filename:    filename,
		interval:    interval,
		version:     version,
		saveTrigger: make(chan bool),
	}

	s.listener = NewListener(fmt.Sprintf("Saver(%s)", filename), s.processUpdates)
	s.listener.RegisterPaths([]string{base})
	go s.saveLoop()

	return s, savedState
}

// Close unregisters the Saver from the statemanager and stops the saving go routine (issuing one last save
// in case there were changes since last save)
func (s *Saver) Close() {
	log.Printf("Saver(%v): Closing", s.filename)
	s.listener.Close()
	s.saveState()
	s.listener = nil
	s.saveTrigger <- true
	log.Printf("Saver(%v): Closed", s.filename)
}

func (s *Saver) processUpdates(updates map[string]*string) {
	s.Lock()
	defer s.Unlock()

	for key, value := range updates {
		if value == nil {
			delete(s.state, key)
		} else {
			s.state[key] = value
		}
	}

	now := time.Now()
	if s.interval == 0 || now.Sub(s.lastSaved) >= s.interval {
		s.lastSaved = now
		s.saveTrigger <- true
	}
}

// func saveInitialize() {
// 	c := make(chan map[string]*string, 10)
// 	go func() {
// 		for {
// 			updates := <-c
// 			for key, value := range updates {
// 				state[key] = value
// 			}
//
// 			saveState()
// 		}
// 	}()
// 	statemanager.RegisterListener(c)
// 	statemanager.RegisterListenerPaths(c, []string{"ScoreBoard"})
// }

func loadState(filename string) map[string]string {
	state := make(map[string]string)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(b, &state)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}

	return state
}

// TODO: implement s.interval in a better way.  only issues save if the last save was more than interval away,
// should possibly issue a save if nothing has changed in that time too.
func (s *Saver) saveLoop() {
	for {
		select {
		case <-s.saveTrigger:
			if s.listener == nil {
				// listener is nil, saver close was requsted, saveState already called from Close()
				return
			}
			s.saveState()
		}
	}
}

func (s *Saver) saveState() {
	w, err := os.Create(s.filename)
	if err != nil {
		log.Print("Cannot save state to disk.", err)
	}
	defer w.Close()

	s.Lock()
	defer s.Unlock()
	b, err := json.Marshal(s.state)
	if err != nil {
		fmt.Println("error:", err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	out.WriteTo(w)
}
