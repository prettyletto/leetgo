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

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			PaddingTop(1)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("219"))

	statusStyle = lipgloss.NewStyle().Width(12)
	lockedStyle = statusStyle.Foreground(lipgloss.Color("240"))
	availStyle  = statusStyle.Foreground(lipgloss.Color("214"))
	activeStyle = statusStyle.Foreground(lipgloss.Color("39"))
	solvedStyle = statusStyle.Foreground(lipgloss.Color("82"))

	diffStyle   = lipgloss.NewStyle().Width(8)
	easyStyle   = diffStyle.Foreground(lipgloss.Color("82"))
	mediumStyle = diffStyle.Foreground(lipgloss.Color("214"))
	hardStyle   = diffStyle.Foreground(lipgloss.Color("196"))

	catStyle = lipgloss.NewStyle().
			Width(20).
			Foreground(lipgloss.Color("243"))

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1).
			MarginTop(1)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("219"))
)

type problemItem struct {
	problem *roadmap.Problem
	status  roadmap.Status
}

func (i problemItem) Title() string {
	return fmt.Sprintf("%d. %s", i.problem.ID, i.problem.Title)
}

func (i problemItem) Description() string {
	status := renderStatus(i.status)
	diff := renderDifficulty(i.problem.Difficulty)
	cat := catStyle.Render(string(i.problem.Category))
	return fmt.Sprintf("%s %s %s", status, diff, cat)
}

func (i problemItem) FilterValue() string {
	return i.problem.Title
}

func renderStatus(s roadmap.Status) string {
	switch s {
	case roadmap.StatusLocked:
		return lockedStyle.Render("[LOCKED]")
	case roadmap.StatusAvailable:
		return availStyle.Render("[READY]")
	case roadmap.StatusInProgress:
		return activeStyle.Render("[ACTIVE]")
	case roadmap.StatusSolved:
		return solvedStyle.Render("[SOLVED]")
	case roadmap.StatusVerified:
		return availStyle.Render("[VERIFIED]")
	default:
		return statusStyle.Render("[???]")
	}
}

func renderDifficulty(d roadmap.Difficulty) string {
	switch d {
	case roadmap.DifficultyEasy:
		return easyStyle.Render("Easy")
	case roadmap.DifficultyMedium:
		return mediumStyle.Render("Medium")
	case roadmap.DifficultyHard:
		return hardStyle.Render("Hard")
	default:
		return diffStyle.Render("???")
	}
}

type viewMode int

const (
	viewList viewMode = iota
	viewGraph
	viewHeatmap
)

type tickMsg time.Time
type submitResultMsg struct {
	problemID int
	slug      string
	language  string
	result    *leetcode.SubmissionResult
	err       error
}

type testRunResultMsg struct {
	problemID  int
	difficulty roadmap.Difficulty
	output     string
	passed     bool
	alreadyRun bool
}

type Model struct {
	list          list.Model
	graphView     *views.GraphView
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
		if status == roadmap.StatusSolved || status == roadmap.StatusVerified {
			solved[id] = true
		}
	}

	sorted, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("sort problems: %w", err)
	}

	items := make([]list.Item, len(sorted))
	for i, p := range sorted {
		status := roadmap.StatusLocked
		if s, ok := progress[p.ID]; ok {
			status = s
		} else if graph.IsUnlocked(p.ID, solved) {
			status = roadmap.StatusAvailable
		}
		items[i] = problemItem{problem: p, status: status}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 24)
	l.Title = rm.Title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	graphView := views.NewGraphView(rm, progress)

	stats, err := db.GetStats(context.Background())
	if err != nil {
		stats = &store.Stats{}
	}
	stats.Total = len(sorted)
	statsBar := views.NewStatsBar(stats)

	notifications := views.NewNotificationManager()
	gamificationEngine := gamification.NewEngine(db, graph)

	streakDays, err := db.GetStreakDays(context.Background())
	if err != nil {
		streakDays = nil
	}
	heatmapView := views.NewHeatmapView(streakDays)

	lcClient, err := leetcode.NewClient()
	if err != nil {
		lcClient = nil
	}

	return &Model{
		list:          l,
		graphView:     graphView,
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
			m.notifications.Add(fmt.Sprintf("Accepted! Runtime: %s, Memory: %s", msg.result.Runtime, msg.result.Memory))
		} else {
			m.recordSolveLog(msg)
			m.notifications.Add(fmt.Sprintf("%s (%d/%d tests passed)", msg.result.Status, msg.result.PassedTests, msg.result.TotalTests))
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-10)
		m.graphView.SetSize(msg.Width, msg.Height)
		m.heatmapView.SetSize(msg.Width, msg.Height)
		m.statsBar.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "g":
			if m.viewMode == viewList {
				m.viewMode = viewGraph
			} else {
				m.viewMode = viewList
			}
			return m, nil

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

		case "up", "k":
			if m.viewMode == viewGraph {
				m.graphView.Scroll(-1)
				return m, nil
			}

		case "down", "j":
			if m.viewMode == viewGraph {
				m.graphView.Scroll(1)
				return m, nil
			}
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
	case viewGraph:
		view = m.graphView.Render()
	case viewHeatmap:
		view = m.heatmapView.Render()
	default:
		view = m.list.View() + "\n" + m.renderProblemDetails()
	}

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
		detailTitleStyle.Render(fmt.Sprintf("#%d %s", p.ID, p.Title)),
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
	items := []string{
		keyStyle.Render("enter") + " start",
		keyStyle.Render("m") + " solve",
		keyStyle.Render("s") + " submit",
		keyStyle.Render("g") + " unlock path",
		keyStyle.Render("h") + " heatmap",
		keyStyle.Render("tab") + " stage",
		keyStyle.Render("0") + " all",
		keyStyle.Render("/") + " filter",
		keyStyle.Render("q") + " quit",
	}
	if m.viewMode == viewGraph {
		items = []string{
			keyStyle.Render("j/k") + " scroll",
			keyStyle.Render("g") + " list",
			keyStyle.Render("h") + " heatmap",
			keyStyle.Render("q") + " quit",
		}
	}
	if m.viewMode == viewHeatmap {
		items = []string{
			keyStyle.Render("h") + " list",
			keyStyle.Render("g") + " unlock path",
			keyStyle.Render("q") + " quit",
		}
	}
	return footerStyle.Render(strings.Join(items, "  "))
}

func (m *Model) renderStageFilter() string {
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
	m.openEditor(stubPath, testPath)

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
}

func (m *Model) handleSubmit() (tea.Model, tea.Cmd) {
	if m.leetcode == nil || !m.leetcode.IsAuthenticated() {
		m.notifications.Add("Not authenticated. Run `leetgo auth` first.")
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
		result, err := client.Submit(context.Background(), slug, lang, string(code))
		return submitResultMsg{problemID: item.problem.ID, slug: slug, language: lang, result: result, err: err}
	}
}

func (m *Model) recordSolveLog(msg submitResultMsg) {
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

func (m *Model) openEditor(stubPath, testPath string) {
	editor := m.config.Editor
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	cmd := editorCommand(editor, []string{stubPath, testPath}, filepath.Dir(stubPath))
	if err := cmd.Start(); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to open editor: %v", err))
	}
}
