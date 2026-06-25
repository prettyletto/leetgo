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
	"github.com/prettyletto/leetgo/internal/recommendation"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
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
	manualSolveNote  string
	manualSolvePhase string // "" = confirm code, "note" = optional note

	submitAnywayMode  bool
	submitAnywayInput string

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

	case testRunResultMsg:
		return s.handleTestRunResult(msg)

	case spinnerTickMsg:
		if s.submitting && s.allowsSpinnerMotion() {
			s.spinnerFrame = (s.spinnerFrame + 1) % len(spinnerFrames)
			return s, spinnerTickCmd()
		}
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		if s.submitAnywayMode {
			return s.handleSubmitAnywayKey(msg)
		}
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
		startedAt := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(startedAt)
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
			duration:   duration,
		}
	}
}

func (s *ProblemDetailScreen) handleSubmit() (Screen, tea.Cmd) {
	return s.handleSubmitWithLocalTests(true)
}

func (s *ProblemDetailScreen) handleSubmitWithLocalTests(runLocalTests bool) (Screen, tea.Cmd) {
	if s.status != roadmap.StatusInProgress && s.status != roadmap.StatusVerified && s.status != roadmap.StatusSolved {
		s.errorMsg = "Start the problem first."
		return s, nil
	}

	if s.leetcode == nil || !s.leetcode.IsAuthenticated() {
		s.errorMsg = "Session expired. Run `leetgo auth` to reconnect."
		return s, nil
	}

	stubPath, _ := s.stubPaths()

	s.submitting = true
	if runLocalTests {
		s.errorMsg = "Running local TestSuite before Submission..."
	} else {
		s.errorMsg = "Submitting to LeetCode without local TestSuite..."
	}
	s.submitAnywayMode = false
	s.submitAnywayInput = ""

	lang := leetcodeLang(s.cfg.Language)
	slug := s.problem.Slug
	client := s.leetcode
	var testCmd *exec.Cmd
	if runLocalTests {
		testCmd = s.testCommand()
	}

	cmds := []tea.Cmd{
		func() tea.Msg {
			var localOutput string
			var localDuration time.Duration
			if runLocalTests {
				localStartedAt := time.Now()
				output, testErr := testCmd.CombinedOutput()
				localDuration = time.Since(localStartedAt)
				localOutput = strings.TrimSpace(string(output))
				if testErr != nil {
					if localOutput != "" {
						localOutput += "\n"
					}
					localOutput += fmt.Sprintf("TestSuite error: %v", testErr)
					return submitResultMsg{
						problemID:         s.problem.ID,
						slug:              slug,
						language:          lang,
						err:               testErr,
						localTestRan:      true,
						localTestPassed:   false,
						localTestOutput:   localOutput,
						localTestDuration: localDuration,
					}
				}
			}

			code, err := os.ReadFile(stubPath)
			if err != nil {
				return submitResultMsg{
					problemID:         s.problem.ID,
					slug:              slug,
					language:          lang,
					err:               err,
					localTestRan:      runLocalTests,
					localTestPassed:   runLocalTests,
					localTestOutput:   localOutput,
					localTestDuration: localDuration,
				}
			}
			startedAt := time.Now()
			result, err := client.Submit(context.Background(), s.problem.ID, slug, lang, string(code))
			return submitResultMsg{
				problemID:         s.problem.ID,
				slug:              slug,
				language:          lang,
				result:            result,
				err:               err,
				duration:          time.Since(startedAt),
				localTestRan:      runLocalTests,
				localTestPassed:   runLocalTests,
				localTestOutput:   localOutput,
				localTestDuration: localDuration,
			}
		},
	}
	if s.allowsSpinnerMotion() {
		cmds = append([]tea.Cmd{spinnerTickCmd()}, cmds...)
	}
	return s, tea.Batch(cmds...)
}

func (s *ProblemDetailScreen) allowsSpinnerMotion() bool {
	return s.cfg.MotionPreference != "off"
}

