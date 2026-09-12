package cli

import (
	"bytes"
	"testing"
)

func TestStatusCmd_FailsWithoutChangeset(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "main.go", "package main\n")
	gitCommit(t, dir, "add main.go")
	runGitCmd(t, dir, "branch", "feature")
	runGitCmd(t, dir, "checkout", "-q", "feature")
	writeRepoFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	gitCommit(t, dir, "update main.go")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--since=main"})

	if err := root.Execute(); err == nil {
		t.Fatalf("expected status to fail without a changeset file")
	}
}

func TestStatusCmd_PassesWithChangeset(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "main.go", "package main\n")
	gitCommit(t, dir, "add main.go")
	runGitCmd(t, dir, "branch", "feature")
	runGitCmd(t, dir, "checkout", "-q", "feature")
	writeRepoFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeRepoFile(t, dir, ".changeset/abc123.md", "---\n\"pkg\": patch\n---\n\nUpdate main.\n")
	gitCommit(t, dir, "update main and add changeset")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--since=main"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out.String())
	}
}

func TestStatusCmd_PassesOnExemptBranch(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "main.go", "package main\n")
	gitCommit(t, dir, "add main.go")
	runGitCmd(t, dir, "branch", "changeset-release/main")
	runGitCmd(t, dir, "checkout", "-q", "changeset-release/main")
	writeRepoFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	gitCommit(t, dir, "update main.go on release branch")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--since=main"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected exempt branch to pass, got error: %v\n%s", err, out.String())
	}
}

func TestStatusCmd_PassesOnDocsOnlyDiff(t *testing.T) {
	dir := initRepo(t)
	writeRepoFile(t, dir, "main.go", "package main\n")
	gitCommit(t, dir, "add main.go")
	runGitCmd(t, dir, "branch", "feature")
	runGitCmd(t, dir, "checkout", "-q", "feature")
	writeRepoFile(t, dir, "README.md", "# Docs update\n")
	gitCommit(t, dir, "update docs")

	root := NewRootCmd("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--since=main"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected docs-only diff to pass, got error: %v\n%s", err, out.String())
	}
}
