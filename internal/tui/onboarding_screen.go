package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
)

type onboardingStep int

const (
	stepDisplayName onboardingStep = iota
	stepGitExport
	stepWorkspaceLang
	stepRoadmapCarousel
	stepThemeSelection
)

var onboardingTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("205")).
	Align(lipgloss.Center).
	MarginBottom(1)

var onboardingPromptStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("219")).
	Align(lipgloss.Center).
	MarginTop(1).
	MarginBottom(1)

var onboardingInputStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("238")).
	Padding(0, 1).
	Width(56)

var onboardingFooterStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("245")).
	Align(lipgloss.Center).
	PaddingTop(1)

var onboardingKeyStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("219"))

var onboardingErrorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")).
	Align(lipgloss.Center).
	MarginTop(1)

var onboardingBodyStyle = lipgloss.NewStyle().
	Align(lipgloss.Center)

var onboardingShellStyle = lipgloss.NewStyle().
	Align(lipgloss.Center)

var carouselCardStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	Padding(0, 1).
	Width(40)

var carouselFocusedCardStyle = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder()).
	BorderForeground(lipgloss.Color("219")).
	Padding(0, 1).
	Width(36)

var carouselPreviewCardStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1).
	Width(22)

var carouselStageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("243")).
	Width(28)

type OnboardingScreen struct {
	cfg       *config.Config
	step      onboardingStep
	errorMsg  string
	languages []string
	roadmaps  []*roadmap.Roadmap
	theme     *Theme

	displayNameInput string

	gitExportChoice int
	gitExportRepo   string

	workspaceInput string
	languageIndex  int

	roadmapFocus int

	themeFocus int
	width      int
	height     int
}

func NewOnboardingScreen(cfg *config.Config, languages []string, roadmaps []*roadmap.Roadmap, themes ...*Theme) *OnboardingScreen {
	theme := firstTheme(themes)
	workspaceInput := onboardingWorkspaceDefault(cfg.Workspace)
	s := &OnboardingScreen{
		cfg:              cfg,
		step:             stepDisplayName,
		languages:        languages,
		roadmaps:         roadmaps,
		theme:            theme,
		displayNameInput: cfg.DisplayName,
		workspaceInput:   workspaceInput,
		roadmapFocus:     0,
		themeFocus:       0,
		width:            100,
		height:           30,
	}

	if cfg.GitExportEnabled {
		s.gitExportChoice = 0
		s.gitExportRepo = cfg.GitExportRepo
	} else {
		s.gitExportChoice = 1
		s.gitExportRepo = ""
	}

	for i, lang := range languages {
		if lang == cfg.Language {
			s.languageIndex = i
			break
		}
	}

	for i, rm := range roadmaps {
		if rm.Recommended {
			s.roadmapFocus = i
			break
		}
	}
	for i, rm := range roadmaps {
		if rm.ID == cfg.Roadmap {
			s.roadmapFocus = i
			break
		}
	}

	for i, theme := range config.ValidThemes {
		if theme == cfg.Theme {
			s.themeFocus = i
			break
		}
	}
	return s
}

func onboardingWorkspaceDefault(saved string) string {
	workspace := strings.TrimSpace(saved)
	if workspace != "" && !isGoTestTempWorkspace(workspace) {
		return workspace
	}
	defaultWorkspace, err := config.DefaultWorkspace()
	if err != nil {
		return workspace
	}
	return defaultWorkspace
}

func isGoTestTempWorkspace(path string) bool {
	cleaned := filepath.Clean(path)
	tempDir := filepath.Clean(os.TempDir())
	rel, err := filepath.Rel(tempDir, cleaned)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}
	firstPart := strings.Split(rel, string(os.PathSeparator))[0]
	return strings.HasPrefix(firstPart, "Test")
}

func firstTheme(themes []*Theme) *Theme {
	if len(themes) > 0 && themes[0] != nil {
		return themes[0]
	}
	theme, _ := LookupTheme("rpg-skill-tree")
	return theme
}

func (s *OnboardingScreen) Init() tea.Cmd {
	return nil
}

