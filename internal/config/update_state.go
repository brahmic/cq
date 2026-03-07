package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UpdateState struct {
	LatestVersion    string    `json:"latest_version"`
	LastCheckedAt    time.Time `json:"last_checked_at"`
	DismissedVersion string    `json:"dismissed_version"`
}

func LoadUpdateState() (UpdateState, error) {
	path, err := updateStatePath()
	if err != nil {
		return UpdateState{}, err
	}

	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return UpdateState{}, nil
		}
		return UpdateState{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	state := UpdateState{
		LatestVersion:    strings.TrimSpace(asString(root["latest_version"])),
		DismissedVersion: strings.TrimSpace(asString(root["dismissed_version"])),
	}
	if lastChecked := strings.TrimSpace(asString(root["last_checked_at"])); lastChecked != "" {
		parsed, err := time.Parse(time.RFC3339, lastChecked)
		if err != nil {
			return UpdateState{}, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		state.LastCheckedAt = parsed
	}

	return state, nil
}

func SaveUpdateState(state UpdateState) error {
	path, err := updateStatePath()
	if err != nil {
		return err
	}

	root := map[string]any{
		"latest_version":    strings.TrimSpace(state.LatestVersion),
		"dismissed_version": strings.TrimSpace(state.DismissedVersion),
		"last_checked_at":   "",
	}
	if !state.LastCheckedAt.IsZero() {
		root["last_checked_at"] = state.LastCheckedAt.UTC().Format(time.RFC3339)
	}

	if err := writeJSONMap(path, root); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func SetDismissedUpdateVersion(version string) error {
	state, err := LoadUpdateState()
	if err != nil {
		return err
	}
	state.DismissedVersion = strings.TrimSpace(version)
	return SaveUpdateState(state)
}

func updateStatePath() (string, error) {
	dir, err := codexQuotaConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update_state.json"), nil
}
