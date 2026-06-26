package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/gamification"
	"github.com/prettyletto/leetgo/internal/recommendation"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

type DashboardScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap

	actions        []recommendation.NextAction
	focusIndex     int
	stats          *store.Stats
	achievementIDs []string

	comingSoon []comingSoonItem

	languageMode  bool
	languageIndex int
	languages     []string
	activePane    dashboardPane

	width  int
	height int
}

type dashboardPane int

const (
	dashboardPaneActions dashboardPane = iota
	dashboardPaneProfile
	dashboardPaneRoadmap
)

type comingSoonItem struct {
	problem  *roadmap.Problem
	blockers []*roadmap.Problem
	status   roadmap.Status
}

func NewDashboardScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap) *DashboardScreen {
	s := &DashboardScreen{
		cfg:        cfg,
		theme:      theme,
		db:         db,
		roadmap:    rm,
		focusIndex: 0,
		languages:  supportedLanguages(),
	}
	s.refresh(context.Background())
	return s
}

func supportedLanguages() []string {
	return []string{"go", "python", "typescript", "javascript", "java", "cpp", "rust", "csharp"}
}

func (s *DashboardScreen) refresh(ctx context.Context) {
	calc := recommendation.NewCalculator(s.db, s.roadmap)
	actions, err := calc.Calculate(ctx)
	if err == nil {
		if !s.cfg.GitExportEnabled || strings.TrimSpace(s.cfg.GitExportRepo) == "" {
			actions = filterExports(actions)
		}
		s.actions = actions
		s.clampFocus()
	}

	stats, err := s.db.GetStats(ctx)
	if err == nil {
		stats.Total = len(s.roadmap.Graph.Problems)
		s.stats = stats
	}

	achievementIDs, err := s.db.GetAchievements(ctx)
	if err == nil {
		s.achievementIDs = achievementIDs
	}

	s.comingSoon = s.buildComingSoon()
}

func (s *DashboardScreen) buildComingSoon() []comingSoonItem {
	solvedMap := s.solvedMap()
	progress, _ := s.db.GetAllProgress(context.Background())
	if progress == nil {
		progress = make(map[int]roadmap.Status)
	}

	sorted, err := s.roadmap.Graph.TopologicalSort()
	if err != nil {
		return nil
	}

	var items []comingSoonItem
	for _, p := range sorted {
		if solvedMap[p.ID] {
			continue
		}
		if progress[p.ID] == roadmap.StatusInProgress {
			continue
		}

		var blockers []*roadmap.Problem
		for _, prereqID := range p.Prerequisites {
			if !solvedMap[prereqID] {
				if bp, ok := s.roadmap.Graph.Problems[prereqID]; ok {
					blockers = append(blockers, bp)
				}
			}
		}

		if len(blockers) >= 1 && len(blockers) <= 2 {
			status := progress[p.ID]
			if status == "" {
				status = roadmap.StatusLocked
			}
			items = append(items, comingSoonItem{problem: p, blockers: blockers, status: status})
			if len(items) >= 3 {
				break
			}
		}
	}
	return items
}

func (s *DashboardScreen) Init() tea.Cmd {
	return nil
}

