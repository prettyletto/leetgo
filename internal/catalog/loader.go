package catalog

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"gopkg.in/yaml.v3"
)

const DefaultRoadmapID = "from-zero-to-hero"

//go:embed data/roadmaps/*.yaml
var catalogFS embed.FS

type rawCatalog struct {
	ID             string         `yaml:"id"`
	Title          string         `yaml:"title"`
	Description    string         `yaml:"description"`
	Tagline        string         `yaml:"tagline"`
	Audience       string         `yaml:"audience"`
	Promise        string         `yaml:"promise"`
	Recommended    bool           `yaml:"recommended"`
	EstimatedHours *int           `yaml:"estimated_hours"`
	DifficultyMix  map[string]int `yaml:"difficulty_mix"`
	Highlights     []string       `yaml:"highlights"`
	Stages         []rawStage     `yaml:"stages"`
	Problems       []rawProblem   `yaml:"problems"`

	RoadmapTimeEstimate *string  `yaml:"roadmap_time_estimate"`
	NextRoadmaps        []string `yaml:"next_roadmaps"`
}

type rawStage struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type rawProblem struct {
	ID            int    `yaml:"id"`
	Title         string `yaml:"title"`
	Slug          string `yaml:"slug"`
	Difficulty    string `yaml:"difficulty"`
	Category      string `yaml:"category"`
	Stage         string `yaml:"stage"`
	Prerequisites []int  `yaml:"prerequisites"`

	PracticeFocus       string   `yaml:"practice_focus"`
	ProblemTimeEstimate string   `yaml:"problem_time_estimate"`
	Summary             string   `yaml:"summary"`
	WhyNow              string   `yaml:"why_now"`
	UnlockImpact        string   `yaml:"unlock_impact"`
	Hints               []string `yaml:"hints"`
}

func Load() (*roadmap.Graph, error) {
	rm, err := LoadRoadmap(DefaultRoadmapID)
	if err != nil {
		return nil, err
	}
	return rm.Graph, nil
}

func LoadRoadmap(id string) (*roadmap.Roadmap, error) {
	if strings.TrimSpace(id) == "" {
		id = DefaultRoadmapID
	}

	data, err := catalogFS.ReadFile(fmt.Sprintf("data/roadmaps/%s.yaml", id))
	if err != nil {
		return nil, fmt.Errorf("read embedded roadmap %q: %w", id, err)
	}
	return ParseRoadmap(data)
}

func ListRoadmaps() ([]*roadmap.Roadmap, error) {
	entries, err := catalogFS.ReadDir("data/roadmaps")
	if err != nil {
		return nil, fmt.Errorf("read roadmaps: %w", err)
	}

	roadmaps := make([]*roadmap.Roadmap, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := catalogFS.ReadFile("data/roadmaps/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read roadmap %s: %w", entry.Name(), err)
		}
		rm, err := ParseRoadmap(data)
		if err != nil {
			return nil, fmt.Errorf("parse roadmap %s: %w", entry.Name(), err)
		}
		roadmaps = append(roadmaps, rm)
	}
	sort.Slice(roadmaps, func(i, j int) bool {
		return roadmaps[i].ID < roadmaps[j].ID
	})
	if err := ValidateRoadmapSet(roadmaps); err != nil {
		return nil, err
	}
	return roadmaps, nil
}

func Parse(data []byte) (*roadmap.Graph, error) {
	rm, err := ParseRoadmap(data)
	if err != nil {
		return nil, err
	}
	return rm.Graph, nil
}

