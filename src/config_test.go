package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Clear env vars
	os.Unsetenv("INPUT_OUTPUT-DIR")
	os.Unsetenv("INPUT_KUSTOMIZE-VERSION")
	os.Unsetenv("INPUT_ENABLE-HELM")
	os.Unsetenv("INPUT_LOAD-RESTRICTOR")
	os.Unsetenv("INPUT_WORKING-DIRECTORY")
	os.Unsetenv("INPUT_BUILD-ALL")
	os.Unsetenv("INPUT_CHANGED-ONLY")
	os.Unsetenv("INPUT_FAIL-ON-ERROR")
	os.Unsetenv("INPUT_FAIL-FAST")

	// Test defaults
	config := LoadConfig()
	if config.OutputDir != "kustomize-builds" {
		t.Errorf("Expected default OutputDir 'kustomize-builds', got '%s'", config.OutputDir)
	}
	if config.EnableHelm != true {
		t.Errorf("Expected default EnableHelm true, got %v", config.EnableHelm)
	}
	if config.ChangedOnly != true {
		t.Errorf("Expected default ChangedOnly true, got %v", config.ChangedOnly)
	}

	// Test with INPUT_ env vars (hyphenated)
	os.Setenv("INPUT_OUTPUT-DIR", "custom-out")
	os.Setenv("INPUT_ENABLE-HELM", "false")
	os.Setenv("INPUT_CHANGED-ONLY", "true")

	config = LoadConfig()
	if config.OutputDir != "custom-out" {
		t.Errorf("Expected OutputDir 'custom-out', got '%s'", config.OutputDir)
	}
	if config.EnableHelm != false {
		t.Errorf("Expected EnableHelm false, got %v", config.EnableHelm)
	}
	if config.ChangedOnly != true {
		t.Errorf("Expected ChangedOnly true, got %v", config.ChangedOnly)
	}

	// Test with INPUT_ env vars (underscored) - GitHub Actions might do this?
	os.Unsetenv("INPUT_OUTPUT-DIR")
	os.Setenv("INPUT_OUTPUT_DIR", "custom-out-underscore")

	config = LoadConfig()
	if config.OutputDir != "custom-out-underscore" {
		t.Errorf("Expected OutputDir 'custom-out-underscore', got '%s'", config.OutputDir)
	}

	// Test legacy env vars (local testing)
	os.Unsetenv("INPUT_OUTPUT_DIR")
	os.Setenv("OUTPUT_DIR", "legacy-out")

	config = LoadConfig()
	if config.OutputDir != "legacy-out" {
		t.Errorf("Expected OutputDir 'legacy-out', got '%s'", config.OutputDir)
	}
}