func (s *ProblemDetailScreen) handleSubmitResult(msg submitResultMsg) (Screen, tea.Cmd) {
	s.submitting = false
	if msg.localTestRan {
		s.recordSubmitLocalAttempt(msg)
		if !msg.localTestPassed {
			s.errorMsg = "Local TestSuite failed. Fix your solution before Submission, or type SUBMIT to submit anyway."
			s.submitAnywayMode = true
			s.submitAnywayInput = ""
			if msg.localTestOutput != "" {
				return s, func() tea.Msg { return GlobalNotificationMsg{Message: msg.localTestOutput} }
			}
			return s, nil
		}
		if err := s.markVerifiedFromSubmitPrecheck(msg); err != nil {
			s.errorMsg = err.Error()
			return s, nil
		}
	}

	if msg.err != nil {
		s.errorMsg = fmt.Sprintf("Submit failed: %v", msg.err)
		return s, nil
	}

	s.recordSolveLog(msg)

	if msg.result.StatusCode == 10 {
		s.markAcceptedSubmission(msg.problemID, msg.duration)
		if s.errorMsg == "" {
			s.errorMsg = acceptedMessage(msg.result)
		}
	} else {
		s.errorMsg = fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests)
		if msg.result.Error != "" {
			s.errorMsg += ": " + msg.result.Error
		}
	}

	return s, nil
}

func (s *ProblemDetailScreen) handleSubmitAnywayKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		s.submitAnywayMode = false
		s.submitAnywayInput = ""
		s.errorMsg = "Submit anyway cancelled."
		return s, nil
	case tea.KeyEnter:
		if s.submitAnywayInput == "SUBMIT" {
			return s.handleSubmitWithLocalTests(false)
		}
		return s, nil
	case tea.KeyBackspace:
		if len(s.submitAnywayInput) > 0 {
			s.submitAnywayInput = s.submitAnywayInput[:len(s.submitAnywayInput)-1]
		}
		return s, nil
	case tea.KeyRunes:
		s.submitAnywayInput += string(msg.Runes)
		return s, nil
	default:
		return s, nil
	}
}

func (s *ProblemDetailScreen) recordSubmitLocalAttempt(msg submitResultMsg) {
	duration := msg.localTestDuration
	if duration < 0 {
		duration = 0
	}
	_ = s.db.RecordAttempt(context.Background(), &store.AttemptRecord{
		ProblemID: msg.problemID,
		Timestamp: time.Now().Add(-duration),
		Duration:  duration,
		Passed:    msg.localTestPassed,
	})
}

func (s *ProblemDetailScreen) markVerifiedFromSubmitPrecheck(msg submitResultMsg) error {
	ctx := context.Background()
	if s.status == roadmap.StatusSolved {
		return nil
	}
	claimed, err := s.db.HasRewardEvent(ctx, msg.problemID, "verify")
	if err != nil {
		return fmt.Errorf("failed to check verify reward: %w", err)
	}
	if !claimed {
		xp := store.XPForDifficulty(s.problem.Difficulty) * 70 / 100
		if xp > 0 {
			if err := s.db.AddXP(ctx, xp); err != nil {
				return fmt.Errorf("failed to add verify XP: %w", err)
			}
		}
		if err := s.db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: msg.problemID, Kind: "verify", XP: xp}); err != nil {
			return fmt.Errorf("failed to record verify reward: %w", err)
		}
	}
	if err := s.db.SetProgress(ctx, msg.problemID, roadmap.StatusVerified); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	s.status = roadmap.StatusVerified
	return nil
}

func acceptedMessage(result *leetcode.SubmissionResult) string {
	parts := []string{"Accepted!"}
	if result.Runtime != "" {
		parts = append(parts, "Runtime: "+result.Runtime)
	}
	if result.Memory != "" {
		parts = append(parts, "Memory: "+result.Memory)
	}
	return strings.Join(parts, " ")
}

func (s *ProblemDetailScreen) handleTestRunResult(msg testRunResultMsg) (Screen, tea.Cmd) {
	ctx := context.Background()
	s.recordLocalAttempt(ctx, msg)

	if !msg.passed {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: msg.output}
		}
	}

	if s.status == roadmap.StatusSolved {
		return s.handleReviewTestPass(ctx, msg)
	}

	if s.status == roadmap.StatusVerified {
		return s, func() tea.Msg {
			result := msg.output + "\n\nTests passed. Reward already claimed."
			return GlobalNotificationMsg{Message: result}
		}
	}

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

