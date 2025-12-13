package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallKustomize_EmptyVersion(t *testing.T) {
	installer := NewKustomizeInstaller()
	_, err := installer.Install("", "")
	if err == nil {
		t.Error("expected error for empty version, got nil")
	}
	if err.Error() != "kustomize version is empty" {
		t.Errorf("expected 'kustomize version is empty', got '%v'", err)
	}
}

func TestInstallKustomize_AlreadyInstalled_Match(t *testing.T) {
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "/usr/local/bin/kustomize", nil
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			// Simulate version output
			return []byte("v5.0.0"), nil
		},
	}
	installer := &KustomizeInstaller{
		Cmd:        cmdRunner,
		Downloader: &MockDownloader{},
		FS:         &MockFileSystem{},
	}

	path, err := installer.Install("v5.0.0", "")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if path != "/usr/local/bin/kustomize" {
		t.Errorf("expected /usr/local/bin/kustomize, got %s", path)
	}
}

func TestInstallKustomize_AlreadyInstalled_Mismatch(t *testing.T) {
	// If installed version doesn't match, it should proceed to download
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "/usr/local/bin/kustomize", nil
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			if args[0] == "version" {
				return []byte("v4.0.0"), nil // Mismatch
			}
			// Tar extraction
			return []byte("extracted"), nil
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

	path, err := installer.Install("v5.0.0", "")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	// It should return the new path (which defaults to /usr/local/bin/kustomize in the code)
	if path != "/usr/local/bin/kustomize" {
		t.Errorf("expected /usr/local/bin/kustomize, got %s", path)
	}
}

func TestVerifySHA256(t *testing.T) {
	// Create a temp file
	tmp, err := os.CreateTemp("", "sha-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	content := []byte("hello world")
	if _, err := tmp.Write(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	// Calculate SHA256
	h := sha256.New()
	h.Write(content)
	expected := hex.EncodeToString(h.Sum(nil))

	// Test match
	if err := verifySHA256(tmp.Name(), expected); err != nil {
		t.Errorf("expected match, got error: %v", err)
	}

	// Test mismatch
	if err := verifySHA256(tmp.Name(), "badsha"); err == nil {
		t.Error("expected mismatch error, got nil")
	}

	// Test invalid length
	if err := verifySHA256(tmp.Name(), "short"); err == nil {
		t.Error("expected invalid length error, got nil")
	}
}

func TestInstallKustomize_SHA256_Verification(t *testing.T) {
	// Create a dummy file to simulate download
	tmp, err := os.CreateTemp("", "kustomize-dummy")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	content := []byte("dummy content")
	tmp.Write(content)
	tmp.Close()

	// Calculate SHA
	h := sha256.New()
	h.Write(content)
	validSHA := hex.EncodeToString(h.Sum(nil))

	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "", errors.New("not installed")
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			return []byte("extracted"), nil
		},
	}
	downloader := &MockDownloader{
		DownloadFunc: func(url, dest string) error {
			// Copy our dummy file to dest to simulate download
			input, _ := os.ReadFile(tmp.Name())
			return os.WriteFile(dest, input, 0644)
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

	// Test with valid SHA
	_, err = installer.Install("v5.0.0", validSHA)
	if err != nil {
		t.Errorf("expected success with valid SHA, got: %v", err)
	}

	// Test with invalid SHA
	_, err = installer.Install("v5.0.0", "badsha"+validSHA[6:]) // Make it same length but different
	if err == nil {
		t.Error("expected error with invalid SHA, got nil")
	}
}

func TestInstallKustomize_TarExtractionFail_Fallback(t *testing.T) {
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "", errors.New("not installed")
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			// Fail first tar attempt (to /usr/local/bin)
			if args[len(args)-1] == "/usr/local/bin" {
				return []byte("permission denied"), errors.New("tar failed")
			}
			// Succeed second attempt (to temp dir)
			return []byte("extracted"), nil
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

	path, err := installer.Install("v5.0.0", "")
	if err != nil {
		t.Fatalf("expected success with fallback, got error: %v", err)
	}

	// Path should NOT be /usr/local/bin/kustomize
	if path == "/usr/local/bin/kustomize" {
		t.Error("expected fallback path, got /usr/local/bin/kustomize")
	}
	if filepath.Base(path) != "kustomize" {
		t.Errorf("expected kustomize binary, got %s", path)
	}
}
