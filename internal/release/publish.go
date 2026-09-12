package release

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/sabcorecom/changeset/internal/changelog"
	"github.com/sabcorecom/changeset/internal/gitutil"
)

// Publish parses the top "## vX.Y.Z" header from CHANGELOG.md under cwd,
// creates and pushes the matching git tag, then runs
// `goreleaser release --clean` with stdout/stderr wired to stdout/stderr.
func Publish(cwd string, stdout, stderr io.Writer) (version string, err error) {
	version, err = changelog.ParseTopVersion(filepath.Join(cwd, "CHANGELOG.md"))
	if err != nil {
		return "", err
	}

	if err := gitutil.CreateTag(version); err != nil {
		return "", err
	}
	if err := gitutil.PushTag(version); err != nil {
		return "", err
	}

	if _, err := exec.LookPath("goreleaser"); err != nil {
		return "", fmt.Errorf("goreleaser not found in PATH: %w", err)
	}

	releaseCmd := exec.Command("goreleaser", "release", "--clean")
	releaseCmd.Dir = cwd
	releaseCmd.Stdout = stdout
	releaseCmd.Stderr = stderr
	if err := releaseCmd.Run(); err != nil {
		return "", fmt.Errorf("run goreleaser: %w", err)
	}

	return version, nil
}
