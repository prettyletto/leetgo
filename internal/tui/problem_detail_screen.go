package tui

import (
	"context"
	"encoding/json"
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

	errorMsg        string
	testResultMsg   string
	submitResultMsg string
	submitting      bool
	spinnerFrame    int

	problemDescription string
	descriptionStatus  string
	scrapedExamples    []generator.ExampleSpec

	manualSolveMode  bool
	manualSolveInput string
	manualSolveNote  string
	manualSolvePhase string // "" = confirm code, "note" = optional note

	submitAnywayMode  bool
	submitAnywayInput string

	regenerateMode  bool
	regenerateInput string

	debugPickerMode bool
	debugCaseIndex  int
	debugLaunchMode editorLaunchMode

	width  int
	height int

	returnScreen string
	returnStage  string

	detailTab    int
	detailScroll int
}

const (
	problemDetailTabContext = iota
	problemDetailTabProgression
)

type spinnerTickMsg time.Time

type problemDescriptionMsg struct {
	problemID   int
	description string
	err         error
}

func NewProblemDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap, problemID int) *ProblemDetailScreen {
	p, ok := rm.Graph.Problems[problemID]
	if !ok {
		p = &roadmap.Problem{ID: problemID, Title: "Unknown Problem"}
	}

	s := &ProblemDetailScreen{
		cfg:          cfg,
		theme:        theme,
		db:           db,
		roadmap:      rm,
		problem:      p,
		returnScreen: ScreenStageDetail,
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
	return s.fetchProblemDescriptionCmd()
}

func (s *ProblemDetailScreen) fetchProblemDescriptionCmd() tea.Cmd {
	problemID := s.problem.ID
	slug := s.problem.Slug
	client := s.leetcode
	if client == nil || slug == "" {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		desc, err := client.ProblemDescription(ctx, slug)
		if err != nil {
			return problemDescriptionMsg{problemID: problemID, err: err}
		}
		return problemDescriptionMsg{problemID: problemID, description: desc.ContentText}
	}
}

func (s *ProblemDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateMsg:
		return s, nil

	case submitResultMsg:
		return s.handleSubmitResult(msg)

	case testRunResultMsg:
		return s.handleTestRunResult(msg)

	case problemDescriptionMsg:
		if msg.problemID != s.problem.ID {
			return s, nil
		}
		if msg.err != nil {
			s.descriptionStatus = "Using bundled brief; full LeetCode description unavailable."
			return s, nil
		}
		s.problemDescription = msg.description
		s.descriptionStatus = "Full LeetCode description loaded."
		if spec, ok := generator.SpecForProblem(s.problem); ok {
			if examples, err := scrapedExamplesFromDescription(spec, msg.description); err == nil {
				s.mergeExternalExamples(examples)
				_, testPath := s.stubPaths()
				_, _ = appendScrapedLeetCodeExamples(s.problem, s.cfg.Language, testPath, examples)
			}
		}
		s.detailScroll = 0
		return s, nil

	case spinnerTickMsg:
		if s.submitting && s.allowsSpinnerMotion() {
			s.spinnerFrame = (s.spinnerFrame + 1) % len(spinnerFrames)
			return s, spinnerTickCmd()
		}
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.detailScroll = 0
		return s, nil

	case tea.KeyMsg:
		if s.submitting {
			switch msg.String() {
			case "q", "ctrl+c":
				return s, tea.Quit
			}
			return s, nil
		}
		if s.testResultMsg != "" {
			switch msg.String() {
			case "esc", "backspace":
				s.testResultMsg = ""
				return s, nil
			case "q", "ctrl+c":
				return s, tea.Quit
			}
		}
		if s.submitResultMsg != "" {
			switch msg.String() {
			case "esc", "backspace":
				s.submitResultMsg = ""
				return s, nil
			case "d":
				if len(s.scrapedExamples) > 0 && !strings.Contains(strings.ToLower(s.submitResultMsg), "accepted") {
					s.submitResultMsg = ""
					return s.handleDebug(editorLaunchAttached)
				}
			case "D":
				if len(s.scrapedExamples) > 0 && !strings.Contains(strings.ToLower(s.submitResultMsg), "accepted") {
					s.submitResultMsg = ""
					return s.handleDebug(editorLaunchDetached)
				}
			case "q", "ctrl+c":
				return s, tea.Quit
			}
			return s, nil
		}
		if s.submitAnywayMode {
			return s.handleSubmitAnywayKey(msg)
		}
		if s.regenerateMode {
			return s.handleRegenerateKey(msg)
		}
		if s.debugPickerMode {
			return s.handleDebugPickerKey(msg)
		}
		if s.manualSolveMode {
			return s.handleManualSolveKey(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "esc", "backspace":
			return s, func() tea.Msg {
				return s.backNavigationMsg()
			}

		case "enter":
			return s.handlePrimaryAction()

		case "e":
			return s.handleOpenEditor(editorLaunchAttached)

		case "E", "o":
			return s.handleOpenEditor(editorLaunchDetached)

		case "d":
			return s.handleDebug(editorLaunchAttached)

		case "D":
			return s.handleDebug(editorLaunchDetached)

		case "x":
			return s.handleRunTests()

		case "s":
			return s.handleSubmit()

		case "m":
			return s.handleMarkSolved()

		case "R":
			return s.handleRegenerateFiles()

		case "left", "h":
			if s.canCycleContextProgression() {
				s.detailTab = problemDetailTabContext
				s.detailScroll = 0
			}
			return s, nil

		case "right", "l":
			if s.canCycleContextProgression() {
				s.detailTab = problemDetailTabProgression
				s.detailScroll = 0
			}
			return s, nil

		case "up", "k":
			if s.canScrollContextProgression() && s.detailScroll > 0 {
				s.detailScroll--
			}
			return s, nil

		case "down", "j":
			if s.canScrollContextProgression() {
				s.detailScroll++
			}
			return s, nil

		}
	}
	return s, nil
}

func (s *ProblemDetailScreen) backNavigationMsg() tea.Msg {
	returnScreen := s.returnScreen
	if returnScreen == "" {
		returnScreen = ScreenStageDetail
	}
	msg := NavigateMsg{ScreenID: returnScreen}
	if returnScreen == ScreenStageDetail {
		msg.Stage = s.returnStage
		if msg.Stage == "" {
			msg.Stage = s.problemStage()
		}
	}
	return msg
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
		return s.handleOpenEditor(editorLaunchAttached)
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
	if len(s.scrapedExamples) > 0 {
		_, _ = appendScrapedLeetCodeExamples(s.problem, s.cfg.Language, testPath, s.scrapedExamples)
	}

	s.status = roadmap.StatusInProgress
	s.errorMsg = "Started. Files opened in editor."

	return s, tea.Batch(
		func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Started %s — files generated", s.problem.Title)}
		},
		s.openEditorCmd(stubPath, editorLaunchAttached),
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

func (s *ProblemDetailScreen) handleOpenEditor(mode editorLaunchMode) (Screen, tea.Cmd) {
	stubPath, _ := s.stubPaths()
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}

	cmd := s.openEditorCmd(stubPath, mode)
	s.errorMsg = ""
	return s, cmd
}

func (s *ProblemDetailScreen) openEditorCmd(stubPath string, mode editorLaunchMode) tea.Cmd {
	editor := s.cfg.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	effectiveMode := editorEffectiveLaunchMode(editor, mode)
	cmd := editorCommand(editor, []string{stubPath}, s.problemDir(), effectiveMode)
	if effectiveMode == editorLaunchAttached {
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to open editor: %v", err)}
			}
			return nil
		})
	}
	return func() tea.Msg {
		if err := startEditorCommand(cmd); err != nil {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to open editor: %v", err)}
		}
		return nil
	}
}

