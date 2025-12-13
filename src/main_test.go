package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetOutput(t *testing.T) {
	// Create a temp file for GITHUB_OUTPUT
	tmpFile, err := os.CreateTemp("", "github_output")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Set env var
	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	setOutput("test-name", "test-value")

	// Read file content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "test-name<<GH_OUTPUT_") {
		t.Errorf("expected output to contain test-name, got %s", s)
	}
	if !strings.Contains(s, "test-value") {
		t.Errorf("expected output to contain test-value, got %s", s)
	}
}

func TestRun_InstallFail(t *testing.T) {
	// Mock installer to fail
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "", errors.New("not installed")
		},
	}
	downloader := &MockDownloader{
		DownloadFunc: func(url, dest string) error {
			return errors.New("download failed")
		},
	}
	fs := &MockFileSystem{}

	installer := &KustomizeInstaller{
		Cmd:        cmdRunner,
		Downloader: downloader,
		FS:         fs,
	}

	cfg := Config{
		KustomizeVersion: "v5.0.0",
	}

	// Mock builder (should not be called)
	builder := func(roots []string, conf Config, kustomizePath string) Summary {
		t.Error("builder should not be called")
		return Summary{}
	}

	err := Run(cfg, installer, builder)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to install kustomize") {
		t.Errorf("expected error to contain 'failed to install kustomize', got %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	// Setup temp dir for workspace
	tmpDir, err := os.MkdirTemp("", "workspace")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy kustomization.yaml
	kustDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(kustDir, 0755); err != nil {
		t.Fatalf("failed to create base dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kustDir, "kustomization.yaml"), []byte("resources:\n- deployment.yaml"), 0644); err != nil {
		t.Fatalf("failed to write kustomization.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kustDir, "deployment.yaml"), []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: test"), 0644); err != nil {
		t.Fatalf("failed to write deployment.yaml: %v", err)
	}

	// Mock installer to succeed
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "/usr/local/bin/kustomize", nil
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("v5.0.0"), nil
		},
	}
	downloader := &MockDownloader{
		DownloadFunc: func(url, dest string) error {
			return nil
		},
	}
	fs := &MockFileSystem{
		ChmodFunc: func(name string, mode os.FileMode) error {
			return nil
		},
	}

	installer := &KustomizeInstaller{
		Cmd:        cmdRunner,
		Downloader: downloader,
		FS:         fs,
	}

	cfg := Config{
		WorkingDir:       tmpDir,
		OutputDir:        filepath.Join(tmpDir, "output"),
		KustomizeVersion: "v5.0.0",
		BuildAll:         false,
		FailOnError:      true,
	}

	// Mock builder
	builder := func(roots []string, conf Config, kustomizePath string) Summary {
		// Verify roots
		if len(roots) != 1 {
			t.Errorf("expected 1 root, got %d", len(roots))
		}
		// Simulate success
		return Summary{
			Success: 1,
			Roots:   1,
		}
	}

	err = Run(cfg, installer, builder)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify output dir exists
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		t.Errorf("expected output dir to exist")
	}

	// Verify summary file exists
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "_summary.json")); os.IsNotExist(err) {
		t.Errorf("expected summary file to exist")
	}
}
