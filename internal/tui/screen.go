package tui

import tea "github.com/charmbracelet/bubbletea"

type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
}

type NavigateMsg struct {
	ScreenID  string
	Stage     string
	ProblemID int
}

type GlobalNotificationMsg struct {
	Message string
}

type ThemeChangedMsg struct {
	ThemeID string
}

const (
	ScreenOnboarding    = "onboarding"
	ScreenLegacyList    = "legacy-list"
	ScreenDashboard     = "dashboard"
	ScreenRoadmapDetail = "roadmap-detail"
	ScreenStageDetail   = "stage-detail"
	ScreenProblemDetail = "problem-detail"
	ScreenSolveLog      = "solve-log"
)
