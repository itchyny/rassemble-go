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
			name:     "single literal with flag",
			patterns: []string{"(?i:abc)"},
			expected: "(?i:ABC)",
		},
		{
			name:     "single literal with multiple flags",
			patterns: []string{"(?ims:^a.b.c$)"},
			expected: "(?ims:^A.B.C$)",
		},
		{
			name:     "multiple literals",
			patterns: []string{"abc", "def", "ghi"},
			expected: "abc|def|ghi",
		},
		{
			name:     "multiple literals with same flag",
			patterns: []string{"(?i:abc)", "(?i:def)", "(?i:ghi)"},
			expected: "(?i:ABC|DEF|GHI)",
		},
		{
			name:     "multiple literals with different flags",
			patterns: []string{"(?i:abc)", "(?m:d.$)", "(?s:^g.i)"},
			expected: "(?-s:(?i:ABC)|(?m:d.$)|(?ms:^g.i))",
		},
		{
			name:     "multiple characters with different flags",
			patterns: []string{"a", "b", "(?i:c)?", "d"},
			expected: "[Ca-d]?",
		},
		{
			name:     "same literals",
			patterns: []string{"abc", "def", "abc", "def"},
			expected: "abc|def",
		},
		{
			name:     "same prefixes with different length",
			patterns: []string{"abc", "ab", "ad", "a"},
			expected: "a(?:bc?|d)?",
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
			name:     "same prefixes with same flag",
			patterns: []string{"(?i:abc)", "(?i:ab)", "(?i:ad)", "(?i:a)"},
			expected: "(?i:A(?:BC?|D)?)",
		},
		{
			name:     "same prefixes with various flags",
			patterns: []string{"(?i:abc)", "(?:a.*b$)", "(?im:ad$|ae)", "(?sm:a.$)", "(?U:a.*c$)"},
			expected: "(?-s:(?im:A(?:BC|D$|E))|(?m:a(?:.*b|(?s:.)|.*?c)$))",
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
			patterns: []string{"abc(?:def)*", "abcdef", "abc"},
			expected: "abc(?:def)*",
		},
		{
			name:     "merge literal to plus",
			patterns: []string{"abc(?:def)+", "abcdef", "abc"},
			expected: "abc(?:def)*",
		},
		{
			name:     "merge literal to alternate",
			patterns: []string{"abc(?:de|f)", "abc"},
			expected: "abc(?:de|f)?",
		},
		{
			name:     "merge literal to alternate with plus",
			patterns: []string{"abc(?:de|[f-h]+)", "abc", "abc"},
			expected: "abc(?:de|[f-h]*)",
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
			name:     "add character class to a character class",
			patterns: []string{"[a-c]", "[e-f]", "d"},
			expected: "[a-f]",
		},
		{
			name:     "add complex character class to a complex character class",
			patterns: []string{"[i-kea-c]", "[f-hd]"},
			expected: "[a-k]",
		},
		{
			name:     "add character to character class negation",
			patterns: []string{"[^0-9]", "3", "5"},
			expected: "[^0-246-9]",
		},
		{
			name:     "add quest of character to character class negation",
			patterns: []string{"[^0-9]", "3?", "5"},
			expected: "[^0-246-9]?",
		},
		{
			name:     "add character to character class negation to match anything",
			patterns: []string{"[^0]", "0"},
			expected: "(?s:.)",
		},
		{
			name:     "merge literal prefix rather than character class",
			patterns: []string{"a", "c", "e", "ab", "cd", "ef"},
			expected: "ab?|cd?|ef?",
		},
		{
			name:     "add character class to a character class negation",
			patterns: []string{"[a-d]", "[^c-g]", "f"},
			expected: "[^eg]",
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
			name:     "add case insensitive literal to a literal",
			patterns: []string{"a", "(?i:b)", "c"},
			expected: "[Ba-c]",
		},
		{
			name:     "add case insensitive literal to a character class",
			patterns: []string{"(?i:a)", "[b-e]", "(?i:f)"},
			expected: "[AFa-f]",
		},
		{
			name:     "add case insensitive character class to a character class",
			patterns: []string{"[a-c]", "(?i:[d-f])", "[g-i]"},
			expected: "[D-Fa-i]",
		},
		{
			name:     "add case insensitive literal to a quest of character class",
			patterns: []string{"(?i:a)", "[b-e]?", "(?i:f)"},
			expected: "[AFa-f]?",
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
			expected: "10?|[02-9]",
		},
		{
			name:     "numbers with prefix",
			patterns: []string{"a2", "a1", "a0", "a8", "a3", "a5", "a6", "a4", "a7", "a11", "a2", "a9", "a0", "a10"},
			expected: "a(?:[02-9]|1[01]?)",
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
			name:     "add empty literal to alternate with quest",
			patterns: []string{"abc", "b", "", ""},
			expected: "abc|b?",
		},
		{
			name:     "add empty literal to alternate with plus and star",
			patterns: []string{"abc", "b+", "", ""},
			expected: "abc|b*",
		},
		{
			name:     "add literal to empty literal",
			patterns: []string{"", "abc", ""},
			expected: "(?:abc)?",
		},
		{
			name:     "add literal with a flag to empty literal",
			patterns: []string{"", "(?i:abc)", ""},
			expected: "(?i:(?:ABC)?)",
		},
		{
			name:     "add quest to empty literal",
			patterns: []string{"", "(?:abc)?"},
			expected: "(?:abc)?",
		},
		{
			name:     "add quest to quest",
			patterns: []string{"(?:abc)?", "(?:abc)?"},
			expected: "(?:abc)?",
		},
		{
			name:     "add star to empty literal",
			patterns: []string{"", "(?:abc)*"},
			expected: "(?:abc)*",
		},
		{
			name:     "add star and plus to character class",
			patterns: []string{"a[a-c]c", "a[a-c]*c", "a[a-c]+c", "a[a-d]+c"},
			expected: "a(?:[a-c]*|[a-d]+)c",
		},
		{
			name:     "add quest and plus to character class",
			patterns: []string{"a[a-c]c", "aac", "a[a-c]?c", "a[a-c]+c", "a[a-d]c"},
			expected: "a(?:[a-c]*|[a-d])c",
		},
		{
			name:     "add quest of character class to literal",
			patterns: []string{"abc", "a[a-c]?c"},
			expected: "a[a-c]?c",
		},
		{
			name:     "add quest of character class to character class",
			patterns: []string{"abc", "adc", "a[a-c]?c"},
			expected: "a[a-d]?c",
		},
		{
			name:     "add plus to empty literal",
			patterns: []string{"", "(?:abc)+"},
			expected: "(?:abc)*",
		},
		{
			name:     "add plus to literal",
			patterns: []string{"abc", "(?:ab)+", "(?:abc)+"},
			expected: "(?:abc)+|(?:ab)+",
		},
		{
			name:     "add plus to quest",
			patterns: []string{"(?:abc)?", "(?:ab)+", "(?:abc)+"},
			expected: "(?:abc)*|(?:ab)+",
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
			name:     "regexps with same prefixes and flags",
			patterns: []string{"(?i:a*b+c*)", "(?i:a*b+(?-i:c*d*))", "(?i:a*)(?i:b+)", "a*", "A*"},
			expected: "A*(?:(?i:B+)(?:(?i:C*)|c*d*))?|a*", // bug in regexp/syntax (golang/go#59007)
		},
		{
			name:     "regexps with same prefixes and different flags",
			patterns: []string{"a?(?i:b+c*)", "(?i:a?)(?i:b+c*d*)", "(?i:a?)b+", "a?"},
			expected: "a?(?i:(?:B+C*)?)|(?i:A?)(?:(?i:B+C*D*)|b+)",
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
