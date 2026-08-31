package engine_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// makeArchive builds a component tar.gz and returns its sha256.
func makeArchive(t *testing.T, path string, files map[string]string) string {
	t.Helper()
	var names []string
	for n := range files {
		names = append(names, n)
	}
	// stable order
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	var buf strings.Builder
	gz := gzip.NewWriter(&stringWriter{&buf})
	tw := tar.NewWriter(gz)
	for _, n := range names {
		content := files[n]
		if err := tw.WriteHeader(&tar.Header{Name: n, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gz.Close()
	raw := []byte(buf.String())
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

type stringWriter struct{ b *strings.Builder }

func (w *stringWriter) Write(p []byte) (int, error) { return w.b.Write(p) }

const regExtManifest = `name: reg-ext
kind: extension
version: %s
subsystem: instructions
rung: instruction
rent:
  class: amplifier
provides:
  - AGENTS.md.d/60-reg.md
description: a registry extension
`

// writeRegistry builds a local file registry with reg-ext versions.
func writeRegistry(t *testing.T, dir string, versions map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	type rel struct {
		URL         string `json:"url"`
		Sha256      string `json:"sha256"`
		Kind        string `json:"kind"`
		Description string `json:"description,omitempty"`
	}
	idx := map[string]map[string]map[string]rel{"components": {"reg-ext": {}}}
	for v, fragment := range versions {
		archive := "reg-ext-" + v + ".tar.gz"
		sha := makeArchive(t, filepath.Join(dir, archive), map[string]string{
			"component.yaml":        strings.Replace(regExtManifest, "%s", v, 1),
			"AGENTS.md.d/60-reg.md": fragment,
		})
		idx["components"]["reg-ext"][v] = rel{URL: archive, Sha256: sha, Kind: "extension"}
	}
	b, _ := json.Marshal(idx)
	path := filepath.Join(dir, "index.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRegistryLifecycle(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	reg := writeRegistry(t, filepath.Join(repo, "registry"),
		map[string]string{"1.0.0": "## Reg rule\n\nBe reg v1.\n"})

	// info before anything is written
	e := envelope.New("info")
	g.Info(e, "reg-ext", reg)
	if !e.OK {
		t.Fatalf("info failed: %+v", e.Diagnostics)
	}
	res := e.Result.(map[string]any)
	if res["version"] != "1.0.0" || res["kind"] != "extension" {
		t.Fatalf("info: unexpected %v", res)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor")); err == nil {
		t.Fatal("info must not vendor anything")
	}

	// add: vendors, declares, installs nothing
	e = envelope.New("add")
	g.Add(e, "reg-ext", reg) // no version -> latest
	if !e.OK {
		t.Fatalf("add failed: %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0/component.yaml")); err != nil {
		t.Fatal("add must vendor the component")
	}
	cm := filepath.Join(repo, "CLAUDE.md")
	if _, err := os.Stat(cm); err == nil {
		t.Fatal("add must not install anything")
	}

	// apply: the fragment lands; the lock pins the vendor sha
	apply(t, g)
	b, _ := os.ReadFile(cm)
	if !strings.Contains(string(b), "Be reg v1.") {
		t.Fatal("registry extension fragment must render")
	}
	lock, _ := os.ReadFile(filepath.Join(repo, "harness.lock"))
	if !strings.Contains(string(lock), `"source": "registry:reg-ext@1.0.0"`) || !strings.Contains(string(lock), `"sha": "sha256:`) {
		t.Fatalf("lock must pin the registry layer, got:\n%s", lock)
	}

	// doctor: clean, then tampered vendor detected
	ed := envelope.New("doctor")
	g.Doctor(ed)
	if !ed.OK {
		t.Fatalf("doctor on clean install: %+v", ed.Diagnostics)
	}
	vf := filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0/AGENTS.md.d/60-reg.md")
	if err := os.WriteFile(vf, []byte("## Reg rule\n\nEVIL.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ed = envelope.New("doctor")
	g.Doctor(ed)
	if ed.OK || !diagCodes(ed)["VENDOR_TAMPERED"] {
		t.Fatalf("doctor must detect vendor tampering, got %+v", ed.Diagnostics)
	}
	// restore by re-adding (re-fetch)
	e = envelope.New("add")
	g.Add(e, "reg-ext@1.0.0", reg)
	if !e.OK {
		t.Fatalf("re-add failed: %+v", e.Diagnostics)
	}
	ed = envelope.New("doctor")
	g.Doctor(ed)
	if !ed.OK {
		t.Fatalf("doctor after re-fetch: %+v", ed.Diagnostics)
	}

	// upgrade: two-phase, then vendors the new version; apply merges
	// around a local edit elsewhere in the file
	reg = writeRegistry(t, filepath.Join(repo, "registry"), map[string]string{
		"1.0.0": "## Reg rule\n\nBe reg v1.\n",
		"1.1.0": "## Reg rule\n\nBe reg v2.\n",
	})
	e = envelope.New("upgrade")
	g.Upgrade(e, false, reg)
	if e.ConfirmRequired == nil {
		t.Fatalf("upgrade without --yes must confirm, got %+v", e.Diagnostics)
	}
	orig, _ := os.ReadFile(cm)
	_ = os.WriteFile(cm, append([]byte("## Local addendum\n\nkeep me\n\n"), orig...), 0o644)
	e = envelope.New("upgrade")
	g.Upgrade(e, true, reg)
	if !e.OK {
		t.Fatalf("upgrade --yes failed: %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0")); err == nil {
		t.Fatal("old vendored version must be removed after upgrade")
	}
	ea := envelope.New("apply")
	g.Apply(ea, true)
	if !ea.OK || !diagCodes(ea)["MERGED"] {
		t.Fatalf("apply after upgrade must merge around the local edit, got %+v", ea.Diagnostics)
	}
	after, _ := os.ReadFile(cm)
	if !strings.Contains(string(after), "keep me") || !strings.Contains(string(after), "Be reg v2.") {
		t.Fatalf("both the local edit and the upgrade must survive:\n%s", after)
	}

	// remove: undeclares; apply removes the files and the fragment
	e = envelope.New("remove")
	g.Remove(e, "reg-ext")
	if !e.OK {
		t.Fatalf("remove failed: %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.1.0")); err == nil {
		t.Fatal("vendor dir must be gone after remove")
	}
	ea = envelope.New("apply")
	g.Apply(ea, true)
	// CLAUDE.md still carries the local addendum -> drift; the reg
	// fragment disappearing is an upstream change -> merge removes it
	after, _ = os.ReadFile(cm)
	if strings.Contains(string(after), "Be reg v2.") {
		t.Fatalf("removed extension's fragment must leave the instruction file:\n%s", after)
	}
	if !strings.Contains(string(after), "keep me") {
		t.Fatal("local edit must survive the removal merge")
	}
}

// A sha mismatch between index and archive must refuse to install.
func TestRegistryShaMismatch(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	reg := writeRegistry(t, filepath.Join(repo, "registry"),
		map[string]string{"1.0.0": "## Reg rule\n\nBe reg v1.\n"})
	// corrupt the archive after indexing
	archive := filepath.Join(repo, "registry", "reg-ext-1.0.0.tar.gz")
	b, _ := os.ReadFile(archive)
	_ = os.WriteFile(archive, append(b, 0x00), 0o644)
	e := envelope.New("add")
	g.Add(e, "reg-ext@1.0.0", reg)
	if e.OK || !diagCodes(e)["FETCH_FAILED"] {
		t.Fatalf("sha mismatch must refuse, got %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0")); err == nil {
		t.Fatal("a refused fetch must leave no vendor dir behind")
	}
}

// The debt ledger's evidence line was "registry.httpGet is
// unexercised". These two tests retire it: the same lifecycle and the
// same refusal, over a real HTTP server instead of a filesystem path.
func TestRegistryOverHTTP(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	regDir := filepath.Join(repo, "registry-src")
	writeRegistry(t, regDir, map[string]string{"1.0.0": "## Reg rule\n\nBe reg v1.\n"})
	srv := httptest.NewServer(http.FileServer(http.Dir(regDir)))
	defer srv.Close()

	e := envelope.New("add")
	g.Add(e, "reg-ext", srv.URL+"/index.json")
	if !e.OK {
		t.Fatalf("add over HTTP failed: %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0/component.yaml")); err != nil {
		t.Fatal("add over HTTP must vendor the component")
	}
}

func TestRegistryOverHTTPShaMismatch(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	regDir := filepath.Join(repo, "registry-src")
	writeRegistry(t, regDir, map[string]string{"1.0.0": "## Reg rule\n\nBe reg v1.\n"})
	archive := filepath.Join(regDir, "reg-ext-1.0.0.tar.gz")
	b, _ := os.ReadFile(archive)
	if err := os.WriteFile(archive, append(b, 0x00), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.FileServer(http.Dir(regDir)))
	defer srv.Close()

	e := envelope.New("add")
	g.Add(e, "reg-ext@1.0.0", srv.URL+"/index.json")
	if e.OK || !diagCodes(e)["FETCH_FAILED"] {
		t.Fatalf("sha mismatch over HTTP must refuse, got %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/vendor/reg-ext@1.0.0")); err == nil {
		t.Fatal("a refused HTTP fetch must leave no vendor dir behind")
	}
}
