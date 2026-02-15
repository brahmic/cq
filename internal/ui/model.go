package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/deLiseLINO/codex-quota/internal/api"
	"github.com/deLiseLINO/codex-quota/internal/auth"
	"github.com/deLiseLINO/codex-quota/internal/config"
)

type Model struct {
	defaultProgress    progress.Model
	shortProgress      progress.Model
	Data               api.UsageData
	Loading            bool
	DeleteConfirm      bool
	ApplyConfirm       bool
	ShowInfo           bool
	Notice             string
	noticeSeq          int
	Err                error
	Width              int
	Height             int
	Accounts           []*config.Account
	SourcesByAccountID map[string][]string
	ActiveAccountIx    int
}

func InitialModel(accounts []*config.Account, sourcesByAccountID map[string][]string) Model {
	defaultProgress := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	shortProgress := progress.New(
		progress.WithGradient("#4285F4", "#34A853"),
		progress.WithoutPercentage(),
	)

	return Model{
		defaultProgress:    defaultProgress,
		shortProgress:      shortProgress,
		Loading:            len(accounts) > 0,
		Accounts:           accounts,
		SourcesByAccountID: sourcesByAccountID,
		ActiveAccountIx:    0,
	}
}

func (m Model) Init() tea.Cmd {
	titleCmd := tea.SetWindowTitle("🚀 Codex Quota Monitor")
	if account := m.activeAccount(); account != nil {
		return tea.Batch(titleCmd, FetchDataCmd(account))
	}
	return titleCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "x", "delete":
			if len(m.Accounts) == 0 {
				return m, nil
			}
			account := m.activeAccount()
			if account == nil {
				return m, nil
			}
			if account.Source != config.SourceManaged {
				m.DeleteConfirm = false
				m.ApplyConfirm = false
				m.ShowInfo = false
				m.Err = nil
				m.Notice = "cannot delete this account (read-only); only accounts added via n can be deleted"
				m.noticeSeq++
				return m, scheduleNoticeClearCmd(m.noticeSeq)
			}

			m.DeleteConfirm = true
			m.ApplyConfirm = false
			m.ShowInfo = false
			m.Notice = ""
			return m, nil

		case "esc":
			if m.DeleteConfirm {
				m.DeleteConfirm = false
				return m, nil
			}
			if m.ApplyConfirm {
				m.ApplyConfirm = false
				return m, nil
			}
			if m.ShowInfo {
				m.ShowInfo = false
				return m, nil
			}
			if m.Err != nil {
				m.Err = nil
				return m, nil
			}
			if m.Notice != "" {
				m.Notice = ""
				return m, nil
			}
			return m, tea.Quit

		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			if m.activeAccount() == nil {
				return m, nil
			}
			m.Loading = true
			m.Err = nil
			m.DeleteConfirm = false
			m.ApplyConfirm = false
			m.Notice = ""
			return m, FetchDataCmd(m.activeAccount())

		case "i":
			m.ShowInfo = !m.ShowInfo
			m.DeleteConfirm = false
			m.ApplyConfirm = false
			m.Notice = ""
			return m, nil

		case "n":
			m.Loading = true
			m.Err = nil
			m.DeleteConfirm = false
			m.ApplyConfirm = false
			m.ShowInfo = false
			m.Notice = ""
			return m, AddAccountCmd()

		case "o":
			if m.activeAccount() == nil {
				return m, nil
			}
			m.ApplyConfirm = true
			m.DeleteConfirm = false
			m.ShowInfo = false
			m.Notice = ""
			return m, nil

		case "enter":
			if m.DeleteConfirm {
				account := m.activeAccount()
				if account == nil {
					m.DeleteConfirm = false
					return m, nil
				}

				nextKey := ""
				if len(m.Accounts) > 1 {
					if m.ActiveAccountIx < len(m.Accounts)-1 {
						nextKey = m.Accounts[m.ActiveAccountIx+1].Key
					} else if m.ActiveAccountIx > 0 {
						nextKey = m.Accounts[m.ActiveAccountIx-1].Key
					}
				}

				m.Loading = true
				m.Err = nil
				m.Notice = ""
				m.ShowInfo = false
				m.ApplyConfirm = false
				m.DeleteConfirm = false
				m.Data = api.UsageData{}
				return m, DeleteManagedAccountCmd(account.AccountID, nextKey)
			}

			if m.ApplyConfirm {
				account := m.activeAccount()
				if account == nil {
					m.ApplyConfirm = false
					return m, nil
				}

				m.Loading = true
				m.Err = nil
				m.DeleteConfirm = false
				m.ShowInfo = false
				m.Notice = ""
				m.ApplyConfirm = false
				return m, ApplyToOpenCodeCmd(account)
			}

			return m, nil

		case "right", "l":
			if len(m.Accounts) > 1 {
				m.ActiveAccountIx = (m.ActiveAccountIx + 1) % len(m.Accounts)
				m.Loading = true
				m.Err = nil
				m.DeleteConfirm = false
				m.ApplyConfirm = false
				m.Notice = ""
				m.Data = api.UsageData{}
				return m, FetchDataCmd(m.activeAccount())
			}

		case "left", "h":
			if len(m.Accounts) > 1 {
				m.ActiveAccountIx = (m.ActiveAccountIx - 1 + len(m.Accounts)) % len(m.Accounts)
				m.Loading = true
				m.Err = nil
				m.DeleteConfirm = false
				m.ApplyConfirm = false
				m.Notice = ""
				m.Data = api.UsageData{}
				return m, FetchDataCmd(m.activeAccount())
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		barWidth := msg.Width - 72
		if barWidth < 20 {
			barWidth = 20
		}
		if barWidth > 50 {
			barWidth = 50
		}
		m.defaultProgress.Width = barWidth
		m.shortProgress.Width = barWidth

	case AccountsMsg:
		m.Accounts = msg.Accounts
		m.SourcesByAccountID = msg.SourcesByAccountID
		m.ActiveAccountIx = 0
		m.Data = api.UsageData{}

		if len(m.Accounts) == 0 {
			m.Loading = false
			m.Err = fmt.Errorf("no accounts found; press n to add account")
			m.Notice = ""
			return m, nil
		}

		if msg.ActiveKey != "" {
			for i, account := range m.Accounts {
				if account != nil && account.Key == msg.ActiveKey {
					m.ActiveAccountIx = i
					break
				}
			}
		}

		m.Loading = true
		m.Err = nil
		m.Notice = msg.Notice
		fetchCmd := FetchDataCmd(m.activeAccount())
		if msg.Notice != "" {
			m.noticeSeq++
			return m, tea.Batch(fetchCmd, scheduleNoticeClearCmd(m.noticeSeq))
		}
		return m, fetchCmd

	case DataMsg:
		m.applyAccountSnapshot(msg.AccountKey, msg.Account)
		if msg.AccountKey != "" && msg.AccountKey != m.activeAccountKey() {
			return m, nil
		}
		m.Data = msg.Data
		m.Loading = false
		m.Err = nil
		return m, nil

	case NoticeMsg:
		m.Loading = false
		m.Err = nil
		m.Notice = msg.Text
		if msg.Text == "" {
			return m, nil
		}
		m.noticeSeq++
		return m, scheduleNoticeClearCmd(m.noticeSeq)

	case NoticeTimeoutMsg:
		if msg.Seq != m.noticeSeq {
			return m, nil
		}
		m.Notice = ""
		return m, nil

	case ErrMsg:
		if msg.AccountKey != "" && msg.AccountKey != m.activeAccountKey() {
			return m, nil
		}
		m.Loading = false
		m.Err = msg.Err
		m.Notice = ""
		return m, nil

	case progress.FrameMsg:
		defaultModel, defaultCmd := m.defaultProgress.Update(msg)
		m.defaultProgress = defaultModel.(progress.Model)

		shortModel, shortCmd := m.shortProgress.Update(msg)
		m.shortProgress = shortModel.(progress.Model)

		return m, tea.Batch(defaultCmd, shortCmd)
	}

	return m, nil
}

