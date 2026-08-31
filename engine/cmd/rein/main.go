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
  note      append a note entry to the harness journal
  attest    record a proof a procedure performed (subject: gate-can-fail)
  journal   read the harness history back out (newest first, offloaded;
            --kind/--since/--path filter, --limit caps inline entries)
  version   engine version

flags:
  --dir <path>       target repo (default: cwd)
  --adapter <name>   host adapter for init (default: claude-code)
  --registry <src>   registry index URL or path (overrides harness.yaml
                     and the built-in default registry)
  --yes              confirm side-effectful commands
  --human            human-readable output instead of the JSON envelope
`

func main() {
	fs := flag.NewFlagSet("rein", flag.ContinueOnError)
	dir := fs.String("dir", ".", "target repo")
	adapterName := fs.String("adapter", "claude-code", "host adapter (init)")
	registrySrc := fs.String("registry", "", "registry index URL or path")
	yes := fs.Bool("yes", false, "confirm side effects")
	kind := fs.String("kind", "", "journal: exact kind filter")
	since := fs.String("since", "", "journal: RFC3339 or YYYY-MM-DD lower bound")
	pathF := fs.String("path", "", "journal: only entries touching this path")
	limit := fs.Int("limit", 20, "journal: inline entry cap (0 = counts only)")
	human := fs.Bool("human", false, "human-readable output")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	if len(os.Args) < 2 {
		fs.Usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	// allow flags after positionals (Go's flag package stops at the
	// first non-flag; an agent writing `rein add x --registry y`
	// must not have the flag silently dropped)
	var positionals []string
	rest := os.Args[2:]
	for {
		if err := fs.Parse(rest); err != nil {
			os.Exit(2)
		}
		if fs.NArg() == 0 {
			break
		}
		positionals = append(positionals, fs.Arg(0))
		rest = fs.Args()[1:]
	}

	e := envelope.New(cmd)
	g := &engine.Engine{Repo: *dir, Content: content.FS}

	switch cmd {
	case "init":
		g.Init(e, *adapterName)
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
	case "inspect":
		g.Inspect(e)
	case "journal":
		g.Journal(e, engine.JournalOpts{Kind: *kind, Since: *since, Path: *pathF, Limit: *limit})
	case "attest":
		if len(positionals) == 0 {
			e.Fail("MISSING_ARGUMENT", "attest needs a subject", "e.g. `rein attest gate-can-fail`")
			break
		}
		g.Attest(e, positionals[0])
	case "note":
		g.Note(e, strings.Join(positionals, " "))
	case "version":
		e.Result = map[string]string{"version": engine.Version}
	default:
		e.Fail("UNKNOWN_COMMAND", "no such command: "+cmd, "run `rein` with no arguments for usage")
	}
	os.Exit(e.Emit(*human))
}
