package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// rein eval (spec/eval.md): the paired measurement instrument. The
// one hard line — the engine never runs a task. It holds the task
// set, assigns the next run deterministically and counterbalanced so
// the operator cannot cherry-pick, reads the active condition from
// the installed harness instead of trusting a flag, records runs
// append-only, and computes paired aggregates that carry their
// misreadings. Judgment executes; computation keeps score.

const (
	// EvalFormatVersion matches spec/eval.md's contract version.
	EvalFormatVersion = "0.1.0"
	EvalTasksPath     = ".rein/eval/tasks.jsonl" // curated by the human, committed
	EvalRunsPath      = ".rein/eval/runs.jsonl"  // append-only, engine-written
)

// evalArms is the default comparison (spec rule 6): the minimal
// profile is the permanent control arm. Per-component conditions
// arrive when a real experiment demands them.
var evalArms = [2]string{"standard", "minimal"}

// The rule-4/rule-5 misreadings, attached to every read: a paired
// number without its characteristic lies invites the wrong verdict.
var evalMisreadings = []string{
	"small N decides nothing alone — few pairs flatter whichever arm got lucky",
	"a tie is information, not noise: the harness may not bind on that task class",
	"runs by different operators or models are different experiments — groups are reported separately, never pooled",
	"no blinding is possible: the operator sees the installed condition; mechanical done-checks and fixed assignment are the compensation, not a cure",
}

type EvalTask struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	DoneCheck string `json:"done_check"`
	Source    string `json:"source"`
}

type evalRun struct {
	At        string `json:"at,omitempty"`
	Engine    string `json:"engine,omitempty"`
	Kind      string `json:"kind,omitempty"` // "" = run; "void" retracts Run
	ID        string `json:"id,omitempty"`
	Run       string `json:"run,omitempty"`
	Task      string `json:"task,omitempty"`
	Condition string `json:"condition,omitempty"`
	Exit      *int   `json:"exit,omitempty"`
	Seconds   int    `json:"seconds,omitempty"`
	Operator  string `json:"operator,omitempty"`
	Note      string `json:"note,omitempty"`
}

