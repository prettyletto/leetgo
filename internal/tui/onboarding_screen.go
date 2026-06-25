package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/recommendation"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
	"github.com/prettyletto/leetgo/internal/workspace"
)

type onboardingStep int

const (
	stepDisplayName onboardingStep = iota
	stepGitExport
	stepWorkspaceLang
	stepRoadmapCarousel
	stepThemeSelection
	stepSession
	stepCompletion
)

var onboardingInputStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("238")).
	Padding(0, 1).
	Width(56)

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
	db        store.Store

	displayNameInput string

	gitExportChoice int
	gitExportRepo   string

	workspaceInput string
	languageIndex  int

	roadmapFocus int

	themeFocus int
	width      int
	height     int

	sessionChoice int

	nextAction     *recommendation.NextAction
	onboardingDone bool
}

func NewOnboardingScreen(cfg *config.Config, languages []string, roadmaps []*roadmap.Roadmap, db store.Store, activeRoadmap *roadmap.Roadmap, themes ...*Theme) *OnboardingScreen {
	cfg.ApplyDefaults()
	theme := firstTheme(themes)
	workspaceInput := onboardingWorkspaceDefault(cfg.Workspace)
	s := &OnboardingScreen{
		cfg:              cfg,
		step:             stepDisplayName,
		languages:        languages,
		roadmaps:         roadmaps,
		theme:            theme,
		db:               db,
		displayNameInput: cfg.DisplayName,
		workspaceInput:   workspaceInput,
		roadmapFocus:     0,
		sessionChoice:    1,
		width:            128,
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
	theme, _ := LookupTheme("adaptive")
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
		case stepSession:
			s.handleSessionKey(msg)
		case stepCompletion:
			s.handleCompletionKey(msg)
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
	case stepCompletion:
		return s.onboardingDone
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
		s.cfg.ApplyDefaults()
		s.step = stepSession
	case stepSession:
		s.step = stepCompletion
		s.calculateNextAction()
	case stepCompletion:
		if s.onboardingDone {
			return nil, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenDashboard}
			}
		}
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
		s.onboardingDone = true
	}
	return s, nil
}

func (s *OnboardingScreen) calculateNextAction() {
	if s.db == nil {
		return
	}
	rm, err := catalog.LoadRoadmap(s.cfg.Roadmap)
	if err != nil {
		return
	}
	ctx := context.Background()
	calc := recommendation.NewCalculator(s.db, rm)
	actions, err := calc.Calculate(ctx)
	if err != nil || len(actions) == 0 {
		return
	}
	a := actions[0]
	s.nextAction = &a
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
	case "p":
		if s.cfg.SymbolMode == "plain" {
			s.cfg.SymbolMode = "rich"
		} else {
			s.cfg.SymbolMode = "plain"
		}
	}
}

func (s *OnboardingScreen) View() string {
	var title string
	var subtitle string
	switch s.step {
	case stepDisplayName:
		title = "Step 1/6: Who are you, challenger?"
		subtitle = "Create your local profile before the Dashboard opens."
	case stepGitExport:
		title = "Step 2/6: Git Export Backup"
		subtitle = "Optional backup for generated work and progress snapshots."
	case stepWorkspaceLang:
		title = "Step 3/6: Workspace & Language"
		subtitle = "Confirm where Problem files live and which language to generate."
	case stepRoadmapCarousel:
		title = "Step 4/6: Choose Your Roadmap"
		subtitle = "Pick the guided Problem progression you want Leetgo to rank against."
	case stepThemeSelection:
		title = "Appearance Preview"
		subtitle = "Leetgo now uses one adaptive appearance system. Preview symbol modes here."
	case stepSession:
		title = "Step 5/6: LeetCode Session"
		subtitle = "Accepted submissions unlock trusted Solve progress and submission XP."
	case stepCompletion:
		title = "Step 6/6: Ready to Start"
		subtitle = "Review your setup and launch into the guided Dashboard."
	}

	header := renderScreenHeader(s.theme, title, subtitle)
	var body string

	switch s.step {
	case stepDisplayName:
		body = s.renderDisplayName()
	case stepGitExport:
		body = s.renderGitExport()
	case stepWorkspaceLang:
		body = s.renderWorkspaceLang()
	case stepRoadmapCarousel:
		body = s.renderRoadmapCarousel()
	case stepThemeSelection:
		body = s.renderThemeSelection()
	case stepSession:
		body = s.renderSession()
	case stepCompletion:
		body = s.renderCompletion()
	}

	if s.errorMsg != "" {
		errorBody := lipgloss.NewStyle().Foreground(s.theme.Danger).Render(s.errorMsg)
		body += "\n\n" + renderThemedPanel(s.theme, "Validation", errorBody, false)
	}

	return renderScreenShell(s.theme, s.width, s.height, header, body, s.renderFooter())
}

