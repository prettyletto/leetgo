package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/gamification"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
	"github.com/prettyletto/leetgo/internal/workspace"
)

type problemItem struct {
	problem *roadmap.Problem
	status  roadmap.Status
	theme   *Theme
}

func (i problemItem) Title() string {
	return fmt.Sprintf("%d. %s", i.problem.ID, i.problem.Title)
}

func (i problemItem) Description() string {
	status := renderStatus(i.status, i.theme)
	diff := renderDifficulty(i.problem.Difficulty, i.theme)
	cat := lipgloss.NewStyle().Width(20).Foreground(i.theme.Muted).Render(string(i.problem.Category))
	return fmt.Sprintf("%s %s %s", status, diff, cat)
}

func (i problemItem) FilterValue() string {
	return i.problem.Title
}

func renderStatus(s roadmap.Status, theme *Theme) string {
	statusStyle := lipgloss.NewStyle().Width(12)
	if theme == nil {
		theme = &Theme{
			PrimaryAccent:   lipgloss.Color("205"),
			SecondaryAccent: lipgloss.Color("219"),
			Muted:           lipgloss.Color("245"),
			Success:         lipgloss.Color("82"),
			Warning:         lipgloss.Color("214"),
			Danger:          lipgloss.Color("196"),
		}
	}
	switch s {
	case roadmap.StatusLocked:
		return statusStyle.Foreground(theme.Muted).Render("[LOCKED]")
	case roadmap.StatusAvailable:
		return statusStyle.Foreground(theme.Warning).Render("[READY]")
	case roadmap.StatusInProgress:
		return statusStyle.Foreground(theme.PrimaryAccent).Render("[ACTIVE]")
	case roadmap.StatusSolved:
		return statusStyle.Foreground(theme.Success).Render("[SOLVED]")
	case roadmap.StatusVerified:
		return statusStyle.Foreground(theme.Warning).Render("[VERIFIED]")
	default:
		return statusStyle.Render("[???]")
	}
}

func renderDifficulty(d roadmap.Difficulty, _ *Theme) string {
	diffStyle := lipgloss.NewStyle().Width(8)
	switch d {
	case roadmap.DifficultyEasy:
		return diffStyle.Render("Easy")
	case roadmap.DifficultyMedium:
		return diffStyle.Render("Medium")
	case roadmap.DifficultyHard:
		return diffStyle.Render("Hard")
	default:
		return diffStyle.Render("???")
	}
}

type viewMode int

const (
	viewList viewMode = iota
	viewHeatmap
)

type tickMsg time.Time
type submitResultMsg struct {
	problemID int
	slug      string
	language  string
	result    *leetcode.SubmissionResult
	err       error
	duration  time.Duration

	localTestRan      bool
	localTestPassed   bool
	localTestOutput   string
	localTestDuration time.Duration
}

type testRunResultMsg struct {
	problemID  int
	difficulty roadmap.Difficulty
	output     string
	passed     bool
	duration   time.Duration
	alreadyRun bool
}

type Model struct {
	list          list.Model
	heatmapView   *views.HeatmapView
	statsBar      *views.StatsBar
	notifications *views.NotificationManager
	gamification  *gamification.Engine
	leetcode      *leetcode.Client
	viewMode      viewMode
	roadmap       *roadmap.Roadmap
	graph         *roadmap.Graph
	allItems      []list.Item
	stageFilter   int
	store         store.Store
	workspace     *workspace.Manager
	config        *config.Config
	theme         *Theme
	quitting      bool
	submitting    bool
}