func (s *ProblemDetailScreen) handleDebug(mode editorLaunchMode) (Screen, tea.Cmd) {
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}
	editor := s.configuredEditor()
	if !isNeovimEditor(editor) {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: "Debug is only available for Neovim for now."}
		}
	}
	if s.cfg.Language != "go" {
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: "Debug is only available for Go Problems in Neovim for now."}
		}
	}
	spec, ok := s.debugProblemSpec()
	if !ok || len(spec.Examples) == 0 || spec.IsDesign {
		s.errorMsg = "No debuggable test cases are available for this Problem."
		return s, nil
	}
	s.debugPickerMode = true
	s.debugCaseIndex = 0
	s.debugLaunchMode = mode
	s.errorMsg = ""
	return s, nil
}

func (s *ProblemDetailScreen) handleDebugPickerKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	spec, ok := s.debugProblemSpec()
	if !ok || len(spec.Examples) == 0 {
		s.debugPickerMode = false
		return s, nil
	}
	switch msg.String() {
	case "esc", "backspace":
		s.debugPickerMode = false
		return s, nil
	case "q", "ctrl+c":
		return s, tea.Quit
	case "up", "k":
		if s.debugCaseIndex > 0 {
			s.debugCaseIndex--
		}
		return s, nil
	case "down", "j":
		if s.debugCaseIndex < len(spec.Examples)-1 {
			s.debugCaseIndex++
		}
		return s, nil
	case "enter":
		idx := s.debugCaseIndex
		mode := s.debugLaunchMode
		s.debugPickerMode = false
		return s, s.openDebugCmd(spec, idx, mode)
	default:
		return s, nil
	}
}

func (s *ProblemDetailScreen) debugProblemSpec() (*generator.ProblemSpec, bool) {
	spec, ok := generator.SpecForProblem(s.problem)
	if !ok {
		return nil, false
	}
	examples := append([]generator.ExampleSpec{}, spec.Examples...)
	if s.cfg.Language == "go" {
		_, testPath := s.stubPaths()
		if localExamples, err := localGoTestExamples(spec, testPath); err == nil {
			examples = appendUniqueExamples(spec, examples, localExamples)
		}
	}
	examples = appendUniqueExamples(spec, examples, s.scrapedExamples)
	copySpec := *spec
	copySpec.Examples = examples
	return &copySpec, true
}

