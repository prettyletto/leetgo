package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

func viewPalette(theme *Theme) views.Palette {
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
	switch theme.ID {
	case "clean-productivity":
		return "Problem Detail", "Problem Brief", "Files"
	case "cyber-dashboard":
		return "Target Detail", "Signal Notes", "Workspace Artifacts"
	default:
		return "Skill Tile", "Training Notes", "Training Files"
	}
}

func themeStageLabels(theme *Theme) (stagePrefix, grid, recommended, review string) {
	switch theme.ID {
	case "clean-productivity":
		return "Stage", "Problems", "Recommended", "Review"
	case "cyber-dashboard":
		return "Sector", "Target Grid", "Priority Target", "Review Signal"
	default:
		return "Zone", "Encounter Grid", "Recommended Encounter", "Review Shrine"
	}
}

func themeRoadmapLabels(theme *Theme) (titlePrefix, stagePrefix, locked string) {
	switch theme.ID {
	case "clean-productivity":
		return "Roadmap", "Stage", "Upcoming"
	case "cyber-dashboard":
		return "Signal Map", "Sector", "Locked Signals"
	default:
		return "World Map", "Zone", "Locked Branches"
	}
}

func renderStatusPill(theme *Theme, symbols SymbolSet, status roadmap.Status) string {
	label := statusLabel(status)
	if symbols.Locked != "[L]" {
		label = statusSymbol(symbols, status) + " " + label
	}
	return views.StatusPill(label, statusColor(theme, status))
}

func renderThemedPanel(theme *Theme, title, body string, focused bool) string {
	switch theme.ID {
	case "clean-productivity":
		return views.Panel(title, body, viewPalette(theme), focused)
	case "cyber-dashboard":
		return views.PixelFrame("SYS: "+title, body, viewPalette(theme))
	default:
		return views.PixelFrame(title, body, viewPalette(theme))
	}
}