func (s *OnboardingScreen) renderDisplayName() string {
	input := onboardingInputStyle.Width(s.formWidth()).Render(s.displayNameInput + "█")
	body := strings.Join([]string{
		"Enter your display name:",
		"",
		input,
		"",
		s.theme.Subtitle.Render(fmt.Sprintf("Maximum %d characters.", config.MaxDisplayNameLength)),
	}, "\n")
	return renderThemedPanel(s.theme, "Profile", body, true)
}

func (s *OnboardingScreen) renderGitExport() string {
	yesLabel := lipgloss.NewStyle().Foreground(s.theme.Muted).Render("Yes, use Git Export backup")
	noLabel := lipgloss.NewStyle().Foreground(s.theme.Muted).Render("Not now")
	if s.gitExportChoice == 0 {
		yesLabel = renderSelectableBlock(s.theme, true, "Yes, use Git Export backup")
	} else {
		noLabel = renderSelectableBlock(s.theme, true, "Not now")
	}

	lines := []string{"Do you want Leetgo to back up progress to a Git repo?", "", yesLabel, noLabel}

	if s.gitExportChoice == 0 {
		lines = append(lines, "")
		lines = append(lines, "Git repository path:")
		input := onboardingInputStyle.Width(s.formWidth()).Render(s.gitExportRepo + "█")
		lines = append(lines, input)
		lines = append(lines, s.theme.Subtitle.Render("Must contain a .git directory."))
	}

	return renderThemedPanel(s.theme, "Backup", strings.Join(lines, "\n"), true)
}

func (s *OnboardingScreen) renderWorkspaceLang() string {
	langLines := make([]string, len(s.languages))
	for i, lang := range s.languages {
		line := lang
		if i == s.languageIndex {
			langLines[i] = renderSelectableBlock(s.theme, true, line)
			continue
		}
		langLines[i] = lipgloss.NewStyle().Foreground(s.theme.Muted).Render(renderSelectableBlock(s.theme, false, line))
	}

	left := renderThemedPanel(s.theme, "Workspace", strings.Join([]string{
		"Workspace path:",
		"",
		onboardingInputStyle.Width(s.formWidth()).Render(s.workspaceInput + "█"),
	}, "\n"), true)
	right := renderThemedPanel(s.theme, "Language", strings.Join(langLines, "\n"), false)
	if s.width >= 96 {
		return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	}
	return left + "\n\n" + right
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
	if s.contentWidth() < 108 {
		return centerCard
	}

	carousel := lipgloss.JoinHorizontal(lipgloss.Top,
		leftCard,
		"  ",
		centerCard,
		"  ",
		rightCard,
	)
	return carousel
}

func (s *OnboardingScreen) renderFocusedCard(rm *roadmap.Roadmap, total int) string {
	title := rm.Title
	if rm.Recommended {
		title += " " + roadmapMarker(true)
	}

	cardWidth := s.focusedRoadmapCardWidth() - 12
	if cardWidth < 26 {
		cardWidth = 26
	}
	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\nProblems: %d | Est. hours: %d",
		wrapText(title, cardWidth),
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

	return lipgloss.NewStyle().Width(s.focusedRoadmapCardWidth()).Render(renderThemedPanel(s.theme, "Roadmap", content, true))
}

