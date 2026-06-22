package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/store"
)

type SolveLogScreen struct {
	cfg   *config.Config
	theme *Theme
	db    store.Store

	logs       []*store.SolveLogRecord
	focusIndex int

	width  int
	height int
}

func NewSolveLogScreen(cfg *config.Config, theme *Theme, db store.Store) *SolveLogScreen {
	s := &SolveLogScreen{
		cfg:   cfg,
		theme: theme,
		db:    db,
	}
	s.refresh()
	return s
}

func (s *SolveLogScreen) refresh() {
	logs, err := s.db.GetSolveLogs(context.Background())
	if err != nil {
		s.logs = nil
		return
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].SubmittedAt.After(logs[j].SubmittedAt)
	})

	s.logs = logs
}

func (s *SolveLogScreen) Init() tea.Cmd {
	return nil
}

func (s *SolveLogScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
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
				return NavigateMsg{ScreenID: ScreenDashboard}
			}

		case "j", "down":
			if len(s.logs) > 0 {
				s.focusIndex = (s.focusIndex + 1) % len(s.logs)
			}

		case "k", "up":
			if len(s.logs) > 0 {
				s.focusIndex--
				if s.focusIndex < 0 {
					s.focusIndex = len(s.logs) - 1
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

func (s *SolveLogScreen) View() string {
	var lines []string

	lines = append(lines, s.theme.Title.MarginBottom(1).Render("Solve Log"))

	if len(s.logs) == 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render("No Solve Logs yet."))
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			Render("Solve Logs are created when you submit a solution to LeetCode."))
	} else {
		lines = append(lines, fmt.Sprintf("Recent submissions: %d", len(s.logs)))
		lines = append(lines, "")

		for i, log := range s.logs {
			line := s.renderLogLine(log, i == s.focusIndex)
			lines = append(lines, line)
		}
	}

	footer := s.renderFooter()
	return strings.Join(lines, "\n") + "\n" + footer
}

func (s *SolveLogScreen) renderLogLine(log *store.SolveLogRecord, focused bool) string {
	when := log.SubmittedAt.Format("2006-01-02 15:04")

	result := log.Status
	if log.StatusCode == 10 {
		result = fmt.Sprintf("Accepted · %s · %s", log.Runtime, log.Memory)
	} else {
		result = fmt.Sprintf("%s (%d/%d tests)", log.Status, log.PassedTests, log.TotalTests)
	}

	prefix := fmt.Sprintf("#%d %s", log.ProblemID, log.Slug)
	line := fmt.Sprintf("  %s  %s  %s  %s", when, prefix, log.Language, result)

	if focused {
		return lipgloss.NewStyle().
			Foreground(s.theme.SecondaryAccent).
			Bold(true).
			Render(line)
	}
	return lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Render(line)
}

func (s *SolveLogScreen) renderFooter() string {
	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("t") + " theme",
		s.theme.Key.Render("esc") + " dashboard",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}
