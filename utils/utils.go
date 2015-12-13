package utils

import "path/filepath"

var baseFilePath string

// BaseFilePath returns the base directory for loading or saving files
func BaseFilePath() string { return baseFilePath }

// SetBaseFilePath sets the base directory for loading or saving files
func SetBaseFilePath(p ...string) {
	path := filepath.Join(p...)
	baseFilePath = path
}

func Path(p ...string) string {
	ret := baseFilePath
	for _, p := range p {
		if p[0] != '.' {
			ret = filepath.Join(ret, p)
		}
	}
	return ret
}

// ParseIDs returns all values within () in the input string.
// Example
// Scoreboard.Team(1).Skater(abc123).Name returns ["1", "abc123"]
func ParseIDs(k string) []string {
	var ret []string
	startPos := -1
	for idx, c := range k {
		if startPos == -1 && c == '(' {
			startPos = idx + 1
		} else if startPos != -1 && c == ')' {
			ret = append(ret, k[startPos:idx])
			startPos = -1
		}
	}
	return ret
}
