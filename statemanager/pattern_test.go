package statemanager

import "testing"

func TestPatternMatch(t *testing.T) {
	cases := []struct {
		pattern  string
		value    string
		expected bool
	}{
		{"Scoreboard.Team(*)", "Scoreboard.Team(1)", true},
		{"Scoreboard.Team(*)", "Scoreboard.Team(2).Name", true},
		{"Scoreboard.Team(*).*", "Scoreboard.Team(1)", false},
		{"Scoreboard.Team(*).*", "Scoreboard.Team(2).Name", true},
		{"Scoreboard.Team(1).*", "Scoreboard.Team(1).Name", true},
		{"Scoreboard.Team(1).*", "Scoreboard.Team(2).Name", false},
		{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(operator).Name.Key(blue)", true},
		{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(operator).Name.Key(blue).Color", true},
		{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(overlay).Name.Key(blue)", false},
		{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(overlay)", false},
		{"", "Scoreboard.Team(1).Color(overlay)", true},
		{"ScoreBoard", "ScoreBoard.State", true},
		{"ScoreBoard", "ScoreBoard", true},
	}
	for _, c := range cases {
		r := PatternMatch(c.value, c.pattern)
		if r != c.expected {
			t.Errorf("CheckPattern('%v', '%v') expected %v got %v", c.value, c.pattern, c.expected, r)
		}
	}
}
