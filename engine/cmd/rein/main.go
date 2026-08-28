// rein — the FreeRein engine. Agent-facing by contract: one JSON
// envelope per invocation (spec/cli-envelope.md).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

const usage = `rein — harness lifecycle manager (walking skeleton)

usage: rein <command> [flags]

commands:
  init      declare a harness in this repo (writes harness.yaml)
  plan      show what apply would change (adds/changes/drift/removes)
  apply     render and install the resolved harness (--yes to confirm)
  dump      print the resolved composition (detail to .rein/out/dump.json)
  doctor    check installed harness: drift, tampering, stale rent
  add       fetch a registry component into the vendor tree and declare it
  remove    undeclare a component (files removed by the next apply)
  info      show a component's manifest before installing anything
  upgrade   check registry components for newer versions (--yes to vendor)
  adapters  list embedded host adapters
  version   engine version

flags:
  --dir <path>       target repo (default: cwd)
  --adapter <name>   host adapter for init (default: claude-code)
  --registry <src>   registry index URL or path (overrides harness.yaml)
  --yes              confirm side-effectful commands
  --human            human-readable output instead of the JSON envelope
`

func main() {
	fs := flag.NewFlagSet("rein", flag.ContinueOnError)
	dir := fs.String("dir", ".", "target repo")
	adapterName := fs.String("adapter", "claude-code", "host adapter (init)")
	registrySrc := fs.String("registry", "", "registry index URL or path")
	yes := fs.Bool("yes", false, "confirm side effects")
	human := fs.Bool("human", false, "human-readable output")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	if len(os.Args) < 2 {
		fs.Usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
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
		if fs.Arg(0) == "" {
			e.Fail("MISSING_ARGUMENT", cmd+" needs a component argument", "e.g. `rein "+cmd+" name@version`")
			break
		}
		switch cmd {
		case "add":
			g.Add(e, fs.Arg(0), *registrySrc)
		case "remove":
			g.Remove(e, fs.Arg(0))
		case "info":
			g.Info(e, fs.Arg(0), *registrySrc)
		}
	case "upgrade":
		g.Upgrade(e, *yes, *registrySrc)
	case "adapters":
		g.Adapters(e)
	case "version":
		e.Result = map[string]string{"version": engine.Version}
	default:
		e.Fail("UNKNOWN_COMMAND", "no such command: "+cmd, "run `rein` with no arguments for usage")
	}
	os.Exit(e.Emit(*human))
}