func (s *DashboardScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg.(type) {
	case NavigateMsg:
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		if s.languageMode {
			return s.handleLanguageKey(msg)
		}
		if delta, ok := dashboardFocusDelta(msg); ok {
			s.moveFocus(delta)
			return s, nil
		}
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "L":
			s.enterLanguageMode()
			return s, nil

		case "h", "left":
			if s.usesCompactPanes() {
				s.cyclePane(-1)
				return s, nil
			}

		case "l", "right":
			if s.usesCompactPanes() {
				s.cyclePane(1)
				return s, nil
			}
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenLegacyList}
			}

		case "enter":
			if s.focusIndex < len(s.actions) {
				action := s.actions[s.focusIndex]
				switch action.Kind {
				case recommendation.KindContinue, recommendation.KindStart, recommendation.KindSubmit, recommendation.KindManualSolve:
					return s, func() tea.Msg {
						return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: action.ProblemID}
					}
				case recommendation.KindReview:
					s.ensureReviewCycle(action)
					return s, func() tea.Msg {
						return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: action.ProblemID}
					}
				case recommendation.KindConnectLeetCode:
					return s, func() tea.Msg {
						return GlobalNotificationMsg{Message: "Run leetgo auth to connect your LeetCode Session for Submission XP."}
					}
				case recommendation.KindViewRoadmapCompletion:
					return s, func() tea.Msg {
						return NavigateMsg{ScreenID: ScreenCompletion}
					}
				case recommendation.KindExport:
					return s, func() tea.Msg {
						return GlobalNotificationMsg{Message: "Git Export from Dashboard is not ready yet. Use leetgo git-export <repo-dir> --commit."}
					}
				case recommendation.KindInspect:
					return s, func() tea.Msg {
						return NavigateMsg{ScreenID: ScreenSolveLog}
					}
				}
			}

		case "r":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenRoadmapDetail}
			}

		case "s":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenSolveLog}
			}
		}
	}
	return s, nil
}

func (s *DashboardScreen) ensureReviewCycle(action recommendation.NextAction) {
	if action.Kind != recommendation.KindReview || action.ProblemID == 0 {
		return
	}
	ctx := context.Background()
	reason := "weakness"
	if action.ReasonType == recommendation.ReasonValidatesManualSolve {
		reason = "manual_solve_validation"
	} else if strings.Contains(action.Reason, "failed-attempt") {
		reason = "failed_attempts"
	}
	cycles, err := s.db.GetReviewCyclesForProblem(ctx, action.ProblemID)
	if err == nil {
		for _, rc := range cycles {
			if rc.CompletedAt == nil && rc.Reason == reason {
				return
			}
		}
	}
	_ = s.db.CreateReviewCycle(ctx, &store.ReviewCycle{
		ProblemID: action.ProblemID,
		Reason:    reason,
		RoadmapID: s.roadmap.ID,
	})
}

func dashboardFocusDelta(msg tea.KeyMsg) (int, bool) {
	switch msg.Type {
	case tea.KeyDown, tea.KeyShiftDown, tea.KeyCtrlDown, tea.KeyCtrlShiftDown:
		return 1, true
	case tea.KeyUp, tea.KeyShiftUp, tea.KeyCtrlUp, tea.KeyCtrlShiftUp:
		return -1, true
	}

	switch msg.String() {
	case "down", "shift+down", "ctrl+down", "ctrl+shift+down", "alt+down", "alt+shift+down", "alt+ctrl+down", "alt+ctrl+shift+down", "j":
		return 1, true
	case "up", "shift+up", "ctrl+up", "ctrl+shift+up", "alt+up", "alt+shift+up", "alt+ctrl+up", "alt+ctrl+shift+up", "k":
		return -1, true
	default:
		return 0, false
	}
}

func (s *DashboardScreen) moveFocus(delta int) {
	n := len(s.actions)
	if n == 0 {
		return
	}

	maxShow := 5
	rendered := n
	if rendered > maxShow {
		rendered = maxShow
	}

	s.focusIndex = (s.focusIndex + delta) % rendered
	if s.focusIndex < 0 {
		s.focusIndex += rendered
	}
}

func (s *DashboardScreen) clampFocus() {
	n := len(s.actions)
	if n == 0 {
		return
	}

	maxShow := 5
	if s.focusIndex >= n || (n > maxShow && s.focusIndex >= maxShow) {
		s.focusIndex = 0
	}
}

