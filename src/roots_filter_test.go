package main

import (
	"encoding/json"
	"testing"
)

func TestSelectRootsForChangedFiles_PicksDeepestMatchingRootAndPreservesOrder(t *testing.T) {
	roots := []string{"apps", "apps/foo", "apps/foo/overlays/dev", "cluster"}
	changed := []string{
		"apps/foo/overlays/dev/kustomization.yaml",
		"apps/foo/base/deploy.yaml",
		"cluster/ns.yaml",
		"README.md", // no match
	}

	got := selectRootsForChangedFiles(roots, changed)
	expected := []string{"apps/foo", "apps/foo/overlays/dev", "cluster"}

	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("index %d: expected %q got %q (all=%v)", i, expected[i], got[i], got)
		}
	}
}

func TestSelectRootsForChangedFiles_NoMatchesReturnsEmpty(t *testing.T) {
	roots := []string{"apps/foo"}
	changed := []string{"docs/readme.md"}
	got := selectRootsForChangedFiles(roots, changed)
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "[]" {
		t.Fatalf("expected JSON [] (not null), got %s (slice=%v)", string(b), got)
	}
}

func TestSelectRootsForChangedFiles_EmptyChangedFilesReturnsEmptySliceNotNil(t *testing.T) {
	roots := []string{"apps/foo"}
	var changed []string
	got := selectRootsForChangedFiles(roots, changed)
	if got == nil {
		t.Fatalf("expected non-nil empty slice")
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "[]" {
		t.Fatalf("expected JSON [], got %s (slice=%v)", string(b), got)
	}
}

func TestSelectRootsForChangedFiles_DotRootMatchesEverything(t *testing.T) {
	roots := []string{".", "apps/foo"}
	changed := []string{"docs/readme.md"}
	got := selectRootsForChangedFiles(roots, changed)
	expected := []string{"."}
	if len(got) != len(expected) || got[0] != expected[0] {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
