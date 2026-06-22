package recommendation

import (
	"context"
	"fmt"
	"sort"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
)

type ActionKind string

const (
	KindContinue ActionKind = "continue"
	KindStart    ActionKind = "start"
	KindReview   ActionKind = "review"
	KindExport   ActionKind = "export"
	KindInspect  ActionKind = "inspect"
)

type NextAction struct {
	ID        string
	Kind      ActionKind
	ProblemID int
	Title     string
	Reason    string
	Priority  int
	Stage     string
	Category  string
	Slug      string
}

type Calculator struct {
	store   store.Store
	roadmap *roadmap.Roadmap
}

func NewCalculator(s store.Store, rm *roadmap.Roadmap) *Calculator {
	return &Calculator{store: s, roadmap: rm}
}

func (c *Calculator) Calculate(ctx context.Context) ([]NextAction, error) {
	progress, err := c.store.GetAllProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}

	graph := c.roadmap.Graph

	solved := make(map[int]bool)
	for id, status := range progress {
		if status == roadmap.StatusSolved {
			solved[id] = true
		}
	}

	topo, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}

	topoIndex := make(map[int]int, len(topo))
	for i, p := range topo {
		topoIndex[p.ID] = i
	}

	var actions []NextAction

	actions = append(actions, c.continueActions(progress, topoIndex)...)
	actions = append(actions, c.startActions(progress, solved, topoIndex)...)
	actions = append(actions, c.reviewActions(ctx)...)
	actions = append(actions, c.maintenanceActions()...)

	for i := range actions {
		actions[i].Priority = i
	}

	return actions, nil
}

func (c *Calculator) continueActions(progress map[int]roadmap.Status, topoIndex map[int]int) []NextAction {
	graph := c.roadmap.Graph

	type inProgressEntry struct {
		problem *roadmap.Problem
		topoIdx int
	}

	var entries []inProgressEntry
	for id, status := range progress {
		if status != roadmap.StatusInProgress {
			continue
		}
		problem, ok := graph.Problems[id]
		if !ok {
			continue
		}
		entries = append(entries, inProgressEntry{problem: problem, topoIdx: topoIndex[id]})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].topoIdx < entries[j].topoIdx
	})

	actions := make([]NextAction, 0, len(entries))
	for _, entry := range entries {
		p := entry.problem
		actions = append(actions, NextAction{
			ID:        fmt.Sprintf("continue-%d", p.ID),
			Kind:      KindContinue,
			ProblemID: p.ID,
			Title:     p.Title,
			Reason:    "You were working on this problem.",
			Stage:     p.Stage,
			Category:  string(p.Category),
			Slug:      p.Slug,
		})
	}
	return actions
}

func (c *Calculator) startActions(progress map[int]roadmap.Status, solved map[int]bool, topoIndex map[int]int) []NextAction {
	graph := c.roadmap.Graph

	stageOrder := make(map[string]int, len(c.roadmap.Stages))
	for _, stage := range c.roadmap.Stages {
		stageOrder[stage.ID] = stage.Order
	}

	type availableEntry struct {
		problem    *roadmap.Problem
		topoIdx    int
		stageOrder int
	}

	var entries []availableEntry
	for _, p := range graph.Problems {
		if progress[p.ID] == roadmap.StatusSolved {
			continue
		}
		if progress[p.ID] == roadmap.StatusInProgress {
			continue
		}
		if !graph.IsUnlocked(p.ID, solved) {
			continue
		}
		entries = append(entries, availableEntry{
			problem:    p,
			topoIdx:    topoIndex[p.ID],
			stageOrder: stageOrder[p.Stage],
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].stageOrder != entries[j].stageOrder {
			return entries[i].stageOrder < entries[j].stageOrder
		}
		return entries[i].topoIdx < entries[j].topoIdx
	})

	actions := make([]NextAction, 0, len(entries))
	for _, entry := range entries {
		p := entry.problem
		actions = append(actions, NextAction{
			ID:        fmt.Sprintf("start-%d", p.ID),
			Kind:      KindStart,
			ProblemID: p.ID,
			Title:     p.Title,
			Reason:    "This problem is ready for you.",
			Stage:     p.Stage,
			Category:  string(p.Category),
			Slug:      p.Slug,
		})
	}
	return actions
}

func (c *Calculator) reviewActions(ctx context.Context) []NextAction {
	return nil
}

func (c *Calculator) maintenanceActions() []NextAction {
	return []NextAction{
		{
			ID:     "export",
			Kind:   KindExport,
			Title:  "Git Export via CLI",
			Reason: "Run leetgo git-export <repo-dir> --commit to back up progress.",
		},
		{
			ID:     "inspect",
			Kind:   KindInspect,
			Title:  "Review Solve Log",
			Reason: "Inspect your submission history and notes.",
		},
	}
}
