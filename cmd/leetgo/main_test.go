package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = old
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return string(out)
}

func TestSetConfigValue_Language(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	require.NoError(t, setConfigValue(cfg, "language", "rust"))
	assert.Equal(t, "rust", cfg.Language)
}

func TestSetConfigValue_Roadmap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	require.NoError(t, setConfigValue(cfg, "roadmap", "hard-mode"))
	assert.Equal(t, "hard-mode", cfg.Roadmap)
}

func TestSetConfigValue_InvalidLanguage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := setConfigValue(cfg, "language", "ruby")
	assert.ErrorContains(t, err, "unsupported language")
}

func TestSetConfigValue_WorkspacePersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := filepath.Join(home, "workspace")
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	require.NoError(t, setConfigValue(cfg, "workspace", workspace))
	loaded, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, workspace, loaded.Workspace)
}

func TestShowSolveLog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.RecordSolveLog(context.Background(), &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: time.Now(),
	}))

	require.NoError(t, showSolveLog([]string{"1"}))
}

func TestFindProblem_ByIDAndSlug(t *testing.T) {
	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	byID, err := findProblem(rm.Graph, "1")
	require.NoError(t, err)
	assert.Equal(t, "two-sum", byID.Slug)

	bySlug, err := findProblem(rm.Graph, "two-sum")
	require.NoError(t, err)
	assert.Equal(t, 1, bySlug.ID)
}

func TestTestCommand(t *testing.T) {
	tests := []struct {
		language string
		name     string
	}{
		{"go", "go"},
		{"python", "python"},
		{"typescript", "npm"},
		{"javascript", "npm"},
		{"java", "mvn"},
		{"cpp", "sh"},
		{"rust", "sh"},
		{"csharp", "dotnet"},
	}
	for _, tt := range tests {
		cmd, err := testCommand(tt.language, "/tmp/problem")
		require.NoError(t, err)
		assert.Equal(t, tt.name, filepath.Base(cmd.Path))
		assert.Equal(t, "/tmp/problem", cmd.Dir)
	}
}

func TestTestCommand_UnsupportedLanguage(t *testing.T) {
	_, err := testCommand("ruby", "/tmp/problem")
	assert.ErrorContains(t, err, "unsupported language")
}

func TestPrintTestPassedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printTestPassed(problem, "+7 XP", "run `leetgo submit .` for LeetCode confirmation and bonus XP")
	})

	assert.Contains(t, out, "Leetgo TestSuite passed for #1 Two Sum")
	assert.Contains(t, out, "Status: Verified")
	assert.Contains(t, out, "Reward: +7 XP")
	assert.Contains(t, out, "Next: run `leetgo submit .` for LeetCode confirmation and bonus XP")
	assert.Contains(t, out, "Reward Moment")
	assert.Contains(t, out, "Problem Verified")
	assert.Contains(t, out, "Output: static non-TTY")
}

func TestPrintTestPassedAlreadyClaimedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printTestPassedWithStatus(problem, "Solved", "already claimed", "run `leetgo submit .`")
	})

	assert.Contains(t, out, "Leetgo TestSuite passed for #1 Two Sum")
	assert.Contains(t, out, "Status: Solved")
	assert.Contains(t, out, "Reward: already claimed")
	assert.Contains(t, out, "Next: run `leetgo submit .`")
}

func TestPrintTestFailedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printTestFailed(problem, "InProgress", "FAIL")
	})

	assert.Contains(t, out, "Leetgo TestSuite failed for #1 Two Sum")
	assert.Contains(t, out, "Status: InProgress")
	assert.Contains(t, out, "FAIL")
}

func TestPrintSubmitAcceptedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printSubmitAccepted(problem, "+3 XP", "1 ms", "2 MB")
	})

	assert.Contains(t, out, "LeetCode Accepted for #1 Two Sum")
	assert.Contains(t, out, "Status: Solved")
	assert.Contains(t, out, "Reward: +3 XP")
	assert.Contains(t, out, "Runtime: 1 ms")
	assert.Contains(t, out, "Memory: 2 MB")
	assert.Contains(t, out, "Reward Moment")
	assert.Contains(t, out, "Problem Solved")
}

func TestPrintSubmitAcceptedAlreadyClaimedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printSubmitAccepted(problem, "already claimed", "1 ms", "2 MB")
	})

	assert.Contains(t, out, "LeetCode Accepted for #1 Two Sum")
	assert.Contains(t, out, "Status: Solved")
	assert.Contains(t, out, "Reward: already claimed")
	assert.Contains(t, out, "Runtime: 1 ms")
	assert.Contains(t, out, "Memory: 2 MB")
}

func TestPrintSubmitUnavailableOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}

	out := captureStdout(t, func() {
		printSubmitUnavailable(problem, "Verified")
	})

	assert.Contains(t, out, "Submission unavailable for #1 Two Sum")
	assert.Contains(t, out, "Status: Verified")
	assert.Contains(t, out, "Reward: none")
	assert.Contains(t, out, "Error: Session expired. Run `leetgo auth` to reconnect.")
}

func TestPrintSubmitRejectedOutput(t *testing.T) {
	problem := &roadmap.Problem{ID: 1, Title: "Two Sum"}
	result := &leetcode.SubmissionResult{Status: "Wrong Answer", PassedTests: 50, TotalTests: 63}

	out := captureStdout(t, func() {
		printSubmitRejected(problem, "Verified", result)
	})

	assert.Contains(t, out, "LeetCode rejected #1 Two Sum")
	assert.Contains(t, out, "Status: Verified")
	assert.Contains(t, out, "Wrong Answer (50/63 tests passed)")
}

func TestCLIStatusLabel(t *testing.T) {
	assert.Equal(t, "Verified", cliStatusLabel(roadmap.StatusVerified))
	assert.Equal(t, "Solved", cliStatusLabel(roadmap.StatusSolved))
	assert.Equal(t, "InProgress", cliStatusLabel(roadmap.StatusInProgress))
}

func TestParseGitExportArgs(t *testing.T) {
	parsed, err := parseGitExportArgs([]string{"/tmp/repo", "--commit"})
	require.NoError(t, err)
	assert.Equal(t, "/tmp/repo", parsed.repoDir)
	assert.True(t, parsed.commit)
}

func TestParseGitExportArgs_UnknownOption(t *testing.T) {
	_, err := parseGitExportArgs([]string{"/tmp/repo", "--push"})
	assert.ErrorContains(t, err, "unknown git-export option")
}

func TestSetConfigValue_Theme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	require.NoError(t, setConfigValue(cfg, "theme", "cyber-dashboard"))
	assert.Equal(t, "adaptive", cfg.Theme)
}

func TestSetConfigValue_InvalidTheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := setConfigValue(cfg, "theme", "neon-hacker")
	assert.ErrorContains(t, err, "unknown theme")
}

func TestSetConfigValue_SymbolMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:        t.TempDir(),
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "rich",
		MotionPreference: "normal",
	}

	require.NoError(t, setConfigValue(cfg, "symbol-mode", "plain"))
	assert.Equal(t, "plain", cfg.SymbolMode)
}

func TestSetConfigValue_InvalidSymbolMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:        t.TempDir(),
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "rich",
		MotionPreference: "normal",
	}

	err := setConfigValue(cfg, "symbol-mode", "sparkles")
	assert.ErrorContains(t, err, "unknown symbol_mode")
}

func TestSetConfigValue_MotionPreference(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:        t.TempDir(),
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "rich",
		MotionPreference: "normal",
	}

	require.NoError(t, setConfigValue(cfg, "motion", "reduced"))
	assert.Equal(t, "reduced", cfg.MotionPreference)
}

func TestSetConfigValue_InvalidMotionPreference(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:        t.TempDir(),
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		SymbolMode:       "rich",
		MotionPreference: "normal",
	}

	err := setConfigValue(cfg, "motion", "maximum")
	assert.ErrorContains(t, err, "unknown motion_preference")
}

func TestSetConfigValue_DisplayName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	require.NoError(t, setConfigValue(cfg, "display-name", "Ada"))
	assert.Equal(t, "Ada", cfg.DisplayName)
}

func TestSetConfigValue_GitExportEnabled(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, ".git"), 0o755))

	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:     t.TempDir(),
		Language:      "go",
		Roadmap:       "from-zero-to-hero",
		Theme:         "rpg-skill-tree",
		GitExportRepo: tmp,
	}

	require.NoError(t, setConfigValue(cfg, "git-export-enabled", "true"))
	assert.True(t, cfg.GitExportEnabled)

	require.NoError(t, setConfigValue(cfg, "git-export-enabled", "false"))
	assert.False(t, cfg.GitExportEnabled)
}

func TestSetConfigValue_GitExportEnabledInvalid(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := setConfigValue(cfg, "git-export-enabled", "maybe")
	assert.ErrorContains(t, err, "must be true or false")
}

func TestSetConfigValue_GitExportRepo(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, ".git"), 0o755))

	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace:        t.TempDir(),
		Language:         "go",
		Roadmap:          "from-zero-to-hero",
		Theme:            "rpg-skill-tree",
		GitExportEnabled: true,
		GitExportRepo:    tmp,
	}

	require.NoError(t, setConfigValue(cfg, "git-export-repo", tmp))
	assert.Equal(t, tmp, cfg.GitExportRepo)
}

