package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/workspace"
)

type ProblemDetailScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap

	problem *roadmap.Problem
	status  roadmap.Status

	leetcode  *leetcode.Client
	workspace *workspace.Manager
	solveLogs []*store.SolveLogRecord

	errorMsg     string
	submitting   bool
	spinnerFrame int

	width  int
	height int
}

type spinnerTickMsg time.Time

func NewProblemDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap, problemID int) *ProblemDetailScreen {
	p, ok := rm.Graph.Problems[problemID]
	if !ok {
		p = &roadmap.Problem{ID: problemID, Title: "Unknown Problem"}
	}

	s := &ProblemDetailScreen{
		cfg:     cfg,
		theme:   theme,
		db:      db,
		roadmap: rm,
		problem: p,
	}

	s.refresh()

	return s
}

func (s *ProblemDetailScreen) refresh() {
	ctx := context.Background()

	progress, err := s.db.GetProgress(ctx, s.problem.ID)
	if err == nil && progress != nil {
		s.status = progress.Status
	} else {
		s.effectiveStatus()
	}

	s.workspace = workspace.New(s.cfg.Workspace, generator.New())

	lcClient, err := leetcode.NewClient()
	if err == nil {
		s.leetcode = lcClient
	}

	logs, err := s.db.GetSolveLogs(ctx)
	if err == nil {
		var filtered []*store.SolveLogRecord
		for _, log := range logs {
			if log.ProblemID == s.problem.ID {
				filtered = append(filtered, log)
			}
		}
		if len(filtered) > 5 {
			filtered = filtered[:5]
		}
		s.solveLogs = filtered
	}
}

func (s *ProblemDetailScreen) effectiveStatus() {
	progress, err := s.db.GetAllProgress(context.Background())
	if err != nil {
		progress = make(map[int]roadmap.Status)
	}

	if status, ok := progress[s.problem.ID]; ok && status != "" {
		s.status = status
		return
	}

	for _, prereq := range s.problem.Prerequisites {
		if progress[prereq] != roadmap.StatusSolved {
			s.status = roadmap.StatusLocked
			return
		}
	}
	s.status = roadmap.StatusAvailable
}

func (s *ProblemDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *ProblemDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateMsg:
		return s, nil

	case submitResultMsg:
		return s.handleSubmitResult(msg)

	case spinnerTickMsg:
		if s.submitting {
			s.spinnerFrame = (s.spinnerFrame + 1) % len(spinnerFrames)
			return s, spinnerTickCmd()
		}
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "esc", "backspace":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenStageDetail, Stage: s.problemStage()}
			}

		case "enter":
			return s.handlePrimaryAction()

		case "o":
			return s.handleOpenEditor()

		case "x":
			return s.handleRunTests()

		case "s":
			return s.handleSubmit()

		case "m":
			return s.handleMarkSolved()

		case "t":
			currentIdx := sliceIndex(config.ValidThemes, s.cfg.Theme)
			nextIdx := (currentIdx + 1) % len(config.ValidThemes)
			return s, func() tea.Msg {
				return ThemeChangedMsg{ThemeID: config.ValidThemes[nextIdx]}
			}
		}
	}
	return s, nil
}

func (s *ProblemDetailScreen) handlePrimaryAction() (Screen, tea.Cmd) {
	switch s.status {
	case roadmap.StatusLocked:
		s.errorMsg = "Problem is locked — solve prerequisites first."
		return s, nil
	case roadmap.StatusAvailable:
		return s.handleStart()
	default:
		return s.handleOpenEditor()
	}
}

func (s *ProblemDetailScreen) handleStart() (Screen, tea.Cmd) {
	ctx := context.Background()

	if err := s.db.SetProgress(ctx, s.problem.ID, roadmap.StatusInProgress); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to update progress: %v", err)
		return s, nil
	}

	stubPath, testPath, err := s.workspace.Generate(s.problem, generator.Language(s.cfg.Language))
	if err != nil {
		s.errorMsg = fmt.Sprintf("Failed to generate files: %v", err)
		return s, nil
	}

	s.status = roadmap.StatusInProgress
	s.errorMsg = fmt.Sprintf("Started — files generated at %s", s.problemDir())

	return s, tea.Batch(
		func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Started %s — files generated", s.problem.Title)}
		},
		s.openEditorCmd(stubPath, testPath),
	)
}

