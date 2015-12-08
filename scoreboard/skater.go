// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"errors"
	"strings"

	"github.com/rollerderby/crg/statemanager"
)

const (
	positionBench   = "Bench"
	positionJammer  = "Jammer"
	positionPivot   = "Pivot"
	positionBlocker = "Blocker"
)

type skater struct {
	t               *team
	id              string
	base            string
	name            string
	legalName       string
	insuranceNumber string
	number          string
	position        string
	isAlt           bool
	isCaptain       bool
	isAltCaptain    bool
	isBenchStaff    bool
	boxTrips        []*boxTrip
	curBoxTrip      *boxTrip
	stateIDs        map[string]string
}

var errSkaterNotFound = errors.New("Skater Not Found")
var errSkaterInBox = errors.New("Skater In Box")
var errPositionFull = errors.New("Position Full")
var errSkaterOnBench = errors.New("Skater On Bench")
var errSkaterNotInBox = errors.New("Skater Not In Box")

func blankSkater(t *team, id string) *skater {
	s := &skater{
		t:        t,
		base:     t.stateBase() + ".Skater(" + id + ")",
		stateIDs: make(map[string]string),
	}

	s.stateIDs["id"] = s.base + ".ID"
	s.stateIDs["name"] = s.base + ".Name"
	s.stateIDs["legalName"] = s.base + ".LegalName"
	s.stateIDs["insuranceNumber"] = s.base + ".InsuranceNumber"
	s.stateIDs["number"] = s.base + ".Number"
	s.stateIDs["isAlt"] = s.base + ".IsAlt"
	s.stateIDs["isCaptain"] = s.base + ".IsCaptain"
	s.stateIDs["isAltCaptain"] = s.base + ".IsAltCaptain"
	s.stateIDs["isBenchStaff"] = s.base + ".IsBenchStaff"
	s.stateIDs["position"] = s.base + ".Position"
	s.stateIDs["description"] = s.base + ".Description"
	s.stateIDs["shortDescription"] = s.base + ".ShortDescription"
	s.stateIDs["inBox"] = s.base + ".InBox"

	s.setID(id)
	s.setName("")
	s.setLegalName("")
	s.setInsuranceNumber("")
	s.setNumber("")
	s.setIsAlt(false)
	s.setIsCaptain(false)
	s.setIsAltCaptain(false)
	s.setIsBenchStaff(false)
	s.setPosition(positionBench)

	return s
}

func newSkater(t *team, id, name, legalName, insuranceNumber, number string, isAlt, isCaptain, isAltCaptain, isBenchStaff bool) *skater {
	s := blankSkater(t, id)
	s.setName(name)
	s.setLegalName(legalName)
	s.setInsuranceNumber(insuranceNumber)
	s.setNumber(number)
	s.setIsAlt(isAlt)
	s.setIsCaptain(isCaptain)
	s.setIsAltCaptain(isAltCaptain)
	s.setIsBenchStaff(isBenchStaff)

	return s
}

func (s *skater) setID(v string) error {
	s.id = v
	return statemanager.StateUpdateString(s.stateIDs["id"], v)
}

func (s *skater) setName(v string) error {
	s.name = v
	return statemanager.StateUpdateString(s.stateIDs["name"], v)
}

func (s *skater) setLegalName(v string) error {
	s.legalName = v
	return statemanager.StateUpdateString(s.stateIDs["legalName"], v)
}

func (s *skater) setInsuranceNumber(v string) error {
	s.insuranceNumber = v
	return statemanager.StateUpdateString(s.stateIDs["insuranceNumber"], v)
}

func (s *skater) setNumber(v string) error {
	s.number = v
	return statemanager.StateUpdateString(s.stateIDs["number"], v)
}

func (s *skater) setIsAlt(v bool) error {
	s.isAlt = v
	s.setDescription()
	return statemanager.StateUpdateBool(s.stateIDs["isAlt"], v)
}

func (s *skater) setIsCaptain(v bool) error {
	s.isCaptain = v
	s.setDescription()
	return statemanager.StateUpdateBool(s.stateIDs["isCaptain"], v)
}

func (s *skater) setIsAltCaptain(v bool) error {
	s.isAltCaptain = v
	s.setDescription()
	return statemanager.StateUpdateBool(s.stateIDs["isAltCaptain"], v)
}

