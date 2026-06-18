// Package config loads user-defined commands from
// ~/.docker-toolbox/config.yaml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	dirName  = ".docker-toolbox"
	fileName = "config.yaml"
)

// Config is the parsed user configuration.
type Config struct {
	Commands map[string]CustomCommand `yaml:"commands"`
}

// CustomCommand is a user-defined workflow.
type CustomCommand struct {
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}

// Step is one step of a custom command. It is either a plain command
// string or an explicit shell step (`shell: "..."`).
//
// The distinction is part of the hybrid execution model: scalar steps
// are reserved for future SDK-native toolbox verbs, while `shell:` steps
// are always run verbatim in a shell. In v0.1 both forms run as external
// commands and require the global --allow-shell flag.
type Step struct {
	// Raw is set for scalar steps, e.g. `- docker compose up -d`.
	Raw string
	// Shell is set for mapping steps, e.g. `- shell: "docker compose up -d"`.
	Shell string
}

// IsShell reports whether the step is an explicit shell step.
func (s Step) IsShell() bool { return s.Shell != "" }

// Command returns the command line this step should run.
func (s Step) Command() string {
	if s.IsShell() {
		return s.Shell
	}
	return s.Raw
}

// UnmarshalYAML accepts either a scalar string or a mapping with a
// non-empty `shell` key.
func (s *Step) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return value.Decode(&s.Raw)
	case yaml.MappingNode:
		var m struct {
			Shell string `yaml:"shell"`
		}
		if err := value.Decode(&m); err != nil {
			return err
		}
		if m.Shell == "" {
			return fmt.Errorf("step mapping must contain a non-empty 'shell' key")
		}
		s.Shell = m.Shell
		return nil
	default:
		return fmt.Errorf("step must be a string or a mapping with a 'shell' key")
	}
}

// Path returns the absolute path to the user config file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName, fileName), nil
}

// Load reads the user config. A missing file is not an error: it returns
// an empty config so the tool works with no setup.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Commands: map[string]CustomCommand{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if cfg.Commands == nil {
		cfg.Commands = map[string]CustomCommand{}
	}
	return &cfg, nil
}