func (s *OnboardingScreen) renderPreviewCard(rm *roadmap.Roadmap) string {
	title := rm.Title
	if rm.Recommended {
		title += " " + roadmapMarker(true)
	}

	content := fmt.Sprintf("%s\n\n%s",
		wrapText(title, 22),
		wrapText(rm.Tagline, 22),
	)

	nav := fmt.Sprintf("<  >")
	content += "\n\n" + nav

	return lipgloss.NewStyle().Width(28).Render(renderThemedPanel(s.theme, "Preview", content, false))
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
		"adaptive":           "Adaptive",
	}
	themeDescs := map[string]string{
		"rpg-skill-tree":     "Strong status colors, bordered cards, reward animations.",
		"clean-productivity": "Minimal color, low motion, high readability.",
		"cyber-dashboard":    "High contrast, neon accents, subtle ambient motion.",
		"adaptive":           "Adjusts to your terminal's background and color profile.",
	}

	lines := []string{"Adaptive appearance follows your terminal. Use this preview only to confirm symbol rendering.", ""}

	for i, theme := range config.ValidThemes {
		label := themeLabels[theme]
		line := fmt.Sprintf("%s - %s", label, themeDescs[theme])
		if i == s.themeFocus {
			lines = append(lines, renderSelectableBlock(s.theme, true, wrapText(line, s.formWidth())))
			continue
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(renderSelectableBlock(s.theme, false, wrapText(line, s.formWidth()))))
	}

	lines = append(lines, "")
	lines = append(lines, s.renderThemePreview(config.ValidThemes[s.themeFocus]))

	return renderThemedPanel(s.theme, "Appearance", strings.Join(lines, "\n"), false)
}

func (s *OnboardingScreen) renderThemePreview(themeID string) string {
	previewTheme, err := LookupTheme(themeID)
	if err != nil {
		previewTheme = s.theme
	}
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	statusLine := strings.Join([]string{
		renderStatusPill(previewTheme, symbols, roadmap.StatusAvailable),
		renderStatusPill(previewTheme, symbols, roadmap.StatusVerified),
		renderStatusPill(previewTheme, symbols, roadmap.StatusLocked),
	}, "  ")
	progress := views.ProgressBar(3, 5, 10, "█", "░")
	body := strings.Join([]string{
		previewTheme.Key.Render("Recommended") + "  Start Two Sum",
		"Why: learn complement lookup",
		"",
		"Status Preview",
		statusLine,
		"",
		"Progress  [" + lipgloss.NewStyle().Foreground(previewTheme.XP).Render(progress) + "]",
		"",
		"Can you see these symbols?",
		fmt.Sprintf("%s %s %s %s %s", symbols.Solved, symbols.Verified, symbols.Locked, symbols.XP, symbols.Review),
		"Symbol mode: " + s.cfg.SymbolMode,
		views.KeytipFooter(map[string]string{"p": "plain/rich", "enter": "select"}, []string{"p", "enter"}, viewPalette(previewTheme)),
	}, "\n")
	return renderThemedPanel(previewTheme, "Theme Preview", body, false)
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
			s.theme.Key.Render("p") + " plain/rich symbols",
			s.theme.Key.Render("enter") + " continue",
		}
	case stepSession:
		items = []string{
			s.theme.Key.Render("up/down") + " choose",
			s.theme.Key.Render("enter") + " next",
		}
	case stepCompletion:
		items = []string{
			s.theme.Key.Render("enter") + " open dashboard",
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

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  •  "))
}

func (s *OnboardingScreen) promptStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.SecondaryAccent).
		Align(lipgloss.Left).
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
	width := s.width - 8
	if width > 156 {
		return 156
	}
	if width < 48 {
		return 48
	}
	return width
}

func (s *OnboardingScreen) focusedRoadmapCardWidth() int {
	width := s.contentWidth()
	if width > 52 {
		return 52
	}
	if width < 36 {
		return 36
	}
	return width
}

func (s *OnboardingScreen) formWidth() int {
	width := s.contentWidth() - 10
	if width > 76 {
		return 76
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

func (s *OnboardingScreen) handleSessionKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "up", "ctrl+p", "k":
		if s.sessionChoice > 0 {
			s.sessionChoice--
		}
	case "down", "ctrl+n", "j":
		if s.sessionChoice < 1 {
			s.sessionChoice++
		}
	}
}

func (s *OnboardingScreen) handleCompletionKey(msg tea.KeyMsg) {
	if !s.onboardingDone {
		return
	}
	switch msg.String() {
	case "up", "ctrl+p", "k":
		if s.sessionChoice > 0 {
			s.sessionChoice--
		}
	case "down", "ctrl+n", "j":
		if s.sessionChoice < 1 {
			s.sessionChoice++
		}
	case "enter":
		if s.sessionChoice == 0 && s.nextAction != nil && s.nextAction.Kind == recommendation.KindStart {
			s.calcAndStart()
		} else {
			s.sessionChoice = 1
		}
	}
}

