package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

type ambientTickMsg time.Time

type RootModel struct {
	cfg           *config.Config
	screen        Screen
	legacyModel   *Model
	languages     []string
	roadmaps      []*roadmap.Roadmap
	activeRoadmap *roadmap.Roadmap
	db            store.Store
	theme         *Theme
	notifications *views.NotificationManager
	ambientFrame  int
	width         int
	height        int
}

func NewRootModel(cfg *config.Config, legacyModel *Model, db store.Store, languages []string, roadmaps []*roadmap.Roadmap) (*RootModel, error) {
	theme, err := LookupTheme(cfg.Theme)
	if err != nil {
		return nil, err
	}

	activeRoadmap, err := catalog.LoadRoadmap(cfg.Roadmap)
	if err != nil {
		activeRoadmap = legacyModel.roadmap
	}

	var scr Screen
	if cfg.ReadyForDashboard(languages, roadmapIDsFromRoadmaps(roadmaps)) {
		scr = NewDashboardScreen(cfg, theme, db, activeRoadmap)
	} else {
		scr = NewOnboardingScreen(cfg, languages, roadmaps, theme)
	}

	return &RootModel{
		cfg:           cfg,
		screen:        scr,
		legacyModel:   legacyModel,
		db:            db,
		languages:     languages,
		roadmaps:      roadmaps,
		activeRoadmap: activeRoadmap,
		theme:         theme,
		notifications: views.NewNotificationManager(),
	}, nil
}

func roadmapIDsFromRoadmaps(roadmaps []*roadmap.Roadmap) []string {
	ids := make([]string, 0, len(roadmaps))
	for _, rm := range roadmaps {
		if rm != nil {
			ids = append(ids, rm.ID)
		}
	}
	return ids
}

func (m *RootModel) Config() *config.Config {
	return m.cfg
}

func (m *RootModel) Theme() string {
	return m.theme.ID
}

func (m *RootModel) ThemeTokens() *Theme {
	return m.theme
}

func (m *RootModel) Notify(msg string) {
	m.notifications.Add(msg)
}

func (m *RootModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.screen.Init()}
	if m.theme.HasAmbientMotion {
		cmds = append(cmds, ambientTickCmd())
	}
	return tea.Batch(cmds...)
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateMsg:
		return m.handleNavigation(msg)

	case GlobalNotificationMsg:
		m.notifications.Add(msg.Message)
		return m, nil

	case ambientTickMsg:
		m.ambientFrame = (m.ambientFrame + 1) % 60
		var cmd tea.Cmd
		if m.theme.HasAmbientMotion {
			cmd = ambientTickCmd()
		}
		return m, cmd

	case ThemeChangedMsg:
		m2, cmd := m.handleThemeChange(msg)
		if m2 != nil {
			if m.theme.HasAmbientMotion {
				cmd = tea.Batch(cmd, ambientTickCmd())
			}
		}
		return m2, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	newScreen, cmd := m.screen.Update(msg)
	if newScreen != nil {
		m.screen = newScreen
	}
	return m, cmd
}

func (m *RootModel) View() string {
	screenView := m.screen.View()

	if m.theme.HasAmbientMotion {
		screenView = m.renderAmbientBorder(screenView)
	}

	notification := m.notifications.Render()
	if notification != "" {
		screenView += "\n" + notification
	}

	return screenView
}

func (m *RootModel) renderAmbientBorder(content string) string {
	phase := m.ambientFrame
	char := "▓"
	if phase%20 < 10 {
		char = "░"
	}

	edgeColor := m.theme.Border
	edge := lipgloss.NewStyle().
		Foreground(edgeColor).
		Render(strings.Repeat(char, 4))

	return edge + "\n" + content + "\n" + edge
}

func (m *RootModel) handleNavigation(msg NavigateMsg) (tea.Model, tea.Cmd) {
	m.reloadActiveRoadmap()

	switch msg.ScreenID {
	case ScreenOnboarding:
		m.refreshTheme()
		m.screen = NewOnboardingScreen(m.cfg, m.languages, m.roadmaps, m.theme)
		m.applyCurrentSize()
		return m, nil
	case ScreenLegacyList:
		if _, ok := m.screen.(*LegacyProblemListScreen); !ok {
			m.screen = NewLegacyProblemListScreen(m.legacyModel)
			m.applyCurrentSize()
		}
		return m, nil
	case ScreenDashboard:
		m.refreshTheme()
		m.screen = NewDashboardScreen(m.cfg, m.theme, m.db, m.activeRoadmap)
		m.applyCurrentSize()
		return m, nil
	case ScreenRoadmapDetail:
		m.refreshTheme()
		m.screen = NewRoadmapDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap)
		m.applyCurrentSize()
		return m, nil
	case ScreenStageDetail:
		m.refreshTheme()
		m.screen = NewStageDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap, msg.Stage)
		m.applyCurrentSize()
		return m, nil
	case ScreenProblemDetail:
		m.refreshTheme()
		m.screen = NewProblemDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap, msg.ProblemID)
		m.applyCurrentSize()
		return m, nil
	case ScreenSolveLog:
		m.refreshTheme()
		m.screen = NewSolveLogScreen(m.cfg, m.theme, m.db)
		m.applyCurrentSize()
		return m, nil
	}
	return m, nil
}

func (m *RootModel) applyCurrentSize() {
	if m.screen == nil || m.width <= 0 || m.height <= 0 {
		return
	}
	newScreen, _ := m.screen.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	if newScreen != nil {
		m.screen = newScreen
	}
}

func (m *RootModel) handleThemeChange(msg ThemeChangedMsg) (tea.Model, tea.Cmd) {
	theme, err := LookupTheme(msg.ThemeID)
	if err != nil {
		return m, nil
	}

	m.theme = theme
	m.cfg.Theme = msg.ThemeID

	if err := m.cfg.Save(); err != nil {
		m.notifications.Add(fmt.Sprintf("Failed to save theme: %v", err))
	}

	if _, ok := m.screen.(*DashboardScreen); ok {
		m.screen = NewDashboardScreen(m.cfg, m.theme, m.db, m.activeRoadmap)
	} else if _, ok := m.screen.(*OnboardingScreen); ok {
		m.screen = NewOnboardingScreen(m.cfg, m.languages, m.roadmaps, m.theme)
	} else if _, ok := m.screen.(*RoadmapDetailScreen); ok {
		m.screen = NewRoadmapDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap)
	} else if sd, ok := m.screen.(*StageDetailScreen); ok {
		m.screen = NewStageDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap, sd.stageID)
	} else if pd, ok := m.screen.(*ProblemDetailScreen); ok {
		m.screen = NewProblemDetailScreen(m.cfg, m.theme, m.db, m.activeRoadmap, pd.problem.ID)
	} else if _, ok := m.screen.(*SolveLogScreen); ok {
		m.screen = NewSolveLogScreen(m.cfg, m.theme, m.db)
	}
	m.applyCurrentSize()

	return m, nil
}

func (m *RootModel) refreshTheme() {
	theme, err := LookupTheme(m.cfg.Theme)
	if err == nil {
		m.theme = theme
	}
}

func (m *RootModel) reloadActiveRoadmap() {
	rm, err := catalog.LoadRoadmap(m.cfg.Roadmap)
	if err == nil {
		m.activeRoadmap = rm
	}
}

func ambientTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return ambientTickMsg(t)
	})
}
