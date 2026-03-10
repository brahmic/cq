package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const headerUpdateHintGap = 6

func (m Model) View() string {
	var s strings.Builder
	modal := m.currentOverlayModal()

	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	if len(m.Accounts) > 0 {
		if !m.CompactMode {
			s.WriteString(m.renderAccountTabs())
			s.WriteString("\n\n")
		} else {
			s.WriteString("\n")
		}
	}

	if m.CompactMode {
		s.WriteString(m.renderCompactView())
	} else {
		if m.Loading {
			s.WriteString(m.renderWindowsLoadingSkeleton())
		} else if account := m.activeAccount(); account != nil {
			s.WriteString(m.renderWindowsView())
		} else {
			s.WriteString("\n")
		}
	}

	footer := HelpStyle.Render("\n" + m.renderFooter())
	s.WriteString(footer)

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
	baseView = m.overlayUpdateHint(baseView)

	if modal != "" {
		body, footerArea := splitFooterArea(baseView, lipgloss.Height(footer))
		return joinFooterArea(overlayCenter(body, modal, m.Width, m.Height-lipgloss.Height(footer)), footerArea)
	}

	return baseView
}

func (m Model) renderHeader() string {
	return TitleStyle.Render("🚀 Codex Quota Monitor")
}

func (m Model) renderFooter() string {
	if m.CompactMode {
		return "↑↓ Move • Enter Menu • ? Help • q Quit"
	}
	return "←→ Move • Enter Menu • ? Help • q Quit"
}

func (m Model) overlayUpdateHint(base string) string {
	hint := strings.TrimSpace(m.UpdateAvailableHint)
	if hint == "" {
		return base
	}

	lines := strings.Split(base, "\n")
	if len(lines) == 0 {
		return base
	}

	canvasWidth := 0
	for _, line := range lines {
		if width := ansi.StringWidth(line); width > canvasWidth {
			canvasWidth = width
		}
	}
	if canvasWidth == 0 {
		return base
	}

	titleIdx := firstNonEmptyLine(lines)
	if titleIdx < 0 {
		return base
	}

	hintRendered := UpdateHintStyle.Render(hint)
	hintWidth := ansi.StringWidth(hintRendered)
	if hintWidth+2 > canvasWidth {
		return base
	}

	candidates := []int{titleIdx, titleIdx + 1}
	for _, idx := range candidates {
		if idx < 0 || idx >= len(lines) {
			continue
		}
		rightEdge := lineRightEdge(lines[idx])
		startX := canvasWidth - hintWidth
		if idx == titleIdx {
			startX = rightEdge + headerUpdateHintGap
		}
		if startX+hintWidth > canvasWidth {
			continue
		}
		if startX < rightEdge+2 {
			continue
		}

		line := padANSI(lines[idx], canvasWidth)
		left := ansi.Cut(line, 0, startX)
		right := ansi.Cut(line, startX+hintWidth, canvasWidth)
		lines[idx] = left + hintRendered + right
		return strings.Join(lines, "\n")
	}

	return base
}

func firstNonEmptyLine(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(ansi.Strip(line)) != "" {
			return i
		}
	}
	return -1
}

func lineRightEdge(line string) int {
	plain := ansi.Strip(line)
	return ansi.StringWidth(strings.TrimRight(plain, " "))
}

func splitFooterArea(view string, footerHeight int) (string, string) {
	if footerHeight <= 0 {
		return view, ""
	}
	lines := strings.Split(view, "\n")
	if footerHeight >= len(lines) {
		return view, ""
	}
	body := strings.Join(lines[:len(lines)-footerHeight], "\n")
	footer := strings.Join(lines[len(lines)-footerHeight:], "\n")
	return body, footer
}

func joinFooterArea(body, footer string) string {
	if strings.TrimSpace(footer) == "" {
		return body
	}
	if body == "" {
		return footer
	}
	return body + "\n" + footer
}
