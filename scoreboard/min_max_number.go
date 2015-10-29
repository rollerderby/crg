package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type minMaxNumber struct {
	parent   parent
	base     string
	num      int64
	min      int64
	max      int64
	updateOn int64
	stateIDs map[string]string
}

func newMinMaxNumber(p parent, id string, min, max, num, updateOn int64) *minMaxNumber {
	mmn := &minMaxNumber{
		parent:   p,
		base:     fmt.Sprintf("%s.%s", p.stateBase(), id),
		updateOn: updateOn,
		stateIDs: make(map[string]string),
	}

	mmn.stateIDs["min"] = mmn.base + ".Min"
	mmn.stateIDs["max"] = mmn.base + ".Max"
	mmn.stateIDs["num"] = mmn.base + ".Num"

	statemanager.RegisterUpdaterInt64(mmn.stateIDs["min"], 1, mmn.setMin)
	statemanager.RegisterUpdaterInt64(mmn.stateIDs["max"], 2, mmn.setMax)
	statemanager.RegisterUpdaterInt64(mmn.stateIDs["num"], 3, mmn.setNum)

	mmn.setMin(min)
	mmn.setMax(max)
	mmn.setNum(num)

	return mmn
}

func (mmn *minMaxNumber) adjust(down bool, adjust int64) bool {
	last := mmn.num
	defer func() {
		diff := last - mmn.num
		if diff < 0 {
			diff = -diff
		}
		if mmn.num%mmn.updateOn == 0 || diff >= mmn.updateOn {
			statemanager.StateUpdate(mmn.stateIDs["num"], mmn.num)
		}
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

	statemanager.StateUpdate(mmn.stateIDs["min"], mmn.min)
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

	statemanager.StateUpdate(mmn.stateIDs["max"], mmn.max)
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

	statemanager.StateUpdate(mmn.stateIDs["num"], mmn.num)
	return nil
}
