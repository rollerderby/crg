package scoreboard

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rollerderby/crg/statemanager"
)

type masterClock struct {
	parent parent
	clocks map[string]*clock
}

type clock struct {
	parent     parent
	base       string
	name       string
	number     *minMaxNumber
	time       *minMaxNumber
	countdown  bool
	running    bool
	adjustable bool
	stateIDs   map[string]string

	// used for precise timing and undos
	startedSec  int64
	startedTime time.Time
}

const (
	clockPeriod       = "Period"
	clockJam          = "Jam"
	clockLineup       = "Lineup"
	clockTimeout      = "Timeout"
	clockIntermission = "Intermission"
)

const clockTicksPerSecond int64 = 5

var clockTimeTick int64 = 1000 / clockTicksPerSecond
var errClockNotFound = errors.New("Clock not found")

func newMasterClock(p parent) *masterClock {
	mc := &masterClock{
		parent: p,
		clocks: make(map[string]*clock),
	}

	ticksPerSecond := int64(1000)
	minutes2 := 2 * 60 * ticksPerSecond
	minutes15 := 15 * 60 * ticksPerSecond
	minutes30 := 30 * 60 * ticksPerSecond

	mc.clocks[clockPeriod] = newClock(
		p,
		clockPeriod,
		1, 2,
		0, minutes30,
		true,
		false,
	)
	mc.clocks[clockJam] = newClock(
		p,
		clockJam,
		1, 99,
		0, minutes2,
		true,
		false,
	)
	mc.clocks[clockLineup] = newClock(
		p,
		clockLineup,
		1, 99,
		0, minutes30,
		false,
		false,
	)
	mc.clocks[clockTimeout] = newClock(
		p,
		clockTimeout,
		1, 99,
		0, minutes30,
		false,
		false,
	)
	mc.clocks[clockIntermission] = newClock(
		p,
		clockIntermission,
		0, 2,
		0, minutes15,
		true,
		false,
	)
	for _, c := range mc.clocks {
		c.reset(true)
	}

	statemanager.RegisterCommand("ClockAdjustTime", mc.adjustTimeCmd)
	statemanager.RegisterCommand("ClockAdjustNumber", mc.adjustNumberCmd)

	statemanager.RegisterCommand("ClockStart", mc.startCmd)
	statemanager.RegisterCommand("ClockStop", mc.stopCmd)
	statemanager.RegisterCommand("ClockReset", mc.resetCmd)

	go mc.tickClocks()

	return mc
}

func (mc *masterClock) stateBase() string {
	return mc.parent.stateBase()
}

func (mc *masterClock) tickClocks() {
	ticker := time.NewTicker(time.Second / time.Duration(clockTicksPerSecond))
	for now := range ticker.C {
		statemanager.Lock()
		for _, c := range mc.clocks {
			if c.isRunning() {
				c.tick(now, clockTimeTick)
			}
		}
		statemanager.Unlock()
	}
}

func (mc *masterClock) setRunningClocks(clocksToStart []string, clocksToReset []string) error {
	for _, c := range mc.clocks {
		c.stop()
	}
	for _, id := range clocksToReset {
		mc.resetClock(id)
	}
	for _, id := range clocksToStart {
		mc.startClock(id)
	}
	return nil
}

func (mc *masterClock) startClock(id string) error {
	c, ok := mc.clocks[id]
	if !ok {
		return errClockNotFound
	}
	c.start()
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
	c.reset(true)
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

func newClock(p parent, name string, numMin, numMax, timeMin, timeMax int64, countdown, running bool) *clock {
	timeNum := timeMin
	if countdown {
		timeNum = timeMax
	}
	c := &clock{
		parent: p,
		base:   fmt.Sprintf("%s.Clock(%s)", p.stateBase(), name),
	}
	c.number = newMinMaxNumber(c, "Number", numMin, numMax, numMin, 1)
	c.time = newMinMaxNumber(c, "Time", timeMin, timeMax, timeNum, 1000)
	c.stateIDs = make(map[string]string)

	c.stateIDs["name"] = c.base + ".Name"
	c.stateIDs["countdown"] = c.base + ".CountDown"
	c.stateIDs["running"] = c.base + ".Running"
	c.stateIDs["adjustable"] = c.base + ".Adjustable"

	statemanager.RegisterUpdater(c.stateIDs["name"], 0, c.setName)
	statemanager.RegisterUpdaterBool(c.stateIDs["countdown"], 4, c.setCountDown)
	statemanager.RegisterUpdaterBool(c.stateIDs["running"], 4, c.setRunning)

	c.setName(name)
	c.setCountDown(countdown)
	c.setRunning(running)
	c.setAdjustable(false)

	return c
}

func (c *clock) stateBase() string {
	return c.base
}

func (c *clock) setName(name string) error {
	c.name = name
	statemanager.StateUpdate(c.stateIDs["name"], name)
	return nil
}

func (c *clock) setCountDown(countdown bool) error {
	c.countdown = countdown
	statemanager.StateUpdate(c.stateIDs["countdown"], countdown)
	return nil
}

func (c *clock) setRunning(running bool) error {
	c.running = running
	statemanager.StateUpdate(c.stateIDs["running"], running)
	return nil
}

func (c *clock) reset(full bool) {
	if full {
		c.number.setNum(c.number.min)
	}
	if c.countdown {
		c.time.setNum(c.time.max)
	} else {
		c.time.setNum(c.time.min)
	}
	c.stop()
}

func (c *clock) isRunning() bool {
	return c.running
}

func (c *clock) start() {
	c.running = true
	statemanager.StateUpdate(c.stateIDs["running"], c.running)
}

func (c *clock) stop() {
	c.running = false
	statemanager.StateUpdate(c.stateIDs["running"], c.running)
}

func (c *clock) tick(now time.Time, tickDuration int64) {
	if !c.time.adjust(c.countdown, tickDuration) {
		c.stop()
	}
}

func (c *clock) SetTime(time int64) {
}

func (c *clock) setAdjustable(adjustable bool) {
	c.adjustable = adjustable
	statemanager.StateUpdate(c.stateIDs["adjustable"], adjustable)
}
