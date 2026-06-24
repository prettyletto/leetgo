package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ID               string
	Name             string
	HasAmbientMotion bool
	Labels           ThemeLabels
	Palette          TerminalPalette
	PrimaryAccent    lipgloss.Color
	SecondaryAccent  lipgloss.Color
	Border           lipgloss.Color
	Muted            lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Danger           lipgloss.Color
	XP               lipgloss.Color
	Review           lipgloss.Color
	Panel            lipgloss.Style
	FocusedPanel     lipgloss.Style
	CompactPanel     lipgloss.Style
	Title            lipgloss.Style
	Footer           lipgloss.Style
	Key              lipgloss.Style
	Spinner          lipgloss.Style
}

type ThemeLabels struct {
	PrimaryAction    string
	SecondaryActions string
	Profile          string
	RoadmapContext   string
	LockedItems      string
	PreviewAction    string
}

type TerminalPalette struct {
	Primary lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color
	Danger  lipgloss.Color
	Muted   lipgloss.Color
	Border  lipgloss.Color
	XP      lipgloss.Color
	Review  lipgloss.Color
}

type SymbolSet struct {
	Locked     string
	Unlocked   string
	Solved     string
	Verified   string
	InProgress string
	XP         string
	Review     string
	Pointer    string
	Bullet     string
	Separator  string
}

func LookupSymbolSet(mode string) (SymbolSet, error) {
	switch mode {
	case "", "rich":
		return SymbolSet{
			Locked:     "🔒",
			Unlocked:   "◆",
			Solved:     "✓",
			Verified:   "⚠",
			InProgress: "…",
			XP:         "✦",
			Review:     "↻",
			Pointer:    "➤",
			Bullet:     "•",
			Separator:  "│",
		}, nil
	case "plain":
		return SymbolSet{
			Locked:     "[L]",
			Unlocked:   "[ ]",
			Solved:     "[S]",
			Verified:   "[V]",
			InProgress: "[*]",
			XP:         "XP",
			Review:     "R",
			Pointer:    ">",
			Bullet:     "-",
			Separator:  "|",
		}, nil
	default:
		return SymbolSet{}, fmt.Errorf("unknown symbol_mode %q", mode)
	}
}

func LookupTheme(id string) (*Theme, error) {
	switch id {
	case "", "rpg-skill-tree":
		return newTheme(
			"rpg-skill-tree",
			"RPG Skill Tree",
			false,
			lipgloss.Color("205"),
			lipgloss.Color("219"),
			lipgloss.Color("99"),
			lipgloss.Color("245"),
			lipgloss.Color("82"),
			lipgloss.Color("214"),
			lipgloss.Color("196"),
			lipgloss.Color("220"),
			lipgloss.Color("147"),
		), nil
	case "clean-productivity":
		return newTheme(
			"clean-productivity",
			"Clean Productivity",
			false,
			lipgloss.Color("252"),
			lipgloss.Color("245"),
			lipgloss.Color("240"),
			lipgloss.Color("244"),
			lipgloss.Color("71"),
			lipgloss.Color("179"),
			lipgloss.Color("167"),
			lipgloss.Color("220"),
			lipgloss.Color("111"),
		), nil
	case "cyber-dashboard":
		return newTheme(
			"cyber-dashboard",
			"Cyber Dashboard",
			true,
			lipgloss.Color("51"),
			lipgloss.Color("201"),
			lipgloss.Color("45"),
			lipgloss.Color("86"),
			lipgloss.Color("48"),
			lipgloss.Color("226"),
			lipgloss.Color("198"),
			lipgloss.Color("226"),
			lipgloss.Color("135"),
		), nil
	default:
		return nil, fmt.Errorf("unknown theme %q", id)
	}
}

func newTheme(id, name string, ambient bool, primary, secondary, border, muted, success, warning, danger, xp, review lipgloss.Color) *Theme {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, 1)

	focusedPanel := panel.Copy().
		Border(lipgloss.ThickBorder()).
		BorderForeground(secondary)

	compactPanel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(0, 1)

	spinner := lipgloss.NewStyle().
		Foreground(warning).
		Bold(true)

	if id == "rpg-skill-tree" {
		panel = panel.BorderForeground(border).BorderBackground(lipgloss.Color("235"))
		focusedPanel = focusedPanel.BorderForeground(secondary).BorderBackground(lipgloss.Color("235"))
	}

	if id == "clean-productivity" {
		panel = compactPanel
		focusedPanel = compactPanel.Copy().
			Border(lipgloss.ThickBorder()).
			BorderForeground(primary)
	}

	if id == "cyber-dashboard" {
		spinner = spinner.Foreground(lipgloss.Color("51"))
		focusedPanel = focusedPanel.BorderForeground(lipgloss.Color("201"))
	}

	return &Theme{
		ID:               id,
		Name:             name,
		HasAmbientMotion: ambient,
		Labels:           themeLabels(id),
		Palette: TerminalPalette{
			Primary: primary,
			Success: success,
			Warning: warning,
			Danger:  danger,
			Muted:   muted,
			Border:  border,
			XP:      xp,
			Review:  review,
		},
		PrimaryAccent:   primary,
		SecondaryAccent: secondary,
		Border:          border,
		Muted:           muted,
		Success:         success,
		Warning:         warning,
		Danger:          danger,
		XP:              xp,
		Review:          review,
		Panel:           panel,
		FocusedPanel:    focusedPanel,
		CompactPanel:    compactPanel,
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			Align(lipgloss.Center),
		Footer: lipgloss.NewStyle().
			Foreground(muted).
			Align(lipgloss.Center),
		Key: lipgloss.NewStyle().
			Bold(true).
			Foreground(secondary),
		Spinner: spinner,
	}
}

func themeLabels(id string) ThemeLabels {
	switch id {
	case "clean-productivity":
		return ThemeLabels{
			PrimaryAction:    "Recommended",
			SecondaryActions: "Available",
			Profile:          "Profile",
			RoadmapContext:   "Roadmap",
			LockedItems:      "Upcoming",
			PreviewAction:    "Recommended",
		}
	case "cyber-dashboard":
		return ThemeLabels{
			PrimaryAction:    "Primary Signal",
			SecondaryActions: "Secondary Targets",
			Profile:          "Operator",
			RoadmapContext:   "Signal Map",
			LockedItems:      "Locked Signals",
			PreviewAction:    "Primary Signal",
		}
	default:
		return ThemeLabels{
			PrimaryAction:    "Main Quest",
			SecondaryActions: "Side Quests",
			Profile:          "Character HUD",
			RoadmapContext:   "Map Fragment",
			LockedItems:      "Locked Branches",
			PreviewAction:    "Main Quest",
		}
	}
}
