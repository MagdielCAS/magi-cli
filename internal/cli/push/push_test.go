package push

import "testing"

func TestBuildPushArgs(t *testing.T) {
	cases := []struct {
		name        string
		branch      string
		remote      string
		hasUpstream bool
		expected    []string
	}{
		{
			name:        "has upstream",
			branch:      "main",
			remote:      "origin",
			hasUpstream: true,
			expected:    []string{"push"},
		},
		{
			name:        "missing upstream",
			branch:      "feature",
			remote:      "origin",
			hasUpstream: false,
			expected:    []string{"push", "--set-upstream", "origin", "feature"},
		},
	}

	for _, tc := range cases {
		got := buildPushArgs(tc.branch, tc.remote, tc.hasUpstream)
		if len(got) != len(tc.expected) {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Fatalf("%s: expected %v, got %v", tc.name, tc.expected, got)
			}
		}
	}
}
