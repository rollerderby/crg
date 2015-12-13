// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import (
	"log"
	"strings"
	"time"
)

type StateValue interface {
	Value() string
	SetFromString(string) error
	release()
}

type State struct {
	stateNum uint64
	name     string
	strVal   string
	strValSN uint64
	value    StateValue
}

var states map[string]*State

func Delete(k string) error {
	if debugFlag {
		log.Printf("StateDelete: %s", k)
	}
	for key := range states {
		if key == k || strings.Index(key, k+".") == 0 {
			states[key].Delete()
		}
	}
	return nil
}

func (s *State) Delete() {
	s.stateNum = stateNum
	s.value = nil
	stateUpdated = true
}

func (s *State) HasValue() bool {
	return s.value != nil
}

func (s *State) Value() string {
	if s.value == nil {
		return ""
	}

	if s.strValSN != s.stateNum {
		s.strVal = s.value.Value()
	}
	return s.strVal
}

func (s *State) StateNum() uint64 {
	return s.stateNum
}

func (s *State) Set(v string) error {
	if s.value == nil {
		s.value = &StateString{s: s}
	}
	return s.value.SetFromString(v)
}

func (s *State) SetString(v string) {
	sv, ok := s.value.(*StateString)
	if !ok {
		if s.value != nil {
			s.value.release()
		}
		s.value = newStateString(s, v)
	} else {
		sv.Set(v)
	}
}

func (s *State) SetBool(v bool) {
	sv, ok := s.value.(*StateBool)
	if !ok {
		if s.value != nil {
			s.value.release()
		}
		s.value = newStateBool(s, v)
	} else {
		sv.Set(v)
	}
}

func (s *State) SetInt64(v int64) {
	sv, ok := s.value.(*StateInt64)
	if !ok {
		if s.value != nil {
			s.value.release()
		}
		s.value = newStateInt64(s, v)
	} else {
		sv.Set(v)
	}
}

func (s *State) SetTime(v time.Time) {
	sv, ok := s.value.(*StateTime)
	if !ok {
		if s.value != nil {
			s.value.release()
		}
		s.value = newStateTime(s, v)
	} else {
		sv.Set(v)
	}
}

func (s *State) setUpdated() {
	s.stateNum = stateNum
	stateUpdated = true
}

func GetState(k string) *State {
	s, ok := states[k]
	if !ok {
		s = &State{name: k}
		states[k] = s
	}
	return s
}

func SetStateString(k string, v string) *State {
	s := GetState(k)
	s.SetString(v)
	return s
}

func SetStateInt64(k string, v int64) *State {
	s := GetState(k)
	s.SetInt64(v)
	return s
}

func SetStateBool(k string, v bool) *State {
	s := GetState(k)
	s.SetBool(v)
	return s
}

func SetStateTime(k string, v time.Time) *State {
	s := GetState(k)
	s.SetTime(v)
	return s
}
