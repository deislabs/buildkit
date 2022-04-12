package gitutil

import "testing"

func TestIsGitRef(t *testing.T) {
	cases := map[string]bool{
		"https://example.com/":                 false,
		"https://example.com/foo":              false,
		"https://example.com/foo.git":          true,
		"https://example.com/foo.git#deadbeef": true,
		"https://example.com/foo.git/":         false,
		"https://example.com/foo.git.bar":      false,
		"git://example.com/foo":                true,
		"github.com/moby/buildkit":             true,
		"https://github.com/moby/buildkit":     false,
		"https://github.com/moby/buildkit.git": true,
		"git@github.com/moby/buildkit":         true,
	}
	for ref, expected := range cases {
		got := IsGitRef(ref)
		if got != expected {
			t.Errorf("expected IsGitRef(%q) to be %v, got %v", ref, expected, got)
		}
	}
}
