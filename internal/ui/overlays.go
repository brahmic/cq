package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) currentOverlayModal() string {
	if m.ShowInfo {
		return m.renderInfoModal()
	}

	if m.DeleteConfirm {
		return renderMessageModal(
			"Delete account",
			"Delete this account?\n[enter] Confirm   [esc] Cancel",
			WarningStyle,
		)
	}

	if m.ApplyConfirm {
		return renderMessageModal(
			"Apply to OpenCode",
			"Apply this account to OpenCode auth?\n[enter] Confirm   [esc] Cancel",
			WarningStyle,
		)
	}

	if m.Err != nil {
		return renderMessageModal("Error", m.Err.Error(), ErrorStyle)
	}

	if m.Notice != "" {
		return renderMessageModal("Notice", m.Notice, NoticeStyle)
	}

	if m.activeAccount() == nil {
		return renderMessageModal("No accounts", "No accounts loaded.\nPress n to add account.", WarningStyle)
	}

	return ""
}

func renderMessageModal(title, message string, titleStyle lipgloss.Style) string {
	content := strings.Join([]string{
		titleStyle.Render(title),
		InfoValueStyle.Render(message),
	}, "\n\n")
	return InfoBoxStyle.Copy().Width(64).Render(content)
}

func (m Model) renderInfoModal() string {
	account := m.activeAccount()

	email := "n/a"
	accountID := "n/a"
	source := "n/a"
	if account != nil {
		if account.Email != "" {
			email = account.Email
		}
		if account.AccountID != "" {
			accountID = account.AccountID
		}
		source = account.SourceLabel()
		if account.AccountID != "" && m.SourcesByAccountID != nil {
			if sources := m.SourcesByAccountID[account.AccountID]; len(sources) > 0 {
				source = strings.Join(sources, ", ")
			}
		}
	}

	plan := m.Data.PlanType
	if plan == "" {
		plan = "n/a"
	}

	allowed := "n/a"
	limitReached := "n/a"
	if m.Data.PlanType != "" || len(m.Data.Windows) > 0 {
		allowed = boolText(m.Data.Allowed)
		limitReached = boolText(m.Data.LimitReached)
	}

	lines := []string{
		InfoTitleStyle.Render("Additional info"),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("email:"), InfoValueStyle.Render(email)),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("account_id:"), InfoValueStyle.Render(accountID)),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("source:"), InfoValueStyle.Render(source)),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("plan_type:"), InfoValueStyle.Render(plan)),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("allowed:"), InfoValueStyle.Render(allowed)),
		fmt.Sprintf("%s %s", InfoKeyStyle.Render("limit_reached:"), InfoValueStyle.Render(limitReached)),
	}

	content := strings.Join(lines, "\n")
	return InfoBoxStyle.Copy().Width(60).Render(content)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func overlayCenter(base, modal string, width, height int) string {
	canvasWidth := width
	if canvasWidth < lipgloss.Width(base) {
		canvasWidth = lipgloss.Width(base)
	}
	if canvasWidth < lipgloss.Width(modal)+2 {
		canvasWidth = lipgloss.Width(modal) + 2
	}

	canvasHeight := height
	if canvasHeight < lipgloss.Height(base) {
		canvasHeight = lipgloss.Height(base)
	}
	if canvasHeight < lipgloss.Height(modal)+2 {
		canvasHeight = lipgloss.Height(modal) + 2
	}

	baseCanvas := lipgloss.Place(canvasWidth, canvasHeight, lipgloss.Left, lipgloss.Top, base)
	baseLines := strings.Split(baseCanvas, "\n")
	modalLines := strings.Split(modal, "\n")

	modalWidth := lipgloss.Width(modal)
	modalHeight := len(modalLines)
	startX := (canvasWidth - modalWidth) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (canvasHeight - modalHeight) / 2
	if startY < 0 {
		startY = 0
	}

	for i, modalLine := range modalLines {
		y := startY + i
		if y < 0 || y >= len(baseLines) {
			continue
		}

		line := padANSI(baseLines[y], canvasWidth)
		modalLine = padANSI(modalLine, modalWidth)

		left := ansi.Cut(line, 0, startX)
		right := ansi.Cut(line, startX+modalWidth, canvasWidth)
		baseLines[y] = left + modalLine + right
	}

	return strings.Join(baseLines, "\n")
}

func padANSI(line string, targetWidth int) string {
	currentWidth := ansi.StringWidth(line)
	if currentWidth >= targetWidth {
		return line
	}
	return line + strings.Repeat(" ", targetWidth-currentWidth)
}
