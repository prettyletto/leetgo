package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

func viewPalette(theme *Theme) views.Palette {
	if theme == nil {
		return views.Palette{
			Primary: "111",
			Success: "114",
			Warning: "221",
			Danger:  "210",
			Muted:   "245",
			Border:  "239",
			XP:      "223",
			Review:  "182",
		}
	}
	return views.Palette{
		Primary: theme.Palette.Primary,
		Success: theme.Palette.Success,
		Warning: theme.Palette.Warning,
		Danger:  theme.Palette.Danger,
		Muted:   theme.Palette.Muted,
		Border:  theme.Palette.Border,
		XP:      theme.Palette.XP,
		Review:  theme.Palette.Review,
	}
}

func statusColor(theme *Theme, status roadmap.Status) lipgloss.Color {
	switch status {
	case roadmap.StatusSolved:
		return theme.Success
	case roadmap.StatusVerified:
		return theme.Warning
	case roadmap.StatusInProgress:
		return theme.PrimaryAccent
	case roadmap.StatusAvailable:
		return theme.PrimaryAccent
	case roadmap.StatusLocked:
		return theme.Muted
	default:
		return theme.Muted
	}
}

func statusLabel(status roadmap.Status) string {
	switch status {
	case roadmap.StatusSolved:
		return "SOLVED"
	case roadmap.StatusVerified:
		return "VERIFIED !"
	case roadmap.StatusInProgress:
		return "ACTIVE"
	case roadmap.StatusAvailable:
		return "READY"
	case roadmap.StatusLocked:
		return "LOCKED"
	default:
		return "LOCKED"
	}
}

func statusSymbol(symbols SymbolSet, status roadmap.Status) string {
	switch status {
	case roadmap.StatusSolved:
		return symbols.Solved
	case roadmap.StatusVerified:
		return symbols.Verified
	case roadmap.StatusInProgress:
		return symbols.InProgress
	case roadmap.StatusAvailable:
		return symbols.Unlocked
	case roadmap.StatusLocked:
		return symbols.Locked
	default:
		return symbols.Locked
	}
}

func themeProblemLabels(theme *Theme) (screen, brief, files string) {
	return "Problem Detail", "Problem Brief", "Workspace Files"
}

func themeStageLabels(theme *Theme) (stagePrefix, grid, recommended, review string) {
	return "Stage", "Problems", "Recommended", "Review"
}

func themeRoadmapLabels(theme *Theme) (titlePrefix, stagePrefix, locked string) {
	return "Roadmap", "Stage", "Upcoming"
}

func renderStatusPill(theme *Theme, symbols SymbolSet, status roadmap.Status) string {
	label := statusLabel(status)
	if symbols.Locked != "[L]" {
		label = statusSymbol(symbols, status) + " " + label
	}
	return views.StatusPill(label, statusColor(theme, status))
}

func renderThemedPanel(theme *Theme, title, body string, focused bool) string {
	return views.Panel(title, body, viewPalette(theme), focused)
}

func renderSelectedLine(theme *Theme, text string) string {
	return lipgloss.NewStyle().
		Background(theme.SelectionBg).
		Foreground(theme.SelectionFg).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

func renderSelectableLine(theme *Theme, focused bool, text string) string {
	if focused {
		return renderSelectedLine(theme, text)
	}
	return text
}

func renderSelectableBlock(theme *Theme, focused bool, body string) string {
	lines := strings.Split(body, "\n")
	if !focused {
		for i, line := range lines {
			lines[i] = lipgloss.NewStyle().Padding(0, 1).Render(line)
		}
		return strings.Join(lines, "\n")
	}

	rail := lipgloss.NewStyle().Foreground(theme.PrimaryAccent).Bold(true).Render("▎")

	maxLen := 0
	for _, line := range lines {
		w := lipgloss.Width(line)
		if w > maxLen {
			maxLen = w
		}
	}

	selected := lipgloss.NewStyle().
		Background(theme.SelectionBg).
		Foreground(theme.SelectionFg).
		Bold(true).
		Padding(0, 1).
		Width(maxLen + 2)

	for i, line := range lines {
		if line == "" {
			line = " "
		}
		lines[i] = rail + selected.Render(line)
	}

	return strings.Join(lines, "\n")
}

func renderScreenHeader(theme *Theme, title, subtitle string) string {
	parts := []string{theme.Title.Render(title)}
	if strings.TrimSpace(subtitle) != "" {
		parts = append(parts, theme.Subtitle.Render(subtitle))
	}
	return strings.Join(parts, "\n")
}

func renderScreenShell(theme *Theme, width, height int, header, body, footer string) string {
	innerWidth := responsiveShellContentWidth(width)
	style := lipgloss.NewStyle().Padding(1, 2)
	if innerWidth > 0 {
		style = style.Width(innerWidth).MaxWidth(innerWidth)
	}
	header = strings.TrimSpace(header)
	body = strings.TrimSpace(body)
	footer = strings.TrimSpace(footer)
	if height > 0 && height < 28 {
		body = fitBodyToShellHeight(header, body, footer, height)
	}
	content := style.Render(header + "\n\n" + body + "\n\n" + footer)
	if width <= 0 || height <= 0 {
		return content
	}
	horizontal := lipgloss.Center
	vertical := lipgloss.Center
	if width < 118 || height < 28 {
		horizontal = lipgloss.Left
		vertical = lipgloss.Top
	}
	return lipgloss.Place(width, height, horizontal, vertical, content)
}

func fitBodyToShellHeight(header, body, footer string, height int) string {
	headerLines := renderedBlockLineCount(header)
	footerLines := renderedBlockLineCount(footer)
	// Separators, shell padding, and a small safety margin for panel borders.
	maxBodyLines := height - headerLines - footerLines - 7
	if maxBodyLines < 1 {
		maxBodyLines = 1
	}
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) <= maxBodyLines {
		return body
	}
	if maxBodyLines == 1 {
		return "..."
	}
	trimmed := append([]string{}, bodyLines[:maxBodyLines-1]...)
	trimmed = append(trimmed, "...")
	return strings.Join(trimmed, "\n")
}

func renderedBlockLineCount(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func responsiveShellContentWidth(width int) int {
	if width <= 0 {
		return 0
	}
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentWidth > 142 {
		contentWidth = 142
	}
	return contentWidth
}
