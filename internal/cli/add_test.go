package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func initWithConfig(t *testing.T, dir string) {
	t.Helper()
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("init: %v\n%s", err, out.String())
	}
}

func listChangesetFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, ".changeset"))
	if err != nil {
		t.Fatalf("read .changeset: %v", err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	return names
}

func TestAddCmd_WithFlags(t *testing.T) {
	dir := initRepo(t)
	initWithConfig(t, dir)

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"add", "--bump=patch", "--message=Fix a bug."})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if got := listChangesetFiles(t, dir); len(got) != 1 {
		t.Fatalf("expected 1 changeset file, got %v", got)
	}
}

func TestAddCmd_InteractivePrompt(t *testing.T) {
	dir := initRepo(t)
	initWithConfig(t, dir)

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("minor\nAdd a new feature.\n"))
	root.SetArgs([]string{"add"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if got := listChangesetFiles(t, dir); len(got) != 1 {
		t.Fatalf("expected 1 changeset file, got %v", got)
	}
}

func TestAddCmd_FailsWithInvalidBump(t *testing.T) {
	dir := initRepo(t)
	initWithConfig(t, dir)

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"add", "--bump=nonsense", "--message=x"})

	if err := root.Execute(); err == nil {
		t.Fatalf("expected error for invalid bump")
	}
}

func TestAddCmd_FailsWithoutConfig(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"add", "--bump=patch", "--message=x"})

	if err := root.Execute(); err == nil {
		t.Fatalf("expected error when .changeset/config.json is missing")
	}
}
