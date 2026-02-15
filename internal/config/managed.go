package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type managedStore struct {
	Accounts []managedAccount `json:"accounts"`
}

type managedAccount struct {
	Label        string `json:"label,omitempty"`
	Email        string `json:"email,omitempty"`
	AccountID    string `json:"account_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at_ms,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
}

func LoadManagedAccounts() ([]*Account, error) {
	path, err := managedAccountsPath()
	if err != nil {
		return nil, err
	}

	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Account{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	store := managedStore{}
	if rawAccounts, ok := root["accounts"]; ok {
		store.Accounts, err = decodeManagedAccounts(rawAccounts)
		if err != nil {
			return nil, fmt.Errorf("failed to decode accounts in %s: %w", path, err)
		}
	}

	accounts := make([]*Account, 0, len(store.Accounts))
	for _, item := range store.Accounts {
		if strings.TrimSpace(item.AccessToken) == "" {
			continue
		}
		account := &Account{
			Label:        strings.TrimSpace(item.Label),
			Email:        strings.TrimSpace(item.Email),
			AccountID:    strings.TrimSpace(item.AccountID),
			AccessToken:  strings.TrimSpace(item.AccessToken),
			RefreshToken: strings.TrimSpace(item.RefreshToken),
			ClientID:     strings.TrimSpace(item.ClientID),
			Source:       SourceManaged,
			FilePath:     path,
			Writable:     true,
		}
		if item.ExpiresAt > 0 {
			account.ExpiresAt = time.UnixMilli(item.ExpiresAt)
		}

		claims := ParseAccessToken(account.AccessToken)
		if account.AccountID == "" {
			account.AccountID = claims.AccountID
		}
		if account.ClientID == "" {
			account.ClientID = claims.ClientID
		}
		if account.Email == "" {
			account.Email = claims.Email
		}
		if account.ExpiresAt.IsZero() {
			account.ExpiresAt = claims.ExpiresAt
		}

		accounts = append(accounts, account)
	}

	sort.Slice(accounts, func(i, j int) bool {
		return strings.ToLower(accounts[i].Label) < strings.ToLower(accounts[j].Label)
	})

	return accounts, nil
}

func UpsertManagedAccount(account *Account) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}
	if strings.TrimSpace(account.AccessToken) == "" {
		return fmt.Errorf("access token is empty")
	}
	if strings.TrimSpace(account.AccountID) == "" {
		claims := ParseAccessToken(account.AccessToken)
		account.AccountID = claims.AccountID
	}
	if strings.TrimSpace(account.AccountID) == "" {
		return fmt.Errorf("account_id is missing")
	}

	path, err := managedAccountsPath()
	if err != nil {
		return err
	}

	store := managedStore{}
	root, err := readJSONMap(path)
	if err == nil {
		if rawAccounts, ok := root["accounts"]; ok {
			store.Accounts, err = decodeManagedAccounts(rawAccounts)
			if err != nil {
				return fmt.Errorf("failed to decode accounts in %s: %w", path, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	item := managedAccount{
		Label:        strings.TrimSpace(account.Label),
		Email:        strings.TrimSpace(account.Email),
		AccountID:    strings.TrimSpace(account.AccountID),
		AccessToken:  strings.TrimSpace(account.AccessToken),
		RefreshToken: strings.TrimSpace(account.RefreshToken),
		ClientID:     strings.TrimSpace(account.ClientID),
	}
	if !account.ExpiresAt.IsZero() {
		item.ExpiresAt = account.ExpiresAt.UnixMilli()
	}

	updated := false
	for i := range store.Accounts {
		if strings.TrimSpace(store.Accounts[i].AccountID) == item.AccountID {
			store.Accounts[i] = mergeManagedAccount(store.Accounts[i], item)
			updated = true
			break
		}
	}
	if !updated {
		store.Accounts = append(store.Accounts, item)
	}

	if err := writeJSONMap(path, map[string]any{"accounts": store.Accounts}); err != nil {
		return err
	}

	return nil
}

func mergeManagedAccount(existing, incoming managedAccount) managedAccount {
	merged := existing

	if strings.TrimSpace(merged.Label) == "" {
		merged.Label = incoming.Label
	}
	if strings.TrimSpace(merged.Email) == "" {
		merged.Email = incoming.Email
	}
	if strings.TrimSpace(merged.ClientID) == "" {
		merged.ClientID = incoming.ClientID
	}
	if strings.TrimSpace(merged.RefreshToken) == "" {
		merged.RefreshToken = incoming.RefreshToken
	}

	if merged.ExpiresAt == 0 {
		merged.ExpiresAt = incoming.ExpiresAt
	}

	if incoming.ExpiresAt > 0 && (merged.ExpiresAt == 0 || incoming.ExpiresAt > merged.ExpiresAt) {
		merged.AccessToken = incoming.AccessToken
		merged.ExpiresAt = incoming.ExpiresAt
		if strings.TrimSpace(incoming.RefreshToken) != "" {
			merged.RefreshToken = incoming.RefreshToken
		}
		if strings.TrimSpace(incoming.ClientID) != "" {
			merged.ClientID = incoming.ClientID
		}
	}

	if strings.TrimSpace(merged.AccessToken) == "" {
		merged.AccessToken = incoming.AccessToken
		if merged.ExpiresAt == 0 {
			merged.ExpiresAt = incoming.ExpiresAt
		}
	}

	return merged
}

func saveManagedAccount(account *Account) error {
	return UpsertManagedAccount(account)
}

func DeleteManagedAccount(accountID string) error {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return fmt.Errorf("account_id is empty")
	}

	path, err := managedAccountsPath()
	if err != nil {
		return err
	}

	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	store := managedStore{}
	if rawAccounts, ok := root["accounts"]; ok {
		store.Accounts, err = decodeManagedAccounts(rawAccounts)
		if err != nil {
			return fmt.Errorf("failed to decode accounts in %s: %w", path, err)
		}
	}

	filtered := make([]managedAccount, 0, len(store.Accounts))
	for _, item := range store.Accounts {
		if strings.TrimSpace(item.AccountID) == accountID {
			continue
		}
		filtered = append(filtered, item)
	}

	if len(filtered) == len(store.Accounts) {
		return nil
	}

	root["accounts"] = filtered
	return writeJSONMap(path, root)
}

func ApplyAccountToOpenCode(account *Account) (string, error) {
	if account == nil {
		return "", fmt.Errorf("account is nil")
	}
	path := opencodeAuthPath()
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("OpenCode auth path is unknown")
	}

	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			root = make(map[string]any)
		} else {
			return "", fmt.Errorf("failed to read %s: %w", path, err)
		}
	}

	openai := asMap(root["openai"])
	if openai == nil {
		openai = make(map[string]any)
		root["openai"] = openai
	}

	openai["access"] = account.AccessToken
	if account.RefreshToken != "" {
		openai["refresh"] = account.RefreshToken
	}
	if account.AccountID != "" {
		openai["accountId"] = account.AccountID
	}
	if account.Email != "" {
		openai["email"] = account.Email
	}
	if !account.ExpiresAt.IsZero() {
		openai["expires"] = account.ExpiresAt.UnixMilli()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("failed to ensure directory for %s: %w", path, err)
	}

	if err := writeJSONMap(path, root); err != nil {
		return "", err
	}

	return path, nil
}

func managedAccountsPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("failed to locate user config dir")
		}
		base = filepath.Join(home, ".config")
	}

	dir := filepath.Join(base, "codex-quota")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}
	return filepath.Join(dir, "accounts.json"), nil
}

func decodeManagedAccounts(raw any) ([]managedAccount, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	accounts := make([]managedAccount, 0)
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}
