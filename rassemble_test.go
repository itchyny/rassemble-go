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
			name:     "empty literals",
			patterns: []string{"", ""},
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
			patterns: []string{"abcd", "abcf", "abc", "abce", "abcgh", "abdc"},
			expected: "ab(?:c(?:[d-f]|gh)?|dc)",
		},
		{
			name:     "same prefixes with same length",
			patterns: []string{"abcde", "abcfg", "abcgh"},
			expected: "abc(?:de|fg|gh)",
		},
		{
			name:     "same prefixes in increasing length order",
			patterns: []string{"a", "ab", "abc", "abcd"},
			expected: "a(?:b(?:cd?)?)?",
		},
		{
			name:     "same prefixes in decreasing length order",
			patterns: []string{"abcd", "abc", "ab", "a"},
			expected: "a(?:b(?:cd?)?)?",
		},
		{
			name:     "same prefix and suffix",
			patterns: []string{"abcdefg", "abcfg", "abefg", "befg", "beefg"},
			expected: "(?:ab(?:c(?:de)?|e)|bee?)fg",
		},
		{
			name:     "same prefix and suffix with double quests",
			patterns: []string{"abcd", "abd", "acd", "ad"},
			expected: "ab?c?d",
		},
		{
			name:     "same prefix and suffix with triple quests",
			patterns: []string{"abcde", "acde", "abde", "abce", "abe", "ace", "ade", "ae"},
			expected: "ab?c?d?e",
		},
		{
			name:     "multiple prefix groups",
			patterns: []string{"abc", "ab", "abcd", "a", "bcd", "bcdef", "cdef", "cdeh"},
			expected: "a(?:b(?:cd?)?)?|bcd(?:ef)?|cde[fh]",
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
			name:     "merge literal to alternate in quest",
			patterns: []string{"abc(?:de|fh)?", "abcff", "abcf", "abchh"},
			expected: "abc(?:de|f[fh]?|hh)?",
		},
		{
			name:     "merge literal to quest with suffix",
			patterns: []string{"abc(?:def)?ghi", "abcd"},
			expected: "abc(?:(?:def)?ghi|d)",
		},
		{
			name:     "merge literal to alternate with same prefix",
			patterns: []string{"abcfd|def", "abcdef", "abcfe"},
			expected: "abc(?:f[de]|def)|def",
		},
		{
			name:     "merge literal to alternate with different prefix",
			patterns: []string{"abc|def", "ghi"},
			expected: "abc|def|ghi",
		},
		{
			name:     "character class",
			patterns: []string{"a", "1", "z", "2"},
			expected: "[12az]",
		},
		{
			name:     "character class with prefix",
			patterns: []string{"aa", "ab"},
			expected: "a[ab]",
		},
		{
			name:     "character class in prefix",
			patterns: []string{"abcde", "abc", "bbcde", "bbc", "cbcde", "cbc"},
			expected: "[a-c]bc(?:de)?",
		},
		{
			name:     "add character class to a character",
			patterns: []string{"d?", "[a-c]", "e"},
			expected: "[a-e]?",
		},
		{
			name:     "unmerge character class",
			patterns: []string{"a", "c", "e", "ab", "cd", "ef"},
			expected: "ab?|cd?|ef?",
		},
		{
			name:     "successive character class",
			patterns: []string{"aa", "ab", "ac"},
			expected: "a[a-c]",
		},
		{
			name:     "successive character class in random order",
			patterns: []string{"ac", "aa", "ae", "ab", "ad"},
			expected: "a[a-e]",
		},
		{
			name:     "numbers",
			patterns: []string{"1", "9", "2", "6", "3"},
			expected: "[1-369]",
		},
		{
			name:     "numbers 0 to 5",
			patterns: []string{"0", "4", "3", "5", "1", "2"},
			expected: "[0-5]",
		},
		{
			name:     "numbers 0 to 10",
			patterns: []string{"1", "9", "2", "6", "3", "7", "10", "8", "0", "5", "4"},
			expected: "[0-9]|10",
		},
		{
			name:     "numbers with prefix",
			patterns: []string{"a2", "a1", "a0", "a8", "a3", "a5", "a6", "a4", "a7", "a2", "a9", "a0", "a10"},
			expected: "a(?:[0-9]|10)",
		},
		{
			name:     "add empty literal to quest",
			patterns: []string{"abc", "", ""},
			expected: "(?:abc)?",
		},
		{
			name:     "add empty literal to plus and star",
			patterns: []string{"(?:abc)+", "", ""},
			expected: "(?:abc)*",
		},
		{
			name:     "add empty literal to character class",
			patterns: []string{"[135]", "", "7"},
			expected: "[1357]?",
		},
		{
			name:     "add empty literal to alternate",
			patterns: []string{"a", "[bcd]", "c+", ""},
			expected: "[a-d]|c*",
		},
		{
			name:     "add literal to empty literal",
			patterns: []string{"", "abc", ""},
			expected: "(?:abc)?",
		},
		{
			name:     "add quest to empty literal",
			patterns: []string{"", "(?:abc)?"},
			expected: "(?:abc)?",
		},
		{
			name:     "add star to empty literal",
			patterns: []string{"", "(?:abc)*"},
			expected: "(?:abc)*",
		},
		{
			name:     "add plus to empty literal",
			patterns: []string{"", "(?:abc)+"},
			expected: "(?:abc)*",
		},
		{
			name:     "add character class to empty literal",
			patterns: []string{"", "[a-c]"},
			expected: "[a-c]?",
		},
		{
			name:     "add alternate",
			patterns: []string{"a", "[a-c]|bb", "cc|d"},
			expected: "[a-d]|bb|cc",
		},
		{
			name:     "merge suffix",
			patterns: []string{"abcde", "cde", "bde"},
			expected: "(?:(?:ab)?c|b)de",
		},
		{
			name:     "merge suffix in increasing length order",
			patterns: []string{"e", "de", "cde", "bcde", "abcde"},
			expected: "(?:(?:(?:a?b)?c)?d)?e",
		},
		{
			name:     "merge suffix in decreasing length order",
			patterns: []string{"abcde", "bcde", "cde", "de", "e"},
			expected: "(?:(?:(?:a?b)?c)?d)?e",
		},
		{
			name:     "regexps matching head",
			patterns: []string{"a?", "a?b*c+"},
			expected: "a?(?:b*c+)?",
		},
		{
			name:     "regexps with same prefix",
			patterns: []string{"a?b+cd", "a?b+c*", "a?b*c+"},
			expected: "a?(?:b+(?:cd|c*)|b*c+)",
		},
		{
			name:     "regexps with same prefixes",
			patterns: []string{"a?b+c*", "a?b+c*d*", "a?b+", "a?"},
			expected: "a?(?:b+(?:c*d*)?)?",
		},
		{
			name:     "regexps with same literal prefix",
			patterns: []string{"abcd*e*", "abcde*f*", "abefg?", "ab"},
			expected: "ab(?:c(?:d*e*|de*f*)|efg?)?",
		},
		{
			name:     "regexps with same suffix",
			patterns: []string{"ab*c", "c+", "bab?c", "a+c", "cbc+", "dbc+", "ab*c", "c*d+", "d+"},
			expected: "(?:ab*|bab?|a+)c|(?:[cd]b)?c+|c*d+",
		},
		{
			name:     "regexps with same literal suffix",
			patterns: []string{"ab*cde", "bcde", "a*de", "cde"},
			expected: "(?:(?:ab*|b)?c|a*)de",
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
	if _, err := Join([]string{"*"}); err == nil {
		t.Fatalf("expected an error")
	}
}
