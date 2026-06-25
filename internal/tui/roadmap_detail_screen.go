package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/tui/views"
)

type RoadmapDetailScreen struct {
	cfg     *config.Config
	theme   *Theme
	db      store.Store
	roadmap *roadmap.Roadmap

	focusIndex int
	problems   []*roadmap.Problem

	progress     map[int]roadmap.Status
	scrollOffset int

	width  int
	height int
}

func NewRoadmapDetailScreen(cfg *config.Config, theme *Theme, db store.Store, rm *roadmap.Roadmap) *RoadmapDetailScreen {
	progress, _ := db.GetAllProgress(context.Background())
	if progress == nil {
		progress = make(map[int]roadmap.Status)
	}

	s := &RoadmapDetailScreen{
		cfg:      cfg,
		theme:    theme,
		db:       db,
		roadmap:  rm,
		progress: progress,
	}

	sorted, err := rm.Graph.TopologicalSort()
	if err == nil {
		s.problems = s.problemsInStageOrder(sorted)
	}

	return s
}

func (s *RoadmapDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *RoadmapDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg.(type) {
	case NavigateMsg:
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "esc", "backspace":
			return s, func() tea.Msg {
				return NavigateMsg{ScreenID: ScreenDashboard}
			}

		case "enter":
			if s.focusIndex < len(s.problems) {
				p := s.problems[s.focusIndex]
				if s.effectiveStatus(p) == roadmap.StatusLocked {
					missing := s.missingPrerequisites(p)
					if len(missing) > 0 {
						stage := s.problemStageID(p)
						return s, func() tea.Msg {
							return NavigateMsg{ScreenID: ScreenStageDetail, Stage: stage}
						}
					}
				}
				stage := s.problemStageID(p)
				return s, func() tea.Msg {
					return NavigateMsg{ScreenID: ScreenStageDetail, Stage: stage}
				}
			}
			return s, nil

		case "j", "down":
			s.moveFocus(1)

		case "k", "up":
			s.moveFocus(-1)

		case "ctrl+d":
			s.scrollOffset += s.bodyHeight() / 2

		case "ctrl+u":
			s.scrollOffset -= s.bodyHeight() / 2

		}
	}
	return s, nil
}

func (s *RoadmapDetailScreen) View() string {
	header := s.renderHeader()
	body := s.renderListView()
	footer := s.renderFooter()

	return renderScreenShell(s.theme, s.width, s.height, header, body, footer)
}

func (s *RoadmapDetailScreen) renderHeader() string {
	titlePrefix, _, _ := themeRoadmapLabels(s.theme)
	subtitle := ""
	if s.roadmap.Tagline != "" {
		subtitle = s.roadmap.Tagline
	}

	if s.roadmap.Promise != "" {
		if subtitle != "" {
			subtitle += "\n"
		}
		subtitle += s.roadmap.Promise
	}

	return renderScreenHeader(s.theme, titlePrefix+": "+s.roadmap.Title, subtitle)
}

