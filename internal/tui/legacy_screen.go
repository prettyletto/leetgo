package tui

import tea "github.com/charmbracelet/bubbletea"

type LegacyProblemListScreen struct {
	model *Model
}

func NewLegacyProblemListScreen(model *Model) *LegacyProblemListScreen {
	return &LegacyProblemListScreen{model: model}
}

func (s *LegacyProblemListScreen) Init() tea.Cmd {
	return s.model.Init()
}

func (s *LegacyProblemListScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if _, ok := msg.(NavigateMsg); ok {
		return s, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "l", "esc", "backspace":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenDashboard}
			}
		}
	}

	newModel, cmd := s.model.Update(msg)
	if updated, ok := newModel.(*Model); ok {
		s.model = updated
	}
	return s, cmd
}

func (s *LegacyProblemListScreen) View() string {
	return s.model.View()
}