func (s *ProblemDetailScreen) recordLocalAttempt(ctx context.Context, msg testRunResultMsg) {
	duration := msg.duration
	if duration < 0 {
		duration = 0
	}
	_ = s.db.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: msg.problemID,
		Timestamp: time.Now().Add(-duration),
		Duration:  duration,
		Passed:    msg.passed,
	})
}

func (s *ProblemDetailScreen) handleReviewTestPass(ctx context.Context, msg testRunResultMsg) (Screen, tea.Cmd) {
	cycles, err := s.db.GetReviewCyclesForProblem(ctx, msg.problemID)
	if err != nil || len(cycles) == 0 {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: msg.output + "\n\nTests passed."}
		}
	}

	var completedCount int
	var rewardedXP int
	for _, rc := range cycles {
		if rc.CompletedAt != nil {
			continue
		}
		if err := s.db.CompleteReviewCycle(ctx, rc.ID); err != nil {
			continue
		}
		completedCount++

		if rc.RewardedAt != nil {
			continue
		}
		reviewXP := 5
		if err := s.db.AddXP(ctx, reviewXP); err != nil {
			continue
		}
		rewardedXP += reviewXP
		if err := s.db.RewardReviewCycle(ctx, rc.ID); err != nil {
			continue
		}
	}

	if completedCount == 0 {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: msg.output + "\n\nTests passed. Review already completed."}
		}
	}

	result := views.RenderRewardMoment(views.RewardMoment{
		Title:   "Review Complete",
		Subject: fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title),
		XP:      rewardedXP,
		Reward:  fmt.Sprintf("Tests passed. %d Review Cycle(s) completed", completedCount),
		Next:    s.nextRecommendationTitle(ctx),
		Actions: []string{"enter open", "x run tests", "s submit"},
	}, viewPalette(s.theme))
	if msg.output != "" {
		result = msg.output + "\n\n" + result
	}
	return s, func() tea.Msg {
		return GlobalNotificationMsg{Message: result}
	}
}

func (s *ProblemDetailScreen) recordSolveLog(msg submitResultMsg) {
	if msg.result == nil {
		return
	}
	_ = s.db.RecordAttempt(context.Background(), &store.AttemptRecord{
		ProblemID: msg.problemID,
		Timestamp: time.Now().Add(-msg.duration),
		Duration:  msg.duration,
		Passed:    msg.result.StatusCode == 10,
	})
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

func (s *ProblemDetailScreen) markAcceptedSubmission(problemID int, duration time.Duration) {
	ctx := context.Background()

	alreadyClaimed, err := s.db.HasRewardEvent(ctx, problemID, "submit")
	if err != nil {
		return
	}

	if err := s.db.SetProgress(ctx, problemID, roadmap.StatusSolved); err != nil {
		return
	}
	s.status = roadmap.StatusSolved

	s.recordAcceptedProvenance(ctx, problemID)

	if alreadyClaimed {
		s.errorMsg = views.RenderRewardMoment(views.RewardMoment{
			Title:    "Problem Solved",
			Subject:  fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title),
			Reward:   "Accepted by LeetCode. Reward already claimed.",
			Duration: duration,
			Next:     s.nextRecommendationTitle(ctx),
			Reason:   "Accepted Solve is trusted completion.",
			Actions:  []string{"esc back", "s practice log"},
		}, viewPalette(s.theme))
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
	s.errorMsg = views.RenderRewardMoment(views.RewardMoment{
		Title:    "Problem Solved",
		Subject:  fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title),
		XP:       totalXP,
		Reward:   "Accepted by LeetCode",
		Duration: duration,
		Unlocked: s.directUnlockLabels(),
		Next:     s.nextRecommendationTitle(ctx),
		Reason:   "Accepted Solve unlocks dependent Problems and counts toward Roadmap completion.",
		Actions:  []string{"esc back", "s practice log", "r roadmap"},
	}, viewPalette(s.theme))
}

func (s *ProblemDetailScreen) recordAcceptedProvenance(ctx context.Context, problemID int) {
	sp, err := s.db.GetSolveProvenance(ctx, problemID)
	if err != nil {
		return
	}

	logs, err := s.db.GetSolveLogsForProblem(ctx, problemID)
	var logID *int
	if err == nil && len(logs) > 0 {
		latestAccepted := -1
		for _, l := range logs {
			if l.StatusCode == 10 && l.ID > latestAccepted {
				latestAccepted = l.ID
			}
		}
		if latestAccepted > 0 {
			logID = &latestAccepted
		}
	}

	if sp != nil && sp.Kind == "manual" {
		if err := s.db.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			Note:       sp.Note,
			SolveLogID: logID,
		}); err != nil {
			return
		}
	} else {
		if err := s.db.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			SolveLogID: logID,
		}); err != nil {
			return
		}
	}
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
	s.manualSolveNote = ""
	s.manualSolvePhase = ""
	s.errorMsg = ""

	return s, nil
}

