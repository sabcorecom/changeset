package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a git repo in a temp dir, chdirs into it for the
// duration of the test, and restores the original cwd on cleanup.
func initRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGitCmd(t, dir, "init", "-q", "-b", "main")
	runGitCmd(t, dir, "config", "user.email", "test@example.com")
	runGitCmd(t, dir, "config", "user.name", "Test")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restore chdir: %v", err)
		}
	})
	return dir
}

func runGitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func writeRepoFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func gitCommit(t *testing.T, dir, message string) {
	t.Helper()
	runGitCmd(t, dir, "add", "-A")
	runGitCmd(t, dir, "commit", "-q", "-m", message)
}

func TestBump_NoChangesetsIsNoop(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	next, section, err := Bump(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "" || section != "" {
		t.Fatalf("expected empty result for no changesets, got next=%q section=%q", next, section)
	}
	if _, err := os.Stat(filepath.Join(dir, "CHANGELOG.md")); err == nil {
		t.Fatalf("expected no CHANGELOG.md to be created when there are no changesets")
	}
}

func TestBump_SingleChangeset(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	next, section, err := Bump(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "v0.0.1" {
		t.Fatalf("got next=%q, want v0.0.1", next)
	}
	if !strings.Contains(section, "Fix one.") {
		t.Fatalf("expected section to contain the changeset summary, got:\n%s", section)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	if !strings.Contains(string(data), "## v0.0.1") {
		t.Fatalf("expected a v0.0.1 header, got:\n%s", data)
	}

	if _, err := os.Stat(filepath.Join(dir, ".changeset", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("expected consumed changeset file to be removed, stat err=%v", err)
	}
}

func TestBump_MultipleChangesetsTakesMaxBump(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")
	writeRepoFile(t, dir, ".changeset/b.md", "---\n\"pkg\": patch\n---\n\nFix two.\n")
	writeRepoFile(t, dir, ".changeset/c.md", "---\n\"pkg\": minor\n---\n\nAdd feature.\n")

	next, _, err := Bump(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "v0.1.0" {
		t.Fatalf("got next=%q, want v0.1.0 (patch+patch+minor -> minor bump)", next)
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

func TestBump_MalformedChangesetFileErrors(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	gitCommit(t, dir, "init module")

	writeRepoFile(t, dir, ".changeset/a.md", "not a valid changeset\n")

	if _, _, err := Bump(dir); err == nil {
		t.Fatalf("expected error for malformed changeset file, got nil")
	}
}
