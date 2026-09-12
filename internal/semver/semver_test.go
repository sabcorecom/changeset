package semver

import (
	"testing"

	"github.com/sabcorecom/changeset/internal/changeset"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag                 string
		major, minor, patch int
		wantErr             bool
	}{
		{tag: "v1.2.3", major: 1, minor: 2, patch: 3},
		{tag: "v0.0.0", major: 0, minor: 0, patch: 0},
		{tag: "v0.10.0", major: 0, minor: 10, patch: 0},
		{tag: "1.2.3", wantErr: true},
		{tag: "v1.2.3-rc1", wantErr: true},
		{tag: "v1.2", wantErr: true},
		{tag: "vabc.2.3", wantErr: true},
		{tag: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			major, minor, patch, err := ParseTag(tt.tag)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for tag %q", tt.tag)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Fatalf("got %d.%d.%d, want %d.%d.%d", major, minor, patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		name    string
		current string
		bump    changeset.Bump
		want    string
		wantErr bool
	}{
		{name: "no tags, major", current: "", bump: changeset.Major, want: "v1.0.0"},
		{name: "no tags, minor", current: "", bump: changeset.Minor, want: "v0.1.0"},
		{name: "no tags, patch", current: "", bump: changeset.Patch, want: "v0.0.1"},
		{name: "patch bump", current: "v1.2.3", bump: changeset.Patch, want: "v1.2.4"},
		{name: "minor rollover not major", current: "v1.9.9", bump: changeset.Minor, want: "v1.10.0"},
		{name: "minor bump resets patch", current: "v1.2.3", bump: changeset.Minor, want: "v1.3.0"},
		{name: "major bump resets minor and patch", current: "v1.2.3", bump: changeset.Major, want: "v2.0.0"},
		{name: "invalid current tag", current: "v1.2.3-rc1", bump: changeset.Patch, wantErr: true},
		{name: "invalid bump", current: "v1.2.3", bump: changeset.Bump("nonsense"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Next(tt.current, tt.bump)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
