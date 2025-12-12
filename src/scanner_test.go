package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestFindKustomizationFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Define the file structure to create
	files := []string{
		"kustomization.yaml",
		"app1/kustomization.yaml",
		"app2/kustomization.yml",
		"app3/subdir/kustomization.yaml",
		"other/readme.md",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", f, err)
		}
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	// Expected result (sorted)
	expected := []string{
		filepath.Join(tmpDir, "app1/kustomization.yaml"),
		filepath.Join(tmpDir, "app2/kustomization.yml"),
		filepath.Join(tmpDir, "app3/subdir/kustomization.yaml"),
		filepath.Join(tmpDir, "kustomization.yaml"),
	}
	sort.Strings(expected)

	// Run the function
	found, err := findKustomizationFiles(tmpDir)
	if err != nil {
		t.Fatalf("findKustomizationFiles returned error: %v", err)
	}
	sort.Strings(found)

	// Verify results
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("Expected %v, got %v", expected, found)
	}
}

func TestFindKustomizationFiles_SkipsDotGit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := []string{
		".git/kustomization.yaml",
		"app/kustomization.yaml",
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", f, err)
		}
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	found, err := findKustomizationFiles(tmpDir)
	if err != nil {
		t.Fatalf("findKustomizationFiles returned error: %v", err)
	}
	for _, p := range found {
		if filepath.Base(filepath.Dir(p)) == ".git" {
			t.Fatalf("expected .git to be skipped, but found %q", p)
		}
	}

	expected := []string{filepath.Join(tmpDir, "app/kustomization.yaml")}
	if !reflect.DeepEqual(found, expected) {
		t.Fatalf("Expected %v, got %v", expected, found)
	}
}

func TestFindKustomizationFilesWithExclusions_SkipsOutputDirAndDotGit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputDir := "kustomize-builds"
	files := []string{
		".git/kustomization.yaml",
		outputDir + "/kustomization.yaml",
		"apps/a/kustomization.yaml",
		"apps/b/kustomization.yml",
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", f, err)
		}
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	found, err := findKustomizationFilesWithExclusions(tmpDir, []string{".git", outputDir})
	if err != nil {
		t.Fatalf("findKustomizationFilesWithExclusions returned error: %v", err)
	}

	expected := []string{
		filepath.Join(tmpDir, "apps/a/kustomization.yaml"),
		filepath.Join(tmpDir, "apps/b/kustomization.yml"),
	}
	sort.Strings(expected)
	if !reflect.DeepEqual(found, expected) {
		t.Fatalf("Expected %v, got %v", expected, found)
	}
}

func TestFindKustomizationFilesWithExclusions_OnlySkipsOutputDirAtRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputDir := "kustomize-builds"
	files := []string{
		outputDir + "/kustomization.yaml",           // should be skipped
		"apps/" + outputDir + "/kustomization.yaml", // must NOT be skipped
		"apps/other/kustomization.yaml",             // control
		".git/kustomization.yaml",                   // must be skipped
		"apps/.git/kustomization.yaml",              // must be skipped (basename-based)
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", f, err)
		}
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	found, err := findKustomizationFilesWithExclusions(tmpDir, []string{".git", outputDir})
	if err != nil {
		t.Fatalf("findKustomizationFilesWithExclusions returned error: %v", err)
	}

	expected := []string{
		filepath.Join(tmpDir, "apps/"+outputDir+"/kustomization.yaml"),
		filepath.Join(tmpDir, "apps/other/kustomization.yaml"),
	}
	sort.Strings(expected)
	if !reflect.DeepEqual(found, expected) {
		t.Fatalf("Expected %v, got %v", expected, found)
	}
}

func TestDedupeTopLevelDirs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"app1", "app2", "app3"},
			expected: []string{"app1", "app2", "app3"},
		},
		{
			name:     "Nested directories",
			input:    []string{"app1", "app1/subdir", "app2"},
			expected: []string{"app1", "app2"},
		},
		{
			name:     "Deeply nested",
			input:    []string{"a", "a/b", "a/b/c", "d"},
			expected: []string{"a", "d"},
		},
		{
			name:     "Unrelated prefixes",
			input:    []string{"app", "apple"},
			expected: []string{"app", "apple"},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedupeTopLevelDirs(tt.input)
			sort.Strings(result)
			sort.Strings(tt.expected)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