func NewModel(cfg *config.Config, db store.Store) (*Model, error) {
	rm, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		return nil, fmt.Errorf("load catalog: %w", err)
	}
	graph := rm.Graph

	gen := generator.New()
	ws := workspace.New(cfg.Workspace, gen)

	progress, err := db.GetAllProgress(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}

	solved := make(map[int]bool)
	for id, status := range progress {
		if status == roadmap.StatusSolved {
			solved[id] = true
		}
	}

	sorted, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("sort problems: %w", err)
	}

	theme, _ := LookupTheme(cfg.Theme)

	items := make([]list.Item, len(sorted))
	for i, p := range sorted {
		status := roadmap.StatusLocked
		if s, ok := progress[p.ID]; ok {
			status = s
		} else if graph.IsUnlocked(p.ID, solved) {
			status = roadmap.StatusAvailable
		}
		items[i] = problemItem{problem: p, status: status, theme: theme}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 24)
	l.Title = rm.Title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	stats, err := db.GetStats(context.Background())
	if err != nil {
		stats = &store.Stats{}
	}
	stats.Total = len(sorted)
	statsBar := views.NewStatsBar(stats)
	statsBar.SetPalette(viewPalette(theme))

	notifications := views.NewNotificationManager()
	notifications.SetPalette(viewPalette(theme))
	gamificationEngine := gamification.NewEngine(db, graph)

	streakDays, err := db.GetStreakDays(context.Background())
	if err != nil {
		streakDays = nil
	}
	heatmapView := views.NewHeatmapView(streakDays)
	heatmapView.SetPalette(viewPalette(theme))

	lcClient, err := leetcode.NewClient()
	if err != nil {
		lcClient = nil
	}

	return &Model{
		list:          l,
		heatmapView:   heatmapView,
		statsBar:      statsBar,
		notifications: notifications,
		gamification:  gamificationEngine,
		leetcode:      lcClient,
		viewMode:      viewList,
		roadmap:       rm,
		graph:         graph,
		allItems:      items,
		stageFilter:   -1,
		store:         db,
		workspace:     ws,
		config:        cfg,
		theme:         theme,
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tickCmd()

	case submitResultMsg:
		m.submitting = false
		if msg.err != nil {
			m.notifications.Add(fmt.Sprintf("Submit failed: %v", msg.err))
		} else if msg.result.StatusCode == 10 {
			m.recordSolveLog(msg)
			m.markAcceptedSubmission(msg.problemID)
			m.notifications.Add(acceptedMessage(msg.result))
		} else {
			m.recordSolveLog(msg)
			message := fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests)
			if msg.result.Error != "" {
				message += ": " + msg.result.Error
			}
			m.notifications.Add(message)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-10)
		m.heatmapView.SetSize(msg.Width, msg.Height)
		m.statsBar.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "h":
			if m.viewMode == viewHeatmap {
				m.viewMode = viewList
			} else {
				m.viewMode = viewHeatmap
			}
			return m, nil

		case "tab":
			if m.viewMode == viewList {
				m.cycleStageFilter(1)
			}
			return m, nil

		case "shift+tab":
			if m.viewMode == viewList {
				m.cycleStageFilter(-1)
			}
			return m, nil

		case "0":
			if m.viewMode == viewList {
				m.stageFilter = -1
				m.applyStageFilter()
			}
			return m, nil

		case "s":
			if m.viewMode == viewList && !m.submitting {
				return m.handleSubmit()
			}
			return m, nil

		case "m":
			if m.viewMode == viewList {
				return m.handleMarkSolved()
			}
			return m, nil

		case "enter":
			if m.viewMode == viewList {
				return m.handleSelect()
			}
			return m, nil

		}
	}

	if m.viewMode == viewList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var view string
	switch m.viewMode {
	case viewHeatmap:
		view = m.heatmapView.Render()
	default:
		view = m.list.View() + "\n" + m.renderProblemDetails()
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.PrimaryAccent).
		MarginBottom(1)

	header := titleStyle.Render("Leetgo · "+m.roadmap.Title) + "\n" + m.statsBar.Render()
	if m.viewMode == viewList {
		header += "\n" + m.renderStageFilter()
	}
	footer := m.renderKeytips()

	notification := m.notifications.Render()
	if notification != "" {
		notification = "\n" + notification
	}

	return header + "\n" + view + "\n" + footer + notification
}

