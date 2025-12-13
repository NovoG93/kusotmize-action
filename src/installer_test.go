package main

import (
	"errors"
	"os"
	"testing"
)

// MockCommandRunner
type MockCommandRunner struct {
	LookPathFunc func(file string) (string, error)
	RunFunc      func(name string, args ...string) ([]byte, error)
}

func (m *MockCommandRunner) LookPath(file string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	return "", errors.New("not found")
}

func (m *MockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return nil, nil
}

// MockDownloader
type MockDownloader struct {
	DownloadFunc func(url string, dest string) error
}

func (m *MockDownloader) Download(url string, dest string) error {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(url, dest)
	}
	return nil
}

// MockFileSystem
type MockFileSystem struct {
	ChmodFunc func(name string, mode os.FileMode) error
}

func (m *MockFileSystem) Chmod(name string, mode os.FileMode) error {
	if m.ChmodFunc != nil {
		return m.ChmodFunc(name, mode)
	}
	return nil
}

func TestInstallKustomize_Success(t *testing.T) {
	// Setup mocks
	cmdRunner := &MockCommandRunner{
		LookPathFunc: func(file string) (string, error) {
			return "", errors.New("not installed")
		},
		RunFunc: func(name string, args ...string) ([]byte, error) {
			// Mock tar extraction success
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

	// We pass empty SHA256 to skip verification logic which reads from disk
	// If we wanted to test SHA verification, we'd need to mock file opening or write a real file.
	// For now, let's assume empty SHA skips verification or we can write a dummy file if needed.
	// Looking at verifySHA256: if expected is empty, it returns nil.
	path, err := installer.Install("v5.0.0", "")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if path == "" {
		t.Error("expected path, got empty string")
	}
}

func TestInstallKustomize_DownloadFail(t *testing.T) {
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

	_, err := installer.Install("v5.0.0", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "download failed" {
		t.Errorf("expected 'download failed', got '%v'", err)
	}
}

func TestInstallKustomize_ChmodFail(t *testing.T) {
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
			return nil
		},
	}
	fs := &MockFileSystem{
		ChmodFunc: func(name string, mode os.FileMode) error {
			return errors.New("chmod failed")
		},
	}

	installer := &KustomizeInstaller{
		Cmd:        cmdRunner,
		Downloader: downloader,
		FS:         fs,
	}

	_, err := installer.Install("v5.0.0", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "chmod failed" {
		t.Errorf("expected 'chmod failed', got '%v'", err)
	}
}