func appendUniqueExamples(spec *generator.ProblemSpec, examples []generator.ExampleSpec, candidates []generator.ExampleSpec) []generator.ExampleSpec {
	if len(candidates) == 0 {
		return examples
	}
	existing := make(map[string]bool)
	for _, ex := range examples {
		if sig, err := exampleValueSignature(spec, ex); err == nil {
			existing[sig] = true
		}
	}
	for _, ex := range candidates {
		sig, err := exampleValueSignature(spec, ex)
		if err != nil || existing[sig] {
			continue
		}
		examples = append(examples, ex)
		existing[sig] = true
	}
	return examples
}

func (s *ProblemDetailScreen) mergeExternalExamples(examples []generator.ExampleSpec) {
	if len(examples) == 0 {
		return
	}
	spec, ok := generator.SpecForProblem(s.problem)
	if !ok {
		return
	}
	existing := make(map[string]bool)
	for _, ex := range spec.Examples {
		if sig, err := exampleValueSignature(spec, ex); err == nil {
			existing[sig] = true
		}
	}
	for _, ex := range s.scrapedExamples {
		if sig, err := exampleValueSignature(spec, ex); err == nil {
			existing[sig] = true
		}
	}
	for _, ex := range examples {
		sig, err := exampleValueSignature(spec, ex)
		if err != nil || existing[sig] {
			continue
		}
		s.scrapedExamples = append(s.scrapedExamples, ex)
		existing[sig] = true
	}
}

func (s *ProblemDetailScreen) openDebugCmd(spec *generator.ProblemSpec, exampleIndex int, mode editorLaunchMode) tea.Cmd {
	stubPath, _ := s.stubPaths()
	editor := s.configuredEditor()
	caseFile, err := s.writeDebugCase(spec, exampleIndex)
	if err != nil {
		return func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to prepare debug case: %v", err)}
		}
	}
	debugTestPath, err := writeGoDebugTestFile(s.problemDir(), spec, exampleIndex)
	if err != nil {
		return func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to prepare Go debug test: %v", err)}
		}
	}
	bootstrapPath, err := writeNeovimDebugBootstrap(s.problemDir(), spec)
	if err != nil {
		return func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to prepare Neovim debug bootstrap: %v", err)}
		}
	}
	cmd := debugEditorCommand(editor, stubPath, bootstrapPath, neovimFuncSearchCommand(spec), s.problemDir(), mode)
	cmd.Env = append(os.Environ(),
		"LEETGO_DEBUG_CASE_FILE="+caseFile,
		"LEETGO_DEBUG_TEST_FILE="+debugTestPath,
		"LEETGO_DEBUG_BOOTSTRAP="+bootstrapPath,
		fmt.Sprintf("LEETGO_DEBUG_CASE_INDEX=%d", exampleIndex),
		"LEETGO_DEBUG_PROBLEM_SLUG="+s.problem.Slug,
	)
	if mode == editorLaunchAttached {
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to open Neovim debugger: %v", err)}
			}
			return nil
		})
	}
	return func() tea.Msg {
		if err := startEditorCommand(cmd); err != nil {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to open Neovim debugger: %v", err)}
		}
		return nil
	}
}

type debugCaseFile struct {
	ProblemID int               `json:"problem_id"`
	Slug      string            `json:"slug"`
	Title     string            `json:"title"`
	Index     int               `json:"index"`
	Name      string            `json:"name"`
	Inputs    map[string]string `json:"inputs"`
	Expect    string            `json:"expect"`
}

