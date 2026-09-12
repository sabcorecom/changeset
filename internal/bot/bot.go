// Package bot is the CI-only entrypoint for the Version PR bot: on a push
// to the base branch, it either regenerates the single, always-reused
// changeset-release/main branch and opens or updates its pull request
// (when changesets are pending), or triggers the publish flow (when the
// push is the just-merged Version PR itself, so there is nothing left to
// version). It is invoked by the composite GitHub Action as `changeset bot`.
package bot

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/sabcorecom/changeset/internal/changeset"
	"github.com/sabcorecom/changeset/internal/config"
	"github.com/sabcorecom/changeset/internal/gate"
	"github.com/sabcorecom/changeset/internal/github"
	"github.com/sabcorecom/changeset/internal/gitutil"
	"github.com/sabcorecom/changeset/internal/release"
)

// CommitTitle is used both as the commit message on the Version PR branch
// and as the pull request title, per epic RFC §3.
const CommitTitle = "chore: version packages"

// pullRequestAPI is the subset of internal/github.Client the bot needs,
// defined here at its point of use so bot depends on an interface rather
// than a concrete HTTP client; *github.Client satisfies it structurally.
type pullRequestAPI interface {
	FindOpenPullRequest(head, base string) (*github.PullRequest, error)
	CreatePullRequest(head, base, title, body string) (*github.PullRequest, error)
	UpdatePullRequest(number int, title, body string) error
}

// Run executes one bot decision: publish if there are no pending
// changesets under cwd, otherwise regenerate the Version PR branch and
// open or update its pull request via gh.
func Run(cwd string, gh pullRequestAPI, stdout, stderr io.Writer) error {
	files, err := changeset.List(filepath.Join(cwd, ".changeset"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		_, err := release.Publish(cwd, stdout, stderr)
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	branch, err := gitutil.CurrentBranch()
	if err != nil {
		return err
	}
	if branch != cfg.BaseBranch {
		return fmt.Errorf("bot must run on base branch %q, got %q", cfg.BaseBranch, branch)
	}

	if err := gitutil.CreateOrResetBranch(gate.VersionPRBranch); err != nil {
		return err
	}

	_, section, err := release.Bump(cwd)
	if err != nil {
		return err
	}

	if err := gitutil.CommitAll(CommitTitle); err != nil {
		return err
	}
	if err := gitutil.ForcePushBranch(gate.VersionPRBranch); err != nil {
		return err
	}

	existing, err := gh.FindOpenPullRequest(gate.VersionPRBranch, cfg.BaseBranch)
	if err != nil {
		return err
	}

	if existing != nil {
		return gh.UpdatePullRequest(existing.Number, CommitTitle, section)
	}

	_, err = gh.CreatePullRequest(gate.VersionPRBranch, cfg.BaseBranch, CommitTitle, section)
	return err
}
