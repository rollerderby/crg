// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

package statemanager

import "strings"

type patternMatcher interface {
	Matches(string) bool
	Pattern() string
}

type blankMatcher struct {
	pattern string
}
type simpleMatcher struct {
	pattern    string
	patternDot string
}
type complexMatcher struct {
	pattern string
}

// newPatternMatch will return a PatternMatcher capable of matching
// patterns to values
//
// Sample patterns
// "" will match anything
// key(*) will match anything inside the parens
// key.* will match anything starting with "key."
// key(*).* will match the combination of the two
func newPatternMatcher(pattern string) patternMatcher {
	if pattern == "" {
		return &blankMatcher{pattern: pattern}
	}

	if strings.Index(pattern, "(*)") == -1 && !strings.HasSuffix(pattern, ".*") {
		return &simpleMatcher{pattern: pattern, patternDot: pattern + "."}
	}
	return &complexMatcher{pattern: pattern}
}

func (bm *blankMatcher) Matches(string) bool { return true }
func (bm *blankMatcher) Pattern() string     { return bm.pattern }

func (sm *simpleMatcher) Matches(value string) bool {
	return value == sm.pattern || strings.HasPrefix(value, sm.patternDot)
}
func (sm *simpleMatcher) Pattern() string { return sm.pattern }

func (cm *complexMatcher) Pattern() string { return cm.pattern }
func (cm *complexMatcher) Matches(value string) bool {
	pattern := cm.pattern
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
