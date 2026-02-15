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

func LoadAllAccountsWithSources() (AccountsLoadResult, error) {
	accounts := make([]*Account, 0, 6)

	appAccounts, err := LoadManagedAccounts()
	if err != nil {
		return AccountsLoadResult{}, err
	}
	accounts = append(accounts, appAccounts...)

	opencodePaths := opencodeAuthPaths()
	writable := firstExistingPath(opencodePaths)
	if writable == "" && len(opencodePaths) > 0 {
		writable = opencodePaths[0]
	}

	for _, path := range opencodePaths {
		openCodeMain, err := loadOpenCodeAccountFile(path, SourceOpenCode, path == writable)
		if err != nil {
			return AccountsLoadResult{}, err
		}
		if openCodeMain != nil {
			accounts = append(accounts, openCodeMain)
		}
	}

	codexAccount, err := loadCodexAccountFile(codexAuthPath())
	if err != nil {
		return AccountsLoadResult{}, err
	}
	if codexAccount != nil {
		accounts = append(accounts, codexAccount)
	}

	sourcesByAccountID := make(map[string][]string)
	for _, account := range accounts {
		if account == nil || account.AccountID == "" {
			continue
		}
		sourcesByAccountID[account.AccountID] = appendUniqueString(sourcesByAccountID[account.AccountID], account.SourceLabel())
	}

	accounts = dedupeAccounts(accounts)
	for _, account := range accounts {
		finalizeAccount(account)
	}

	sort.Slice(accounts, func(i, j int) bool {
		return strings.ToLower(accounts[i].Label) < strings.ToLower(accounts[j].Label)
	})

	return AccountsLoadResult{Accounts: accounts, SourcesByAccountID: sourcesByAccountID}, nil
}

func loadOpenCodeAccountFile(path string, source Source, writable bool) (*Account, error) {
	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	openai := asMap(root["openai"])
	if openai == nil {
		return nil, nil
	}

	account := buildOpenAIAccount(openai, source, path, writable)
	if account == nil {
		return nil, nil
	}

	return account, nil
}

func buildOpenAIAccount(openai map[string]any, source Source, path string, writable bool) *Account {
	accessToken := strings.TrimSpace(asString(openai["access"]))
	if accessToken == "" {
		return nil
	}

	account := &Account{
		AccessToken:  accessToken,
		RefreshToken: strings.TrimSpace(asString(openai["refresh"])),
		AccountID:    strings.TrimSpace(asString(openai["accountId"])),
		Email:        strings.TrimSpace(asString(openai["email"])),
		Source:       source,
		FilePath:     path,
		Writable:     writable,
	}

	if expiresMillis, ok := asInt64(openai["expires"]); ok && expiresMillis > 0 {
		account.ExpiresAt = time.UnixMilli(expiresMillis)
	}

	claims := ParseAccessToken(accessToken)
	if account.AccountID == "" {
		account.AccountID = claims.AccountID
	}
	if account.ClientID == "" {
		account.ClientID = claims.ClientID
	}
	if account.ExpiresAt.IsZero() {
		account.ExpiresAt = claims.ExpiresAt
	}
	if account.Email == "" {
		account.Email = claims.Email
	}

	return account
}

func loadCodexAccountFile(path string) (*Account, error) {
	root, err := readJSONMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	tokens := asMap(root["tokens"])
	if tokens == nil {
		return nil, nil
	}

	accessToken := strings.TrimSpace(asString(tokens["access_token"]))
	if accessToken == "" {
		return nil, nil
	}

	account := &Account{
		AccessToken:  accessToken,
		RefreshToken: strings.TrimSpace(asString(tokens["refresh_token"])),
		AccountID:    strings.TrimSpace(asString(tokens["account_id"])),
		Source:       SourceCodex,
		FilePath:     path,
		Writable:     true,
	}

	claims := ParseAccessToken(accessToken)
	if account.AccountID == "" {
		account.AccountID = claims.AccountID
	}
	account.ClientID = claims.ClientID
	account.ExpiresAt = claims.ExpiresAt

	return account, nil
}

func saveOpenCodeAccount(account *Account) error {
	root, err := readJSONMap(account.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", account.FilePath, err)
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
	if !account.ExpiresAt.IsZero() {
		openai["expires"] = account.ExpiresAt.UnixMilli()
	}

	return writeJSONMap(account.FilePath, root)
}

func saveCodexAccount(account *Account) error {
	root, err := readJSONMap(account.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", account.FilePath, err)
	}

	tokens := asMap(root["tokens"])
	if tokens == nil {
		tokens = make(map[string]any)
		root["tokens"] = tokens
	}

	tokens["access_token"] = account.AccessToken
	if account.RefreshToken != "" {
		tokens["refresh_token"] = account.RefreshToken
	}
	if account.AccountID != "" {
		tokens["account_id"] = account.AccountID
	}

	root["last_refresh"] = time.Now().UTC().Format(time.RFC3339)

	return writeJSONMap(account.FilePath, root)
}

func finalizeAccount(account *Account) {
	if account == nil {
		return
	}

	if account.Label == "" {
		if account.Email != "" {
			account.Label = account.Email
		} else if account.AccountID != "" {
			account.Label = shortAccountID(account.AccountID)
		} else {
			account.Label = account.SourceLabel()
		}
	}

	if account.Key == "" {
		if account.AccountID != "" {
			account.Key = account.AccountID
		} else {
			account.Key = fmt.Sprintf("%s:%s", account.Source, filepath.Base(account.FilePath))
		}
	}
}

func appendUniqueString(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}

func readJSONMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	root := make(map[string]any)
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	return root, nil
}

func writeJSONMap(path string, root map[string]any) error {
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %w", path, err)
	}

	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	cleanup = false
	return nil
}