func AddAccountCmd() tea.Cmd {
	return func() tea.Msg {
		account, err := auth.LoginOpenAICodex()
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("login failed: %w", err)}
		}
		if err := config.UpsertManagedAccount(account); err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to save account: %w", err)}
		}

		result, err := config.LoadAllAccountsWithSources()
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to reload accounts: %w", err)}
		}

		note := "account added"
		if account.Email != "" {
			note = "account added: " + account.Email
		}

		return AccountsMsg{ActiveKey: account.AccountID, Accounts: result.Accounts, SourcesByAccountID: result.SourcesByAccountID, Notice: note}
	}
}

func ApplyToOpenCodeCmd(account *config.Account) tea.Cmd {
	accountSnapshot := cloneAccount(account)
	if accountSnapshot == nil {
		return nil
	}

	return func() tea.Msg {
		appliedPath, err := config.ApplyAccountToOpenCode(accountSnapshot)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("apply to OpenCode failed: %w", err)}
		}

		result, loadErr := config.LoadAllAccountsWithSources()
		if loadErr != nil {
			return NoticeMsg{Text: fmt.Sprintf("applied to OpenCode: %s", appliedPath)}
		}

		note := fmt.Sprintf("applied to OpenCode: %s", filepath.Base(appliedPath))

		return AccountsMsg{ActiveKey: accountSnapshot.AccountID, Accounts: result.Accounts, SourcesByAccountID: result.SourcesByAccountID, Notice: note}
	}
}

