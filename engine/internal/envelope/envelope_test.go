package envelope_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// spec/cli-envelope.md rules 1, 6, 7: failures stay well-formed, exit
// codes are contractual, fix is omitted when empty and result/confirm
// are explicitly present.
func TestEmitContract(t *testing.T) {
	e := envelope.New("plan")
	e.Diag(envelope.Info, "UP_TO_DATE", "nothing to do", "")
	var out bytes.Buffer
	if code := e.EmitTo(&out, false); code != 0 {
		t.Fatalf("ok envelope must exit 0, got %d", code)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &wire); err != nil {
		t.Fatalf("emit must be one JSON document: %v", err)
	}
	for _, key := range []string{"ok", "command", "result", "diagnostics", "confirm_required"} {
		if _, present := wire[key]; !present {
			t.Fatalf("the two branch fields and the core are always present; %q missing in %s", key, out.Bytes())
		}
	}
	if strings.Contains(out.String(), `"fix"`) {
		t.Fatalf("an empty fix is omitted, not serialized (rule 7): %s", out.String())
	}

	e = envelope.New("apply")
	e.Fail("WRITE_FAILED", "disk full", "free space and re-run")
	out.Reset()
	if code := e.EmitTo(&out, false); code != 1 {
		t.Fatalf("failed envelope must exit 1, got %d", code)
	}
	if !strings.Contains(out.String(), `"fix": "free space and re-run"`) {
		t.Fatalf("a carried fix must serialize: %s", out.String())
	}
}

func TestEmitHuman(t *testing.T) {
	e := envelope.New("doctor")
	e.Diag(envelope.Warning, "DRIFT", "CLAUDE.md drifted", "review and merge")
	e.Result = map[string]int{"findings": 1}
	var out bytes.Buffer
	if code := e.EmitTo(&out, true); code != 0 {
		t.Fatalf("warnings alone stay exit 0, got %d", code)
	}
	for _, want := range []string{"rein doctor: ok", "[warning] DRIFT", "fix: review and merge", "findings"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("human render must carry %q, got:\n%s", want, out.String())
		}
	}
}
