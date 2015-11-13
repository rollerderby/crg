package statemanager

import "testing"

var cases = []*struct {
	pattern  string
	value    string
	expected bool
	pm       patternMatcher
}{
	{"Scoreboard.Team(*)", "Scoreboard.Team(1)", true, nil},
	{"Scoreboard.Team(*)", "Scoreboard.Team(2).Name", true, nil},
	{"Scoreboard.Team(*).*", "Scoreboard.Team(1)", false, nil},
	{"Scoreboard.Team(*).*", "Scoreboard.Team(2).Name", true, nil},
	{"Scoreboard.Team(1).*", "Scoreboard.Team(1).Name", true, nil},
	{"Scoreboard.Team(1).*", "Scoreboard.Team(2).Name", false, nil},
	{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(operator).Name.Key(blue)", true, nil},
	{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(operator).Name.Key(blue).Color", true, nil},
	{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(overlay).Name.Key(blue)", false, nil},
	{"Scoreboard.Team(*).Color(operator).Name.Key(*)", "Scoreboard.Team(1).Color(overlay)", false, nil},
	{"", "Scoreboard.Team(1).Color(overlay)", true, nil},
	{"ScoreBoard", "ScoreBoard.State", true, nil},
	{"ScoreBoard", "ScoreBoard", true, nil},
}

func TestPatternMatcher(t *testing.T) {
	for _, c := range cases {
		c.pm = newPatternMatcher(c.pattern)
		r := c.pm.Matches(c.value)
		if r != c.expected {
			t.Errorf("CheckPattern('%v', '%v') expected %v got %v", c.value, c.pattern, c.expected, r)
		}
		t.Logf("ParseIDs: %v %+v", c.value, ParseIDs(c.value))
	}
}

func BenchmarkBlankPatternMatcher(b *testing.B) {
	var pm = newPatternMatcher("")
	for n := 0; n < b.N; n++ {
		pm.Matches("Scoreboard.Team(1).Color(operator).Name.Key(blue).Color")
	}
}

func BenchmarkSimplePatternMatcher(b *testing.B) {
	var pm = newPatternMatcher("Scoreboard.Team(1).Color(operator).Name")
	for n := 0; n < b.N; n++ {
		pm.Matches("Scoreboard.Team(1).Color(operator).Name.Key(blue).Color")
	}
}

func BenchmarkComplexPatternMatcher(b *testing.B) {
	var pm = newPatternMatcher("Scoreboard.Team(*).Color(*).Name.*")
	for n := 0; n < b.N; n++ {
		pm.Matches("Scoreboard.Team(1).Color(operator).Name.Key(blue).Color")
	}
}
