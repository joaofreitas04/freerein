package publish_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/publish"
	"github.com/joaofreitas04/freerein/engine/internal/registry"
)

func writeToy(t *testing.T, dir, fragment string) {
	t.Helper()
	manifest := `name: toy-ext
kind: extension
version: 1.0.0
subsystem: instructions
rung: instruction
rent:
  class: amplifier
provides:
  - AGENTS.md.d/60-toy.md
description: a published toy
`
	if err := os.MkdirAll(filepath.Join(dir, "AGENTS.md.d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "component.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md.d/60-toy.md"), []byte(fragment), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPublishDeterminismAndRefusal(t *testing.T) {
	work := t.TempDir()
	comp := filepath.Join(work, "toy-ext")
	reg := filepath.Join(work, "reg")
	writeToy(t, comp, "## Toy\n\nv1 rule\n")

	if err := publish.Run(reg, []string{comp}); err != nil {
		t.Fatal(err)
	}
	idx1, err := registry.LoadIndex(filepath.Join(reg, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	rel1, err := idx1.Lookup("toy-ext", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	// determinism: republishing identical content keeps the sha
	if err := publish.Run(reg, []string{comp}); err != nil {
		t.Fatalf("idempotent republish must succeed: %v", err)
	}
	idx2, _ := registry.LoadIndex(filepath.Join(reg, "index.json"))
	rel2, _ := idx2.Lookup("toy-ext", "1.0.0")
	if rel1.Sha256 != rel2.Sha256 {
		t.Fatal("republishing identical content must keep the sha stable")
	}
	// changed content without a version bump: refused, and the
	// published archive must stay fetchable under its old sha
	writeToy(t, comp, "## Toy\n\nCHANGED rule\n")
	if err := publish.Run(reg, []string{comp}); err == nil {
		t.Fatal("republish with changed content must be refused")
	}
	dest := t.TempDir()
	if err := idx1.Fetch(rel1, dest); err != nil {
		t.Fatalf("archive on disk must still match the index pin after a refused publish: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "AGENTS.md.d/60-toy.md")); err != nil || string(b) != "## Toy\n\nv1 rule\n" {
		t.Fatal("fetched content must be the originally published version")
	}
	// a bumped version publishes alongside the old one
	writeToy(t, comp, "## Toy\n\nCHANGED rule\n")
	b, _ := os.ReadFile(filepath.Join(comp, "component.yaml"))
	_ = os.WriteFile(filepath.Join(comp, "component.yaml"),
		[]byte(string(b[:0])+replaceVersion(string(b), "1.0.0", "1.1.0")), 0o644)
	if err := publish.Run(reg, []string{comp}); err != nil {
		t.Fatalf("bumped version must publish: %v", err)
	}
	idx3, _ := registry.LoadIndex(filepath.Join(reg, "index.json"))
	if v, _, _ := idx3.Latest("toy-ext"); v != "1.1.0" {
		t.Fatalf("latest must be 1.1.0, got %s", v)
	}
}

func replaceVersion(s, from, to string) string {
	out := ""
	for _, line := range splitLines(s) {
		if line == "version: "+from {
			line = "version: " + to
		}
		out += line + "\n"
	}
	return out
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
