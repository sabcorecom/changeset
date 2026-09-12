package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sabcorecom/changeset/internal/changeset"
)

func TestPrependRelease_CreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")

	err := PrependRelease(path, "v1.0.0", map[changeset.Bump][]string{
		changeset.Major: {"Breaking change."},
		changeset.Patch: {"Fix bug."},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	content := string(data)

	want := "# Changelog\n\n## v1.0.0\n\n### Major Changes\n\n- Breaking change.\n\n### Patch Changes\n\n- Fix bug.\n"
	if content != want {
		t.Fatalf("got:\n%q\nwant:\n%q", content, want)
	}
}

func TestPrependRelease_InsertsAboveExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")

	if err := PrependRelease(path, "v1.0.0", map[changeset.Bump][]string{
		changeset.Patch: {"First release."},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := PrependRelease(path, "v1.1.0", map[changeset.Bump][]string{
		changeset.Minor: {"New feature."},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	content := string(data)

	idxNew := strings.Index(content, "## v1.1.0")
	idxOld := strings.Index(content, "## v1.0.0")
	if idxNew == -1 || idxOld == -1 || idxNew > idxOld {
		t.Fatalf("expected v1.1.0 section above v1.0.0 section, got:\n%s", content)
	}
}

func TestPrependRelease_RejectsInvalidVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	err := PrependRelease(path, "1.0.0", map[changeset.Bump][]string{changeset.Patch: {"x"}})
	if err == nil {
		t.Fatalf("expected error for version missing v prefix")
	}
}

func TestParseTopVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := PrependRelease(path, "v2.3.4", map[changeset.Bump][]string{changeset.Minor: {"x"}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	version, err := ParseTopVersion(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "v2.3.4" {
		t.Fatalf("got %q, want v2.3.4", version)
	}
}

func TestParseTopVersion_ReturnsTopmostAfterMultipleReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := PrependRelease(path, "v1.0.0", map[changeset.Bump][]string{changeset.Patch: {"x"}}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := PrependRelease(path, "v2.0.0", map[changeset.Bump][]string{changeset.Major: {"y"}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	version, err := ParseTopVersion(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "v2.0.0" {
		t.Fatalf("got %q, want v2.0.0", version)
	}
}

func TestParseTopVersion_ErrorsWithoutHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(path, []byte("# Changelog\n\nno release headers here\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := ParseTopVersion(path); err == nil {
		t.Fatalf("expected error when no version header present")
	}
}

func TestParseTopVersion_ErrorsOnMissingFile(t *testing.T) {
	if _, err := ParseTopVersion(filepath.Join(t.TempDir(), "missing.md")); err == nil {
		t.Fatalf("expected error for missing file")
	}
}