func (s *ProblemDetailScreen) handleManualSolveKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	if s.manualSolvePhase == "note" {
		switch msg.Type {
		case tea.KeyEsc:
			s.manualSolveMode = false
			s.manualSolveInput = ""
			s.manualSolveNote = ""
			s.manualSolvePhase = ""
			return s, nil

		case tea.KeyEnter:
			s.confirmManualSolve()
			return s, nil

		case tea.KeyBackspace:
			if len(s.manualSolveNote) > 0 {
				s.manualSolveNote = s.manualSolveNote[:len(s.manualSolveNote)-1]
			}
			return s, nil

		case tea.KeyRunes:
			s.manualSolveNote += string(msg.Runes)
			return s, nil
		}
		return s, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		s.manualSolveMode = false
		s.manualSolveInput = ""
		s.manualSolveNote = ""
		s.manualSolvePhase = ""
		return s, nil

	case tea.KeyEnter:
		if s.manualSolveInput == "SOLVE" {
			s.manualSolvePhase = "note"
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

	if err := s.db.RecordSolveProvenance(ctx, &store.SolveProvenance{
		ProblemID: s.problem.ID,
		Kind:      "manual",
		Note:      s.manualSolveNote,
	}); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to record solve provenance: %v", err)
		s.manualSolveMode = false
		return
	}

	if err := s.db.UpdateStreak(ctx); err != nil {
		s.errorMsg += fmt.Sprintf(" | Failed to update streak: %v", err)
	}

	s.status = roadmap.StatusSolved
	s.manualSolveMode = false
	s.manualSolveInput = ""
	s.manualSolveNote = ""
	s.errorMsg = views.RenderRewardMoment(views.RewardMoment{
		Title:    "Manual Solve",
		Subject:  fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title),
		Reward:   "Manually marked Solved. No XP awarded.",
		Unlocked: s.directUnlockLabels(),
		Next:     s.nextRecommendationTitle(ctx),
		Reason:   "Manual Solve unlocks progression, but Accepted Submission is still recommended for confidence and XP.",
		Actions:  []string{"s submit for Accepted Solve", "x run local TestSuite", "esc back"},
	}, viewPalette(s.theme))
}

func (s *ProblemDetailScreen) directUnlockLabels() []string {
	unlocks := s.directUnlocks()
	labels := make([]string, 0, len(unlocks))
	for _, p := range unlocks {
		labels = append(labels, fmt.Sprintf("#%d %s", p.ID, p.Title))
	}
	return labels
}

