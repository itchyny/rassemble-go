package rassemble

import "testing"

func TestJoin(t *testing.T) {
	testCases := []struct {
		name     string
		patterns []string
		expected string
	}{
		{
			name:     "empty",
			patterns: []string{},
			expected: "",
		},
		{
			name:     "empty literal",
			patterns: []string{""},
			expected: "(?:)",
		},
		{
			name:     "single literal",
			patterns: []string{"abc"},
			expected: "abc",
		},
		{
			name:     "multiple literals",
			patterns: []string{"abc", "def", "ghi"},
			expected: "abc|def|ghi",
		},
		{
			name:     "same literals",
			patterns: []string{"abc", "def", "abc", "def"},
			expected: "abc|def",
		},
		{
			name:     "same prefixes with different length",
			patterns: []string{"abcd", "abcf", "abdc"},
			expected: "ab(?:c(?:d|f)|dc)",
		},
		{
			name:     "same prefixes with same length",
			patterns: []string{"abcde", "abcfg", "abcgh"},
			expected: "abc(?:de|fg|gh)",
		},
		{
			name:     "same prefixes in increasing order",
			patterns: []string{"a", "ab", "abc", "abcd"},
			expected: "a(?:b?|bcd?)",
		},
		{
			name:     "same prefixes in decreasing order",
			patterns: []string{"abcd", "abc", "ab", "a"},
			expected: "a(?:b(?:cd?)?)?",
		},
		{
			name:     "multiple prefix groups",
			patterns: []string{"abc", "ab", "abcd", "a", "bcd", "bcdef", "cdef", "cdeh"},
			expected: "a(?:b(?:c?|cd))?|bcd(?:ef)?|cde(?:f|h)",
		},
		{
			name:     "merge literal to quest",
			patterns: []string{"abc(?:def)?", "abc"},
			expected: "abc(?:def)?",
		},
		{
			name:     "merge literal to star",
			patterns: []string{"abc(?:def)*", "abc"},
			expected: "abc(?:def)*",
		},
		{
			name:     "merge literal to plus",
			patterns: []string{"abc(?:def)+", "abc"},
			expected: "abc(?:def)*",
		},
		{
			name:     "merge literal to alternate",
			patterns: []string{"abc(?:de|f)", "abc"},
			expected: "abc(?:de|f)?",
		},
		{
			name:     "merge literal to concat",
			patterns: []string{"abca*b*", "abc"},
			expected: "abc(?:a*b*)?",
		},
		{
			name:     "merge literal to concat",
			patterns: []string{"abca*b*", "abcde"},
			expected: "abc(?:a*b*|de)",
		},
		{
			name:     "merge literal to quest with suffix",
			patterns: []string{"abc(?:def)?ghi", "abcd"},
			expected: "abc(?:(?:def)?ghi|d)",
		},
		{
			name:     "numbers",
			patterns: []string{"a0", "a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10"},
			expected: "a(?:0|10?|2|3|4|5|6|7|8|9)",
		},
		{
			name:     "regexps",
			patterns: []string{"ab*c", "abc+", "bc+"},
			expected: "ab*c|abc+|bc+",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Join(tc.patterns)
			if err != nil {
				t.Fatalf("got an error: %s", err)
			}
			if got != tc.expected {
				t.Errorf("expected: %s, got: %s", tc.expected, got)
			}
		})
	}
}
