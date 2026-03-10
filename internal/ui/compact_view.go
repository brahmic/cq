package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deLiseLINO/codex-quota/internal/config"
)

func (m Model) renderCompactView() string {
	if len(m.Accounts) == 0 {
		return "No accounts.\n"
	}

	accountWidth := 30
	if m.Width > 0 && m.Width < 120 {
		accountWidth = 24
	}
	normalRows := make([]int, 0, len(m.Accounts))
	exhaustedRows := make([]int, 0, len(m.Accounts))

	for i, acc := range m.Accounts {
		if acc == nil {
			continue
		}
		if m.isCompactAccountExhausted(acc.Key) {
			exhaustedRows = append(exhaustedRows, i)
			continue
		}
		normalRows = append(normalRows, i)
	}

	var s strings.Builder
	m.renderCompactRows(&s, normalRows, accountWidth)

	if len(exhaustedRows) > 0 {
		if len(normalRows) > 0 {
			s.WriteString("\n")
		}
		s.WriteString(CompactExhaustedHeaderStyle.Render("Exhausted accounts"))
		s.WriteString("\n")
		m.renderCompactRows(&s, exhaustedRows, accountWidth)
	}

	return s.String()
}

func (m Model) renderCompactRows(s *strings.Builder, rowIndexes []int, accountWidth int) {
	for _, i := range rowIndexes {
		if i < 0 || i >= len(m.Accounts) {
			continue
		}
		acc := m.Accounts[i]
		if acc == nil {
			continue
		}
		s.WriteString(m.renderCompactAccountRow(i, acc, accountWidth))
		s.WriteString("\n")
	}
}

func (m Model) renderCompactAccountRow(index int, acc *config.Account, accountWidth int) string {
	var s strings.Builder
	isActive := index == m.ActiveAccountIx
	prefix := "  "
	if isActive {
		prefix = "> "
	}

	name := acc.Label
	if name == "" {
		name = acc.SourceLabel()
	}
	subscribed := m.hasSubscription(acc)
	badgeWidth := m.activeSourceBadgesDisplayWidth(acc)
	nameWidth := accountWidth
	if badgeWidth > 0 {
		nameWidth = accountWidth - badgeWidth - 1
		if nameWidth < 4 {
			nameWidth = 4
		}
	}
	name = truncateLabel(name, nameWidth-1)
	alignedName := fmt.Sprintf("%-*s", nameWidth, name)

	s.WriteString(prefix)
	if badgeWidth > 0 {
		s.WriteString(m.renderActiveSourceBadges(acc, isActive))
		s.WriteString(" ")
	}
	if subscribed && isActive {
		s.WriteString(SubscribedLabelActiveStyle.Render(alignedName))
	} else if subscribed {
		s.WriteString(SubscribedLabelMutedStyle.Render(alignedName))
	} else if isActive {
		s.WriteString(TabActiveStyle.Render(alignedName))
	} else {
		s.WriteString(LabelStyle.Render(alignedName))
	}
	s.WriteString(" ")

	if err := m.ErrorsMap[acc.Key]; err != nil {
		status := truncateLabel("Error: "+err.Error(), 24)
		s.WriteString(m.renderCompactStatusRow(status, subscribed))
		return s.String()
	}
	if m.LoadingMap[acc.Key] {
		s.WriteString(m.renderCompactStatusRow("Loading...", subscribed))
		return s.String()
	}

	data, ok := m.UsageData[acc.Key]
	if !ok {
		s.WriteString(m.renderCompactStatusRow("Queued...", subscribed))
		return s.String()
	}

	window, ok := compactPrimaryWindow(data)
	if !ok {
		s.WriteString(m.renderCompactStatusRow("No quota data", subscribed))
		return s.String()
	}

	ratio := m.compactBarRatio(acc.Key, clampRatio(window.LeftPercent/100))
	s.WriteString(renderSmoothBar(m.defaultBarWidth(), ratio, defaultBarGradientStart, defaultBarGradientEnd))
	s.WriteString(" ")
	s.WriteString(m.renderCompactPercent(fmt.Sprintf("%.1f%%", window.LeftPercent), subscribed))
	s.WriteString(ResetTimeStyle.Render(formatResetText(window.ResetAt)))
	return s.String()
}

func (m Model) isCompactAccountExhausted(accountKey string) bool {
	if accountKey == "" {
		return false
	}
	if m.ExhaustedSticky[accountKey] {
		return true
	}
	if m.LoadingMap[accountKey] {
		return false
	}
	if err := m.ErrorsMap[accountKey]; err != nil {
		return false
	}

	data, ok := m.UsageData[accountKey]
	if !ok {
		return false
	}
	return isConfirmedExhausted(data)
}

func (m Model) renderCompactStatusRow(status string, subscribed bool) string {
	row := renderSmoothBar(m.defaultBarWidth(), 0, defaultBarGradientStart, defaultBarGradientEnd)
	row += " "
	row += m.renderCompactPercent("...", subscribed)
	row += ResetTimeStyle.Render(truncateLabel(status, 24))
	return TabInactiveStyle.Render(row)
}

func (m Model) renderCompactPercent(value string, subscribed bool) string {
	if !subscribed {
		return PercentStyle.Render(value)
	}

	return PercentStyle.Copy().Foreground(lipgloss.Color("177")).Render(value)
}
