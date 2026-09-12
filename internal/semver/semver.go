// Package semver parses and increments the strict vMAJOR.MINOR.PATCH git
// tag format used as the single source of truth for this repository's
// version. There is no version file: only git tags carry the version.
package semver

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/sabcorecom/changeset/internal/changeset"
)

var tagPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// ParseTag parses a strict "vMAJOR.MINOR.PATCH" tag. Any deviation
// (missing "v" prefix, pre-release/build suffix, non-numeric parts) is a
// hard error rather than a best-effort truncated parse.
func ParseTag(tag string) (major, minor, patch int, err error) {
	m := tagPattern.FindStringSubmatch(tag)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("invalid version tag %q: expected format vMAJOR.MINOR.PATCH", tag)
	}

	major, err = strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version in tag %q: %w", tag, err)
	}
	minor, err = strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version in tag %q: %w", tag, err)
	}
	patch, err = strconv.Atoi(m[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version in tag %q: %w", tag, err)
	}
	return major, minor, patch, nil
}

// Next computes the next version tag given the current tag (empty string
// meaning no tags exist yet, i.e. a base of v0.0.0) and a bump type.
func Next(current string, bump changeset.Bump) (string, error) {
	if err := bump.Validate(); err != nil {
		return "", err
	}

	major, minor, patch := 0, 0, 0
	if current != "" {
		var err error
		major, minor, patch, err = ParseTag(current)
		if err != nil {
			return "", err
		}
	}

	switch bump {
	case changeset.Major:
		major++
		minor, patch = 0, 0
	case changeset.Minor:
		minor++
		patch = 0
	case changeset.Patch:
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}