func (s *DashboardScreen) View() string {
	greeting := "Dashboard"
	if s.cfg.DisplayName != "" {
		greeting = fmt.Sprintf("Welcome back, %s", s.cfg.DisplayName)
	}

	headerLines := []string{s.theme.Title.Render(greeting)}
	if s.roadmap != nil {
		headerLines = append(headerLines, s.theme.Subtitle.Render(s.roadmap.Title+"  •  guided practice roadmap"))
	}
	header := strings.Join(headerLines, "\n")

	if s.languageMode {
		return s.renderLanguagePicker(header)
	}

	shellWidth := s.width - 10
	if shellWidth > 120 {
		shellWidth = 120
	}
	if shellWidth < 46 {
		shellWidth = 46
	}

	compactHeight := s.height > 0 && s.height < 34
	sidebarWidth := 42
	mainWidth := shellWidth
	if compactHeight && s.width >= 90 {
		sidebarWidth = 34
		mainWidth = shellWidth - sidebarWidth - 2
	} else if s.width >= 118 {
		mainWidth = shellWidth - sidebarWidth - 2
	} else if s.width >= 78 {
		sidebarWidth = mainWidth
	}
	if mainWidth < 34 {
		mainWidth = 34
	}

	if s.usesCompactPanes() {
		body := s.renderActiveCompactPane(shellWidth)
		return renderScreenShell(s.theme, s.width, s.height, header, body, s.renderFooter())
	}

	main := s.renderCenter(mainWidth)
	sidebar := s.renderSidebar(sidebarWidth)
	if compactHeight && s.width >= 90 {
		main = s.renderCompactCenter(mainWidth)
		sidebar = s.renderCompactSidebar(sidebarWidth)
	}
	footer := s.renderFooter()

	var body string
	if compactHeight && s.width >= 90 {
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(mainWidth).Render(main),
			"  ",
			lipgloss.NewStyle().Width(sidebarWidth).Render(sidebar),
		)
	} else if s.width >= 118 {
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(mainWidth).Render(main),
			"  ",
			lipgloss.NewStyle().Width(sidebarWidth).Render(sidebar),
		)
	} else if s.width >= 78 {
		body = main + "\n\n" + sidebar
	} else {
		body = main
	}

	return renderScreenShell(s.theme, s.width, s.height, header, body, footer)
}

func (s *DashboardScreen) centerContent(content string) string {
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *DashboardScreen) renderHUD() string {
	return s.renderHUDWithWidth(30)
}

func (s *DashboardScreen) renderHUDWithWidth(width int) string {
	var lines []string
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	lines = append(lines, s.theme.Key.Render(s.cfg.DisplayName))

	if s.stats != nil {
		lines = append(lines, fmt.Sprintf("%s Level %d", symbols.XP, s.stats.Level))
		lines = append(lines, s.renderXPProgress())
		lines = append(lines, fmt.Sprintf("Streak: %d days", s.stats.Streak))
		lines = append(lines, fmt.Sprintf("Verified: %d", s.stats.Verified))
		lines = append(lines, fmt.Sprintf("Solved: %d", s.stats.Solved))
		progress := s.stats.Verified + s.stats.Solved
		lines = append(lines, fmt.Sprintf("Progress: %d/%d", progress, s.stats.Total))
	}

	lines = append(lines, fmt.Sprintf("Roadmap: %s", s.cfg.Roadmap))
	if latest := s.renderLatestAchievement(); latest != "" {
		lines = append(lines, wrapText(latest, width-4))
	}

	content := strings.Join(lines, "\n")
	return renderRoadmapPanel(s.theme, "Profile", content, false, width)
}

func (s *DashboardScreen) renderCompactHUD(width int) string {
	var lines []string
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	name := s.cfg.DisplayName
	if name == "" {
		name = "Profile"
	}
	lines = append(lines, s.theme.Key.Render(name))
	if s.stats != nil {
		progress := s.stats.Verified + s.stats.Solved
		lines = append(lines, fmt.Sprintf("%s L%d  %d/%d", symbols.XP, s.stats.Level, progress, s.stats.Total))
		lines = append(lines, fmt.Sprintf("Streak: %d  XP: %d", s.stats.Streak, s.stats.TotalXP))
	}
	return renderRoadmapPanel(s.theme, "Profile", strings.Join(lines, "\n"), false, width)
}

