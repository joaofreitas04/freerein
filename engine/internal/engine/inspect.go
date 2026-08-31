package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// rein inspect (spec/inspection.md): the discovery half of setup made
// mechanical. Detection is pure file inspection — inspect never
// executes project code; a test candidate is a command derived from a
// manifest, and running it is the setup procedure's judgment call.

type Candidate struct {
	Command string `json:"command"`
	Source  string `json:"source"`
}

type CorpusFile struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

type Churn struct {
	Path    string `json:"path"`
	Changes int    `json:"changes"`
}

// InspectFormatVersion is semver'd with spec/inspection.md rule 5:
// consumers may branch on it, and additions bump the minor.
const InspectFormatVersion = "0.2.0"

type LangStat struct {
	Language string `json:"language"`
	Files    int    `json:"files"`
	Lines    int    `json:"lines"`
	NonBlank int    `json:"non_blank"`
	Bytes    int64  `json:"bytes"`
}

type ClassifiedFile struct {
	Path  string `json:"path"`
	State string `json:"state"`
}

// Measure is the native size-and-composition family (spec Families
// table + rule 5): every file lands in exactly one state, nothing is
// silently skipped, and the counting method travels in the report
// beside the numbers.
type Measure struct {
	Method     string           `json:"method"`
	Languages  []LangStat       `json:"languages"`
	States     map[string]int   `json:"states"`
	Classified []ClassifiedFile `json:"classified,omitempty"`
	TotalFiles int              `json:"total_files"`
	TotalLines int              `json:"total_lines"`
	TotalBytes int64            `json:"total_bytes"`
}

type InspectReport struct {
	FormatVersion string `json:"formatVersion"`
	Engine        string `json:"engine"`
	Toolchain     struct {
		Languages []string `json:"languages"`
		Manifests []string `json:"manifests"`
		Monorepo  bool     `json:"monorepo"`
	} `json:"toolchain"`
	Tests struct {
		Configs    []string    `json:"configs"`
		Candidates []Candidate `json:"candidates"`
	} `json:"tests"`
	Measure        Measure         `json:"measure"`
	CI             []string        `json:"ci"`
	LintFormat     []string        `json:"lint_format"`
	Instruction    []CorpusFile    `json:"instruction_corpus"`
	ConfigSurfaces []string        `json:"config_surfaces"`
	HighTouch      []Churn         `json:"high_touch,omitempty"`
	DocsTree       string          `json:"docs_tree,omitempty"`
	Affordances    map[string]bool `json:"affordances"`
	Notes          []string        `json:"notes,omitempty"`
}

// ---------- detection tables (files only, never execution) ----------

var langManifests = []struct{ file, lang string }{
	{"go.mod", "go"}, {"go.work", "go"},
	{"package.json", "node"},
	{"Cargo.toml", "rust"},
	{"pyproject.toml", "python"}, {"setup.py", "python"}, {"requirements.txt", "python"},
	{"pom.xml", "jvm"}, {"build.gradle", "jvm"}, {"build.gradle.kts", "jvm"},
	{"Gemfile", "ruby"},
	{"composer.json", "php"},
	{"mix.exs", "elixir"},
}

var monorepoMarkers = []string{"go.work", "pnpm-workspace.yaml", "lerna.json", "turbo.json", "nx.json"}

var testConfigs = []string{
	"pytest.ini", "tox.ini", "conftest.py",
	"jest.config.js", "jest.config.ts", "jest.config.mjs", "jest.config.cjs",
	"vitest.config.ts", "vitest.config.js", "playwright.config.ts",
	".rspec", "phpunit.xml", "phpunit.xml.dist",
}

var ciConfigs = []string{
	".gitlab-ci.yml", ".circleci/config.yml", "Jenkinsfile",
	"azure-pipelines.yml", ".buildkite/pipeline.yml",
}

