// Package adapter implements spec/host-adapter.md: a declarative
// mapping from host-neutral artifacts to host-specific paths.
// Adapters carry no logic.
package adapter

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"

	"gopkg.in/yaml.v3"
)

type InstructionFile struct {
	Path     string `yaml:"path"`
	Imports  bool   `yaml:"imports"`
	MaxBytes int    `yaml:"max_bytes"`
}

type Skills struct {
	Dir             string `yaml:"dir"`
	Frontmatter     string `yaml:"frontmatter"`
	UserInvokedFlag string `yaml:"user_invoked_flag"`
}

type Hooks struct {
	Events                    []string `yaml:"events"`
	PostCompactionReinjection bool     `yaml:"post_compaction_reinjection"`
}

type Settings struct {
	Path        string `yaml:"path"`
	Permissions bool   `yaml:"permissions"`
}

type Adapter struct {
	Name            string          `yaml:"name"`
	Version         string          `yaml:"version"`
	InstructionFile InstructionFile `yaml:"instruction_file"`
	Skills          Skills          `yaml:"skills"`
	Hooks           Hooks           `yaml:"hooks"`
	Settings        Settings        `yaml:"settings"`
	StateDir        string          `yaml:"state_dir"`
	Degradations    []string        `yaml:"degradations"`
}

func Load(fsys fs.FS, name string) (*Adapter, error) {
	raw, err := fs.ReadFile(fsys, path.Join("adapters", name+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("adapter %q: %w", name, err)
	}
	var a Adapter
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&a); err != nil {
		return nil, fmt.Errorf("adapter %q: %w", name, err)
	}
	if a.Name == "" || a.InstructionFile.Path == "" {
		return nil, fmt.Errorf("adapter %q: name and instruction_file.path are required", name)
	}
	return &a, nil
}
