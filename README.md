# changeset

A Go tool for managing versioned changelogs from per-change changeset
files, in the spirit of [`@changesets/cli`](https://github.com/changesets/changesets)
but without a Node.js dependency. The repository's version lives only in
git tags — there is no `package.json`-style version file.

## Commands

- `changeset init` — create `.changeset/config.json` (requires a `go.mod`
  in the current directory, used to infer the package name).
- `changeset add` — create a new changeset file. Prompts interactively
  for a bump type (`patch`/`minor`/`major`) and a summary, or accepts
  `--bump` and `--message` flags for non-interactive use.
- `changeset status --since=<ref>` — fail if the diff against `<ref>`
  contains non-docs changes without a changeset file. Exempts docs-only
  diffs and the `changeset-release/main` branch. Intended for a CI gate:
  `changeset status --since=origin/main`.
- `changeset version` — compute the next version from the latest git tag
  plus the bump implied by accumulated `.changeset/*.md` files, write a
  `## vX.Y.Z` release section to `CHANGELOG.md` (grouped into Major/Minor/
  Patch Changes), and delete the consumed changeset files.
- `changeset publish` — parse the top `## vX.Y.Z` header from
  `CHANGELOG.md`, create and push the matching git tag, then run
  `goreleaser release --clean`.
- `changeset bot` — **CI-only**, not meant to be run manually. On a push
  to the base branch: if `.changeset/*.md` files are accumulated,
  regenerate the `changeset-release/main` branch (`changeset version`'s
  bump, committed as `chore: version packages`, force-pushed) and open or
  update the single Version PR on that branch; if none are accumulated,
  the push is the just-merged Version PR itself, so run `changeset publish`
  instead. Requires `GITHUB_TOKEN` (standard Actions token with
  `contents: write` and `pull-requests: write`, no custom PAT) and
  `GITHUB_REPOSITORY`, both of which a GitHub Actions workflow sets
  automatically. See [GitHub Action](#github-action) below for the
  recommended way to run it in CI.

## GitHub Action

`action/action.yml` in this repo is a composite GitHub Action, referenced
by consumers as `sabcorecom/changeset/action@v1`. It installs the
`changeset` binary via `go install` and runs `changeset bot` as a single
step — no separate install/build step needed:

```yaml
name: changeset
on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

concurrency:
  group: changeset-bot-${{ github.ref }}
  cancel-in-progress: false

jobs:
  version-or-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: sabcorecom/changeset/action@v1
```

Inputs (both optional):

- `github-token` — token used for the Version PR. Defaults to
  `${{ github.token }}`, which is sufficient; no custom PAT is needed.
- `version` — git ref of this module that `go install` fetches. Defaults
  to the ref this action was invoked at.

Prerequisites the action does not provide, and the consumer must set up:

- `actions/checkout` must run first, with `fetch-depth: 0` (full tag
  history is required to compute the next version from git tags) and the
  default `persist-credentials: true` (needed to push the Version PR
  branch and, later, release tags).
- If/when the publish path runs, `goreleaser` must already be installed
  and configured in the job — this action does not install it (fails
  fast with a clear error otherwise).
- The `concurrency` block shown above is recommended so two overlapping
  bot runs don't race on the same Version PR branch.

The action only wraps `changeset bot`. The PR-gate check
(`changeset status --since=origin/main`) is a separate CI step the
consumer adds on their own, not part of this action.

## Building

```sh
make build   # builds bin/changeset with version info from git
make test    # runs the test suite
make check   # build + vet + lint + test
```

Version, commit, and build date are injected into the binary via
`-ldflags` at build time (see `Makefile`); nothing is stored in the repo.
