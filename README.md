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

## Building

```sh
make build   # builds bin/changeset with version info from git
make test    # runs the test suite
make check   # build + vet + lint + test
```

Version, commit, and build date are injected into the binary via
`-ldflags` at build time (see `Makefile`); nothing is stored in the repo.