func (s *OnboardingScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateMsg:
		return s, nil
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return s, tea.Quit
		case "q":
			if s.canQuitWithQ() {
				return s, tea.Quit
			}
		case "esc":
			if s.step > 0 {
				s.step--
				s.errorMsg = ""
			}
			return s, nil
		case "enter":
			return s.handleNext()
		}

		switch s.step {
		case stepDisplayName:
			s.handleDisplayNameKey(msg)
		case stepGitExport:
			s.handleGitExportKey(msg)
		case stepWorkspaceLang:
			s.handleWorkspaceLangKey(msg)
		case stepRoadmapCarousel:
			s.handleRoadmapKey(msg)
		case stepThemeSelection:
			s.handleThemeKey(msg)
		}
		return s, nil
	}
	return s, nil
}

func (s *OnboardingScreen) canQuitWithQ() bool {
	switch s.step {
	case stepDisplayName, stepWorkspaceLang:
		return false
	case stepGitExport:
		return s.gitExportChoice != 0
	default:
		return true
	}
}

func (s *OnboardingScreen) roadmapIDs() []string {
	ids := make([]string, len(s.roadmaps))
	for i, rm := range s.roadmaps {
		ids[i] = rm.ID
	}
	return ids
}

func (s *OnboardingScreen) handleNext() (Screen, tea.Cmd) {
	s.errorMsg = ""
	switch s.step {
	case stepDisplayName:
		name := strings.TrimSpace(s.displayNameInput)
		if name == "" {
			s.errorMsg = "Display name is required."
			return s, nil
		}
		if len(name) > config.MaxDisplayNameLength {
			name = name[:config.MaxDisplayNameLength]
		}
		s.cfg.DisplayName = name
		s.step = stepGitExport
	case stepGitExport:
		if s.gitExportChoice == 0 {
			repo := strings.TrimSpace(s.gitExportRepo)
			if repo == "" {
				s.errorMsg = "Git repository path is required."
				return s, nil
			}
			if _, err := os.Stat(repo); err != nil {
				s.errorMsg = "Path does not exist."
				return s, nil
			}
			if _, err := os.Stat(filepath.Join(repo, ".git")); err != nil {
				s.errorMsg = "Not a git repository (missing .git)."
				return s, nil
			}
			s.cfg.GitExportEnabled = true
			s.cfg.GitExportRepo = repo
		} else {
			s.cfg.GitExportEnabled = false
			s.cfg.GitExportRepo = ""
		}
		s.step = stepWorkspaceLang
	case stepWorkspaceLang:
		s.cfg.Workspace = strings.TrimSpace(s.workspaceInput)
		if s.cfg.Workspace == "" {
			s.errorMsg = "Workspace path is required."
			return s, nil
		}
		s.cfg.Language = s.languages[s.languageIndex]
		s.step = stepRoadmapCarousel
	case stepRoadmapCarousel:
		s.cfg.Roadmap = s.roadmaps[s.roadmapFocus].ID
		s.step = stepThemeSelection
	case stepThemeSelection:
		s.cfg.Theme = config.ValidThemes[s.themeFocus]
		s.cfg.OnboardingComplete = true
		s.cfg.OnboardingVersion = config.CurrentOnboardingVersion
		if err := s.cfg.Validate(s.languages, s.roadmapIDs()); err != nil {
			s.cfg.OnboardingComplete = false
			s.cfg.OnboardingVersion = 0
			s.errorMsg = fmt.Sprintf("Config validation failed: %v", err)
			return s, nil
		}
		if err := s.cfg.Save(); err != nil {
			s.cfg.OnboardingComplete = false
			s.cfg.OnboardingVersion = 0
			s.errorMsg = fmt.Sprintf("Failed to save config: %v", err)
			return s, nil
		}
		return nil, func() tea.Msg {
			return NavigateMsg{ScreenID: ScreenDashboard}
		}
	}
	return s, nil
}

func (s *OnboardingScreen) handleDisplayNameKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyBackspace:
		if len(s.displayNameInput) > 0 {
			s.displayNameInput = s.displayNameInput[:len(s.displayNameInput)-1]
		}
	case tea.KeySpace:
		if len(s.displayNameInput) < config.MaxDisplayNameLength {
			s.displayNameInput += " "
		}
	default:
		if len(msg.Runes) > 0 && len(s.displayNameInput) < config.MaxDisplayNameLength {
			s.displayNameInput += string(msg.Runes)
		}
	}
}

