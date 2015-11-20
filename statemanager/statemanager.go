// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

// Package statemanager stores the global state and provides
// mechanisms for updating the state, listing for state updates,
// registering commands, and triggering commands
package statemanager

import (
	"errors"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type state struct {
	stateNum  uint64
	name      string
	value     *string
	valueType string
}

var states map[string]*state
var lock sync.Mutex
var cond *sync.Cond
var debug = false

var stateNum = uint64(1)
var stateUpdated = false

// ErrNotFound is returned when the key name is not in the current state
var ErrNotFound = errors.New("State Not Found")

// ErrUpdaterNotFound is returned when an updater cannot be located for the key name
var ErrUpdaterNotFound = errors.New("Updater Not Found")

// ErrUnknownType is returned when the type passed to the statemanager is not
// one of the supported types (currently string, int64, bool)
var ErrUnknownType = errors.New("Unknown State Type")

var baseFilePath = ""

// SetDebug turns debugging information on or off (sent to log)
func SetDebug(d bool) { debug = d }

// BaseFilePath returns the base directory for loading or saving files
func BaseFilePath() string { return baseFilePath }

// SetBaseFilePath sets the base directory for loading or saving files
func SetBaseFilePath(p ...string) {
	path := filepath.Join(p...)
	log.Printf("statemanager: Setting BaseFilePath to '%v'", path)
	baseFilePath = path
}

// Initialize starts up the statemanager
func Initialize() {
	cond = sync.NewCond(&lock)
	states = make(map[string]*state)

	go flushListeners()
}

// Lock places the statemanager in a locked state, should be called before
// any updates to the state are made, but only once.
func Lock() {
	lock.Lock()
}

// Unlock removes the lock from the statemanager and starts the processing
// of any updates to listeners waiting for changes
func Unlock() {
	cond.Signal()
	lock.Unlock()
}

// StateUpdate sets the state for keyName to value.
// Passing nil as value will mark the state to nil
// and all states that starts with keyName + "."
func StateUpdate(keyName string, value interface{}) error {
	s, ok := states[keyName]
	if !ok {
		s = &state{name: keyName}
		states[keyName] = s
	}
	if debug {
		if s.value == nil {
			log.Printf("StateUpdate(%s, %v, %T) s.value=%v", keyName, value, value, s.value)
		} else {
			log.Printf("StateUpdate(%s, %v, %T) s.value='%v'", keyName, value, value, *s.value)
		}
	}

	if value == nil {
		for key := range states {
			if key == keyName || strings.Index(key, keyName+".") == 0 {
				s := states[key]
				s.stateNum = stateNum
				s.value = nil
				stateUpdated = true
			}
		}
	} else {
		var newValue string
		switch v := value.(type) {
		case string:
			newValue = v
			s.valueType = "string"
		case int64:
			newValue = strconv.FormatInt(v, 10)
			s.valueType = "int64"
		case bool:
			newValue = strconv.FormatBool(v)
			s.valueType = "bool"
		case time.Time:
			newValue = v.UTC().Format(time.RFC3339)
			s.valueType = "time"
		default:
			log.Printf("StateUpdate: Unknown type '%T' for %v", value, keyName)
			return ErrUnknownType
		}
		if s.value == nil || *s.value != newValue {
			if debug {
				log.Printf("StateUpdate: setting %s to %v", keyName, value)
			}
			s.stateNum = stateNum
			s.value = &newValue
			stateUpdated = true
		}
	}
	return nil
}

// ParseIDs returns all values within () in the input string.
// Example
// Scoreboard.Team(1).Skater(abc123).Name returns ["1", "abc123"]
func ParseIDs(k string) []string {
	var ret []string
	startPos := -1
	for idx, c := range k {
		if startPos == -1 && c == '(' {
			startPos = idx + 1
		} else if startPos != -1 && c == ')' {
			ret = append(ret, k[startPos:idx])
			startPos = -1
		}
	}
	return ret
}
