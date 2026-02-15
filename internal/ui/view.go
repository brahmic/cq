package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("🚀 Codex Quota Monitor"))
	s.WriteString("\n")

	if len(m.Accounts) > 0 {
		s.WriteString(renderAccountTabs(m.Accounts, m.ActiveAccountIx, m.Width))
		s.WriteString("\n\n")
	}

	if m.Loading {
		s.WriteString("Loading...")
		s.WriteString("\n")
	} else if account := m.activeAccount(); account != nil {
		s.WriteString(m.renderWindowsView())
	} else {
		s.WriteString("\n")
	}

	s.WriteString(HelpStyle.Render("\n[r] refresh • [i] additional info • [n] add account • [o] apply to opencode • [x] delete account • [←/→] switch • [q] quit"))

	content := s.String()
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	containerStyle := lipgloss.NewStyle().Padding(1, 2)
	if m.Width > contentWidth+4 && m.Height > contentHeight+2 {
		containerStyle = containerStyle.
			Width(m.Width).
			Height(m.Height).
			Align(lipgloss.Center, lipgloss.Center)
	}

	baseView := containerStyle.Render(content)

	if modal := m.currentOverlayModal(); modal != "" {
		return overlayCenter(baseView, modal, m.Width, m.Height)
	}

	return baseView
}
