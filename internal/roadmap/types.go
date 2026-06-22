package roadmap

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type Status string

const (
	StatusLocked     Status = "locked"
	StatusAvailable  Status = "available"
	StatusInProgress Status = "in_progress"
	StatusSolved     Status = "solved"
)

type Category string

type Stage struct {
	ID          string
	Title       string
	Description string
	Order       int
}

type Problem struct {
	ID            int
	Title         string
	Slug          string
	Difficulty    Difficulty
	Category      Category
	Stage         string
	Prerequisites []int
}

type Roadmap struct {
	ID             string
	Title          string
	Description    string
	Tagline        string
	Audience       string
	Promise        string
	Recommended    bool
	EstimatedHours int
	DifficultyMix  map[Difficulty]int
	Highlights     []string
	Stages         []Stage
	Graph          *Graph
}

type Graph struct {
	Problems map[int]*Problem
}

func NewGraph(problems []*Problem) *Graph {
	g := &Graph{
		Problems: make(map[int]*Problem, len(problems)),
	}
	for _, p := range problems {
		g.Problems[p.ID] = p
	}
	return g
}

func (g *Graph) IsUnlocked(problemID int, solved map[int]bool) bool {
	p, ok := g.Problems[problemID]
	if !ok {
		return false
	}
	for _, prereq := range p.Prerequisites {
		if !solved[prereq] {
			return false
		}
	}
	return true
}

func (g *Graph) Available(solved map[int]bool) []*Problem {
	var available []*Problem
	for _, p := range g.Problems {
		if solved[p.ID] {
			continue
		}
		if g.IsUnlocked(p.ID, solved) {
			available = append(available, p)
		}
	}
	return available
}
