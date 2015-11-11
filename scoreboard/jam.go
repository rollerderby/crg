package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

const (
	maxPasses  = 9
	posJammer  = "Jammer"
	posPivot   = "Pivot"
	posBlocker = "Blocker"
)

type jamInfo struct {
	sb     *Scoreboard
	period int64
	jam    int64
	team   []*jamTeam
}

type jamTeam struct {
	base     string
	starPass bool
	scores   [maxPasses]*passScore
	skaters  [5]*skaterPosition
}

type skaterPosition struct {
	s    *skater
	p    string
	slot int64
}

type passScore struct {
	base          string
	score         int64
	afterStarPass bool
}

func newJam(sb *Scoreboard, period, jam int64) *jamInfo {
	j := &jamInfo{
		sb:     sb,
		period: period,
		jam:    jam,
		team:   make([]*jamTeam, 2),
	}

	base := fmt.Sprintf("%v.Period(%v).Jam(%v)", sb.stateBase(), period, jam)
	j.team[0] = newJamTeam(sb, base+".Team(1)")
	j.team[1] = newJamTeam(sb, base+".Team(2)")

	return j
}

func newJamTeam(sb *Scoreboard, base string) *jamTeam {
	jt := &jamTeam{base: base}

	jt.setStarPass(false)
	for i := 0; i < maxPasses; i++ {
		jt.scores[i] = &passScore{}
		jt.scores[i].setScore(-1)
		jt.scores[i].setAfterStarPass(false)
	}

	return jt
}

func (jt *jamTeam) setStarPass(v bool) {
	jt.starPass = v
	statemanager.StateUpdate(jt.base+".StarPass", v)
}

func (ps *passScore) setScore(v int64) {
	ps.score = v
	statemanager.StateUpdate(ps.base+".Score", v)
}

func (ps *passScore) setAfterStarPass(v bool) {
	ps.afterStarPass = v
	statemanager.StateUpdate(ps.base+".AfterStarPass", v)
}

func (jt *jamTeam) addSkater(s *skater, p string) {
}
