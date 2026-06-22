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

	manualSolveMode  bool
	manualSolveInput string

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
		if progress[prereq] != roadmap.StatusSolved && progress[prereq] != roadmap.StatusVerified {
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

	case testRunResultMsg:
		return s.handleTestRunResult(msg)

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
		if s.manualSolveMode {
			return s.handleManualSolveKey(msg)
		}
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
	case roadmap.StatusVerified:
		return s.handleSubmit()
	default:
		return s.handleOpenEditor()
	}
}

func (s *ProblemDetailScreen) handleStart() (Screen, tea.Cmd) {
	ctx := context.Background()

	if err := workspace.EnsureManifestWritable(s.problemDir(), s.problem.ID); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to write manifest: %v", err)
		return s, nil
	}

	if err := s.db.SetProgress(ctx, s.problem.ID, roadmap.StatusInProgress); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to update progress: %v", err)
		return s, nil
	}

	stubPath, testPath, err := s.workspace.Generate(s.problem, generator.Language(s.cfg.Language))
	if err != nil {
		s.errorMsg = fmt.Sprintf("Failed to generate files: %v", err)
		return s, nil
	}

	if err := s.writeManifest(stubPath, testPath); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to write manifest: %v", err)
		return s, nil
	}

	s.status = roadmap.StatusInProgress
	s.errorMsg = "Started. Files opened in editor."

	return s, tea.Batch(
		func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Started %s — files generated", s.problem.Title)}
		},
		s.openEditorCmd(stubPath, testPath),
	)
}

func (s *ProblemDetailScreen) writeManifest(stubPath, testPath string) error {
	dir := s.problemDir()
	stage := s.problem.Stage
	if stage == "" {
		stage = string(s.problem.Category)
	}
	m := &workspace.Manifest{
		ProblemID:     s.problem.ID,
		Slug:          s.problem.Slug,
		Roadmap:       s.roadmap.ID,
		Stage:         stage,
		Language:      s.cfg.Language,
		StubPath:      filepath.Base(stubPath),
		TestsuitePath: filepath.Base(testPath),
	}
	return workspace.WriteManifest(dir, m)
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
	return func() tea.Msg {
		cmd := editorCommand(editor, []string{stubPath, testPath}, s.problemDir())
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

	difficulty := s.problem.Difficulty
	problemID := s.problem.ID
	cmd := s.testCommand()
	return s, func() tea.Msg {
		output, err := cmd.CombinedOutput()
		result := strings.TrimSpace(string(output))
		passed := err == nil
		if err != nil {
			result += fmt.Sprintf("\nTestSuite error: %v", err)
		}
		return testRunResultMsg{
			problemID:  problemID,
			difficulty: difficulty,
			output:     result,
			passed:     passed,
		}
	}
}

func (s *ProblemDetailScreen) handleSubmit() (Screen, tea.Cmd) {
	if s.status != roadmap.StatusInProgress && s.status != roadmap.StatusVerified && s.status != roadmap.StatusSolved {
		s.errorMsg = "Start the problem first."
		return s, nil
	}

	if s.leetcode == nil || !s.leetcode.IsAuthenticated() {
		s.errorMsg = "Session expired. Run `leetgo auth` and try again."
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
		if s.errorMsg == "" {
			s.errorMsg = fmt.Sprintf("Accepted! Runtime: %s, Memory: %s", msg.result.Runtime, msg.result.Memory)
		}
	} else {
		s.errorMsg = fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests)
	}

	return s, nil
}

