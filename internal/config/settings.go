package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	CheckForUpdateOnStartup bool `json:"check_for_update_on_startup"`
}

func DefaultSettings() Settings {
	return Settings{
		CheckForUpdateOnStartup: true,
	}
}

func LoadSettings() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings(), err
	}

	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), fmt.Errorf("failed to read %s: %w", path, err)
	}

	settings := DefaultSettings()
	if check, ok := root["check_for_update_on_startup"].(bool); ok {
		settings.CheckForUpdateOnStartup = check
	}

	return settings, nil
}

func SaveSettings(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}

	root := map[string]any{
		"check_for_update_on_startup": settings.CheckForUpdateOnStartup,
	}
	if err := writeJSONMap(path, root); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func settingsPath() (string, error) {
	dir, err := codexQuotaConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}
