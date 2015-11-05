package scoreboard

import (
	"errors"
	"log"
	"time"
)

type stateSnapshot struct {
	state              string
	canUndo            bool
	startTime, endTime time.Time
	clocks             map[string]*stateSnapshotClock
}

type stateSnapshotClock struct {
	number    int64
	startTime int64
	endTime   int64
}

var errCannotRollback = errors.New("Cannot rollback state")

func newStateSnapshot(sb *Scoreboard, canUndo bool) *stateSnapshot {
	ss := &stateSnapshot{
		state:   sb.state,
		canUndo: canUndo,
	}

	ss.startTime = time.Now()
	for _, c := range sb.clocks.clocks {
		ss.clocks[c.name] = &stateSnapshotClock{
			number:    c.number.num,
			startTime: c.time.num,
		}
	}

	return ss
}

func (ss *stateSnapshot) end(sb *Scoreboard) {
	ss.endTime = time.Now()
	for _, c := range sb.clocks.clocks {
		ss.clocks[c.name].number = c.number.num
		ss.clocks[c.name].endTime = c.time.num
	}
}

func (ss *stateSnapshot) rollback(sb *Scoreboard) error {
	log.Print("stateSnapshot.rollback: NOT IMPLEMENTED")
	return errCannotRollback
}

func (ss *stateSnapshot) period() int64 {
	return ss.clocks[clockPeriod].number
}

func (ss *stateSnapshot) jam() int64 {
	return ss.clocks[clockJam].number
}