func (s *OnboardingScreen) handleGitExportKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "ctrl+p":
		if s.gitExportChoice > 0 {
			s.gitExportChoice--
		}
	case "down", "ctrl+n":
		if s.gitExportChoice < 1 {
			s.gitExportChoice++
		}
	case "y", "1":
		s.gitExportChoice = 0
	case "n", "0":
		s.gitExportChoice = 1
	default:
		if s.gitExportChoice == 0 {
			if msg.Type == tea.KeyBackspace {
				if len(s.gitExportRepo) > 0 {
					s.gitExportRepo = s.gitExportRepo[:len(s.gitExportRepo)-1]
				}
			} else if msg.Type == tea.KeySpace || len(msg.Runes) > 0 {
				s.gitExportRepo += string(msg.Runes)
			}
		}
	}
}

func (s *OnboardingScreen) handleWorkspaceLangKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "ctrl+p":
		if s.languageIndex > 0 {
			s.languageIndex--
		}
	case "down", "ctrl+n":
		if s.languageIndex < len(s.languages)-1 {
			s.languageIndex++
		}
	default:
		switch msg.Type {
		case tea.KeyBackspace:
			if len(s.workspaceInput) > 0 {
				s.workspaceInput = s.workspaceInput[:len(s.workspaceInput)-1]
			}
		case tea.KeySpace:
			s.workspaceInput += " "
		default:
			if len(msg.Runes) > 0 {
				s.workspaceInput += string(msg.Runes)
			}
		}
	}
}

func (s *OnboardingScreen) handleRoadmapKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "left", "<", "h":
		s.roadmapFocus--
		if s.roadmapFocus < 0 {
			s.roadmapFocus = len(s.roadmaps) - 1
		}
	case "right", ">", "l":
		s.roadmapFocus++
		if s.roadmapFocus >= len(s.roadmaps) {
			s.roadmapFocus = 0
		}
	}
}

func (s *OnboardingScreen) handleThemeKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "left", "<", "h":
		if s.themeFocus > 0 {
			s.themeFocus--
		}
	case "right", ">", "l":
		if s.themeFocus < len(config.ValidThemes)-1 {
			s.themeFocus++
		}
	}
}

func (s *OnboardingScreen) View() string {
	var title string
	switch s.step {
	case stepDisplayName:
		title = "Step 1/5: Who are you, challenger?"
	case stepGitExport:
		title = "Step 2/5: Git Export Backup"
	case stepWorkspaceLang:
		title = "Step 3/5: Workspace & Language"
	case stepRoadmapCarousel:
		title = "Step 4/5: Choose Your Roadmap"
	case stepThemeSelection:
		title = "Step 5/5: Pick a Theme"
	}

	contentWidth := s.contentWidth()
	content := s.theme.Title.Width(contentWidth).MarginBottom(1).Render(title) + "\n"

	switch s.step {
	case stepDisplayName:
		content += s.renderDisplayName()
	case stepGitExport:
		content += s.renderGitExport()
	case stepWorkspaceLang:
		content += s.renderWorkspaceLang()
	case stepRoadmapCarousel:
		content += s.renderRoadmapCarousel()
	case stepThemeSelection:
		content += s.renderThemeSelection()
	}

	if s.errorMsg != "" {
		content += "\n" + lipgloss.NewStyle().Foreground(s.theme.Danger).Align(lipgloss.Center).Width(contentWidth).Render(s.errorMsg)
	}

	content += "\n" + s.renderFooter()
	return lipgloss.Place(
		s.width,
		s.height,
		lipgloss.Center,
		lipgloss.Center,
		onboardingShellStyle.Width(contentWidth).Render(content),
	)
}

func (s *OnboardingScreen) renderDisplayName() string {
	input := onboardingInputStyle.Width(s.formWidth()).Render(s.displayNameInput + "█")
	lines := []string{
		s.promptStyle().Width(s.contentWidth()).Render("Enter your display name:"),
		lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, input),
		"",
		s.theme.Footer.Width(s.contentWidth()).Render(fmt.Sprintf("max %d characters", config.MaxDisplayNameLength)),
	}
	return strings.Join(lines, "\n")
}

