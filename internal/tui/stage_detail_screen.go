package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

type StageDetailScreen struct {
	cfg          *config.Config
	theme        *Theme
	db           store.Store
	roadmap      *roadmap.Roadmap
	stageID      string
	returnScreen string
	returnStage  string

	focusIndex int
	problems   []*roadmap.Problem
	progress   map[int]roadmap.Status
	activePane stageDetailPane

	width  int
	height int
}

type stageDetailPane int

const (
	stagePaneProblems stageDetailPane = iota
	stagePaneSummary
	stagePaneReview
)

func NewStageDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap, stageID string) *StageDetailScreen {
	progress, _ := db.GetAllProgress(context.Background())
	if progress == nil {
		progress = make(map[int]roadmap.Status)
	}

	s := &StageDetailScreen{
		cfg:          cfg,
		theme:        theme,
		db:           db,
		roadmap:      rm,
		stageID:      stageID,
		returnScreen: ScreenRoadmapDetail,
		progress:     progress,
	}

	sorted, err := rm.Graph.TopologicalSort()
	if err == nil {
		for _, p := range sorted {
			effectiveStage := p.Stage
			if effectiveStage == "" {
				effectiveStage = string(p.Category)
			}
			if effectiveStage == stageID {
				s.problems = append(s.problems, p)
			}
		}
	}

	return s
}

func (s *StageDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *StageDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
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

		case "esc", "backspace":
			return s, func() tea.Msg {
				return s.backNavigationMsg()
			}

		case "enter":
			if s.focusIndex < len(s.problems) {
				p := s.problems[s.focusIndex]
				return s, func() tea.Msg {
					return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: p.ID, ReturnScreen: s.returnScreen, ReturnStage: s.returnStage}
				}
			}
			return s, nil

		case "j", "down":
			if len(s.problems) > 0 {
				s.focusIndex = (s.focusIndex + 1) % len(s.problems)
			}

		case "k", "up":
			if len(s.problems) > 0 {
				s.focusIndex--
				if s.focusIndex < 0 {
					s.focusIndex = len(s.problems) - 1
				}
			}

		case "h", "left":
			if s.usesCompactPanes() {
				s.cyclePane(-1)
			}

		case "l", "right":
			if s.usesCompactPanes() {
				s.cyclePane(1)
			}

		}
	}
	return s, nil
}

func (s *StageDetailScreen) backNavigationMsg() tea.Msg {
	returnScreen := s.returnScreen
	if returnScreen == "" {
		returnScreen = ScreenRoadmapDetail
	}
	return NavigateMsg{ScreenID: returnScreen, Stage: s.returnStage}
}

func (s *StageDetailScreen) View() string {
	stageTitle := s.stageTitle()
	stageDesc := s.stageDescription()
	stagePrefix, gridLabel, recommendedLabel, reviewLabel := themeStageLabels(s.theme)

	solved, verified, inProgress, available, locked, total := s.statusCounts()
	header := renderScreenHeader(s.theme, stagePrefix+": "+stageTitle, stageDesc)
	summaryLines := []string{fmt.Sprintf("Solved: %d  Verified: %d  In Progress: %d  Available: %d  Locked: %d  Total: %d",
		solved, verified, inProgress, available, locked, total)}

	if total > 0 {
		percentage := float64(solved) / float64(total) * 100
		bar := s.renderProgressBar(percentage)
		summaryLines = append(summaryLines, bar)
	}
	panelWidth := s.stagePanelWidth()
	summary := renderRoadmapPanel(s.theme, "Stage Summary", strings.Join(summaryLines, "\n"), false, panelWidth)

	var problemLines []string
	problemLines = append(problemLines, s.theme.Subtitle.Render(gridLabel))

	var firstAvailable *roadmap.Problem
	for _, p := range s.problems {
		if s.effectiveStatus(p) == roadmap.StatusAvailable && firstAvailable == nil {
			firstAvailable = p
		}
	}

	for i, p := range s.problems {
		isRecommended := firstAvailable != nil && p.ID == firstAvailable.ID
		line := s.renderProblemLine(p, i == s.focusIndex, isRecommended, recommendedLabel)
		problemLines = append(problemLines, line)
	}

	if len(s.problems) == 0 {
		problemLines = append(problemLines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render("  No problems in this stage."))
	}
	problems := renderRoadmapPanel(s.theme, "Stage Progress", strings.Join(problemLines, "\n"), false, panelWidth)

	reviewLines := s.renderReviewShrine(reviewLabel)
	parts := []string{summary, problems}
	if len(reviewLines) > 0 {
		parts = append(parts, renderRoadmapPanel(s.theme, "Review", strings.Join(reviewLines, "\n"), false, panelWidth))
	}
	if s.usesCompactPanes() {
		s.clampPane(reviewLines)
		switch s.activePane {
		case stagePaneSummary:
			parts = []string{summary}
		case stagePaneReview:
			parts = []string{renderRoadmapPanel(s.theme, "Review", strings.Join(reviewLines, "\n"), false, panelWidth)}
		default:
			parts = []string{problems}
		}
	}

	footer := s.renderFooter()
	return renderScreenShell(s.theme, s.width, s.height, header, strings.Join(parts, "\n\n"), footer)
}

func (s *StageDetailScreen) statusCounts() (solved, verified, inProgress, available, locked, total int) {
	total = len(s.problems)
	for _, p := range s.problems {
		status := s.effectiveStatus(p)
		switch status {
		case roadmap.StatusSolved:
			solved++
		case roadmap.StatusVerified:
			verified++
		case roadmap.StatusInProgress:
			inProgress++
		case roadmap.StatusAvailable:
			available++
		default:
			locked++
		}
	}
	return
}

