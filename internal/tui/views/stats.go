package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/store"
)

var (
	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	xpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	levelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	streakStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	progressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	progressBarBG = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type StatsBar struct {
	stats *store.Stats
	width int
}

func NewStatsBar(stats *store.Stats) *StatsBar {
	return &StatsBar{
		stats: stats,
		width: 80,
	}
}

func (s *StatsBar) SetWidth(width int) {
	s.width = width
}

func (s *StatsBar) Render() string {
	if s.stats == nil {
		return ""
	}

	level := levelStyle.Render(fmt.Sprintf("LVL %d", s.stats.Level))
	streak := streakStyle.Render(fmt.Sprintf("🔥 %d", s.stats.Streak))
	solved := statsStyle.Render(fmt.Sprintf("%d/%d solved", s.stats.Solved, s.stats.Total))

	xpBar := s.renderXPBar()

	left := fmt.Sprintf("%s  %s  %s", level, streak, solved)
	right := xpBar

	padding := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}

	return left + strings.Repeat(" ", padding) + right
}

func (s *StatsBar) renderXPBar() string {
	currentLevelXP := store.LevelToXP(s.stats.Level)
	nextLevelXP := store.LevelToXP(s.stats.Level + 1)
	xpInLevel := s.stats.TotalXP - currentLevelXP
	xpNeeded := nextLevelXP - currentLevelXP

	percentage := float64(xpInLevel) / float64(xpNeeded)
	barWidth := 20
	filled := int(percentage * float64(barWidth))

	bar := progressBarStyle.Render(strings.Repeat("█", filled)) +
		progressBarBG.Render(strings.Repeat("░", barWidth-filled))

	xp := xpStyle.Render(fmt.Sprintf("%d XP", s.stats.TotalXP))

	return fmt.Sprintf("%s [%s]", xp, bar)
}
