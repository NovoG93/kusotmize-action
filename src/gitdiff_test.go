package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetChangedFilesLastCommit_ReturnsRepoRootRelativeSlashPaths(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	// Base commit
	mustWriteFile(t, filepath.Join(repoDir, "apps/a/kustomization.yaml"), "resources: []\n")
	mustWriteFile(t, filepath.Join(repoDir, "apps/a/deploy.yaml"), "apiVersion: v1\nkind: ConfigMap\n")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "base")

	// Change commit
	mustWriteFile(t, filepath.Join(repoDir, "apps/a/deploy.yaml"), "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: changed\n")
	mustWriteFile(t, filepath.Join(repoDir, "apps/b/other.txt"), "new")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "change")

	changed, err := getChangedFilesLastCommit(repoDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changed) == 0 {
		t.Fatalf("expected changed files, got none")
	}
	for _, p := range changed {
		if strings.Contains(p, "\\") {
			t.Fatalf("expected slash path, got %q", p)
		}
		if strings.HasPrefix(p, "/") {
			t.Fatalf("expected repo-root relative path, got %q", p)
		}
		if strings.HasPrefix(p, "./") {
			t.Fatalf("expected normalized path without ./ prefix, got %q", p)
		}
	}
	if !contains(changed, "apps/a/deploy.yaml") {
		t.Fatalf("expected apps/a/deploy.yaml in changed set, got %v", changed)
	}
	if !contains(changed, "apps/b/other.txt") {
		t.Fatalf("expected apps/b/other.txt in changed set, got %v", changed)
	}
}

func TestGetChangedFilesLastCommit_WhenHeadMinus1Missing_ReturnsHelpfulError(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	mustWriteFile(t, filepath.Join(repoDir, "README.md"), "hello")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "initial")

	_, err := getChangedFilesLastCommit(repoDir)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "HEAD~1") {
		t.Fatalf("expected error to mention HEAD~1, got %q", msg)
	}
	if !strings.Contains(msg, "fetch-depth") {
		t.Fatalf("expected error to mention fetch-depth, got %q", msg)
	}
}

func TestGetChangedFilesLastCommit_IncludesDeletedFiles(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	// Base commit
	mustWriteFile(t, filepath.Join(repoDir, "apps/c/delete.txt"), "bye")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "base")

	// Delete commit
	runGit(t, repoDir, "rm", "apps/c/delete.txt")
	runGit(t, repoDir, "commit", "-m", "delete")

	changed, err := getChangedFilesLastCommit(repoDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !contains(changed, "apps/c/delete.txt") {
		t.Fatalf("expected deleted file in changed set, got %v", changed)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
