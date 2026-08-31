package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// spec/eval.md: the engine never runs a task — it assigns, records,
// and keeps score. These tests hold the contract's mechanical half:
// deterministic counterbalanced assignment, conditions read from the
// installed harness, append-only runs with void-by-entry, and
// aggregates that carry their misreadings.

func evalTasks(t *testing.T, repo, lines string) {
	t.Helper()
	seedFile(t, repo, engine.EvalTasksPath, lines)
}

func switchProfile(t *testing.T, g *engine.Engine, repo, profile string) {
	t.Helper()
	cfgPath := filepath.Join(repo, "harness.yaml")
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(cfg), "adapter: claude-code\n",
		"adapter: claude-code\nprofile: "+profile+"\n", 1)
	if err := os.WriteFile(cfgPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	apply(t, g)
}

// Assignment is deterministic and counterbalanced: the first arm
// alternates per task, and the operator asks — never picks.
func TestEvalNextCounterbalanced(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)

	e := envelope.New("eval")
	g.EvalNext(e)
	if !e.OK || !diagCodes(e)["EVAL_NO_TASKS"] {
		t.Fatalf("an empty task set is a finding, not a failure, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	evalTasks(t, repo,
		`{"id":"t1","statement":"land a one-line fix","done_check":"true","source":"debt row"}`+"\n"+
			`{"id":"t2","statement":"answer a repo question","done_check":"true","source":"session shape"}`+"\n")

	e = envelope.New("eval")
	g.EvalNext(e)
	if !e.OK {
		t.Fatalf("next failed: %+v", e.Diagnostics)
	}
	r := e.Result.(map[string]any)
	if r["task"] != "t1" || r["condition"] != "standard" || r["installed"] != "standard" {
		t.Fatalf("first task's first arm is standard (index parity), got %+v", r)
	}

	rec := envelope.New("eval")
	g.EvalRecord(rec, "t1", 0, 0, "", "")
	if !rec.OK {
		t.Fatalf("record failed: %+v", rec.Diagnostics)
	}
	b, _ := os.ReadFile(filepath.Join(repo, engine.EvalRunsPath))
	if !strings.Contains(string(b), `"condition":"standard"`) {
		t.Fatalf("the condition is read from the installed harness, never asserted: %s", b)
	}

	e = envelope.New("eval")
	g.EvalNext(e)
	r = e.Result.(map[string]any)
	if r["task"] != "t1" || r["condition"] != "minimal" {
		t.Fatalf("the pair completes before the next task starts, got %+v", r)
	}
	if !strings.Contains(r["action"].(string), "profile") {
		t.Fatalf("a condition not installed must direct the switch through plan/apply, got %+v", r)
	}

	switchProfile(t, g, repo, "minimal")
	rec = envelope.New("eval")
	g.EvalRecord(rec, "t1", 1, 0, "", "")
	if !rec.OK {
		t.Fatalf("record under minimal failed: %+v", rec.Diagnostics)
	}

	e = envelope.New("eval")
	g.EvalNext(e)
	r = e.Result.(map[string]any)
	if r["task"] != "t2" || r["condition"] != "minimal" {
		t.Fatalf("t2's first arm is minimal (counterbalanced), got %+v", r)
	}
}

