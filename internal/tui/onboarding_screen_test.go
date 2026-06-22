package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRoadmaps(t *testing.T) []*roadmap.Roadmap {
	t.Helper()
	roadmaps, err := catalog.ListRoadmaps()
	require.NoError(t, err)
	return roadmaps
}

func testLanguages() []string {
	return []string{"go", "python", "typescript", "java", "cpp", "javascript", "rust", "csharp"}
}

func TestOnboardingPrefill_DisplayName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "  Ada  "

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, "  Ada  ", s.displayNameInput, "raw input preserved; trimming happens on next")
}

func TestOnboardingPrefill_Workspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Workspace = "/home/ada/my-workspace"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, "/home/ada/my-workspace", s.workspaceInput)
}

func TestOnboardingPrefill_Language(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Language = "typescript"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	langs := testLanguages()
	assert.Equal(t, "typescript", langs[s.languageIndex])
}

func TestOnboardingPrefill_RoadmapFromConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Roadmap = "interview-sprint"

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	assert.Equal(t, "interview-sprint", rmaps[s.roadmapFocus].ID)
}

func TestOnboardingPrefill_RoadmapFallsBackToRecommended(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Roadmap = "nonexistent"

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	assert.Equal(t, "from-zero-to-hero", rmaps[s.roadmapFocus].ID)
	assert.True(t, rmaps[s.roadmapFocus].Recommended)
}

func TestOnboardingPrefill_Theme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Theme = "cyber-dashboard"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, "cyber-dashboard", config.ValidThemes[s.themeFocus])
}

func TestOnboardingPrefill_GitExportEnabled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	gitDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.GitExportEnabled = true
	cfg.GitExportRepo = gitDir

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, 0, s.gitExportChoice, "should be 'Yes' when GitExport is enabled")
	assert.Equal(t, gitDir, s.gitExportRepo, "should prefill the repo path")
}

func TestOnboardingPrefill_GitExportDisabled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.GitExportEnabled = false

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, 1, s.gitExportChoice, "should be 'Not now' when GitExport is disabled")
	assert.Empty(t, s.gitExportRepo)
}

func TestOnboardingPrefill_GitExportClearsRepoOnNotNow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.GitExportEnabled = false
	cfg.GitExportRepo = "/old/repo"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, 1, s.gitExportChoice)
	assert.Empty(t, s.gitExportRepo, "should clear repo when git export is disabled")
}

func TestOnboarding_EnterDisplayName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	assert.Equal(t, stepDisplayName, s.step)

	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	next, cmd := s.handleNext()
	assert.Equal(t, stepGitExport, s.step)
	assert.Equal(t, "Ada", s.cfg.DisplayName)
	assert.Nil(t, cmd)
	_, isOnboarding := next.(*OnboardingScreen)
	assert.True(t, isOnboarding)
}

func TestOnboarding_DisplayNameRequired(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.displayNameInput = "   "
	s.handleNext()
	assert.Equal(t, stepDisplayName, s.step, "should not advance when name is empty")
	assert.Contains(t, s.errorMsg, "required")
}

func TestOnboarding_GitExportNotNow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport
	s.gitExportChoice = 1
	s.handleNext()
	assert.Equal(t, stepWorkspaceLang, s.step)
	assert.False(t, s.cfg.GitExportEnabled)
	assert.Empty(t, s.cfg.GitExportRepo)
}

func TestOnboarding_GitExportOptIn(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	gitDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, ".git"), 0o755))

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport
	s.gitExportChoice = 0
	s.gitExportRepo = gitDir
	s.handleNext()
	assert.Equal(t, stepWorkspaceLang, s.step)
	assert.True(t, s.cfg.GitExportEnabled)
	assert.Equal(t, gitDir, s.cfg.GitExportRepo)
}

func TestOnboarding_GitExportRequiresRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport
	s.gitExportChoice = 0
	s.gitExportRepo = ""
	s.handleNext()
	assert.Equal(t, stepGitExport, s.step, "should not advance without repo path")
	assert.Contains(t, s.errorMsg, "required")
}

func TestOnboarding_GitExportNonExistentPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport
	s.gitExportChoice = 0
	s.gitExportRepo = "/nonexistent/path"
	s.handleNext()
	assert.Equal(t, stepGitExport, s.step)
	assert.Contains(t, s.errorMsg, "not exist")
}

func TestOnboarding_GitExportNotGitRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	nonGitDir := t.TempDir()
	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport
	s.gitExportChoice = 0
	s.gitExportRepo = nonGitDir
	s.handleNext()
	assert.Equal(t, stepGitExport, s.step)
	assert.Contains(t, s.errorMsg, "git repository")
}

func TestOnboarding_WorkspaceRequired(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepWorkspaceLang
	s.workspaceInput = ""
	s.handleNext()
	assert.Equal(t, stepWorkspaceLang, s.step)
	assert.Contains(t, s.errorMsg, "required")
}

func TestOnboarding_RoadmapCarouselWrapsLeft(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 0

	s.handleRoadmapKey(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, len(rmaps)-1, s.roadmapFocus, "left from first should wrap to last")
}

func TestOnboarding_RoadmapCarouselWrapsRight(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = len(rmaps) - 1

	s.handleRoadmapKey(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 0, s.roadmapFocus, "right from last should wrap to first")
}

func TestOnboarding_RoadmapCarouselHKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 1

	s.handleRoadmapKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	assert.Equal(t, 0, s.roadmapFocus, "h should move left")
}

func TestOnboarding_RoadmapCarouselLKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 0

	s.handleRoadmapKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	assert.Equal(t, 1, s.roadmapFocus, "l should move right")
}