func (s *RoadmapDetailScreen) renderListView() string {
	solvedCount := s.buildStageSolvedCount()
	stagesContent, problemLineMap, _ := s.buildStagesContent(solvedCount)

	_, stagePrefix, _ := themeRoadmapLabels(s.theme)
	avail := s.bodyHeight()
	leftWidth, rightWidth := s.listPaneWidths()

	fp := s.focusedProblem()
	focusedID := 0
	if fp != nil {
		focusedID = fp.ID
	}

	var rightPanel string
	if s.width >= 110 {
		if fp != nil && s.effectiveStatus(fp) == roadmap.StatusLocked {
			rightPanel = s.renderBlockedInfo(fp, rightWidth)
		} else {
			summary := s.renderSummaryPanel(solvedCount, rightWidth)
			comingSoon := s.buildComingSoon()
			var upcoming string
			if len(comingSoon) > 0 {
				_, _, lockedLabel := themeRoadmapLabels(s.theme)
				upcoming = s.renderUpcomingPanel(comingSoon, lockedLabel, rightWidth)
			}
			if upcoming != "" {
				rightPanel = summary + "\n\n" + upcoming
			} else {
				rightPanel = summary
			}
		}
	}

	visible := s.stagesVisibleFromBudget(avail, stagesContent)
	if s.width >= 110 {
		rightHeight := strings.Count(rightPanel, "\n") + 1 + 6
		stagesH := avail
		if rightHeight > stagesH {
			stagesH = rightHeight
		}
		visible = s.stagesVisibleFromBudget(stagesH, stagesContent)
	}

	s.clampScroll(len(stagesContent), visible)
	scrolled, sourceLines := s.windowLinesWithSource(stagesContent, visible, problemLineMap)
	scrolled = s.applySelection(scrolled, sourceLines, problemLineMap, focusedID)
	indicator := s.scrollIndicator(len(stagesContent), visible, stagePrefix)

	stagesText := strings.Join(scrolled, "\n")
	if indicator != "" {
		stagesText = indicator + "\n" + stagesText
	}
	stagesPanel := renderRoadmapPanel(s.theme, "Stage Progress", strings.TrimSpace(stagesText), false, leftWidth)

	if s.width >= 110 {
		left := lipgloss.NewStyle().Width(leftWidth).Render(stagesPanel)
		right := lipgloss.NewStyle().Width(rightWidth).Render(rightPanel)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	}

	return stagesPanel
}

func (s *RoadmapDetailScreen) applySelection(lines []string, sourceLines []int, problemLineMap map[int]int, focusedID int) []string {
	if focusedID == 0 {
		return lines
	}
	result := make([]string, len(lines))
	for i, line := range lines {
		if i < len(sourceLines) {
			if pid, ok := problemLineMap[sourceLines[i]]; ok && pid == focusedID {
				result[i] = renderSelectableBlock(s.theme, true, line)
				continue
			}
		}
		result[i] = lipgloss.NewStyle().Padding(0, 1).Render(line)
	}
	return result
}

func (s *RoadmapDetailScreen) stagesVisibleFromBudget(budget int, stagesContent []string) int {
	budget -= 3
	if len(stagesContent) > budget {
		budget--
	}
	if budget < 1 {
		budget = 1
	}
	return budget
}

func (s *RoadmapDetailScreen) bodyHeight() int {
	if s.height <= 0 {
		return 24
	}
	headerLines := strings.Count(s.renderHeader(), "\n") + 1
	footerLines := strings.Count(s.renderFooter(), "\n") + 1
	overhead := headerLines + footerLines + 6
	h := s.height - overhead
	if h < 10 {
		h = 10
	}
	return h
}

func (s *RoadmapDetailScreen) clampScroll(totalLines, visibleLines int) {
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	if totalLines <= visibleLines {
		s.scrollOffset = 0
		return
	}
	maxScroll := totalLines - visibleLines + 1
	if s.scrollOffset > maxScroll {
		s.scrollOffset = maxScroll
	}
}

func (s *RoadmapDetailScreen) windowLines(lines []string, visible int, problemLineMap map[int]int) []string {
	window, _ := s.windowLinesWithSource(lines, visible, problemLineMap)
	return window
}

func (s *RoadmapDetailScreen) windowLinesWithSource(lines []string, visible int, problemLineMap map[int]int) ([]string, []int) {
	start := s.scrollOffset
	if start < 0 {
		start = 0
	}
	if start >= len(lines) {
		return nil, nil
	}
	sticky := s.currentStageHeaderLine(lines, problemLineMap, start)
	if sticky >= 0 && sticky < start && visible > 1 {
		end := start + visible - 1
		if end > len(lines) {
			end = len(lines)
		}
		window := make([]string, 0, visible)
		sources := make([]int, 0, visible)
		window = append(window, lines[sticky])
		sources = append(sources, sticky)
		for i := start; i < end; i++ {
			window = append(window, lines[i])
			sources = append(sources, i)
		}
		return window, sources
	}
	end := start + visible
	if end > len(lines) {
		end = len(lines)
	}
	window := make([]string, 0, end-start)
	sources := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		window = append(window, lines[i])
		sources = append(sources, i)
	}
	return window, sources
}

