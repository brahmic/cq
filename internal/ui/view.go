package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const headerUpdateHintGap = 6

func (m Model) View() string {
	var s strings.Builder

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

	footer := "[r] refresh • [R] refresh all • [i] info • [n] add • [enter/o] apply • [x] del • [v] view • [↑↓←→] switch • [q] quit"
	s.WriteString(HelpStyle.Render("\n" + footer))

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

	if modal := m.currentOverlayModal(); modal != "" {
		return overlayCenter(baseView, modal, m.Width, m.Height)
	}

	return baseView
}

func (m Model) renderHeader() string {
	return TitleStyle.Render("🚀 Codex Quota Monitor")
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
