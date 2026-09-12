package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCmd_BumpsFromAccumulatedChangesets(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")
	writeRepoFile(t, dir, ".changeset/b.md", "---\n\"pkg\": patch\n---\n\nFix two.\n")
	writeRepoFile(t, dir, ".changeset/c.md", "---\n\"pkg\": minor\n---\n\nAdd feature.\n")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	if !strings.Contains(string(data), "## v0.1.0") {
		t.Fatalf("expected a v0.1.0 header (patch+patch+minor -> minor bump from v0.0.0), got:\n%s", data)
	}

	entries, err := os.ReadDir(filepath.Join(dir, ".changeset"))
	if err != nil {
		t.Fatalf("read .changeset: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			t.Fatalf("expected used changeset files to be removed, found %s", e.Name())
		}
	}
}

func TestVersionCmd_NoChangesetsIsNoop(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if _, err := os.Stat(filepath.Join(dir, "CHANGELOG.md")); err == nil {
		t.Fatalf("expected no CHANGELOG.md to be created when there are no changesets")
	}
}
