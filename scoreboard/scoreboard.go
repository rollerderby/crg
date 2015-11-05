package scoreboard

import (
	"log"

	"github.com/rollerderby/crg/statemanager"
)

// Scoreboard is the realtime scoreboard controller.  All aspects of the
// live state for the scoreboard are contained within and exported
// via github.com/rollerderby/crg/statemanager
type Scoreboard struct {
	stateIDs map[string]string
	teams    []*team
	clocks   *masterClock
	state    string
}

const (
	stateNotRunning   = ""
	statePreGame      = "PreGame"
	stateJam          = "Jam"
	stateLineup       = "Lineups"
	stateTimeout      = "Timeout"
	stateIntermission = "Intermission"
	stateUnofficial   = "UnofficialFinal"
	stateFinal        = "Final"
)

type parent interface {
	stateBase() string
}

// New initialized a default state for the scoreboard.  Additional setup
// of the scoreboard is required from either a saved state or via
// the web interface.  Returns a *Scoreboard
func New() *Scoreboard {
	statemanager.Lock()
	defer statemanager.Unlock()

	sb := &Scoreboard{}
	sb.teams = append(sb.teams, newTeam(sb, 1), newTeam(sb, 2))
	sb.clocks = newMasterClock(sb)

	sb.stateIDs = make(map[string]string)
	sb.stateIDs["state"] = sb.stateBase() + ".State"

	statemanager.RegisterUpdater(sb.stateIDs["state"], 0, sb.setState)

	statemanager.RegisterCommand("StartJam", sb.startJam)
	statemanager.RegisterCommand("StopJam", sb.stopJam)
	statemanager.RegisterCommand("Timeout", sb.timeout)
	statemanager.RegisterCommand("UndoStateChange", sb.undoStateChange)

	sb.setState(stateNotRunning)

	return sb
}

func (sb *Scoreboard) snapshotStateStart() {
}

func (sb *Scoreboard) snapshotStateEnd(canUndo bool) {
}

func (sb *Scoreboard) clocksExpired() {
	switch sb.state {
	case stateLineup:
		if !sb.clocks.period.running {
			// Period clock ended, go to intermission or unofficial
			sb.endOfPeriod(false)
		} else {
			// Lineup expired, start jam!
			sb.startJam(nil)
		}
	case stateJam:
		if !sb.clocks.jam.running {
			if !sb.clocks.period.running {
				// Period clock is out, go to intermission or unofficial
				sb.endOfPeriod(false)
				return
			}
			sb.stopJam(nil)
		}
	case stateIntermission:
		if sb.clocks.intermission.number.num == 1 {
			sb.endOfIntermission()
		}
	}
}

func (sb *Scoreboard) endOfPeriod(canUndo bool) {
	sb.snapshotStateEnd(canUndo)
	defer sb.snapshotStateStart()
	if sb.clocks.period.number.num == 1 {
		sb.setState(stateIntermission)

		// Reset & start intermission clock
		sb.clocks.intermission.reset(false, false)
		sb.clocks.setRunningClocks(clockIntermission)
	} else {
		sb.setState(stateUnofficial)
	}
}

func (sb *Scoreboard) endOfIntermission() {
	sb.clocks.period.reset(false, true)
	sb.clocks.jam.reset(true, false)
}

func (sb *Scoreboard) stateBase() string {
	return "ScoreBoard"
}

func (sb *Scoreboard) setState(state string) error {
	log.Printf("scoreboard: setState(%+v)", state)
	sb.state = state
	statemanager.StateUpdate(sb.stateIDs["state"], state)

	adjustable := false
	if state == stateTimeout {
		adjustable = true
	}
	sb.clocks.setClockAdjustable(clockPeriod, adjustable)
	return nil
}

func (sb *Scoreboard) startJam(_ []string) error {
	switch sb.state {
	case stateJam:
		return nil
	case stateIntermission:
		sb.endOfIntermission()
	}

	sb.snapshotStateEnd(true)
	defer sb.snapshotStateStart()
	sb.setState(stateJam)

	// Reset jam clock and increment jam number
	sb.clocks.jam.reset(false, true)
	// Start clocks Period, Jam
	sb.clocks.setRunningClocks(clockPeriod, clockJam)
	return nil
}

func (sb *Scoreboard) stopJam(_ []string) error {
	if sb.state != stateJam {
		return nil
	}

	if !sb.clocks.period.running {
		// Period clock is out, go to intermission or unofficial
		sb.endOfPeriod(true)
		return nil
	}

	// Not the end of a period, start lineups
	sb.snapshotStateEnd(sb.clocks.jam.time.num == sb.clocks.jam.time.min)
	defer sb.snapshotStateStart()
	sb.setState(stateLineup)

	// Reset lineup clock
	sb.clocks.lineup.reset(false, false)
	// Start clocks Period, Lineup
	sb.clocks.setRunningClocks(clockPeriod, clockLineup)
	return nil
}

func (sb *Scoreboard) timeout(_ []string) error {
	if sb.state == stateTimeout {
		return nil
	}
	sb.snapshotStateEnd(true)
	defer sb.snapshotStateStart()
	sb.setState(stateTimeout)

	// Reset timeout clock
	sb.clocks.timeout.reset(false, false)
	// Start clocks Timeout
	sb.clocks.setRunningClocks(clockTimeout)
	return nil
}

func (sb *Scoreboard) undoStateChange(_ []string) error {
	log.Print("Scoreboard.undoStateChange: NOT IMPLEMENTED")
	return nil
}
