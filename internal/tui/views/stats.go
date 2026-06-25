package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/store"
)

type StatsBar struct {
	stats   *store.Stats
	width   int
	palette Palette
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

func (s *StatsBar) SetPalette(p Palette) {
	s.palette = p
}

func (s *StatsBar) Render() string {
	if s.stats == nil {
		return ""
	}

	levelStyle := lipgloss.NewStyle().Foreground(s.palette.Success).Bold(true)
	streakStyle := lipgloss.NewStyle().Foreground(s.palette.Danger).Bold(true)
	statsStyle := lipgloss.NewStyle().Foreground(s.palette.Muted)
	xpStyle := lipgloss.NewStyle().Foreground(s.palette.XP).Bold(true)
	barFillStyle := lipgloss.NewStyle().Foreground(s.palette.XP)
	barEmptyStyle := lipgloss.NewStyle().Foreground(s.palette.Muted)

	level := levelStyle.Render(fmt.Sprintf("LVL %d", s.stats.Level))
	streak := streakStyle.Render(fmt.Sprintf(" %d", s.stats.Streak))
	solved := statsStyle.Render(fmt.Sprintf("%d/%d solved", s.stats.Solved, s.stats.Total))

	xpBar := s.renderXPBar(barFillStyle, barEmptyStyle, xpStyle)

	left := fmt.Sprintf("%s  •  %s  •  %s", level, streak, solved)
	right := xpBar

	padding := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}

	return left + strings.Repeat(" ", padding) + right
}

func (s *StatsBar) renderXPBar(fillStyle, emptyStyle, xpStyle lipgloss.Style) string {
	currentLevelXP := store.LevelToXP(s.stats.Level)
	nextLevelXP := store.LevelToXP(s.stats.Level + 1)
	xpInLevel := s.stats.TotalXP - currentLevelXP
	xpNeeded := nextLevelXP - currentLevelXP

	percentage := float64(xpInLevel) / float64(xpNeeded)
	barWidth := 20
	filled := int(percentage * float64(barWidth))

	bar := fillStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", barWidth-filled))

	xp := xpStyle.Render(fmt.Sprintf("%d XP", s.stats.TotalXP))

	return fmt.Sprintf("%s [%s]", xp, bar)
}