func (m *Model) renderProblemDetails() string {
	item, ok := m.list.SelectedItem().(problemItem)
	if !ok {
		return ""
	}
	p := item.problem
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(m.theme.SecondaryAccent).Render(fmt.Sprintf("#%d %s", p.ID, p.Title)),
		fmt.Sprintf("Status: %s  Difficulty: %s  Stage: %s  Category: %s", string(item.status), string(p.Difficulty), m.stageTitle(p.Stage), string(p.Category)),
	}
	if len(p.Prerequisites) == 0 {
		lines = append(lines, "Prerequisites: none")
	} else {
		lines = append(lines, "Prerequisites: "+strings.Join(m.prerequisiteLabels(p.Prerequisites), ", "))
	}
	if item.status == roadmap.StatusLocked {
		missing := m.missingPrerequisiteLabels(p)
		if len(missing) > 0 {
			lines = append(lines, "Blocked by: "+strings.Join(missing, ", "))
		}
	}
	if item.status == roadmap.StatusAvailable {
		lines = append(lines, "Next: press enter to Start and generate the Stub/TestSuite.")
	}
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1).
		MarginTop(1)
	return detailStyle.Width(max(40, m.list.Width()-4)).Render(strings.Join(lines, "\n"))
}

func (m *Model) stageTitle(stageID string) string {
	if stageID == "" {
		return "uncategorized"
	}
	for _, stage := range m.roadmap.Stages {
		if stage.ID == stageID {
			return stage.Title
		}
	}
	return stageID
}

func (m *Model) prerequisiteLabels(ids []int) []string {
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.graph.Problems[id]; ok {
			labels = append(labels, fmt.Sprintf("#%d %s", p.ID, p.Title))
		} else {
			labels = append(labels, fmt.Sprintf("#%d", id))
		}
	}
	return labels
}

func (m *Model) missingPrerequisiteLabels(p *roadmap.Problem) []string {
	progress, err := m.store.GetAllProgress(context.Background())
	if err != nil {
		return nil
	}
	var missing []string
	for _, id := range p.Prerequisites {
		if progress[id] == roadmap.StatusSolved {
			continue
		}
		if prereq, ok := m.graph.Problems[id]; ok {
			missing = append(missing, fmt.Sprintf("#%d %s", prereq.ID, prereq.Title))
		} else {
			missing = append(missing, fmt.Sprintf("#%d", id))
		}
	}
	return missing
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) renderKeytips() string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.SecondaryAccent)
	footerStyle := lipgloss.NewStyle().Foreground(m.theme.Muted).PaddingTop(1)

	items := []string{
		keyStyle.Render("enter") + " start",
		keyStyle.Render("m") + " solve",
		keyStyle.Render("s") + " submit",
		keyStyle.Render("h") + " heatmap",
		keyStyle.Render("tab") + " stage",
		keyStyle.Render("0") + " all",
		keyStyle.Render("/") + " filter",
		keyStyle.Render("q") + " quit",
	}
	if m.viewMode == viewHeatmap {
		items = []string{
			keyStyle.Render("h") + " list",
			keyStyle.Render("q") + " quit",
		}
	}
	return footerStyle.Render(strings.Join(items, "  "))
}

func (m *Model) renderStageFilter() string {
	footerStyle := lipgloss.NewStyle().Foreground(m.theme.Muted).PaddingTop(1)
	if m.stageFilter < 0 || m.stageFilter >= len(m.roadmap.Stages) {
		return footerStyle.Render("Stage: All")
	}
	stage := m.roadmap.Stages[m.stageFilter]
	return footerStyle.Render(fmt.Sprintf("Stage: %s", stage.Title))
}

func (m *Model) cycleStageFilter(delta int) {
	if len(m.roadmap.Stages) == 0 {
		return
	}
	if m.stageFilter == -1 && delta < 0 {
		m.stageFilter = len(m.roadmap.Stages) - 1
	} else {
		m.stageFilter += delta
	}
	if m.stageFilter >= len(m.roadmap.Stages) {
		m.stageFilter = -1
	}
	if m.stageFilter < -1 {
		m.stageFilter = len(m.roadmap.Stages) - 1
	}
	m.applyStageFilter()
}

func (m *Model) applyStageFilter() {
	if m.stageFilter < 0 || m.stageFilter >= len(m.roadmap.Stages) {
		m.list.SetItems(m.allItems)
		return
	}
	stageID := m.roadmap.Stages[m.stageFilter].ID
	items := make([]list.Item, 0, len(m.allItems))
	for _, item := range m.allItems {
		problem, ok := item.(problemItem)
		if !ok {
			continue
		}
		if problem.problem.Stage == stageID {
			items = append(items, item)
		}
	}
	m.list.SetItems(items)
}