func (s *RoadmapDetailScreen) currentStageHeaderLine(lines []string, problemLineMap map[int]int, start int) int {
	for i := start; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if _, ok := problemLineMap[i]; ok {
			continue
		}
		return i
	}
	return -1
}

func (s *RoadmapDetailScreen) scrollIndicator(total, visible int, stagePrefix string) string {
	if total <= visible {
		return ""
	}
	if s.scrollOffset > 0 {
		indicator := lipgloss.NewStyle().Foreground(s.theme.Muted).Render("↑ more " + stagePrefix + " above")
		if s.scrollOffset+visible < total {
			indicator += "  " + lipgloss.NewStyle().Foreground(s.theme.Muted).Render("↓ more below")
		}
		return indicator
	}
	if s.scrollOffset+visible < total {
		return lipgloss.NewStyle().Foreground(s.theme.Muted).Render("↓ more " + stagePrefix + " below")
	}
	return ""
}

func (s *RoadmapDetailScreen) buildStagesContent(solvedCount map[string]int) ([]string, map[int]int, map[int]bool) {
	_, stagePrefix, _ := themeRoadmapLabels(s.theme)
	var lines []string
	problemLineMap := make(map[int]int)
	problemFirstLine := make(map[int]bool)

	groups := s.groupProblemsByStage()
	renderStage := func(stageID, title string) {
		problems, ok := groups[stageID]
		if !ok || len(problems) == 0 {
			return
		}

		total := len(problems)
		solved := solvedCount[stageID]

		header := fmt.Sprintf("%s  %d/%d solved",
			s.theme.Key.Render(stagePrefix+": "+title),
			solved, total,
		)

		if s.roadmap.Stages != nil && len(s.roadmap.Stages) > 1 {
			percentage := float64(solved) / float64(total) * 100
			bar := s.renderMiniBar(percentage)
			header += "  " + bar
		}

		lines = append(lines, header)

		for _, p := range problems {
			first := true
			for _, line := range s.renderProblemLines(p) {
				if first {
					problemLineMap[len(lines)] = p.ID
					problemFirstLine[len(lines)] = true
					first = false
				}
				lines = append(lines, line)
			}
		}

		lines = append(lines, "")
	}

	seenStages := make(map[string]bool)
	for _, stage := range s.roadmap.Stages {
		seenStages[stage.ID] = true
		renderStage(stage.ID, stage.Title)
	}
	for _, p := range s.problems {
		stageID := s.problemStageID(p)
		if seenStages[stageID] {
			continue
		}
		seenStages[stageID] = true
		renderStage(stageID, s.stageDisplayTitle(stageID))
	}
	return lines, problemLineMap, problemFirstLine
}

func (s *RoadmapDetailScreen) renderUpcomingPanel(comingSoon []comingSoonItem, lockedLabel string, panelWidth int) string {
	width := panelBodyWidth(panelWidth)
	var upcomingLines []string
	upcomingLines = append(upcomingLines, s.theme.Key.Render(lockedLabel))
	for _, item := range comingSoon {
		line := fmt.Sprintf("#%d %s", item.problem.ID, item.problem.Title)
		upcomingLines = append(upcomingLines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(wrapText(line, width)))
		upcomingLines = append(upcomingLines, s.theme.Key.Render("Blocked by"))
		for _, b := range item.blockers {
			status := ""
			if s.progress[b.ID] == roadmap.StatusVerified {
				status = " (Verified: Submit or Manual Solve to open gate)"
			}
			blocker := fmt.Sprintf("  - #%d %s%s", b.ID, b.Title, status)
			upcomingLines = append(upcomingLines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(wrapText(blocker, width)))
		}
		upcomingLines = append(upcomingLines, "")
	}
	upcomingLines = trimTrailingBlankLines(upcomingLines)
	return renderRoadmapPanel(s.theme, "Upcoming", strings.Join(upcomingLines, "\n"), false, panelWidth)
}

