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
	"sync"
)

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

// ErrStateInvalid is returned when state value is already deleted
var ErrStateInvalid = errors.New("State Invalid")

// SetDebug turns debugging information on or off (sent to log)
func SetDebug(d bool) { debugFlag = d }

// Initialize starts up the state
func Initialize() {
	cond = sync.NewCond(&lock)
	states = make(map[string]*State)

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
