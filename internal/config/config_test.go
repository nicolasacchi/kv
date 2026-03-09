package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAPIKey_FlagOverride(t *testing.T) {
	key, err := LoadAPIKey("flag-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "flag-key" {
		t.Errorf("expected flag-key, got %s", key)
	}
}

func TestLoadAPIKey_EnvVar(t *testing.T) {
	t.Setenv("KLAVIYO_API_KEY", "env-key")
	key, err := LoadAPIKey("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "env-key" {
		t.Errorf("expected env-key, got %s", key)
	}
}

func TestLoadAPIKey_KVEnvVar(t *testing.T) {
	t.Setenv("KV_API_KEY", "kv-env-key")
	key, err := LoadAPIKey("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "kv-env-key" {
		t.Errorf("expected kv-env-key, got %s", key)
	}
}

func TestLoadAPIKey_FlagOverridesEnv(t *testing.T) {
	t.Setenv("KLAVIYO_API_KEY", "env-key")
	key, err := LoadAPIKey("flag-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "flag-key" {
		t.Errorf("flag should override env, got %s", key)
	}
}

func TestLoadAPIKey_MissingKey(t *testing.T) {
	// Ensure no env vars are set
	t.Setenv("KLAVIYO_API_KEY", "")
	t.Setenv("KV_API_KEY", "")
	// Point config to nonexistent dir
	t.Setenv("HOME", t.TempDir())

	_, err := LoadAPIKey("", "")
	if err == nil {
		t.Fatal("expected error when no key available")
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "***"},
		{"1234567890", "***"},
		{"pk_abc123456789xyz", "pk_abc12***9xyz"},
	}

	for _, tt := range tests {
		result := MaskKey(tt.input)
		if result != tt.expected {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAddAndRemoveProject(t *testing.T) {
	// Use temp dir as home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Add a project
	err := AddProject("test", "pk_test123", "2024-10-15")
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	// Verify file was created
	cfgPath := filepath.Join(tmpDir, ".config", "kv", "config.toml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}

	// List projects
	cfg, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if cfg.DefaultProject != "test" {
		t.Errorf("expected default project 'test', got %q", cfg.DefaultProject)
	}
	if cfg.Projects["test"].APIKey != "pk_test123" {
		t.Errorf("expected API key 'pk_test123', got %q", cfg.Projects["test"].APIKey)
	}

	// Load API key
	key, err := LoadAPIKey("", "test")
	if err != nil {
		t.Fatalf("LoadAPIKey: %v", err)
	}
	if key != "pk_test123" {
		t.Errorf("expected pk_test123, got %s", key)
	}

	// Remove project
	err = RemoveProject("test")
	if err != nil {
		t.Fatalf("RemoveProject: %v", err)
	}

	cfg, err = ListProjects()
	if err != nil {
		t.Fatalf("ListProjects after remove: %v", err)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected 0 projects after remove, got %d", len(cfg.Projects))
	}
}

func TestSetDefaultProject(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	AddProject("first", "key1", "")
	AddProject("second", "key2", "")

	err := SetDefaultProject("second")
	if err != nil {
		t.Fatalf("SetDefaultProject: %v", err)
	}

	cfg, _ := ListProjects()
	if cfg.DefaultProject != "second" {
		t.Errorf("expected default 'second', got %q", cfg.DefaultProject)
	}
}