func (s *ProblemDetailScreen) writeDebugCase(spec *generator.ProblemSpec, exampleIndex int) (string, error) {
	if exampleIndex < 0 || exampleIndex >= len(spec.Examples) {
		return "", fmt.Errorf("debug case index out of range")
	}
	ex := spec.Examples[exampleIndex]
	inputs := make(map[string]string, len(spec.Params))
	for _, p := range spec.Params {
		inputs[p.Name] = ex.Input[p.Name]
	}
	data := debugCaseFile{
		ProblemID: s.problem.ID,
		Slug:      s.problem.Slug,
		Title:     s.problem.Title,
		Index:     exampleIndex,
		Name:      debugExampleName(ex, exampleIndex),
		Inputs:    inputs,
		Expect:    ex.Expect,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	dir := s.problemDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, ".leetgo-debug.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *ProblemDetailScreen) configuredEditor() string {
	editor := s.cfg.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	return editor
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
	_, canVerify, _, reason := generator.AutomationSupport(s.problem)
	if !canVerify {
		s.errorMsg = reason + ". Submit to LeetCode directly when ready."
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
	_, canVerify, canSubmit, reason := generator.AutomationSupport(s.problem)
	if !canSubmit {
		s.errorMsg = reason
		return s, nil
	}
	if runLocalTests && !canVerify {
		s.errorMsg = reason + ". Type SUBMIT to submit without local verification."
		s.submitAnywayMode = true
		s.submitAnywayInput = ""
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
			s.submitResultMsg = s.errorMsg
			return s, nil
		}
	}

	if msg.err != nil {
		s.errorMsg = fmt.Sprintf("Submit failed: %v", msg.err)
		s.submitResultMsg = s.errorMsg
		return s, nil
	}

	s.recordSolveLog(msg)

	if msg.result.StatusCode == 10 {
		s.markAcceptedSubmission(msg.problemID, msg.duration)
		if s.errorMsg == "" {
			s.errorMsg = acceptedMessage(msg.result)
		}
		s.submitResultMsg = acceptedMessage(msg.result)
		if strings.Contains(strings.ToLower(s.errorMsg), "already claimed") {
			s.submitResultMsg += "\n" + s.errorMsg
		}
	} else {
		s.errorMsg = fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests)
		if msg.result.Error != "" {
			s.errorMsg += ": " + msg.result.Error
		}
		_, testPath := s.stubPaths()
		importMsg, err := appendFailedLeetCodeCase(s.problem, s.cfg.Language, testPath, msg.result)
		if err != nil {
			importMsg = fmt.Sprintf("Failed to add remote testcase locally: %v", err)
		}
		if ex, err := failedLeetCodeExample(s.problem, msg.result); err == nil {
			s.mergeExternalExamples([]generator.ExampleSpec{ex})
		}
		if importMsg != "" {
			s.errorMsg += "\n" + importMsg
		}
		s.submitResultMsg = s.errorMsg
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

func (s *ProblemDetailScreen) handleRegenerateFiles() (Screen, tea.Cmd) {
	if s.status == roadmap.StatusLocked {
		s.errorMsg = "Problem is locked."
		return s, nil
	}
	s.regenerateMode = true
	s.regenerateInput = ""
	s.errorMsg = ""
	return s, nil
}

func (s *ProblemDetailScreen) handleRegenerateKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		s.regenerateMode = false
		s.regenerateInput = ""
		s.errorMsg = "Regenerate cancelled."
		return s, nil
	case tea.KeyEnter:
		if s.regenerateInput == "REGENERATE" {
			s.regenerateMode = false
			s.regenerateInput = ""
			return s.regenerateFiles()
		}
		return s, nil
	case tea.KeyBackspace:
		if len(s.regenerateInput) > 0 {
			s.regenerateInput = s.regenerateInput[:len(s.regenerateInput)-1]
		}
		return s, nil
	case tea.KeyRunes:
		s.regenerateInput += string(msg.Runes)
		return s, nil
	default:
		return s, nil
	}
}

func (s *ProblemDetailScreen) regenerateFiles() (Screen, tea.Cmd) {
	stubPath, testPath, err := s.workspace.Generate(s.problem, generator.Language(s.cfg.Language))
	if err != nil {
		s.errorMsg = fmt.Sprintf("Failed to regenerate files: %v", err)
		return s, nil
	}
	if err := s.writeManifest(stubPath, testPath); err != nil {
		s.errorMsg = fmt.Sprintf("Failed to write manifest: %v", err)
		return s, nil
	}
	if len(s.scrapedExamples) > 0 {
		if msg, err := appendScrapedLeetCodeExamples(s.problem, s.cfg.Language, testPath, s.scrapedExamples); err != nil {
			s.errorMsg = fmt.Sprintf("Regenerated files, but failed to add scraped tests: %v", err)
			return s, nil
		} else if msg != "" {
			s.errorMsg = "Regenerated files. " + msg
			return s, nil
		}
	}
	s.errorMsg = "Regenerated Stub and TestSuite files."
	return s, nil
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
		s.testResultMsg = "Local TestSuite failed.\n\n" + msg.output
		return s, nil
	}

	if s.status == roadmap.StatusSolved {
		return s.handleReviewTestPass(ctx, msg)
	}

	if s.status == roadmap.StatusVerified {
		s.testResultMsg = msg.output + "\n\nTests passed. Reward already claimed."
		return s, nil
	}

	alreadyClaimed, err := s.db.HasRewardEvent(ctx, msg.problemID, "verify")
	if err != nil {
		s.errorMsg = fmt.Sprintf("Failed to check verify reward: %v", err)
		return s, nil
	}

	if alreadyClaimed {
		s.testResultMsg = msg.output + "\n\nVerify XP already claimed for this problem."
		return s, nil
	}

	xp := store.XPForDifficulty(msg.difficulty) * 70 / 100
	if xp > 0 {
		if err := s.db.AddXP(ctx, xp); err != nil {
			s.errorMsg = fmt.Sprintf("Failed to add verify XP: %v", err)
			return s, nil
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
	s.testResultMsg = fmt.Sprintf("%s\n\n+%d XP (verify: %d%%) Tests passed — Problem Verified", msg.output, xp, 70)
	return s, nil
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
		s.testResultMsg = msg.output + "\n\nTests passed."
		return s, nil
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
		s.testResultMsg = msg.output + "\n\nTests passed. Review already completed."
		return s, nil
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
	s.testResultMsg = result
	return s, nil
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
		cmd = exec.Command("go", "test", ".", "-timeout", "5s")
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
		cmd = exec.Command("go", "test", ".", "-timeout", "5s")
	}
	cmd.Dir = dir
	return cmd
}

