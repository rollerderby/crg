// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

// Package state stores the global state and provides
// mechanisms for updating the state, listing for state updates,
// registering commands, and triggering commands
package state

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type state struct {
	stateNum    uint64
	name        string
	isEmpty     bool
	valueString string
	valueInt64  int64
	valueBool   bool
	valueTime   time.Time
	t           string
}

var states map[string]*state
var lock sync.Mutex
var cond *sync.Cond
var debugFlag = false

var stateNum = uint64(1)
var stateUpdated = false

// ErrNotFound is returned when the key name is not in the current state
var ErrNotFound = errors.New("State Not Found")

// ErrUpdaterNotFound is returned when an updater cannot be located for the key name
var ErrUpdaterNotFound = errors.New("Updater Not Found")

// ErrUnknownType is returned when the type passed to the state is not
// one of the supported types (currently string, int64, bool)
var ErrUnknownType = errors.New("Unknown State Type")

// SetDebug turns debugging information on or off (sent to log)
func SetDebug(d bool) { debugFlag = d }

// Initialize starts up the state
func Initialize() {
	cond = sync.NewCond(&lock)
	states = make(map[string]*state)

	go flushListeners()
}

// Lock places the state in a locked state, should be called before
// any updates to the state are made, but only once.
func Lock() {
	lock.Lock()
}

// Unlock removes the lock from the state and starts the processing
// of any updates to listeners waiting for changes
func Unlock() {
	cond.Signal()
	lock.Unlock()
}

func (s *state) Value() (string, bool) {
	if s.isEmpty {
		return "", true
	}
	switch s.t {
	case "string":
		return s.valueString, false
	case "int64":
		return strconv.FormatInt(s.valueInt64, 10), false
	case "bool":
		return strconv.FormatBool(s.valueBool), false
	case "time":
		return s.valueTime.UTC().Format(time.RFC3339), false
	}

	// Unknown type, return empty
	return "", true
}

// StateUpdate sets the state for keyName to value.
// Passing nil as value will mark the state to nil
// and all states that starts with keyName + "."
func stateUpdate(k string, v interface{}) error {
	log.Printf("StateUpdate(%v): Using Old Interface", k)
	if v == nil {
		if debugFlag {
			log.Printf("StateUpdate: DELETING SUBTREE %s", k)
		}
		for key := range states {
			if key == k || strings.Index(key, k+".") == 0 {
				s := states[key]
				s.stateNum = stateNum
				s.isEmpty = true
				stateUpdated = true
			}
		}
		return nil
	}

	switch v := v.(type) {
	case string:
		return StateUpdateString(k, v)
	case int64:
		return StateUpdateInt64(k, v)
	case bool:
		return StateUpdateBool(k, v)
	case time.Time:
		return StateUpdateTime(k, v)
	default:
		log.Printf("StateUpdate: Unknown type '%T' for %v", v, k)
		return ErrUnknownType
	}
}

func StateDelete(k string) error {
	if debugFlag {
		log.Printf("StateDelete: %s", k)
	}
	for key := range states {
		if key == k || strings.Index(key, k+".") == 0 {
			s := states[key]
			s.stateNum = stateNum
			s.isEmpty = true
			stateUpdated = true
		}
	}
	return nil
}

func StateUpdateString(k string, v string) error {
	s, ok := states[k]
	if !ok {
		s = &state{name: k, isEmpty: true}
		states[k] = s
	}
	if s.isEmpty || s.t != "string" || s.valueString != v {
		s.t = "string"
		s.valueString = v
		s.isEmpty = false
		s.stateNum = stateNum
		stateUpdated = true
	}
	return nil
}

func StateUpdateInt64(k string, v int64) error {
	s, ok := states[k]
	if !ok {
		s = &state{name: k, isEmpty: true}
		states[k] = s
	}
	if s.isEmpty || s.t != "int64" || s.valueInt64 != v {
		s.t = "int64"
		s.valueInt64 = v
		s.isEmpty = false
		s.stateNum = stateNum
		stateUpdated = true
	}
	return nil
}

func StateUpdateBool(k string, v bool) error {
	s, ok := states[k]
	if !ok {
		s = &state{name: k, isEmpty: true}
		states[k] = s
	}
	if s.isEmpty || s.t != "bool" || s.valueBool != v {
		s.t = "bool"
		s.valueBool = v
		s.isEmpty = false
		s.stateNum = stateNum
		stateUpdated = true
	}
	return nil
}

func StateUpdateTime(k string, v time.Time) error {
	s, ok := states[k]
	if !ok {
		s = &state{name: k, isEmpty: true}
		states[k] = s
	}
	if s.isEmpty || s.t != "time" || s.valueTime != v {
		s.t = "time"
		s.valueTime = v
		s.isEmpty = false
		s.stateNum = stateNum
		stateUpdated = true
	}
	return nil
}
