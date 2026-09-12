// Package changeset parses and writes .changeset/*.md files in the
// format used by @changesets/cli: a YAML frontmatter map of package
// name to bump type, followed by a free-text summary.
package changeset

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Bump is a semantic version bump type.
type Bump string

const (
	Patch Bump = "patch"
	Minor Bump = "minor"
	Major Bump = "major"
)

// Validate reports whether b is one of the known bump types.
func (b Bump) Validate() error {
	switch b {
	case Patch, Minor, Major:
		return nil
	default:
		return fmt.Errorf("invalid bump %q: must be one of patch, minor, major", string(b))
	}
}

func (b Bump) rank() int {
	switch b {
	case Major:
		return 3
	case Minor:
		return 2
	case Patch:
		return 1
	default:
		return 0
	}
}

// MaxBump returns the higher-priority bump of a and b (major > minor > patch).
func MaxBump(a, b Bump) Bump {
	if a.rank() >= b.rank() {
		return a
	}
	return b
}

// Changeset is a single parsed changeset entry.
type Changeset struct {
	Bump    Bump
	Summary string
}

// File is a changeset file together with its parsed contents.
type File struct {
	Path      string
	Changeset *Changeset
}

// Parse reads a changeset document from r.
func Parse(r io.Reader) (*Changeset, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read changeset: %w", err)
	}
	return parseBytes(data)
}

func parseBytes(data []byte) (*Changeset, error) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")

	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("missing opening frontmatter delimiter (---)")
	}

	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return nil, fmt.Errorf("missing closing frontmatter delimiter (---)")
	}

	frontmatter := strings.Join(lines[1:closeIdx], "\n")
	summary := strings.TrimSpace(strings.Join(lines[closeIdx+1:], "\n"))

	var bumps map[string]string
	if err := yaml.Unmarshal([]byte(frontmatter), &bumps); err != nil {
		return nil, fmt.Errorf("parse frontmatter yaml: %w", err)
	}
	if len(bumps) == 0 {
		return nil, fmt.Errorf("frontmatter must contain at least one package entry")
	}

	var result Bump
	first := true
	for name, raw := range bumps {
		b := Bump(raw)
		if err := b.Validate(); err != nil {
			return nil, fmt.Errorf("package %q: %w", name, err)
		}
		if first {
			result = b
			first = false
		} else {
			result = MaxBump(result, b)
		}
	}

	if summary == "" {
		return nil, fmt.Errorf("changeset summary must not be empty")
	}

	return &Changeset{Bump: result, Summary: summary}, nil
}

// ParseFile reads and parses a single changeset file at path.
func ParseFile(path string) (*Changeset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open changeset file %s: %w", path, err)
	}
	defer f.Close()

	cs, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cs, nil
}

// List returns every changeset file in dir, parsed. It fails fast on the
// first malformed .md file rather than skipping it. A missing dir is not
// an error: it is treated as an empty changeset list.
func List(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read changeset dir %s: %w", dir, err)
	}

	var files []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "config.json" || name == "README.md" {
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		path := filepath.Join(dir, name)
		cs, err := ParseFile(path)
		if err != nil {
			return nil, err
		}
		files = append(files, File{Path: path, Changeset: cs})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

// Write creates a new uniquely-named changeset file in dir for pkgName,
// returning the created file's base name.
func Write(dir, pkgName string, bump Bump, summary string) (string, error) {
	if err := bump.Validate(); err != nil {
		return "", err
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return "", fmt.Errorf("changeset summary must not be empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create changeset dir %s: %w", dir, err)
	}

	const maxAttempts = 10
	var path, name string
	for i := 0; i < maxAttempts; i++ {
		candidate, err := randomName()
		if err != nil {
			return "", err
		}
		p := filepath.Join(dir, candidate+".md")
		if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
			path, name = p, candidate+".md"
			break
		}
	}
	if path == "" {
		return "", fmt.Errorf("failed to generate a unique changeset filename after %d attempts", maxAttempts)
	}

	content := fmt.Sprintf("---\n%q: %s\n---\n\n%s\n", pkgName, bump, summary)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write changeset file %s: %w", path, err)
	}
	return name, nil
}
