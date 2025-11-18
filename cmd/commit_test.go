package cmd

import "testing"

func TestNormalizeCommitMessage(t *testing.T) {
	input := "feat(app): ✨ add\n\nextra details"
	got := normalizeCommitMessage(input)
	want := "feat(app): ✨ add"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestValidateCommitFormat(t *testing.T) {
	valid := "fix(cli): 🐛 prevent crash"
	if err := validateCommitFormat(valid); err != nil {
		t.Fatalf("expected message to be valid, got error %v", err)
	}

	cases := map[string]string{
		"missing type":        "(cli): ✨ nope",
		"no scope":            "feat: ✨ nope",
		"no emoji":            "feat(cli): add stuff",
		"bad type":            "unknown(cli): ✅ yep",
		"empty":               "",
		"missing colon":       "feat(cli) ✨ foo",
		"description missing": "feat(cli): ✨ ",
	}

	for name, msg := range cases {
		if err := validateCommitFormat(msg); err == nil {
			t.Fatalf("expected %s case to fail", name)
		}
	}
}