// loadEvalTasks returns the valid tasks in file order plus the ids of
// rows missing their done_check — rule 1: a task without a runnable
// check is not a task yet, and it is flagged, never silently used.
func (g *Engine) loadEvalTasks() (tasks []EvalTask, incomplete []string) {
	b, err := os.ReadFile(filepath.Join(g.Repo, EvalTasksPath))
	if err != nil {
		return nil, nil
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var t EvalTask
		if json.Unmarshal([]byte(line), &t) != nil || t.ID == "" {
			incomplete = append(incomplete, strings.TrimSpace(line))
			continue
		}
		if strings.TrimSpace(t.DoneCheck) == "" || strings.TrimSpace(t.Statement) == "" {
			incomplete = append(incomplete, t.ID)
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, incomplete
}

// loadEvalRuns returns non-voided runs in file order, the voided
// count, and the malformed-line count. A bad run is corrected by a
// later void entry, never rewritten — the journal's discipline.
func (g *Engine) loadEvalRuns() (runs []evalRun, voided, malformed int) {
	b, err := os.ReadFile(filepath.Join(g.Repo, EvalRunsPath))
	if err != nil {
		return nil, 0, 0
	}
	var all []evalRun
	voidedIDs := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var r evalRun
		if json.Unmarshal([]byte(line), &r) != nil {
			malformed++
			continue
		}
		if r.Kind == "void" {
			voidedIDs[r.Run] = true
			continue
		}
		all = append(all, r)
	}
	for _, r := range all {
		if voidedIDs[r.ID] {
			voided++
			continue
		}
		runs = append(runs, r)
	}
	return runs, voided, malformed
}

// installedProfile reads the declared condition; "" means standard.
func evalCondition(cfg *Config) string {
	if cfg.Profile == "" {
		return "standard"
	}
	return cfg.Profile
}

func (g *Engine) evalConfig(e *envelope.Envelope) *Config {
	cfg, err := g.readConfig()
	if os.IsNotExist(err) {
		e.Fail("NOT_INITIALIZED", ConfigName+" not found in "+g.Repo, "run `rein init` first")
		return nil
	}
	if err != nil {
		e.Fail("CONFIG_INVALID", err.Error(), "fix "+ConfigName+" and re-run")
		return nil
	}
	return cfg
}

const evalTaskShape = `one JSON object per line in ` + EvalTasksPath +
	`: {"id","statement","done_check","source"} — tasks are the repo's own work ` +
	`(debt rows, recurring session shapes), and a task without a runnable done_check is not a task yet`

// EvalNext assigns the next run: first task in file order whose pair
// is incomplete, first arm alternating by task position (rule 2 —
// choosing your own arm is the eval's version of the proposer
// accepting). The engine assigns; the operator runs and records.
func (g *Engine) EvalNext(e *envelope.Envelope) {
	cfg := g.evalConfig(e)
	if cfg == nil {
		return
	}
	tasks, incomplete := g.loadEvalTasks()
	evalIncompleteDiag(e, incomplete)
	if len(tasks) == 0 {
		e.Diag(envelope.Info, "EVAL_NO_TASKS", "no tasks in "+EvalTasksPath, "curate the set from the repo's own work: "+evalTaskShape)
		e.Result = map[string]any{"tasks": 0}
		return
	}
	runs, _, _ := g.loadEvalRuns()
	covered := map[string]map[string]bool{}
	for _, r := range runs {
		if covered[r.Task] == nil {
			covered[r.Task] = map[string]bool{}
		}
		covered[r.Task][r.Condition] = true
	}
	installed := evalCondition(cfg)
	for i, t := range tasks {
		first, second := evalArms[i%2], evalArms[(i+1)%2]
		arm := ""
		switch {
		case !covered[t.ID][first]:
			arm = first
		case !covered[t.ID][second]:
			arm = second
		default:
			continue
		}
		action := "run the task under the installed composition; when its done_check has an exit code, `rein eval record --task " + t.ID + " --exit <code>`"
		if arm != installed {
			action = "switch first — set `profile: " + arm + "` in " + ConfigName + " and `rein apply --yes` (the switch lands in the journal) — then run the task and `rein eval record --task " + t.ID + " --exit <code>`"
		}
		e.Result = map[string]any{
			"task": t.ID, "statement": t.Statement, "done_check": t.DoneCheck,
			"condition": arm, "installed": installed, "action": action,
		}
		return
	}
	e.Diag(envelope.Info, "EVAL_COMPLETE", "every task has runs under both arms", "read the differential with `rein eval`, or curate more tasks: "+evalTaskShape)
	e.Result = map[string]any{"tasks": len(tasks), "assignment": nil}
}

func evalIncompleteDiag(e *envelope.Envelope, incomplete []string) {
	if len(incomplete) == 0 {
		return
	}
	e.Diag(envelope.Warning, "EVAL_TASK_INCOMPLETE",
		fmt.Sprintf("%d task row(s) unusable (missing done_check, statement, or id): %s", len(incomplete), strings.Join(incomplete, ", ")),
		"outcomes are what the check says, so a task without a runnable done_check is excluded from assignment and pairing — complete the row; "+evalTaskShape)
}

// EvalRecord appends one run under the condition the installed
// harness shows (rule 3: installed, not claimed — there is no
// condition flag to assert). A declared-but-unapplied composition is
// refused: the condition would be ambiguous.
func (g *Engine) EvalRecord(e *envelope.Envelope, taskID string, exitCode, seconds int, operator, note string) {
	var missing []string
	if strings.TrimSpace(taskID) == "" {
		missing = append(missing, "--task")
	}
	if exitCode < 0 {
		missing = append(missing, "--exit")
	}
	if len(missing) > 0 {
		e.Fail("MISSING_ARGUMENT", "a run is refused with a hole: "+strings.Join(missing, ", ")+" missing",
			"--task names the assigned task; --exit is the done_check's exit code — the outcome is what the check said, not what the operator felt")
		return
	}
	cfg := g.evalConfig(e)
	if cfg == nil {
		return
	}
	tasks, _ := g.loadEvalTasks()
	known := false
	for _, t := range tasks {
		if t.ID == taskID {
			known = true
		}
	}
	if !known {
		e.Fail("NOT_FOUND", "no task "+taskID+" in "+EvalTasksPath,
			"`rein eval next` assigns; "+evalTaskShape)
		return
	}
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	if p, err := g.computePlan(r); err != nil || !p.empty() {
		e.Fail("EVAL_CONDITION_UNAPPLIED",
			"the declared composition differs from what is installed — the run's condition would be a claim, not a fact",
			"conditions are read from the installed harness: `rein apply --yes` (or revert "+ConfigName+"), then record")
		return
	}
	condition := evalCondition(cfg)
	at := time.Now().UTC().Format(time.RFC3339)
	prior, _, _ := g.loadEvalRuns()
	sum := sha256.Sum256([]byte(taskID + "|" + condition + "|" + at + "|" + strconv.Itoa(len(prior))))
	id := "r-" + hex.EncodeToString(sum[:])[:8]
	entry := evalRun{At: at, Engine: g.engineID(), ID: id, Task: taskID,
		Condition: condition, Exit: &exitCode, Seconds: seconds, Operator: operator, Note: note}
	line, _ := json.Marshal(entry)
	full := filepath.Join(g.Repo, EvalRunsPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on .rein/eval")
		return
	}
	f, err := os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+EvalRunsPath)
		return
	}
	_, werr := f.Write(append(line, '\n'))
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		e.Fail("WRITE_FAILED", werr.Error(), "check permissions on "+EvalRunsPath)
		return
	}
	e.Result = map[string]any{"id": id, "task": taskID, "condition": condition,
		"done": exitCode == 0, "recorded": EvalRunsPath, "next": "`rein eval next` for the assignment"}
}

