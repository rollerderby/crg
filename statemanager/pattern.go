package statemanager

import "strings"

// PatternMatch will check value to see if the pattern matches
//
// Sample patterns
// key(*) will match anything inside the parens
// key.* will match anything starting with "key."
// key(*).* will match the combination of the two
func PatternMatch(value, pattern string) bool {
	// Special case, if pattern is empty, it matches
	if pattern == "" {
		return true
	}
	for {
		id := strings.Index(pattern, "(*)")
		if id == -1 {
			break
		}
		id = id + 1
		// check if value is long enough
		if len(value) < id {
			return false
		}

		// check everything leading up to the open paren
		if value[:id] != pattern[:id] {
			return false
		}

		// prefix matched, now look for the close paren
		rparen := strings.IndexRune(value[id:], ')')
		if rparen == -1 {
			return false
		}

		value = value[id+rparen+1:]
		pattern = pattern[id+2:]
	}

	if value == pattern {
		return true
	}
	if strings.HasPrefix(value, pattern+".") {
		return true
	}

	// match if pattern is empty and value is empty or starts with a dot
	if len(pattern) == 0 {
		return len(value) == 0 || value[0] == '.'
	}

	// look for trailing *
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(value, pattern[:len(pattern)-1])
	}

	return value == pattern
}
