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
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

type RoadmapCompletionScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap
	width   int
	height  int
}

func NewRoadmapCompletionScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap) *RoadmapCompletionScreen {
	return &RoadmapCompletionScreen{cfg: cfg, theme: theme, db: db, roadmap: rm}
}

func (s *RoadmapCompletionScreen) Init() tea.Cmd { return nil }

func (s *RoadmapCompletionScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateMsg:
		return s, nil
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "esc", "backspace", "d":
			return s, func() tea.Msg { return NavigateMsg{ScreenID: ScreenDashboard} }
		case "r":
			return s, func() tea.Msg { return NavigateMsg{ScreenID: ScreenRoadmapDetail} }
		}
	}
	return s, nil
}

func (s *RoadmapCompletionScreen) View() string {
	ctx := context.Background()
	stats, _ := s.db.GetStats(ctx)
	progress, _ := s.db.GetAllProgress(ctx)
	provenance, _ := s.db.GetSolveProvenanceAll(ctx)
	cycles, _ := s.db.GetReviewCycles(ctx)

	total := len(s.roadmap.Graph.Problems)
	solved := 0
	for id := range s.roadmap.Graph.Problems {
		if progress[id] == roadmap.StatusSolved {
			solved++
		}
	}

	accepted, manual := 0, 0
	for id := range s.roadmap.Graph.Problems {
		if sp := provenance[id]; sp != nil {
			if sp.Kind == "accepted" {
				accepted++
			} else if sp.Kind == "manual" {
				manual++
			}
		}
	}

	activeReviews := 0
	for _, rc := range cycles {
		if rc.CompletedAt == nil {
			activeReviews++
		}
	}

	var lines []string
	lines = append(lines, s.theme.Title.Render("Roadmap Completion"))
	reward := views.RewardMoment{
		Title:   "Roadmap Completion",
		Subject: s.roadmap.Title,
		Actions: []string{"r roadmap", "esc dashboard", "q quit"},
	}
	if stats != nil {
		reward.XP = stats.TotalXP
	}
	if len(s.roadmap.NextRoadmaps) > 0 {
		reward.Next = s.roadmap.NextRoadmaps[0]
	}
	if solved == total {
		reward.Reward = "Roadmap complete"
	} else {
		reward.Reward = "Roadmap in progress"
	}
	reward.AdditionalHighlights = []string{
		fmt.Sprintf("Problems Solved: %d/%d", solved, total),
		fmt.Sprintf("Accepted Solves: %d", accepted),
		fmt.Sprintf("Manual Solves: %d", manual),
		fmt.Sprintf("Total Solve Duration: %s", s.solveDuration(ctx).Round(time.Minute)),
		fmt.Sprintf("Active Review Cycles: %d", activeReviews),
	}
	lines = append(lines, views.RenderRewardMoment(reward, viewPalette(s.theme)))
	lines = append(lines, "")
	if solved == total {
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Success).Render("Roadmap complete"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Warning).Render("Roadmap in progress"))
	}
	lines = append(lines, fmt.Sprintf("Problems Solved: %d/%d", solved, total))
	lines = append(lines, fmt.Sprintf("Accepted Solves: %d", accepted))
	lines = append(lines, fmt.Sprintf("Manual Solves: %d", manual))
	if stats != nil {
		lines = append(lines, fmt.Sprintf("Total XP: %d", stats.TotalXP))
	}
	lines = append(lines, fmt.Sprintf("Total Solve Duration: %s", s.solveDuration(ctx).Round(time.Minute)))
	lines = append(lines, fmt.Sprintf("Active Review Cycles: %d", activeReviews))

	if strongest := s.strongestCategory(progress); strongest != "" {
		lines = append(lines, fmt.Sprintf("Strongest Category: %s", strongest))
	}
	if len(s.roadmap.NextRoadmaps) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Suggested next Roadmap: %s", s.roadmap.NextRoadmaps[0]))
	}
	if manual > 0 {
		lines = append(lines, "")
		lines = append(lines, "Recommended next: validate Manual Solves with Accepted Submissions for confidence and XP.")
	}
	if activeReviews > 0 {
		lines = append(lines, "Recommended review: finish active Review Cycles before starting harder Roadmaps.")
	}

	footer := s.theme.Footer.PaddingTop(1).Render(strings.Join([]string{
		s.theme.Key.Render("r") + " roadmap",
		s.theme.Key.Render("esc") + " dashboard",
		s.theme.Key.Render("q") + " quit",
	}, "  "))
	return strings.Join(lines, "\n") + "\n" + footer
}

func (s *RoadmapCompletionScreen) solveDuration(ctx context.Context) time.Duration {
	var total time.Duration
	for id := range s.roadmap.Graph.Problems {
		attempts, err := s.db.GetAttempts(ctx, id)
		if err != nil {
			continue
		}
		for _, a := range attempts {
			total += a.Duration
		}
	}
	return total
}

func (s *RoadmapCompletionScreen) strongestCategory(progress map[int]roadmap.Status) string {
	counts := make(map[roadmap.Category]int)
	for id, p := range s.roadmap.Graph.Problems {
		if progress[id] == roadmap.StatusSolved {
			counts[p.Category]++
		}
	}
	type entry struct {
		category roadmap.Category
		count    int
	}
	entries := make([]entry, 0, len(counts))
	for category, count := range counts {
		entries = append(entries, entry{category: category, count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].category < entries[j].category
	})
	if len(entries) == 0 {
		return ""
	}
	return string(entries[0].category)
}