func (s *RoadmapDetailScreen) scrollToProblem(problemID int) {
	content, problemLineMap, _ := s.buildStagesContent(s.buildStageSolvedCount())
	visible := s.stagesVisible()
	lineIdx := -1
	for i := range content {
		if problemLineMap[i] == problemID {
			lineIdx = i
			break
		}
	}
	if lineIdx < 0 {
		return
	}

	if lineIdx < s.scrollOffset {
		s.scrollOffset = lineIdx
	} else if lineIdx >= s.scrollOffset+visible {
		s.scrollOffset = lineIdx - visible + 1
	}
	s.clampScroll(len(content), visible)

	for attempts := 0; attempts < len(content) && !s.problemVisibleInWindow(problemID, content, problemLineMap, visible); attempts++ {
		previous := s.scrollOffset
		if lineIdx < s.scrollOffset {
			s.scrollOffset--
		} else {
			s.scrollOffset++
		}
		s.clampScroll(len(content), visible)
		if s.scrollOffset == previous {
			return
		}
	}
}

func (s *RoadmapDetailScreen) problemVisibleInWindow(problemID int, content []string, problemLineMap map[int]int, visible int) bool {
	_, sources := s.windowLinesWithSource(content, visible, problemLineMap)
	for _, source := range sources {
		if problemLineMap[source] == problemID {
			return true
		}
	}
	return false
}

func (s *RoadmapDetailScreen) stagesVisible() int {
	avail := s.bodyHeight()
	stagesContent, _, _ := s.buildStagesContent(s.buildStageSolvedCount())
	return s.stagesVisibleFromBudget(avail, stagesContent)
}

func (s *RoadmapDetailScreen) moveFocus(delta int) {
	if len(s.problems) == 0 {
		return
	}
	next := s.focusIndex + delta
	if next < 0 {
		next = 0
	}
	if next >= len(s.problems) {
		next = len(s.problems) - 1
	}
	if next == s.focusIndex {
		return
	}
	s.focusIndex = next
	s.scrollToProblem(s.problems[s.focusIndex].ID)
}

func (s *RoadmapDetailScreen) focusedProblem() *roadmap.Problem {
	if s.focusIndex < 0 || s.focusIndex >= len(s.problems) {
		return nil
	}
	return s.problems[s.focusIndex]
}

