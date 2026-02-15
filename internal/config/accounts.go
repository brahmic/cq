package config

import "time"

type Source string

const (
	SourceManaged  Source = "managed"
	SourceOpenCode Source = "opencode"
	SourceCodex    Source = "codex"
)

type Account struct {
	Key          string
	Label        string
	Email        string
	AccountID    string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	ClientID     string
	Source       Source
	FilePath     string
	Writable     bool
}

type AccessTokenClaims struct {
	ClientID  string
	AccountID string
	ExpiresAt time.Time
	Email     string
}

type AccountsLoadResult struct {
	Accounts           []*Account
	SourcesByAccountID map[string][]string
}

func (a *Account) SourceLabel() string {
	switch a.Source {
	case SourceManaged:
		return "app"
	case SourceOpenCode:
		return "opencode"
	case SourceCodex:
		return "codex"
	default:
		return "unknown"
	}
}

func LoadAllAccounts() ([]*Account, error) {
	result, err := LoadAllAccountsWithSources()
	if err != nil {
		return nil, err
	}
	return result.Accounts, nil
}

func SaveAccount(account *Account) error {
	if account == nil || !account.Writable {
		return nil
	}

	switch account.Source {
	case SourceManaged:
		return saveManagedAccount(account)
	case SourceOpenCode:
		if account.FilePath == "" {
			return nil
		}
		return saveOpenCodeAccount(account)
	case SourceCodex:
		if account.FilePath == "" {
			return nil
		}
		return saveCodexAccount(account)
	default:
		return nil
	}
}
