// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"fmt"
	"log"
	"strconv"

	"github.com/rollerderby/crg/state"
)

const (
	starPassNo    = "No"
	starPassYes   = "Yes"
	starPassOther = "Other"
)

type jam struct {
	sb       *Scoreboard
	lastJam  *jam
	idx      int64
	period   int64
	jam      int64
	base     string
	teams    [2]jamTeam
	stateIDs map[string]string
}

type jamTeam struct {
	starPass string
	jammer   string
	pivot    string
	noPivot  bool
	blockers []string
	scores   [9]int64
}

func blankJam(sb *Scoreboard) *jam {
	j := &jam{
		sb:       sb,
		idx:      int64(len(sb.jams)),
		stateIDs: make(map[string]string),
		base:     fmt.Sprintf("%v.Jam(%v)", sb.stateBase(), len(sb.jams)),
	}

	j.stateIDs["idx"] = j.base + ".Idx"
	j.stateIDs["period"] = j.base + ".Period"
	j.stateIDs["jam"] = j.base + ".Jam"

	j.setPeriod(0)
	j.setJam(0)

	j.lastJam = sb.activeJam
	sb.jams = append(sb.jams, j)
	sb.activeJam = j

	return j
}

func newJam(sb *Scoreboard) *jam {
	j := blankJam(sb)

	if len(sb.jams) == 1 {
		j.setPeriod(1)
		j.setJam(1)
	} else {
		j.setPeriod(sb.masterClock.period.number.num)
		j.setJam(sb.masterClock.jam.number.num + 1)
	}

	for _, t := range sb.teams {
		for _, s := range t.skaters {
			s.setPosition(positionBench)
			j.setTeamPosition(s)
		}
	}
	return j
}

func (j *jam) updateJam() {
	j.setPeriod(j.sb.masterClock.period.number.num)
	j.setJam(j.sb.masterClock.jam.number.num)
}

func (j *jam) delete() {
	state.StateDelete(j.base)
	j.sb = nil
}

func (j *jam) reinstatePositions() {
	reinstate := func(t *team, jt *jamTeam) {
		set := func(id, position string) error {
			if s, ok := t.skaters[id]; ok {
				s.position = position
				return nil
			}
			return errSkaterNotFound
		}
		log.Printf("UNBENCH! %+v", jt)

		for _, s := range t.skaters {
			s.position = positionBench
		}

		set(jt.jammer, positionJammer)
		set(jt.pivot, positionPivot)
		for _, blocker := range jt.blockers {
			set(blocker, positionBlocker)
		}

		for _, s := range t.skaters {
			s.setPosition(s.position)
		}
	}
	reinstate(j.sb.teams[0], &j.teams[0])
	reinstate(j.sb.teams[1], &j.teams[1])
}

func (j *jam) clearTeamPositions(t *team) {
	base := fmt.Sprintf("%v.Team(%v)", j.base, t.id)
	j.teams[t.id-1].jammer = ""
	state.StateDelete(base + ".Jammer")
	j.teams[t.id-1].pivot = ""
	state.StateDelete(base + ".Pivot")
	for idx, _ := range j.teams[t.id-1].blockers {
		state.StateDelete(fmt.Sprintf("%v.Blocker(%v)", base, idx))
	}
	j.teams[t.id-1].blockers = nil
}

func (j *jam) setTeamPosition(s *skater) {
	base := fmt.Sprintf("%v.Team(%v)", j.base, s.t.id)
	switch s.position {
	case positionJammer:
		j.teams[s.t.id-1].jammer = s.id
		state.StateUpdateString(base+".Jammer", s.id)
	case positionPivot:
		j.teams[s.t.id-1].jammer = s.id
		state.StateUpdateString(base+".Pivot", s.id)
	case positionBlocker:
		j.teams[s.t.id-1].blockers = append(j.teams[s.t.id-1].blockers, s.id)
		state.StateUpdateString(fmt.Sprintf("%v.Blocker(%v)", base, len(j.teams[s.t.id-1].blockers)-1), s.id)
	}
}

func (j *jam) setPeriod(v int64) error {
	j.period = v
	return state.StateUpdateInt64(j.stateIDs["period"], v)
}

func (j *jam) setJam(v int64) error {
	j.jam = v
	return state.StateUpdateInt64(j.stateIDs["jam"], v)
}

/* helper functions to find the jam for registerupdaters */
func (sb *Scoreboard) findJam(k string) *jam {
	ids := state.ParseIDs(k)
	if len(ids) == 0 {
		return nil
	}
	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return nil
	}

	// generate blank snapshots if needed
	for i := int64(len(sb.jams)); i <= id; i++ {
		sb.jams = append(sb.jams, blankJam(sb))
	}

	return sb.jams[id]
}
