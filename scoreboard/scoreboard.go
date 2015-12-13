package scoreboard

import "github.com/rollerderby/crg/state"

// Initialize creates the basic structure for the scoreboard, clearing the old state first and making a new, default state
func Initialize() {
	state.Delete("Scoreboard")
}
