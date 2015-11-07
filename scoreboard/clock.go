package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type clock struct {
	sb         *Scoreboard
	base       string
	name       string
	number     *minMaxNumber
	time       *minMaxNumber
	countdown  bool
	running    bool
	adjustable bool
	stateIDs   map[string]string
}

func newClock(sb *Scoreboard, name string, numMin, numMax, timeMin, timeMax int64, countdown, running bool) *clock {
	timeNum := timeMin
	if countdown {
		timeNum = timeMax
	}
	c := &clock{
		sb:   sb,
		base: fmt.Sprintf("%s.Clock(%s)", sb.stateBase(), name),
	}
	c.number = newMinMaxNumber(c, "Number", false, numMin, numMax, numMin, 1)
	c.time = newMinMaxNumber(c, "Time", countdown, timeMin, timeMax, timeNum, 1000)
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
	c.time.setCountDown(countdown)
	return nil
}

func (c *clock) setRunning(running bool) error {
	c.running = running
	statemanager.StateUpdate(c.stateIDs["running"], running)
	return nil
}

func (c *clock) reset(full bool, incNumber bool) {
	if full {
		c.number.setNum(c.number.min)
	}
	if c.countdown {
		if incNumber && c.time.num != c.time.max {
			c.number.incNum()
		}
		c.time.setNum(c.time.max)
	} else {
		if incNumber && c.time.num != c.time.min {
			c.number.incNum()
		}
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

// returns true if clock timedout
func (c *clock) tick(tickDuration int64) bool {
	if !c.time.adjust(c.countdown, tickDuration) {
		c.stop()
		return true
	}
	return false
}

func (c *clock) SetTime(time int64) {
}

func (c *clock) setAdjustable(adjustable bool) {
	c.adjustable = adjustable
	statemanager.StateUpdate(c.stateIDs["adjustable"], adjustable)
}

func (c *clock) clone() *clock {
	return nil
}
