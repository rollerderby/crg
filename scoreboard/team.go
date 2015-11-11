package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type team struct {
	sb                     *Scoreboard
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

func newTeam(sb *Scoreboard, id uint8) *team {
	t := &team{
		sb:       sb,
		base:     fmt.Sprintf("%s.Team(%d)", sb.stateBase(), id),
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
	statemanager.RegisterCommand(t.stateIDs["timeouts"]+".Start", t.startTimeout)
	statemanager.RegisterCommand(t.stateIDs["officialReviews"]+".Start", t.startOfficialReview)
	statemanager.RegisterCommand(t.stateIDs["officialReviews"]+".Retained", t.retainOfficialReview)
	statemanager.RegisterCommand(t.base+".DeleteSkater", t.deleteSkater)

	// Setup Updaters for skaters (functions located in skater.go)
	statemanager.RegisterPatternUpdaterString(t.base+".Skater(*).ID", 0, t.sSetID)
	statemanager.RegisterPatternUpdaterString(t.base+".Skater(*).Name", 0, t.sSetName)
	statemanager.RegisterPatternUpdaterString(t.base+".Skater(*).LegalName", 0, t.sSetLegalName)
	statemanager.RegisterPatternUpdaterString(t.base+".Skater(*).InsuranceNumber", 0, t.sSetInsuranceNumber)
	statemanager.RegisterPatternUpdaterString(t.base+".Skater(*).Number", 0, t.sSetNumber)
	statemanager.RegisterPatternUpdaterBool(t.base+".Skater(*).IsAlt", 0, t.sSetIsAlt)
	statemanager.RegisterPatternUpdaterBool(t.base+".Skater(*).IsCaptain", 0, t.sSetIsCaptain)
	statemanager.RegisterPatternUpdaterBool(t.base+".Skater(*).IsAltCaptain", 0, t.sSetIsAltCaptain)
	statemanager.RegisterPatternUpdaterBool(t.base+".Skater(*).IsBenchStaff", 0, t.sSetIsBenchStaff)

	t.reset()
	return t
}

func (t *team) reset() {
	t.setName(fmt.Sprintf("Team %d", t.id))
	t.setScore(0)
	t.setLastScore(0)
	t.setTimeouts(3)
	t.setOfficialReviews(1)
	t.setOfficialReviewRetained(false)
}

func (t *team) deleteSkater(data []string) error {
	if len(data) < 1 {
		return errSkaterNotFound
	}

	if _, ok := t.skaters[data[0]]; !ok {
		return errSkaterNotFound
	}
	delete(t.skaters, data[0])
	return statemanager.StateUpdate(t.base+".Skater("+data[0]+")", nil)
}

func (t *team) stateBase() string {
	return t.base
}

func (t *team) startTimeout(_ []string) error {
	state := stateTTO1
	if t.id == 2 {
		state = stateTTO2
	}
	return t.sb.timeout([]string{state})
}

func (t *team) startOfficialReview(_ []string) error {
	state := stateOR1
	if t.id == 2 {
		state = stateOR2
	}
	return t.sb.timeout([]string{state})
}

func (t *team) retainOfficialReview(_ []string) error {
	if t.officialReviews == 0 && !t.officialReviewRetained {
		t.setOfficialReviews(1)
		t.setOfficialReviewRetained(true)
	} else if t.officialReviews == 1 && t.officialReviewRetained {
		t.setOfficialReviews(0)
		t.setOfficialReviewRetained(false)
	}
	return nil
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