func (s *ProblemDetailScreen) nextRecommendationTitle(ctx context.Context) string {
	calc := recommendation.NewCalculator(s.db, s.roadmap)
	actions, err := calc.Calculate(ctx)
	if err != nil || len(actions) == 0 {
		return ""
	}
	return actions[0].Title
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
	if s.submitAnywayMode {
		return s.renderSubmitAnywayConfirmation()
	}
	if s.manualSolveMode {
		return s.renderManualSolveConfirmation()
	}

	screenLabel, briefLabel, filesLabel := themeProblemLabels(s.theme)
	header := renderScreenHeader(s.theme, fmt.Sprintf("%s: #%d %s", screenLabel, s.problem.ID, s.problem.Title), fmt.Sprintf("Difficulty: %s  •  Category: %s  •  Stage: %s", s.problem.Difficulty, s.problem.Category, s.stageName()))
	if s.problem.ProblemTimeEstimate != "" {
		header += "\n" + s.theme.Subtitle.Render("Time estimate: "+s.problem.ProblemTimeEstimate)
	}

	statusLabel := s.renderStatusDetail()
	leftSections := []string{statusLabel}
	var contextLines []string

	contextLines = append(contextLines, s.renderProblemBriefBlock(briefLabel))

	if len(s.problem.Prerequisites) == 0 {
		contextLines = append(contextLines, "Requires: none")
	} else {
		prereqs := s.prerequisiteLabels()
		contextLines = append(contextLines, "Requires: "+strings.Join(prereqs, ", "))
	}

	if s.status == roadmap.StatusLocked {
		blocked := s.missingPrerequisites()
		if len(blocked) > 0 {
			contextLines = append(contextLines, lipgloss.NewStyle().
				Foreground(s.theme.Warning).
				Render("Blocked by: "+strings.Join(blocked, ", ")+". Clear a prerequisite to unlock this Problem."))
		}
	}
	leftSections = append(leftSections, renderThemedPanel(s.theme, "Context", strings.Join(contextLines, "\n\n"), false))

	var progressionLines []string
	unlocks := s.directUnlocks()
	if len(unlocks) > 0 {
		progressionLines = append(progressionLines, s.theme.Key.Render("Unlocks"))
		for _, u := range unlocks {
			progressionLines = append(progressionLines, fmt.Sprintf("  #%d %s (%s)", u.ID, u.Title, u.Difficulty))
		}
	}

	indirect := s.indirectUnlocks(2)
	if len(indirect) > 0 {
		if len(progressionLines) > 0 {
			progressionLines = append(progressionLines, "")
		}
		progressionLines = append(progressionLines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render("Builds Toward"))
		for _, u := range indirect {
			progressionLines = append(progressionLines, fmt.Sprintf("  #%d %s", u.ID, u.Title))
		}
	}
	if len(progressionLines) > 0 {
		leftSections = append(leftSections, renderThemedPanel(s.theme, "Progression", strings.Join(progressionLines, "\n"), false))
	}

	var workspaceLines []string
	if s.status != roadmap.StatusLocked {
		stubPath, testPath := s.stubPaths()
		workspaceLines = append(workspaceLines,
			fmt.Sprintf("Stub: %s", stubPath),
			fmt.Sprintf("Test: %s", testPath),
		)
		leftSections = append(leftSections, renderThemedPanel(s.theme, filesLabel, strings.Join(workspaceLines, "\n"), false))
	}

	rightSections := []string{}
	practiceLog := BuildPracticeLog(s.db, s.problem.ID)
	if len(practiceLog) > 0 {
		var logLines []string
		for _, entry := range practiceLog {
			when := entry.Timestamp.Format("2006-01-02 15:04")
			line := fmt.Sprintf("  %s  %s", s.practiceLogSymbol(entry), when)
			line += "  " + entry.Kind
			if entry.Detail != "" {
				line += " · " + entry.Detail
			}
			logLines = append(logLines, line)
		}
		rightSections = append(rightSections, renderThemedPanel(s.theme, "Practice Log", strings.Join(logLines, "\n"), false))
	}

	statusLines := []string{}
	if s.submitting {
		if s.allowsSpinnerMotion() {
			frame := spinnerFrames[s.spinnerFrame%len(spinnerFrames)]
			statusLines = append(statusLines, s.theme.Spinner.Render(frame+" Submitting to LeetCode..."))
		} else {
			statusLines = append(statusLines, s.theme.Spinner.Render("Submitting to LeetCode..."))
		}
	}

	if s.errorMsg != "" {
		style := lipgloss.NewStyle().Foreground(s.theme.Muted)
		if strings.Contains(s.errorMsg, "failed") || strings.Contains(s.errorMsg, "locked") || strings.Contains(s.errorMsg, "Failed") {
			style = lipgloss.NewStyle().Foreground(s.theme.Danger)
		} else if strings.Contains(s.errorMsg, "Accepted") || strings.Contains(s.errorMsg, "XP") {
			style = lipgloss.NewStyle().Foreground(s.theme.Success)
		} else if strings.Contains(s.errorMsg, "Start") || strings.Contains(s.errorMsg, "Started") {
			style = lipgloss.NewStyle().Foreground(s.theme.Warning)
		}
		statusLines = append(statusLines, style.Render(s.errorMsg))
	}
	if len(statusLines) > 0 {
		rightSections = append(rightSections, renderThemedPanel(s.theme, "Status", strings.Join(statusLines, "\n"), false))
	}

	footer := s.renderFooter()
	left := strings.Join(leftSections, "\n\n")
	right := strings.Join(rightSections, "\n\n")
	body := left
	if right != "" {
		if s.width >= 110 {
			leftWidth := maxInt(46, s.width-44)
			body = lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(leftWidth).Render(left),
				"  ",
				lipgloss.NewStyle().Width(34).Render(right),
			)
		} else {
			body = left + "\n\n" + right
		}
	}
	return renderScreenShell(s.theme, s.width, s.height, header, body, footer)
}