func (s *DashboardScreen) renderXPProgress() string {
	if s.stats == nil {
		return ""
	}

	currentLevelXP := store.LevelToXP(s.stats.Level)
	nextLevelXP := store.LevelToXP(s.stats.Level + 1)
	xpInLevel := s.stats.TotalXP - currentLevelXP
	xpNeeded := nextLevelXP - currentLevelXP

	if xpNeeded <= 0 {
		return fmt.Sprintf("%d XP", s.stats.TotalXP)
	}

	barWidth := 10
	percentage := float64(xpInLevel) / float64(xpNeeded)
	filled := int(percentage * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := views.ProgressBar(filled, barWidth, barWidth, "█", "░")
	return fmt.Sprintf("[%s] %d/%d XP", lipgloss.NewStyle().Foreground(s.theme.XP).Render(bar), xpInLevel, xpNeeded)
}

func (s *DashboardScreen) renderLatestAchievement() string {
	if len(s.achievementIDs) == 0 {
		return ""
	}

	latestID := s.achievementIDs[len(s.achievementIDs)-1]
	ach, ok := gamification.Achievements[latestID]
	if !ok {
		return ""
	}

	return lipgloss.NewStyle().
		Foreground(s.theme.Warning).
		Render(fmt.Sprintf("Latest achievement: %s %s", ach.Icon, ach.Name))
}

func (s *DashboardScreen) renderRoadmapContext() string {
	return s.renderRoadmapContextWithWidth(30)
}

func (s *DashboardScreen) renderRoadmapContextWithWidth(width int) string {
	var lines []string
	lines = append(lines, s.theme.Key.Render(s.roadmap.Title))

	solvedMap := s.solvedMap()
	currentStage := ""
	stageCompletions := make(map[string][2]int)

	for _, stage := range s.roadmap.Stages {
		total := 0
		solved := 0
		for _, p := range s.roadmap.Graph.Problems {
			if p.Stage == stage.ID {
				total++
				if solvedMap[p.ID] {
					solved++
				}
				if currentStage == "" && !solvedMap[p.ID] {
					currentStage = stage.Title
				}
			}
		}
		stageCompletions[stage.Title] = [2]int{solved, total}
	}

	if currentStage == "" && len(s.roadmap.Stages) > 0 {
		currentStage = s.roadmap.Stages[0].Title
	}

	if currentStage != "" {
		lines = append(lines, fmt.Sprintf("Current Stage: %s", currentStage))
		completion := stageCompletions[currentStage]
		lines = append(lines, fmt.Sprintf("Progress: %d/%d solved", completion[0], completion[1]))
	}

	blocker := s.findNextBlocker(solvedMap)
	if blocker != "" {
		lines = append(lines, "")
		lines = append(lines, wrapText(blocker, width-4))
	}

	if len(s.comingSoon) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render("Upcoming:"))
		for _, cs := range s.comingSoon[:min(2, len(s.comingSoon))] {
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(
				wrapText(fmt.Sprintf("  #%d %s", cs.problem.ID, cs.problem.Title), width-4)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, s.theme.Subtitle.Render("Press r for Roadmap Detail"))

	content := strings.Join(lines, "\n")
	return renderRoadmapPanel(s.theme, "Roadmap", content, false, width)
}

func (s *DashboardScreen) renderCompactRoadmapContext(width int) string {
	var lines []string
	lines = append(lines, s.theme.Key.Render(s.roadmap.Title))

	solvedMap := s.solvedMap()
	currentStage := ""
	for _, stage := range s.roadmap.Stages {
		for _, p := range s.roadmap.Graph.Problems {
			if p.Stage == stage.ID && !solvedMap[p.ID] {
				currentStage = stage.Title
				break
			}
		}
		if currentStage != "" {
			break
		}
	}
	if currentStage != "" {
		lines = append(lines, wrapText("Stage: "+currentStage, width-4))
	}
	if len(s.comingSoon) > 0 {
		cs := s.comingSoon[0]
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(wrapText(fmt.Sprintf("Next locked: #%d %s", cs.problem.ID, cs.problem.Title), width-4)))
	}
	lines = append(lines, s.theme.Subtitle.Render("r roadmap"))

	return renderRoadmapPanel(s.theme, "Roadmap", strings.Join(lines, "\n"), false, width)
}

