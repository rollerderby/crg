// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import "log"

// Listener allows functions to listen for changes in the state of the scoreboard
type Listener struct {
	name     string
	stateNum uint64
	ch       chan map[string]*State
	matchers []patternMatcher
	paths    []string
}

var listeners []*Listener

// NewListener creates a listener with name describing the listener (for log messages)
// and cb is a callback function which gets called on changes to the state filtered
// by Listener.RegisterPaths
func NewListener(name string, cb func(map[string]*State)) *Listener {
	lock.Lock()
	defer lock.Unlock()

	l := &Listener{
		name:     name,
		stateNum: 0,
		ch:       make(chan map[string]*State),
		matchers: nil,
		paths:    nil,
	}

	listeners = append(listeners, l)
	// Process update goroutine
	go func() {
		for {
			updates := <-l.ch
			cb(updates)
			if updates == nil {
				return
			}
		}
	}()
	return l
}

// NewStringListener creates a listener with name describing the listener (for log messages)
// and cb is a callback function which gets called on changes to the state filtered
// by Listener.RegisterPaths
func NewStringListener(name string, cb func(map[string]string)) *Listener {
	var callbackConverter = func(u1 map[string]*State) {
		u2 := make(map[string]string)
		for k, s := range u1 {
			if s.HasValue() {
				u2[k] = ""
			} else {
				u2[k] = s.Value()
			}
		}
		cb(u2)
	}

	return NewListener(name, callbackConverter)
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
		if debugFlag {
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
		if debugFlag {
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
	var u map[string]*State
	if l.stateNum < stateNum || paths != nil {
		for stateName, s := range states {
			needed := l.stateNum < s.StateNum()
			matched := false

			for idx, pm := range l.matchers {
				p := l.paths[idx]
				if pm.Matches(stateName) {
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
					u = make(map[string]*State)
				}
				u[stateName] = s
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