func ParseRoadmap(data []byte) (*roadmap.Roadmap, error) {
	var raw rawCatalog
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse roadmap: %w", err)
	}
	if err := validateRawCatalog(raw); err != nil {
		return nil, err
	}
	if raw.ID == "" {
		raw.ID = DefaultRoadmapID
	}
	if raw.Title == "" {
		raw.Title = "From Zero To Hero"
	}

	problems := make([]*roadmap.Problem, len(raw.Problems))
	for i, r := range raw.Problems {
		stage := r.Stage
		if stage == "" {
			stage = r.Category
		}
		problems[i] = &roadmap.Problem{
			ID:                  r.ID,
			Title:               r.Title,
			Slug:                r.Slug,
			Difficulty:          roadmap.Difficulty(r.Difficulty),
			Category:            roadmap.Category(r.Category),
			Stage:               stage,
			Prerequisites:       r.Prerequisites,
			PracticeFocus:       r.PracticeFocus,
			ProblemTimeEstimate: r.ProblemTimeEstimate,
			Summary:             r.Summary,
			WhyNow:              r.WhyNow,
			UnlockImpact:        r.UnlockImpact,
			Hints:               r.Hints,
		}
	}

	stages := make([]roadmap.Stage, len(raw.Stages))
	for i, s := range raw.Stages {
		stages[i] = roadmap.Stage{
			ID:          s.ID,
			Title:       s.Title,
			Description: s.Description,
			Order:       i,
		}
	}

	difficultyMix := make(map[roadmap.Difficulty]int, len(raw.DifficultyMix))
	for k, v := range raw.DifficultyMix {
		difficultyMix[roadmap.Difficulty(k)] = v
	}

	estimatedHours := 0
	if raw.EstimatedHours != nil {
		estimatedHours = *raw.EstimatedHours
	}

	roadmapTimeEstimate := ""
	if raw.RoadmapTimeEstimate != nil {
		roadmapTimeEstimate = *raw.RoadmapTimeEstimate
	}

	rm := &roadmap.Roadmap{
		ID:                  raw.ID,
		Title:               raw.Title,
		Description:         raw.Description,
		Tagline:             raw.Tagline,
		Audience:            raw.Audience,
		Promise:             raw.Promise,
		Recommended:         raw.Recommended,
		EstimatedHours:      estimatedHours,
		DifficultyMix:       difficultyMix,
		Highlights:          raw.Highlights,
		Stages:              stages,
		Graph:               roadmap.NewGraph(problems),
		RoadmapTimeEstimate: roadmapTimeEstimate,
		NextRoadmaps:        raw.NextRoadmaps,
	}
	if _, err := rm.Graph.TopologicalSort(); err != nil {
		return nil, fmt.Errorf("validate roadmap %q: %w", rm.ID, err)
	}
	return rm, nil
}

func validateRawCatalog(raw rawCatalog) error {
	if raw.ID == "" {
		return fmt.Errorf("roadmap id is required")
	}
	if raw.Title == "" {
		return fmt.Errorf("roadmap title is required")
	}
	if raw.Tagline == "" {
		return fmt.Errorf("roadmap tagline is required for %q", raw.ID)
	}
	if raw.Audience == "" {
		return fmt.Errorf("roadmap audience is required for %q", raw.ID)
	}
	if raw.Promise == "" {
		return fmt.Errorf("roadmap promise is required for %q", raw.ID)
	}
	if raw.EstimatedHours != nil && *raw.EstimatedHours <= 0 {
		return fmt.Errorf("roadmap estimated_hours must be positive for %q", raw.ID)
	}
	if raw.RoadmapTimeEstimate != nil && strings.TrimSpace(*raw.RoadmapTimeEstimate) == "" {
		return fmt.Errorf("roadmap_time_estimate must be non-empty for %q", raw.ID)
	}
	if len(raw.Highlights) < 2 || len(raw.Highlights) > 3 {
		return fmt.Errorf("roadmap highlights must contain 2-3 items for %q, got %d", raw.ID, len(raw.Highlights))
	}
	if len(raw.DifficultyMix) > 0 {
		total := 0
		for _, v := range raw.DifficultyMix {
			total += v
		}
		if total != 100 {
			return fmt.Errorf("roadmap difficulty_mix must add to 100 for %q, got %d", raw.ID, total)
		}
	}

	stageIDs := make(map[string]bool, len(raw.Stages))
	for _, stage := range raw.Stages {
		if stage.ID == "" {
			return fmt.Errorf("stage id is required")
		}
		if stageIDs[stage.ID] {
			return fmt.Errorf("duplicate stage id %q", stage.ID)
		}
		stageIDs[stage.ID] = true
	}

	problemIDs := make(map[int]bool, len(raw.Problems))
	for _, problem := range raw.Problems {
		if problem.ID == 0 {
			return fmt.Errorf("problem id is required")
		}
		if problemIDs[problem.ID] {
			return fmt.Errorf("duplicate problem id %d", problem.ID)
		}
		problemIDs[problem.ID] = true
		if problem.Slug == "" {
			return fmt.Errorf("problem %d slug is required", problem.ID)
		}
		if len(stageIDs) > 0 && problem.Stage != "" && !stageIDs[problem.Stage] {
			return fmt.Errorf("problem %d references unknown stage %q", problem.ID, problem.Stage)
		}
	}
	return nil
}

func ValidateRoadmapSet(roadmaps []*roadmap.Roadmap) error {
	recommendedCount := 0
	for _, rm := range roadmaps {
		if rm.Recommended {
			recommendedCount++
		}
	}
	if recommendedCount > 1 {
		return fmt.Errorf("only one bundled roadmap may be marked recommended, found %d", recommendedCount)
	}

	validIDs := make(map[string]bool, len(roadmaps))
	for _, rm := range roadmaps {
		validIDs[rm.ID] = true
	}
	for _, rm := range roadmaps {
		for _, nextID := range rm.NextRoadmaps {
			if !validIDs[nextID] {
				return fmt.Errorf("roadmap %q references unknown next_roadmap %q", rm.ID, nextID)
			}
		}
	}
	return nil
}