func (s *RoadmapDetailScreen) renderBlockedInfo(p *roadmap.Problem, panelWidth int) string {
	width := panelBodyWidth(panelWidth)
	var lines []string
	lines = append(lines, s.theme.Key.Render("Problem"))
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(s.theme.Warning).Render(wrapText(fmt.Sprintf("#%d %s", p.ID, p.Title), width)))

	missing := s.missingPrerequisites(p)
	if len(missing) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.theme.Key.Render("Blocked by"))
		for _, m := range missing {
			lines = append(lines, lipgloss.NewStyle().Foreground(s.theme.Muted).Render(wrapText("  - "+m, width)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, s.theme.Subtitle.Render("Press enter to open the stage."))

	return renderRoadmapPanel(s.theme, "Blocked", strings.Join(lines, "\n"), true, panelWidth)
}

func (s *RoadmapDetailScreen) renderSummaryPanel(solvedCount map[string]int, panelWidth int) string {
	total := len(s.problems)
	solved := 0
	available := 0
	verified := 0
	for _, p := range s.problems {
		switch s.effectiveStatus(p) {
		case roadmap.StatusSolved:
			solved++
		case roadmap.StatusAvailable:
			available++
		case roadmap.StatusVerified:
			verified++
		}
	}
	currentStage := ""
	for _, stage := range s.roadmap.Stages {
		if solvedCount[stage.ID] < len(s.groupProblemsByStage()[stage.ID]) {
			currentStage = stage.Title
			break
		}
	}
	lines := []string{
		fmt.Sprintf("Problems: %d", total),
		fmt.Sprintf("Solved: %d", solved),
		fmt.Sprintf("Available: %d", available),
		fmt.Sprintf("Verified: %d", verified),
	}
	if currentStage != "" {
		lines = append(lines, "", fmt.Sprintf("Current Stage: %s", currentStage))
	}
	return renderRoadmapPanel(s.theme, "Overview", strings.Join(lines, "\n"), false, panelWidth)
}

func (s *RoadmapDetailScreen) listPaneWidths() (int, int) {
	left := 78
	right := 56
	shellWidth := s.width - 8
	if shellWidth <= 0 {
		return left, right
	}
	total := left + right + 2
	if shellWidth >= total {
		return left, right
	}

	right = 44
	left = shellWidth - right - 2
	if left < 62 {
		left = 62
		right = shellWidth - left - 2
	}
	if right < 36 {
		right = 36
	}
	return left, right
}

func renderRoadmapPanel(theme *Theme, title, body string, focused bool, width int) string {
	innerWidth := panelBodyWidth(width)
	if innerWidth <= 0 {
		return renderThemedPanel(theme, title, body, focused)
	}

	lines := strings.Split(body, "\n")
	lineStyle := lipgloss.NewStyle().Width(innerWidth)
	for i, line := range lines {
		lines[i] = lineStyle.Render(line)
	}

	return renderThemedPanel(theme, title, strings.Join(lines, "\n"), focused)
}

func panelBodyWidth(panelWidth int) int {
	inner := panelWidth - 4
	if inner < 20 {
		inner = 20
	}
	return inner
}

func (s *RoadmapDetailScreen) buildComingSoon() []comingSoonItem {
	solvedMap := make(map[int]bool)
	for id, status := range s.progress {
		if status == roadmap.StatusSolved {
			solvedMap[id] = true
		}
	}

	sorted, err := s.roadmap.Graph.TopologicalSort()
	if err != nil {
		return nil
	}

	var items []comingSoonItem
	for _, p := range sorted {
		if solvedMap[p.ID] {
			continue
		}
		if s.progress[p.ID] == roadmap.StatusInProgress {
			continue
		}

		var blockers []*roadmap.Problem
		for _, prereqID := range p.Prerequisites {
			if !solvedMap[prereqID] {
				if bp, ok := s.roadmap.Graph.Problems[prereqID]; ok {
					blockers = append(blockers, bp)
				}
			}
		}

		if len(blockers) >= 1 && len(blockers) <= 2 {
			items = append(items, comingSoonItem{problem: p, blockers: blockers})
			if len(items) >= 3 {
				break
			}
		}
	}
	return items
}

func (s *RoadmapDetailScreen) groupProblemsByStage() map[string][]*roadmap.Problem {
	groups := make(map[string][]*roadmap.Problem)
	for _, p := range s.problems {
		stage := s.problemStageID(p)
		groups[stage] = append(groups[stage], p)
	}
	return groups
}

func (s *RoadmapDetailScreen) problemsInStageOrder(problems []*roadmap.Problem) []*roadmap.Problem {
	groups := make(map[string][]*roadmap.Problem)
	for _, p := range problems {
		stage := s.problemStageID(p)
		groups[stage] = append(groups[stage], p)
	}

	ordered := make([]*roadmap.Problem, 0, len(problems))
	seenStages := make(map[string]bool)
	for _, stage := range s.roadmap.Stages {
		ordered = append(ordered, groups[stage.ID]...)
		seenStages[stage.ID] = true
	}
	for _, p := range problems {
		stage := s.problemStageID(p)
		if !seenStages[stage] {
			ordered = append(ordered, p)
		}
	}
	return ordered
}

func (s *RoadmapDetailScreen) buildStageSolvedCount() map[string]int {
	solvedCount := make(map[string]int)
	for _, p := range s.problems {
		stage := s.problemStageID(p)
		if s.progress[p.ID] == roadmap.StatusSolved {
			solvedCount[stage]++
		}
	}
	return solvedCount
}

func (s *RoadmapDetailScreen) problemStageID(p *roadmap.Problem) string {
	stage := p.Stage
	if stage == "" {
		stage = string(p.Category)
	}
	if stage == "heap" && s.hasStage("heap-priority-queue") {
		return "heap-priority-queue"
	}
	return stage
}

func (s *RoadmapDetailScreen) hasStage(id string) bool {
	for _, stage := range s.roadmap.Stages {
		if stage.ID == id {
			return true
		}
	}
	return false
}

func (s *RoadmapDetailScreen) stageDisplayTitle(id string) string {
	for _, stage := range s.roadmap.Stages {
		if stage.ID == id {
			return stage.Title
		}
	}
	parts := strings.FieldsFunc(id, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func (s *RoadmapDetailScreen) renderProblemLines(p *roadmap.Problem) []string {
	status := s.effectiveStatus(p)
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	marker := renderStatusPill(s.theme, symbols, status)
	label := fmt.Sprintf("#%d %s", p.ID, p.Title)

	prefix := marker + " "
	wrapWidth := s.stageProblemTextWidth() - lipgloss.Width(prefix)
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	wrapped := strings.Split(wrapText(label, wrapWidth), "\n")
	if len(wrapped) == 0 {
		return []string{prefix}
	}

	lines := make([]string, 0, len(wrapped))
	lines = append(lines, prefix+wrapped[0])
	indent := strings.Repeat(" ", lipgloss.Width(prefix))
	for _, line := range wrapped[1:] {
		lines = append(lines, indent+line)
	}
	return lines
}

func (s *RoadmapDetailScreen) stageProblemTextWidth() int {
	if s.width >= 110 {
		left, _ := s.listPaneWidths()
		w := left - 8
		if w > 0 {
			return w
		}
	}
	w := s.width - 16
	if w < 40 {
		w = 40
	}
	if w > 72 {
		w = 72
	}
	return w
}

func (s *RoadmapDetailScreen) effectiveStatus(p *roadmap.Problem) roadmap.Status {
	if status, ok := s.progress[p.ID]; ok && status != "" {
		return status
	}

	locked := false
	for _, prereq := range p.Prerequisites {
		if s.progress[prereq] != roadmap.StatusSolved {
			locked = true
			break
		}
	}
	if locked {
		return roadmap.StatusLocked
	}
	return roadmap.StatusAvailable
}

func (s *RoadmapDetailScreen) renderStatusMarker(status roadmap.Status) string {
	symbols, _ := LookupSymbolSet(s.cfg.SymbolMode)
	return lipgloss.NewStyle().Width(14).Render(renderStatusPill(s.theme, symbols, status))
}

func trimTrailingBlankLines(lines []string) []string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[:end]
}

func (s *RoadmapDetailScreen) missingPrerequisites(p *roadmap.Problem) []string {
	var missing []string
	for _, id := range p.Prerequisites {
		if s.progress[id] == roadmap.StatusSolved {
			continue
		}
		if prereq, ok := s.roadmap.Graph.Problems[id]; ok {
			missing = append(missing, fmt.Sprintf("#%d %s", prereq.ID, prereq.Title))
		} else {
			missing = append(missing, fmt.Sprintf("#%d", id))
		}
	}
	return missing
}

func (s *RoadmapDetailScreen) renderMiniBar(percentage float64) string {
	width := 10
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := views.ProgressBar(filled, width, width, "█", "░")
	return fmt.Sprintf("[%s] %.0f%%", lipgloss.NewStyle().Foreground(s.theme.XP).Render(bar), percentage)
}

func (s *RoadmapDetailScreen) renderFooter() string {
	items := []string{
		s.theme.Key.Render("j/k") + " navigate",
		s.theme.Key.Render("enter") + " problem",
		s.theme.Key.Render("esc") + " dashboard",
		s.theme.Key.Render("q") + " quit",
	}

	return s.theme.Footer.PaddingTop(1).Render(strings.Join(items, "  •  "))
}

func sliceIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
