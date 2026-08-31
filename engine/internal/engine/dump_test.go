package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// Dump prints the composition's fixed point (resolution rule 3) and
// offloads the full detail, announcing it with a fix (envelope rule
// 5 + rescoped rule 2).
func TestDumpPrintsTheFixedPoint(t *testing.T) {
	g, repo := appliedRepo(t)
	e := envelope.New("dump")
	g.Dump(e)
	if !e.OK {
		t.Fatalf("dump failed: %+v", e.Diagnostics)
	}
	var offload *envelope.Diagnostic
	for i := range e.Diagnostics {
		if e.Diagnostics[i].Code == "OUTPUT_OFFLOADED" {
			offload = &e.Diagnostics[i]
		}
	}
	if offload == nil || offload.Fix == "" {
		t.Fatalf("dump must announce its offload with a fix, got %+v", e.Diagnostics)
	}
	r := e.Result.(map[string]any)
	detail := r["detail"].(string)
	b, err := os.ReadFile(filepath.Join(repo, detail))
	if err != nil {
		t.Fatalf("the announced detail file must exist: %v", err)
	}
	var full struct {
		Adapter string           `json:"adapter"`
		Files   []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(b, &full); err != nil {
		t.Fatal(err)
	}
	if full.Adapter != "claude-code" || len(full.Files) == 0 {
		t.Fatalf("detail must carry the composition, got %+v", full)
	}
	// the summary in the envelope names every rendered path
	paths := map[string]bool{}
	for _, f := range r["files"].([]map[string]any) {
		paths[f["path"].(string)] = true
	}
	for _, want := range []string{"CLAUDE.md", "scripts/verify"} {
		if !paths[want] {
			t.Fatalf("summary must name %s, got %v", want, paths)
		}
	}
}

func TestAdaptersLists(t *testing.T) {
	g, _ := newRepo(t, "claude-code")
	e := envelope.New("adapters")
	g.Adapters(e)
	if !e.OK {
		t.Fatalf("adapters failed: %+v", e.Diagnostics)
	}
	b, _ := json.Marshal(e.Result)
	for _, want := range []string{"claude-code", "codex"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("adapters must list %s, got %s", want, b)
		}
	}
}
