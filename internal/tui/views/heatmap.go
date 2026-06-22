package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	heatmapEmpty  = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	heatmapLow    = lipgloss.NewStyle().Foreground(lipgloss.Color("22"))
	heatmapMedium = lipgloss.NewStyle().Foreground(lipgloss.Color("64"))
	heatmapHigh   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	heatmapLabel  = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Width(3)
	heatmapHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Bold(true)
	heatmapToday  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	heatmapStats  = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

type HeatmapView struct {
	days   []time.Time
	width  int
	height int
}

func NewHeatmapView(days []time.Time) *HeatmapView {
	return &HeatmapView{
		days:   days,
		width:  80,
		height: 24,
	}
}

func (v *HeatmapView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *HeatmapView) Render() string {
	daySet := make(map[string]bool)
	for _, d := range v.days {
		daySet[d.Format("2006-01-02")] = true
	}

	today := time.Now()
	startDate := today.AddDate(0, -6, 0)
	startDate = startOfSunday(startDate)

	var lines []string

	monthLabels := v.renderMonthLabels(startDate, today)
	lines = append(lines, monthLabels)

	dayNames := []string{"Mon", "Wed", "Fri"}
	dayIndices := []int{1, 3, 5}
	header := "   "
	for _, name := range dayNames {
		header += heatmapLabel.Render(name) + "    "
	}
	lines = append(lines, heatmapHeader.Render(header))

	weeks := v.computeWeeks(startDate, today)

	for row := 0; row < 7; row++ {
		var cells []string
		dayLabel := ""
		for i, idx := range dayIndices {
			if row == 0 && i == 0 {
				continue
			}
			if idx == row {
				dayLabel = dayNames[i]
				break
			}
		}
		cells = append(cells, heatmapLabel.Render(dayLabel))

		for _, week := range weeks {
			if row < len(week) {
				day := week[row]
				if day.After(today) {
					cells = append(cells, " ")
				} else if daySet[day.Format("2006-01-02")] {
					if day.YearDay() == today.YearDay() && day.Year() == today.Year() {
						cells = append(cells, heatmapToday.Render("■"))
					} else {
						cells = append(cells, heatmapHigh.Render("■"))
					}
				} else {
					cells = append(cells, heatmapEmpty.Render("·"))
				}
			} else {
				cells = append(cells, " ")
			}
		}
		lines = append(lines, strings.Join(cells, ""))
	}

	lines = append(lines, "")
	lines = append(lines, v.renderLegend())
	lines = append(lines, v.renderStats())

	return strings.Join(lines, "\n")
}

func (v *HeatmapView) renderMonthLabels(start, end time.Time) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	var parts []string
	current := start
	lastMonth := -1

	for current.Before(end) || current.Equal(end) {
		if current.Month() != time.Month(lastMonth+1) || lastMonth == -1 {
			if current.Month() != time.Month(lastMonth) {
				parts = append(parts, heatmapLabel.Render(months[current.Month()-1]))
				lastMonth = int(current.Month()) - 1
			} else {
				parts = append(parts, "   ")
			}
		} else {
			parts = append(parts, "   ")
		}
		current = current.AddDate(0, 0, 7)
	}

	return "   " + strings.Join(parts, "")
}

func (v *HeatmapView) computeWeeks(start, end time.Time) [][]time.Time {
	var weeks [][]time.Time
	current := start

	for current.Before(end) || current.Equal(end) {
		var week []time.Time
		for i := 0; i < 7; i++ {
			week = append(week, current)
			current = current.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}

	return weeks
}

func (v *HeatmapView) renderLegend() string {
	return heatmapStats.Render("Less ") +
		heatmapEmpty.Render("·") + " " +
		heatmapHigh.Render("■") +
		heatmapStats.Render(" More")
}

func (v *HeatmapView) renderStats() string {
	total := len(v.days)
	current := v.currentStreak()
	longest := v.longestStreak()

	return fmt.Sprintf(
		"\n%s  Total days: %d  |  Current streak: %d  |  Longest streak: %d",
		heatmapStats.Render("📊"),
		total, current, longest,
	)
}

func (v *HeatmapView) currentStreak() int {
	if len(v.days) == 0 {
		return 0
	}

	streak := 0
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	daySet := make(map[string]bool)
	for _, d := range v.days {
		daySet[d.Format("2006-01-02")] = true
	}

	if !daySet[today] && !daySet[yesterday] {
		return 0
	}

	check := time.Now()
	if !daySet[today] {
		check = check.AddDate(0, 0, -1)
	}

	for daySet[check.Format("2006-01-02")] {
		streak++
		check = check.AddDate(0, 0, -1)
	}

	return streak
}

func (v *HeatmapView) longestStreak() int {
	if len(v.days) == 0 {
		return 0
	}

	daySet := make(map[string]bool)
	for _, d := range v.days {
		daySet[d.Format("2006-01-02")] = true
	}

	longest := 0
	current := 0

	sorted := make([]time.Time, len(v.days))
	copy(sorted, v.days)

	for i, day := range sorted {
		if i == 0 {
			current = 1
		} else {
			prev := sorted[i-1]
			if day.Sub(prev).Hours() < 48 {
				current++
			} else {
				current = 1
			}
		}
		if current > longest {
			longest = current
		}
	}

	return longest
}

func startOfSunday(t time.Time) time.Time {
	for t.Weekday() != time.Sunday {
		t = t.AddDate(0, 0, -1)
	}
	return t
}
