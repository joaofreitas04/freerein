package main

import (
	"flag"
	"io"
	"testing"
)

func newFS() (*flag.FlagSet, *string, *bool) {
	fs := flag.NewFlagSet("rein", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	registry := fs.String("registry", "", "")
	yes := fs.Bool("yes", false, "")
	return fs, registry, yes
}

// An agent writing `rein add x --registry y` must not have the flag
// silently dropped — flags parse wherever they appear.
func TestSplitArgsInterleaves(t *testing.T) {
	cases := []struct {
		args        []string
		positionals []string
		registry    string
		yes         bool
	}{
		{[]string{"comp", "--registry", "https://r/index.json"}, []string{"comp"}, "https://r/index.json", false},
		{[]string{"--registry", "./local", "comp"}, []string{"comp"}, "./local", false},
		{[]string{"a", "--yes", "b", "--registry", "r"}, []string{"a", "b"}, "r", true},
		{[]string{}, nil, "", false},
	}
	for _, c := range cases {
		fs, registry, yes := newFS()
		got, err := splitArgs(fs, c.args)
		if err != nil {
			t.Fatalf("%v: %v", c.args, err)
		}
		if len(got) != len(c.positionals) {
			t.Fatalf("%v: want positionals %v, got %v", c.args, c.positionals, got)
		}
		for i := range got {
			if got[i] != c.positionals[i] {
				t.Fatalf("%v: want positionals %v, got %v", c.args, c.positionals, got)
			}
		}
		if *registry != c.registry || *yes != c.yes {
			t.Fatalf("%v: flags dropped (registry=%q yes=%v)", c.args, *registry, *yes)
		}
	}
}

func TestSplitArgsRejectsUnknownFlag(t *testing.T) {
	fs, _, _ := newFS()
	if _, err := splitArgs(fs, []string{"comp", "--bogus"}); err == nil {
		t.Fatal("an unknown flag must error (exit 2 path)")
	}
}
