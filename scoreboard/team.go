package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type team struct {
	parent                 parent
	base                   string
	id                     uint8
	name                   string
	score                  int64
	lastScore              int64
	timeouts               int64
	officialReviews        int64
	officialReviewRetained bool
	settings               map[string]*setting
	skaters                map[string]*skater
	stateIDs               map[string]string
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
	t.stateIDs["jamScore"] = fmt.Sprintf("%s.JamScore", t.base)
	t.stateIDs["timeouts"] = fmt.Sprintf("%s.Timeouts", t.base)
	t.stateIDs["officialReviews"] = fmt.Sprintf("%s.OfficialReviews", t.base)
	t.stateIDs["officialReviewRetained"] = fmt.Sprintf("%s.OfficialReviewRetained", t.base)

	statemanager.StateUpdate(t.stateIDs["id"], int64(id))

	statemanager.RegisterUpdaterString(t.stateIDs["name"], 0, t.setName)
	statemanager.RegisterUpdaterInt64(t.stateIDs["score"], 0, t.setScore)
	statemanager.RegisterUpdaterInt64(t.stateIDs["lastScore"], 0, t.setLastScore)
	statemanager.RegisterUpdaterInt64(t.stateIDs["timeouts"], 0, t.setTimeouts)
	statemanager.RegisterUpdaterInt64(t.stateIDs["officialReviews"], 0, t.setOfficialReviews)
	statemanager.RegisterUpdaterBool(t.stateIDs["officialReviewRetained"], 0, t.setOfficialReviewRetained)

	statemanager.RegisterCommand(t.stateIDs["score"]+".Inc", t.incScore)
	statemanager.RegisterCommand(t.stateIDs["score"]+".Dec", t.decScore)
	statemanager.RegisterCommand(t.stateIDs["lastScore"]+".Inc", t.incLastScore)
	statemanager.RegisterCommand(t.stateIDs["lastScore"]+".Dec", t.decLastScore)

	t.setName(fmt.Sprintf("Team %d", id))
	t.setScore(0)
	t.setLastScore(0)
	t.setTimeouts(3)
	t.setOfficialReviews(1)
	t.setOfficialReviewRetained(false)

	return t
}

func (t *team) setName(name string) error {
	t.name = name
	statemanager.StateUpdate(t.stateIDs["name"], name)
	return nil
}

func (t *team) setScore(v int64) error {
	if v < 0 {
		return nil
	}
	t.score = v
	if v < t.lastScore {
		t.setLastScore(v)
	}
	statemanager.StateUpdate(t.stateIDs["score"], v)
	statemanager.StateUpdate(t.stateIDs["jamScore"], t.score-t.lastScore)
	return nil
}

func (t *team) setLastScore(v int64) error {
	if v < 0 {
		return nil
	}
	if v > t.score {
		return nil
	}
	t.lastScore = v
	statemanager.StateUpdate(t.stateIDs["lastScore"], v)
	statemanager.StateUpdate(t.stateIDs["jamScore"], t.score-t.lastScore)
	return nil
}

func (t *team) setTimeouts(v int64) error {
	t.timeouts = v
	statemanager.StateUpdate(t.stateIDs["timeouts"], v)
	return nil
}

func (t *team) setOfficialReviews(v int64) error {
	t.officialReviews = v
	statemanager.StateUpdate(t.stateIDs["officialReviews"], v)
	return nil
}

func (t *team) setOfficialReviewRetained(v bool) error {
	t.officialReviewRetained = v
	statemanager.StateUpdate(t.stateIDs["officialReviewRetained"], v)
	return nil
}

func (t *team) useTimeout() bool {
	if t.timeouts > 0 {
		t.setTimeouts(t.timeouts - 1)
		return true
	}
	return false
}

func (t *team) useOfficialReview() bool {
	if t.officialReviews > 0 {
		t.setOfficialReviews(t.officialReviews - 1)
		return true
	}
	return false
}

func (t *team) incScore(_ []string) error {
	t.setScore(t.score + 1)
	return nil
}

func (t *team) decScore(_ []string) error {
	t.setScore(t.score - 1)
	return nil
}

func (t *team) incLastScore(_ []string) error {
	t.setLastScore(t.lastScore + 1)
	return nil
}

func (t *team) decLastScore(_ []string) error {
	t.setLastScore(t.lastScore - 1)
	return nil
}
