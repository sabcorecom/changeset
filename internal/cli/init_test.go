package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd_CreatesConfig(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if _, err := os.Stat(filepath.Join(dir, ".changeset", "config.json")); err != nil {
		t.Fatalf("expected config.json to be created: %v", err)
	}
}

func TestInitCmd_FailsWithoutGoMod(t *testing.T) {
	initRepo(t)

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init"})

	if err := root.Execute(); err == nil {
		t.Fatalf("expected error without go.mod")
	}
}

func TestInitCmd_FailsIfConfigAlreadyExists(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first init: %v", err)
	}

	root = NewRootCmd("test")
	out.Reset()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err == nil {
		t.Fatalf("expected error on second init")
	}
}