var lintFormatConfigs = []string{
	".golangci.yml", ".golangci.yaml",
	".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yml", ".eslintrc.yaml",
	"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "eslint.config.ts",
	".prettierrc", ".prettierrc.json", ".prettierrc.yml", ".prettierrc.yaml", ".prettierrc.js",
	"biome.json", "biome.jsonc",
	"ruff.toml", ".ruff.toml", ".flake8",
	".rubocop.yml", "rustfmt.toml", ".clang-format", ".editorconfig",
}

var rootInstructionFiles = []string{
	"CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "AGENT.md", "GEMINI.md",
	"CONTEXT.md",
	".cursorrules", ".windsurfrules", ".clinerules",
	".github/copilot-instructions.md",
}

var configSurfaces = []string{
	".mcp.json", ".claude/settings.json", ".claude/settings.local.json",
	".gemini/settings.json", ".cursor/mcp.json",
}

// directories the bounded walk never enters
// directories the walks never enter (spec rule 1): every
// dot-directory is pruned generically via pruneDir, plus these
// dependency/build names.
var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, "third_party": true,
	"dist": true, "build": true, "target": true, "venv": true,
	"__pycache__": true, "coverage": true,
}

func pruneDir(name string) bool {
	return strings.HasPrefix(name, ".") || skipDirs[name] ||
		strings.HasSuffix(name, ".egg-info")
}

func (g *Engine) fileExists(rel string) bool {
	_, err := os.Stat(filepath.Join(g.Repo, rel))
	return err == nil
}

func (g *Engine) detectToolchain() (langs, manifests []string, monorepo bool) {
	seen := map[string]bool{}
	for _, m := range langManifests {
		if g.fileExists(m.file) {
			manifests = append(manifests, m.file)
			if !seen[m.lang] {
				seen[m.lang] = true
				langs = append(langs, m.lang)
			}
		}
	}
	for _, m := range monorepoMarkers {
		if g.fileExists(m) {
			monorepo = true
		}
	}
	sort.Strings(langs)
	return
}

// candidateProjectCap bounds per-project candidate lists; the cut
// announces itself in notes (spec/inspection.md rule 5).
const candidateProjectCap = 5

func (g *Engine) detectTests() (configs []string, candidates []Candidate, note string) {
	for _, f := range testConfigs {
		if g.fileExists(f) {
			configs = append(configs, f)
		}
	}
	if g.fileExists("go.mod") {
		candidates = append(candidates, Candidate{Command: "go test ./...", Source: "go.mod"})
	}
	if g.fileExists("Cargo.toml") {
		candidates = append(candidates, Candidate{Command: "cargo test", Source: "Cargo.toml"})
	}
	if b, err := os.ReadFile(filepath.Join(g.Repo, "package.json")); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(b, &pkg) == nil {
			names := make([]string, 0, len(pkg.Scripts))
			for n := range pkg.Scripts {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				if strings.TrimSpace(pkg.Scripts[n]) == "" {
					continue
				}
				switch {
				case n == "test":
					candidates = append(candidates, Candidate{Command: "npm test", Source: "package.json scripts.test"})
				case strings.HasPrefix(n, "test:"):
					candidates = append(candidates, Candidate{Command: "npm run " + n, Source: "package.json scripts." + n})
				}
			}
		}
	}
	// nx mediates its workspace's tasks and angular.json project
	// entries may be path strings with targets living elsewhere, so
	// nx.json's existence alone derives the canonical whole-workspace
	// invocation — the same move as go.mod deriving `go test ./...`
	// without promising tests exist [real 10]. Plain angular has no
	// run-all; project test targets derive per-project candidates
	// [bench 1].
	if g.fileExists("nx.json") {
		candidates = append(candidates, Candidate{Command: "npx nx run-many -t test", Source: "nx.json"})
	} else if b, err := os.ReadFile(filepath.Join(g.Repo, "angular.json")); err == nil {
		var ws struct {
			Projects map[string]json.RawMessage `json:"projects"`
		}
		if json.Unmarshal(b, &ws) == nil {
			var withTest []string
			for name, raw := range ws.Projects {
				var p struct {
					Architect map[string]json.RawMessage `json:"architect"`
					Targets   map[string]json.RawMessage `json:"targets"`
				}
				if json.Unmarshal(raw, &p) != nil {
					continue
				}
				if _, ok := p.Architect["test"]; !ok {
					if _, ok = p.Targets["test"]; !ok {
						continue
					}
				}
				withTest = append(withTest, name)
			}
			sort.Strings(withTest)
			if len(withTest) > candidateProjectCap {
				note = fmt.Sprintf("per-project test candidates capped at %d of %d angular.json test targets", candidateProjectCap, len(withTest))
				withTest = withTest[:candidateProjectCap]
			}
			for _, name := range withTest {
				candidates = append(candidates, Candidate{Command: "npx ng test " + name, Source: "angular.json projects." + name})
			}
		}
	}
	for _, f := range []string{"pytest.ini", "tox.ini", "conftest.py"} {
		if g.fileExists(f) {
			candidates = append(candidates, Candidate{Command: "pytest", Source: f})
			break
		}
	}
	if g.fileExists("Gemfile") && g.fileExists(".rspec") {
		candidates = append(candidates, Candidate{Command: "bundle exec rspec", Source: ".rspec"})
	}
	if g.fileExists("gradlew") {
		candidates = append(candidates, Candidate{Command: "./gradlew test", Source: "gradlew"})
	} else if g.fileExists("pom.xml") {
		candidates = append(candidates, Candidate{Command: "mvn -q test", Source: "pom.xml"})
	}
	return
}

