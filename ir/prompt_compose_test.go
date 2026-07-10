package ir

import "testing"

func TestComposePrompt(t *testing.T) {
	cases := []struct {
		name                          string
		prefix, body, include, suffix string
		want                          string
	}{
		{"body only", "", "do X", "", "", "do X"},
		{"prefix+body+suffix", "P", "B", "", "S", "P\n\nB\n\nS"},
		{"include after body before suffix", "", "B", "I", "S", "B\n\nI\n\nS"},
		{"empty parts no blank lines", "", "B", "", "", "B"},
		{"suffix always last", "P", "B", "I", "S", "P\n\nB\n\nI\n\nS"},
		{"all empty", "", "", "", "", ""},
		{"prefix only", "P", "", "", "", "P"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ComposePrompt(c.prefix, c.body, c.include, c.suffix); got != c.want {
				t.Errorf("ComposePrompt = %q, want %q", got, c.want)
			}
		})
	}
}
