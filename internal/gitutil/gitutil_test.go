package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// chdirTemp creates a fresh git repo in a temp dir, chdirs into it for the
// duration of the test, and restores the original cwd on cleanup.
func chdirTemp(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	run(t, dir, "config", "user.email", "test@example.com")
	run(t, dir, "config", "user.name", "Test")

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

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func commit(t *testing.T, dir, file, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	run(t, dir, "add", file)
	run(t, dir, "commit", "-q", "-m", "commit "+file)
}

func TestLatestTag_NoTags(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "a.txt", "a")

	tag, ok, err := LatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok || tag != "" {
		t.Fatalf("expected no tag, got %q ok=%v", tag, ok)
	}
}

func TestLatestTag_VersionSorted(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "a.txt", "a")
	run(t, dir, "tag", "v0.1.0")
	run(t, dir, "tag", "v0.2.0")
	run(t, dir, "tag", "v0.10.0")

	tag, ok, err := LatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected a tag to be found")
	}
	if tag != "v0.10.0" {
		t.Fatalf("got %q, want v0.10.0 (version sort, not lexicographic)", tag)
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "a.txt", "a")

	branch, err := CurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Fatalf("got %q, want main", branch)
	}
}

func TestChangedFiles(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "base.txt", "base")
	run(t, dir, "branch", "feature")
	run(t, dir, "checkout", "-q", "feature")
	commit(t, dir, "new.go", "package main")

	files, err := ChangedFiles("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "new.go" {
		t.Fatalf("got %v, want [new.go]", files)
	}
}

func TestCreateTag_IdempotentOnSameHEAD(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "a.txt", "a")

	if err := CreateTag("v1.0.0"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if err := CreateTag("v1.0.0"); err != nil {
		t.Fatalf("expected idempotent no-op, got error: %v", err)
	}
}

func TestCreateTag_ErrorsWhenTagPointsElsewhere(t *testing.T) {
	dir := chdirTemp(t)
	commit(t, dir, "a.txt", "a")

	if err := CreateTag("v1.0.0"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	commit(t, dir, "b.txt", "b")
	if err := CreateTag("v1.0.0"); err == nil {
		t.Fatalf("expected error when tag already points at a different commit")
	}
}
