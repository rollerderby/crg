// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import "time"

type StateTime struct {
	s *State
	v time.Time
}

func newStateTime(s *State, v time.Time) *StateTime {
	si := &StateTime{s: s}
	si.Set(v)
	return si
}

func (st *StateTime) Value() string {
	return st.v.UTC().Format(time.RFC3339)
}

func (st *StateTime) Set(v time.Time) error {
	if st.s == nil {
		return ErrStateInvalid
	}
	st.v = v
	st.s.setUpdated()
	return nil
}

func (st *StateTime) SetFromString(v string) error {
	if v, err := time.Parse(time.RFC3339, v); err != nil {
		st.Set(v)
		return nil
	} else {
		return err
	}
}

func (st *StateTime) release() {
	st.s = nil
}
