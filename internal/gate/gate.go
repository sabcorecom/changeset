// Package gate implements the PR gate policy used by `changeset status`:
// fail the diff unless it either only touches docs, is on the exempt
// Version PR branch, or includes a new changeset file.
package gate

import (
	"fmt"
	"path/filepath"
	"strings"
)

const exemptBranch = "changeset-release/main"

// IsExemptBranch reports whether branch is the Version PR bot's branch,
// which is exempt from the changeset gate because it is itself the
// commit that consumes accumulated changesets.
func IsExemptBranch(branch string) bool {
	return branch == exemptBranch
}

// IsDocOnly reports whether every changed file is a markdown document.
// An empty file list is vacuously doc-only: there is nothing to gate.
func IsDocOnly(files []string) bool {
	for _, f := range files {
		if !strings.HasSuffix(f, ".md") {
			return false
		}
	}
	return true
}

func hasNewChangesetFile(files []string) bool {
	for _, f := range files {
		dir, name := filepath.Split(f)
		if filepath.Clean(dir) != ".changeset" {
			continue
		}
		if name == "README.md" {
			continue
		}
		if strings.HasSuffix(name, ".md") {
			return true
		}
	}
	return false
}

// Decide applies the gate policy for a diff against changedFiles on
// branch, returning whether it passes and a human-readable reason.
func Decide(branch string, changedFiles []string) (pass bool, reason string) {
	if IsExemptBranch(branch) {
		return true, fmt.Sprintf("branch %q is exempt from the changeset gate", branch)
	}
	if IsDocOnly(changedFiles) {
		return true, "all changed files are docs-only"
	}
	if hasNewChangesetFile(changedFiles) {
		return true, "a changeset file was added"
	}
	return false, "no changeset file found for non-docs changes; run `changeset add`"
}