type evalPair struct {
	Task     string         `json:"task"`
	Operator string         `json:"operator"`
	Outcome  string         `json:"outcome"` // standard | minimal | tie-both-done | tie-both-failed
	Exits    map[string]int `json:"exits"`
	Seconds  map[string]int `json:"seconds,omitempty"`
}

type evalGroup struct {
	Operator     string `json:"operator"`
	Pairs        int    `json:"pairs"`
	StandardWins int    `json:"standard_wins"`
	MinimalWins  int    `json:"minimal_wins"`
	BothDone     int    `json:"both_done"`
	BothFailed   int    `json:"both_failed"`
}

// Eval reads the differential back: latest non-voided run per
// (task, operator, condition), paired where both arms exist, grouped
// by operator — never pooled — with the misreadings in the result
// itself. The eval is an instrument, not a judge: it never closes a
// proposal.
func (g *Engine) Eval(e *envelope.Envelope) {
	if cfg := g.evalConfig(e); cfg == nil {
		return
	}
	tasks, incomplete := g.loadEvalTasks()
	evalIncompleteDiag(e, incomplete)
	runs, voided, malformed := g.loadEvalRuns()
	if malformed > 0 {
		e.Diag(envelope.Warning, "EVAL_LINE_UNREADABLE",
			fmt.Sprintf("%d unreadable line(s) in %s skipped", malformed, EvalRunsPath),
			"repair by appending a corrected entry — the run store is append-only, never edited in place")
	}
	valid := map[string]bool{}
	for _, t := range tasks {
		valid[t.ID] = true
	}
	// latest run per (task, operator, condition): file order, last wins
	latest := map[string]map[string]map[string]evalRun{}
	for _, r := range runs {
		if !valid[r.Task] || r.Exit == nil {
			continue
		}
		if latest[r.Task] == nil {
			latest[r.Task] = map[string]map[string]evalRun{}
		}
		if latest[r.Task][r.Operator] == nil {
			latest[r.Task][r.Operator] = map[string]evalRun{}
		}
		latest[r.Task][r.Operator][r.Condition] = r
	}
	var pairs []evalPair
	groups := map[string]*evalGroup{}
	for _, t := range tasks {
		for op, byCond := range latest[t.ID] {
			std, okS := byCond[evalArms[0]]
			min, okM := byCond[evalArms[1]]
			if !okS || !okM {
				continue
			}
			stdDone, minDone := *std.Exit == 0, *min.Exit == 0
			outcome := "tie-both-failed"
			switch {
			case stdDone && minDone:
				outcome = "tie-both-done"
			case stdDone:
				outcome = evalArms[0]
			case minDone:
				outcome = evalArms[1]
			}
			p := evalPair{Task: t.ID, Operator: op, Outcome: outcome,
				Exits: map[string]int{evalArms[0]: *std.Exit, evalArms[1]: *min.Exit}}
			if std.Seconds > 0 && min.Seconds > 0 {
				p.Seconds = map[string]int{evalArms[0]: std.Seconds, evalArms[1]: min.Seconds}
			}
			pairs = append(pairs, p)
			grp := groups[op]
			if grp == nil {
				grp = &evalGroup{Operator: op}
				groups[op] = grp
			}
			grp.Pairs++
			switch outcome {
			case evalArms[0]:
				grp.StandardWins++
			case evalArms[1]:
				grp.MinimalWins++
			case "tie-both-done":
				grp.BothDone++
			default:
				grp.BothFailed++
			}
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Task != pairs[j].Task {
			return pairs[i].Task < pairs[j].Task
		}
		return pairs[i].Operator < pairs[j].Operator
	})
	var groupList []evalGroup
	for _, grp := range groups {
		groupList = append(groupList, *grp)
	}
	sort.Slice(groupList, func(i, j int) bool { return groupList[i].Operator < groupList[j].Operator })
	if groupList == nil {
		groupList = []evalGroup{}
	}
	full := map[string]any{
		"formatVersion": EvalFormatVersion, "engine": g.engineID(),
		"arms": evalArms, "tasks": len(tasks), "runs": len(runs),
		"voided": voided, "pairs": pairs, "groups": groupList,
		"misreadings": evalMisreadings,
	}
	outPath := filepath.Join(OutDir, "eval.json")
	b, _ := json.MarshalIndent(full, "", "  ")
	if err := os.MkdirAll(filepath.Join(g.Repo, OutDir), 0o755); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	if err := os.WriteFile(filepath.Join(g.Repo, outPath), append(b, '\n'), 0o644); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	e.Diag(envelope.Info, "OUTPUT_OFFLOADED", "full differential written to "+outPath,
		"read "+outPath+" for per-pair detail omitted from this envelope")
	e.Result = map[string]any{
		"detail": outPath, "tasks": len(tasks), "runs": len(runs),
		"voided": voided, "pairs": len(pairs), "groups": len(groupList),
		"misreadings": evalMisreadings,
	}
}
