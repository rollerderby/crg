package leagues

import (
	"errors"

	"github.com/rollerderby/crg/statemanager"
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
	statemanager.RegisterPatternUpdaterString("Leagues.League(*).Name", 1, leagueSetName)

	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).ID", 0, personSetID)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).Name", 0, personSetName)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).LegalName", 0, personSetLegalName)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).InsuranceNumber", 0, personSetInsuranceNumber)
	statemanager.RegisterPatternUpdaterString("Leagues.Person(*).Number", 0, personSetNumber)
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
	return statemanager.StateUpdate(l.stateIDs["id"], v)
}

func (l *League) Name() string { return l.name }
func (l *League) SetName(v string) error {
	l.name = v
	return statemanager.StateUpdate(l.stateIDs["name"], v)
}

/* Helper functions to find the League for RegisterUpdaters */
func findLeague(k string) *League {
	ids := statemanager.ParseIDs(k)
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