func (s *ProblemDetailScreen) renderSubmitAnywayConfirmation() string {
	var lines []string
	lines = append(lines, s.theme.Title.Render("Submit Anyway"))
	lines = append(lines, fmt.Sprintf("Problem: #%d %s", s.problem.ID, s.problem.Title))
	lines = append(lines, "")
	warning := strings.Join([]string{
		lipgloss.NewStyle().Foreground(s.theme.Danger).Bold(true).Render("Local TestSuite failed."),
		"Submitting anyway skips local verification and sends your current Stub to LeetCode.",
		"Use this only when generated tests are wrong or you intentionally want LeetCode as source of truth.",
	}, "\n")
	lines = append(lines, renderThemedPanel(s.theme, "Serious Warning", warning, false))
	lines = append(lines, "")
	lines = append(lines, "Type SUBMIT to confirm:")
	lines = append(lines, s.theme.FocusedPanel.Render(s.submitAnywayInput))
	lines = append(lines, "")
	lines = append(lines, s.theme.Footer.Render(s.theme.Key.Render("enter")+" confirm  "+s.theme.Key.Render("esc")+" cancel"))
	return strings.Join(lines, "\n")
}

func (s *ProblemDetailScreen) renderProblemBriefBlock(label string) string {
	if s.problem.Summary == "" && s.problem.PracticeFocus == "" {
		return renderThemedPanel(s.theme, label, "No brief available yet.", false)
	}
	var lines []string
	if s.problem.Summary != "" {
		lines = append(lines, s.problem.Summary)
	}
	if s.problem.PracticeFocus != "" {
		lines = append(lines, "", fmt.Sprintf("Practice Focus: %s", s.problem.PracticeFocus))
	}
	if s.problem.WhyNow != "" {
		lines = append(lines, fmt.Sprintf("Why now: %s", s.problem.WhyNow))
	}
	if s.problem.UnlockImpact != "" {
		lines = append(lines, fmt.Sprintf("Unlock Impact: %s", s.problem.UnlockImpact))
	}
	return renderThemedPanel(s.theme, label, strings.Join(lines, "\n"), false)
}

func (s *ProblemDetailScreen) renderProblemBrief(lines *[]string, label string) {
	if s.problem.Summary == "" && s.problem.PracticeFocus == "" {
		return
	}
	*lines = append(*lines, "")
	*lines = append(*lines, s.theme.Key.Render(label))
	if s.problem.Summary != "" {
		*lines = append(*lines, fmt.Sprintf("  %s", s.problem.Summary))
	}
	if s.problem.PracticeFocus != "" {
		*lines = append(*lines, fmt.Sprintf("  Practice Focus: %s", s.problem.PracticeFocus))
	}
	if s.problem.WhyNow != "" {
		*lines = append(*lines, fmt.Sprintf("  Why now: %s", s.problem.WhyNow))
	}
	if s.problem.UnlockImpact != "" {
		*lines = append(*lines, fmt.Sprintf("  Unlock Impact: %s", s.problem.UnlockImpact))
	}
}

func (s *ProblemDetailScreen) renderStatusDetail() string {
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	switch s.status {
	case roadmap.StatusSolved:
		label := renderStatusPill(s.theme, symbols, roadmap.StatusSolved)
		sp, _ := s.db.GetSolveProvenance(context.Background(), s.problem.ID)
		if sp != nil {
			if sp.Kind == "manual" {
				label = views.StatusPill("MANUAL SOLVE", s.theme.Success)
			} else {
				label = views.StatusPill("ACCEPTED SOLVE", s.theme.Success)
			}
		}
		return renderThemedPanel(s.theme, "Status", label+"\nPractice Log and unlock impact are now the priority.", false)
	case roadmap.StatusVerified:
		return renderThemedPanel(s.theme, "Status", renderStatusPill(s.theme, symbols, roadmap.StatusVerified)+"\nPrimary: Submit to LeetCode. Secondary: Manual Solve if needed.", false)
	case roadmap.StatusInProgress:
		return renderThemedPanel(s.theme, "Status", renderStatusPill(s.theme, symbols, roadmap.StatusInProgress)+"\nKeep file paths and TestSuite actions close.", false)
	case roadmap.StatusAvailable:
		return renderThemedPanel(s.theme, "Status", renderStatusPill(s.theme, symbols, roadmap.StatusAvailable)+"\nRead the Problem Brief, then start the Problem.", false)
	default:
		return renderThemedPanel(s.theme, "Status", renderStatusPill(s.theme, symbols, roadmap.StatusLocked)+"\nClear the blocker before starting.", false)
	}
}

