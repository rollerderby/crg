// Package statemanager stores the global state and provides
// mechanisms for updating the state, listing for state updates,
// registering commands, and triggering commands
package statemanager

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
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
var ErrNotFound = errors.New("State Not Found")
var ErrUpdaterNotFound = errors.New("Updater Not Found")
var ErrUnknownType = errors.New("Unknown State Type")

func Initialize() {
	cond = sync.NewCond(&lock)
	states = make(map[string]*state)

	go flushListeners()
}

func Lock() {
	lock.Lock()
}

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
		s.stateNum = stateNum
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
		default:
			return ErrUnknownType
		}
		if s.value == nil || *s.value != newValue {
			if debug {
				log.Printf("StateUpdate: setting %s to %v", keyName, value)
			}
			s.value = &newValue
			stateUpdated = true
		}
	}
	return nil
}
