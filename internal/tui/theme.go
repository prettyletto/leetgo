package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ID               string
	Name             string
	HasAmbientMotion bool
	PrimaryAccent    lipgloss.Color
	SecondaryAccent  lipgloss.Color
	Border           lipgloss.Color
	Muted            lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Danger           lipgloss.Color
	Panel            lipgloss.Style
	FocusedPanel     lipgloss.Style
	CompactPanel     lipgloss.Style
	Title            lipgloss.Style
	Footer           lipgloss.Style
	Key              lipgloss.Style
	Spinner          lipgloss.Style
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
		), nil
	default:
		return nil, fmt.Errorf("unknown theme %q", id)
	}
}

func newTheme(id, name string, ambient bool, primary, secondary, border, muted, success, warning, danger lipgloss.Color) *Theme {
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
		PrimaryAccent:    primary,
		SecondaryAccent:  secondary,
		Border:           border,
		Muted:            muted,
		Success:          success,
		Warning:          warning,
		Danger:           danger,
		Panel:            panel,
		FocusedPanel:     focusedPanel,
		CompactPanel:     compactPanel,
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
