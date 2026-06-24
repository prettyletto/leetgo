package tui

import (
	"context"
	"fmt"
	"slices"
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

	width  int
	height int
}

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
	}
	s.refresh(context.Background())
	return s
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
		if delta, ok := dashboardFocusDelta(msg); ok {
			s.moveFocus(delta)
			return s, nil
		}
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "l":
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

		case "t":
			currentIdx := slices.Index(config.ValidThemes, s.cfg.Theme)
			nextIdx := (currentIdx + 1) % len(config.ValidThemes)
			nextTheme := config.ValidThemes[nextIdx]
			return s, func() tea.Msg {
				return ThemeChangedMsg{ThemeID: nextTheme}
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
	greeting := "Welcome"
	if s.cfg.DisplayName != "" {
		greeting = fmt.Sprintf("Welcome, %s", s.cfg.DisplayName)
	}

	header := s.theme.Title.MarginBottom(1).Render(greeting)
	var content string

	if s.width > 100 {
		left := s.renderHUD()
		center := s.renderCenter()
		right := s.renderRoadmapContext()
		body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", center, "  ", right)
		content = header + "\n" + body + "\n" + s.renderFooter()
		return s.centerContent(content)
	}

	if s.width > 60 {
		center := s.renderCenter()
		left := s.renderHUD()
		right := s.renderRoadmapContext()
		rails := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
		content = header + "\n" + center + "\n" + rails + "\n" + s.renderFooter()
		return s.centerContent(content)
	}

	center := s.renderCenter()
	content = header + "\n" + center + "\n" + s.renderFooter()
	return s.centerContent(content)
}

func (s *DashboardScreen) centerContent(content string) string {
	if s.width <= 0 || s.height <= 0 {
		return content
	}
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, content)
}

func (s *DashboardScreen) renderHUD() string {
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

	lines = append(lines, s.renderLatestAchievement())

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(27).Render(renderThemedPanel(s.theme, s.theme.Labels.Profile, content, false))
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
		Render(fmt.Sprintf("Latest: %s %s", ach.Icon, ach.Name))
}

func (s *DashboardScreen) renderRoadmapContext() string {
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
		lines = append(lines, blocker)
	}

	if len(s.comingSoon) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(s.theme.Labels.LockedItems+":"))
		for _, cs := range s.comingSoon[:min(2, len(s.comingSoon))] {
			blockerLabels := make([]string, len(cs.blockers))
			for i, b := range cs.blockers {
				blockerLabels[i] = fmt.Sprintf("#%d %s", b.ID, b.Title)
			}
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(
				fmt.Sprintf("  #%d %s", cs.problem.ID, cs.problem.Title)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Press r for Roadmap Detail")

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(27).Render(renderThemedPanel(s.theme, s.theme.Labels.RoadmapContext, content, false))
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

func (s *DashboardScreen) renderCenter() string {
	var sections []string

	primary := s.renderPrimaryAction()
	if primary != "" {
		sections = append(sections, primary)
	}

	also := s.renderAlsoAvailable()
	if also != "" {
		sections = append(sections, also)
	}

	return strings.Join(sections, "\n")
}

func (s *DashboardScreen) renderPrimaryAction() string {
	if len(s.actions) == 0 {
		return s.theme.Panel.Width(44).Render("No actions available.\n\nStart a problem from the list view (press l).")
	}

	action := s.actions[0]
	focused := s.focusIndex == 0
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)

	var style, labelStyle lipgloss.Style
	_ = symbols
	marker := ">"
	if focused {
		style = s.theme.FocusedPanel.Width(44)
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(s.theme.SecondaryAccent)
	} else {
		style = s.theme.Panel.Width(44)
		labelStyle = lipgloss.NewStyle().Foreground(s.theme.Muted)
	}

	label := formatActionLabel(action.Kind)
	header := lipgloss.NewStyle().Bold(true).Foreground(s.theme.PrimaryAccent).Render(s.theme.Labels.PrimaryAction)

	lines := fmt.Sprintf("%s %s %s\n%s  %s\n       %s",
		marker,
		labelStyle.Render(label),
		s.theme.Key.Render(action.Title),
		header,
		lipgloss.NewStyle().Foreground(s.theme.Muted).Render(formatReasonType(action.ReasonType)),
		action.Reason,
	)

	if action.ProblemID > 0 {
		detail := fmt.Sprintf("%s · %s", action.Stage, action.Category)
		if detail != " · " {
			lines += "\n" + lipgloss.NewStyle().Foreground(s.theme.Muted).Render("       "+detail)
		}
	}

	panelTitle := s.theme.Labels.PrimaryAction
	if s.theme.ID == "rpg-skill-tree" {
		panelTitle = "Quest Board"
	}
	return renderThemedPanel(s.theme, panelTitle, style.Render(lines), focused) + "\n"
}

func (s *DashboardScreen) renderAlsoAvailable() string {
	if len(s.actions) <= 1 {
		return ""
	}

	rest := s.actions[1:]
	maxShow := 4
	if len(rest) > maxShow {
		rest = rest[:maxShow]
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(s.theme.Muted).Render(s.theme.Labels.SecondaryActions))

	for i, action := range rest {
		idx := i + 1
		focused := s.focusIndex == idx

		marker := "  "
		style := lipgloss.NewStyle()
		if focused {
			marker = "> "
			style = lipgloss.NewStyle().
				Foreground(s.theme.SecondaryAccent).Bold(true)
		} else {
			style = lipgloss.NewStyle().Foreground(s.theme.Muted)
		}

		label := formatActionLabel(action.Kind)
		line := fmt.Sprintf("%s%s  %s", marker, label, action.Title)
		lines = append(lines, style.Render(line))
		if focused {
			reasonLine := fmt.Sprintf("   %s", action.Reason)
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(reasonLine))
		}
	}

	return views.Panel("", strings.Join(lines, "\n"), viewPalette(s.theme), false)
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

func (s *DashboardScreen) renderFooter() string {
	keys := map[string]string{"up/down": "navigate", "j/k": "navigate", "enter": "select", "t": "theme", "r": "roadmap", "s": "practice log", "l": "list", "q": "quit"}
	order := []string{"up/down", "j/k", "enter", "t", "r", "s", "l", "q"}
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
