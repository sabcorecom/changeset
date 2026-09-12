package changeset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Changeset
		wantErr bool
	}{
		{
			name: "single package patch",
			input: "---\n" +
				"\"leap-autotests\": patch\n" +
				"---\n\n" +
				"Increase retry budgets.\n",
			want: &Changeset{Bump: Patch, Summary: "Increase retry budgets."},
		},
		{
			name: "multiple packages takes max bump",
			input: "---\n" +
				"\"pkg-a\": patch\n" +
				"\"pkg-b\": major\n" +
				"---\n\n" +
				"Breaking change.\n",
			want: &Changeset{Bump: Major, Summary: "Breaking change."},
		},
		{
			name:    "missing opening delimiter",
			input:   "\"pkg\": patch\n---\n\nSummary\n",
			wantErr: true,
		},
		{
			name:    "missing closing delimiter",
			input:   "---\n\"pkg\": patch\n\nSummary\n",
			wantErr: true,
		},
		{
			name:    "empty frontmatter map",
			input:   "---\n---\n\nSummary\n",
			wantErr: true,
		},
		{
			name:    "invalid bump value",
			input:   "---\n\"pkg\": foo\n---\n\nSummary\n",
			wantErr: true,
		},
		{
			name:    "empty summary",
			input:   "---\n\"pkg\": patch\n---\n\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Bump != tt.want.Bump || got.Summary != tt.want.Summary {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("setup write %s: %v", name, err)
		}
	}

	write("config.json", `{"packageName":"x"}`)
	write("README.md", "# ignored on purpose, not a changeset")
	write("a1b2c3.md", "---\n\"pkg\": patch\n---\n\nFix bug.\n")
	write("d4e5f6.md", "---\n\"pkg\": minor\n---\n\nAdd feature.\n")

	files, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 changeset files, got %d: %+v", len(files), files)
	}
}

func TestList_FailsFastOnMalformedFile(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte("not a changeset"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := List(dir); err == nil {
		t.Fatalf("expected List to fail on malformed changeset file, got nil error")
	}
}

func TestList_MissingDirIsEmpty(t *testing.T) {
	files, err := List(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty list, got %+v", files)
	}
}

func TestWrite_RoundTrips(t *testing.T) {
	dir := t.TempDir()

	name, err := Write(dir, "my-package", Minor, "Add a new thing.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cs, err := ParseFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("failed to parse written changeset: %v", err)
	}
	if cs.Bump != Minor || cs.Summary != "Add a new thing." {
		t.Fatalf("got %+v", cs)
	}
}

func TestWrite_UniqueNames(t *testing.T) {
	dir := t.TempDir()

	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		name, err := Write(dir, "pkg", Patch, "change")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[name] {
			t.Fatalf("duplicate changeset filename: %s", name)
		}
		seen[name] = true
	}
}

func TestWrite_RejectsInvalidBump(t *testing.T) {
	if _, err := Write(t.TempDir(), "pkg", Bump("nonsense"), "summary"); err == nil {
		t.Fatalf("expected error for invalid bump")
	}
}

func TestWrite_RejectsEmptySummary(t *testing.T) {
	if _, err := Write(t.TempDir(), "pkg", Patch, "   "); err == nil {
		t.Fatalf("expected error for empty summary")
	}
}

func TestMaxBump(t *testing.T) {
	if MaxBump(Patch, Major) != Major {
		t.Fatalf("expected major to win")
	}
	if MaxBump(Minor, Patch) != Minor {
		t.Fatalf("expected minor to win")
	}
	if MaxBump(Patch, Patch) != Patch {
		t.Fatalf("expected patch")
	}
}
