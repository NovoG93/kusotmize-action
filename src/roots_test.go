package main

import "testing"

func TestMapRootsToRepoRootRelative(t *testing.T) {
	tests := []struct {
		name       string
		workingDir string
		roots      []string
		expected   []string
	}{
		{
			name:       "workingDir dot leaves roots unchanged except empty becomes dot",
			workingDir: ".",
			roots:      []string{"", "apps/a", "apps/b"},
			expected:   []string{".", "apps/a", "apps/b"},
		},
		{
			name:       "workingDir prefixes all roots",
			workingDir: "deploy",
			roots:      []string{"", "apps/a", "apps/b"},
			expected:   []string{"deploy", "deploy/apps/a", "deploy/apps/b"},
		},
		{
			name:       "workingDir is normalized (./ and trailing slash)",
			workingDir: "./deploy/",
			roots:      []string{"", "apps/a"},
			expected:   []string{"deploy", "deploy/apps/a"},
		},
		{
			name:       "root dot is treated as workingDir",
			workingDir: "infra",
			roots:      []string{"."},
			expected:   []string{"infra"},
		},
		{
			name:       "empty workingDir behaves like dot",
			workingDir: "",
			roots:      []string{"", "apps/a"},
			expected:   []string{".", "apps/a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapRootsToRepoRootRelative(tt.workingDir, tt.roots)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d (%v)", len(tt.expected), len(got), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Fatalf("index %d: expected %q, got %q (all=%v)", i, tt.expected[i], got[i], got)
				}
			}
		})
	}
}
