package ui

import (
	"fmt"
	"strings"

	"github.com/deLiseLINO/codex-quota/internal/api"
)

func (m Model) renderWindowsView() string {
	if len(m.Data.Windows) == 0 {
		return "No quota data.\n"
	}

	var s strings.Builder

	for i, window := range m.Data.Windows {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString(GroupHeaderStyle.Render(windowHeader(window)))
		s.WriteString("\n")
		s.WriteString(m.renderWindowRow(window))
		s.WriteString("\n")
	}

	return s.String()
}

func (m Model) renderWindowsLoadingSkeleton() string {
	windows := make([]api.QuotaWindow, 0, 2)
	if account := m.activeAccount(); account != nil && m.isPaidByKnownPlan(account.Key) {
		windows = append(windows, api.QuotaWindow{Label: "5 hour usage limit", WindowSec: 18000})
	}
	windows = append(windows, api.QuotaWindow{Label: "Weekly usage limit", WindowSec: 604800})
	var s strings.Builder
	for i, window := range windows {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString(GroupHeaderStyle.Render(windowHeader(window)))
		s.WriteString("\n")
		s.WriteString(m.renderWindowStatusRow(window, "Loading..."))
		s.WriteString("\n")
	}
	return s.String()
}

func (m Model) renderWindowRow(window api.QuotaWindow) string {
	var s strings.Builder

	ratio := clampRatio(window.LeftPercent / 100)
	ratio = m.tabWindowRatio(m.activeAccountKey(), window, ratio)

	name := window.Label
	if len(name) > 33 {
		name = name[:30] + "..."
	}
	alignedName := fmt.Sprintf("%-35s", name)
	barWidth := m.barWidthForWindow(window.WindowSec)
	gradientStart, gradientEnd := barGradientForWindow(window.WindowSec)

	s.WriteString(windowRowIndent)
	s.WriteString(LabelStyle.Render(alignedName))
	s.WriteString(" ")
	s.WriteString(renderSmoothBar(barWidth, ratio, gradientStart, gradientEnd))
	s.WriteString(" ")
	s.WriteString(PercentStyle.Render(fmt.Sprintf("%.1f%%", window.LeftPercent)))
	s.WriteString(ResetTimeStyle.Render(formatResetText(window.ResetAt)))

	return s.String()
}

func (m Model) renderWindowStatusRow(window api.QuotaWindow, status string) string {
	var s strings.Builder
	name := window.Label
	if len(name) > 33 {
		name = name[:30] + "..."
	}
	alignedName := fmt.Sprintf("%-35s", name)
	barWidth := m.barWidthForWindow(window.WindowSec)
	gradientStart, gradientEnd := barGradientForWindow(window.WindowSec)

	s.WriteString(windowRowIndent)
	s.WriteString(LabelStyle.Render(alignedName))
	s.WriteString(" ")
	s.WriteString(renderSmoothBar(barWidth, 0, gradientStart, gradientEnd))
	s.WriteString(" ")
	s.WriteString(PercentStyle.Render("..."))
	s.WriteString(ResetTimeStyle.Render(status))
	return s.String()
}