func (s *ProblemDetailScreen) handleOpenEditor() (Screen, tea.Cmd) {
	stubPath, testPath := s.stubPaths()
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}

	cmd := s.openEditorCmd(stubPath, testPath)
	s.errorMsg = ""
	return s, cmd
}

func (s *ProblemDetailScreen) openEditorCmd(stubPath, testPath string) tea.Cmd {
	editor := s.cfg.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return func() tea.Msg {
			return GlobalNotificationMsg{Message: "No editor configured. Set $EDITOR or run `leetgo init`."}
		}
	}

	parts := strings.Fields(editor)
	args := append(parts[1:], stubPath, testPath)

	return func() tea.Msg {
		cmd := exec.Command(parts[0], args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to open editor: %v", err)}
		}
		return nil
	}
}

func (s *ProblemDetailScreen) handleRunTests() (Screen, tea.Cmd) {
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}

	dir := s.problemDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		s.errorMsg = "Problem files not generated. Start the problem first."
		return s, nil
	}

	cmd := s.testCommand()
	return s, func() tea.Msg {
		output, err := cmd.CombinedOutput()
		result := strings.TrimSpace(string(output))
		if err != nil {
			result += fmt.Sprintf("\nTestSuite error: %v", err)
		}
		return GlobalNotificationMsg{Message: result}
	}
}

func (s *ProblemDetailScreen) handleSubmit() (Screen, tea.Cmd) {
	if s.status != roadmap.StatusInProgress && s.status != roadmap.StatusSolved {
		s.errorMsg = "Start the problem first."
		return s, nil
	}

	if s.leetcode == nil || !s.leetcode.IsAuthenticated() {
		s.errorMsg = "Not authenticated. Run `leetgo auth` first."
		return s, nil
	}

	stubPath, _ := s.stubPaths()
	code, err := os.ReadFile(stubPath)
	if err != nil {
		s.errorMsg = fmt.Sprintf("Could not read solution file: %v", err)
		return s, nil
	}

	s.submitting = true
	s.errorMsg = "Submitting to LeetCode..."

	lang := leetcodeLang(s.cfg.Language)
	slug := s.problem.Slug
	client := s.leetcode

	return s, tea.Batch(
		spinnerTickCmd(),
		func() tea.Msg {
			result, err := client.Submit(context.Background(), slug, lang, string(code))
			return submitResultMsg{problemID: s.problem.ID, slug: slug, language: lang, result: result, err: err}
		},
	)
}

func (s *ProblemDetailScreen) handleSubmitResult(msg submitResultMsg) (Screen, tea.Cmd) {
	s.submitting = false

	if msg.err != nil {
		s.errorMsg = fmt.Sprintf("Submit failed: %v", msg.err)
		return s, nil
	}

	s.recordSolveLog(msg)

	if msg.result.StatusCode == 10 {
		s.markAcceptedSubmission(msg.problemID)
		s.errorMsg = fmt.Sprintf("Accepted! Runtime: %s, Memory: %s", msg.result.Runtime, msg.result.Memory)
	} else {
		s.errorMsg = fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests)
	}

	return s, nil
}

func (s *ProblemDetailScreen) recordSolveLog(msg submitResultMsg) {
	if msg.result == nil {
		return
	}
	log := &store.SolveLogRecord{
		ProblemID:   msg.problemID,
		Slug:        msg.slug,
		Language:    msg.language,
		Status:      msg.result.Status,
		StatusCode:  msg.result.StatusCode,
		Runtime:     msg.result.Runtime,
		Memory:      msg.result.Memory,
		TotalTests:  msg.result.TotalTests,
		PassedTests: msg.result.PassedTests,
		Error:       msg.result.Error,
		SubmittedAt: time.Now(),
	}
	if err := s.db.RecordSolveLog(context.Background(), log); err != nil {
		s.errorMsg += fmt.Sprintf(" | Failed to record solve log: %v", err)
	}
}

