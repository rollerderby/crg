package statemanager

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
	groupPriority uint8
	isPattern     bool
}
type stateUpdaterArray []*stateUpdater

var updaters = make(map[string]*stateUpdater)

func findUpdater(keyName string) *stateUpdater {
	updater, ok := updaters[keyName]
	if !ok {
		// not found, look for pattern match
		for _, u := range updaters {
			if u.isPattern {
				result := PatternMatch(keyName, u.name)
				if result {
					updater = u
					break
				}
			}
		}
	}

	return updater
}

// StateSet attempts to update the state to value
// using keyName to lookup a handler.  It returns an
// error on failure.
func StateSet(keyName string, value string) error {
	updater := findUpdater(keyName)
	if updater == nil {
		return ErrUpdaterNotFound
	}

	state, ok := states[keyName]
	if ok && state.value != nil && *state.value == value {
		// nothing to do, move along
		return nil
	}

	switch cb := updater.updater.(type) {
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
	log.Printf("StateSet: Unknown type '%T' for %v", updater.updater, keyName)
	return ErrUnknownType
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
		updater := findUpdater(keyName)
		if updater != nil {
			u = append(u, updater)
			um[updater] = append(um[updater], keyName)
		}
	}

	sort.Sort(stateUpdaterArray(u))
	for _, updater := range u {
		for _, keyName := range um[updater] {
			err := StateSet(keyName, values[keyName])
			if err != nil {
				log.Print("StateSetGroup: Cannot set state: ", err)
			}
		}
	}
}

// RegisterUpdaterString adds a string updater to the statemanager.
func RegisterUpdaterString(name string, groupPriority uint8, u UpdaterStringFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterInt64 adds an int64 updater to the statemanager.
func RegisterUpdaterInt64(name string, groupPriority uint8, u UpdaterInt64Func) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterBool adds a bool updater to the statemanager.
func RegisterUpdaterBool(name string, groupPriority uint8, u UpdaterBoolFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterTime adds a time updater to the statemanager.
func RegisterUpdaterTime(name string, groupPriority uint8, u UpdaterTimeFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority}
}

// RegisterPatternUpdaterString adds a string updater to the statemanager.
func RegisterPatternUpdaterString(name string, groupPriority uint8, u UpdaterPatternStringFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true}
}

// RegisterPatternUpdaterInt64 adds an int64 updater to the statemanager.
func RegisterPatternUpdaterInt64(name string, groupPriority uint8, u UpdaterPatternInt64Func) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true}
}

// RegisterPatternUpdaterBool adds a bool updater to the statemanager.
func RegisterPatternUpdaterBool(name string, groupPriority uint8, u UpdaterPatternBoolFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true}
}

// RegisterPatternUpdaterTime adds a time updater to the statemanager.
func RegisterPatternUpdaterTime(name string, groupPriority uint8, u UpdaterPatternTimeFunc) {
	updaters[name] = &stateUpdater{updater: u, name: name, groupPriority: groupPriority, isPattern: true}
}

// UnregisterUpdater removes an updater from the statemanager.
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
