package scoreboard

import (
	"errors"
	"strconv"
	"time"

	"github.com/rollerderby/crg/statemanager"
)

type masterClock struct {
	sb         *Scoreboard
	syncClocks bool
	clocks     map[string]*clock

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

var clockTimeTick = 1000 / clockTicksPerSecond
var errClockNotFound = errors.New("Clock not found")

func newMasterClock(sb *Scoreboard) *masterClock {
	mc := &masterClock{
		sb:         sb,
		syncClocks: true,
		clocks:     make(map[string]*clock),
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

	statemanager.RegisterCommand("ClockAdjustTime", mc.adjustTimeCmd)
	statemanager.RegisterCommand("ClockAdjustNumber", mc.adjustNumberCmd)

	go mc.tickClocks()

	return mc
}

func (mc *masterClock) stateBase() string {
	return mc.sb.stateBase()
}

func (mc *masterClock) tickClocks() {
	ticker := time.NewTicker(time.Second / time.Duration(clockTicksPerSecond))
	for now := range ticker.C {
		clockExpired := false
		statemanager.Lock()
		for _, c := range mc.clocks {
			if c.isRunning() {
				if c.tick(now, clockTimeTick) {
					clockExpired = true
				}
			}
		}
		if clockExpired {
			mc.sb.clocksExpired()
		}
		statemanager.Unlock()
	}
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

func (mc *masterClock) adjustTimeCmd(data []string) error {
	if len(data) < 1 {
		return errClockNotFound
	}
	id := data[0]
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}

	by, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return err
	}
	c.time.adjust(false, by)
	return nil
}

func (mc *masterClock) adjustNumberCmd(data []string) error {
	if len(data) < 1 {
		return errClockNotFound
	}
	id := data[0]
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}

	by, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return err
	}
	c.number.adjust(false, by)
	return nil
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
