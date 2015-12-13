// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package state

import "strconv"

type StateBool struct {
	s *State
	v bool
}

func newStateBool(s *State, v bool) *StateBool {
	si := &StateBool{s: s}
	si.Set(v)
	return si
}

func (sb *StateBool) Value() string {
	return strconv.FormatBool(sb.v)
}

func (sb *StateBool) Set(v bool) error {
	if sb.s == nil {
		return ErrStateInvalid
	}
	sb.v = v
	sb.s.setUpdated()
	return nil
}

func (sb *StateBool) SetFromString(v string) error {
	if v, err := strconv.ParseBool(v); err != nil {
		sb.Set(v)
		return nil
	} else {
		return err
	}
}

func (sb *StateBool) release() {
	sb.s = nil
}