func (s *skater) setIsBenchStaff(v bool) error {
	s.isBenchStaff = v
	s.setDescription()
	return statemanager.StateUpdateBool(s.stateIDs["isBenchStaff"], v)
}

func (s *skater) inBox() bool {
	return s.curBoxTrip != nil
}

func (s *skater) setInBox(v bool) error {
	if s.position == positionBench && v {
		return errSkaterOnBench
	}

	if !v {
		if s.curBoxTrip == nil {
			return errSkaterNotInBox
		}
		s.curBoxTrip.setOutJamIdx(int64(len(s.t.sb.jams) - 1))
		s.curBoxTrip = nil
	} else {
		s.curBoxTrip = newBoxTrip(s, int64(len(s.t.sb.jams)-1), false, s.t.starPass)
		s.boxTrips = append(s.boxTrips, s.curBoxTrip)
	}

	if s.position == positionJammer || s.position == positionPivot {
		s.t.updatePositions()
	}

	return statemanager.StateUpdateBool(s.stateIDs["inBox"], v)
}

func (s *skater) setPosition(v string) error {
	var set = func(v string) error {
		updatePositions := v == positionJammer || v == positionPivot || s.position == positionJammer || s.position == positionPivot
		s.position = v
		if updatePositions {
			s.t.updatePositions()
		}
		return statemanager.StateUpdateString(s.stateIDs["position"], v)
	}

	if v == s.position {
		// Nothing to see, move along
		return nil
	}

	if s.inBox() {
		return errSkaterInBox
	}

	if v == positionBench {
		return set(v)
	}
	if v == positionJammer {
		s2, ok := s.t.skaters[s.t.jammer]
		if ok {
			if err := s2.setPosition(positionBench); err != nil {
				return err
			}
		}
		return set(v)
	}
	if v == positionPivot {
		s2, ok := s.t.skaters[s.t.pivot]
		if ok {
			to := positionBench
			if s.position == positionBlocker {
				to = positionBlocker
			}
			if err := s2.setPosition(to); err != nil {
				return err
			}
		}
		return set(v)
	}

	// Must be blocker
	open := 3
	if s.t.pivot == "" || s.position == positionPivot {
		open = 4
	}
	for _, s2 := range s.t.skaters {
		if s2.position == positionBlocker {
			open = open - 1
		}
	}
	if open < 1 {
		return errPositionFull
	}
	return set(v)
}

func (s *skater) setDescription() {
	var long, short []string
	if s.isAlt {
		long = append(long, "Alternate")
		short = append(short, "Alt")
	}
	if s.isCaptain {
		long = append(long, "Captain")
		short = append(short, "C")
	}
	if s.isAltCaptain {
		long = append(long, "Alternate Captain")
		short = append(short, "A")
	}
	if s.isBenchStaff {
		long = append(long, "Bench Staff")
		short = append(short, "B")
	}
	statemanager.StateUpdateString(s.stateIDs["description"], strings.Join(long, ", "))
	statemanager.StateUpdateString(s.stateIDs["shortDescription"], strings.Join(short, ""))
}

/* Helper functions to find the skater for RegisterUpdaters */
func (t *team) findSkater(k string) *skater {
	ids := statemanager.ParseIDs(k)
	if len(ids) < 2 {
		return nil
	}
	id := ids[1]

	s, ok := t.skaters[id]
	if !ok {
		s = blankSkater(t, id)
		t.skaters[id] = s
	}
	return s
}

func (t *team) sSetID(k, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setID(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetName(k, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setName(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetLegalName(k, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setLegalName(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetInsuranceNumber(k, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setInsuranceNumber(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetNumber(k, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setNumber(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetIsAlt(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsAlt(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetIsCaptain(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsCaptain(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetIsAltCaptain(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsAltCaptain(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetIsBenchStaff(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsBenchStaff(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetPosition(k string, v string) error {
	if s := t.findSkater(k); s != nil {
		s.setPosition(v)
		return nil
	}
	return errSkaterNotFound
}
func (t *team) sSetInBox(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setInBox(v)
		return nil
	}
	return errSkaterNotFound
}
