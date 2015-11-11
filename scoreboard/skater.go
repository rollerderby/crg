package scoreboard

import (
	"errors"
	"strings"

	"github.com/rollerderby/crg/statemanager"
)

type skater struct {
	t               *team
	id              string
	name            string
	legalName       string
	insuranceNumber string
	number          string
	isCaptain       bool
	isAlt           bool
	isBenchStaff    bool
	stateIDs        map[string]string
}

var errSkaterNotFound = errors.New("Skater Not Found")

func blankSkater(t *team, id string) *skater {
	s := &skater{
		t:        t,
		stateIDs: make(map[string]string),
	}

	base := t.stateBase() + ".Skater(" + id + ")"
	s.stateIDs["id"] = base + ".ID"
	s.stateIDs["name"] = base + ".Name"
	s.stateIDs["legalName"] = base + ".LegalName"
	s.stateIDs["insuranceNumber"] = base + ".InsuranceNumber"
	s.stateIDs["number"] = base + ".Number"
	s.stateIDs["isCaptain"] = base + ".IsCaptain"
	s.stateIDs["isAlt"] = base + ".IsAlt"
	s.stateIDs["isBenchStaff"] = base + ".IsBenchStaff"

	s.setID(id)
	s.setName("")
	s.setLegalName("")
	s.setInsuranceNumber("")
	s.setNumber("")
	s.setIsCaptain(false)
	s.setIsAlt(false)
	s.setIsBenchStaff(false)

	return s
}

func newSkater(t *team, id, name, legalName, insuranceNumber, number string, isCaptain, isAlt, isBenchStaff bool) *skater {
	s := blankSkater(t, id)
	s.setName(name)
	s.setLegalName(legalName)
	s.setInsuranceNumber(insuranceNumber)
	s.setNumber(number)
	s.setIsCaptain(isCaptain)
	s.setIsAlt(isAlt)
	s.setIsBenchStaff(isBenchStaff)

	return s
}

func (s *skater) setID(v string) error {
	s.id = v
	return statemanager.StateUpdate(s.stateIDs["id"], v)
}

func (s *skater) setName(v string) error {
	s.name = v
	return statemanager.StateUpdate(s.stateIDs["name"], v)
}

func (s *skater) setLegalName(v string) error {
	s.legalName = v
	return statemanager.StateUpdate(s.stateIDs["legalName"], v)
}

func (s *skater) setInsuranceNumber(v string) error {
	s.insuranceNumber = v
	return statemanager.StateUpdate(s.stateIDs["insuranceNumber"], v)
}

func (s *skater) setNumber(v string) error {
	s.number = v
	return statemanager.StateUpdate(s.stateIDs["number"], v)
}

func (s *skater) setIsCaptain(v bool) error {
	s.isCaptain = v
	return statemanager.StateUpdate(s.stateIDs["isCaptain"], v)
}

func (s *skater) setIsAlt(v bool) error {
	s.isAlt = v
	return statemanager.StateUpdate(s.stateIDs["isAlt"], v)
}

func (s *skater) setIsBenchStaff(v bool) error {
	s.isBenchStaff = v
	return statemanager.StateUpdate(s.stateIDs["isBenchStaff"], v)
}

/* Helper functions to find the skater for RegisterUpdaters */
func (t *team) findSkater(k string) *skater {
	k = k[len(t.base+".Skater("):]
	end := strings.IndexRune(k, ')')
	if end <= 0 {
		return nil
	}
	id := k[:end]

	s, ok := t.skaters[id]
	if !ok {
		s = blankSkater(t, id)
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
func (t *team) sSetIsCaptain(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsCaptain(v)
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
func (t *team) sSetIsBenchStaff(k string, v bool) error {
	if s := t.findSkater(k); s != nil {
		s.setIsBenchStaff(v)
		return nil
	}
	return errSkaterNotFound
}