func (g *Engine) detectCI() []string {
	var out []string
	wf := filepath.Join(g.Repo, ".github", "workflows")
	if entries, err := os.ReadDir(wf); err == nil {
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yml") || strings.HasSuffix(e.Name(), ".yaml")) {
				out = append(out, ".github/workflows/"+e.Name())
			}
		}
	}
	for _, f := range ciConfigs {
		if g.fileExists(f) {
			out = append(out, f)
		}
	}
	sort.Strings(out)
	return out
}

func (g *Engine) detectFiles(list []string) []string {
	var out []string
	for _, f := range list {
		if g.fileExists(f) {
			out = append(out, f)
		}
	}
	return out
}

func (g *Engine) detectInstructionCorpus() []CorpusFile {
	var out []CorpusFile
	add := func(rel string) {
		if fi, err := os.Stat(filepath.Join(g.Repo, rel)); err == nil && !fi.IsDir() {
			out = append(out, CorpusFile{Path: rel, Bytes: fi.Size()})
		}
	}
	for _, f := range rootInstructionFiles {
		add(f)
	}
	if fi, err := os.Stat(filepath.Join(g.Repo, ".cursor", "rules")); err == nil && fi.IsDir() {
		out = append(out, CorpusFile{Path: ".cursor/rules/", Bytes: 0})
	}
	// nested AGENTS.md / CLAUDE.md, bounded depth, skipping dependency dirs
	root := g.Repo
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if pruneDir(d.Name()) || strings.Count(rel, "/") >= 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(rel, "/") && (d.Name() == "AGENTS.md" || d.Name() == "CLAUDE.md") {
			add(rel)
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// detectHighTouch reads local git history only — never the network.
// Best-effort: no history means no churn map, reported as a note.
const highTouchCap = 10

func (g *Engine) detectHighTouch() ([]Churn, string) {
	cmd := exec.Command("git", "-C", g.Repo, "log", "--name-only", "--pretty=format:", "-n", "500")
	b, err := cmd.Output()
	if err != nil {
		return nil, "git history unavailable — high-touch map omitted"
	}
	counts := map[string]int{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ".rein/") || line == "harness.lock" {
			continue
		}
		counts[line]++
	}
	var out []Churn
	for p, n := range counts {
		out = append(out, Churn{Path: p, Changes: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Changes != out[j].Changes {
			return out[i].Changes > out[j].Changes
		}
		return out[i].Path < out[j].Path
	})
	if len(out) > highTouchCap {
		total := len(out)
		out = out[:highTouchCap]
		return out, fmt.Sprintf("high-touch list capped at %d of %d changed paths", highTouchCap, total)
	}
	return out, ""
}

func (g *Engine) detectDocsTree() string {
	for _, dir := range []string{"docs", "doc"} {
		matches, _ := filepath.Glob(filepath.Join(g.Repo, dir, "*.md"))
		if len(matches) > 0 {
			return dir + "/"
		}
	}
	return ""
}

func (g *Engine) buildInspectReport() *InspectReport {
	r := &InspectReport{FormatVersion: InspectFormatVersion, Engine: g.engineID(), Affordances: map[string]bool{}}
	r.Toolchain.Languages, r.Toolchain.Manifests, r.Toolchain.Monorepo = g.detectToolchain()
	var measureNote string
	r.Measure, measureNote = g.detectMeasure()
	if measureNote != "" {
		r.Notes = append(r.Notes, measureNote)
	}
	var testsNote string
	r.Tests.Configs, r.Tests.Candidates, testsNote = g.detectTests()
	if testsNote != "" {
		r.Notes = append(r.Notes, testsNote)
	}
	r.CI = g.detectCI()
	r.LintFormat = g.detectFiles(lintFormatConfigs)
	r.Instruction = g.detectInstructionCorpus()
	r.ConfigSurfaces = g.detectFiles(configSurfaces)
	var note string
	r.HighTouch, note = g.detectHighTouch()
	if note != "" {
		r.Notes = append(r.Notes, note)
	}
	r.DocsTree = g.detectDocsTree()
	ok, _ := g.probe("git")
	r.Affordances["git"] = ok
	r.Affordances["test-runner"] = len(r.Tests.Candidates) > 0 || len(r.Tests.Configs) > 0
	r.Affordances["ci"] = len(r.CI) > 0
	r.Affordances["linter"] = len(r.LintFormat) > 0
	r.Affordances["docs-tree"] = r.DocsTree != ""
	return r
}

// Inspect works before `rein init` — discovery precedes declaration.
func (g *Engine) Inspect(e *envelope.Envelope) {
	r := g.buildInspectReport()
	outPath := filepath.Join(OutDir, "inspect.json")
	b, _ := json.MarshalIndent(r, "", "  ")
	if err := os.MkdirAll(filepath.Join(g.Repo, OutDir), 0o755); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	if err := os.WriteFile(filepath.Join(g.Repo, outPath), append(b, '\n'), 0o644); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	e.Diag(envelope.Info, "OUTPUT_OFFLOADED", "full report written to "+outPath,
		"read "+outPath+" for the full report omitted from this envelope")
	top := ""
	if len(r.Measure.Languages) > 0 {
		top = r.Measure.Languages[0].Language
	}
	e.Result = map[string]any{
		"detail":            outPath,
		"affordances":       r.Affordances,
		"languages":         r.Toolchain.Languages,
		"top_language":      top,
		"total_lines":       r.Measure.TotalLines,
		"states":            r.Measure.States,
		"test_candidates":   len(r.Tests.Candidates),
		"ci_configs":        len(r.CI),
		"instruction_files": len(r.Instruction),
		"config_surfaces":   len(r.ConfigSurfaces),
		"docs_tree":         r.DocsTree,
	}
}

// ---------- measure: every file lands in exactly one state ----------

const (
	oversizeBytes       = 8 << 20 // beyond this, bytes are counted and lines are not
	binaryProbeBytes    = 8192
	generatedProbeBytes = 2048
	classifiedCap       = 200
)

// measureMethod travels inside the report (spec rule 5): counting is
// opinion, and a count whose method lives elsewhere is not comparable.
const measureMethod = "lines = physical newline count; non_blank = lines with any non-whitespace byte; " +
	"no code/comment split (a lexer's job this surface deliberately does not take on); " +
	"binary = no Unicode BOM and a zero byte in the first 8192 bytes; " +
	"generated = a known derived-file name, or a generation marker in the first 2048 bytes; " +
	"duplicate = equal size and equal SHA-256, first path kept and later paths marked; " +
	"oversize = over 8 MiB (bytes counted, lines not); " +
	"unknown = no language mapping for the name or shebang; " +
	"error = the file could not be read"

var extLanguages = map[string]string{
	".go": "Go", ".py": "Python",
	".js": "JavaScript", ".jsx": "JavaScript", ".mjs": "JavaScript", ".cjs": "JavaScript",
	".ts": "TypeScript", ".tsx": "TypeScript",
	".rb": "Ruby", ".rs": "Rust", ".java": "Java", ".kt": "Kotlin", ".kts": "Kotlin",
	".c": "C", ".h": "C/C++ header", ".cpp": "C++", ".cc": "C++", ".hpp": "C/C++ header",
	".cs": "C#", ".php": "PHP", ".swift": "Swift", ".scala": "Scala",
	".sh": "Shell", ".bash": "Shell", ".zsh": "Shell",
	".sql": "SQL", ".html": "HTML", ".htm": "HTML",
	".css": "CSS", ".scss": "CSS", ".sass": "CSS", ".less": "CSS",
	".md": "Markdown", ".markdown": "Markdown",
	".yaml": "YAML", ".yml": "YAML", ".json": "JSON", ".toml": "TOML", ".xml": "XML",
	".proto": "Protocol Buffers", ".tf": "Terraform",
	".ex": "Elixir", ".exs": "Elixir", ".erl": "Erlang", ".hs": "Haskell",
	".lua": "Lua", ".pl": "Perl", ".r": "R",
	".dart": "Dart", ".vue": "Vue", ".svelte": "Svelte", ".zig": "Zig",
	".txt": "Text",
}

var nameLanguages = map[string]string{
	"Makefile": "Makefile", "Dockerfile": "Dockerfile", "Jenkinsfile": "Groovy",
	"Rakefile": "Ruby", "Gemfile": "Ruby",
}

var generatedExactNames = map[string]bool{
	"package-lock.json": true, "pnpm-lock.yaml": true, "go.sum": true,
}

var generatedSuffixes = []string{".lock", ".min.js", ".min.css", ".g.dart", ".pb.go"}

var generatedMarkers = []string{
	"do not edit", "@generated", "code generated by",
	"automatically generated", "auto-generated",
}

func generatedName(name string) bool {
	if generatedExactNames[name] {
		return true
	}
	for _, suf := range generatedSuffixes {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func generatedContent(b []byte) bool {
	head := b
	if len(head) > generatedProbeBytes {
		head = head[:generatedProbeBytes]
	}
	lower := strings.ToLower(string(head))
	for _, m := range generatedMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

var textBOMs = [][]byte{
	{0xFF, 0xFE, 0x00, 0x00}, {0x00, 0x00, 0xFE, 0xFF}, // UTF-32 first: its BOMs prefix UTF-16's
	{0xEF, 0xBB, 0xBF},         // UTF-8
	{0xFF, 0xFE}, {0xFE, 0xFF}, // UTF-16
}

func isBinary(b []byte) bool {
	for _, bom := range textBOMs {
		if bytes.HasPrefix(b, bom) {
			return false
		}
	}
	head := b
	if len(head) > binaryProbeBytes {
		head = head[:binaryProbeBytes]
	}
	return bytes.IndexByte(head, 0) >= 0
}

func countLines(b []byte) (physical, nonBlank int) {
	for len(b) > 0 {
		i := bytes.IndexByte(b, '\n')
		line := b
		if i >= 0 {
			line = b[:i]
			b = b[i+1:]
		} else {
			b = nil
		}
		physical++
		if len(bytes.TrimSpace(line)) > 0 {
			nonBlank++
		}
	}
	return
}

var shebangLanguages = []struct{ marker, lang string }{
	{"python", "Python"}, {"node", "JavaScript"}, {"ruby", "Ruby"},
	{"perl", "Perl"}, {"bash", "Shell"}, {"sh", "Shell"},
}

// languageFor maps by well-known filename, then extension, then — for
// extensionless scripts — the shebang line (reading, never running).
func languageFor(name string, content []byte) (string, bool) {
	if l, ok := nameLanguages[name]; ok {
		return l, true
	}
	if l, ok := extLanguages[strings.ToLower(filepath.Ext(name))]; ok {
		return l, true
	}
	if bytes.HasPrefix(content, []byte("#!")) {
		first := content
		if i := bytes.IndexByte(first, '\n'); i >= 0 {
			first = first[:i]
		}
		line := string(first)
		for _, sb := range shebangLanguages {
			if strings.Contains(line, sb.marker) {
				return sb.lang, true
			}
		}
	}
	return "", false
}

// detectMeasure reads file contents (the one family that does) but
// still executes nothing. Every file visited lands in exactly one
// state; the classified list makes the non-analyzed ones inspectable.
func (g *Engine) detectMeasure() (Measure, string) {
	m := Measure{
		Method: measureMethod,
		States: map[string]int{
			"analyzed": 0, "empty": 0, "binary": 0, "oversize": 0,
			"generated": 0, "duplicate": 0, "unknown": 0, "error": 0,
		},
	}
	byLang := map[string]*LangStat{}
	seen := map[[32]byte]bool{}
	classifiedTotal := 0
	classify := func(rel, state string) {
		m.States[state]++
		classifiedTotal++
		if len(m.Classified) < classifiedCap {
			m.Classified = append(m.Classified, ClassifiedFile{Path: rel, State: state})
		}
	}
	root := g.Repo
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		name := d.Name()
		if d.IsDir() {
			if pruneDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil // never follow a link out of the tree being described
		}
		info, ierr := d.Info()
		if ierr != nil {
			classify(rel, "error")
			return nil
		}
		size := info.Size()
		m.TotalFiles++
		m.TotalBytes += size
		switch {
		case size == 0:
			classify(rel, "empty")
			return nil
		case generatedName(name):
			classify(rel, "generated")
			return nil
		case size > oversizeBytes:
			classify(rel, "oversize")
			return nil
		}
		b, ferr := os.ReadFile(p)
		if ferr != nil {
			classify(rel, "error")
			return nil
		}
		if isBinary(b) {
			classify(rel, "binary")
			return nil
		}
		if generatedContent(b) {
			classify(rel, "generated")
			return nil
		}
		h := sha256.Sum256(b)
		if seen[h] {
			classify(rel, "duplicate")
			return nil
		}
		seen[h] = true
		lang, known := languageFor(name, b)
		if !known {
			classify(rel, "unknown")
			return nil
		}
		st := byLang[lang]
		if st == nil {
			st = &LangStat{Language: lang}
			byLang[lang] = st
		}
		physical, nonBlank := countLines(b)
		st.Files++
		st.Lines += physical
		st.NonBlank += nonBlank
		st.Bytes += size
		m.States["analyzed"]++
		m.TotalLines += physical
		return nil
	})
	for _, st := range byLang {
		m.Languages = append(m.Languages, *st)
	}
	sort.Slice(m.Languages, func(i, j int) bool {
		if m.Languages[i].Lines != m.Languages[j].Lines {
			return m.Languages[i].Lines > m.Languages[j].Lines
		}
		return m.Languages[i].Language < m.Languages[j].Language
	})
	sort.Slice(m.Classified, func(i, j int) bool { return m.Classified[i].Path < m.Classified[j].Path })
	note := ""
	if classifiedTotal > classifiedCap {
		note = fmt.Sprintf("classified list capped at %d of %d non-analyzed files", classifiedCap, classifiedTotal)
	}
	return m, note
}
