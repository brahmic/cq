package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deLiseLINO/codex-quota/internal/config"
	"github.com/deLiseLINO/codex-quota/internal/ui"
)

func main() {
	loadResult, err := config.LoadAllAccountsWithSources()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load accounts: %v\n", err)
	}

	p := tea.NewProgram(ui.InitialModel(loadResult.Accounts, loadResult.SourcesByAccountID), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
