// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

type StateString struct {
	s *State
	v string
}

func newStateString(s *State, v string) *StateString {
	si := &StateString{s: s}
	si.Set(v)
	return si
}

func (ss *StateString) Value() string {
	return ss.v
}

func (ss *StateString) Set(v string) error {
	if ss.s == nil {
		return ErrStateInvalid
	}
	ss.v = v
	ss.s.setUpdated()
	return nil
}

func (ss *StateString) SetFromString(v string) error {
	return ss.Set(v)
}

func (ss *StateString) release() {
	ss.s = nil
}
