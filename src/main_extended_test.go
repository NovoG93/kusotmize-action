package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_BuildAll(t *testing.T) {
	// Setup temp dir for workspace
	tmpDir, err := os.MkdirTemp("", "workspace-buildall")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested kustomizations
	// base/kustomization.yaml
	// overlay/kustomization.yaml
	dirs := []string{"base", "overlay"}
	for _, d := range dirs {
		p := filepath.Join(tmpDir, d)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, "kustomization.yaml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Mock installer
	installer := &KustomizeInstaller{
		Cmd: &MockCommandRunner{
			LookPathFunc: func(file string) (string, error) { return "/bin/kustomize", nil },
			RunFunc:      func(name string, args ...string) ([]byte, error) { return []byte("v5.0.0"), nil },
		},
		Downloader: &MockDownloader{},
		FS:         &MockFileSystem{},
	}

	cfg := Config{
		WorkingDir:       tmpDir,
		OutputDir:        filepath.Join(tmpDir, "output"),
		KustomizeVersion: "v5.0.0",
		BuildAll:         true, // Enable BuildAll
	}

	builder := func(roots []string, conf Config, kustomizePath string) Summary {
		if len(roots) != 2 {
			t.Errorf("expected 2 roots, got %d", len(roots))
		}
		return Summary{Success: 2, Roots: 2}
	}

	if err := Run(cfg, installer, builder); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestRun_ChangedOnly(t *testing.T) {
	// Setup temp dir for workspace
	tmpDir, err := os.MkdirTemp("", "workspace-changedonly")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %s", args, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "you@example.com")
	runGit("config", "user.name", "Your Name")

	// Create base/kustomization.yaml and commit it
	baseDir := filepath.Join(tmpDir, "base")
	os.MkdirAll(baseDir, 0755)
	os.WriteFile(filepath.Join(baseDir, "kustomization.yaml"), []byte(""), 0644)
	runGit("add", ".")
	runGit("commit", "-m", "initial commit")

	// Create overlay/kustomization.yaml and commit it (this will be the "changed" one)
	overlayDir := filepath.Join(tmpDir, "overlay")
	os.MkdirAll(overlayDir, 0755)
	os.WriteFile(filepath.Join(overlayDir, "kustomization.yaml"), []byte(""), 0644)
	runGit("add", ".")
	runGit("commit", "-m", "add overlay")

	// Mock installer
	installer := &KustomizeInstaller{
		Cmd: &MockCommandRunner{
			LookPathFunc: func(file string) (string, error) { return "/bin/kustomize", nil },
			RunFunc:      func(name string, args ...string) ([]byte, error) { return []byte("v5.0.0"), nil },
		},
		Downloader: &MockDownloader{},
		FS:         &MockFileSystem{},
	}

	// Change to temp dir so "." works
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		WorkingDir:       ".",
		OutputDir:        filepath.Join(tmpDir, "output"),
		KustomizeVersion: "v5.0.0",
		ChangedOnly:      true, // Enable ChangedOnly
	}

	builder := func(roots []string, conf Config, kustomizePath string) Summary {
		// Should only pick up overlay because it was changed in the last commit
		// Wait, getChangedFilesLastCommit compares HEAD to HEAD~1.
		// So overlay/kustomization.yaml should be in the diff.
		if len(roots) != 1 {
			t.Errorf("expected 1 root, got %d: %v", len(roots), roots)
		} else {
			if !strings.Contains(roots[0], "overlay") {
				t.Errorf("expected overlay root, got %s", roots[0])
			}
		}
		return Summary{Success: 1, Roots: 1}
	}

	if err := Run(cfg, installer, builder); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestRun_FailOnError(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "workspace-fail")
	defer os.RemoveAll(tmpDir)

	installer := &KustomizeInstaller{
		Cmd: &MockCommandRunner{
			LookPathFunc: func(file string) (string, error) { return "/bin/kustomize", nil },
			RunFunc:      func(name string, args ...string) ([]byte, error) { return []byte("v5.0.0"), nil },
		},
		Downloader: &MockDownloader{},
		FS:         &MockFileSystem{},
	}

	cfg := Config{
		WorkingDir:       tmpDir,
		OutputDir:        filepath.Join(tmpDir, "output"),
		KustomizeVersion: "v5.0.0",
		FailOnError:      true,
	}

	builder := func(roots []string, conf Config, kustomizePath string) Summary {
		return Summary{Failed: 1, Roots: 1}
	}

	err := Run(cfg, installer, builder)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "kustomize build failed") {
		t.Errorf("expected 'kustomize build failed', got %v", err)
	}
}
