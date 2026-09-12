// Package gitutil wraps the subset of git plumbing commands needed to
// read the repository's version state and cut releases. It shells out to
// the git binary rather than reimplementing git internals.
package gitutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// LatestTag returns the highest version-sorted tag matching "v*". It
// returns ("", false, nil) if the repository has no such tags — that is
// a legitimate state, not an error.
func LatestTag() (string, bool, error) {
	out, err := runGit("tag", "--list", "v*", "--sort=-v:refname")
	if err != nil {
		return "", false, fmt.Errorf("list git tags: %w", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return "", false, nil
	}

	lines := strings.Split(trimmed, "\n")
	return lines[0], true, nil
}

// ChangedFiles returns the files changed on HEAD relative to the merge
// base with since (e.g. "origin/main"), matching what a PR diff would
// show.
func ChangedFiles(since string) ([]string, error) {
	out, err := runGit("diff", "--name-only", since+"...HEAD")
	if err != nil {
		return nil, fmt.Errorf("diff against %s: %w", since, err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// CurrentBranch returns the checked-out branch name.
func CurrentBranch() (string, error) {
	out, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CreateTag creates a lightweight tag v pointing at HEAD. It is
// idempotent: re-running it when v already points at HEAD is a no-op.
// If v exists and points elsewhere, that is an error rather than a
// silent overwrite.
func CreateTag(v string) error {
	head, err := runGit("rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("resolve HEAD: %w", err)
	}
	head = strings.TrimSpace(head)

	existing, err := runGit("rev-parse", "--verify", "refs/tags/"+v)
	if err == nil {
		existing = strings.TrimSpace(existing)
		if existing == head {
			return nil
		}
		return fmt.Errorf("tag %s already exists and points at %s, not HEAD (%s)", v, existing, head)
	}

	if _, err := runGit("tag", v, head); err != nil {
		return fmt.Errorf("create tag %s: %w", v, err)
	}
	return nil
}

// PushTag pushes tag v to the "origin" remote.
func PushTag(v string) error {
	if _, err := runGit("push", "origin", v); err != nil {
		return fmt.Errorf("push tag %s: %w", v, err)
	}
	return nil
}

// CreateOrResetBranch checks out branch name, creating it if absent or
// resetting it to HEAD if it already exists.
func CreateOrResetBranch(name string) error {
	if _, err := runGit("checkout", "-B", name); err != nil {
		return fmt.Errorf("checkout branch %s: %w", name, err)
	}
	return nil
}

// CommitAll stages every change in the working tree and commits it with
// message. There is no "nothing to commit" special-casing: callers only
// invoke this after a real change, so an empty diff is a caller bug and
// should surface as an error rather than be silently skipped.
func CommitAll(message string) error {
	if _, err := runGit("add", "-A"); err != nil {
		return fmt.Errorf("stage changes: %w", err)
	}
	if _, err := runGit("commit", "-m", message); err != nil {
		return fmt.Errorf("commit %q: %w", message, err)
	}
	return nil
}

// ForcePushBranch force-pushes the local branch name to the "origin"
// remote's branch of the same name.
func ForcePushBranch(name string) error {
	if _, err := runGit("push", "--force", "origin", name+":"+name); err != nil {
		return fmt.Errorf("force-push branch %s: %w", name, err)
	}
	return nil
}
