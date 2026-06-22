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

	width  int
	height int
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
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "l":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenLegacyList}
			}

		case "j", "down":
			s.moveFocus(1)

		case "k", "up":
			s.moveFocus(-1)

		case "enter":
			if s.focusIndex < len(s.actions) {
				action := s.actions[s.focusIndex]
				switch action.Kind {
				case recommendation.KindContinue, recommendation.KindStart:
					return s, func() tea.Msg {
						return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: action.ProblemID}
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

func (s *DashboardScreen) moveFocus(delta int) {
	n := len(s.actions)
	if n == 0 {
		return
	}

	rendered := n
	maxShow := 5
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

	if s.width > 100 {
		left := s.renderHUD()
		center := s.renderNextActions()
		right := s.renderRoadmapContext()
		body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", center, "  ", right)
		return header + "\n" + body + "\n" + s.renderFooter()
	}

	if s.width > 60 {
		center := s.renderNextActions()
		left := s.renderHUD()
		right := s.renderRoadmapContext()
		rails := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
		return header + "\n" + center + "\n" + rails + "\n" + s.renderFooter()
	}

	center := s.renderNextActions()
	return header + "\n" + center + "\n" + s.renderFooter()
}

func (s *DashboardScreen) renderHUD() string {
	var lines []string
	lines = append(lines, s.theme.Key.Render(s.cfg.DisplayName))

	if s.stats != nil {
		lines = append(lines, fmt.Sprintf("Level %d", s.stats.Level))
		lines = append(lines, s.renderXPProgress())
		lines = append(lines, fmt.Sprintf("Streak: %d days", s.stats.Streak))
		lines = append(lines, fmt.Sprintf("Solved: %d/%d", s.stats.Solved, s.stats.Total))
	}

	lines = append(lines, fmt.Sprintf("Roadmap: %s", s.cfg.Roadmap))

	lines = append(lines, s.renderLatestAchievement())

	content := strings.Join(lines, "\n")
	return s.theme.Panel.Width(25).Render(content)
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

	bar := lipgloss.NewStyle().
		Foreground(s.theme.Success).
		Render(strings.Repeat("█", filled))
	bg := lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Render(strings.Repeat("░", barWidth-filled))

	return fmt.Sprintf("[%s%s] %d/%d XP", bar, bg, xpInLevel, xpNeeded)
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
		lines = append(lines, fmt.Sprintf("Stage: %s", currentStage))
		completion := stageCompletions[currentStage]
		lines = append(lines, fmt.Sprintf("Progress: %d/%d solved", completion[0], completion[1]))
	}

	blocker := s.findNextBlocker(solvedMap)
	if blocker != "" {
		lines = append(lines, "")
		lines = append(lines, blocker)
	}

	lines = append(lines, "")
	lines = append(lines, "Press r for Roadmap Detail")

	content := strings.Join(lines, "\n")
	return s.theme.Panel.Width(25).Render(content)
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
			return lipgloss.NewStyle().Foreground(s.theme.Danger).Render(
				fmt.Sprintf("Blocker: #%d %s\n  locked by #%d %s",
					p.ID, p.Title, blockingPrereq.ID, blockingPrereq.Title))
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

func (s *DashboardScreen) renderNextActions() string {
	if len(s.actions) == 0 {
		content := fmt.Sprintf("No actions available.\n\nStart a problem from the list view (press l).")
		return s.theme.Panel.Width(40).Render(content)
	}

	maxShow := 5
	showActions := s.actions
	if len(showActions) > maxShow {
		showActions = showActions[:maxShow]
	}

	actionLines := make([]string, 0, len(showActions))
	for i, action := range showActions {
		line := s.renderActionCard(action, i == s.focusIndex)
		actionLines = append(actionLines, line)
	}

	content := strings.Join(actionLines, "")
	return content
}

func (s *DashboardScreen) renderActionCard(action recommendation.NextAction, focused bool) string {
	var style, labelStyle lipgloss.Style
	if focused {
		style = s.theme.FocusedPanel.Width(40)
		labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(s.theme.SecondaryAccent)
	} else {
		style = s.theme.Panel.Width(40)
		labelStyle = lipgloss.NewStyle().
			Foreground(s.theme.Muted)
	}

	label := string(action.Kind)
	label = strings.ToUpper(label[:1]) + label[1:]

	lines := fmt.Sprintf("%s  %s\n%s",
		labelStyle.Render(label),
		s.theme.Key.Render(action.Title),
		action.Reason,
	)

	if action.ProblemID > 0 {
		detail := fmt.Sprintf("%s · %s", action.Stage, action.Category)
		if detail != " · " {
			lines += "\n" + lipgloss.NewStyle().Foreground(s.theme.Muted).Render(detail)
		}
	}

	return style.Render(lines) + "\n"
}

func (s *DashboardScreen) renderFooter() string {
	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("enter") + " select",
		s.theme.Key.Render("t") + " theme",
		s.theme.Key.Render("r") + " roadmap",
		s.theme.Key.Render("s") + " solve log",
		s.theme.Key.Render("l") + " list",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
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
