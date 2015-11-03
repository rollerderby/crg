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
	stateJam          = "Jam"
	stateLineup       = "Lineups"
	stateTimeout      = "Timeout"
	stateIntermission = "Intermission"
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
	statemanager.RegisterCommand("UndoStartJam", sb.startJamUndo)
	statemanager.RegisterCommand("StopJam", sb.stopJam)
	statemanager.RegisterCommand("UndoStopJam", sb.stopJamUndo)
	statemanager.RegisterCommand("Timeout", sb.timeout)
	statemanager.RegisterCommand("UndoTimeout", sb.timeoutUndo)

	sb.setState(stateNotRunning)

	return sb
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
	if sb.state == stateJam {
		return nil
	}
	sb.setState(stateJam)

	// Start clocks Period, Jam
	// Reset clocks Jam
	return sb.clocks.setRunningClocks([]string{clockPeriod, clockJam}, []string{clockJam})
}

func (sb *Scoreboard) startJamUndo(_ []string) error {
	return nil
}

func (sb *Scoreboard) stopJam(_ []string) error {
	if sb.state != stateJam {
		return nil
	}
	sb.setState(stateLineup)

	// Start clocks Period, Lineup
	// Reset clocks Lineup
	return sb.clocks.setRunningClocks([]string{clockPeriod, clockLineup}, []string{clockLineup})
}

func (sb *Scoreboard) stopJamUndo(_ []string) error {
	return nil
}

func (sb *Scoreboard) timeout(_ []string) error {
	if sb.state == stateTimeout {
		return nil
	}
	sb.setState(stateTimeout)

	// Start clocks Timeout
	// Reset clocks Timeout
	return sb.clocks.setRunningClocks([]string{clockTimeout}, []string{clockTimeout})
}

func (sb *Scoreboard) timeoutUndo(_ []string) error {
	return nil
}
