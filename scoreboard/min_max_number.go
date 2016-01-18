// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type minMaxNumber struct {
	parent    parent
	base      string
	countdown bool
	num       int64
	min       int64
	max       int64
	updateOn  int64
	stateIDs  map[string]string
}

func newMinMaxNumber(p parent, id string, countdown bool, min, max, num, updateOn int64) *minMaxNumber {
	mmn := &minMaxNumber{
		parent:    p,
		base:      fmt.Sprintf("%s.%s", p.stateBase(), id),
		countdown: countdown,
		updateOn:  updateOn,
		stateIDs:  make(map[string]string),
	}

	mmn.stateIDs["min"] = mmn.base + ".Min"
	mmn.stateIDs["max"] = mmn.base + ".Max"
	mmn.stateIDs["num"] = mmn.base + ".Num"
	mmn.stateIDs["precise"] = mmn.base + ".PreciseNum"

	statemanager.RegisterUpdaterInt64(mmn.stateIDs["min"], 1, mmn.setMin)
	statemanager.RegisterUpdaterInt64(mmn.stateIDs["max"], 2, mmn.setMax)
	statemanager.RegisterUpdaterInt64(mmn.stateIDs["precise"], 3, mmn.setNum)

	mmn.setMin(min)
	mmn.setMax(max)
	mmn.setNum(num)

	return mmn
}

func (mmn *minMaxNumber) sendNumStateUpdate() {
	statemanager.StateUpdateInt64(mmn.stateIDs["precise"], mmn.num)
	diff := mmn.num % mmn.updateOn
	if mmn.countdown {
		diff = -((mmn.updateOn - diff) % mmn.updateOn)
	}

	num := mmn.num - diff
	if num < mmn.min {
		num = mmn.min
	} else if num > mmn.max {
		num = mmn.max
	}
	statemanager.StateUpdateInt64(mmn.stateIDs["num"], num)
}

func (mmn *minMaxNumber) adjust(down bool, adjust int64) bool {
	defer func() {
		mmn.sendNumStateUpdate()
	}()
	if adjust < 0 {
		adjust = -adjust
		down = !down
	}
	if down {
		if mmn.num <= mmn.min+adjust {
			mmn.num = mmn.min
			return false
		}
		mmn.num = mmn.num - adjust
	} else {
		if mmn.num >= mmn.max-adjust {
			mmn.num = mmn.max
			return false
		}
		mmn.num = mmn.num + adjust
	}
	return true
}

func (mmn *minMaxNumber) setMin(v int64) error {
	mmn.min = v
	if mmn.min > mmn.max {
		mmn.setMax(mmn.min)
	}
	if mmn.min > mmn.num {
		mmn.setNum(mmn.num)
	}

	statemanager.StateUpdateInt64(mmn.stateIDs["min"], mmn.min)
	return nil
}

func (mmn *minMaxNumber) setMax(v int64) error {
	mmn.max = v
	if mmn.min > mmn.max {
		mmn.setMin(mmn.max)
	}
	if mmn.num > mmn.max {
		mmn.setNum(mmn.max)
	}

	statemanager.StateUpdateInt64(mmn.stateIDs["max"], mmn.max)
	return nil
}

func (mmn *minMaxNumber) setNum(v int64) error {
	if v < mmn.min {
		v = mmn.min
	} else if v > mmn.max {
		v = mmn.max
	}
	if mmn.num != v {
		mmn.num = v
	}

	mmn.sendNumStateUpdate()
	return nil
}

func (mmn *minMaxNumber) setCountDown(countdown bool) {
	mmn.countdown = countdown
	mmn.sendNumStateUpdate()
}

func (mmn *minMaxNumber) incNum() error {
	return mmn.setNum(mmn.num + 1)
}