func (s *StageDetailScreen) stageTitle() string {
	for _, stage := range s.roadmap.Stages {
		if stage.ID == s.stageID {
			return stage.Title
		}
	}
	description := s.stageID
	if len(description) > 0 {
		description = strings.ToUpper(description[:1]) + description[1:]
	}
	return description
}

func (s *StageDetailScreen) stageDescription() string {
	for _, stage := range s.roadmap.Stages {
		if stage.ID == s.stageID {
			return stage.Description
		}
	}
	return ""
}

func (s *StageDetailScreen) completionCount() (int, int) {
	solved := 0
	for _, p := range s.problems {
		if s.progress[p.ID] == roadmap.StatusSolved {
			solved++
		}
	}
	return solved, len(s.problems)
}

func (s *StageDetailScreen) renderProgressBar(percentage float64) string {
	width := 20
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := views.ProgressBar(filled, width, width, "█", "░")
	return fmt.Sprintf("[%s] %.0f%%", lipgloss.NewStyle().Foreground(s.theme.XP).Render(bar), percentage)
}

func (s *StageDetailScreen) renderProblemLine(p *roadmap.Problem, focused bool, isRecommended bool, recommendedLabel string) string {
	status := s.effectiveStatus(p)
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	marker := renderStatusPill(s.theme, symbols, status)
	prefix := marker + " "
	label := fmt.Sprintf("#%d %s", p.ID, p.Title)
	wrapWidth := s.stageProblemTextWidth() - lipgloss.Width(prefix)
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	wrapped := strings.Split(wrapText(label, wrapWidth), "\n")
	line := prefix
	if len(wrapped) > 0 {
		line += wrapped[0]
	}

	if isRecommended && status == roadmap.StatusAvailable {
		line += "  " + lipgloss.NewStyle().
			Foreground(s.theme.Warning).
			Render("["+recommendedLabel+"]")
	}

	lines := []string{line}
	indent := strings.Repeat(" ", lipgloss.Width(prefix))
	for _, wrappedLine := range wrapped[1:] {
		lines = append(lines, indent+wrappedLine)
	}

	return renderSelectableBlock(s.theme, focused, strings.Join(lines, "\n"))
}

func (s *StageDetailScreen) stagePanelWidth() int {
	width := responsiveShellContentWidth(s.width)
	if width <= 0 {
		return 96
	}
	width -= 4
	if width < 40 {
		return 40
	}
	return width
}

func (s *StageDetailScreen) stageProblemTextWidth() int {
	w := s.stagePanelWidth() - 8
	if w < 40 {
		w = 40
	}
	if w > 72 {
		w = 72
	}
	return w
}

func (s *StageDetailScreen) effectiveStatus(p *roadmap.Problem) roadmap.Status {
	if status, ok := s.progress[p.ID]; ok && status != "" {
		return status
	}

	for _, prereq := range p.Prerequisites {
		if s.progress[prereq] != roadmap.StatusSolved {
			return roadmap.StatusLocked
		}
	}
	return roadmap.StatusAvailable
}

func (s *StageDetailScreen) renderStatusMarker(status roadmap.Status) string {
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	return lipgloss.NewStyle().Width(14).Render(renderStatusPill(s.theme, symbols, status))
}

func (s *StageDetailScreen) renderReviewShrine(label string) []string {
	cycles, err := s.db.GetReviewCycles(context.Background())
	if err != nil || len(cycles) == 0 {
		return nil
	}
	ids := make(map[int]bool)
	for _, p := range s.problems {
		ids[p.ID] = true
	}
	var lines []string
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	for _, cycle := range cycles {
		if cycle.CompletedAt != nil || !ids[cycle.ProblemID] {
			continue
		}
		if len(lines) == 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Review).Bold(true).Render(label))
		}
		if p, ok := s.roadmap.Graph.Problems[cycle.ProblemID]; ok {
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Review).Render(fmt.Sprintf("  %s #%d %s", symbols.Review, p.ID, p.Title)))
		}
	}
	return lines
}

func (s *StageDetailScreen) cyclePane(delta int) {
	panes := s.availablePanes()
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

func (s *StageDetailScreen) clampPane(reviewLines []string) {
	if s.activePane == stagePaneReview && len(reviewLines) == 0 {
		s.activePane = stagePaneProblems
	}
}

func (s *StageDetailScreen) availablePanes() []stageDetailPane {
	panes := []stageDetailPane{stagePaneProblems, stagePaneSummary}
	if len(s.renderReviewShrine("Review")) > 0 {
		panes = append(panes, stagePaneReview)
	}
	return panes
}

func (s *StageDetailScreen) usesCompactPanes() bool {
	return (s.width > 0 && s.width < 90) || (s.height > 0 && s.height < 28)
}

func (s *StageDetailScreen) missingPrerequisites(p *roadmap.Problem) []string {
	var missing []string
	for _, id := range p.Prerequisites {
		if s.progress[id] == roadmap.StatusSolved {
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

func (s *StageDetailScreen) renderFooter() string {
	if s.usesCompactPanes() {
		items := []string{
			s.theme.Key.Render("h/l") + " pane",
			s.theme.Key.Render("j/k") + " navigate",
			s.theme.Key.Render("enter") + " problem",
			s.theme.Key.Render("esc") + " roadmap",
			s.theme.Key.Render("q") + " quit",
		}
		return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  •  "))
	}

	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("enter") + " problem",
		s.theme.Key.Render("esc") + " roadmap",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  •  "))
}