func TestRunOnboard_SetsIncomplete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.OnboardingComplete = true
	cfg.OnboardingVersion = config.CurrentOnboardingVersion
	cfg.DisplayName = "Ada"

	require.NoError(t, runOnboard(cfg, nil))

	loaded, err := config.Load()
	require.NoError(t, err)
	assert.False(t, loaded.OnboardingComplete, "onboarding should be set to incomplete")
	assert.Zero(t, loaded.OnboardingVersion, "rerun should clear onboarding version")
	assert.Equal(t, "Ada", loaded.DisplayName, "existing display_name should be preserved")
}

func TestRunOnboard_FreshClearsProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.DefaultConfig()
	require.NoError(t, err)
	cfg.OnboardingComplete = true
	cfg.OnboardingVersion = config.CurrentOnboardingVersion
	cfg.DisplayName = "Ada"
	cfg.Language = "python"
	cfg.Roadmap = "interview-sprint"
	cfg.Theme = "cyber-dashboard"
	cfg.SymbolMode = "plain"
	cfg.MotionPreference = "off"
	cfg.GitExportEnabled = true
	cfg.GitExportRepo = "/tmp/repo"

	require.NoError(t, runOnboard(cfg, []string{"--fresh"}))

	loaded, err := config.Load()
	require.NoError(t, err)
	assert.False(t, loaded.OnboardingComplete)
	assert.Zero(t, loaded.OnboardingVersion)
	assert.Empty(t, loaded.DisplayName, "fresh should clear display name")
	assert.Equal(t, "go", loaded.Language, "fresh should reset language")
	assert.Contains(t, loaded.Workspace, "leetgo-workspace", "fresh should use the standard workspace default")
	assert.Equal(t, "from-zero-to-hero", loaded.Roadmap, "fresh should reset roadmap")
	assert.Equal(t, "adaptive", loaded.Theme, "fresh should reset theme")
	assert.Equal(t, "rich", loaded.SymbolMode, "fresh should reset symbol mode")
	assert.Equal(t, "normal", loaded.MotionPreference, "fresh should reset motion preference")
	assert.False(t, loaded.GitExportEnabled, "fresh should disable git export")
	assert.Empty(t, loaded.GitExportRepo, "fresh should clear git export repo")
}

func TestRunOnboard_UnknownOption(t *testing.T) {
	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	err = runOnboard(cfg, []string{"--nope"})
	assert.ErrorContains(t, err, "unknown onboard option")
}

func TestDefaultConfig_FreshDefaults(t *testing.T) {
	cfg, err := config.DefaultConfig()
	require.NoError(t, err)

	assert.False(t, cfg.OnboardingComplete, "fresh config should not be complete")
	assert.Zero(t, cfg.OnboardingVersion, "fresh config should not have completed Onboarding version")
	assert.Empty(t, cfg.DisplayName, "fresh config should have empty display name")
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "from-zero-to-hero", cfg.Roadmap)
	assert.Equal(t, "adaptive", cfg.Theme)
	assert.False(t, cfg.GitExportEnabled)
}

func TestResolveProblem_ByID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	problem, language, dir, err := resolveProblem("1", cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, problem.ID)
	assert.Equal(t, "two-sum", problem.Slug)
	assert.Equal(t, "go", language)
	assert.Contains(t, dir, "1-two-sum")
}

func TestResolveProblem_BySlug(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "python",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	problem, language, _, err := resolveProblem("two-sum", cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, problem.ID)
	assert.Equal(t, "python", language)
}

func TestResolveProblem_Invalid(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	_, _, _, err := resolveProblem("nonexistent", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no manifest found")
}

func TestResolveProblem_ManifestInPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	problemDir := t.TempDir()
	require.NoError(t, os.MkdirAll(problemDir, 0o755))

	m := &workspace.Manifest{
		ProblemID:     1,
		Slug:          "two-sum",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "python",
		StubPath:      "two_sum.py",
		TestsuitePath: "two_sum_test.py",
	}
	require.NoError(t, workspace.WriteManifest(problemDir, m))

	problem, language, dir, err := resolveProblem(problemDir, cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, problem.ID)
	assert.Equal(t, "two-sum", problem.Slug)
	assert.Equal(t, "python", language)
	assert.Equal(t, problemDir, dir)
}

func TestResolveProblem_ManifestParentWalking(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	problemDir := t.TempDir()
	require.NoError(t, os.MkdirAll(problemDir, 0o755))

	m := &workspace.Manifest{
		ProblemID:     1,
		Slug:          "two-sum",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "go",
		StubPath:      "two_sum.go",
		TestsuitePath: "two_sum_test.go",
	}
	require.NoError(t, workspace.WriteManifest(problemDir, m))

	subdir := filepath.Join(problemDir, "src", "nested")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	problem, _, dir, err := resolveProblem(subdir, cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, problem.ID)
	assert.Equal(t, problemDir, dir)
}

