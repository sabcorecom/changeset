package bot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sabcorecom/changeset/internal/github"
)

// initRepo creates a git repo with a bare "origin" remote in temp dirs,
// chdirs into the working repo for the duration of the test, and
// restores the original cwd on cleanup.
func initRepo(t *testing.T) (dir, bare string) {
	t.Helper()

	dir = t.TempDir()
	runGitCmd(t, dir, "init", "-q", "-b", "main")
	runGitCmd(t, dir, "config", "user.email", "test@example.com")
	runGitCmd(t, dir, "config", "user.name", "Test")

	bare = t.TempDir()
	runGitCmd(t, bare, "init", "-q", "--bare")
	runGitCmd(t, dir, "remote", "add", "origin", bare)

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
	return dir, bare
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

func writeConfig(t *testing.T, dir, baseBranch string) {
	t.Helper()
	writeRepoFile(t, dir, ".changeset/config.json", fmt.Sprintf("{\"packageName\":\"pkg\",\"baseBranch\":%q}\n", baseBranch))
}

func gitCommit(t *testing.T, dir, message string) {
	t.Helper()
	runGitCmd(t, dir, "add", "-A")
	runGitCmd(t, dir, "commit", "-q", "-m", message)
}

// withFakeGoreleaser installs a fake "goreleaser" script on PATH so the
// publish path's exec.Command call succeeds without the real tool.
func withFakeGoreleaser(t *testing.T, exitCode int) {
	t.Helper()
	binDir := t.TempDir()
	script := filepath.Join(binDir, "goreleaser")
	content := "#!/bin/sh\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake goreleaser: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

type createCall struct {
	head, base, title, body string
}

type updateCall struct {
	number      int
	title, body string
}

// fakeGitHub is an in-memory pullRequestAPI double for exercising Run
// without touching the network.
type fakeGitHub struct {
	existing *github.PullRequest
	findErr  error

	createCalls []createCall
	updateCalls []updateCall
}

func (f *fakeGitHub) FindOpenPullRequest(head, base string) (*github.PullRequest, error) {
	return f.existing, f.findErr
}

func (f *fakeGitHub) CreatePullRequest(head, base, title, body string) (*github.PullRequest, error) {
	f.createCalls = append(f.createCalls, createCall{head, base, title, body})
	return &github.PullRequest{Number: 99}, nil
}

func (f *fakeGitHub) UpdatePullRequest(number int, title, body string) error {
	f.updateCalls = append(f.updateCalls, updateCall{number, title, body})
	return nil
}

func TestRun_NoChangesets_RunsPublish(t *testing.T) {
	dir, bare := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeConfig(t, dir, "main")
	writeRepoFile(t, dir, "CHANGELOG.md", "# Changelog\n\n## v0.1.0\n\n### Patch Changes\n\n- Fix.\n")
	gitCommit(t, dir, "init")
	withFakeGoreleaser(t, 0)

	gh := &fakeGitHub{}
	var out bytes.Buffer
	if err := Run(dir, gh, &out, &out); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if tagOut := runGitCmd(t, dir, "tag", "--list", "v0.1.0"); tagOut == "" {
		t.Fatalf("expected publish path to create tag v0.1.0")
	}
	if len(gh.createCalls) != 0 || len(gh.updateCalls) != 0 {
		t.Fatalf("expected no PR API calls on the publish path, got create=%v update=%v", gh.createCalls, gh.updateCalls)
	}
	if branchOut := runGitCmd(t, bare, "branch", "--list", "changeset-release/main"); strings.TrimSpace(branchOut) != "" {
		t.Fatalf("expected no changeset-release/main branch to be pushed on the publish path")
	}
}

func TestRun_Changesets_NoExistingPR_CreatesPR(t *testing.T) {
	dir, bare := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeConfig(t, dir, "main")
	gitCommit(t, dir, "init")

	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	gh := &fakeGitHub{}
	var out bytes.Buffer
	if err := Run(dir, gh, &out, &out); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if len(gh.createCalls) != 1 {
		t.Fatalf("expected exactly one CreatePullRequest call, got %d", len(gh.createCalls))
	}
	call := gh.createCalls[0]
	if call.head != "changeset-release/main" || call.base != "main" || call.title != CommitTitle {
		t.Fatalf("got %+v", call)
	}
	if len(gh.updateCalls) != 0 {
		t.Fatalf("expected no UpdatePullRequest calls, got %v", gh.updateCalls)
	}

	if branchOut := runGitCmd(t, bare, "branch", "--list", "changeset-release/main"); strings.TrimSpace(branchOut) == "" {
		t.Fatalf("expected changeset-release/main to be pushed to origin")
	}
}

func TestRun_Changesets_ExistingPR_UpdatesPR(t *testing.T) {
	dir, _ := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeConfig(t, dir, "main")
	gitCommit(t, dir, "init")
	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	gh := &fakeGitHub{existing: &github.PullRequest{Number: 42}}
	var out bytes.Buffer
	if err := Run(dir, gh, &out, &out); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if len(gh.updateCalls) != 1 || gh.updateCalls[0].number != 42 {
		t.Fatalf("expected UpdatePullRequest(42, ...), got %v", gh.updateCalls)
	}
	if len(gh.createCalls) != 0 {
		t.Fatalf("expected CreatePullRequest to never be called when a PR already exists, got %v", gh.createCalls)
	}
}

func TestRun_MultipleOpenPRs_ErrorsBeforeCreatingOrUpdating(t *testing.T) {
	dir, _ := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeConfig(t, dir, "main")
	gitCommit(t, dir, "init")
	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	gh := &fakeGitHub{findErr: fmt.Errorf("found 2 open pull requests for head changeset-release/main, base main: expected at most one")}
	var out bytes.Buffer
	if err := Run(dir, gh, &out, &out); err == nil {
		t.Fatalf("expected error when the API reports more than one open pull request")
	}

	if len(gh.createCalls) != 0 || len(gh.updateCalls) != 0 {
		t.Fatalf("expected no create/update call once the invariant is violated, got create=%v update=%v", gh.createCalls, gh.updateCalls)
	}
}

func TestRun_WrongBranch_Errors(t *testing.T) {
	dir, bare := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeConfig(t, dir, "main")
	gitCommit(t, dir, "init")
	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	runGitCmd(t, dir, "checkout", "-q", "-b", "not-main")

	gh := &fakeGitHub{}
	var out bytes.Buffer
	if err := Run(dir, gh, &out, &out); err == nil {
		t.Fatalf("expected error when current branch is not the configured base branch")
	}

	if len(gh.createCalls) != 0 || len(gh.updateCalls) != 0 {
		t.Fatalf("expected no PR API calls, got create=%v update=%v", gh.createCalls, gh.updateCalls)
	}
	if branchOut := runGitCmd(t, bare, "branch", "--list", "changeset-release/main"); strings.TrimSpace(branchOut) != "" {
		t.Fatalf("expected no changeset-release/main branch to be pushed to origin")
	}
}
