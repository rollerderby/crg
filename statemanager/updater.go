package statemanager

import (
	"log"
	"sort"
	"strconv"
	"time"
)

// UpdaterFunc is a callback type for string updates from the state engine
type UpdaterFunc func(string) error

// UpdaterInt64Func is a callback type for int64 updates from the state engine
type UpdaterInt64Func func(int64) error

// UpdaterBoolFunc is a callback type for bool updates from the state engine
type UpdaterBoolFunc func(bool) error

// UpdaterTimeFunc is a callback type for time updates from the state engine
type UpdaterTimeFunc func(time.Time) error

type stateUpdater struct {
	stringUpdater UpdaterFunc
	int64Updater  UpdaterInt64Func
	boolUpdater   UpdaterBoolFunc
	timeUpdater   UpdaterTimeFunc
	name          string
	groupPriority uint8
}
type stateUpdaterArray []*stateUpdater

var updaters = make(map[string]*stateUpdater)

// StateSet attempts to update the state to value
// using keyName to lookup a handler.  It returns an
// error on failure.
func StateSet(keyName string, value string) error {
	state, ok := states[keyName]
	if !ok {
		return ErrNotFound
	}
	if state.value != nil && *state.value == value {
		// nothing to do, move along
		return nil
	}

	updater, ok := updaters[keyName]
	if !ok {
		return ErrUpdaterNotFound
	}

	switch state.valueType {
	case "string":
		if debug {
			log.Printf("Calling stringUpdater(%v) for %s (%s)", value, keyName, *state.value)
		}
		return updater.stringUpdater(value)
	case "int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		if debug {
			log.Printf("Calling int64Updater(%v) for %s (%s)", v, keyName, *state.value)
		}
		return updater.int64Updater(v)
	case "bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		if debug {
			log.Printf("Calling boolUpdater(%v) for %s (%s)", v, keyName, *state.value)
		}
		return updater.boolUpdater(v)
	case "time":
		v, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		if debug {
			log.Printf("Calling timeUpdater(%v) for %s (%s)", v, keyName, *state.value)
		}
		return updater.timeUpdater(v)
	default:
		log.Printf("StateSet: Unknown type '%v' for %v", state.valueType, keyName)
		return ErrUnknownType
	}
	return ErrNotFound
}

// StateSetGroup attempts to update the state using a
// map of key/values to lookup handlers.  Calls
// StateSet for each key/values using the groupPriority
// of the registered updater to call lower numbers (higher
// priority) first.  Allows setting things like min/max values
// before the actual number.
func StateSetGroup(values map[string]string) {
	var u []*stateUpdater

	for keyName := range values {
		updater, ok := updaters[keyName]
		if ok {
			u = append(u, updater)
		}
	}

	sort.Sort(stateUpdaterArray(u))
	for _, updater := range u {
		keyName := updater.name
		err := StateSet(keyName, values[keyName])
		if err != nil {
			log.Print("StateSetGroup: Cannot set state: ", err)
		}
	}
}

// RegisterUpdater adds a string updater to the statemanager.
func RegisterUpdater(name string, groupPriority uint8, u UpdaterFunc) {
	updaters[name] = &stateUpdater{stringUpdater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterInt64 adds an int64 updater to the statemanager.
func RegisterUpdaterInt64(name string, groupPriority uint8, u UpdaterInt64Func) {
	updaters[name] = &stateUpdater{int64Updater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterBool adds a bool updater to the statemanager.
func RegisterUpdaterBool(name string, groupPriority uint8, u UpdaterBoolFunc) {
	updaters[name] = &stateUpdater{boolUpdater: u, name: name, groupPriority: groupPriority}
}

// RegisterUpdaterTime adds a time updater to the statemanager.
func RegisterUpdaterTime(name string, groupPriority uint8, u UpdaterTimeFunc) {
	updaters[name] = &stateUpdater{timeUpdater: u, name: name, groupPriority: groupPriority}
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
