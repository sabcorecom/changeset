package github

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient("test-token", "acme", "widgets", srv.URL), srv
}

func TestFindOpenPullRequest_None(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("got method %s, want GET", r.Method)
		}
		if r.URL.Path != "/repos/acme/widgets/pulls" {
			t.Fatalf("got path %s, want /repos/acme/widgets/pulls", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("head") != "acme:changeset-release/main" {
			t.Fatalf("got head=%q, want acme:changeset-release/main", q.Get("head"))
		}
		if q.Get("base") != "main" {
			t.Fatalf("got base=%q, want main", q.Get("base"))
		}
		if q.Get("state") != "open" {
			t.Fatalf("got state=%q, want open", q.Get("state"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("got Authorization=%q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-GitHub-Api-Version") != apiVersion {
			t.Fatalf("got X-GitHub-Api-Version=%q, want %q", r.Header.Get("X-GitHub-Api-Version"), apiVersion)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	pr, err := client.FindOpenPullRequest("changeset-release/main", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr != nil {
		t.Fatalf("got %+v, want nil", pr)
	}
}

func TestFindOpenPullRequest_One(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"number": 42, "html_url": "https://github.com/acme/widgets/pull/42"}]`))
	})

	pr, err := client.FindOpenPullRequest("changeset-release/main", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr == nil || pr.Number != 42 {
		t.Fatalf("got %+v, want number=42", pr)
	}
}

func TestFindOpenPullRequest_MultipleIsError(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"number": 1}, {"number": 2}]`))
	})

	pr, err := client.FindOpenPullRequest("changeset-release/main", "main")
	if err == nil {
		t.Fatalf("expected error when 2+ open PRs are found, got pr=%+v", pr)
	}
}

func TestCreatePullRequest(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("got method %s, want POST", r.Method)
		}
		if r.URL.Path != "/repos/acme/widgets/pulls" {
			t.Fatalf("got path %s, want /repos/acme/widgets/pulls", r.URL.Path)
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var got map[string]string
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		want := map[string]string{
			"title": "chore: version packages",
			"body":  "release notes",
			"head":  "changeset-release/main",
			"base":  "main",
		}
		for k, v := range want {
			if got[k] != v {
				t.Fatalf("got %s=%q, want %q", k, got[k], v)
			}
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number": 7, "html_url": "https://github.com/acme/widgets/pull/7"}`))
	})

	pr, err := client.CreatePullRequest("changeset-release/main", "main", "chore: version packages", "release notes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.Number != 7 {
		t.Fatalf("got number=%d, want 7", pr.Number)
	}
}

func TestUpdatePullRequest(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("got method %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/repos/acme/widgets/pulls/7" {
			t.Fatalf("got path %s, want /repos/acme/widgets/pulls/7", r.URL.Path)
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var got map[string]string
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if got["title"] != "chore: version packages" || got["body"] != "updated notes" {
			t.Fatalf("got body=%+v", got)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"number": 7}`))
	})

	if err := client.UpdatePullRequest(7, "chore: version packages", "updated notes"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_ErrorStatusIsReturnedAsError(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message": "Validation Failed"}`))
	})

	if _, err := client.CreatePullRequest("head", "base", "title", "body"); err == nil {
		t.Fatalf("expected error for non-2xx response")
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("GITHUB_REPOSITORY", "acme/widgets")

		if _, err := NewClientFromEnv(); err == nil {
			t.Fatalf("expected error when GITHUB_TOKEN is unset")
		}
	})

	t.Run("missing repository", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "tok")
		t.Setenv("GITHUB_REPOSITORY", "")

		if _, err := NewClientFromEnv(); err == nil {
			t.Fatalf("expected error when GITHUB_REPOSITORY is unset")
		}
	})

	t.Run("malformed repository", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "tok")
		t.Setenv("GITHUB_REPOSITORY", "not-owner-slash-repo")

		if _, err := NewClientFromEnv(); err == nil {
			t.Fatalf("expected error when GITHUB_REPOSITORY has no owner/repo slash")
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "tok")
		t.Setenv("GITHUB_REPOSITORY", "acme/widgets")
		t.Setenv("GITHUB_API_URL", "")

		client, err := NewClientFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.owner != "acme" || client.repo != "widgets" || client.token != "tok" {
			t.Fatalf("got %+v", client)
		}
		if client.baseURL != defaultBaseURL {
			t.Fatalf("got baseURL=%q, want default %q", client.baseURL, defaultBaseURL)
		}
	})

	t.Run("custom API base URL (GHES)", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "tok")
		t.Setenv("GITHUB_REPOSITORY", "acme/widgets")
		t.Setenv("GITHUB_API_URL", "https://ghes.example.com/api/v3")

		client, err := NewClientFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.baseURL != "https://ghes.example.com/api/v3" {
			t.Fatalf("got baseURL=%q, want the overridden GITHUB_API_URL", client.baseURL)
		}
	})
}