func (m *Model) handleSelect() (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(problemItem)
	if !ok {
		return m, nil
	}

	if item.status == roadmap.StatusLocked {
		m.notifications.Add("Problem is locked — solve prerequisites first.")
		return m, nil
	}

	ctx := context.Background()

	if err := workspace.EnsureManifestWritable(m.workspace.ProblemDir(item.problem), item.problem.ID); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to write manifest: %v", err))
		return m, nil
	}

	if item.status == roadmap.StatusAvailable {
		if err := m.store.SetProgress(ctx, item.problem.ID, roadmap.StatusInProgress); err != nil {
			m.notifications.Add(fmt.Sprintf("Failed to update progress: %v", err))
			return m, nil
		}
	}

	stubPath, testPath, err := m.workspace.Generate(item.problem, generator.Language(m.config.Language))
	if err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to generate files: %v", err))
		return m, nil
	}

	if err := m.writeManifest(item.problem, stubPath, testPath); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to write manifest: %v", err))
		return m, nil
	}

	m.notifications.Add(fmt.Sprintf("Started %s — files generated", item.problem.Title))
	m.openEditor(stubPath)

	return m, nil
}

func (m *Model) writeManifest(problem *roadmap.Problem, stubPath, testPath string) error {
	dir := m.workspace.ProblemDir(problem)
	stage := problem.Stage
	if stage == "" {
		stage = string(problem.Category)
	}
	manifest := &workspace.Manifest{
		ProblemID:     problem.ID,
		Slug:          problem.Slug,
		Roadmap:       m.roadmap.ID,
		Stage:         stage,
		Language:      m.config.Language,
		StubPath:      filepath.Base(stubPath),
		TestsuitePath: filepath.Base(testPath),
	}
	return workspace.WriteManifest(dir, manifest)
}

func (m *Model) handleMarkSolved() (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(problemItem)
	if !ok {
		return m, nil
	}

	if item.status == roadmap.StatusSolved {
		m.notifications.Add("Already solved!")
		return m, nil
	}

	if item.status == roadmap.StatusLocked {
		m.notifications.Add("Problem is locked.")
		return m, nil
	}

	ctx := context.Background()

	if err := m.store.SetProgress(ctx, item.problem.ID, roadmap.StatusSolved); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to update progress: %v", err))
		return m, nil
	}

	if err := m.store.RecordSolveProvenance(ctx, &store.SolveProvenance{
		ProblemID: item.problem.ID,
		Kind:      "manual",
	}); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to record solve provenance: %v", err))
		return m, nil
	}

	if err := m.store.UpdateStreak(ctx); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to update streak: %v", err))
		return m, nil
	}

	m.notifications.Add(fmt.Sprintf("Marked %s solved — manual solve awards no XP", item.problem.Title))

	unlocked, err := m.gamification.OnProblemSolved(ctx, item.problem.ID)
	if err == nil {
		for _, id := range unlocked {
			if a, ok := gamification.Achievements[id]; ok {
				m.notifications.AddAchievement(a)
				if err := m.store.UnlockAchievement(ctx, id); err != nil {
					m.notifications.Add(fmt.Sprintf("Failed to save achievement: %v", err))
				}
			}
		}
	}

	m.refreshStats()
	return m, nil
}

func (m *Model) refreshStats() {
	stats, err := m.store.GetStats(context.Background())
	if err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to refresh stats: %v", err))
		return
	}
	stats.Total = len(m.graph.Problems)
	m.statsBar = views.NewStatsBar(stats)
	m.statsBar.SetPalette(viewPalette(m.theme))
}

