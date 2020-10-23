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
