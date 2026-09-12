package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGoMod(t *testing.T, dir, module string) {
	t.Helper()
	content := "module " + module + "\n\ngo 1.25\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func TestInit_CreatesConfig(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "github.com/sabcorecom/changeset")

	if err := Init(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load after init: %v", err)
	}
	if cfg.PackageName != "changeset" {
		t.Fatalf("got packageName %q, want changeset", cfg.PackageName)
	}
	if cfg.BaseBranch != "main" {
		t.Fatalf("got baseBranch %q, want main", cfg.BaseBranch)
	}
}

func TestInit_FailsWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err == nil {
		t.Fatalf("expected error when go.mod is absent")
	}
}

func TestInit_FailsIfConfigAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "github.com/example/foo")

	if err := Init(dir); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if err := Init(dir); err == nil {
		t.Fatalf("expected error on second init, config already exists")
	}
}

func TestLoad_FailsOnMissingConfig(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatalf("expected error for missing config")
	}
}

func TestLoad_FailsOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".changeset"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(Path(dir), []byte("not json"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatalf("expected error for invalid JSON config")
	}
}

func TestLoad_FailsOnEmptyPackageName(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".changeset"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(Path(dir), []byte(`{"packageName":"","baseBranch":"main"}`), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := Load(dir); err == nil {
		t.Fatalf("expected error for empty packageName")
	}
}