func (s *ProblemDetailScreen) View() string {
	if s.submitAnywayMode {
		return s.renderSubmitAnywayConfirmation()
	}
	if s.regenerateMode {
		return s.renderRegenerateConfirmation()
	}
	if s.testResultMsg != "" {
		return s.renderTestResult()
	}
	if s.submitResultMsg != "" {
		return s.renderSubmitResult()
	}
	if s.submitting {
		return s.renderSubmitProgress()
	}
	if s.debugPickerMode {
		return s.renderDebugPicker()
	}
	if s.manualSolveMode {
		return s.renderManualSolveConfirmation()
	}

	screenLabel, briefLabel, filesLabel := themeProblemLabels(s.theme)
	header := s.renderProblemDetailHeader(screenLabel)
	if s.problem.ProblemTimeEstimate != "" {
		header += "\n" + s.theme.Subtitle.Render("Time estimate: "+s.problem.ProblemTimeEstimate)
	}

	statusLabel := s.renderStatusDetailBody()
	transientStatusLines := s.renderTransientStatusLines()
	if len(transientStatusLines) > 0 {
		statusLabel += "\n\n" + strings.Join(transientStatusLines, "\n")
	}
	var contextLines []string

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

	var workspaceLines []string
	if s.status != roadmap.StatusLocked {
		stubPath, testPath := s.stubPaths()
		workspaceLines = append(workspaceLines,
			fmt.Sprintf("Stub: %s", stubPath),
			fmt.Sprintf("Test: %s", testPath),
		)
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
		rightSections = append(rightSections, renderProblemDetailPanel(s.theme, "Practice Log", strings.Join(logLines, "\n"), false, s.problemDetailBodyWidth()))
	}

	footer := s.renderFooter()
	bodyWidth := s.problemDetailBodyWidth()
	bodySections := []string{
		renderProblemDetailPanel(s.theme, "Status", statusLabel, false, minProblemDetailInt(58, bodyWidth)),
		s.renderContextProgressionRow(briefLabel, contextLines, progressionLines, bodyWidth),
	}
	if len(workspaceLines) > 0 {
		bodySections = append(bodySections, renderProblemDetailPanel(s.theme, filesLabel, strings.Join(workspaceLines, "\n"), false, bodyWidth))
	}
	if len(rightSections) > 0 {
		bodySections = append(bodySections, rightSections...)
	}
	body := strings.Join(bodySections, "\n\n")
	return renderScreenShell(s.theme, s.width, s.height, header, body, footer)
}

