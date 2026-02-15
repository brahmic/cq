package ui

import (
	"strings"

	"github.com/deLiseLINO/codex-quota/internal/config"
)

func renderAccountTabs(accounts []*config.Account, activeIndex, width int) string {
	if len(accounts) == 0 {
		return ""
	}

	maxVisible := 3
	if width >= 180 {
		maxVisible = 5
	} else if width >= 130 {
		maxVisible = 4
	}

	start := 0
	end := len(accounts)
	if len(accounts) > maxVisible {
		half := maxVisible / 2
		start = activeIndex - half
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(accounts) {
			end = len(accounts)
			start = end - maxVisible
		}
	}

	tabs := make([]string, 0, maxVisible+2)
	if start > 0 {
		tabs = append(tabs, TabInactiveStyle.Render("..."))
	}

	for i := start; i < end; i++ {
		account := accounts[i]
		label := account.Label
		if label == "" {
			label = account.SourceLabel()
		}
		label = truncateLabel(label, 28)

		style := TabInactiveStyle
		if i == activeIndex {
			style = TabActiveStyle
		}

		tabs = append(tabs, style.Render(label))
	}

	if end < len(accounts) {
		tabs = append(tabs, TabInactiveStyle.Render("..."))
	}

	return strings.Join(tabs, " • ")
}

func truncateLabel(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}
