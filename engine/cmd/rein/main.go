// rein — the FreeRein engine. Agent-facing by contract: one JSON
// envelope per invocation (spec/cli-envelope.md).
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

const usage = `rein — harness lifecycle manager (walking skeleton)

usage: rein <command> [flags]

commands:
  init      declare a harness in this repo (writes harness.yaml)
  inspect   survey the repo — toolchain, tests, ci, instruction corpus
            (detection only; never executes project code)
  plan      show what apply would change (adds/changes/drift/removes)
  apply     render and install the resolved harness (--yes to confirm)
  dump      print the resolved composition (detail to .rein/out/dump.json)
  doctor    check installed harness: drift, tampering, stale rent
  add       fetch a registry component into the vendor tree and declare it
  remove    undeclare a component (files removed by the next apply)
  info      show a component's manifest before installing anything
  upgrade   check registry components for newer versions (--yes to vendor)
  adapters  list embedded host adapters
  probes    list the affordance vocabulary requires: entries may use
  gaps      join absent affordances to the registry's addresses —
            coverage facts and creation candidates, never lift claims
            (works before init; --registry overrides the source)
  note      append a note entry to the harness journal
  attest    record a proof a procedure performed (subject: gate-can-fail)
  cite      record that a fragment shaped an action (no argument: read
            the counts back, lowest first, with decay candidates)
  propose   open an evolve proposal — all six fields required
            (--surface --change --prediction --measurement --baseline
            --nominated-by); one open proposal at a time
  verdict   close a proposal on its measurement's result
            (--proposal --outcome accepted|rejected --evidence)
  journal   read the harness history back out (newest first, offloaded;
            --kind/--since/--path filter, --limit caps inline entries)
  report    assemble a field report for an installed component from the
            lock, journal, and citation stats (--subsystem
            --failure-class --reproduction --disposition); written to
            .rein/out/report/, never transmitted
  eval      the paired measurement instrument (spec/eval.md): "eval
            next" assigns task + condition (deterministic,
            counterbalanced); "eval record --task <id> --exit <code>"
            appends a run under the installed condition (--seconds
            --operator --note optional); bare "eval" reads pairs,
            groups, and misreadings — the engine never runs a task
  version   engine version

flags:
  --dir <path>       target repo (default: cwd)
  --adapter <name>   host adapter for init (default: claude-code)
  --profile <name>   core profile for init: standard, or minimal — the
                     control condition, not a starter tier
  --registry <src>   registry index URL or path (overrides harness.yaml
                     and the built-in default registry)
  --yes              confirm side-effectful commands
  --human            human-readable output instead of the JSON envelope
`

// splitArgs parses flags interleaved with positionals. Go's flag
// package stops at the first non-flag; an agent writing
// `rein add x --registry y` must not have the flag silently dropped.
func splitArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return positionals, nil
		}
		positionals = append(positionals, fs.Arg(0))
		rest = fs.Args()[1:]
	}
}

// resolveCommand finds the command wherever it sits in the vector:
// agents write `rein --dir x inspect` as often as `rein inspect
// --dir x`, and both must dispatch [real 9]. Empty command with a
// nil error is the usage path.
func resolveCommand(fs *flag.FlagSet, args []string) (string, []string, error) {
	positionals, err := splitArgs(fs, args)
	if err != nil {
		return "", nil, err
	}
	if len(positionals) == 0 {
		return "", nil, nil
	}
	return positionals[0], positionals[1:], nil
}