func (s *DashboardScreen) findNextBlocker(solvedMap map[int]bool) string {
	progress, err := s.db.GetAllProgress(context.Background())
	if err != nil {
		return ""
	}

	sorted, err := s.roadmap.Graph.TopologicalSort()
	if err != nil {
		return ""
	}

	for _, p := range sorted {
		if solvedMap[p.ID] {
			continue
		}
		if progress[p.ID] == roadmap.StatusInProgress {
			continue
		}

		if len(p.Prerequisites) == 0 {
			continue
		}

		var blockingPrereq *roadmap.Problem
		for _, prereqID := range p.Prerequisites {
			if !solvedMap[prereqID] {
				blockingPrereq = s.roadmap.Graph.Problems[prereqID]
				break
			}
		}

		if blockingPrereq != nil {
			status := progress[blockingPrereq.ID]
			blockerLabel := fmt.Sprintf("#%d %s", blockingPrereq.ID, blockingPrereq.Title)
			if status == roadmap.StatusVerified {
				blockerLabel += " (Verified)"
			}
			return lipgloss.NewStyle().Foreground(s.theme.Warning).Render(
				fmt.Sprintf("Blocker: #%d %s\n  locked by %s",
					p.ID, p.Title, blockerLabel))
		}
	}
	return ""
}

func (s *DashboardScreen) solvedMap() map[int]bool {
	progress, err := s.db.GetAllProgress(context.Background())
	if err != nil {
		return nil
	}
	solved := make(map[int]bool, len(progress))
	for id, status := range progress {
		if status == roadmap.StatusSolved {
			solved[id] = true
		}
	}
	return solved
}

func (s *DashboardScreen) renderCenter(width int) string {
	var sections []string
	sections = append(sections, s.renderSectionLead(width))

	primary := s.renderPrimaryAction(width)
	if primary != "" {
		sections = append(sections, primary)
	}

	also := s.renderAlsoAvailable(width)
	if also != "" {
		sections = append(sections, also)
	}

	return strings.Join(sections, "\n")
}

func (s *DashboardScreen) renderCompactCenter(width int) string {
	var sections []string
	primary := s.renderPrimaryAction(width)
	if primary != "" {
		sections = append(sections, primary)
	}
	also := s.renderCompactAlsoAvailable(width)
	if also != "" {
		sections = append(sections, also)
	}
	return strings.Join(sections, "\n")
}

func (s *DashboardScreen) renderSectionLead(width int) string {
	copy := "Your next best action is ranked first. Continue in-progress work before starting something new."
	if len(s.actions) == 0 {
		copy = "No next action is ready yet. Open the problem list to generate or inspect workspace files."
	}
	return s.theme.Subtitle.Render(wrapText(copy, width))
}

func (s *DashboardScreen) renderPrimaryAction(width int) string {
	if len(s.actions) == 0 {
		body := wrapText("No actions available yet.", width-8) + "\n\n" + wrapText("Open the problem list with l to start a Problem manually.", width-8)
		return renderRoadmapPanel(s.theme, "Recommended", body, true, width)
	}

	action := s.actions[0]
	focused := s.focusIndex == 0
	marker := s.theme.Key.Render("Recommended")
	if !focused {
		marker = s.theme.Subtitle.Render("Recommended")
	}

	label := formatActionLabel(action.Kind)
	lines := []string{
		marker,
		lipgloss.NewStyle().Bold(true).Foreground(s.theme.PrimaryAccent).Render(action.Title),
		lipgloss.NewStyle().Foreground(s.theme.Muted).Render(label),
		"",
		formatReasonType(action.ReasonType),
		wrapText(action.Reason, width-8),
	}

	if action.ProblemID > 0 {
		detail := fmt.Sprintf("%s · %s", action.Stage, action.Category)
		if detail != " · " {
			lines = append(lines, "", lipgloss.NewStyle().Foreground(s.theme.Muted).Render(detail))
		}
	}

	body := renderSelectableBlock(s.theme, focused, strings.Join(lines, "\n"))
	return renderRoadmapPanel(s.theme, "Recommended", body, focused, width)
}

