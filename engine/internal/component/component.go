// Package component implements spec/component-manifest.md: parsing and
// validation of component.yaml. Unknown keys are rejected, not ignored.
package component

import (
	"fmt"
	"io/fs"
	"path"
	"sort"

	"gopkg.in/yaml.v3"
)

type Rent struct {
	Class   string `yaml:"class"`
	Expires string `yaml:"expires"`
}

type Manifest struct {
	Name        string   `yaml:"name"`
	Kind        string   `yaml:"kind"`
	Version     string   `yaml:"version"`
	Subsystem   string   `yaml:"subsystem"`
	Rung        string   `yaml:"rung"`
	Rent        Rent     `yaml:"rent"`
	Provides    []string `yaml:"provides"`
	Seeds       []string `yaml:"seeds"`
	Requires    []string `yaml:"requires"`
	Conflicts   []string `yaml:"conflicts"`
	Description string   `yaml:"description"`
}

// Loaded is a manifest plus its payload files (path -> content),
// where each payload path is relative to the target repo root.
type Loaded struct {
	Manifest Manifest
	Files    map[string][]byte
}

var (
	kinds      = set("core", "extension", "preset", "bundle")
	subsystems = set("instructions", "tools", "environment", "state", "feedback")
	rungs      = set("instruction", "conditional", "permission", "hook", "isolation")
	rentKinds  = set("compensation", "amplifier")
	// The probe vocabulary (spec rule 4). Grown deliberately, in
	// lockstep with the engine's probes.
	Probes = set("git")
)

func set(ss ...string) map[string]bool {
	m := map[string]bool{}
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func (m *Manifest) Validate() error {
	switch {
	case m.Name == "":
		return fmt.Errorf("name is required")
	case !kinds[m.Kind]:
		return fmt.Errorf("%s: kind %q not one of core|extension|preset|bundle", m.Name, m.Kind)
	case m.Version == "":
		return fmt.Errorf("%s: version is required", m.Name)
	case !subsystems[m.Subsystem]:
		return fmt.Errorf("%s: subsystem %q not one of instructions|tools|environment|state|feedback", m.Name, m.Subsystem)
	case !rungs[m.Rung]:
		return fmt.Errorf("%s: rung %q not one of instruction|conditional|permission|hook|isolation", m.Name, m.Rung)
	case !rentKinds[m.Rent.Class]:
		return fmt.Errorf("%s: rent.class %q not one of compensation|amplifier", m.Name, m.Rent.Class)
	case m.Rent.Class == "compensation" && m.Rent.Expires == "":
		return fmt.Errorf("%s: rent.class compensation requires rent.expires (the re-check trigger)", m.Name)
	}
	for _, r := range m.Requires {
		if !Probes[r] {
			return fmt.Errorf("%s: requires %q is not in the probe vocabulary", m.Name, r)
		}
	}
	provided := map[string]bool{}
	for _, p := range m.Provides {
		provided[p] = true
	}
	for _, s := range m.Seeds {
		if !provided[s] {
			return fmt.Errorf("%s: seeds %q must also be listed in provides", m.Name, s)
		}
	}
	return nil
}

func (m *Manifest) Ref() string { return m.Name + "@" + m.Version }

// LoadDir loads one component from a directory: component.yaml plus
// payload files. Every payload file must be listed in provides, and
// every provides entry must exist (spec: the manifest is enforced).
func LoadDir(fsys fs.FS, dir string) (*Loaded, error) {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "component.yaml"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", dir, err)
	}
	var m Manifest
	dec := yaml.NewDecoder(bytesReader(raw))
	dec.KnownFields(true) // unknown keys are rejected, not ignored
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("%s/component.yaml: %w", dir, err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	lc := &Loaded{Manifest: m, Files: map[string][]byte{}}
	err = fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Base(p) == "component.yaml" {
			return err
		}
		rel := p[len(dir)+1:]
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		lc.Files[rel] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	declared := map[string]bool{}
	for _, p := range m.Provides {
		declared[p] = true
		if _, ok := lc.Files[p]; !ok {
			return nil, fmt.Errorf("%s: provides %q but ships no such file", m.Name, p)
		}
	}
	for p := range lc.Files {
		if !declared[p] {
			return nil, fmt.Errorf("%s: ships %q but does not declare it in provides", m.Name, p)
		}
	}
	return lc, nil
}

// LoadAll loads every component directory directly under root, sorted
// by name for determinism.
func LoadAll(fsys fs.FS, root string) ([]*Loaded, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}
	var out []*Loaded
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		lc, err := LoadDir(fsys, path.Join(root, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, lc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Manifest.Name < out[j].Manifest.Name })
	return out, nil
}
