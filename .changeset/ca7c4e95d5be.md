---
"changeset": minor
---

Add `changeset bot`, a CI-only command that maintains the single Version PR on push to the base branch (regenerating `changeset-release/main` from accumulated changesets) or runs `changeset publish` when that PR was just merged.