func (s *ProblemDetailScreen) practiceLogSymbol(entry ProblemLogEntry) string {
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	switch {
	case strings.Contains(entry.Kind, "accepted") || entry.Kind == "Accepted Solve":
		return symbols.Solved
	case strings.Contains(entry.Kind, "failed"):
		return "!"
	case entry.Kind == "Manual Solve":
		return "M"
	case strings.Contains(entry.Kind, "passed"):
		return symbols.Verified
	default:
		return symbols.Bullet
	}
}

func (s *ProblemDetailScreen) directUnlocks() []*roadmap.Problem {
	var result []*roadmap.Problem
	progress, _ := s.db.GetAllProgress(context.Background())
	for _, p := range s.roadmap.Graph.Problems {
		for _, prereq := range p.Prerequisites {
			if prereq == s.problem.ID {
				if progress[p.ID] != roadmap.StatusSolved {
					result = append(result, p)
				}
				break
			}
		}
	}
	return result
}

func (s *ProblemDetailScreen) indirectUnlocks(depth int) []*roadmap.Problem {
	if depth <= 0 {
		return nil
	}
	var result []*roadmap.Problem
	progress, _ := s.db.GetAllProgress(context.Background())
	seen := make(map[int]bool)
	for _, p := range s.roadmap.Graph.Problems {
		for _, prereq := range p.Prerequisites {
			if prereq == s.problem.ID {
				if !seen[p.ID] && progress[p.ID] != roadmap.StatusSolved {
					seen[p.ID] = true
					result = append(result, p)
				}
				for _, pp := range s.roadmap.Graph.Problems {
					for _, ppr := range pp.Prerequisites {
						if ppr == p.ID && !seen[pp.ID] && progress[pp.ID] != roadmap.StatusSolved {
							seen[pp.ID] = true
							result = append(result, pp)
						}
					}
				}
			}
		}
	}
	if len(result) > 3 {
		result = result[:3]
	}
	return result
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
	if s.manualSolveMode {
		return s.renderManualSolveFooter()
	}

	primaryLabel := "open"
	switch s.status {
	case roadmap.StatusAvailable:
		primaryLabel = "start"
	case roadmap.StatusVerified:
		primaryLabel = "submit (primary)"
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
		s.theme.Key.Render("esc") + " back",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}

func (s *ProblemDetailScreen) renderManualSolveConfirmation() string {
	var lines []string

	lines = append(lines, s.theme.Title.Render(fmt.Sprintf("#%d %s", s.problem.ID, s.problem.Title)))
	lines = append(lines, "")

	if s.manualSolvePhase == "note" {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render("Optional note (press Enter to skip):"))
		lines = append(lines, "")

		inputDisplay := s.manualSolveNote
		if inputDisplay == "" {
			inputDisplay = "_"
		}
		inputStyle := lipgloss.NewStyle().
			Foreground(s.theme.PrimaryAccent).
			Bold(true)
		lines = append(lines, inputStyle.Render(inputDisplay))
	} else {
		warning := lipgloss.NewStyle().
			Foreground(s.theme.Warning).
			Bold(true).
			Render("Mark as manually solved? This will unlock dependent Problems, but you will not earn XP unless LeetCode accepts a Submission later.")
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
	}

	content := strings.Join(lines, "\n")
	footer := s.renderManualSolveFooter()
	return content + "\n" + footer
}

func (s *ProblemDetailScreen) renderManualSolveFooter() string {
	if s.manualSolvePhase == "note" {
		items := []string{
			s.theme.Key.Render("enter") + " confirm",
			s.theme.Key.Render("esc") + " cancel",
		}
		return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
	}
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
