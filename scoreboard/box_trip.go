// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package scoreboard

import (
	"fmt"

	"github.com/rollerderby/crg/statemanager"
)

type boxTrip struct {
	t        *team
	s        *skater
	in       boxTripTime
	out      boxTripTime
	stateIDs map[string]string
}

type boxTripTime struct {
	period        int64
	jam           int64
	betweenJams   bool
	afterStarPass bool
}

func blankBoxTrip(t *team, idx int64) *boxTrip {
	bt := &boxTrip{stateIDs: make(map[string]string)}
	base := fmt.Sprintf("%v.BoxTrip(%v)", t.base, idx)
	bt.stateIDs["skater"] = base + ".Skater"
	bt.stateIDs["in.period"] = base + ".In.Period"
	bt.stateIDs["in.jam"] = base + ".In.Jam"
	bt.stateIDs["in.betweenJams"] = base + ".In.BetweenJams"
	bt.stateIDs["in.afterStarPass"] = base + ".In.AfterStarPass"
	bt.stateIDs["out.period"] = base + ".Out.Period"
	bt.stateIDs["out.jam"] = base + ".Out.Jam"
	bt.stateIDs["out.betweenJams"] = base + ".Out.BetweenJams"
	bt.stateIDs["out.afterStarPass"] = base + ".Out.AfterStarPass"

	return bt
}

func newBoxTrip(t *team, idx int64, s *skater, period, jam int64, betweenJams, afterStarPass bool) *boxTrip {
	bt := blankBoxTrip(t, idx)

	bt.setSkater(s)
	bt.setInPeriod(period)
	bt.setInJam(jam)
	bt.setInBetweenJams(betweenJams)
	bt.setInAfterStarPass(afterStarPass)
	bt.setOutPeriod(0)
	bt.setOutJam(0)
	bt.setOutBetweenJams(false)
	bt.setOutAfterStarPass(false)

	return bt
}

func (bt *boxTrip) setSkater(s *skater) {
	bt.s = s
	statemanager.StateUpdate(bt.stateIDs["skater"], s.id)
}

func (bt *boxTrip) setInPeriod(v int64) {
	bt.in.period = v
	statemanager.StateUpdate(bt.stateIDs["in.period"], v)
}

func (bt *boxTrip) setInJam(v int64) {
	bt.in.jam = v
	statemanager.StateUpdate(bt.stateIDs["in.jam"], v)
}

func (bt *boxTrip) setInBetweenJams(v bool) {
	bt.in.betweenJams = v
	statemanager.StateUpdate(bt.stateIDs["in.betweenJams"], v)
}

func (bt *boxTrip) setInAfterStarPass(v bool) {
	bt.in.afterStarPass = v
	statemanager.StateUpdate(bt.stateIDs["in.afterStarPass"], v)
}

func (bt *boxTrip) setOutPeriod(v int64) {
	bt.out.period = v
	statemanager.StateUpdate(bt.stateIDs["out.period"], v)
}

func (bt *boxTrip) setOutJam(v int64) {
	bt.out.jam = v
	statemanager.StateUpdate(bt.stateIDs["out.jam"], v)
}

func (bt *boxTrip) setOutBetweenJams(v bool) {
	bt.out.betweenJams = v
	statemanager.StateUpdate(bt.stateIDs["out.betweenJams"], v)
}

func (bt *boxTrip) setOutAfterStarPass(v bool) {
	bt.out.afterStarPass = v
	statemanager.StateUpdate(bt.stateIDs["out.afterStarPass"], v)
}
