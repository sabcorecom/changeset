package release

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// initRepoWithOrigin is like initRepo but also wires a bare "origin"
// remote so PushTag has somewhere to push to.
func initRepoWithOrigin(t *testing.T) string {
	t.Helper()
	dir := initRepo(t)

	bare := t.TempDir()
	runGitCmd(t, bare, "init", "-q", "--bare")
	runGitCmd(t, dir, "remote", "add", "origin", bare)

	return dir
}

// withFakeGoreleaser installs a fake "goreleaser" script on PATH that
// exits with the given code, so Publish's exec.Command call succeeds
// without depending on the real tool being installed.
func withFakeGoreleaser(t *testing.T, exitCode int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake goreleaser shim is a POSIX shell script")
	}

	binDir := t.TempDir()
	script := filepath.Join(binDir, "goreleaser")
	content := "#!/bin/sh\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake goreleaser: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestPublish_TagsAndRunsGoreleaser(t *testing.T) {
	dir := initRepoWithOrigin(t)
	writeRepoFile(t, dir, "CHANGELOG.md", "# Changelog\n\n## v0.1.0\n\n### Minor Changes\n\n- Add feature.\n")
	gitCommit(t, dir, "add changelog")

	withFakeGoreleaser(t, 0)

	var out bytes.Buffer
	version, err := Publish(dir, &out, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}
	if version != "v0.1.0" {
		t.Fatalf("got version=%q, want v0.1.0", version)
	}

	tagOut := runGitCmd(t, dir, "tag", "--list", "v0.1.0")
	if tagOut == "" {
		t.Fatalf("expected tag v0.1.0 to be created")
	}
}

func TestPublish_FailsWithoutGoreleaserInPath(t *testing.T) {
	dir := initRepoWithOrigin(t)
	writeRepoFile(t, dir, "CHANGELOG.md", "# Changelog\n\n## v0.1.0\n\n### Patch Changes\n\n- Fix.\n")
	gitCommit(t, dir, "add changelog")

	emptyPathDir := t.TempDir()
	t.Setenv("PATH", emptyPathDir)

	var out bytes.Buffer
	if _, err := Publish(dir, &out, &out); err == nil {
		t.Fatalf("expected error when goreleaser is missing from PATH")
	}
}

func TestPublish_FailsWithoutVersionHeader(t *testing.T) {
	dir := initRepoWithOrigin(t)
	writeRepoFile(t, dir, "CHANGELOG.md", "# Changelog\n\nnothing to release\n")
	gitCommit(t, dir, "add changelog")

	withFakeGoreleaser(t, 0)

	var out bytes.Buffer
	if _, err := Publish(dir, &out, &out); err == nil {
		t.Fatalf("expected error when CHANGELOG.md has no release header")
	}
}
