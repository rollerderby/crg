// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"fmt"
	"strconv"

	"github.com/rollerderby/crg/statemanager"
)

type jam struct {
	sb       *Scoreboard
	idx      int64
	period   int64
	jam      int64
	stateIDs map[string]string
}

func blankJam(sb *Scoreboard) *jam {
	j := &jam{
		sb:       sb,
		idx:      int64(len(sb.jams)),
		stateIDs: make(map[string]string),
	}

	base := fmt.Sprintf("%v.Jam(%v)", sb.stateBase(), j.idx)
	j.stateIDs["idx"] = base + ".Idx"
	j.stateIDs["period"] = base + ".Period"
	j.stateIDs["jam"] = base + ".Jam"

	j.setPeriod(0)
	j.setJam(0)

	sb.jams = append(sb.jams, j)

	return j
}

func newJam(sb *Scoreboard) *jam {
	j := blankJam(sb)

	if len(sb.jams) == 1 {
		j.setPeriod(1)
		j.setJam(1)
	} else {
		j.setPeriod(sb.masterClock.period.number.num)
		j.setJam(sb.masterClock.period.number.num + 1)
	}

	return j
}

func (j *jam) setPeriod(v int64) error {
	j.period = v
	return statemanager.StateUpdateInt64(j.stateIDs["period"], v)
}

func (j *jam) setJam(v int64) error {
	j.jam = v
	return statemanager.StateUpdateInt64(j.stateIDs["jam"], v)
}

/* helper functions to find the jam for registerupdaters */
func (sb *Scoreboard) findJam(k string) *jam {
	ids := statemanager.ParseIDs(k)
	if len(ids) == 0 {
		return nil
	}
	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return nil
	}

	// generate blank snapshots if needed
	for i := int64(len(sb.jams)); i <= id; i++ {
		sb.jams = append(sb.jams, blankJam(sb))
	}

	return sb.jams[id]
}