func (s *ProblemDetailScreen) renderSubmitResult() string {
	width := responsiveShellContentWidth(s.width)
	if width <= 0 {
		width = 86
	}
	panelWidth := minProblemDetailInt(width-4, 86)
	if panelWidth < 40 {
		panelWidth = 40
	}

	title := "Submission Result"
	lower := strings.ToLower(s.submitResultMsg)
	if strings.Contains(lower, "accepted") {
		title = "Submission Accepted"
	} else if strings.Contains(lower, "failed") {
		title = "Submission Failed"
	} else if strings.Contains(lower, "wrong answer") || strings.Contains(lower, "runtime error") || strings.Contains(lower, "time limit") {
		title = "Submission Rejected"
	}

	body := renderProblemDetailPreformattedPanel(s.theme, title, strings.TrimSpace(s.submitResultMsg), true, panelWidth)
	footerItems := []string{
		s.theme.Key.Render("esc") + " details",
		s.theme.Key.Render("q") + " quit",
	}
	if title != "Submission Accepted" && len(s.scrapedExamples) > 0 {
		footerItems = append([]string{s.theme.Key.Render("d") + " debug", s.theme.Key.Render("D") + " debug window"}, footerItems...)
	}
	footer := s.theme.Footer.PaddingTop(1).Render(strings.Join(footerItems, "  "))
	content := body + "\n\n" + footer
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *ProblemDetailScreen) renderSubmitProgress() string {
	width := responsiveShellContentWidth(s.width)
	if width <= 0 {
		width = 86
	}
	panelWidth := minProblemDetailInt(width-4, 86)
	if panelWidth < 40 {
		panelWidth = 40
	}

	status := strings.TrimSpace(s.errorMsg)
	if status == "" {
		status = "Submitting to LeetCode..."
	}
	if s.allowsSpinnerMotion() {
		frame := spinnerFrames[s.spinnerFrame%len(spinnerFrames)]
		status = frame + " " + status
	}

	lines := []string{
		s.theme.Spinner.Render(status),
		"",
		lipgloss.NewStyle().Foreground(s.theme.Muted).Render("Keep this screen open while Leetgo runs local checks and submits your current Stub."),
	}
	body := renderProblemDetailPanel(s.theme, "Submitting", strings.Join(lines, "\n"), true, panelWidth)
	footer := s.theme.Footer.PaddingTop(1).Render(strings.Join([]string{
		s.theme.Key.Render("q") + " quit",
	}, "  "))
	content := body + "\n\n" + footer
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *ProblemDetailScreen) renderTransientStatusLines() []string {
	var lines []string
	if s.submitting {
		if s.allowsSpinnerMotion() {
			frame := spinnerFrames[s.spinnerFrame%len(spinnerFrames)]
			lines = append(lines, s.theme.Spinner.Render(frame+" Submitting to LeetCode..."))
		} else {
			lines = append(lines, s.theme.Spinner.Render("Submitting to LeetCode..."))
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
		lines = append(lines, style.Render(s.errorMsg))
	}
	return lines
}

func (s *ProblemDetailScreen) renderTestResult() string {
	width := responsiveShellContentWidth(s.width)
	if width <= 0 {
		width = 86
	}
	panelWidth := minProblemDetailInt(width-4, 86)
	if panelWidth < 40 {
		panelWidth = 40
	}
	title := "TestSuite Results"
	lower := strings.ToLower(s.testResultMsg)
	if strings.Contains(lower, "failed") {
		title = "TestSuite Failed"
	} else if strings.Contains(lower, "passed") {
		title = "TestSuite Passed"
	}
	body := renderProblemDetailPreformattedPanel(s.theme, title, strings.TrimSpace(s.testResultMsg), true, panelWidth)
	footer := s.theme.Footer.PaddingTop(1).Render(strings.Join([]string{
		s.theme.Key.Render("esc") + " details",
		s.theme.Key.Render("q") + " quit",
	}, "  "))
	content := body + "\n\n" + footer
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *ProblemDetailScreen) renderProblemDetailHeader(screenLabel string) string {
	title := s.theme.Title.Render(fmt.Sprintf("%s: #%d %s", screenLabel, s.problem.ID, s.problem.Title))
	subtitle := strings.Join([]string{
		s.theme.Subtitle.Render("Difficulty: ") + renderProblemDetailDifficultyTag(s.theme, s.problem.Difficulty),
		s.theme.Subtitle.Render("Category: " + string(s.problem.Category)),
		s.theme.Subtitle.Render("Stage: " + s.stageName()),
	}, s.theme.Subtitle.Render("  •  "))
	return title + "\n" + subtitle
}

func (s *ProblemDetailScreen) renderContextProgressionRow(briefLabel string, contextLines, progressionLines []string, bodyWidth int) string {
	activeHeight := s.activeDetailPanelBodyLines()
	if len(progressionLines) == 0 {
		return s.renderContextPanel(briefLabel, contextLines, bodyWidth, activeHeight, s.detailScroll)
	}

	progressionBody := strings.Join(progressionLines, "\n")
	if bodyWidth >= 104 {
		progressionWidth := 40
		contextWidth := bodyWidth - progressionWidth - 2
		return lipgloss.JoinHorizontal(lipgloss.Top,
			s.renderContextPanel(briefLabel, contextLines, contextWidth, activeHeight, s.detailScroll),
			"  ",
			renderProblemDetailPanel(s.theme, "Progression", progressionBody, false, progressionWidth),
		)
	}

	if s.detailTab == problemDetailTabProgression {
		s.detailScroll = clampProblemDetailScroll(progressionBody, bodyWidth, activeHeight, s.detailScroll)
		return renderProblemDetailPanelScrollable(s.theme, "Progression", progressionBody, false, bodyWidth, activeHeight, s.detailScroll)
	}
	return s.renderContextPanel(briefLabel, contextLines, bodyWidth, activeHeight, s.detailScroll)
}

func (s *ProblemDetailScreen) canCycleContextProgression() bool {
	return s.problemDetailBodyWidth() < 104 && s.hasProgressionPanel()
}

func (s *ProblemDetailScreen) canScrollContextProgression() bool {
	return s.problemDetailBodyWidth() < 104 || s.problemDescription != ""
}

func (s *ProblemDetailScreen) hasProgressionPanel() bool {
	return len(s.directUnlocks()) > 0 || len(s.indirectUnlocks(2)) > 0
}

func (s *ProblemDetailScreen) renderContextPanel(briefLabel string, contextLines []string, width, maxBodyLines, scroll int) string {
	parts := []string{s.renderProblemBriefBlock(briefLabel, panelBodyWidth(width))}
	if len(contextLines) > 0 {
		parts = append(parts, strings.Join(contextLines, "\n\n"))
	}
	body := strings.Join(parts, "\n\n")
	s.detailScroll = clampProblemDetailScroll(body, width, maxBodyLines, scroll)
	return renderProblemDetailPanelScrollable(s.theme, "Context", body, false, width, maxBodyLines, s.detailScroll)
}

func (s *ProblemDetailScreen) activeDetailPanelBodyLines() int {
	if (s.problemDetailBodyWidth() >= 104 && s.problemDescription == "") || s.height <= 0 {
		return 0
	}
	lines := s.height - 22
	if lines < 6 {
		return 6
	}
	if lines > 14 {
		return 14
	}
	return lines
}

func (s *ProblemDetailScreen) problemDetailBodyWidth() int {
	width := responsiveShellContentWidth(s.width)
	if width <= 0 {
		return 96
	}
	width -= 4
	if width < 20 {
		return 20
	}
	return width
}

func renderProblemDetailPanel(theme *Theme, title, body string, focused bool, width int) string {
	return renderProblemDetailPanelScrollable(theme, title, body, focused, width, 0, 0)
}

func renderProblemDetailPreformattedPanel(theme *Theme, title, body string, focused bool, width int) string {
	innerWidth := panelBodyWidth(width)
	lines := problemDetailPreformattedLines(body, innerWidth)
	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().Width(innerWidth).Render(line)
	}
	return renderThemedPanel(theme, title, strings.Join(lines, "\n"), focused)
}

