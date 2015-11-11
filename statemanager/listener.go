package statemanager

import (
	"fmt"
	"log"
	"sort"
)

// Listener allows functions to listen for changes in the state of the scoreboard
type Listener struct {
	name     string
	callback func(map[string]*string)
	stateNum uint64
	ch       chan map[string]*string
	matchers []patternMatcher
	paths    []string
}

var listeners []*Listener

// NewListener creates a listener with name describing the listener (for log messages)
// and cb is a callback function which gets called on changes to the state filtered
// by Listener.RegisterPaths
func NewListener(name string, cb func(map[string]*string)) *Listener {
	lock.Lock()
	defer lock.Unlock()

	l := &Listener{
		name:     name,
		stateNum: 0,
		callback: cb,
		ch:       make(chan map[string]*string, 10),
		matchers: nil,
		paths:    nil,
	}

	listeners = append(listeners, l)
	go l.processUpdates()
	return l
}

// Close closes the listener.  After this call it the callback will never be called for
// this listener again.
func (l *Listener) Close() {
	lock.Lock()
	defer lock.Unlock()

	lLen := len(listeners)
	for i, l2 := range listeners {
		if l == l2 {
			listeners[i], listeners[lLen-1] = listeners[lLen-1], nil
			listeners = listeners[:lLen-1]
			return
		}
	}
}

func (l *Listener) processUpdates() {
	for {
		updates := <-l.ch
		if debug {
			var values []string
			for k, v := range updates {
				if v != nil {
					values = append(values, fmt.Sprintf("%v=\"%v\"", k, *v))
				} else {
					values = append(values, fmt.Sprintf("%v=nil", k))
				}
			}
			sort.StringSlice(values).Sort()
			log.Printf("Processing %v updates for %v  %+v", len(updates), l.name, values)
		}

		l.callback(updates)
		if updates == nil {
			return
		}
	}
}

func (l *Listener) findPatternMatcher(path string) (int, patternMatcher) {
	for idx, pm := range l.matchers {
		if pm.Pattern() == path {
			return idx, pm
		}
	}
	return -1, nil
}

// RegisterPaths adds the paths to the listener to get updates.  See PatternMatch for examples
// of how the pattern matching is done.  The callback will immediately be called with and
// matching paths before returning to the caller.
func (l *Listener) RegisterPaths(paths []string) {
	if paths == nil || len(paths) == 0 {
		return
	}

	lock.Lock()
	defer lock.Unlock()
	for _, p := range paths {
		if debug {
			log.Printf("RegisterListenerPaths: %v", p)
		}
		idx, _ := l.findPatternMatcher(p)
		if idx == -1 {
			l.paths = append(l.paths, p)
			l.matchers = append(l.matchers, newPatternMatcher(p))
		}
	}
	l.flush(paths)
}

// UnregisterPaths removes the paths from the listener.
func (l *Listener) UnregisterPaths(paths []string) {
	if paths == nil || len(paths) == 0 {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	for _, p := range paths {
		if debug {
			log.Printf("UnregisterListenerPaths: %v", p)
		}
		idx, _ := l.findPatternMatcher(p)
		if idx != -1 {
			// TODO: DELETE ENTRY FROM SLICE!
			l.paths[idx] = ""
			l.matchers[idx] = nil
		}
	}
}

func (l *Listener) flush(paths []string) {
	var u map[string]*string
	if l.stateNum < stateNum || paths != nil {
		for _, s := range states {
			needed := l.stateNum < s.stateNum
			matched := false

			for idx, pm := range l.matchers {
				p := l.paths[idx]
				if pm.Matches(s.name) {
					matched = true

					// Matched, now look for just registered
					for _, p2 := range paths {
						if p == p2 {
							needed = true
							break
						}
					}
					break
				}
			}

			if matched && needed {
				if u == nil {
					u = make(map[string]*string)
				}
				u[s.name] = s.value
			}
		}
		if u != nil {
			if paths == nil {
				l.stateNum = stateNum
			}
			l.ch <- u
		}
	}
}

func flushListeners() {
	lock.Lock()
	defer lock.Unlock()

	for {
		cond.Wait()
		if stateUpdated {
			for _, l := range listeners {
				l.flush(nil)
			}
			stateNum = stateNum + 1
			stateUpdated = false
		}
	}
}