func (s *DashboardScreen) renderAlsoAvailable(width int) string {
	if len(s.actions) <= 1 {
		return ""
	}

	rest := s.actions[1:]
	maxShow := 4
	if s.height > 0 && s.height < 30 && s.width < 118 {
		maxShow = 2
	}
	if len(rest) > maxShow {
		rest = rest[:maxShow]
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(s.theme.Muted).Render("Queue"))

	for i, action := range rest {
		idx := i + 1
		focused := s.focusIndex == idx

		style := lipgloss.NewStyle()
		if focused {
			style = lipgloss.NewStyle().Bold(true)
		} else {
			style = lipgloss.NewStyle().Foreground(s.theme.Muted)
		}

		label := formatActionLabel(action.Kind)
		lineText := wrapText(fmt.Sprintf("%s  (%s)", action.Title, label), width-8)
		lines = append(lines, renderSelectableBlock(s.theme, focused, style.Render(lineText)))
	}

	return renderRoadmapPanel(s.theme, "Up next", strings.Join(lines, "\n"), false, width)
}

func (s *DashboardScreen) renderCompactAlsoAvailable(width int) string {
	if len(s.actions) <= 1 {
		return ""
	}

	rest := s.actions[1:]
	if len(rest) > 2 {
		rest = rest[:2]
	}

	var lines []string
	for _, action := range rest {
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(wrapText(action.Title+"  ("+formatActionLabel(action.Kind)+")", width-8)))
	}
	return renderRoadmapPanel(s.theme, "Up next", strings.Join(lines, "\n"), false, width)
}

func (s *DashboardScreen) renderSidebar(width int) string {
	sections := []string{s.renderHUDWithWidth(width)}
	if s.width >= 118 {
		sections = append(sections, s.renderRoadmapContextWithWidth(width))
	}
	return strings.Join(sections, "\n\n")
}

func (s *DashboardScreen) renderActiveCompactPane(width int) string {
	s.clampPane()
	switch s.activePane {
	case dashboardPaneProfile:
		return s.renderHUDWithWidth(width)
	case dashboardPaneRoadmap:
		return s.renderRoadmapContextWithWidth(width)
	default:
		return s.renderCenter(width)
	}
}

func (s *DashboardScreen) cyclePane(delta int) {
	panes := []dashboardPane{dashboardPaneActions, dashboardPaneProfile, dashboardPaneRoadmap}
	idx := 0
	for i, pane := range panes {
		if pane == s.activePane {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(panes)
	if idx < 0 {
		idx += len(panes)
	}
	s.activePane = panes[idx]
}

func (s *DashboardScreen) clampPane() {
	if s.activePane < dashboardPaneActions || s.activePane > dashboardPaneRoadmap {
		s.activePane = dashboardPaneActions
	}
}

func (s *DashboardScreen) usesCompactPanes() bool {
	return s.width > 0 && s.width < 78
}

func (s *DashboardScreen) renderCompactSidebar(width int) string {
	sections := []string{s.renderCompactHUD(width)}
	if s.width >= 118 {
		sections = append(sections, s.renderCompactRoadmapContext(width))
	}
	return strings.Join(sections, "\n")
}

func formatReasonType(rt recommendation.ReasonType) string {
	switch rt {
	case recommendation.ReasonUnlocksDependent:
		return "Why: unlocks dependent problems"
	case recommendation.ReasonStrengthensPracticeFocus:
		return "Why: strengthens practice focus"
	case recommendation.ReasonCompletesVerified:
		return "Why: complete verified problem"
	case recommendation.ReasonContinuesInProgress:
		return "Why: continue in-progress work"
	case recommendation.ReasonRepairsWeakness:
		return "Why: repairs weakness"
	case recommendation.ReasonValidatesManualSolve:
		return "Why: validates manual solve"
	case recommendation.ReasonCompletesRoadmap:
		return "Why: completes roadmap"
	default:
		return ""
	}
}

func (s *DashboardScreen) enterLanguageMode() {
	s.languageMode = true
	for i, lang := range s.languages {
		if lang == s.cfg.Language {
			s.languageIndex = i
			return
		}
	}
	s.languageIndex = 0
}

func (s *DashboardScreen) handleLanguageKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		s.languageMode = false
		return s, nil

	case "esc", "backspace", "L":
		s.languageMode = false
		return s, nil

	case "enter":
		return s.selectLanguage()

	case "j", "down":
		if s.languageIndex < len(s.languages)-1 {
			s.languageIndex++
		}

	case "k", "up":
		if s.languageIndex > 0 {
			s.languageIndex--
		}
	}
	return s, nil
}