// Recording refuses holes, unknown tasks, and — rule 3's teeth — a
// declared composition that is not what the lock shows installed.
func TestEvalRecordRefusals(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)
	evalTasks(t, repo, `{"id":"t1","statement":"s","done_check":"true","source":"x"}`+"\n")

	e := envelope.New("eval")
	g.EvalRecord(e, "", -1, 0, "", "")
	if e.OK || !diagCodes(e)["MISSING_ARGUMENT"] {
		t.Fatalf("missing --task/--exit must refuse, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	e = envelope.New("eval")
	g.EvalRecord(e, "ghost", 0, 0, "", "")
	if e.OK || !diagCodes(e)["NOT_FOUND"] {
		t.Fatalf("unknown task must refuse, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	// declare minimal without applying: the condition is ambiguous
	cfgPath := filepath.Join(repo, "harness.yaml")
	cfg, _ := os.ReadFile(cfgPath)
	updated := strings.Replace(string(cfg), "adapter: claude-code\n",
		"adapter: claude-code\nprofile: minimal\n", 1)
	if err := os.WriteFile(cfgPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	e = envelope.New("eval")
	g.EvalRecord(e, "t1", 0, 0, "", "")
	if e.OK || !diagCodes(e)["EVAL_CONDITION_UNAPPLIED"] {
		t.Fatalf("an unapplied declaration must refuse recording, got ok=%v %+v", e.OK, e.Diagnostics)
	}
	apply(t, g)
	e = envelope.New("eval")
	g.EvalRecord(e, "t1", 0, 0, "", "")
	if !e.OK {
		t.Fatalf("record after apply must succeed: %+v", e.Diagnostics)
	}
}

// The read pairs runs per task and operator, excludes voided runs
// with the exclusion announced, flags incomplete tasks, and carries
// its misreadings in the result — never a bare number.
func TestEvalReadPairsAndMisreadings(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)
	evalTasks(t, repo,
		`{"id":"t1","statement":"s","done_check":"true","source":"x"}`+"\n"+
			`{"id":"broken","statement":"no check yet","source":"x"}`+"\n")

	rec := envelope.New("eval")
	g.EvalRecord(rec, "t1", 0, 41, "op-a", "")
	if !rec.OK {
		t.Fatal("standard-arm record failed")
	}
	switchProfile(t, g, repo, "minimal")
	rec = envelope.New("eval")
	g.EvalRecord(rec, "t1", 1, 65, "op-a", "flaky env")
	if !rec.OK {
		t.Fatal("minimal-arm record failed")
	}
	minRunID := rec.Result.(map[string]any)["id"].(string)

	e := envelope.New("eval")
	g.Eval(e)
	if !e.OK || !diagCodes(e)["OUTPUT_OFFLOADED"] {
		t.Fatalf("read must offload, got ok=%v %+v", e.OK, e.Diagnostics)
	}
	if !diagCodes(e)["EVAL_TASK_INCOMPLETE"] {
		t.Fatalf("a task without a done_check is flagged, never silently used, got %+v", e.Diagnostics)
	}
	res := e.Result.(map[string]any)
	if res["pairs"] != 1 {
		t.Fatalf("one completed pair expected, got %+v", res)
	}
	if _, ok := res["misreadings"]; !ok {
		t.Fatal("aggregates carry their misreadings in the result")
	}
	b, err := os.ReadFile(filepath.Join(repo, ".rein/out/eval.json"))
	if err != nil {
		t.Fatalf("offloaded report missing: %v", err)
	}
	var full map[string]any
	if err := json.Unmarshal(b, &full); err != nil {
		t.Fatal(err)
	}
	groups := full["groups"].([]any)
	if len(groups) != 1 {
		t.Fatalf("one operator group expected, got %+v", groups)
	}
	grp := groups[0].(map[string]any)
	if grp["operator"] != "op-a" || grp["standard_wins"].(float64) != 1 {
		t.Fatalf("the pair is a standard win (exit 0 vs 1), got %+v", grp)
	}
	if !strings.Contains(string(b), "blinding") {
		t.Fatal("the no-blinding limit is stated in every report")
	}

	// void the minimal run by appending — never rewriting — and the
	// pair dissolves with the exclusion announced
	f, err := os.OpenFile(filepath.Join(repo, engine.EvalRunsPath), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"kind":"void","run":"` + minRunID + `"}` + "\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	e = envelope.New("eval")
	g.Eval(e)
	res = e.Result.(map[string]any)
	if res["pairs"] != 0 || res["voided"] != 1 {
		t.Fatalf("a voided run dissolves its pair with the count announced, got %+v", res)
	}
}
