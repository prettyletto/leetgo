package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRootModel(t *testing.T, cfg *config.Config) *RootModel {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	if cfg == nil {
		cfg = &config.Config{
			Workspace: t.TempDir(),
			Language:  "go",
			Roadmap:   "from-zero-to-hero",
			Theme:     "rpg-skill-tree",
		}
	}
	if cfg.OnboardingComplete && cfg.OnboardingVersion == 0 {
		cfg.OnboardingVersion = config.CurrentOnboardingVersion
	}

	legacyModel, err := NewModel(cfg, db)
	require.NoError(t, err)

	roadmaps, err := catalog.ListRoadmaps()
	require.NoError(t, err)

	languages := []string{"go", "python", "typescript", "java", "cpp", "javascript", "rust", "csharp"}

	root, err := NewRootModel(cfg, legacyModel, db, languages, roadmaps)
	require.NoError(t, err)
	return root
}

func TestNewRootModel_OnboardingIncomplete(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: false,
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	assert.NotNil(t, m.screen)
	_, ok := m.screen.(*OnboardingScreen)
	assert.True(t, ok, "expected OnboardingScreen when onboarding is incomplete")
}

func TestNewRootModel_OnboardingComplete(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	assert.NotNil(t, m.screen)
	_, ok := m.screen.(*DashboardScreen)
	assert.True(t, ok, "expected DashboardScreen when onboarding is complete")
}

func TestNewRootModel_StaleOnboardingVersionOpensOnboarding(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		OnboardingVersion:  0,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModelWithoutVersionUpgrade(t, cfg)
	assert.NotNil(t, m.screen)
	_, ok := m.screen.(*OnboardingScreen)
	assert.True(t, ok, "stale configs should rerun Onboarding")
}

func newTestRootModelWithoutVersionUpgrade(t *testing.T, cfg *config.Config) *RootModel {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	legacyModel, err := NewModel(cfg, db)
	require.NoError(t, err)

	roadmaps, err := catalog.ListRoadmaps()
	require.NoError(t, err)

	languages := []string{"go", "python", "typescript", "java", "cpp", "javascript", "rust", "csharp"}
	root, err := NewRootModel(cfg, legacyModel, db, languages, roadmaps)
	require.NoError(t, err)
	return root
}

func TestRootModel_View(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: false,
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	view := m.View()
	assert.Contains(t, view, "Who are you, challenger?")
}

func TestRootModel_NavigationToDashboard(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenDashboard})
	root := updated.(*RootModel)
	view := root.View()
	assert.Contains(t, view, "Welcome, Ada")
}

func TestRootModel_NavigationToLegacyList(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenLegacyList})
	root := updated.(*RootModel)
	view := root.View()
	assert.Contains(t, view, "Leetgo")
}

func TestRootModel_NavigationToOnboarding(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenOnboarding})
	root := updated.(*RootModel)
	view := root.View()
	assert.Contains(t, view, "Who are you, challenger?")
	_, ok := root.screen.(*OnboardingScreen)
	assert.True(t, ok, "expected OnboardingScreen after navigation")
}

func TestRootModel_WindowSizeMsg(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	root := updated.(*RootModel)
	assert.Equal(t, 120, root.width)
	assert.Equal(t, 40, root.height)
	view := root.View()
	assert.Contains(t, view, "Welcome, Ada")
}

func TestRootModel_UnsupportedSize(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 59, Height: 18})
	root := updated.(*RootModel)
	view := root.View()
	assert.Contains(t, view, "Unsupported Size")
	assert.Contains(t, view, "59x18")
}

func TestRootModel_UnsupportedSizeRestoresScreen(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	root := updated.(*RootModel)
	assert.Contains(t, root.View(), "Unsupported Size")
	updated, _ = root.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	root = updated.(*RootModel)
	assert.Contains(t, root.View(), "Welcome, Ada")
}

func TestRootModel_ReducedMotionDisablesAmbientBorder(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "cyber-dashboard",
		MotionPreference:   "reduced",
	}

	m := newTestRootModel(t, cfg)
	view := m.View()
	assert.Contains(t, view, "Welcome, Ada")
	assert.NotContains(t, view, "░░░░")
	assert.NotContains(t, view, "▓▓▓▓")
}

func TestRootModel_CyberNormalMotionShowsAmbientBorder(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "cyber-dashboard",
		MotionPreference:   "normal",
	}

	m := newTestRootModel(t, cfg)
	view := m.View()
	assert.Contains(t, view, "Welcome, Ada")
	assert.Contains(t, view, "░░░░")
}

func TestRootModel_Quit(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	assert.NotNil(t, cmd)
}

func TestRootModel_ConfigOwnership(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "cyber-dashboard",
	}

	m := newTestRootModel(t, cfg)
	assert.Equal(t, cfg, m.Config())
	assert.Equal(t, "cyber-dashboard", m.Theme())
	assert.Equal(t, "Cyber Dashboard", m.ThemeTokens().Name)
}

