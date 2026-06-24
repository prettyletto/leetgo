package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg, err := DefaultConfig()
	require.NoError(t, err)
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "from-zero-to-hero", cfg.Roadmap)
	assert.Contains(t, cfg.Workspace, "leetgo-workspace")
	assert.False(t, cfg.OnboardingComplete)
	assert.Zero(t, cfg.OnboardingVersion)
	assert.Empty(t, cfg.DisplayName)
	assert.Equal(t, "rpg-skill-tree", cfg.Theme)
	assert.Equal(t, "rich", cfg.SymbolMode)
	assert.Equal(t, "normal", cfg.MotionPreference)
	assert.False(t, cfg.GitExportEnabled)
	assert.Empty(t, cfg.GitExportRepo)
}

func TestLoad_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "from-zero-to-hero", cfg.Roadmap)
	assert.False(t, cfg.OnboardingComplete)
	assert.Zero(t, cfg.OnboardingVersion)
	assert.Equal(t, "rpg-skill-tree", cfg.Theme)
	assert.Equal(t, "rich", cfg.SymbolMode)
	assert.Equal(t, "normal", cfg.MotionPreference)
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &Config{
		DisplayName:      "Ada",
		Workspace:        filepath.Join(tmp, "my-workspace"),
		Editor:           "nvim",
		Language:         "python",
		Roadmap:          "interview-sprint",
		Theme:            "clean-productivity",
		SymbolMode:       "plain",
		MotionPreference: "reduced",
		GitExportEnabled: false,
	}
	require.NoError(t, cfg.Save())

	data, err := os.ReadFile(filepath.Join(tmp, ".leetgo", "config.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "Ada")
	assert.Contains(t, string(data), "clean-productivity")

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "Ada", loaded.DisplayName)
	assert.Equal(t, "nvim", loaded.Editor)
	assert.Equal(t, "python", loaded.Language)
	assert.Equal(t, "interview-sprint", loaded.Roadmap)
	assert.Equal(t, "clean-productivity", loaded.Theme)
	assert.Equal(t, "plain", loaded.SymbolMode)
	assert.Equal(t, "reduced", loaded.MotionPreference)
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Workspace: "/tmp/leetgo",
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go", "rust"}, []string{"from-zero-to-hero"})
	require.NoError(t, err)
}

func TestValidate_UnsupportedLanguage(t *testing.T) {
	cfg := &Config{
		Workspace: "/tmp/leetgo",
		Language:  "ruby",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "unsupported language")
}

func TestValidate_UnknownRoadmap(t *testing.T) {
	cfg := &Config{
		Workspace: "/tmp/leetgo",
		Language:  "go",
		Roadmap:   "unknown",
		Theme:     "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "unknown roadmap")
}

func TestValidate_UnknownTheme(t *testing.T) {
	cfg := &Config{
		Workspace: "/tmp/leetgo",
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "neon-hacker",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "unknown theme")
}

func TestValidate_UnknownSymbolMode(t *testing.T) {
	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "neon",
		MotionPreference: "normal",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "unknown symbol_mode")
}

func TestValidate_UnknownMotionPreference(t *testing.T) {
	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "rich",
		MotionPreference: "chaotic",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "unknown motion_preference")
}

func TestValidate_OnboardingCompleteWithoutDisplayName(t *testing.T) {
	cfg := &Config{
		OnboardingComplete: true,
		DisplayName:        "",
		Workspace:          "/tmp/leetgo",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "display_name is required")
}

func TestValidate_OnboardingCompleteWithDisplayName(t *testing.T) {
	cfg := &Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          "/tmp/leetgo",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	require.NoError(t, err)
}

func TestValidate_OnboardingCompleteDisplayNameTooLong(t *testing.T) {
	cfg := &Config{
		OnboardingComplete: true,
		DisplayName:        "This name is way too long and exceeds forty characters easily",
		Workspace:          "/tmp/leetgo",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "display_name exceeds")
}

