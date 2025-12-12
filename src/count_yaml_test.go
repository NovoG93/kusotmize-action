package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountYAMLFiles_CountsYamlAndYmlAndIgnoresNonYaml(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	files := []string{
		"a.yaml",
		"b.yml",
		"c.txt",
		"d.json",
	}
	for _, name := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	got, err := countYAMLFiles(dir)
	if err != nil {
		t.Fatalf("countYAMLFiles error: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected 2 YAML files, got %d", got)
	}
}

func TestCountYAMLFiles_ExcludesErrorFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	include := []string{
		"ok.yaml",
		"ok.yml",
	}
	exclude := []string{
		"root_kustomization-err.yaml",
		"root_kustomization-err.yml",
	}

	for _, name := range append(include, exclude...) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	got, err := countYAMLFiles(dir)
	if err != nil {
		t.Fatalf("countYAMLFiles error: %v", err)
	}
	if got != len(include) {
		t.Fatalf("expected %d YAML files (excluding error outputs), got %d", len(include), got)
	}
}

func TestCountYAMLFiles_CountsFilesEndingWithErrSuffix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	paths := []string{
		"service-err.yaml",
		"service-err.yml",
		"normal.yaml",
		"normal.yml",
		"build_kustomization-err.yaml",
	}
	for _, name := range paths {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	got, err := countYAMLFiles(dir)
	if err != nil {
		t.Fatalf("countYAMLFiles error: %v", err)
	}
	if got != 4 {
		t.Fatalf("expected 4 YAML files (excluding only _kustomization-err artifacts), got %d", got)
	}
}

func TestCountYAMLFiles_WorksWithNestedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	paths := []string{
		filepath.Join("nested", "a.yaml"),
		filepath.Join("nested", "deeper", "b.yml"),
		filepath.Join("nested", "deeper", "c.txt"),
		filepath.Join("nested", "deeper", "fail_kustomization-err.yaml"),
	}

	for _, rel := range paths {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	got, err := countYAMLFiles(dir)
	if err != nil {
		t.Fatalf("countYAMLFiles error: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected 2 YAML files in nested dirs, got %d", got)
	}
}
