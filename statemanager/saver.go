package statemanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

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

func NewSaver(filename, base string, interval time.Duration, version bool) (*Saver, map[string]string) {
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

func (s *Saver) Close() {
	s.listener.Close()
	s.listener = nil
	s.saveTrigger <- true
}

func (s *Saver) processUpdates(updates map[string]*string) {
	s.Lock()
	defer s.Unlock()

	for key, value := range updates {
		s.state[key] = value
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

func (s *Saver) saveLoop() {
	for {
		select {
		case <-s.saveTrigger:
			s.saveState()
		}
		if s.listener == nil {
			return
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
	w.Write(b)
}
