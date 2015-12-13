// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import (
	"log"
	"sort"
	"strconv"
	"time"
)

// UpdaterStringFunc is a callback type for string updates from the state engine
type UpdaterStringFunc func(string) error

// UpdaterInt64Func is a callback type for int64 updates from the state engine
type UpdaterInt64Func func(int64) error

// UpdaterBoolFunc is a callback type for bool updates from the state engine
type UpdaterBoolFunc func(bool) error

// UpdaterTimeFunc is a callback type for time updates from the state engine
type UpdaterTimeFunc func(time.Time) error

// UpdaterPatternStringFunc is a callback type for string updates from the state engine
type UpdaterPatternStringFunc func(string, string) error

// UpdaterPatternInt64Func is a callback type for int64 updates from the state engine
type UpdaterPatternInt64Func func(string, int64) error

// UpdaterPatternBoolFunc is a callback type for bool updates from the state engine
type UpdaterPatternBoolFunc func(string, bool) error

// UpdaterPatternTimeFunc is a callback type for time updates from the state engine
type UpdaterPatternTimeFunc func(string, time.Time) error

type stateUpdater struct {
	updater       interface{}
	name          string
	pm            patternMatcher
	groupPriority uint8
	isPattern     bool
}
type stateUpdaterArray []*stateUpdater

var updaters = make(map[string]*stateUpdater)

func findStateUpdater(keyName string) *stateUpdater {
	updater, ok := updaters[keyName]
	if !ok {
		// not found, look for pattern match
		for _, u := range updaters {
			if u.isPattern {
				result := u.pm.Matches(keyName)
				if result {
					updater = u
					break
				}
			}
		}
	}

	return updater
}

func (su *stateUpdater) update(keyName, value string) error {
	// state, ok := states[keyName]
	// if ok && state.value != nil && *state.value == value {
	// 	// nothing to do, move along
	// 	return nil
	// }

	switch cb := su.updater.(type) {
	case UpdaterStringFunc:
		return cb(value)
	case UpdaterPatternStringFunc:
		return cb(keyName, value)
	case UpdaterInt64Func:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return cb(v)
	case UpdaterPatternInt64Func:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return cb(keyName, v)
	case UpdaterBoolFunc:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		return cb(v)
	case UpdaterPatternBoolFunc:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		return cb(keyName, v)
	case UpdaterTimeFunc:
		v, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		return cb(v)
	case UpdaterPatternTimeFunc:
		v, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		return cb(keyName, v)
	}
	log.Printf("StateSet: Unknown type '%T' for %v", su.updater, keyName)
	return ErrUnknownType
}

// StateSet attempts to update the state to value
// using keyName to lookup a handler.  It returns an
// error on failure.
func StateSet(keyName string, value string) error {
	su := findStateUpdater(keyName)
	if su == nil {
		return ErrUpdaterNotFound
	}

	return su.update(keyName, value)
}

// StateSetGroup attempts to update the state using a
// map of key/values to lookup handlers.  Calls
// StateSet for each key/values using the groupPriority
// of the registered updater to call lower numbers (higher
// priority) first.  Allows setting things like min/max values
// before the actual number.
func StateSetGroup(values map[string]string) {
	var u []*stateUpdater
	um := make(map[*stateUpdater][]string)

	for keyName := range values {
		su := findStateUpdater(keyName)
		if su != nil {
			u = append(u, su)
			um[su] = append(um[su], keyName)
		}
	}

	sort.Sort(stateUpdaterArray(u))
	for _, su := range u {
		for _, keyName := range um[su] {
			err := su.update(keyName, values[keyName])
			if err != nil {
				log.Print("StateSetGroup: Cannot set state: ", err)
			}
		}
	}
}

// RegisterUpdaterString adds a string updater to the state.
func RegisterUpdaterString(name string, groupPriority uint8, u UpdaterStringFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, pm: newPatternMatcher(name)}
}

// RegisterUpdaterInt64 adds an int64 updater to the state.
func RegisterUpdaterInt64(name string, groupPriority uint8, u UpdaterInt64Func) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, pm: newPatternMatcher(name)}
}

// RegisterUpdaterBool adds a bool updater to the state.
func RegisterUpdaterBool(name string, groupPriority uint8, u UpdaterBoolFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, pm: newPatternMatcher(name)}
}

// RegisterUpdaterTime adds a time updater to the state.
func RegisterUpdaterTime(name string, groupPriority uint8, u UpdaterTimeFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, pm: newPatternMatcher(name)}
}

// RegisterPatternUpdaterString adds a string updater to the state.
func RegisterPatternUpdaterString(name string, groupPriority uint8, u UpdaterPatternStringFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true, pm: newPatternMatcher(name)}
}

// RegisterPatternUpdaterInt64 adds an int64 updater to the state.
func RegisterPatternUpdaterInt64(name string, groupPriority uint8, u UpdaterPatternInt64Func) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true, pm: newPatternMatcher(name)}
}

// RegisterPatternUpdaterBool adds a bool updater to the state.
func RegisterPatternUpdaterBool(name string, groupPriority uint8, u UpdaterPatternBoolFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true, pm: newPatternMatcher(name)}
}

// RegisterPatternUpdaterTime adds a time updater to the state.
func RegisterPatternUpdaterTime(name string, groupPriority uint8, u UpdaterPatternTimeFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true, pm: newPatternMatcher(name)}
}

// UnregisterUpdater removes an updater from the state.
func UnregisterUpdater(name string) {
	delete(updaters, name)
}

func (a stateUpdaterArray) Len() int      { return len(a) }
func (a stateUpdaterArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a stateUpdaterArray) Less(i, j int) bool {
	if a[i].groupPriority < a[j].groupPriority {
		return true
	} else if a[i].groupPriority > a[j].groupPriority {
		return false
	}
	return a[i].name < a[j].name
}