func TestResolveProblem_DotResolvesCWD(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	problemDir := t.TempDir()
	m := &workspace.Manifest{
		ProblemID:     1,
		Slug:          "two-sum",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "go",
		StubPath:      "two_sum.go",
		TestsuitePath: "two_sum_test.go",
	}
	require.NoError(t, workspace.WriteManifest(problemDir, m))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(problemDir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	problem, language, dir, err := resolveProblem(".", cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, problem.ID)
	assert.Equal(t, "go", language)
	assert.Equal(t, problemDir, dir)
}

func TestRunProblemTests_AlreadyClaimedReturnsNil(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}))
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err = runProblemTests(cfg, []string{"1"})
	assert.Error(t, err, "should error because no problem files exist for problem 1")
}

func TestRunProblemSubmit_UnauthenticatedReturnsNil(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err = runProblemSubmit(cfg, []string{"1"})
	assert.Error(t, err, "should error because no problem files exist")
}

func TestRunProblemSubmit_NoFilesError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err = runProblemSubmit(cfg, []string{"1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunManualSolve_RequiresManualFlag(t *testing.T) {
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := runManualSolve(cfg, []string{"1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--manual")
}

func TestRunManualSolve_UnknownProblem(t *testing.T) {
	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := runManualSolve(cfg, []string{"--manual", "99999"})
	assert.Error(t, err)
}

func TestRunManualSolve_WithNote(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runManualSolve(cfg, []string{"--manual", "--yes", "--note", "browser solve", "1"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Manually Solved")
	assert.Contains(t, output, "browser solve")
	assert.Contains(t, output, "Manual Solve earns no XP")

	ctx := context.Background()
	db2, _ := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	defer db2.Close()

	sp, err := db2.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "manual", sp.Kind)
	assert.Equal(t, "browser solve", sp.Note)

	progress, err := db2.GetProgress(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, progress)
	assert.Equal(t, roadmap.StatusSolved, progress.Status)
}

func TestRunManualSolve_AlreadySolved(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	db.Close()

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runManualSolve(cfg, []string{"--manual", "--yes", "1"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Already Solved")
}

func TestRunSubmit_SkipTestsFlag(t *testing.T) {
	err := runProblemSubmit(&config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}, []string{"--skip-tests", "1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunSubmit_UnknownFlag(t *testing.T) {
	err := runProblemSubmit(&config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}, []string{"--unknown", "1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown submit option")
}

func TestRunNext_ShowsPrimaryAction(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runNext(cfg, []string{})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Next:")
}

func TestRunNext_All(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runNext(cfg, []string{"--all"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Next Actions:")
}

func TestRunNext_StartNotStartAction(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	db, err := store.NewSQLiteStore(filepath.Join(dataDir, "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runNext(cfg, []string{"--start"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "not Start")
}

func TestRunNext_UnknownFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	err := runNext(cfg, []string{"--unknown"})
	assert.Error(t, err)
}

func TestRunInfo_OutsideWorkspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runInfo(cfg, []string{"1"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "#1 Two Sum")
	assert.Contains(t, output, "Difficulty")
	assert.Contains(t, output, "Status")
}

func TestRunInfo_InsideWorkspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	problem := rm.Graph.Problems[1]
	require.NotNil(t, problem)

	manager := workspace.New(cfg.Workspace, generator.New())
	dir := manager.ProblemDir(problem)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, workspace.WriteManifest(dir, &workspace.Manifest{
		ProblemID:     1,
		Slug:          "two-sum",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "go",
		StubPath:      "two_sum.go",
		TestsuitePath: "two_sum_test.go",
	}))

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	require.NoError(t, os.Chdir(dir))

	output := captureStdout(t, func() {
		err := runInfo(cfg, []string{})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "#1 Two Sum")
	assert.Contains(t, output, "Practice Focus")
}

func TestRunInfo_LockedShowsBlockers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runInfo(cfg, []string{"49"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "#49 Group Anagrams")
	assert.Contains(t, output, "Locked")
	assert.Contains(t, output, "Blockers")
}

func TestRunInfo_SolvedShowsProvenance(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir, err := config.DataDir()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dataDir, 0o755))

	dbPath := filepath.Join(dataDir, "leetgo.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, db.RecordSolveProvenance(ctx, &store.SolveProvenance{ProblemID: 1, Kind: "manual", Note: "test note"}))

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
		Roadmap:   "from-zero-to-hero",
		Theme:     "rpg-skill-tree",
	}

	output := captureStdout(t, func() {
		err := runInfo(cfg, []string{"1"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Manual Solve")
	assert.Contains(t, output, "test note")
}
