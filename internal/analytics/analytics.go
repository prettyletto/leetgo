package analytics

import (
	"context"
	"fmt"
	"sort"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
)

type CategoryStats struct {
	Category      roadmap.Category
	Total         int
	Solved        int
	Attempts      int
	Failures      int
	FailRate      float64
	AvgDifficulty float64
}

type Weakness struct {
	Category roadmap.Category
	Severity float64
	Reason   string
}

type Analytics struct {
	store store.Store
	graph *roadmap.Graph
}

func New(s store.Store, g *roadmap.Graph) *Analytics {
	return &Analytics{store: s, graph: g}
}

func (a *Analytics) CategoryStats(ctx context.Context) ([]CategoryStats, error) {
	categories := make(map[roadmap.Category]*CategoryStats)

	for _, p := range a.graph.Problems {
		cat := p.Category
		if _, ok := categories[cat]; !ok {
			categories[cat] = &CategoryStats{Category: cat}
		}
		categories[cat].Total++
	}

	progress, err := a.store.GetAllProgress(ctx)
	if err != nil {
		return nil, err
	}

	for id, status := range progress {
		p := a.graph.Problems[id]
		if p == nil {
			continue
		}
		if status == roadmap.StatusSolved {
			categories[p.Category].Solved++
		}
	}

	for _, p := range a.graph.Problems {
		attempts, err := a.store.GetAttempts(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get attempts for problem %d: %w", p.ID, err)
		}
		stats := categories[p.Category]
		stats.Attempts += len(attempts)
		for _, att := range attempts {
			if !att.Passed {
				stats.Failures++
			}
			if att.SelfReported != "" {
				stats.AvgDifficulty += difficultyScore(att.SelfReported)
			}
		}
	}

	var result []CategoryStats
	for _, stats := range categories {
		if stats.Attempts > 0 {
			stats.FailRate = float64(stats.Failures) / float64(stats.Attempts)
			stats.AvgDifficulty /= float64(stats.Attempts)
		}
		result = append(result, *stats)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FailRate > result[j].FailRate
	})

	return result, nil
}

func (a *Analytics) DetectWeaknesses(ctx context.Context) ([]Weakness, error) {
	stats, err := a.CategoryStats(ctx)
	if err != nil {
		return nil, err
	}

	var weaknesses []Weakness
	for _, s := range stats {
		if s.Attempts < 3 {
			continue
		}

		if s.FailRate > 0.5 {
			weaknesses = append(weaknesses, Weakness{
				Category: s.Category,
				Severity: s.FailRate,
				Reason:   "High failure rate",
			})
		}

		if s.AvgDifficulty > 2.0 {
			weaknesses = append(weaknesses, Weakness{
				Category: s.Category,
				Severity: s.AvgDifficulty / 3.0,
				Reason:   "Self-reported as difficult",
			})
		}
	}

	sort.Slice(weaknesses, func(i, j int) bool {
		return weaknesses[i].Severity > weaknesses[j].Severity
	})

	return weaknesses, nil
}

func difficultyScore(d roadmap.Difficulty) float64 {
	switch d {
	case roadmap.DifficultyEasy:
		return 1.0
	case roadmap.DifficultyMedium:
		return 2.0
	case roadmap.DifficultyHard:
		return 3.0
	default:
		return 0
	}
}
