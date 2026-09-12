package gate

import "testing"

func TestIsDocOnly(t *testing.T) {
	if !IsDocOnly([]string{"README.md", "docs/x.md"}) {
		t.Fatalf("expected docs-only files to pass")
	}
	if IsDocOnly([]string{"main.go", "README.md"}) {
		t.Fatalf("expected non-docs file to fail doc-only check")
	}
}

func TestIsExemptBranch(t *testing.T) {
	if !IsExemptBranch("changeset-release/main") {
		t.Fatalf("expected changeset-release/main to be exempt")
	}
	if IsExemptBranch("feature/x") {
		t.Fatalf("expected feature/x to not be exempt")
	}
}

func TestDecide(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		files    []string
		wantPass bool
	}{
		{
			name:     "exempt branch passes regardless of diff",
			branch:   "changeset-release/main",
			files:    []string{"main.go"},
			wantPass: true,
		},
		{
			name:     "docs only passes",
			branch:   "feature/x",
			files:    []string{"README.md"},
			wantPass: true,
		},
		{
			name:     "code change with changeset file passes",
			branch:   "feature/x",
			files:    []string{"main.go", ".changeset/a1b2c3.md"},
			wantPass: true,
		},
		{
			name:     "code change without changeset file fails",
			branch:   "feature/x",
			files:    []string{"main.go"},
			wantPass: false,
		},
		{
			name:     "changeset README does not count as a changeset file",
			branch:   "feature/x",
			files:    []string{"main.go", ".changeset/README.md"},
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, reason := Decide(tt.branch, tt.files)
			if pass != tt.wantPass {
				t.Fatalf("got pass=%v reason=%q, want pass=%v", pass, reason, tt.wantPass)
			}
			if reason == "" {
				t.Fatalf("expected non-empty reason")
			}
		})
	}
}
