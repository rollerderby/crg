// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"fmt"
	"strconv"

	"github.com/rollerderby/crg/statemanager"
)

type boxTrip struct {
	sb       *Scoreboard
	s        *skater
	in       boxTripTime
	out      boxTripTime
	stateIDs map[string]string
}

type boxTripTime struct {
	jamIdx        int64
	betweenJams   bool
	afterStarPass bool
}

func blankBoxTrip(s *skater) *boxTrip {
	bt := &boxTrip{
		sb:       s.t.sb,
		s:        s,
		stateIDs: make(map[string]string),
	}
	idx := len(s.boxTrips)
	base := fmt.Sprintf("%v.BoxTrip(%v)", s.base, idx)

	bt.stateIDs["skater"] = base + ".Skater"
	bt.stateIDs["in.jamIdx"] = base + ".In.JamIdx"
	bt.stateIDs["in.period"] = base + ".In.Period"
	bt.stateIDs["in.jam"] = base + ".In.Jam"
	bt.stateIDs["in.betweenJams"] = base + ".In.BetweenJams"
	bt.stateIDs["in.afterStarPass"] = base + ".In.AfterStarPass"
	bt.stateIDs["out.jamIdx"] = base + ".Out.JamIdx"
	bt.stateIDs["out.period"] = base + ".Out.Period"
	bt.stateIDs["out.jam"] = base + ".Out.Jam"
	bt.stateIDs["out.betweenJams"] = base + ".Out.BetweenJams"
	bt.stateIDs["out.afterStarPass"] = base + ".Out.AfterStarPass"

	return bt
}

func newBoxTrip(s *skater, jamIdx int64, betweenJams, afterStarPass bool) *boxTrip {
	bt := blankBoxTrip(s)

	bt.setSkater(s)
	bt.setInJamIdx(jamIdx)
	bt.setInBetweenJams(betweenJams)
	bt.setInAfterStarPass(afterStarPass)
	bt.setOutJamIdx(-1)
	bt.setOutBetweenJams(false)
	bt.setOutAfterStarPass(false)

	return bt
}

func (bt *boxTrip) setSkater(s *skater) {
	bt.s = s
	statemanager.StateUpdateString(bt.stateIDs["skater"], s.id)
}

func (bt *boxTrip) setInJamIdx(v int64) error {
	if v >= 0 && v < int64(len(bt.sb.jams)) {
		jam := bt.sb.jams[v]
		bt.in.jamIdx = jam.idx
		statemanager.StateUpdateInt64(bt.stateIDs["in.jamIdx"], jam.idx)
		statemanager.StateUpdateInt64(bt.stateIDs["in.period"], jam.period)
		statemanager.StateUpdateInt64(bt.stateIDs["in.jam"], jam.jam)
		return nil
	}

	bt.in.jamIdx = -1
	statemanager.StateDelete(bt.stateIDs["in.jamIdx"])
	statemanager.StateDelete(bt.stateIDs["in.period"])
	statemanager.StateDelete(bt.stateIDs["in.jam"])
	return nil
}

func (bt *boxTrip) setInBetweenJams(v bool) {
	bt.in.betweenJams = v
	statemanager.StateUpdateBool(bt.stateIDs["in.betweenJams"], v)
}

func (bt *boxTrip) setInAfterStarPass(v bool) {
	bt.in.afterStarPass = v
	statemanager.StateUpdateBool(bt.stateIDs["in.afterStarPass"], v)
}

func (bt *boxTrip) setOutJamIdx(v int64) {
	if v >= 0 && v < int64(len(bt.sb.jams)) {
		jam := bt.sb.jams[v]
		bt.out.jamIdx = jam.idx
		statemanager.StateUpdateInt64(bt.stateIDs["out.jamIdx"], jam.idx)
		statemanager.StateUpdateInt64(bt.stateIDs["out.period"], jam.period)
		statemanager.StateUpdateInt64(bt.stateIDs["out.jam"], jam.jam)
	}

	statemanager.StateDelete(bt.stateIDs["out.jamIdx"])
	statemanager.StateDelete(bt.stateIDs["out.period"])
	statemanager.StateDelete(bt.stateIDs["out.jam"])
}

func (bt *boxTrip) setOutBetweenJams(v bool) {
	bt.out.betweenJams = v
	statemanager.StateUpdateBool(bt.stateIDs["out.betweenJams"], v)
}

func (bt *boxTrip) setOutAfterStarPass(v bool) {
	bt.out.afterStarPass = v
	statemanager.StateUpdateBool(bt.stateIDs["out.afterStarPass"], v)
}

/* Helper functions to find the jam for RegisterUpdaters */
func (s *skater) findBoxTrip(k string) *boxTrip {
	ids := statemanager.ParseIDs(k)
	if len(ids) == 0 {
		return nil
	}
	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return nil
	}

	// Generate blank snapshots if needed
	for i := int64(len(s.boxTrips)); i <= id; i++ {
		s.boxTrips = append(s.boxTrips, blankBoxTrip(s))
	}

	return s.boxTrips[id]
}