func (s *DashboardScreen) selectLanguage() (Screen, tea.Cmd) {
	lang := s.languages[s.languageIndex]
	if lang == s.cfg.Language {
		s.languageMode = false
		return s, nil
	}
	s.cfg.Language = lang
	if err := s.cfg.Save(); err != nil {
		s.languageMode = false
		return s, func() tea.Msg {
			return GlobalNotificationMsg{Message: fmt.Sprintf("Failed to save language: %v", err)}
		}
	}
	s.languageMode = false
	return s, func() tea.Msg {
		return GlobalNotificationMsg{Message: fmt.Sprintf("Language set to %s", lang)}
	}
}

func (s *DashboardScreen) renderLanguagePicker(header string) string {
	var lines []string
	for i, lang := range s.languages {
		marker := "  "
		if lang == s.cfg.Language {
			marker = "* "
		}
		line := marker + lang
		if i == s.languageIndex {
			line = "▎ " + line
		}
		lines = append(lines, line)
	}
	body := renderThemedPanel(s.theme, "Language", strings.Join(lines, "\n"), true)

	marker := s.cfg.Language
	subtitle := s.theme.Subtitle.Render("Current: " + marker)
	footer := s.theme.Footer.PaddingTop(1).Render(strings.Join([]string{
		s.theme.Key.Render("j/k") + " select",
		s.theme.Key.Render("enter") + " confirm",
		s.theme.Key.Render("esc") + " cancel",
	}, "  "))

	content := header + "\n\n" + body + "\n" + subtitle + "\n\n" + footer
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *DashboardScreen) renderFooter() string {
	if s.usesCompactPanes() {
		keys := map[string]string{"h/l": "pane", "j/k": "nav", "enter": "open", "r": "roadmap", "s": "log", "L": "lang", "q": "quit"}
		order := []string{"h/l", "j/k", "enter", "r", "s", "L", "q"}
		return s.theme.Footer.PaddingTop(1).Render(views.KeytipFooter(keys, order, viewPalette(s.theme)))
	}
	if s.width > 0 && s.width < 100 {
		keys := map[string]string{"j/k": "nav", "enter": "open", "r": "roadmap", "s": "log", "L": "lang", "l": "list", "q": "quit"}
		order := []string{"j/k", "enter", "r", "s", "L", "l", "q"}
		return s.theme.Footer.PaddingTop(1).Render(views.KeytipFooter(keys, order, viewPalette(s.theme)))
	}
	keys := map[string]string{"up/down": "navigate", "j/k": "navigate", "enter": "open", "r": "roadmap", "s": "practice log", "L": "language", "l": "list", "q": "quit"}
	order := []string{"up/down", "j/k", "enter", "r", "s", "L", "l", "q"}
	return s.theme.Footer.PaddingTop(1).Render(views.KeytipFooter(keys, order, viewPalette(s.theme)))
}

func filterExports(actions []recommendation.NextAction) []recommendation.NextAction {
	filtered := make([]recommendation.NextAction, 0, len(actions))
	for _, a := range actions {
		if a.Kind == recommendation.KindExport {
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}

func formatActionLabel(kind recommendation.ActionKind) string {
	switch kind {
	case "manual_solve":
		return "Manual Solve"
	case "connect_leetcode":
		return "Connect"
	case "view_roadmap_completion":
		return "Completion"
	default:
		s := string(kind)
		return strings.ToUpper(s[:1]) + s[1:]
	}
}
