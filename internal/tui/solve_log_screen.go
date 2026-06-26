// Package tui implements the terminal user interface for leetgo.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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
	provenance map[int]*store.SolveProvenance
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

	provenance, err := s.db.GetSolveProvenanceAll(context.Background())
	if err == nil {
		s.provenance = provenance
	}
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

		}
	}
	return s, nil
}

func (s *SolveLogScreen) View() string {
	header := renderScreenHeader(s.theme, "Practice Log", "Recent submissions and manual solves across your workspace.")
	var body string

	if len(s.logs) == 0 && len(s.provenance) == 0 {
		body = renderThemedPanel(s.theme, "History", strings.Join([]string{
			lipgloss.NewStyle().
				Foreground(s.theme.Muted).
				Render("No entries yet."),
			lipgloss.NewStyle().
				Foreground(s.theme.Muted).
				Render("Practice Log entries are created when you submit solutions or manually solve problems."),
		}, "\n\n"), false)
	} else {
		var sections []string
		if len(s.logs) > 0 {
			sections = append(sections, renderThemedPanel(s.theme, "Summary", fmt.Sprintf("Recent submissions: %d", len(s.logs)), false))
		}

		var provenanceLines []string
		for _, sp := range s.provenance {
			if sp.Kind == "manual" {
				provenanceLines = append(provenanceLines, s.renderProvenanceLine(sp))
			}
		}
		if len(provenanceLines) > 0 {
			sections = append(sections, renderThemedPanel(s.theme, "Manual Solves", strings.Join(provenanceLines, "\n"), false))
		}

		var logLines []string
		for i, log := range s.logs {
			line := s.renderLogLine(log, i == s.focusIndex)
			logLines = append(logLines, line)
		}
		sections = append(sections, renderThemedPanel(s.theme, "Submissions", strings.Join(logLines, "\n"), false))
		body = strings.Join(sections, "\n\n")
	}

	footer := s.renderFooter()
	return renderScreenShell(s.theme, s.width, s.height, header, body, footer)
}

func (s *SolveLogScreen) renderProvenanceLine(sp *store.SolveProvenance) string {
	when := sp.SolvedAt.Format("2006-01-02 15:04")
	kindLabel := "Manual Solve"
	if sp.Note != "" {
		kindLabel += fmt.Sprintf(" (%s)", sp.Note)
	}
	line := fmt.Sprintf("  %s  #%d %s", when, sp.ProblemID, kindLabel)
	return lipgloss.NewStyle().
		Foreground(s.theme.Review).
		Render(line)
}

func (s *SolveLogScreen) renderLogLine(log *store.SolveLogRecord, focused bool) string {
	when := log.SubmittedAt.Format("2006-01-02 15:04")

	var result string
	if log.StatusCode == 10 {
		result = fmt.Sprintf("✓ Accepted · %s · %s", log.Runtime, log.Memory)
	} else {
		result = fmt.Sprintf("! %s (%d/%d tests)", log.Status, log.PassedTests, log.TotalTests)
	}

	solveLabel := ""
	if sp, ok := s.provenance[log.ProblemID]; ok && sp.Kind == "accepted" && sp.SolveLogID != nil && *sp.SolveLogID == log.ID {
		solveLabel = " [Accepted Solve]"
	}

	prefix := fmt.Sprintf("#%d %s", log.ProblemID, log.Slug)
	line := fmt.Sprintf("  %s  %s  %s  %s%s", when, prefix, log.Language, result, solveLabel)

	if focused {
		return renderSelectableBlock(s.theme, true, line)
	}
	return lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Render(renderSelectableBlock(s.theme, false, line))
}

func (s *SolveLogScreen) renderFooter() string {
	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("esc") + " dashboard",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  •  "))
}

type ProblemLogEntry struct {
	Timestamp time.Time
	Kind      string
	Detail    string
}

func BuildPracticeLog(db store.Store, problemID int) []ProblemLogEntry {
	ctx := context.Background()

	var entries []ProblemLogEntry

	attempts, err := db.GetAttempts(ctx, problemID)
	if err == nil {
		for _, a := range attempts {
			kind := "Local Attempt passed"
			if !a.Passed {
				kind = "Local Attempt failed"
			}
			detail := fmt.Sprintf("Duration: %s", a.Duration.Round(time.Second))
			if a.SelfReported != "" {
				detail += fmt.Sprintf(" · Self-Reported: %s", a.SelfReported)
			}
			entries = append(entries, ProblemLogEntry{
				Timestamp: a.Timestamp,
				Kind:      kind,
				Detail:    detail,
			})
		}
	}

	logs, err := db.GetSolveLogsForProblem(ctx, problemID)
	if err == nil {
		for _, l := range logs {
			kind := fmt.Sprintf("Submission Attempt %s", l.Status)
			if l.StatusCode == 10 {
				kind = "Submission Attempt accepted"
			}
			detail := fmt.Sprintf("Language: %s · %d/%d tests", l.Language, l.PassedTests, l.TotalTests)
			if l.Runtime != "" {
				detail += fmt.Sprintf(" · Runtime: %s", l.Runtime)
			}
			if l.Memory != "" {
				detail += fmt.Sprintf(" · Memory: %s", l.Memory)
			}
			entries = append(entries, ProblemLogEntry{
				Timestamp: l.SubmittedAt,
				Kind:      kind,
				Detail:    detail,
			})
		}
	}

	sp, err := db.GetSolveProvenance(ctx, problemID)
	if err == nil && sp != nil {
		kind := "Accepted Solve"
		if sp.Kind == "manual" {
			kind = "Manual Solve"
		}
		detail := ""
		if sp.Note != "" {
			detail = sp.Note
		}
		entries = append(entries, ProblemLogEntry{
			Timestamp: sp.SolvedAt,
			Kind:      kind,
			Detail:    detail,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	if len(entries) > 3 {
		entries = entries[:3]
	}

	return entries
}
