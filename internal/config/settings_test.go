package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettingsMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if !settings.CheckForUpdateOnStartup {
		t.Fatalf("CheckForUpdateOnStartup = false, want true")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	initial := Settings{CheckForUpdateOnStartup: false}
	if err := SaveSettings(initial); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if loaded.CheckForUpdateOnStartup != initial.CheckForUpdateOnStartup {
		t.Fatalf("CheckForUpdateOnStartup = %v, want %v", loaded.CheckForUpdateOnStartup, initial.CheckForUpdateOnStartup)
	}
}

func TestLoadSettingsInvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	path := filepath.Join(dir, "codex-quota", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	if _, err := LoadSettings(); err == nil {
		t.Fatalf("LoadSettings() error = nil, want non-nil")
	}
}
