package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	// Override config path by using a temp HOME
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{VaultPath: "/tmp/test-vault"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	p := filepath.Join(tmpDir, ".config", "flow.yaml")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.VaultPath != "/tmp/test-vault" {
		t.Errorf("VaultPath = %q, want %q", loaded.VaultPath, "/tmp/test-vault")
	}
}

func TestLoadMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	_, err := Load()
	if err != ErrNotConfigured {
		t.Errorf("Load() = %v, want ErrNotConfigured", err)
	}
}
