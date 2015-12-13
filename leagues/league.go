// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package leagues

import (
	"errors"

	"github.com/rollerderby/crg/state"
	"github.com/rollerderby/crg/utils"
)

type League struct {
	id       string
	name     string
	stateIDs map[string]string
	// members []*leagueMember
	// teams   []*team
}

var leagues = make(map[string]*League)
var errLeagueNotFound = errors.New("League Not Found")

// Initialize the leagues subsystem
func Initialize() {
	state.RegisterPatternUpdaterString("Leagues.League(*).Name", 1, leagueSetName)

	state.RegisterPatternUpdaterString("Leagues.Person(*).ID", 0, personSetID)
	state.RegisterPatternUpdaterString("Leagues.Person(*).Name", 0, personSetName)
	state.RegisterPatternUpdaterString("Leagues.Person(*).LegalName", 0, personSetLegalName)
	state.RegisterPatternUpdaterString("Leagues.Person(*).InsuranceNumber", 0, personSetInsuranceNumber)
	state.RegisterPatternUpdaterString("Leagues.Person(*).Number", 0, personSetNumber)
}

func blankLeague(id string) *League {
	l := &League{
		stateIDs: make(map[string]string),
	}

	base := "Leagues.League(" + id + ")"
	l.stateIDs["id"] = base + ".ID"
	l.stateIDs["name"] = base + ".Name"

	l.SetID(id)
	l.SetName("")

	leagues[id] = l

	return l
}

func NewLeague(id, name, legalName, insuranceNumber, number string) *League {
	l := blankLeague(id)
	l.SetName(name)

	return l
}

func (l *League) ID() string { return l.id }
func (l *League) SetID(v string) error {
	l.id = v
	return state.StateUpdateString(l.stateIDs["id"], v)
}

func (l *League) Name() string { return l.name }
func (l *League) SetName(v string) error {
	l.name = v
	return state.StateUpdateString(l.stateIDs["name"], v)
}

/* Helper functions to find the League for RegisterUpdaters */
func findLeague(k string) *League {
	ids := utils.ParseIDs(k)
	id := ids[0]

	l, ok := leagues[id]
	if !ok {
		l = blankLeague(id)
		leagues[id] = l
	}
	return l
}

func leagueSetID(k, v string) error {
	if l := findLeague(k); l != nil {
		l.SetID(v)
		return nil
	}
	return errLeagueNotFound
}
func leagueSetName(k, v string) error {
	if l := findLeague(k); l != nil {
		l.SetName(v)
		return nil
	}
	return errLeagueNotFound
}
