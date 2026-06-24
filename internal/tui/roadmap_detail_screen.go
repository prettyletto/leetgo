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

type roadmapViewMode int

const (
	rdViewList roadmapViewMode = iota
	rdViewGraph
)

type RoadmapDetailScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap

	viewMode   roadmapViewMode
	focusIndex int
	problems   []*roadmap.Problem

	graphView *views.GraphView
	progress  map[int]roadmap.Status

	width  int
	height int
}

func NewRoadmapDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap) *RoadmapDetailScreen {
	progress, _ := db.GetAllProgress(context.Background())
	if progress == nil {
		progress = make(map[int]roadmap.Status)
	}

	s := &RoadmapDetailScreen{
		cfg:       cfg,
		theme:     theme,
		db:        db,
		roadmap:   rm,
		progress:  progress,
		viewMode:  rdViewList,
		graphView: views.NewGraphView(rm, progress),
	}

	sorted, err := rm.Graph.TopologicalSort()
	if err == nil {
		s.problems = sorted
	}

	return s
}

func (s *RoadmapDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *RoadmapDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg.(type) {
	case NavigateMsg:
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.graphView.SetSize(msg.Width, msg.Height)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "esc", "backspace":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenDashboard}
			}

		case "g":
			if s.viewMode == rdViewList {
				s.viewMode = rdViewGraph
			} else {
				s.viewMode = rdViewList
			}
			return s, nil

		case "enter":
			if s.viewMode == rdViewList && s.focusIndex < len(s.problems) {
				p := s.problems[s.focusIndex]
				stage := p.Stage
				if stage == "" {
					stage = string(p.Category)
				}
				return s, func() tea.Msg {
					return NavigateMsg{ScreenID: ScreenStageDetail, Stage: stage}
				}
			}
			return s, nil

		case "j", "down":
			if s.viewMode == rdViewList {
				if len(s.problems) > 0 {
					s.focusIndex = (s.focusIndex + 1) % len(s.problems)
				}
			}

		case "k", "up":
			if s.viewMode == rdViewList {
				if len(s.problems) > 0 {
					s.focusIndex--
					if s.focusIndex < 0 {
						s.focusIndex = len(s.problems) - 1
					}
				}
			}

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

func (s *RoadmapDetailScreen) View() string {
	header := s.renderHeader()

	var body string
	switch s.viewMode {
	case rdViewGraph:
		body = s.graphView.Render()
	default:
		body = s.renderListView()
	}

	footer := s.renderFooter()

	return header + "\n" + body + "\n" + footer
}

func (s *RoadmapDetailScreen) renderHeader() string {
	titlePrefix, _, _ := themeRoadmapLabels(s.theme)
	title := s.theme.Title.Render(titlePrefix + ": " + s.roadmap.Title)
	subtitle := ""
	if s.roadmap.Tagline != "" {
		subtitle = lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render(s.roadmap.Tagline)
	}

	if s.roadmap.Promise != "" {
		subtitle += "\n" + lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Italic(true).
			Render(s.roadmap.Promise)
	}

	return title + "\n" + subtitle
}

