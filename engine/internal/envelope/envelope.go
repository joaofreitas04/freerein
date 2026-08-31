// Package envelope implements spec/cli-envelope.md: every rein
// invocation emits exactly one JSON document; failures stay
// well-formed; every diagnostic carries a fix.
package envelope

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

type Diagnostic struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Fix      string   `json:"fix,omitempty"`
}

type Envelope struct {
	OK              bool         `json:"ok"`
	Command         string       `json:"command"`
	Result          any          `json:"result"`
	Diagnostics     []Diagnostic `json:"diagnostics"`
	ConfirmRequired any          `json:"confirm_required"`
}

func New(command string) *Envelope {
	return &Envelope{OK: true, Command: command, Diagnostics: []Diagnostic{}}
}

func (e *Envelope) Diag(sev Severity, code, message, fix string) {
	e.Diagnostics = append(e.Diagnostics, Diagnostic{sev, code, message, fix})
	if sev == Error {
		e.OK = false
	}
}

func (e *Envelope) Fail(code, message, fix string) { e.Diag(Error, code, message, fix) }

// Emit writes the envelope to stdout and returns the process exit
// code (spec/cli-envelope.md rule 6: 0 = ok, 1 = not ok; exit 2 is
// main's, for invocations that fail before a command exists).
func (e *Envelope) Emit(human bool) int {
	return e.EmitTo(os.Stdout, human)
}

// EmitTo is Emit against any writer — the seam that makes the
// exit-code and wire contracts testable.
func (e *Envelope) EmitTo(w io.Writer, human bool) int {
	if human {
		e.emitHuman(w)
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(e)
	}
	if e.OK {
		return 0
	}
	return 1
}

func (e *Envelope) emitHuman(w io.Writer) {
	status := "ok"
	if !e.OK {
		status = "FAILED"
	}
	fmt.Fprintf(w, "rein %s: %s\n", e.Command, status)
	for _, d := range e.Diagnostics {
		fmt.Fprintf(w, "  [%s] %s: %s\n", d.Severity, d.Code, d.Message)
		if d.Fix != "" {
			fmt.Fprintf(w, "      fix: %s\n", d.Fix)
		}
	}
	if e.ConfirmRequired != nil {
		b, _ := json.MarshalIndent(e.ConfirmRequired, "  ", "  ")
		fmt.Fprintf(w, "  confirmation required for:\n  %s\n  re-run with --yes\n", b)
	}
	if e.Result != nil {
		b, _ := json.MarshalIndent(e.Result, "  ", "  ")
		fmt.Fprintf(w, "  result:\n  %s\n", b)
	}
}
