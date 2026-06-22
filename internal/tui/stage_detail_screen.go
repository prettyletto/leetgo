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
)

type StageDetailScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap
	stageID string

	focusIndex int
	problems   []*roadmap.Problem
	progress   map[int]roadmap.Status

	width  int
	height int
}

func NewStageDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap, stageID string) *StageDetailScreen {
	progress, _ := db.GetAllProgress(context.Background())
	if progress == nil {
		progress = make(map[int]roadmap.Status)
	}

	s := &StageDetailScreen{
		cfg:      cfg,
		theme:    theme,
		db:       db,
		roadmap:  rm,
		stageID:  stageID,
		progress: progress,
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
				return NavigateMsg{ScreenID: ScreenRoadmapDetail}
			}

		case "enter":
			if s.focusIndex < len(s.problems) {
				p := s.problems[s.focusIndex]
				return s, func() tea.Msg {
					return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: p.ID}
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

func (s *StageDetailScreen) View() string {
	stageTitle := s.stageTitle()
	stageDesc := s.stageDescription()

	var lines []string
	lines = append(lines, s.theme.Title.Render(stageTitle))

	if stageDesc != "" {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render(stageDesc))
	}

	solved, total := s.completionCount()
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s  %d/%d solved",
		s.theme.Key.Render("Completion"),
		solved, total,
	))

	if total > 0 {
		percentage := float64(solved) / float64(total) * 100
		bar := s.renderProgressBar(percentage)
		lines = append(lines, bar)
	}

	lines = append(lines, "")
	lines = append(lines, s.theme.Key.Render("Problems"))

	for i, p := range s.problems {
		line := s.renderProblemLine(p, i == s.focusIndex)
		lines = append(lines, line)
	}

	if len(s.problems) == 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render("  No problems in this stage."))
	}

	footer := s.renderFooter()
	return strings.Join(lines, "\n") + "\n" + footer
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

	bar := lipgloss.NewStyle().
		Foreground(s.theme.Success).
		Render(strings.Repeat("█", filled))
	bg := lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Render(strings.Repeat("░", width-filled))

	return fmt.Sprintf("[%s%s] %.0f%%", bar, bg, percentage)
}

func (s *StageDetailScreen) renderProblemLine(p *roadmap.Problem, focused bool) string {
	status := s.effectiveStatus(p)
	marker := s.renderStatusMarker(status)

	label := fmt.Sprintf("#%d %s · %s", p.ID, p.Title, p.Difficulty)
	labelStyle := lipgloss.NewStyle()

	if focused {
		labelStyle = lipgloss.NewStyle().
			Foreground(s.theme.SecondaryAccent).
			Bold(true)
	}

	if status == roadmap.StatusInProgress {
		labelStyle = labelStyle.Copy().
			Background(s.theme.Warning).
			Foreground(lipgloss.Color("0"))
	}

	if status == roadmap.StatusAvailable && !focused {
		labelStyle = labelStyle.Copy().
			Foreground(s.theme.PrimaryAccent)
	}

	line := fmt.Sprintf("  %s %s", marker, labelStyle.Render(label))

	if status == roadmap.StatusLocked {
		blocked := s.missingPrerequisites(p)
		if len(blocked) > 0 {
			line += "  " + lipgloss.NewStyle().
				Foreground(s.theme.Muted).
				Render("blocked by "+strings.Join(blocked, ", "))
		}
	}

	return line
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
	width := 12
	switch status {
	case roadmap.StatusSolved:
		return lipgloss.NewStyle().
			Foreground(s.theme.Success).
			Width(width).
			Render("[SOLVED]")
	case roadmap.StatusInProgress:
		return lipgloss.NewStyle().
			Foreground(s.theme.Warning).
			Bold(true).
			Width(width).
			Render("[ACTIVE]")
	case roadmap.StatusAvailable:
		return lipgloss.NewStyle().
			Foreground(s.theme.PrimaryAccent).
			Bold(true).
			Width(width).
			Render("[READY]")
	default:
		return lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Width(width).
			Render("[LOCKED]")
	}
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
	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("enter") + " problem",
		s.theme.Key.Render("t") + " theme",
		s.theme.Key.Render("esc") + " roadmap",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}