func (s *RoadmapDetailScreen) renderListView() string {
	var lines []string

	solvedCount := s.buildStageSolvedCount()
	_, stagePrefix, lockedLabel := themeRoadmapLabels(s.theme)

	groups := s.groupProblemsByStage()
	for _, stage := range s.roadmap.Stages {
		problems, ok := groups[stage.ID]
		if !ok || len(problems) == 0 {
			continue
		}

		total := len(problems)
		solved := solvedCount[stage.ID]

		header := fmt.Sprintf("%s  %d/%d solved",
			s.theme.Key.Render(stagePrefix+": "+stage.Title),
			solved, total,
		)

		if s.roadmap.Stages != nil && len(s.roadmap.Stages) > 1 {
			percentage := float64(solved) / float64(total) * 100
			bar := s.renderMiniBar(percentage)
			header += "  " + bar
		}

		lines = append(lines, header)

		for _, p := range problems {
			line := s.renderProblemLine(p)
			lines = append(lines, line)
		}

		lines = append(lines, "")
	}

	comingSoon := s.buildComingSoon()
	if len(comingSoon) > 0 {
		lines = append(lines, s.theme.Key.Render(lockedLabel))
		for _, item := range comingSoon {
			blockerNames := make([]string, len(item.blockers))
			for i, b := range item.blockers {
				status := ""
				if s.progress[b.ID] == roadmap.StatusVerified {
					status = " (Verified: Submit or Manual Solve to open gate)"
				}
				blockerNames[i] = fmt.Sprintf("#%d %s%s", b.ID, b.Title, status)
			}
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(
				fmt.Sprintf("  #%d %s — blocked by %s", item.problem.ID, item.problem.Title, strings.Join(blockerNames, ", "))))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (s *RoadmapDetailScreen) buildComingSoon() []comingSoonItem {
	solvedMap := make(map[int]bool)
	for id, status := range s.progress {
		if status == roadmap.StatusSolved {
			solvedMap[id] = true
		}
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
		if s.progress[p.ID] == roadmap.StatusInProgress {
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
			items = append(items, comingSoonItem{problem: p, blockers: blockers})
			if len(items) >= 3 {
				break
			}
		}
	}
	return items
}

func (s *RoadmapDetailScreen) groupProblemsByStage() map[string][]*roadmap.Problem {
	groups := make(map[string][]*roadmap.Problem)
	for _, p := range s.problems {
		stage := p.Stage
		if stage == "" {
			stage = string(p.Category)
		}
		groups[stage] = append(groups[stage], p)
	}
	return groups
}

func (s *RoadmapDetailScreen) buildStageSolvedCount() map[string]int {
	solvedCount := make(map[string]int)
	for _, p := range s.problems {
		stage := p.Stage
		if stage == "" {
			stage = string(p.Category)
		}
		if s.progress[p.ID] == roadmap.StatusSolved {
			solvedCount[stage]++
		}
	}
	return solvedCount
}

func (s *RoadmapDetailScreen) renderProblemLine(p *roadmap.Problem) string {
	status := s.effectiveStatus(p)
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)

	focused := s.focusIndex < len(s.problems) && s.problems[s.focusIndex].ID == p.ID

	marker := renderStatusPill(s.theme, symbols, status)

	label := fmt.Sprintf("#%d %s", p.ID, p.Title)
	labelStyle := lipgloss.NewStyle()
	if focused {
		labelStyle = lipgloss.NewStyle().
			Foreground(s.theme.SecondaryAccent).
			Bold(true)
	}

	line := fmt.Sprintf("  %s %s", marker, labelStyle.Render(label))

	if status == roadmap.StatusLocked {
		blocked := s.missingPrerequisites(p)
		if len(blocked) > 0 {
			line += lipgloss.NewStyle().
				Foreground(s.theme.Muted).
				Render("  blocked by " + strings.Join(blocked, ", "))
		}
	}

	return line
}

func (s *RoadmapDetailScreen) effectiveStatus(p *roadmap.Problem) roadmap.Status {
	if status, ok := s.progress[p.ID]; ok && status != "" {
		return status
	}

	locked := false
	for _, prereq := range p.Prerequisites {
		if s.progress[prereq] != roadmap.StatusSolved {
			locked = true
			break
		}
	}
	if locked {
		return roadmap.StatusLocked
	}
	return roadmap.StatusAvailable
}

func (s *RoadmapDetailScreen) renderStatusMarker(status roadmap.Status) string {
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	return lipgloss.NewStyle().Width(14).Render(renderStatusPill(s.theme, symbols, status))
}

func (s *RoadmapDetailScreen) missingPrerequisites(p *roadmap.Problem) []string {
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

func (s *RoadmapDetailScreen) renderMiniBar(percentage float64) string {
	width := 10
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := views.ProgressBar(filled, width, width, "█", "░")
	return fmt.Sprintf("[%s] %.0f%%", lipgloss.NewStyle().Foreground(s.theme.XP).Render(bar), percentage)
}

func (s *RoadmapDetailScreen) renderFooter() string {
	viewHint := "graph"
	if s.viewMode == rdViewGraph {
		viewHint = "list"
	}

	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("enter") + " problem",
		s.theme.Key.Render("g") + " " + viewHint,
		s.theme.Key.Render("t") + " theme",
		s.theme.Key.Render("esc") + " dashboard",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}

func sliceIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return 0
}
