package scoreboard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rollerderby/crg/statemanager"
)

type stateSnapshot struct {
	sb         *Scoreboard
	idx        int64
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
	base                   string
	timeouts               int64
	officialReviews        int64
	officialReviewRetained bool
}

type stateSnapshotClock struct {
	base      string
	number    int64
	startTime int64
	endTime   int64
	running   bool
}

var errSnapshotNotFound = errors.New("Snapshot Not Found")
var errSnapshotClockNotFound = errors.New("Snapshot Clock Not Found")
var errSnapshotTeamNotFound = errors.New("Snapshot Team Not Found")

func newStateSnapshot(sb *Scoreboard, idx int64, initializeValues bool) *stateSnapshot {
	ss := &stateSnapshot{
		sb:       sb,
		idx:      idx,
		teams:    make([]*stateSnapshotTeam, 2),
		clocks:   make(map[string]*stateSnapshotClock),
		stateIDs: make(map[string]string),
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

	if initializeValues {
		ss.setState(sb.state)
		ss.setInProgress(true)
		ss.setCanRevert(false)
		ss.setStartTime(time.Now())
		ss.setStartTicks(sb.clocks.ticks)
		ss.setLength(0)
	}

	for id, team := range sb.teams {
		t := &stateSnapshotTeam{base: fmt.Sprintf("%v.Team(%v)", ss.base, id)}
		ss.teams[id] = t

		if initializeValues {
			t.setTimeouts(team.timeouts)
			t.setOfficialReviews(team.officialReviews)
			t.setOfficialReviewRetained(team.officialReviewRetained)
		}
	}

	for id, clock := range sb.clocks.clocks {
		c := &stateSnapshotClock{base: ss.base + ".Clock(" + id + ")"}
		ss.clocks[id] = c

		if initializeValues {
			c.setNumber(clock.number.num)
			c.setStartTime(clock.time.num)
			c.setEndTime(0)
			c.setRunning(clock.running)
		}
	}
	return ss
}

func (ss *stateSnapshot) end(canRevert bool) {
	ss.setEndTicks(ss.sb.clocks.ticks)
	ss.setEndTime(time.Now())
	ss.setCanRevert(canRevert)
	ss.setInProgress(false)
	ss.updateLength()

	for id, c := range ss.clocks {
		c.setEndTime(ss.sb.clocks.clocks[id].time.num)
	}
}

func (ss *stateSnapshot) unend() {
	ss.setCanRevert(false)
	ss.setEndTicks(0)
	ss.setEndTime(time.Time{})
	ss.setInProgress(true)
	ss.updateLength()

	for name, c := range ss.clocks {
		c.setEndTime(0)
		statemanager.StateUpdate(ss.stateIDs[name+".endTime"], c.endTime)
	}
}

func (ss *stateSnapshot) updateLength() {
	if ss.idx != 0 {
		ss.setLength((ss.sb.clocks.ticks - ss.startTicks) * clockTimeTick)
	}
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

func (ss *stateSnapshot) setState(v string) {
	ss.state = v
	statemanager.StateUpdate(ss.stateIDs["state"], v)
}

func (ss *stateSnapshot) setInProgress(v bool) {
	ss.inProgress = v
	statemanager.StateUpdate(ss.stateIDs["inProgress"], v)
}

func (ss *stateSnapshot) setCanRevert(v bool) {
	ss.canRevert = v
	statemanager.StateUpdate(ss.stateIDs["canRevert"], v)
}

func (ss *stateSnapshot) setStartTicks(v int64) {
	ss.startTicks = v
	statemanager.StateUpdate(ss.stateIDs["startTicks"], v)
}

func (ss *stateSnapshot) setEndTicks(v int64) {
	ss.endTicks = v
	statemanager.StateUpdate(ss.stateIDs["endTicks"], v)
}

func (ss *stateSnapshot) setStartTime(v time.Time) {
	ss.startTime = v
	statemanager.StateUpdate(ss.stateIDs["startTime"], v)
}

func (ss *stateSnapshot) setEndTime(v time.Time) {
	ss.endTime = v
	statemanager.StateUpdate(ss.stateIDs["endTime"], v)
}

func (ss *stateSnapshot) setLength(v int64) {
	ss.length = v
	statemanager.StateUpdate(ss.stateIDs["length"], v)
}

func (c *stateSnapshotClock) setNumber(v int64) {
	c.number = v
	statemanager.StateUpdate(c.base+".Number", v)
}

func (c *stateSnapshotClock) setStartTime(v int64) {
	c.startTime = v
	statemanager.StateUpdate(c.base+".StartTime", v)
}

func (c *stateSnapshotClock) setEndTime(v int64) {
	c.endTime = v
	statemanager.StateUpdate(c.base+".EndTime", v)
}

func (c *stateSnapshotClock) setRunning(v bool) {
	c.running = v
	statemanager.StateUpdate(c.base+".Running", v)
}

func (t *stateSnapshotTeam) setTimeouts(v int64) {
	t.timeouts = v
	statemanager.StateUpdate(t.base+".Timeouts", v)
}

func (t *stateSnapshotTeam) setOfficialReviews(v int64) {
	t.officialReviews = v
	statemanager.StateUpdate(t.base+".OfficialReviews", v)
}

func (t *stateSnapshotTeam) setOfficialReviewRetained(v bool) {
	t.officialReviewRetained = v
	statemanager.StateUpdate(t.base+".OfficialReviewRetained", v)
}

func (ss *stateSnapshot) findClock(k string) *stateSnapshotClock {
	for _, c := range ss.clocks {
		if strings.HasPrefix(k, c.base+".") {
			return c
		}
	}
	return nil
}
func (ss *stateSnapshot) findTeam(k string) *stateSnapshotTeam {
	for _, t := range ss.teams {
		if strings.HasPrefix(k, t.base+".") {
			return t
		}
	}
	return nil
}

/* Helper functions to find the stateSnapshot for RegisterUpdaters */
/*
	teams      []*stateSnapshotTeam
	clocks     map[string]*stateSnapshotClock
*/
func (sb *Scoreboard) findStateSnapshot(k string) *stateSnapshot {
	k = k[len(sb.stateBase()+".Snapshot("):]
	end := strings.IndexRune(k, ')')
	if end <= 0 {
		return nil
	}
	id, err := strconv.ParseInt(k[:end], 10, 64)
	if err != nil {
		return nil
	}

	// Generate blank snapshots if needed
	for i := int64(len(sb.snapshots)); i <= id; i++ {
		sb.snapshots = append(sb.snapshots, newStateSnapshot(sb, i, false))
	}

	return sb.snapshots[id]
}
func (sb *Scoreboard) ssSetState(k string, v string) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setState(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetInProgress(k string, v bool) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setInProgress(v)
		if v {
			sb.activeSnapshot = ss
		}
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetCanRevert(k string, v bool) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setCanRevert(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetStartTicks(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setStartTicks(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetEndTicks(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setEndTicks(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetLength(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setLength(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetStartTime(k string, v time.Time) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setStartTime(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) ssSetEndTime(k string, v time.Time) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		ss.setEndTime(v)
		return nil
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sscSetNumber(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if c := ss.findClock(k); c != nil {
			c.setNumber(v)
			return nil
		}
		return errSnapshotClockNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sscSetStartTime(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if c := ss.findClock(k); c != nil {
			c.setStartTime(v)
			return nil
		}
		return errSnapshotClockNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sscSetEndTime(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if c := ss.findClock(k); c != nil {
			c.setEndTime(v)
			return nil
		}
		return errSnapshotClockNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sscSetRunning(k string, v bool) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if c := ss.findClock(k); c != nil {
			c.setRunning(v)
			return nil
		}
		return errSnapshotClockNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sstSetTimeouts(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if t := ss.findTeam(k); t != nil {
			t.setTimeouts(v)
			return nil
		}
		return errSnapshotTeamNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sstSetOfficialReviews(k string, v int64) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if t := ss.findTeam(k); t != nil {
			t.setOfficialReviews(v)
			return nil
		}
		return errSnapshotTeamNotFound
	}
	return errSnapshotNotFound
}
func (sb *Scoreboard) sstSetOfficialReviewRetained(k string, v bool) error {
	if ss := sb.findStateSnapshot(k); ss != nil {
		if t := ss.findTeam(k); t != nil {
			t.setOfficialReviewRetained(v)
			return nil
		}
		return errSnapshotTeamNotFound
	}
	return errSnapshotNotFound
}