func TestNewRootModel_InvalidTheme(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "neon-hacker",
	}
	legacyModel, err := NewModel(cfg, db)
	require.NoError(t, err)
	roadmaps, err := catalog.ListRoadmaps()
	require.NoError(t, err)

	_, err = NewRootModel(cfg, legacyModel, db, []string{"go"}, roadmaps)
	assert.ErrorContains(t, err, "unknown theme")
}

func TestRootModel_ConfigOwnership_DefaultTheme(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: false,
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	assert.Equal(t, "rpg-skill-tree", m.Theme())
	assert.False(t, m.Config().OnboardingComplete)
}

func TestRootModel_GlobalNotification(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	m.Notify("test notification")
	view := m.View()
	assert.Contains(t, view, "test notification")
}

func TestRootModel_GlobalNotificationViaMsg(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(GlobalNotificationMsg{Message: "from screen"})
	root := updated.(*RootModel)
	view := root.View()
	assert.Contains(t, view, "from screen")
}

func TestRootModel_NotificationManagerIsOwned(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	assert.NotNil(t, m.notifications)
	m.Notify("owned by root")
	view := m.View()
	assert.Contains(t, view, "owned by root")
}

func TestRootModel_MultipleNotifications(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	m.Notify("first")
	m.Notify("second")
	view := m.View()
	assert.Contains(t, view, "second")
	assert.NotContains(t, view, "first")
}

func TestRootModel_ScreenDelegation_Init(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	cmd := m.Init()
	assert.Nil(t, cmd, "dashboard placeholder screen Init returns nil")

	cfg2 := &config.Config{
		OnboardingComplete: false,
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}
	m2 := newTestRootModel(t, cfg2)
	cmd2 := m2.Init()
	assert.Nil(t, cmd2, "onboarding screen Init returns nil")
}

func TestRootModel_ScreenDelegation_LegacyListStartStop(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenLegacyList})
	root := updated.(*RootModel)
	_, isLegacy := root.screen.(*LegacyProblemListScreen)
	assert.True(t, isLegacy, "navigation to legacy list should set LegacyProblemListScreen")
	assert.Contains(t, root.View(), "Leetgo")
}

func TestRootModel_NavigateToUnknownScreen(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	screenBefore := m.screen
	updated, _ := m.Update(NavigateMsg{ScreenID: "nonexistent"})
	root := updated.(*RootModel)
	assert.Equal(t, screenBefore, root.screen, "navigating to unknown screen should not change the active screen")
}

func TestRootModel_ThemeCycle(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, cmd := m.Update(ThemeChangedMsg{ThemeID: "clean-productivity"})
	require.NotNil(t, updated)
	assert.Nil(t, cmd)
	root := updated.(*RootModel)
	assert.Equal(t, "clean-productivity", root.cfg.Theme)
	assert.Equal(t, "clean-productivity", root.theme.ID)

	_, isDashboard := root.screen.(*DashboardScreen)
	assert.True(t, isDashboard, "theme change should recreate DashboardScreen")
}

func TestRootModel_ThemeCyclePreservesScreenSize(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 42})
	root := updated.(*RootModel)
	updated, _ = root.Update(ThemeChangedMsg{ThemeID: "clean-productivity"})
	root = updated.(*RootModel)

	dashboard, ok := root.screen.(*DashboardScreen)
	require.True(t, ok)
	assert.Equal(t, 140, dashboard.width)
	assert.Equal(t, 42, dashboard.height)
	assert.Contains(t, dashboard.View(), "Solved:", "wide Dashboard layout should remain after recreation")
}

func TestRootModel_ThemeCycleRecreatesSolveLogWithSize(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 132, Height: 37})
	root := updated.(*RootModel)
	updated, _ = root.Update(NavigateMsg{ScreenID: ScreenSolveLog})
	root = updated.(*RootModel)

	updated, _ = root.Update(ThemeChangedMsg{ThemeID: "clean-productivity"})
	root = updated.(*RootModel)

	solveLog, ok := root.screen.(*SolveLogScreen)
	require.True(t, ok)
	assert.Equal(t, "clean-productivity", solveLog.theme.ID)
	assert.Equal(t, 132, solveLog.width)
	assert.Equal(t, 37, solveLog.height)
}

func TestRootModel_NavigationToStageDetail(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenStageDetail, Stage: "arrays-hashing"})
	root := updated.(*RootModel)
	_, isStageDetail := root.screen.(*StageDetailScreen)
	assert.True(t, isStageDetail, "should navigate to StageDetailScreen")
	view := root.View()
	assert.Contains(t, view, "Arrays & Hashing")
}

func TestRootModel_ReloadsActiveRoadmap(t *testing.T) {
	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "hard-mode",
		Theme:              "rpg-skill-tree",
	}

	m := newTestRootModel(t, cfg)
	updated, _ := m.Update(NavigateMsg{ScreenID: ScreenDashboard})
	root := updated.(*RootModel)
	assert.Equal(t, "hard-mode", root.activeRoadmap.ID, "root should reload active roadmap from config")
	assert.Equal(t, "hard-mode", root.cfg.Roadmap)
}
