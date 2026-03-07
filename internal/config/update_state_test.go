package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUpdateStateMissingFileReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	state, err := LoadUpdateState()
	if err != nil {
		t.Fatalf("LoadUpdateState() error = %v", err)
	}
	if state.LatestVersion != "" || !state.LastCheckedAt.IsZero() || state.DismissedVersion != "" {
		t.Fatalf("LoadUpdateState() = %+v, want zero state", state)
	}
}

func TestSaveAndLoadUpdateState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	now := time.Date(2026, time.March, 7, 14, 0, 0, 0, time.UTC)
	initial := UpdateState{
		LatestVersion:    "0.1.5",
		LastCheckedAt:    now,
		DismissedVersion: "0.1.4",
	}
	if err := SaveUpdateState(initial); err != nil {
		t.Fatalf("SaveUpdateState() error = %v", err)
	}

	loaded, err := LoadUpdateState()
	if err != nil {
		t.Fatalf("LoadUpdateState() error = %v", err)
	}
	if loaded != initial {
		t.Fatalf("LoadUpdateState() = %+v, want %+v", loaded, initial)
	}
}

func TestLoadUpdateStateInvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	path := filepath.Join(dir, "codex-quota", "update_state.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	if _, err := LoadUpdateState(); err == nil {
		t.Fatalf("LoadUpdateState() error = nil, want non-nil")
	}
}

func TestSetDismissedUpdateVersion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CQ_CONFIG_HOME", dir)

	if err := SaveUpdateState(UpdateState{LatestVersion: "0.1.5"}); err != nil {
		t.Fatalf("SaveUpdateState() error = %v", err)
	}
	if err := SetDismissedUpdateVersion("0.1.5"); err != nil {
		t.Fatalf("SetDismissedUpdateVersion() error = %v", err)
	}

	state, err := LoadUpdateState()
	if err != nil {
		t.Fatalf("LoadUpdateState() error = %v", err)
	}
	if state.DismissedVersion != "0.1.5" {
		t.Fatalf("DismissedVersion = %q, want %q", state.DismissedVersion, "0.1.5")
	}
}
