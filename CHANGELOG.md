# Changelog

## v0.1.0

### Minor Changes

- Add goreleaser config and CI workflows wiring this repo to its own Version PR bot and PR gate
- Add `action/action.yml`, a composite GitHub Action (`sabcorecom/changeset/action@v1`) that installs the `changeset` binary via `go install` and runs `changeset bot` as a single CI step.
- Add `changeset bot`, a CI-only command that maintains the single Version PR on push to the base branch (regenerating `changeset-release/main` from accumulated changesets) or runs `changeset publish` when that PR was just merged.

### Patch Changes

- `changeset publish` now only requires and runs `goreleaser` when the target repo has a goreleaser config file (one of the paths goreleaser itself looks for by default, e.g. `.goreleaser.yaml`). Repos without one are tagged and pushed without needing goreleaser installed.