func (s *OnboardingScreen) renderSession() string {
	var content strings.Builder

	content.WriteString("Accepted Submissions unlock Roadmap progress and XP.\n\n")
	content.WriteString(s.theme.Subtitle.Render("You can skip and use Manual Solve, but Manual Solve earns no XP."))
	content.WriteString("\n\n")

	choices := []string{"Connect now", "Skip for now"}
	for i, choice := range choices {
		if i == s.sessionChoice {
			content.WriteString(renderSelectableBlock(s.theme, true, choice))
			content.WriteString("\n")
			continue
		}
		content.WriteString(lipgloss.NewStyle().Foreground(s.theme.Muted).Render(renderSelectableBlock(s.theme, false, choice)))
		content.WriteString("\n")
	}

	if s.errorMsg != "" {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(s.theme.Danger).Align(lipgloss.Center).Render(s.errorMsg))
	}

	return renderThemedPanel(s.theme, "LeetCode Session", content.String(), true)
}

func (s *OnboardingScreen) renderCompletion() string {
	if s.onboardingDone {
		return s.renderCompletionDone()
	}

	var content strings.Builder
	content.WriteString("You're all set!\n\n")
	content.WriteString(fmt.Sprintf("Roadmap: %s\n", s.cfg.Roadmap))
	content.WriteString(fmt.Sprintf("Language: %s\n", s.cfg.Language))
	content.WriteString(fmt.Sprintf("Workspace: %s\n\n", s.cfg.Workspace))

	if s.nextAction != nil {
		content.WriteString("Recommended First Action\n")
		label := formatActionLabel(s.nextAction.Kind)
		content.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s: %s", label, s.nextAction.Title)))
		content.WriteString("\n")
		reason := s.nextAction.Reason
		if len(reason) > 60 {
			reason = reason[:57] + "..."
		}
		content.WriteString(lipgloss.NewStyle().Foreground(s.theme.Muted).Render(reason))
	}

	content.WriteString("\n\nPress enter to save and continue.")

	return renderThemedPanel(s.theme, "Ready", content.String(), true)
}

func (s *OnboardingScreen) renderCompletionDone() string {
	var content strings.Builder

	content.WriteString("Setup complete!")

	if s.nextAction != nil {
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("Recommended: %s", s.nextAction.Title))
		content.WriteString("\n")
		reason := s.nextAction.Reason
		if len(reason) > 60 {
			reason = reason[:57] + "..."
		}
		content.WriteString(lipgloss.NewStyle().Foreground(s.theme.Muted).Render(reason))
		content.WriteString("\n\n")

		choices := []string{"Start now", "Go to Dashboard"}
		for i, choice := range choices {
			if i == s.sessionChoice {
				content.WriteString(renderSelectableBlock(s.theme, true, choice))
				content.WriteString("\n")
				continue
			}
			content.WriteString(lipgloss.NewStyle().Foreground(s.theme.Muted).Render(renderSelectableBlock(s.theme, false, choice)))
			content.WriteString("\n")
		}
	} else {
		content.WriteString("\n\n")
		content.WriteString("Press enter to go to the Dashboard.")
	}

	return renderThemedPanel(s.theme, "Complete", content.String(), true)
}

func (s *OnboardingScreen) calcAndStart() {
	if s.nextAction == nil || s.db == nil {
		return
	}
	rm, err := catalog.LoadRoadmap(s.cfg.Roadmap)
	if err != nil {
		return
	}
	ctx := context.Background()
	problem, ok := rm.Graph.Problems[s.nextAction.ProblemID]
	if !ok {
		return
	}
	manager := workspace.New(s.cfg.Workspace, generator.New())
	dir := manager.ProblemDir(problem)
	workspace.EnsureManifestWritable(dir, problem.ID)
	s.db.SetProgress(ctx, problem.ID, roadmap.StatusInProgress)
	stubPath, testPath, _ := manager.Generate(problem, generator.Language(s.cfg.Language))
	stage := problem.Stage
	if stage == "" {
		stage = string(problem.Category)
	}
	workspace.WriteManifest(dir, &workspace.Manifest{
		ProblemID:     problem.ID,
		Slug:          problem.Slug,
		Roadmap:       s.cfg.Roadmap,
		Stage:         stage,
		Language:      s.cfg.Language,
		StubPath:      filepath.Base(stubPath),
		TestsuitePath: filepath.Base(testPath),
	})
}
