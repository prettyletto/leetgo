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
	if dash.width >= 118 {
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
	assert.NotContains(t, view, "Current Stage:")
}

func TestDashboardWidth118_UsesHorizontalLayout(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 118, Height: 35})
	require.NotNil(t, updated)
	dash := updated.(*DashboardScreen)

	view := dash.View()
	assertResponsiveWidth(t, view, 118)

	mainWidth := dash.width - 10 - 42 - 2
	center := dash.renderCenter(mainWidth)
	sidebar := dash.renderSidebar(42)

	assert.LessOrEqual(t, maxRenderedLineWidth(center), mainWidth)
	assert.LessOrEqual(t, maxRenderedLineWidth(sidebar), 42)
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Profile")
}

func TestDashboardWidth119_UsesHorizontalLayout(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 119, Height: 35})
	require.NotNil(t, updated)
	dash := updated.(*DashboardScreen)

	view := dash.View()
	assertResponsiveWidth(t, view, 119)

	mainWidth := dash.width - 10 - 42 - 2
	center := dash.renderCenter(mainWidth)
	sidebar := dash.renderSidebar(42)

	assert.LessOrEqual(t, maxRenderedLineWidth(center), mainWidth)
	assert.LessOrEqual(t, maxRenderedLineWidth(sidebar), 42)
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Profile")
}

func TestDashboardMediumWidth_HidesRoadmapContext(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 100, Height: 26})
	require.NotNil(t, updated)
	dash := updated.(*DashboardScreen)

	sidebar := dash.renderSidebar(dash.width - 10)

	assert.Contains(t, sidebar, "Profile")
	assert.Contains(t, sidebar, "Streak:")
	assert.NotContains(t, sidebar, "Current Stage:")
}

func TestDashboardWideWidth_ShowsRoadmapContext(t *testing.T) {
	dashboard, _ := newTestDashboard(t)
	updated, _ := dashboard.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	require.NotNil(t, updated)
	dash := updated.(*DashboardScreen)

	sidebar := dash.renderSidebar(42)

	assert.Contains(t, sidebar, "Profile")
	assert.Contains(t, sidebar, "Current Stage:")
}
