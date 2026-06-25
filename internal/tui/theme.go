package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ID               string
	Name             string
	HasAmbientMotion bool
	Palette          TerminalPalette
	PrimaryAccent    lipgloss.Color
	SecondaryAccent  lipgloss.Color
	Border           lipgloss.Color
	Muted            lipgloss.Color
	SelectionBg      lipgloss.Color
	SelectionFg      lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Danger           lipgloss.Color
	XP               lipgloss.Color
	Review           lipgloss.Color
	Panel            lipgloss.Style
	FocusedPanel     lipgloss.Style
	CompactPanel     lipgloss.Style
	Title            lipgloss.Style
	Subtitle         lipgloss.Style
	Footer           lipgloss.Style
	Key              lipgloss.Style
	Spinner          lipgloss.Style
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
	term := detectTerminal()
	switch id {
	case "", "adaptive", "rpg-skill-tree", "clean-productivity", "cyber-dashboard":
		return newAdaptiveTheme(term), nil
	default:
		return nil, fmt.Errorf("unknown theme %q", id)
	}
}

func newAdaptiveTheme(term terminalProfile) *Theme {
	var primary, secondary, border, muted, selectionBg, selectionFg, success, warning, danger, xp, review lipgloss.Color

	if term.HasDarkBackground {
		primary = lipgloss.Color("111")
		secondary = lipgloss.Color("153")
		border = lipgloss.Color("239")
		muted = lipgloss.Color("245")
		selectionBg = lipgloss.Color("238")
		selectionFg = lipgloss.Color("255")
		success = lipgloss.Color("114")
		warning = lipgloss.Color("221")
		danger = lipgloss.Color("210")
		xp = lipgloss.Color("223")
		review = lipgloss.Color("182")
	} else {
		primary = lipgloss.Color("25")
		secondary = lipgloss.Color("32")
		border = lipgloss.Color("250")
		muted = lipgloss.Color("244")
		selectionBg = lipgloss.Color("254")
		selectionFg = lipgloss.Color("235")
		success = lipgloss.Color("28")
		warning = lipgloss.Color("130")
		danger = lipgloss.Color("124")
		xp = lipgloss.Color("130")
		review = lipgloss.Color("61")
	}

	return buildTheme(term, primary, secondary, border, muted, selectionBg, selectionFg, success, warning, danger, xp, review)
}

func buildTheme(term terminalProfile, primary, secondary, border, muted, selectionBg, selectionFg, success, warning, danger, xp, review lipgloss.Color) *Theme {
	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(0, 1)

	focusedPanel := panel.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary)

	compactPanel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(0, 1)

	spinner := lipgloss.NewStyle().
		Foreground(warning).
		Bold(true)

	return &Theme{
		ID:               "adaptive",
		Name:             "Adaptive",
		HasAmbientMotion: false,
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
		SelectionBg:     selectionBg,
		SelectionFg:     selectionFg,
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
			Align(lipgloss.Left),
		Subtitle: lipgloss.NewStyle().
			Foreground(muted),
		Footer: lipgloss.NewStyle().
			Foreground(muted).
			Align(lipgloss.Left),
		Key: lipgloss.NewStyle().
			Bold(true).
			Foreground(secondary),
		Spinner: spinner,
	}
}
