// Package config reads and writes .changeset/config.json, the only
// per-repository configuration `changeset` needs.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultBaseBranch = "main"

// Config is the contents of .changeset/config.json.
type Config struct {
	PackageName string `json:"packageName"`
	BaseBranch  string `json:"baseBranch"`
}

// Path returns the path to the config file rooted at dir.
func Path(dir string) string {
	return filepath.Join(dir, ".changeset", "config.json")
}

// Init creates .changeset/ and a default config.json rooted at dir. It
// fails if dir has no go.mod (there is nowhere to infer a package name
// from) or if a config already exists (it never silently overwrites).
func Init(dir string) error {
	pkgName, err := packageNameFromGoMod(dir)
	if err != nil {
		return fmt.Errorf("determine package name: %w", err)
	}

	path := Path(dir)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}

	cfg := Config{PackageName: pkgName, BaseBranch: defaultBaseBranch}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Load reads and validates the config rooted at dir.
func Load(dir string) (*Config, error) {
	path := Path(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.PackageName == "" {
		return nil, fmt.Errorf("%s: packageName must not be empty", path)
	}
	return &cfg, nil
}

func packageNameFromGoMod(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		modPath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
		if modPath == "" {
			return "", fmt.Errorf("go.mod has an empty module directive")
		}
		parts := strings.Split(modPath, "/")
		return parts[len(parts)-1], nil
	}
	return "", fmt.Errorf("go.mod does not contain a module directive")
}
