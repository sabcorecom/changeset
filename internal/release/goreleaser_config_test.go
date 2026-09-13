package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasGoreleaserConfig(t *testing.T) {
	tests := []struct {
		name       string
		writeFiles map[string]string
		want       bool
	}{
		{
			name:       "no config file",
			writeFiles: nil,
			want:       false,
		},
		{
			name:       "dot config yml",
			writeFiles: map[string]string{".config/goreleaser.yml": ""},
			want:       true,
		},
		{
			name:       "dot config yaml",
			writeFiles: map[string]string{".config/goreleaser.yaml": ""},
			want:       true,
		},
		{
			name:       "dotfile yml",
			writeFiles: map[string]string{".goreleaser.yml": ""},
			want:       true,
		},
		{
			name:       "dotfile yaml",
			writeFiles: map[string]string{".goreleaser.yaml": ""},
			want:       true,
		},
		{
			name:       "plain yml",
			writeFiles: map[string]string{"goreleaser.yml": ""},
			want:       true,
		},
		{
			name:       "plain yaml",
			writeFiles: map[string]string{"goreleaser.yaml": ""},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for relPath, content := range tt.writeFiles {
				full := filepath.Join(dir, relPath)
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatalf("mkdir for %s: %v", relPath, err)
				}
				if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
					t.Fatalf("write %s: %v", relPath, err)
				}
			}

			got, err := hasGoreleaserConfig(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("hasGoreleaserConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasGoreleaserConfig_StatError(t *testing.T) {
	dir := t.TempDir()

	// ".config" is a regular file, not a directory, so stat-ing
	// ".config/goreleaser.yml" underneath it fails with ENOTDIR rather
	// than ENOENT — a real error os.IsNotExist does not match, which
	// hasGoreleaserConfig must surface instead of treating as "absent".
	if err := os.WriteFile(filepath.Join(dir, ".config"), []byte(""), 0o644); err != nil {
		t.Fatalf("write .config file: %v", err)
	}

	_, err := hasGoreleaserConfig(dir)
	if err == nil {
		t.Fatalf("expected error when a config candidate's parent is not a directory")
	}
}
