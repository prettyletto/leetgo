package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/roadmap"
)

var (
	titleLine = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("219"))

	stageLine = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75")).
			PaddingTop(1)

	nodeStyle = lipgloss.NewStyle().
			Bold(true).
			Width(8)

	lockedNode = nodeStyle.
			Foreground(lipgloss.Color("240"))

	availableNode = nodeStyle.
			Foreground(lipgloss.Color("214"))

	activeNode = nodeStyle.
			Foreground(lipgloss.Color("39"))

	solvedNode = nodeStyle.
			Foreground(lipgloss.Color("82"))

	edgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))
)

type GraphView struct {
	roadmap  *roadmap.Roadmap
	graph    *roadmap.Graph
	progress map[int]roadmap.Status
	scrollY  int
	width    int
	height   int
}

func NewGraphView(rm *roadmap.Roadmap, progress map[int]roadmap.Status) *GraphView {
	return &GraphView{
		roadmap:  rm,
		graph:    rm.Graph,
		progress: progress,
		width:    80,
		height:   24,
	}
}

func (v *GraphView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *GraphView) Scroll(delta int) {
	v.scrollY += delta
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

func (v *GraphView) Render() string {
	sorted, err := v.graph.TopologicalSort()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	stageTitles := v.stageTitles()
	byStage := make(map[string][]*roadmap.Problem)
	stageOrder := make([]string, 0, len(v.roadmap.Stages))
	seenStage := make(map[string]bool)
	for _, stage := range v.roadmap.Stages {
		stageOrder = append(stageOrder, stage.ID)
		seenStage[stage.ID] = true
	}
	for _, p := range sorted {
		stage := p.Stage
		if stage == "" {
			stage = string(p.Category)
		}
		byStage[stage] = append(byStage[stage], p)
		if !seenStage[stage] {
			stageOrder = append(stageOrder, stage)
			seenStage[stage] = true
		}
	}

	lines := []string{titleLine.Render("Unlock Path")}
	for _, stageID := range stageOrder {
		problems := byStage[stageID]
		if len(problems) == 0 {
			continue
		}
		title := stageTitles[stageID]
		if title == "" {
			title = stageID
		}
		lines = append(lines, stageLine.Render(title))
		for _, p := range problems {
			lines = append(lines, v.renderProblemLine(p))
		}
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	allLines := strings.Split(content, "\n")

	start := v.scrollY
	end := start + v.height - 4
	if end > len(allLines) {
		end = len(allLines)
	}
	if start >= len(allLines) {
		start = len(allLines) - 1
	}

	visible := allLines[start:end]
	return strings.Join(visible, "\n")
}

func (v *GraphView) renderProblemLine(p *roadmap.Problem) string {
	status := v.statusFor(p)
	marker := v.renderMarker(status)
	label := fmt.Sprintf("#%d %s", p.ID, p.Title)
	if status == roadmap.StatusLocked {
		missing := v.missingPrerequisites(p)
		if len(missing) > 0 {
			label += edgeStyle.Render("  blocked by " + strings.Join(missing, ", "))
		}
	}
	return fmt.Sprintf("  %s %s", marker, label)
}

func (v *GraphView) renderMarker(status roadmap.Status) string {
	switch status {
	case roadmap.StatusSolved:
		return solvedNode.Render("SOLVED")
	case roadmap.StatusInProgress:
		return activeNode.Render("ACTIVE")
	case roadmap.StatusAvailable:
		return availableNode.Render("READY")
	default:
		return lockedNode.Render("LOCKED")
	}
}

func (v *GraphView) statusFor(p *roadmap.Problem) roadmap.Status {
	if status := v.progress[p.ID]; status != "" {
		return status
	}
	for _, prereq := range p.Prerequisites {
		if v.progress[prereq] != roadmap.StatusSolved {
			return roadmap.StatusLocked
		}
	}
	return roadmap.StatusAvailable
}

func (v *GraphView) missingPrerequisites(p *roadmap.Problem) []string {
	var missing []string
	for _, id := range p.Prerequisites {
		if v.progress[id] == roadmap.StatusSolved {
			continue
		}
		if prereq, ok := v.graph.Problems[id]; ok {
			missing = append(missing, fmt.Sprintf("#%d %s", prereq.ID, prereq.Title))
		} else {
			missing = append(missing, fmt.Sprintf("#%d", id))
		}
	}
	return missing
}

func (v *GraphView) stageTitles() map[string]string {
	titles := make(map[string]string, len(v.roadmap.Stages))
	for _, stage := range v.roadmap.Stages {
		titles[stage.ID] = stage.Title
	}
	return titles
}
