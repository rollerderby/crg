package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type team struct {
	parent    parent
	base      string
	id        uint8
	name      string
	score     int64
	lastScore int64
	settings  map[string]*setting
	skaters   map[string]*skater
	stateIDs  map[string]string
}

func newTeam(p parent, id uint8) *team {
	t := &team{
		parent:   p,
		base:     fmt.Sprintf("%s.Team(%d)", p.stateBase(), id),
		id:       id,
		settings: make(map[string]*setting),
		skaters:  make(map[string]*skater),
		stateIDs: make(map[string]string),
	}

	t.stateIDs["id"] = fmt.Sprintf("%s.ID", t.base)
	t.stateIDs["name"] = fmt.Sprintf("%s.Name", t.base)
	t.stateIDs["score"] = fmt.Sprintf("%s.Score", t.base)
	t.stateIDs["lastScore"] = fmt.Sprintf("%s.LastScore", t.base)

	statemanager.StateUpdate(t.stateIDs["id"], int64(id))

	statemanager.RegisterUpdater(t.stateIDs["name"], 0, t.setName)
	statemanager.RegisterUpdaterInt64(t.stateIDs["score"], 0, t.setScore)
	statemanager.RegisterUpdaterInt64(t.stateIDs["lastScore"], 0, t.setLastScore)

	t.setName(fmt.Sprintf("Team %d", id))
	t.setScore(0)
	t.setLastScore(0)

	return t
}

func (t *team) setName(name string) error {
	t.name = name
	statemanager.StateUpdate(t.stateIDs["name"], name)
	return nil
}

func (t *team) setScore(v int64) error {
	t.score = v
	statemanager.StateUpdate(t.stateIDs["score"], v)
	return nil
}

func (t *team) setLastScore(v int64) error {
	t.lastScore = v
	statemanager.StateUpdate(t.stateIDs["lastScore"], v)
	return nil
}