func (s *ProblemDetailScreen) markAcceptedSubmission(problemID int) {
	ctx := context.Background()
	progress, err := s.db.GetProgress(ctx, problemID)
	if err != nil {
		return
	}
	if progress != nil && progress.Status == roadmap.StatusSolved {
		return
	}
	if err := s.db.SetProgress(ctx, problemID, roadmap.StatusSolved); err != nil {
		return
	}

	xp := store.XPForDifficulty(s.problem.Difficulty)
	if err := s.db.AddXP(ctx, xp); err != nil {
		return
	}
	if err := s.db.UpdateStreak(ctx); err != nil {
		return
	}
	s.status = roadmap.StatusSolved
}

func (s *ProblemDetailScreen) handleMarkSolved() (Screen, tea.Cmd) {
	if s.status == roadmap.StatusSolved {
		s.errorMsg = "Already solved!"
		return s, nil
	}
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}

	ctx := context.Background()

	if err := s.db.SetProgress(ctx, s.problem.ID, roadmap.StatusSolved); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to update progress: %v", err)
		return s, nil
	}

	xp := store.XPForDifficulty(s.problem.Difficulty)
	if err := s.db.AddXP(ctx, xp); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to add XP: %v", err)
		return s, nil
	}

	if err := s.db.UpdateStreak(ctx); err != nil {
		s.errorMsg += fmt.Sprintf(" | Failed to update streak: %v", err)
	} else {
		_ = "streak updated"
	}

	s.status = roadmap.StatusSolved
	s.errorMsg = fmt.Sprintf("+%d XP for %s", xp, s.problem.Title)

	return s, nil
}

func (s *ProblemDetailScreen) stubPaths() (string, string) {
	stubName := strings.ReplaceAll(s.problem.Slug, "-", "_")
	stubExt := stubExt(s.cfg.Language)
	dir := s.problemDir()
	return filepath.Join(dir, stubName+stubExt), filepath.Join(dir, stubName+"_test"+stubExt)
}

func (s *ProblemDetailScreen) problemDir() string {
	return s.workspace.ProblemDir(s.problem)
}

func (s *ProblemDetailScreen) testCommand() *exec.Cmd {
	dir := s.problemDir()
	var cmd *exec.Cmd
	switch s.cfg.Language {
	case "go":
		cmd = exec.Command("go", "test", ".")
	case "python":
		cmd = exec.Command("python", "-m", "pytest")
	case "typescript":
		cmd = exec.Command("npm", "test")
	case "javascript":
		cmd = exec.Command("npm", "test")
	case "java":
		cmd = exec.Command("mvn", "test")
	case "cpp":
		cmd = exec.Command("sh", "-c", "g++ -std=c++17 *.cpp -o /tmp/leetgo-cpp-test && /tmp/leetgo-cpp-test")
	case "rust":
		cmd = exec.Command("sh", "-c", "rustc --test *_test.rs -o /tmp/leetgo-rust-test && /tmp/leetgo-rust-test")
	case "csharp":
		cmd = exec.Command("dotnet", "test")
	default:
		cmd = exec.Command("go", "test", ".")
	}
	cmd.Dir = dir
	return cmd
}

func (s *ProblemDetailScreen) View() string {
	var lines []string

	lines = append(lines, s.theme.Title.Render(fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title)))

	statusLabel := s.renderStatus()
	lines = append(lines, statusLabel)

	lines = append(lines, fmt.Sprintf("Difficulty: %s  Category: %s  Stage: %s",
		s.problem.Difficulty,
		s.problem.Category,
		s.stageName(),
	))

	if len(s.problem.Prerequisites) == 0 {
		lines = append(lines, "Prerequisites: none")
	} else {
		prereqs := s.prerequisiteLabels()
		lines = append(lines, "Prerequisites: "+strings.Join(prereqs, ", "))
	}

	if s.status == roadmap.StatusLocked {
		blocked := s.missingPrerequisites()
		if len(blocked) > 0 {
			lines = append(lines, lipgloss.NewStyle().
				Foreground(s.theme.Danger).
				Render("Blocked by: "+strings.Join(blocked, ", ")))
		}
	}

	if s.status != roadmap.StatusLocked {
		stubPath, testPath := s.stubPaths()
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Stub: %s", stubPath))
		lines = append(lines, fmt.Sprintf("Test: %s", testPath))
	}

	if len(s.solveLogs) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.theme.Key.Render("Latest Solve Logs"))
		for _, log := range s.solveLogs {
			when := log.SubmittedAt.Format("2006-01-02 15:04")
			result := log.Status
			if log.StatusCode == 10 {
				result = fmt.Sprintf("Accepted · %s · %s", log.Runtime, log.Memory)
			}
			lines = append(lines, fmt.Sprintf("  %s  %s  %s", when, log.Language, result))
		}
	}

	lines = append(lines, "")

	if s.submitting {
		frame := spinnerFrames[s.spinnerFrame%len(spinnerFrames)]
		lines = append(lines, s.theme.Spinner.Render(frame+" Submitting to LeetCode..."))
	}

	if s.errorMsg != "" {
		lines = append(lines, "")
		style := lipgloss.NewStyle().Foreground(s.theme.Muted)
		if strings.Contains(s.errorMsg, "failed") || strings.Contains(s.errorMsg, "locked") || strings.Contains(s.errorMsg, "Failed") {
			style = lipgloss.NewStyle().Foreground(s.theme.Danger)
		} else if strings.Contains(s.errorMsg, "Accepted") || strings.Contains(s.errorMsg, "XP") {
			style = lipgloss.NewStyle().Foreground(s.theme.Success)
		} else if strings.Contains(s.errorMsg, "Start") || strings.Contains(s.errorMsg, "Started") {
			style = lipgloss.NewStyle().Foreground(s.theme.Warning)
		}
		lines = append(lines, style.Render(s.errorMsg))
	}

	footer := s.renderFooter()
	return strings.Join(lines, "\n") + "\n" + footer
}