func DeleteManagedAccountCmd(accountID string, nextActiveKey string) tea.Cmd {
	return func() tea.Msg {
		if err := config.DeleteManagedAccount(accountID); err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to delete account: %w", err)}
		}

		result, err := config.LoadAllAccountsWithSources()
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to reload accounts: %w", err)}
		}
		return AccountsMsg{ActiveKey: nextActiveKey, Accounts: result.Accounts, SourcesByAccountID: result.SourcesByAccountID, Notice: "account deleted"}
	}
}

func FetchDataCmd(account *config.Account) tea.Cmd {
	accountSnapshot := cloneAccount(account)
	if accountSnapshot == nil {
		return nil
	}

	accountKey := accountSnapshot.Key

	return func() tea.Msg {
		workingAccount := *accountSnapshot

		if auth.IsExpired(&workingAccount) {
			if err := auth.RefreshToken(&workingAccount); err != nil {
				return ErrMsg{AccountKey: accountKey, Err: fmt.Errorf("token refresh failed: %w", err)}
			}
		}

		data, err := api.CallAPI(workingAccount.AccessToken, workingAccount.AccountID)
		if err != nil && api.IsUnauthorized(err) && workingAccount.RefreshToken != "" {
			if refreshErr := auth.RefreshToken(&workingAccount); refreshErr != nil {
				return ErrMsg{AccountKey: accountKey, Err: fmt.Errorf("token refresh failed: %w", refreshErr)}
			}
			data, err = api.CallAPI(workingAccount.AccessToken, workingAccount.AccountID)
		}

		if err != nil {
			return ErrMsg{AccountKey: accountKey, Err: err}
		}

		return DataMsg{AccountKey: accountKey, Data: data, Account: &workingAccount}
	}
}

func ReloadAccountsCmd(activeKey string) tea.Cmd {
	return func() tea.Msg {
		result, err := config.LoadAllAccountsWithSources()
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to reload accounts: %w", err)}
		}
		return AccountsMsg{ActiveKey: activeKey, Accounts: result.Accounts, SourcesByAccountID: result.SourcesByAccountID}
	}
}

func scheduleNoticeClearCmd(seq int) tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(_ time.Time) tea.Msg {
		return NoticeTimeoutMsg{Seq: seq}
	})
}

func cloneAccount(account *config.Account) *config.Account {
	if account == nil {
		return nil
	}

	cloned := *account
	return &cloned
}

func (m *Model) applyAccountSnapshot(accountKey string, snapshot *config.Account) {
	if snapshot == nil || accountKey == "" {
		return
	}

	for _, account := range m.Accounts {
		if account == nil || account.Key != accountKey {
			continue
		}

		account.AccessToken = snapshot.AccessToken
		account.RefreshToken = snapshot.RefreshToken
		account.ExpiresAt = snapshot.ExpiresAt
		if snapshot.ClientID != "" {
			account.ClientID = snapshot.ClientID
		}
		if snapshot.AccountID != "" {
			account.AccountID = snapshot.AccountID
		}
		if snapshot.Email != "" {
			account.Email = snapshot.Email
		}
		if snapshot.Label != "" {
			account.Label = snapshot.Label
		}

		return
	}
}

func (m Model) activeAccount() *config.Account {
	if len(m.Accounts) == 0 {
		return nil
	}
	if m.ActiveAccountIx < 0 || m.ActiveAccountIx >= len(m.Accounts) {
		return nil
	}
	return m.Accounts[m.ActiveAccountIx]
}

func (m Model) activeAccountKey() string {
	account := m.activeAccount()
	if account == nil {
		return ""
	}
	return account.Key
}