func (s *OnboardingScreen) renderGitExport() string {
	yesLabel := "  [Yes, use Git Export backup]"
	noLabel := "  [Not now]"
	if s.gitExportChoice == 0 {
		yesLabel = "> [Yes, use Git Export backup]"
	} else {
		noLabel = "> [Not now]"
	}

	lines := []string{
		s.promptStyle().Width(s.contentWidth()).Render("Do you want Leetgo to back up progress to a Git repo?"),
		lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, yesLabel),
		lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, noLabel),
	}

	if s.gitExportChoice == 0 {
		lines = append(lines, "")
		lines = append(lines, s.promptStyle().Width(s.contentWidth()).Render("Git repository path:"))
		input := onboardingInputStyle.Width(s.formWidth()).Render(s.gitExportRepo + "█")
		lines = append(lines, lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, input))
		lines = append(lines, s.theme.Footer.Width(s.contentWidth()).Render("must contain .git directory"))
	}

	return strings.Join(lines, "\n")
}

func (s *OnboardingScreen) renderWorkspaceLang() string {
	langLines := make([]string, len(s.languages))
	for i, lang := range s.languages {
		marker := "  "
		if i == s.languageIndex {
			marker = "> "
		}
		langLines[i] = marker + lang
	}

	lines := []string{
		s.promptStyle().Width(s.contentWidth()).Render("Workspace path:"),
		lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, onboardingInputStyle.Width(s.formWidth()).Render(s.workspaceInput+"█")),
		"",
		s.promptStyle().Width(s.contentWidth()).Render("Language"),
		lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, strings.Join(langLines, "\n")),
	}
	return strings.Join(lines, "\n")
}

func (s *OnboardingScreen) renderRoadmapCarousel() string {
	total := len(s.roadmaps)

	if total == 1 {
		return s.renderFocusedCard(s.roadmaps[0], total)
	}

	leftIdx := s.roadmapFocus - 1
	if leftIdx < 0 {
		leftIdx = total - 1
	}
	rightIdx := (s.roadmapFocus + 1) % total

	leftCard := s.renderPreviewCard(s.roadmaps[leftIdx])
	centerCard := s.renderFocusedCard(s.roadmaps[s.roadmapFocus], total)
	rightCard := s.renderPreviewCard(s.roadmaps[rightIdx])
	if s.contentWidth() < 88 {
		return lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, centerCard)
	}

	carousel := lipgloss.JoinHorizontal(lipgloss.Center,
		leftCard,
		"  ",
		centerCard,
		"  ",
		rightCard,
	)
	return lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, carousel)
}

func (s *OnboardingScreen) renderFocusedCard(rm *roadmap.Roadmap, total int) string {
	title := rm.Title
	if rm.Recommended {
		title += " " + roadmapMarker(true)
	}

	cardWidth := 32
	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\nProblems: %d | Est. hours: %d",
		title,
		wrapText(rm.Tagline, cardWidth),
		wrapText("Audience: "+rm.Audience, cardWidth),
		wrapText("Promise: "+rm.Promise, cardWidth),
		len(rm.Graph.Problems),
		rm.EstimatedHours,
	)

	if len(rm.DifficultyMix) > 0 {
		mix := formatDifficultyMix(rm.DifficultyMix)
		if mix != "" {
			content += "\nDifficulty: " + mix
		}
	}

	if len(rm.Stages) > 0 {
		content += "\n\nFirst stages:"
		n := len(rm.Stages)
		if n > 3 {
			n = 3
		}
		for i := 0; i < n; i++ {
			content += "\n" + carouselStageStyle.Width(cardWidth).Render("  "+rm.Stages[i].Title)
		}
	}

	if len(rm.Highlights) > 0 {
		content += "\n"
		for _, h := range rm.Highlights {
			content += "\n" + wrapText("  - "+h, cardWidth)
		}
	}

	nav := fmt.Sprintf("< %d/%d >", s.roadmapFocus+1, total)
	content += "\n\n" + nav

	return s.theme.FocusedPanel.Width(36).Render(content)
}