func (m *Model) handleSubmit() (tea.Model, tea.Cmd) {
	if m.leetcode == nil || !m.leetcode.IsAuthenticated() {
		m.notifications.Add("Session expired. Run `leetgo auth` to reconnect.")
		return m, nil
	}

	item, ok := m.list.SelectedItem().(problemItem)
	if !ok {
		return m, nil
	}

	if item.status != roadmap.StatusInProgress && item.status != roadmap.StatusVerified && item.status != roadmap.StatusSolved {
		m.notifications.Add("Start the problem first (press Enter).")
		return m, nil
	}

	stubName := strings.ReplaceAll(item.problem.Slug, "-", "_")
	stubPath := filepath.Join(m.workspace.ProblemDir(item.problem), stubName+stubExt(m.config.Language))

	code, err := os.ReadFile(stubPath)
	if err != nil {
		m.notifications.Add(fmt.Sprintf("Could not read solution file: %v", err))
		return m, nil
	}

	m.submitting = true
	m.notifications.Add("Submitting to LeetCode...")

	lang := leetcodeLang(m.config.Language)
	slug := item.problem.Slug
	client := m.leetcode

	return m, func() tea.Msg {
		startedAt := time.Now()
		result, err := client.Submit(context.Background(), item.problem.ID, slug, lang, string(code))
		return submitResultMsg{problemID: item.problem.ID, slug: slug, language: lang, result: result, err: err, duration: time.Since(startedAt)}
	}
}

func (m *Model) recordSolveLog(msg submitResultMsg) {
	if msg.result == nil {
		return
	}
	_ = m.store.RecordAttempt(context.Background(), &store.AttemptRecord{
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
	if err := m.store.RecordSolveLog(context.Background(), log); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to record solve log: %v", err))
	}
}

func (m *Model) markAcceptedSubmission(problemID int) {
	ctx := context.Background()
	alreadyClaimed, err := m.store.HasRewardEvent(ctx, problemID, "submit")
	if err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to check submit reward: %v", err))
		return
	}

	if err := m.store.SetProgress(ctx, problemID, roadmap.StatusSolved); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to mark solved: %v", err))
		return
	}

	m.recordAcceptedProvenance(ctx, problemID)

	if alreadyClaimed {
		m.notifications.Add("Submit XP already claimed for this problem.")
		m.refreshStats()
		return
	}

	problem, ok := m.graph.Problems[problemID]
	if ok {
		xp := store.XPForDifficulty(problem.Difficulty) * 30 / 100
		if err := m.store.AddXP(ctx, xp); err != nil {
			m.notifications.Add(fmt.Sprintf("Failed to add XP: %v", err))
			return
		}
		event := &store.RewardEvent{
			ProblemID: problemID,
			Kind:      "submit",
			XP:        xp,
		}
		if err := m.store.RecordRewardEvent(ctx, event); err != nil {
			m.notifications.Add(fmt.Sprintf("Failed to record submit event: %v", err))
			return
		}
	}
	if err := m.store.UpdateStreak(ctx); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to update streak: %v", err))
	}
	m.refreshStats()
}

func (m *Model) recordAcceptedProvenance(ctx context.Context, problemID int) {
	sp, err := m.store.GetSolveProvenance(ctx, problemID)
	if err != nil {
		return
	}

	logs, err := m.store.GetSolveLogsForProblem(ctx, problemID)
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
		if err := m.store.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			Note:       sp.Note,
			SolveLogID: logID,
		}); err != nil {
			return
		}
	} else {
		if err := m.store.RecordSolveProvenance(ctx, &store.SolveProvenance{
			ProblemID:  problemID,
			Kind:       "accepted",
			SolveLogID: logID,
		}); err != nil {
			return
		}
	}
}

func stubExt(lang string) string {
	switch lang {
	case "go":
		return ".go"
	case "python":
		return ".py"
	case "typescript":
		return ".ts"
	case "java":
		return ".java"
	case "cpp":
		return ".cpp"
	case "javascript":
		return ".js"
	case "rust":
		return ".rs"
	case "csharp":
		return ".cs"
	default:
		return ".go"
	}
}

func leetcodeLang(lang string) string {
	switch lang {
	case "go":
		return "golang"
	case "python":
		return "python3"
	case "typescript":
		return "typescript"
	case "java":
		return "java"
	case "cpp":
		return "cpp"
	case "javascript":
		return "javascript"
	case "rust":
		return "rust"
	case "csharp":
		return "csharp"
	default:
		return "golang"
	}
}

func (m *Model) openEditor(stubPath string) {
	editor := m.config.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	cmd := editorCommand(editor, []string{stubPath}, filepath.Dir(stubPath), editorLaunchDetached)
	if err := startEditorCommand(cmd); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to open editor: %v", err))
	}
}
