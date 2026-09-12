package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestBotCmd_ChangesetsPending_CreatesVersionPR(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "go.mod", "module example.com/pkg\n\ngo 1.25\n")
	writeRepoFile(t, dir, ".changeset/config.json", "{\"packageName\":\"pkg\",\"baseBranch\":\"main\"}\n")
	gitCommit(t, dir, "init")
	writeRepoFile(t, dir, ".changeset/a.md", "---\n\"pkg\": patch\n---\n\nFix one.\n")

	var createCalled atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/pulls"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pulls"):
			createCalled.Store(true)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"number": 1, "html_url": "https://example.invalid/pull/1"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_REPOSITORY", "acme/widgets")
	t.Setenv("GITHUB_API_URL", srv.URL)

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"bot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}

	if !createCalled.Load() {
		t.Fatalf("expected the fake GitHub API to receive a CreatePullRequest call")
	}
}