func (s *OnboardingScreen) renderPreviewCard(rm *roadmap.Roadmap) string {
	title := rm.Title
	if rm.Recommended {
		title += " " + roadmapMarker(true)
	}

	content := fmt.Sprintf("%s\n\n%s",
		title,
		wrapText(rm.Tagline, 18),
	)

	nav := fmt.Sprintf("<  >")
	content += "\n\n" + nav

	return s.theme.Panel.Width(22).BorderForeground(s.theme.Muted).Render(content)
}

func formatDifficultyMix(mix map[roadmap.Difficulty]int) string {
	parts := make([]string, 0, 3)
	order := []roadmap.Difficulty{roadmap.DifficultyEasy, roadmap.DifficultyMedium, roadmap.DifficultyHard}
	for _, d := range order {
		if v, ok := mix[d]; ok && v > 0 {
			parts = append(parts, fmt.Sprintf("%s %d%%", d, v))
		}
	}
	return strings.Join(parts, " / ")
}

func (s *OnboardingScreen) renderThemeSelection() string {
	themeLabels := map[string]string{
		"rpg-skill-tree":     "RPG Skill Tree",
		"clean-productivity": "Clean Productivity",
		"cyber-dashboard":    "Cyber Dashboard",
	}
	themeDescs := map[string]string{
		"rpg-skill-tree":     "Strong status colors, bordered cards, reward animations.",
		"clean-productivity": "Minimal color, low motion, high readability.",
		"cyber-dashboard":    "High contrast, neon accents, subtle ambient motion.",
	}

	lines := []string{
		s.promptStyle().Width(s.contentWidth()).Render("Choose your visual style:"),
	}

	for i, theme := range config.ValidThemes {
		marker := "  "
		if i == s.themeFocus {
			marker = "> "
		}
		label := themeLabels[theme]
		line := fmt.Sprintf("%s%s - %s", marker, label, themeDescs[theme])
		lines = append(lines, lipgloss.PlaceHorizontal(s.contentWidth(), lipgloss.Center, wrapText(line, s.formWidth())))
	}

	return strings.Join(lines, "\n")
}

func (s *OnboardingScreen) renderFooter() string {
	var items []string
	switch s.step {
	case stepDisplayName:
		items = []string{
			s.theme.Key.Render("type") + " name",
			s.theme.Key.Render("enter") + " next",
		}
	case stepGitExport:
		items = []string{
			s.theme.Key.Render("up/down") + " choose",
			s.theme.Key.Render("ctrl+n/p") + " choose",
			s.theme.Key.Render("type") + " repo path",
			s.theme.Key.Render("enter") + " next",
		}
	case stepWorkspaceLang:
		items = []string{
			s.theme.Key.Render("type") + " workspace",
			s.theme.Key.Render("up/down") + " language",
			s.theme.Key.Render("ctrl+n/p") + " language",
			s.theme.Key.Render("enter") + " next",
		}
	case stepRoadmapCarousel:
		items = []string{
			s.theme.Key.Render("</>") + " browse",
			s.theme.Key.Render("enter") + " select",
		}
	case stepThemeSelection:
		items = []string{
			s.theme.Key.Render("left/right") + " browse",
			s.theme.Key.Render("enter") + " finish",
		}
	}

	if s.step > 0 {
		items = append(items, s.theme.Key.Render("esc")+" back")
	}
	if s.canQuitWithQ() {
		items = append(items, s.theme.Key.Render("q")+" quit")
	} else {
		items = append(items, s.theme.Key.Render("ctrl+c")+" quit")
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  "))
}

func (s *OnboardingScreen) promptStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.SecondaryAccent).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(1)
}

func roadmapMarker(recommended bool) string {
	if recommended {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220")).
			Render("[RECOMMENDED]")
	}
	return ""
}

func (s *OnboardingScreen) contentWidth() int {
	width := s.width - 4
	if width > 112 {
		return 112
	}
	if width < 48 {
		return 48
	}
	return width
}

func (s *OnboardingScreen) formWidth() int {
	width := s.contentWidth() - 8
	if width > 64 {
		return 64
	}
	if width < 36 {
		return 36
	}
	return width
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
