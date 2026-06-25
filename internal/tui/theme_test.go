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
		{"adaptive", "Adaptive", false},
		{"rpg-skill-tree", "Adaptive", false},
		{"clean-productivity", "Adaptive", false},
		{"cyber-dashboard", "Adaptive", false},
		{"", "Adaptive", false},
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
		assert.NotEmpty(t, theme.Palette.Primary)
		assert.NotEmpty(t, theme.Palette.XP)
		assert.NotEmpty(t, theme.Palette.Review)
		assert.NotEmpty(t, theme.XP)
		assert.NotEmpty(t, theme.Review)
		assert.NotEmpty(t, theme.Spinner.Foreground)
		assert.Equal(t, "adaptive", theme.ID)
	}
}

func TestLookupTheme_Invalid(t *testing.T) {
	_, err := LookupTheme("neon-hacker")
	assert.ErrorContains(t, err, "unknown theme")
}

func TestTheme_AdaptiveCompatibility(t *testing.T) {
	adaptive, _ := LookupTheme("adaptive")
	legacy, _ := LookupTheme("rpg-skill-tree")

	assert.Equal(t, adaptive.Name, legacy.Name)
	assert.Equal(t, adaptive.PrimaryAccent, legacy.PrimaryAccent)
	assert.False(t, adaptive.HasAmbientMotion)
	assert.NotEmpty(t, adaptive.Spinner.Render("loading"))
	assert.NotEmpty(t, adaptive.CompactPanel.Render("test"))
}

func TestTheme_FocusedPanelDiffersFromPanel(t *testing.T) {
	theme, _ := LookupTheme("adaptive")

	panel := theme.Panel.Render("test")
	focused := theme.FocusedPanel.Render("test")
	assert.NotEqual(t, panel, focused, "focused should differ from normal panel")
}

func TestLookupSymbolSet(t *testing.T) {
	rich, err := LookupSymbolSet("rich")
	require.NoError(t, err)
	assert.Equal(t, "🔒", rich.Locked)
	assert.Equal(t, "✦", rich.XP)

	plain, err := LookupSymbolSet("plain")
	require.NoError(t, err)
	assert.Equal(t, "[L]", plain.Locked)
	assert.Equal(t, "XP", plain.XP)

	defaultSet, err := LookupSymbolSet("")
	require.NoError(t, err)
	assert.Equal(t, rich, defaultSet)
}

func TestLookupSymbolSet_Invalid(t *testing.T) {
	_, err := LookupSymbolSet("emoji-only")
	assert.ErrorContains(t, err, "unknown symbol_mode")
}
