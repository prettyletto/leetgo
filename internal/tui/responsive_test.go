package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func maxRenderedLineWidth(view string) int {
	maxWidth := 0
	for _, line := range strings.Split(view, "\n") {
		w := lipgloss.Width(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

func renderedLineCount(view string) int {
	return len(strings.Split(strings.TrimRight(view, "\n"), "\n"))
}

func assertResponsiveWidth(t *testing.T, view string, width int) {
	t.Helper()
	assert.LessOrEqual(t, maxRenderedLineWidth(view), width)
}

func TestScreensFitMinimumSupportedWidth(t *testing.T) {
	const width = 60
	const height = 24

	dashboard, _ := newTestDashboard(t)
	roadmapDetail, _ := newTestRoadmapDetail(t)
	stageDetail, _ := newTestStageDetail(t, "arrays-hashing")
	problemDetail, _ := newTestProblemDetail(t, 1)
	solveLog, _ := newTestSolveLog(t)
	onboarding := NewOnboardingScreen(&config.Config{}, testLanguages(), testRoadmaps(t), nil, nil)
	completion := NewRoadmapCompletionScreen(&config.Config{Theme: "adaptive"}, roadmapDetail.theme, roadmapDetail.db, roadmapDetail.roadmap)

	screens := []Screen{dashboard, roadmapDetail, stageDetail, problemDetail, solveLog, onboarding, completion}
	for _, screen := range screens {
		updated, _ := screen.Update(tea.WindowSizeMsg{Width: width, Height: height})
		require.NotNil(t, updated)
		assertResponsiveWidth(t, updated.View(), width)
	}
}

func TestScreenShellUsesWideCenteredLayoutWhenRoomy(t *testing.T) {
	theme, err := LookupTheme("adaptive")
	require.NoError(t, err)

	view := renderScreenShell(theme, 140, 40, "Header", "Body", "Footer")

	assertResponsiveWidth(t, view, 140)
	assert.Contains(t, view, "Header")
}

func TestDashboardSplitPaneWidthUsesFullStackedCards(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 96, Height: 32})
	require.NotNil(t, updated)

	dash := updated.(*DashboardScreen)
	shellWidth := dash.width - 10
	mainWidth := shellWidth
	sidebarWidth := 42
	if dash.width >= 120 {
		mainWidth = shellWidth - sidebarWidth - 2
	} else if dash.width >= 78 {
		sidebarWidth = mainWidth
	}

	assert.Equal(t, 86, mainWidth)
	assert.Equal(t, mainWidth, sidebarWidth)
	assertResponsiveWidth(t, dash.View(), 96)
	assert.NotContains(t, dash.renderFooter(), "practice log")
}

func TestDashboardStackedPanelsFillColumnWidth(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 96, Height: 32})
	require.NotNil(t, updated)
	dash := updated.(*DashboardScreen)

	columnWidth := 86
	center := dash.renderCenter(columnWidth)
	sidebar := dash.renderSidebar(columnWidth)

	assert.Equal(t, columnWidth, maxRenderedLineWidth(center))
	assert.Equal(t, columnWidth, maxRenderedLineWidth(sidebar))
}

func TestDashboardShortSplitPaneFitsHeight(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 96, Height: 28})
	require.NotNil(t, updated)

	view := updated.View()

	assertResponsiveWidth(t, view, 96)
	assert.LessOrEqual(t, renderedLineCount(view), 28)
	assert.Contains(t, view, "Up next")
	assert.Contains(t, view, "Profile")
	assert.Contains(t, view, "Roadmap")
}
