package scoreboard

import (
	"errors"
	"fmt"
	"time"

	"github.com/rollerderby/crg/statemanager"
)

type stateSnapshot struct {
	sb         *Scoreboard
	state      string
	inProgress bool
	canRevert  bool
	startTicks int64
	endTicks   int64
	length     int64
	startTime  time.Time
	endTime    time.Time
	teams      []*stateSnapshotTeam
	clocks     map[string]*stateSnapshotClock
	base       string
	stateIDs   map[string]string
}

type stateSnapshotTeam struct {
	timeouts               int64
	officialReviews        int64
	officialReviewRetained bool
}

type stateSnapshotClock struct {
	number    int64
	startTime int64
	endTime   int64
	running   bool
}

var errCannotRollback = errors.New("Cannot rollback state")

func newStateSnapshot(sb *Scoreboard, idx int) *stateSnapshot {
	ss := &stateSnapshot{
		sb:         sb,
		state:      sb.state,
		inProgress: true,
		canRevert:  false,
		startTime:  time.Now(),
		startTicks: sb.clocks.ticks,
		teams:      make([]*stateSnapshotTeam, 2),
		clocks:     make(map[string]*stateSnapshotClock),
		stateIDs:   make(map[string]string),
		length:     0,
	}

	ss.base = fmt.Sprintf("%v.Snapshot(%v)", sb.stateBase(), idx)
	ss.stateIDs["state"] = ss.base + ".State"
	ss.stateIDs["inProgress"] = ss.base + ".InProgress"
	ss.stateIDs["canRevert"] = ss.base + ".CanRevert"
	ss.stateIDs["startTicks"] = ss.base + ".StartTicks"
	ss.stateIDs["endTicks"] = ss.base + ".EndTicks"
	ss.stateIDs["length"] = ss.base + ".Length"
	ss.stateIDs["startTime"] = ss.base + ".StartTime"
	ss.stateIDs["endTime"] = ss.base + ".EndTime"

	for t, team := range sb.teams {
		ss.teams[t] = &stateSnapshotTeam{
			timeouts:               team.timeouts,
			officialReviews:        team.officialReviews,
			officialReviewRetained: team.officialReviewRetained,
		}
	}

	statemanager.StateUpdate(ss.stateIDs["state"], ss.state)
	statemanager.StateUpdate(ss.stateIDs["inProgress"], ss.inProgress)
	statemanager.StateUpdate(ss.stateIDs["canRevert"], ss.canRevert)
	statemanager.StateUpdate(ss.stateIDs["startTicks"], ss.startTicks)
	statemanager.StateUpdate(ss.stateIDs["endTicks"], ss.endTicks)
	statemanager.StateUpdate(ss.stateIDs["startTime"], ss.startTime)
	statemanager.StateUpdate(ss.stateIDs["endTime"], ss.endTime)
	statemanager.StateUpdate(ss.stateIDs["length"], ss.length)

	for _, c := range sb.clocks.clocks {
		ss.clocks[c.name] = &stateSnapshotClock{
			number:    c.number.num,
			startTime: c.time.num,
			endTime:   0,
			running:   c.running,
		}
		ss.stateIDs[c.name+".number"] = ss.base + ".Clock(" + c.name + ").Number"
		ss.stateIDs[c.name+".startTime"] = ss.base + ".Clock(" + c.name + ").StartTime"
		ss.stateIDs[c.name+".endTime"] = ss.base + ".Clock(" + c.name + ").EndTime"
		ss.stateIDs[c.name+".running"] = ss.base + ".Clock(" + c.name + ").Running"

		statemanager.StateUpdate(ss.stateIDs[c.name+".number"], c.number.num)
		statemanager.StateUpdate(ss.stateIDs[c.name+".startTime"], c.time.num)
		statemanager.StateUpdate(ss.stateIDs[c.name+".endTime"], int64(0))
		statemanager.StateUpdate(ss.stateIDs[c.name+".running"], c.running)
	}
	return ss
}

func (ss *stateSnapshot) end(canRevert bool) {
	ss.updateLength()
	ss.endTicks = ss.sb.clocks.ticks
	ss.endTime = time.Now()
	ss.canRevert = canRevert
	ss.inProgress = false
	for _, c := range ss.sb.clocks.clocks {
		ss.clocks[c.name].endTime = c.time.num
		statemanager.StateUpdate(ss.stateIDs[c.name+".endTime"], c.time.num)
	}

	statemanager.StateUpdate(ss.stateIDs["canRevert"], ss.canRevert)
	statemanager.StateUpdate(ss.stateIDs["inProgress"], ss.inProgress)
	statemanager.StateUpdate(ss.stateIDs["endTicks"], ss.endTicks)
	statemanager.StateUpdate(ss.stateIDs["endTime"], ss.endTime)
}

func (ss *stateSnapshot) unend() {
	ss.updateLength()
	ss.canRevert = false
	ss.endTicks = 0
	ss.endTime = time.Time{}
	ss.inProgress = true
	for name, c := range ss.clocks {
		c.endTime = 0
		statemanager.StateUpdate(ss.stateIDs[name+".endTime"], c.endTime)
	}

	statemanager.StateUpdate(ss.stateIDs["canRevert"], ss.canRevert)
	statemanager.StateUpdate(ss.stateIDs["inProgress"], ss.inProgress)
	statemanager.StateUpdate(ss.stateIDs["endTicks"], ss.endTicks)
	statemanager.StateUpdate(ss.stateIDs["endTime"], ss.endTime)
}

func (ss *stateSnapshot) updateLength() {
	ss.length = (ss.sb.clocks.ticks - ss.startTicks) * clockTimeTick
	statemanager.StateUpdate(ss.stateIDs["length"], ss.length)
}

func (ss *stateSnapshot) delete() {
	statemanager.StateUpdate(ss.base, nil)
}

func (ss *stateSnapshot) period() int64 {
	return ss.clocks[clockPeriod].number
}

func (ss *stateSnapshot) jam() int64 {
	return ss.clocks[clockJam].number
}