func (s *ProblemDetailScreen) renderStatus() string {
	switch s.status {
	case roadmap.StatusSolved:
		return lipgloss.NewStyle().Foreground(s.theme.Success).Bold(true).Render("[SOLVED]")
	case roadmap.StatusInProgress:
		return lipgloss.NewStyle().Foreground(s.theme.Warning).Bold(true).Render("[ACTIVE]")
	case roadmap.StatusAvailable:
		return lipgloss.NewStyle().Foreground(s.theme.PrimaryAccent).Bold(true).Render("[READY]")
	default:
		return lipgloss.NewStyle().Foreground(s.theme.Muted).Bold(true).Render("[LOCKED]")
	}
}

func (s *ProblemDetailScreen) problemStage() string {
	if s.problem.Stage != "" {
		return s.problem.Stage
	}
	return string(s.problem.Category)
}

func (s *ProblemDetailScreen) stageName() string {
	for _, stage := range s.roadmap.Stages {
		if stage.ID == s.problem.Stage || stage.ID == string(s.problem.Category) {
			return stage.Title
		}
	}
	if s.problem.Stage != "" {
		return s.problem.Stage
	}
	return string(s.problem.Category)
}

func (s *ProblemDetailScreen) prerequisiteLabels() []string {
	labels := make([]string, 0, len(s.problem.Prerequisites))
	for _, id := range s.problem.Prerequisites {
		if p, ok := s.roadmap.Graph.Problems[id]; ok {
			labels = append(labels, fmt.Sprintf("#%d %s", p.ID, p.Title))
		} else {
			labels = append(labels, fmt.Sprintf("#%d", id))
		}
	}
	return labels
}

func (s *ProblemDetailScreen) missingPrerequisites() []string {
	progress, _ := s.db.GetAllProgress(context.Background())
	var missing []string
	for _, id := range s.problem.Prerequisites {
		if progress[id] == roadmap.StatusSolved {
			continue
		}
		if prereq, ok := s.roadmap.Graph.Problems[id]; ok {
			missing = append(missing, fmt.Sprintf("#%d %s", prereq.ID, prereq.Title))
		} else {
			missing = append(missing, fmt.Sprintf("#%d", id))
		}
	}
	return missing
}

func (s *ProblemDetailScreen) renderFooter() string {
	primaryLabel := "start"
	switch s.status {
	case roadmap.StatusAvailable:
		primaryLabel = "start"
	case roadmap.StatusLocked:
		primaryLabel = "locked"
	default:
		primaryLabel = "open"
	}

	items := []string{
		s.theme.Key.Render("enter") + " " + primaryLabel,
		s.theme.Key.Render("o") + " open",
		s.theme.Key.Render("x") + " test",
		s.theme.Key.Render("s") + " submit",
		s.theme.Key.Render("m") + " solve",
		s.theme.Key.Render("t") + " theme",
		s.theme.Key.Render("esc") + " back",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg(t)
	})
}
