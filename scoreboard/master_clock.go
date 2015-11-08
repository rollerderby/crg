package scoreboard

import (
	"errors"
	"log"
	"time"

	"github.com/rollerderby/crg/statemanager"
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

const clockTicksPerSecond int64 = 5
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
		0, 2,
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

	statemanager.RegisterUpdaterTime(mc.stateIDs["startTime"], 0, mc.setStartTime)
	statemanager.RegisterUpdaterInt64(mc.stateIDs["ticks"], 0, mc.setTicks)

	mc.setStartTime(time.Now())
	mc.setTicks(0)

	go mc.tickClocks()

	return mc
}

func (mc *masterClock) stateBase() string {
	return mc.sb.stateBase()
}

// Called from tickClocks() and stateSnapshot.rollback()
// statemanager lock MUST be held before calling and
// released after ticker() returns by the caller
func (mc *masterClock) ticker() {
	now := time.Now()
	ticksFromStart := int64(now.Sub(mc.startTime) / durationPerTick)
	ticksToDo := ticksFromStart - mc.ticks

	if ticksToDo > 1 {
		log.Printf("Ticking %v times, now: %v, ticksFromStart: %v", ticksToDo, now, ticksFromStart)
	}
	for i := int64(0); i < ticksToDo; i++ {
		clockExpired := false
		for _, c := range mc.clocks {
			if c.isRunning() {
				if c.tick(clockTimeTick) {
					clockExpired = true
				}
			}
		}
		if clockExpired {
			mc.sb.clocksExpired()
		}
		mc.sb.activeSnapshot.updateLength()
	}

	mc.setTicks(ticksFromStart)
}

func (mc *masterClock) tickClocks() {
	ticker := time.NewTicker(durationPerTick)
	for range ticker.C {
		statemanager.Lock()
		mc.ticker()
		statemanager.Unlock()
	}
}

func (mc *masterClock) setStartTime(v time.Time) error {
	mc.startTime = v
	statemanager.StateUpdate(mc.stateIDs["startTime"], v)
	return nil
}

func (mc *masterClock) setTicks(v int64) error {
	mc.ticks = v
	statemanager.StateUpdate(mc.stateIDs["ticks"], v)
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
	m_ttt := master.time.num % 1000
	if !master.countdown {
		m_ttt = (1000 - m_ttt) % 1000
	}

	// Calc slave time to tick
	s_ttt := slave.time.num % 1000
	if !slave.countdown {
		s_ttt = (1000 - s_ttt) % 1000
	}

	// Calc difference (normalizing between -500 and 500)
	diff := m_ttt - s_ttt
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
