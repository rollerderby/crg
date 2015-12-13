// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import "strconv"

type StateInt64 struct {
	s *State
	v int64
}

func newStateInt64(s *State, v int64) *StateInt64 {
	si := &StateInt64{s: s}
	si.Set(v)
	return si
}

func (si *StateInt64) Value() string {
	return strconv.FormatInt(si.v, 10)
}

func (si *StateInt64) Set(v int64) error {
	if si.s == nil {
		return ErrStateInvalid
	}
	si.v = v
	si.s.setUpdated()
	return nil
}

func (si *StateInt64) SetFromString(v string) error {
	if v, err := strconv.ParseInt(v, 10, 64); err != nil {
		si.Set(v)
		return nil
	} else {
		return err
	}
}

func (si *StateInt64) release() {
	si.s = nil
}
