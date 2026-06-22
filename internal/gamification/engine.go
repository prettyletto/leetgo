package gamification

import (
	"context"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
)

type Achievement struct {
	ID          string
	Name        string
	Description string
	Icon        string
}

var Achievements = map[string]Achievement{
	"first_solve": {
		ID:          "first_solve",
		Name:        "First Blood",
		Description: "Solve your first problem",
		Icon:        "🎯",
	},
	"streak_7": {
		ID:          "streak_7",
		Name:        "Week Warrior",
		Description: "Maintain a 7-day streak",
		Icon:        "🔥",
	},
	"streak_30": {
		ID:          "streak_30",
		Name:        "Monthly Master",
		Description: "Maintain a 30-day streak",
		Icon:        "💎",
	},
	"level_5": {
		ID:          "level_5",
		Name:        "Rising Star",
		Description: "Reach level 5",
		Icon:        "⭐",
	},
	"level_10": {
		ID:          "level_10",
		Name:        "Elite Coder",
		Description: "Reach level 10",
		Icon:        "🏆",
	},
	"all_easy": {
		ID:          "all_easy",
		Name:        "Easy Rider",
		Description: "Solve all easy problems",
		Icon:        "✅",
	},
	"first_hard": {
		ID:          "first_hard",
		Name:        "Challenge Accepted",
		Description: "Solve your first hard problem",
		Icon:        "💪",
	},
	"category_complete": {
		ID:          "category_complete",
		Name:        "Category Master",
		Description: "Complete all problems in a category",
		Icon:        "🎓",
	},
}

type Engine struct {
	store store.Store
	graph *roadmap.Graph
}

func NewEngine(s store.Store, g *roadmap.Graph) *Engine {
	return &Engine{store: s, graph: g}
}

func (e *Engine) OnProblemSolved(ctx context.Context, problemID int) ([]string, error) {
	var unlocked []string

	stats, err := e.store.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	if stats.Solved == 1 {
		unlocked = append(unlocked, "first_solve")
	}

	if stats.Streak >= 7 {
		unlocked = append(unlocked, "streak_7")
	}
	if stats.Streak >= 30 {
		unlocked = append(unlocked, "streak_30")
	}

	if stats.Level >= 5 {
		unlocked = append(unlocked, "level_5")
	}
	if stats.Level >= 10 {
		unlocked = append(unlocked, "level_10")
	}

	problem := e.graph.Problems[problemID]
	if problem != nil && problem.Difficulty == roadmap.DifficultyHard {
		unlocked = append(unlocked, "first_hard")
	}

	return unlocked, nil
}
