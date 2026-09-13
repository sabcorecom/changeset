package release

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sabcorecom/changeset/internal/changelog"
	"github.com/sabcorecom/changeset/internal/gitutil"
)

// goreleaserConfigCandidates are the config file paths goreleaser itself
// probes by default, in the same order, per goreleaser v2's
// loadConfigCheck (cmd/config.go on goreleaser/goreleaser@main). Keeping
// this list in sync with goreleaser's own search order is what makes
// "config file present" a reliable proxy for "this repo wants a
// goreleaser release" (see task RFC section B).
var goreleaserConfigCandidates = []string{
	".config/goreleaser.yml",
	".config/goreleaser.yaml",
	".goreleaser.yml",
	".goreleaser.yaml",
	"goreleaser.yml",
	"goreleaser.yaml",
}

// hasGoreleaserConfig reports whether cwd contains a goreleaser config
// file at one of goreleaser's default search paths.
func hasGoreleaserConfig(cwd string) (bool, error) {
	for _, candidate := range goreleaserConfigCandidates {
		_, err := os.Stat(filepath.Join(cwd, candidate))
		if err == nil {
			return true, nil
		}
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("stat %s: %w", candidate, err)
		}
	}
	return false, nil
}

// Publish parses the top "## vX.Y.Z" header from CHANGELOG.md under cwd,
// creates and pushes the matching git tag, then — only if cwd contains a
// goreleaser config file — runs `goreleaser release --clean` with
// stdout/stderr wired to stdout/stderr. Repos with no goreleaser config
// are tagged and pushed without requiring goreleaser to be installed.
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

	hasConfig, err := hasGoreleaserConfig(cwd)
	if err != nil {
		return "", err
	}
	if !hasConfig {
		return version, nil
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