func (s *ProblemDetailScreen) handleTestRunResult(msg testRunResultMsg) (Screen, tea.Cmd) {
	if !msg.passed {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: msg.output}
		}
	}

	if s.status == roadmap.StatusVerified || s.status == roadmap.StatusSolved {
		return s, func() tea.Msg {
			result := msg.output + "\n\nTests passed. Reward already claimed."
			return GlobalNotificationMsg{Message: result}
		}
	}

	ctx := context.Background()
	alreadyClaimed, err := s.db.HasRewardEvent(ctx, msg.problemID, "verify")
	if err != nil {
		s.errorMsg = fmt.Sprintf("Failed to check verify reward: %v", err)
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: msg.output}
		}
	}

	if alreadyClaimed {
		return s, func() tea.Msg {
			result := msg.output + "\n\nVerify XP already claimed for this problem."
			return GlobalNotificationMsg{Message: result}
		}
	}

	xp := store.XPForDifficulty(msg.difficulty) * 70 / 100
	if xp > 0 {
		if err := s.db.AddXP(ctx, xp); err != nil {
			s.errorMsg = fmt.Sprintf("Failed to add verify XP: %v", err)
			return s, func() tea.Msg {
				return GlobalNotificationMsg{Message: msg.output}
			}
		}
	}

	event := &store.RewardEvent{
		ProblemID: msg.problemID,
		Kind:      "verify",
		XP:        xp,
	}
	if err := s.db.RecordRewardEvent(ctx, event); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to record verify event: %v", err)
		return s, nil
	}

	if err := s.db.SetProgress(ctx, msg.problemID, roadmap.StatusVerified); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to update status: %v", err)
		return s, nil
	}

	if err := s.db.UpdateStreak(ctx); err != nil {
		s.errorMsg += fmt.Sprintf(" | Failed to update streak: %v", err)
	}

	s.status = roadmap.StatusVerified
	return s, func() tea.Msg {
		result := fmt.Sprintf("%s\n\n+%d XP (verify: %d%%) Tests passed — Problem Verified", msg.output, xp, 70)
		return GlobalNotificationMsg{Message: result}
	}
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

	alreadyClaimed, err := s.db.HasRewardEvent(ctx, problemID, "submit")
	if err != nil {
		return
	}

	if err := s.db.SetProgress(ctx, problemID, roadmap.StatusSolved); err != nil {
		return
	}
	s.status = roadmap.StatusSolved

	if alreadyClaimed {
		s.errorMsg = "Accepted by LeetCode. Reward already claimed."
		return
	}

	totalXP := 0
	xp := store.XPForDifficulty(s.problem.Difficulty) * 30 / 100
	if xp > 0 {
		if err := s.db.AddXP(ctx, xp); err != nil {
			return
		}
		totalXP += xp
	}

	event := &store.RewardEvent{
		ProblemID: problemID,
		Kind:      "submit",
		XP:        xp,
	}
	if err := s.db.RecordRewardEvent(ctx, event); err != nil {
		return
	}

	hasVerify, err := s.db.HasRewardEvent(ctx, problemID, "verify")
	if err != nil {
		return
	}
	if !hasVerify {
		verifyXP := store.XPForDifficulty(s.problem.Difficulty) * 70 / 100
		if verifyXP > 0 {
			if err := s.db.AddXP(ctx, verifyXP); err != nil {
				return
			}
			totalXP += verifyXP
		}
		verifyEvent := &store.RewardEvent{
			ProblemID: problemID,
			Kind:      "verify",
			XP:        verifyXP,
		}
		if err := s.db.RecordRewardEvent(ctx, verifyEvent); err != nil {
			return
		}
	}

	if err := s.db.UpdateStreak(ctx); err != nil {
		return
	}
	s.errorMsg = fmt.Sprintf("Accepted by LeetCode. +%d XP.", totalXP)
}

func (s *ProblemDetailScreen) handleMarkSolved() (Screen, tea.Cmd) {
	if s.status == roadmap.StatusSolved {
		s.errorMsg = "Already Solved."
		return s, nil
	}
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}

	s.manualSolveMode = true
	s.manualSolveInput = ""
	s.errorMsg = ""

	return s, nil
}

func (s *ProblemDetailScreen) handleManualSolveKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		s.manualSolveMode = false
		s.manualSolveInput = ""
		return s, nil

	case tea.KeyEnter:
		if s.manualSolveInput == "SOLVE" {
			s.confirmManualSolve()
		}
		return s, nil

	case tea.KeyBackspace:
		if len(s.manualSolveInput) > 0 {
			s.manualSolveInput = s.manualSolveInput[:len(s.manualSolveInput)-1]
		}
		return s, nil

	case tea.KeyRunes:
		s.manualSolveInput += string(msg.Runes)
		return s, nil

	default:
		return s, nil
	}
}

func (s *ProblemDetailScreen) confirmManualSolve() {
	ctx := context.Background()

	if err := s.db.SetProgress(ctx, s.problem.ID, roadmap.StatusSolved); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to update progress: %v", err)
		s.manualSolveMode = false
		return
	}

	if err := s.db.UpdateStreak(ctx); err != nil {
		s.errorMsg += fmt.Sprintf(" | Failed to update streak: %v", err)
	}

	s.status = roadmap.StatusSolved
	s.manualSolveMode = false
	s.manualSolveInput = ""
	s.errorMsg = "Manually marked Solved. No XP awarded."
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
	if s.manualSolveMode {
		return s.renderManualSolveConfirmation()
	}

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
	case roadmap.StatusVerified:
		return lipgloss.NewStyle().Foreground(s.theme.PrimaryAccent).Bold(true).Render("[VERIFIED]")
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
		if progress[id] == roadmap.StatusSolved || progress[id] == roadmap.StatusVerified {
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
	if s.manualSolveMode {
		return s.renderManualSolveFooter()
	}

	primaryLabel := "start"
	switch s.status {
	case roadmap.StatusAvailable:
		primaryLabel = "start"
	case roadmap.StatusVerified:
		primaryLabel = "submit"
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

func (s *ProblemDetailScreen) renderManualSolveConfirmation() string {
	var lines []string

	lines = append(lines, s.theme.Title.Render(fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title)))
	lines = append(lines, "")

	warning := lipgloss.NewStyle().
		Foreground(s.theme.Warning).
		Bold(true).
		Render("Manual Solve bypasses verification and awards no XP.")
	lines = append(lines, warning)
	lines = append(lines, "")

	prompt := lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Render("Type SOLVE to confirm:")
	lines = append(lines, prompt)
	lines = append(lines, "")

	inputDisplay := s.manualSolveInput
	if inputDisplay == "" {
		inputDisplay = "_"
	}
	inputStyle := lipgloss.NewStyle().
		Foreground(s.theme.PrimaryAccent).
		Bold(true)
	lines = append(lines, inputStyle.Render(inputDisplay))

	content := strings.Join(lines, "\n")
	footer := s.renderManualSolveFooter()
	return content + "\n" + footer
}

func (s *ProblemDetailScreen) renderManualSolveFooter() string {
	items := []string{
		s.theme.Key.Render("type SOLVE") + " to confirm",
		s.theme.Key.Render("esc") + " to cancel",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg(t)
	})
}