func TestValidate_OnboardingCompleteDisplayNameAtMaxLength(t *testing.T) {
	name := ""
	for i := 0; i < MaxDisplayNameLength; i++ {
		name += "x"
	}
	cfg := &Config{
		OnboardingComplete: true,
		DisplayName:        name,
		Workspace:          "/tmp/leetgo",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	require.NoError(t, err)
}

func TestValidate_GitExportEnabledWithoutRepo(t *testing.T) {
	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		GitExportEnabled: true,
		GitExportRepo:    "",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "git_export_repo is required")
}

func TestValidate_GitExportEnabledWithInvalidRepo(t *testing.T) {
	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		GitExportEnabled: true,
		GitExportRepo:    "/nonexistent/path",
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "git_export_repo is not accessible")
}

func TestValidate_GitExportEnabledWithNonGitDir(t *testing.T) {
	tmp := t.TempDir()

	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		GitExportEnabled: true,
		GitExportRepo:    tmp,
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	assert.ErrorContains(t, err, "not a git repository")
}

func TestValidate_GitExportEnabledWithValidGitRepo(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, ".git"), 0o755))

	cfg := &Config{
		Workspace:        "/tmp/leetgo",
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		GitExportEnabled: true,
		GitExportRepo:    tmp,
	}

	err := cfg.Validate([]string{"go"}, []string{"from-zero-to-hero"})
	require.NoError(t, err)
}

func TestValidThemes(t *testing.T) {
	assert.Contains(t, ValidThemes, "rpg-skill-tree")
	assert.Contains(t, ValidThemes, "clean-productivity")
	assert.Contains(t, ValidThemes, "cyber-dashboard")
	assert.Len(t, ValidThemes, 3)
}

func TestValidAppearancePreferences(t *testing.T) {
	assert.ElementsMatch(t, []string{"rich", "plain"}, ValidSymbolModes)
	assert.ElementsMatch(t, []string{"normal", "reduced", "off"}, ValidMotionPreferences)
}

func TestLoad_BackwardCompatibility(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dataDir := filepath.Join(tmp, ".leetgo")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	oldConfig := `workspace = "/home/ada/leetgo-workspace"
editor = "nvim"
language = "go"
roadmap = "from-zero-to-hero"
`
	require.NoError(t, os.WriteFile(filepath.Join(dataDir, "config.toml"), []byte(oldConfig), 0o644))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "/home/ada/leetgo-workspace", cfg.Workspace)
	assert.Equal(t, "nvim", cfg.Editor)
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "from-zero-to-hero", cfg.Roadmap)
	assert.False(t, cfg.OnboardingComplete)
	assert.Zero(t, cfg.OnboardingVersion)
	assert.Empty(t, cfg.DisplayName)
	assert.Equal(t, "rpg-skill-tree", cfg.Theme)
	assert.Equal(t, "rich", cfg.SymbolMode)
	assert.Equal(t, "normal", cfg.MotionPreference)
	assert.False(t, cfg.GitExportEnabled)
	assert.Empty(t, cfg.GitExportRepo)
}

func TestReadyForDashboard(t *testing.T) {
	cfg := &Config{
		OnboardingComplete: true,
		OnboardingVersion:  CurrentOnboardingVersion,
		DisplayName:        "Grace",
		Workspace:          "/tmp/leetgo",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	assert.True(t, cfg.ReadyForDashboard([]string{"go"}, []string{"from-zero-to-hero"}))

	cfg.OnboardingVersion = 0
	assert.False(t, cfg.ReadyForDashboard([]string{"go"}, []string{"from-zero-to-hero"}), "stale configs must rerun Onboarding")

	cfg.OnboardingVersion = CurrentOnboardingVersion
	cfg.DisplayName = ""
	assert.False(t, cfg.ReadyForDashboard([]string{"go"}, []string{"from-zero-to-hero"}), "incomplete Profile must rerun Onboarding")
}