func main() {
	fs := flag.NewFlagSet("rein", flag.ContinueOnError)
	dir := fs.String("dir", ".", "target repo")
	adapterName := fs.String("adapter", "claude-code", "host adapter (init)")
	profile := fs.String("profile", "standard", "core profile for init (standard | minimal)")
	registrySrc := fs.String("registry", "", "registry index URL or path")
	yes := fs.Bool("yes", false, "confirm side effects")
	kind := fs.String("kind", "", "journal: exact kind filter")
	since := fs.String("since", "", "journal: RFC3339 or YYYY-MM-DD lower bound")
	pathF := fs.String("path", "", "journal: only entries touching this path")
	limit := fs.Int("limit", 20, "journal: inline entry cap (0 = counts only)")
	surface := fs.String("surface", "", "propose: artifact/component being changed")
	change := fs.String("change", "", "propose: one line, what changes")
	prediction := fs.String("prediction", "", "propose: falsifiable — next time X, Y instead of Z")
	measurement := fs.String("measurement", "", "propose: what decides, named before the change")
	baseline := fs.String("baseline", "", "propose: condition compared against")
	nominatedBy := fs.String("nominated-by", "", "propose: cite-decay|compensation-recheck|overlap|recurrence|human")
	proposalID := fs.String("proposal", "", "verdict: proposal id")
	outcome := fs.String("outcome", "", "verdict: accepted|rejected")
	evidence := fs.String("evidence", "", "verdict: the measurement's result")
	task := fs.String("task", "", "eval record: the assigned task id")
	exitCode := fs.Int("exit", -1, "eval record: the done_check's exit code")
	seconds := fs.Int("seconds", 0, "eval record: wall-clock seconds (optional)")
	operator := fs.String("operator", "", "eval record: operator label — different operators are different experiments")
	noteF := fs.String("note", "", "eval record: free-text note (optional)")
	subsystem := fs.String("subsystem", "", "report: diagnose attribution (instructions|tools|environment|state|feedback)")
	failureClass := fs.String("failure-class", "", "report: generalized statement of the defect")
	reproduction := fs.String("reproduction", "", "report: minimal generalized trigger, inputs as affordances")
	disposition := fs.String("disposition", "", "report: fix|default-change|table-entry|docs")
	human := fs.Bool("human", false, "human-readable output")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	cmd, positionals, err := resolveCommand(fs, os.Args[1:])
	if err != nil {
		os.Exit(2)
	}
	if cmd == "" {
		fs.Usage()
		os.Exit(2)
	}

	e := envelope.New(cmd)
	g := &engine.Engine{Repo: *dir, Content: content.FS}

	switch cmd {
	case "init":
		g.Init(e, *adapterName, *profile)
	case "plan":
		g.Plan(e)
	case "apply":
		g.Apply(e, *yes)
	case "dump":
		g.Dump(e)
	case "doctor":
		g.Doctor(e)
	case "add", "remove", "info":
		if len(positionals) == 0 {
			e.Fail("MISSING_ARGUMENT", cmd+" needs a component argument", "e.g. `rein "+cmd+" name@version`")
			break
		}
		switch cmd {
		case "add":
			g.Add(e, positionals[0], *registrySrc)
		case "remove":
			g.Remove(e, positionals[0])
		case "info":
			g.Info(e, positionals[0], *registrySrc)
		}
	case "upgrade":
		g.Upgrade(e, *yes, *registrySrc)
	case "adapters":
		g.Adapters(e)
	case "probes":
		g.Probes(e)
	case "gaps":
		g.Gaps(e, *registrySrc)
	case "inspect":
		g.Inspect(e)
	case "journal":
		g.Journal(e, engine.JournalOpts{Kind: *kind, Since: *since, Path: *pathF, Limit: *limit})
	case "propose":
		g.Propose(e, engine.ProposalFields{Surface: *surface, Change: *change,
			Prediction: *prediction, Measurement: *measurement,
			Baseline: *baseline, NominatedBy: *nominatedBy})
	case "verdict":
		g.Verdict(e, *proposalID, *outcome, *evidence)
	case "cite":
		id := ""
		if len(positionals) > 0 {
			id = positionals[0]
		}
		g.Cite(e, id)
	case "attest":
		if len(positionals) == 0 {
			e.Fail("MISSING_ARGUMENT", "attest needs a subject", "e.g. `rein attest gate-can-fail`")
			break
		}
		g.Attest(e, positionals[0])
	case "report":
		if len(positionals) == 0 {
			e.Fail("MISSING_ARGUMENT", "report needs a component argument",
				"e.g. `rein report instructions-base --subsystem … --failure-class … --reproduction … --disposition …`")
			break
		}
		g.Report(e, positionals[0], engine.ReportFields{Subsystem: *subsystem,
			FailureClass: *failureClass, Reproduction: *reproduction, Disposition: *disposition})
	case "eval":
		sub := ""
		if len(positionals) > 0 {
			sub = positionals[0]
		}
		switch sub {
		case "":
			g.Eval(e)
		case "next":
			g.EvalNext(e)
		case "record":
			g.EvalRecord(e, *task, *exitCode, *seconds, *operator, *noteF)
		default:
			e.Fail("INVALID_ARGUMENT", "unknown eval action "+sub,
				"eval actions: `rein eval` (read), `rein eval next` (assign), `rein eval record --task <id> --exit <code>`")
		}
	case "note":
		g.Note(e, strings.Join(positionals, " "))
	case "version":
		e.Result = map[string]string{"version": engine.Version}
	default:
		e.Fail("UNKNOWN_COMMAND", "no such command: "+cmd, "run `rein` with no arguments for usage")
	}
	os.Exit(e.Emit(*human))
}