func renderProblemDetailPanelScrollable(theme *Theme, title, body string, focused bool, width, maxBodyLines, scroll int) string {
	innerWidth := panelBodyWidth(width)
	lines := problemDetailPanelLines(body, innerWidth)
	if maxBodyLines > 0 && len(lines) > maxBodyLines {
		visibleLines := maxBodyLines - 1
		if visibleLines < 1 {
			visibleLines = 1
		}
		maxScroll := len(lines) - visibleLines
		if scroll > maxScroll {
			scroll = maxScroll
		}
		if scroll < 0 {
			scroll = 0
		}
		lines = append(lines[scroll:scroll+visibleLines], problemDetailScrollIndicator(theme, scroll, maxScroll))
	}

	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().Width(innerWidth).Render(line)
	}
	return renderThemedPanel(theme, title, strings.Join(lines, "\n"), focused)
}

func clampProblemDetailScroll(body string, width, maxBodyLines, scroll int) int {
	if maxBodyLines <= 0 {
		return 0
	}
	visibleLines := maxBodyLines - 1
	if visibleLines < 1 {
		visibleLines = 1
	}
	lineCount := len(problemDetailPanelLines(body, panelBodyWidth(width)))
	maxScroll := lineCount - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	if scroll < 0 {
		return 0
	}
	return scroll
}

func problemDetailPanelLines(body string, width int) []string {
	var lines []string
	for _, line := range strings.Split(body, "\n") {
		wrapped := wrapProblemDetailLine(line, width)
		lines = append(lines, strings.Split(wrapped, "\n")...)
	}
	return lines
}

func problemDetailPreformattedLines(body string, width int) []string {
	var lines []string
	for _, line := range strings.Split(body, "\n") {
		if width <= 0 || lipgloss.Width(line) <= width {
			lines = append(lines, line)
			continue
		}
		for lipgloss.Width(line) > width {
			cut := problemDetailWrapCut(line, width)
			lines = append(lines, line[:cut])
			line = line[cut:]
		}
		lines = append(lines, line)
	}
	return lines
}

func problemDetailScrollIndicator(theme *Theme, scroll, maxScroll int) string {
	position := fmt.Sprintf("%d/%d", scroll+1, maxScroll+1)
	symbol := "↓"
	if scroll > 0 && scroll < maxScroll {
		symbol = "↑↓"
	} else if scroll >= maxScroll {
		symbol = "↑"
	}
	return lipgloss.NewStyle().Foreground(theme.Muted).Render(symbol + " scroll " + position)
}

func wrapProblemDetailLine(line string, width int) string {
	if width <= 0 || lipgloss.Width(line) <= width {
		return line
	}

	wrapped := wrapText(line, width)
	if strings.Contains(wrapped, "\x1b") {
		return wrapped
	}

	var lines []string
	for _, part := range strings.Split(wrapped, "\n") {
		for lipgloss.Width(part) > width {
			cut := problemDetailWrapCut(part, width)
			lines = append(lines, part[:cut])
			part = part[cut:]
		}
		lines = append(lines, part)
	}
	return strings.Join(lines, "\n")
}

func problemDetailWrapCut(text string, width int) int {
	if width <= 0 {
		return len(text)
	}
	currentWidth := 0
	for idx, r := range text {
		charWidth := lipgloss.Width(string(r))
		if currentWidth+charWidth > width {
			if idx == 0 {
				return len(string(r))
			}
			return idx
		}
		currentWidth += charWidth
	}
	return len(text)
}

func minProblemDetailInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func renderProblemDetailDifficultyTag(theme *Theme, difficulty roadmap.Difficulty) string {
	style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch difficulty {
	case roadmap.DifficultyEasy:
		return style.Foreground(theme.Success).Render("easy")
	case roadmap.DifficultyMedium:
		return style.Foreground(theme.Warning).Render("medium")
	case roadmap.DifficultyHard:
		return style.Foreground(theme.Danger).Render("hard")
	default:
		return style.Foreground(theme.Muted).Render(string(difficulty))
	}
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

func (s *ProblemDetailScreen) renderRegenerateConfirmation() string {
	var lines []string
	lines = append(lines, s.theme.Title.Render("Regenerate Problem Files"))
	lines = append(lines, fmt.Sprintf("Problem: #%d %s", s.problem.ID, s.problem.Title))
	lines = append(lines, "")
	warning := strings.Join([]string{
		lipgloss.NewStyle().Foreground(s.theme.Danger).Bold(true).Render("This overwrites your current Stub and TestSuite files."),
		"Use this when you want a fresh generated solution file and fresh local tests.",
		"Scraped LeetCode examples will be added again when available.",
	}, "\n")
	lines = append(lines, renderThemedPanel(s.theme, "Confirm Regenerate", warning, false))
	lines = append(lines, "")
	lines = append(lines, "Type REGENERATE to confirm:")
	lines = append(lines, s.theme.FocusedPanel.Render(s.regenerateInput))
	lines = append(lines, "")
	lines = append(lines, s.theme.Footer.Render(s.theme.Key.Render("enter")+" confirm  "+s.theme.Key.Render("esc")+" cancel"))
	return strings.Join(lines, "\n")
}

func (s *ProblemDetailScreen) renderDebugPicker() string {
	spec, ok := s.debugProblemSpec()
	if !ok || len(spec.Examples) == 0 {
		return renderThemedPanel(s.theme, "Debug", "No debuggable test cases are available.", true)
	}
	var lines []string
	lines = append(lines, s.theme.Title.Render("Debug with Neovim DAP"))
	lines = append(lines, fmt.Sprintf("Problem: #%d %s", s.problem.ID, s.problem.Title))
	lines = append(lines, "")
	lines = append(lines, "Choose a test case. Leetgo writes a focused Go debug test and starts Neovim DAP directly.")
	lines = append(lines, "")
	muted := lipgloss.NewStyle().Foreground(s.theme.Muted)
	selected := lipgloss.NewStyle().Foreground(s.theme.SelectionFg).Background(s.theme.SelectionBg).Bold(true)
	for i, ex := range spec.Examples {
		label := fmt.Sprintf("%d. %s", i+1, debugExampleName(ex, i))
		preview := debugExamplePreview(spec, ex)
		row := label + "  " + muted.Render(preview)
		if i == s.debugCaseIndex {
			row = selected.Render("> " + row)
		} else {
			row = "  " + row
		}
		lines = append(lines, row)
	}
	lines = append(lines, "")
	lines = append(lines, s.theme.Footer.Render(s.theme.Key.Render("j/k")+" choose  "+s.theme.Key.Render("enter")+" debug  "+s.theme.Key.Render("esc")+" cancel"))
	return strings.Join(lines, "\n")
}

func debugExampleName(ex generator.ExampleSpec, index int) string {
	if name := strings.TrimSpace(ex.Input["_name"]); name != "" {
		return name
	}
	return fmt.Sprintf("case %d", index+1)
}

func debugExamplePreview(spec *generator.ProblemSpec, ex generator.ExampleSpec) string {
	parts := make([]string, 0, len(spec.Params)+1)
	for _, p := range spec.Params {
		parts = append(parts, fmt.Sprintf("%s=%s", p.Name, ex.Input[p.Name]))
	}
	parts = append(parts, "expect="+ex.Expect)
	preview := strings.Join(parts, ", ")
	if len(preview) > 96 {
		return preview[:93] + "..."
	}
	return preview
}

func (s *ProblemDetailScreen) renderProblemBriefBlock(label string, width int) string {
	if s.problemDescription != "" {
		body := s.theme.Subtitle.Render("Source: LeetCode") + "\n\n" + s.problemDescription
		return renderProblemDetailPanel(s.theme, "Problem Description", body, false, width)
	}
	if s.problem.Summary == "" && s.problem.PracticeFocus == "" {
		return renderProblemDetailPanel(s.theme, label, "No brief available yet.", false, width)
	}
	var lines []string
	if s.descriptionStatus != "" {
		lines = append(lines, s.theme.Subtitle.Render(s.descriptionStatus), "")
	}
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
	return renderProblemDetailPanel(s.theme, label, strings.Join(lines, "\n"), false, width)
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
	return renderThemedPanel(s.theme, "Status", s.renderStatusDetailBody(), false)
}

func (s *ProblemDetailScreen) renderStatusDetailBody() string {
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
		return label + "\nPractice Log and unlock impact are now the priority."
	case roadmap.StatusVerified:
		return renderStatusPill(s.theme, symbols, roadmap.StatusVerified) + "\nPrimary: Submit to LeetCode. Secondary: Manual Solve if needed."
	case roadmap.StatusInProgress:
		return renderStatusPill(s.theme, symbols, roadmap.StatusInProgress) + "\nKeep file paths and TestSuite actions close."
	case roadmap.StatusAvailable:
		return renderStatusPill(s.theme, symbols, roadmap.StatusAvailable) + "\nRead the Problem Brief, then start the Problem."
	default:
		return renderStatusPill(s.theme, symbols, roadmap.StatusLocked) + "\nClear the blocker before starting."
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
		s.theme.Key.Render("e") + " edit",
		s.theme.Key.Render("E") + " open window",
		s.theme.Key.Render("d") + " debug",
		s.theme.Key.Render("D") + " debug window",
		s.theme.Key.Render("R") + " regenerate",
		s.theme.Key.Render("x") + " test",
		s.theme.Key.Render("s") + " submit",
		s.theme.Key.Render("m") + " solve",
	}
	if s.canCycleContextProgression() {
		items = append(items, s.theme.Key.Render("left/right")+" context/progression")
	}
	if s.canScrollContextProgression() {
		items = append(items, s.theme.Key.Render("up/down")+" scroll")
	}
	items = append(items,
		s.theme.Key.Render("esc")+" back",
		s.theme.Key.Render("q")+" quit",
	)

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
