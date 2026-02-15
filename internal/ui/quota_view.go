package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/deLiseLINO/codex-quota/internal/api"
)

func (m Model) renderWindowsView() string {
	if len(m.Data.Windows) == 0 {
		return "No quota data.\n"
	}

	var s strings.Builder

	for _, window := range m.Data.Windows {
		s.WriteString(GroupHeaderStyle.Render(windowHeader(window)))
		s.WriteString("\n")
		s.WriteString(m.renderWindowRow(window))
		s.WriteString("\n")
	}

	return s.String()
}

func (m Model) renderWindowRow(window api.QuotaWindow) string {
	var s strings.Builder

	ratio := window.LeftPercent / 100
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	name := window.Label
	if len(name) > 33 {
		name = name[:30] + "..."
	}
	alignedName := fmt.Sprintf("%-35s", name)
	bar := m.defaultProgress
	if window.WindowSec == 18000 {
		bar = m.shortProgress
	}

	s.WriteString("    ")
	s.WriteString(LabelStyle.Render(alignedName))
	s.WriteString(" ")
	s.WriteString(bar.ViewAs(ratio))
	s.WriteString(" ")
	s.WriteString(PercentStyle.Render(fmt.Sprintf("%.1f%%", window.LeftPercent)))
	s.WriteString(ResetTimeStyle.Render(formatResetText(window.ResetAt)))
	s.WriteString("\n")

	return s.String()
}

func formatResetText(resetAt time.Time) string {
	if resetAt.IsZero() {
		return "Resets unknown"
	}

	remaining := time.Until(resetAt)
	if remaining <= 0 {
		return "Resets now"
	}

	localReset := resetAt.Local()
	now := time.Now().Local()
	absolute := ""

	if sameDay(localReset, now) {
		absolute = localReset.Format("15:04")
	} else if remaining <= 7*24*time.Hour {
		absolute = localReset.Format("Mon 15:04")
	} else {
		absolute = localReset.Format("01-02 15:04")
	}

	return fmt.Sprintf("Resets %s (%s)", absolute, formatRemainingShort(remaining))
}

func formatRemainingShort(remaining time.Duration) string {
	if remaining <= 0 {
		return "now"
	}

	if remaining < time.Minute {
		return "<1m"
	}

	totalMinutes := int(remaining.Minutes())
	if totalMinutes < 60 {
		return fmt.Sprintf("%dm", totalMinutes)
	}

	totalHours := int(remaining.Hours())
	if totalHours < 24 {
		mins := totalMinutes % 60
		if mins == 0 {
			return fmt.Sprintf("%dh", totalHours)
		}
		return fmt.Sprintf("%dh %dm", totalHours, mins)
	}

	days := totalHours / 24
	hours := totalHours % 24
	if hours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, hours)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func windowHeader(window api.QuotaWindow) string {
	if window.WindowSec == 18000 {
		return "5 hour"
	}
	if window.WindowSec == 604800 {
		return "Weekly"
	}
	return window.Label
}
