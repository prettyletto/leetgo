package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupTheme(t *testing.T) {
	tests := []struct {
		id      string
		name    string
		ambient bool
	}{
		{"rpg-skill-tree", "RPG Skill Tree", false},
		{"clean-productivity", "Clean Productivity", false},
		{"cyber-dashboard", "Cyber Dashboard", true},
		{"", "RPG Skill Tree", false},
	}

	for _, tt := range tests {
		theme, err := LookupTheme(tt.id)
		require.NoError(t, err)
		assert.Equal(t, tt.name, theme.Name)
		assert.Equal(t, tt.ambient, theme.HasAmbientMotion)
		assert.NotEmpty(t, theme.PrimaryAccent)
		assert.NotEmpty(t, theme.SecondaryAccent)
		assert.NotEmpty(t, theme.Border)
		assert.NotEmpty(t, theme.Muted)
		assert.NotEmpty(t, theme.Spinner.Foreground)
	}
}

func TestLookupTheme_Invalid(t *testing.T) {
	_, err := LookupTheme("neon-hacker")
	assert.ErrorContains(t, err, "unknown theme")
}

func TestTheme_SpinnerPerTheme(t *testing.T) {
	rpg, _ := LookupTheme("rpg-skill-tree")
	clean, _ := LookupTheme("clean-productivity")
	cyber, _ := LookupTheme("cyber-dashboard")

	assert.NotNil(t, rpg.Spinner.Render("loading"))
	assert.NotNil(t, clean.Spinner.Render("loading"))
	assert.NotNil(t, cyber.Spinner.Render("loading"))
}

func TestTheme_CompactPanel(t *testing.T) {
	rpg, _ := LookupTheme("rpg-skill-tree")
	clean, _ := LookupTheme("clean-productivity")
	cyber, _ := LookupTheme("cyber-dashboard")

	assert.NotNil(t, rpg.CompactPanel.Render("test"))
	assert.NotNil(t, clean.CompactPanel.Render("test"))
	assert.NotNil(t, cyber.CompactPanel.Render("test"))
}

func TestTheme_AmbientMotion(t *testing.T) {
	rpg, _ := LookupTheme("rpg-skill-tree")
	clean, _ := LookupTheme("clean-productivity")
	cyber, _ := LookupTheme("cyber-dashboard")

	assert.False(t, rpg.HasAmbientMotion)
	assert.False(t, clean.HasAmbientMotion)
	assert.True(t, cyber.HasAmbientMotion)
}

func TestTheme_FocusedPanelDiffersFromPanel(t *testing.T) {
	rpg, _ := LookupTheme("rpg-skill-tree")
	clean, _ := LookupTheme("clean-productivity")

	rpgPanel := rpg.Panel.Render("test")
	rpgFocused := rpg.FocusedPanel.Render("test")
	assert.NotEqual(t, rpgPanel, rpgFocused, "focused should differ from normal panel in RPG theme")

	cleanPanel := clean.Panel.Render("test")
	cleanFocused := clean.FocusedPanel.Render("test")
	assert.NotEqual(t, cleanPanel, cleanFocused, "focused should differ from normal panel in Clean theme")
}
