// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"errors"
	"log"
	"time"

	"github.com/rollerderby/crg/state"
)

type masterClock struct {
	sb         *Scoreboard
	syncClocks bool
	clocks     map[string]*clock
	startTime  time.Time
	ticks      int64
	stateIDs   map[string]string

	period       *clock
	jam          *clock
	lineup       *clock
	timeout      *clock
	intermission *clock
}

const (
	clockPeriod       = "Period"
	clockJam          = "Jam"
	clockLineup       = "Lineup"
	clockTimeout      = "Timeout"
	clockIntermission = "Intermission"
)

const clockTicksPerSecond int64 = 10
const durationPerTick = time.Second / time.Duration(clockTicksPerSecond)

var clockTimeTick = 1000 / clockTicksPerSecond
var errClockNotFound = errors.New("Clock not found")

func newMasterClock(sb *Scoreboard) *masterClock {
	mc := &masterClock{
		sb:         sb,
		syncClocks: true,
		clocks:     make(map[string]*clock),
		stateIDs:   make(map[string]string),
	}

	ticksPerSecond := int64(1000)
	minutes2 := 2 * 60 * ticksPerSecond
	minutes15 := 15 * 60 * ticksPerSecond
	minutes30 := 30 * 60 * ticksPerSecond

	mc.period = newClock(
		sb,
		clockPeriod,
		1, 2,
		0, minutes30,
		true,
		false,
	)
	mc.jam = newClock(
		sb,
		clockJam,
		1, 99,
		0, minutes2,
		true,
		false,
	)
	mc.lineup = newClock(
		sb,
		clockLineup,
		1, 99,
		0, minutes30,
		false,
		false,
	)
	mc.timeout = newClock(
		sb,
		clockTimeout,
		1, 99,
		0, minutes30,
		false,
		false,
	)
	mc.intermission = newClock(
		sb,
		clockIntermission,
		1, 2,
		0, minutes15,
		true,
		false,
	)
	mc.clocks[clockPeriod] = mc.period
	mc.clocks[clockJam] = mc.jam
	mc.clocks[clockLineup] = mc.lineup
	mc.clocks[clockTimeout] = mc.timeout
	mc.clocks[clockIntermission] = mc.intermission
	for _, c := range mc.clocks {
		c.reset(true, false)
	}

	mc.stateIDs["startTime"] = sb.stateBase() + ".MasterClock.StartTime"
	mc.stateIDs["ticks"] = sb.stateBase() + ".MasterClock.Ticks"

	state.RegisterUpdaterTime(mc.stateIDs["startTime"], 0, mc.setStartTime)
	state.RegisterUpdaterInt64(mc.stateIDs["ticks"], 0, mc.setTicks)

	go mc.tickClocks()

	return mc
}

func (mc *masterClock) reset() {
	for _, c := range mc.clocks {
		c.reset(true, false)
	}

	mc.setStartTime(time.Now())
	mc.setTicks(0)
}

func (mc *masterClock) stateBase() string {
	return mc.sb.stateBase()
}

// Called from tickClocks() and stateSnapshot.rollback()
// state lock MUST be held before calling and
// released after ticker() returns by the caller
func (mc *masterClock) ticker() {
	now := time.Now()
	ticksFromStart := int64(now.Sub(mc.startTime) / durationPerTick)
	ticksToDo := ticksFromStart - mc.ticks
	var clocks []*clock
	for _, c := range mc.clocks {
		clocks = append(clocks, c)
	}

	if ticksToDo > 1 {
		log.Printf("Ticking %v times, now: %v, ticksFromStart: %v", ticksToDo, now, ticksFromStart)
	}
	for i := int64(0); i < ticksToDo; i++ {
		clockExpired := false
		for _, c := range clocks {
			if c.isRunning() {
				if c.tick(clockTimeTick) {
					clockExpired = true
				}
			}
		}
		mc.setTicks(mc.ticks + 1)
		mc.sb.activeSnapshot.updateLength()
		if clockExpired {
			mc.sb.clocksExpired()
		}
	}
}

func (mc *masterClock) tickClocks() {
	ticker := time.NewTicker(durationPerTick)
	for range ticker.C {
		state.Lock()
		mc.ticker()
		state.Unlock()
	}
}

func (mc *masterClock) setStartTime(v time.Time) error {
	mc.startTime = v
	state.StateUpdateTime(mc.stateIDs["startTime"], v)
	return nil
}

func (mc *masterClock) setTicks(v int64) error {
	mc.ticks = v
	state.StateUpdateInt64(mc.stateIDs["ticks"], v)
	return nil
}

func (mc *masterClock) setRunningClocks(clocksToStart ...string) {
	for _, c := range mc.clocks {
		c.stop()
	}
	for _, id := range clocksToStart {
		mc.startClock(id)
	}
}

func (mc *masterClock) startClock(id string) error {
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}
	mc.triggerClockStart(c)
	return nil
}

func (mc *masterClock) startCmd(data []string) error {
	if len(data) < 1 {
		return errClockNotFound
	}
	return mc.startClock(data[0])
}

func (mc *masterClock) stopClock(id string) error {
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}
	c.stop()
	return nil
}

func (mc *masterClock) setClockAdjustable(id string, adjustable bool) error {
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}
	c.setAdjustable(adjustable)
	return nil
}

func (mc *masterClock) stopCmd(data []string) error {
	if len(data) < 1 {
		return errClockNotFound
	}
	return mc.stopClock(data[0])
}

func (mc *masterClock) resetClock(id string) error {
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}
	c.reset(true, false)
	return nil
}

func (mc *masterClock) resetCmd(data []string) error {
	if len(data) < 1 {
		return errClockNotFound
	}
	return mc.resetClock(data[0])
}

func calculateClockOffset(master, slave *clock) int64 {
	// figure out when to start the clock based on the master clock

	// Calc master time to tick
	masterTimeToTick := master.time.num % 1000
	if !master.countdown {
		masterTimeToTick = (1000 - masterTimeToTick) % 1000
	}

	// Calc slave time to tick
	slaveTimeToTick := slave.time.num % 1000
	if !slave.countdown {
		slaveTimeToTick = (1000 - slaveTimeToTick) % 1000
	}

	// Calc difference (normalizing between -500 and 500)
	diff := masterTimeToTick - slaveTimeToTick
	if diff < -500 {
		diff = diff + 1000
	} else if diff > 500 {
		diff = diff - 1000
	}

	// Invert if slave is countdown
	if slave.countdown {
		diff = -diff
	}

	return diff
}

func (mc *masterClock) triggerClockStart(c *clock) {
	if mc.syncClocks && c != mc.period && mc.period.running {
		offset := calculateClockOffset(mc.period, c)
		c.time.num = c.time.num - offset
	}

	c.setRunning(true)
}

func (mc *masterClock) CurrentTime() time.Time {
	return mc.startTime.Add(time.Duration(mc.ticks) * durationPerTick)
}
