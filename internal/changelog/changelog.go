// Package changelog owns the single source of truth for the CHANGELOG.md
// release header format ("## vMAJOR.MINOR.PATCH"): version writes it,
// publish parses it back, and both go through this package so the format
// can never drift between the two.
package changelog

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sabcorecom/changeset/internal/changeset"
)

const (
	changelogTitle = "# Changelog"
	versionCore    = `v\d+\.\d+\.\d+`
)

var (
	versionPattern = regexp.MustCompile(`^` + versionCore + `$`)
	headerPattern  = regexp.MustCompile(`^## (` + versionCore + `)$`)
)

var groupOrder = []struct {
	bump  changeset.Bump
	label string
}{
	{changeset.Major, "Major Changes"},
	{changeset.Minor, "Minor Changes"},
	{changeset.Patch, "Patch Changes"},
}

// PrependRelease inserts a new "## version" release section, grouped by
// bump type into Major/Minor/Patch Changes, above any existing release
// sections in the CHANGELOG.md at path. It creates the file if absent.
func PrependRelease(path, version string, groups map[changeset.Bump][]string) error {
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("invalid version %q: expected format vMAJOR.MINOR.PATCH", version)
	}

	var existingBody string
	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
		content := string(existing)
		content = strings.TrimPrefix(content, changelogTitle)
		existingBody = strings.TrimLeft(content, "\n")
	case os.IsNotExist(err):
		existingBody = ""
	default:
		return fmt.Errorf("read changelog %s: %w", path, err)
	}

	var sb strings.Builder
	sb.WriteString(changelogTitle)
	sb.WriteString("\n\n")
	sb.WriteString(FormatSection(version, groups))
	if existingBody != "" {
		sb.WriteString("\n")
		sb.WriteString(existingBody)
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("write changelog %s: %w", path, err)
	}
	return nil
}

// FormatSection renders a single "## version" release section, grouped by
// bump type into Major/Minor/Patch Changes, without touching any existing
// CHANGELOG.md content.
func FormatSection(version string, groups map[changeset.Bump][]string) string {
	var sb strings.Builder
	sb.WriteString("## " + version + "\n")

	for _, g := range groupOrder {
		entries := groups[g.bump]
		if len(entries) == 0 {
			continue
		}
		sb.WriteString("\n### " + g.label + "\n\n")
		for _, e := range entries {
			sb.WriteString("- " + e + "\n")
		}
	}

	return sb.String()
}

// ParseTopVersion returns the version from the first "## vMAJOR.MINOR.PATCH"
// release header found in the CHANGELOG.md at path.
func ParseTopVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read changelog %s: %w", path, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if m := headerPattern.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("no release header (## vMAJOR.MINOR.PATCH) found in %s", path)
}