func TestOnboarding_RoadmapCarouselRendersThreeCards(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	require.GreaterOrEqual(t, len(rmaps), 3, "need at least 3 bundled roadmaps")

	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 0

	view := s.View()
	assert.Contains(t, view, "Step 4/5")
	assert.Contains(t, view, "From Zero To Hero")
	assert.Contains(t, view, "Interview Sprint")
	assert.Contains(t, view, "Hard Mode")
}

func TestOnboarding_RoadmapCarouselFocusShowsAll(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	require.GreaterOrEqual(t, len(rmaps), 3)

	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 1

	view := s.View()
	assert.Contains(t, view, "From Zero To Hero", "preview card should show")
	assert.Contains(t, view, "Interview Sprint", "focused card should show")
	assert.Contains(t, view, "Hard Mode", "preview card should show")
}

func TestOnboarding_RoadmapConfirmSetsRoadmap(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 2
	s.handleNext()
	assert.Equal(t, stepThemeSelection, s.step)
	assert.Equal(t, "interview-sprint", s.cfg.Roadmap)
}

func TestOnboarding_CompletionSavesAndNavigates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "TestUser"
	cfg.Workspace = t.TempDir()
	cfg.Language = "go"
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	nextScreen, cmd := s.handleNext()
	assert.Nil(t, nextScreen, "screen should be nil on completion (root replaces it)")
	require.NotNil(t, cmd, "should return a navigation command")
	assert.True(t, s.cfg.OnboardingComplete, "should mark onboarding complete on success")
}

func TestOnboarding_CompletionValidationFails_NoWorkspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "TestUser"
	cfg.Workspace = ""
	cfg.Language = "go"
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	next, cmd := s.handleNext()
	assert.NotNil(t, next, "should stay on the screen")
	assert.Nil(t, cmd)
	assert.False(t, s.cfg.OnboardingComplete, "should not mark complete on validation failure")
	assert.Contains(t, s.errorMsg, "validation failed")
	assert.Equal(t, stepThemeSelection, s.step, "should stay on final step")
}

func TestOnboarding_CompletionValidationFails_UnsupportedLanguage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "TestUser"
	cfg.Workspace = t.TempDir()
	cfg.Language = "ruby"
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	next, cmd := s.handleNext()
	assert.NotNil(t, next)
	assert.Nil(t, cmd)
	assert.False(t, s.cfg.OnboardingComplete)
	assert.Contains(t, s.errorMsg, "validation failed")
	assert.Contains(t, s.errorMsg, "ruby")
}

func TestOnboarding_CompletionValidationFails_UnknownRoadmap(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "TestUser"
	cfg.Workspace = t.TempDir()
	cfg.Language = "go"
	cfg.Roadmap = "unknown-roadmap"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	next, cmd := s.handleNext()
	assert.NotNil(t, next)
	assert.Nil(t, cmd)
	assert.False(t, s.cfg.OnboardingComplete)
	assert.Contains(t, s.errorMsg, "validation failed")
}

func TestOnboarding_CompletionValidationFails_DisplayNameEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = ""
	cfg.Workspace = t.TempDir()
	cfg.Language = "go"
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	next, cmd := s.handleNext()
	assert.NotNil(t, next)
	assert.Nil(t, cmd)
	assert.False(t, s.cfg.OnboardingComplete)
	assert.Contains(t, s.errorMsg, "validation failed")
}

func TestOnboarding_QuitDoesNotMarkComplete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.OnboardingComplete = false

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
	assert.False(t, s.cfg.OnboardingComplete, "quitting should not mark Onboarding complete")
}

func TestOnboarding_EscGoesBack(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepGitExport

	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, stepDisplayName, s.step, "esc should go back one step")
}

func TestOnboarding_EscStaysOnFirstStep(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepDisplayName

	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, stepDisplayName, s.step, "esc on first step should stay put")
}

func TestOnboarding_FullFlowNoGitExport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.Workspace = t.TempDir()

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))

	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("T")})
	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	s.handleDisplayNameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	s.handleNext()

	s.handleNext()

	s.handleNext()

	s.handleNext()

	_, cmd := s.handleNext()
	assert.NotNil(t, cmd)
	assert.True(t, s.cfg.OnboardingComplete)
	assert.Equal(t, "Test", s.cfg.DisplayName)
	assert.False(t, s.cfg.GitExportEnabled)
	assert.Equal(t, "go", s.cfg.Language)
	assert.Equal(t, "from-zero-to-hero", s.cfg.Roadmap)
	assert.Equal(t, "rpg-skill-tree", s.cfg.Theme)
}

func TestOnboarding_CompletionReturnsNavigateCmd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.DisplayName = "Ada"
	cfg.Workspace = t.TempDir()
	cfg.Language = "go"
	cfg.Roadmap = "from-zero-to-hero"
	cfg.Theme = "rpg-skill-tree"

	s := NewOnboardingScreen(cfg, testLanguages(), testRoadmaps(t))
	s.step = stepThemeSelection
	s.themeFocus = 0

	_, cmd := s.handleNext()
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok, "should return NavigateMsg")
	assert.Equal(t, ScreenDashboard, navigate.ScreenID)
}

func TestOnboarding_RoadmapCarouselViewContainsIndex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	rmaps := testRoadmaps(t)
	s := NewOnboardingScreen(cfg, testLanguages(), rmaps)
	s.step = stepRoadmapCarousel
	s.roadmapFocus = 0

	view := s.View()
	assert.Contains(t, view, "1/"+itoa(len(rmaps)))

	s.roadmapFocus = len(rmaps) - 1
	view = s.View()
	assert.Contains(t, view, itoa(len(rmaps))+"/"+itoa(len(rmaps)))
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }
