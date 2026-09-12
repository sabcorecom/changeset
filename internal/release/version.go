// Package release holds the bump-and-tag logic shared by the `version`/
// `publish` CLI commands and the CI-only `bot` command: computing the next
// version from accumulated changesets, writing it to CHANGELOG.md, and
// cutting the git tag + goreleaser run once that version is merged.
package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sabcorecom/changeset/internal/changelog"
	"github.com/sabcorecom/changeset/internal/changeset"
	"github.com/sabcorecom/changeset/internal/gitutil"
	"github.com/sabcorecom/changeset/internal/semver"
)

// Bump computes the next version from the latest git tag plus the bump
// implied by accumulated .changeset/*.md files under cwd, writes a
// "## vX.Y.Z" release section to CHANGELOG.md, and deletes the consumed
// changeset files. It returns the empty string for both next and section
// (and a nil error) when there are no changesets to release.
func Bump(cwd string) (next string, section string, err error) {
	files, err := changeset.List(filepath.Join(cwd, ".changeset"))
	if err != nil {
		return "", "", err
	}
	if len(files) == 0 {
		return "", "", nil
	}

	overallBump := files[0].Changeset.Bump
	groups := map[changeset.Bump][]string{}
	for _, f := range files {
		overallBump = changeset.MaxBump(overallBump, f.Changeset.Bump)
		groups[f.Changeset.Bump] = append(groups[f.Changeset.Bump], f.Changeset.Summary)
	}

	current, _, err := gitutil.LatestTag()
	if err != nil {
		return "", "", err
	}

	next, err = semver.Next(current, overallBump)
	if err != nil {
		return "", "", err
	}

	if err := changelog.PrependRelease(filepath.Join(cwd, "CHANGELOG.md"), next, groups); err != nil {
		return "", "", err
	}

	for _, f := range files {
		if err := os.Remove(f.Path); err != nil {
			return "", "", fmt.Errorf("remove changeset file %s: %w", f.Path, err)
		}
	}

	return next, changelog.FormatSection(next, groups), nil
}
