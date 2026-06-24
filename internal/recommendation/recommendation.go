package recommendation

import (
	"context"
	"fmt"
	"sort"

	"github.com/prettyletto/leetgo/internal/analytics"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
)

type ActionKind string

const (
	KindContinue              ActionKind = "continue"
	KindStart                 ActionKind = "start"
	KindSubmit                ActionKind = "submit"
	KindManualSolve           ActionKind = "manual_solve"
	KindReview                ActionKind = "review"
	KindConnectLeetCode       ActionKind = "connect_leetcode"
	KindExport                ActionKind = "export"
	KindInspect               ActionKind = "inspect"
	KindViewRoadmapCompletion ActionKind = "view_roadmap_completion"
)

type ReasonType string

const (
	ReasonUnlocksDependent         ReasonType = "unlocks_dependent"
	ReasonStrengthensPracticeFocus ReasonType = "strengthens_practice_focus"
	ReasonCompletesVerified        ReasonType = "completes_verified"
	ReasonContinuesInProgress      ReasonType = "continues_in_progress"
	ReasonRepairsWeakness          ReasonType = "repairs_weakness"
	ReasonValidatesManualSolve     ReasonType = "validates_manual_solve"
	ReasonCompletesRoadmap         ReasonType = "completes_roadmap"
)

type NextAction struct {
	ID            string
	Kind          ActionKind
	ProblemID     int
	Title         string
	Reason        string
	ReasonType    ReasonType
	Priority      int
	Stage         string
	Category      string
	Slug          string
	Difficulty    roadmap.Difficulty
	PracticeFocus string
	UnlockImpact  string
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

	prov, _ := c.store.GetSolveProvenanceAll(ctx)
	if prov == nil {
		prov = make(map[int]*store.SolveProvenance)
	}

	weaknesses, _ := analytics.New(c.store, graph).DetectWeaknesses(ctx)
	weaknessCats := make(map[roadmap.Category]bool)
	for _, w := range weaknesses {
		weaknessCats[w.Category] = true
	}

	allSolved := true
	for _, p := range graph.Problems {
		if !solved[p.ID] {
			allSolved = false
			break
		}
	}

	var actions []NextAction
	var priorityOffset int

	// Priority 1: Verified problems (Submit or ManualSolve)
	submitActions := c.submitActions(progress, prov)
	actions = append(actions, submitActions...)
	priorityOffset += len(submitActions)

	// Priority 2: Recent InProgress
	continueActions := c.continueActions(progress, topoIndex)
	actions = append(actions, continueActions...)
	priorityOffset += len(continueActions)

	// Priority 3: Critical Review (weakness blocks progression)
	reviewActions := c.reviewActions(ctx, progress, solved, weaknessCats)
	actions = append(actions, reviewActions...)
	priorityOffset += len(reviewActions)

	// Priority 4-5: Available starts (with gradual ranking)
	startActions := c.startActions(progress, solved, topoIndex, weaknessCats)
	actions = append(actions, startActions...)
	priorityOffset += len(startActions)

	// Priority 6: Maintenance
	maintenance := c.maintenanceActions(progress, prov, allSolved)
	actions = append(actions, maintenance...)
	priorityOffset += len(maintenance)

	for i := range actions {
		actions[i].Priority = i
	}

	return actions, nil
}

func (c *Calculator) submitActions(progress map[int]roadmap.Status, prov map[int]*store.SolveProvenance) []NextAction {
	graph := c.roadmap.Graph

	var entries []struct {
		problem  *roadmap.Problem
		stageOrd int
		topoIdx  int
		isManual bool
	}

	stageOrder := c.stageOrder()
	for id, status := range progress {
		if status != roadmap.StatusVerified {
			continue
		}
		problem, ok := graph.Problems[id]
		if !ok {
			continue
		}

		entries = append(entries, struct {
			problem  *roadmap.Problem
			stageOrd int
			topoIdx  int
			isManual bool
		}{
			problem:  problem,
			stageOrd: stageOrder[problem.Stage],
			isManual: false,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].stageOrd != entries[j].stageOrd {
			return entries[i].stageOrd < entries[j].stageOrd
		}
		return true
	})

	actions := make([]NextAction, 0, len(entries)+len(entries))
	for _, entry := range entries {
		p := entry.problem
		actions = append(actions, NextAction{
			ID:         fmt.Sprintf("submit-%d", p.ID),
			Kind:       KindSubmit,
			ProblemID:  p.ID,
			Title:      p.Title,
			Reason:     fmt.Sprintf("Verified locally — submit to LeetCode for full credit."),
			ReasonType: ReasonCompletesVerified,
			Stage:      p.Stage,
			Category:   string(p.Category),
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
		})
		actions = append(actions, NextAction{
			ID:         fmt.Sprintf("manual_solve-%d", p.ID),
			Kind:       KindManualSolve,
			ProblemID:  p.ID,
			Title:      p.Title,
			Reason:     fmt.Sprintf("Mark as manually solved — unlocks dependents but awards no XP."),
			ReasonType: ReasonCompletesVerified,
			Stage:      p.Stage,
			Category:   string(p.Category),
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
		})
	}
	return actions
}

func (c *Calculator) continueActions(progress map[int]roadmap.Status, topoIndex map[int]int) []NextAction {
	graph := c.roadmap.Graph

	type entry struct {
		problem *roadmap.Problem
		topoIdx int
	}

	var entries []entry
	for id, status := range progress {
		if status != roadmap.StatusInProgress {
			continue
		}
		problem, ok := graph.Problems[id]
		if !ok {
			continue
		}
		entries = append(entries, entry{problem: problem, topoIdx: topoIndex[id]})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].topoIdx < entries[j].topoIdx
	})

	actions := make([]NextAction, 0, len(entries))
	for _, e := range entries {
		p := e.problem
		actions = append(actions, NextAction{
			ID:         fmt.Sprintf("continue-%d", p.ID),
			Kind:       KindContinue,
			ProblemID:  p.ID,
			Title:      p.Title,
			Reason:     c.continueReason(p),
			ReasonType: ReasonContinuesInProgress,
			Stage:      p.Stage,
			Category:   string(p.Category),
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
		})
	}
	return actions
}

func (c *Calculator) continueReason(p *roadmap.Problem) string {
	if p.PracticeFocus != "" {
		return fmt.Sprintf("You were working on this — practice focus: %s.", p.PracticeFocus)
	}
	return "You were working on this problem."
}

func (c *Calculator) startActions(progress map[int]roadmap.Status, solved map[int]bool, topoIndex map[int]int, weaknessCats map[roadmap.Category]bool) []NextAction {
	graph := c.roadmap.Graph
	stageOrder := c.stageOrder()

	type entry struct {
		problem       *roadmap.Problem
		topoIdx       int
		stageOrd      int
		difficulty    int
		unlockCount   int
		indirectCount int
		isWeakness    bool
	}

	var entries []entry
	for _, p := range graph.Problems {
		if progress[p.ID] == roadmap.StatusSolved || progress[p.ID] == roadmap.StatusVerified {
			continue
		}
		if progress[p.ID] == roadmap.StatusInProgress {
			continue
		}
		if !graph.IsUnlocked(p.ID, solved) {
			continue
		}

		unlockCount, indirectCount := c.dependentCounts(p.ID, solved)

		entries = append(entries, entry{
			problem:       p,
			topoIdx:       topoIndex[p.ID],
			stageOrd:      stageOrder[p.Stage],
			difficulty:    difficultyRank(p.Difficulty),
			unlockCount:   unlockCount,
			indirectCount: indirectCount,
			isWeakness:    weaknessCats[p.Category],
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		ei, ej := entries[i], entries[j]
		if ei.isWeakness != ej.isWeakness {
			return ei.isWeakness
		}
		if ei.stageOrd != ej.stageOrd {
			return ei.stageOrd < ej.stageOrd
		}
		if ei.difficulty != ej.difficulty {
			return ei.difficulty < ej.difficulty
		}
		if ei.unlockCount != ej.unlockCount {
			return ei.unlockCount > ej.unlockCount
		}
		return ei.topoIdx < ej.topoIdx
	})

	var critical []NextAction
	var regular []NextAction
	for _, e := range entries {
		p := e.problem
		reason := c.startReason(p, e.unlockCount, e.indirectCount, e.isWeakness)
		action := NextAction{
			ID:            fmt.Sprintf("start-%d", p.ID),
			Kind:          KindStart,
			ProblemID:     p.ID,
			Title:         p.Title,
			Reason:        reason,
			ReasonType:    ReasonStrengthensPracticeFocus,
			Stage:         p.Stage,
			Category:      string(p.Category),
			Slug:          p.Slug,
			Difficulty:    p.Difficulty,
			PracticeFocus: p.PracticeFocus,
			UnlockImpact:  p.UnlockImpact,
		}
		if e.isWeakness || e.unlockCount > 0 {
			action.ReasonType = ReasonUnlocksDependent
			critical = append(critical, action)
		} else {
			regular = append(regular, action)
		}
	}

	result := make([]NextAction, 0, len(critical)+len(regular))
	result = append(result, critical...)
	result = append(result, regular...)
	return result
}

func (c *Calculator) startReason(p *roadmap.Problem, unlockCount, indirectCount int, isWeakness bool) string {
	if p.PracticeFocus != "" {
		if isWeakness {
			return fmt.Sprintf("Recommended to repair weakness in %s.", p.Category)
		}
		if unlockCount > 0 {
			verb := "unlocks"
			if unlockCount == 1 {
				verb = "unlocks"
			}
			return fmt.Sprintf("Strengthens %s — %s %d direct dependent(s).", p.PracticeFocus, verb, unlockCount)
		}
		return fmt.Sprintf("Strengthens %s.", p.PracticeFocus)
	}
	if unlockCount > 0 {
		return fmt.Sprintf("Unlocks %d dependent problem(s).", unlockCount)
	}
	return "This problem is ready for you."
}

func (c *Calculator) reviewActions(ctx context.Context, progress map[int]roadmap.Status, solved map[int]bool, weaknessCats map[roadmap.Category]bool) []NextAction {
	graph := c.roadmap.Graph

	// Get existing review cycles to avoid duplicates
	existingCycles, _ := c.store.GetReviewCycles(ctx)
	cycleMap := make(map[int]map[string]bool)
	for _, rc := range existingCycles {
		if rc.CompletedAt == nil {
			if cycleMap[rc.ProblemID] == nil {
				cycleMap[rc.ProblemID] = make(map[string]bool)
			}
			cycleMap[rc.ProblemID][rc.Reason] = true
		}
	}

	type entry struct {
		problem *roadmap.Problem
		reason  string
		topoIdx int
	}

	var entries []entry

	// Weakness repair: revisit solved problems in weak categories
	for _, p := range graph.Problems {
		if !weaknessCats[p.Category] {
			continue
		}
		if !solved[p.ID] {
			continue
		}
		if cycleMap[p.ID] != nil && cycleMap[p.ID]["weakness"] {
			continue
		}
		entries = append(entries, entry{problem: p, reason: "weakness"})
	}

	// Manual solve validation: recommend submitting manual solves
	prov, _ := c.store.GetSolveProvenanceAll(ctx)
	for _, p := range graph.Problems {
		sp, ok := prov[p.ID]
		if !ok || sp.Kind != "manual" {
			continue
		}
		if cycleMap[p.ID] != nil && cycleMap[p.ID]["manual_solve_validation"] {
			continue
		}
		entries = append(entries, entry{problem: p, reason: "manual_solve_validation"})
	}

	// Failed attempts: problems with high fail rate
	for _, p := range graph.Problems {
		if !solved[p.ID] {
			continue
		}
		attempts, _ := c.store.GetAttempts(ctx, p.ID)
		fails := 0
		for _, a := range attempts {
			if !a.Passed {
				fails++
			}
		}
		if fails < 3 {
			continue
		}
		if cycleMap[p.ID] != nil && cycleMap[p.ID]["failed_attempts"] {
			continue
		}
		entries = append(entries, entry{problem: p, reason: "failed_attempts"})
	}

	topo, _ := graph.TopologicalSort()
	topoIndex := make(map[int]int, len(topo))
	for i, p := range topo {
		topoIndex[p.ID] = i
	}
	for i := range entries {
		entries[i].topoIdx = topoIndex[entries[i].problem.ID]
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].topoIdx < entries[j].topoIdx
	})

	if len(entries) > 3 {
		entries = entries[:3]
	}

	actions := make([]NextAction, 0, len(entries))
	for _, e := range entries {
		p := e.problem
		reasonText := ""
		reasonType := ReasonRepairsWeakness
		switch e.reason {
		case "weakness":
			reasonText = fmt.Sprintf("Repair weakness in %s by revisiting.", p.Category)
		case "manual_solve_validation":
			reasonText = fmt.Sprintf("Validate Manual Solve with an Accepted Submission to earn XP.")
			reasonType = ReasonValidatesManualSolve
		case "failed_attempts":
			reasonText = "High failed-attempt history — review for deeper mastery."
		}
		actions = append(actions, NextAction{
			ID:         fmt.Sprintf("review-%d", p.ID),
			Kind:       KindReview,
			ProblemID:  p.ID,
			Title:      p.Title,
			Reason:     reasonText,
			ReasonType: reasonType,
			Stage:      p.Stage,
			Category:   string(p.Category),
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
		})
	}
	return actions
}

func (c *Calculator) maintenanceActions(progress map[int]roadmap.Status, prov map[int]*store.SolveProvenance, allSolved bool) []NextAction {
	var actions []NextAction

	hasLeetCode := false
	for _, sp := range prov {
		if sp.Kind == "accepted" {
			hasLeetCode = true
			break
		}
	}

	if !hasLeetCode {
		verifiedCount := 0
		for _, status := range progress {
			if status == roadmap.StatusVerified {
				verifiedCount++
			}
		}
		if verifiedCount > 0 {
			actions = append(actions, NextAction{
				ID:         "connect_leetcode",
				Kind:       KindConnectLeetCode,
				Title:      "Connect LeetCode Session",
				Reason:     fmt.Sprintf("You have %d Verified problem(s) waiting for Submission. Accepted Solves earn XP and unlock progress.", verifiedCount),
				ReasonType: ReasonCompletesVerified,
			})
		}
	}

	if allSolved && len(c.roadmap.NextRoadmaps) > 0 {
		actions = append(actions, NextAction{
			ID:         "view_roadmap_completion",
			Kind:       KindViewRoadmapCompletion,
			Title:      "View Roadmap Completion",
			Reason:     fmt.Sprintf("All problems Solved! Consider continuing with: %s.", c.roadmap.NextRoadmaps[0]),
			ReasonType: ReasonCompletesRoadmap,
		})
	}

	actions = append(actions, NextAction{
		ID:     "export",
		Kind:   KindExport,
		Title:  "Git Export via CLI",
		Reason: "Run leetgo git-export <repo-dir> --commit to back up progress.",
	})

	actions = append(actions, NextAction{
		ID:     "inspect",
		Kind:   KindInspect,
		Title:  "Review Practice Log",
		Reason: "Inspect your submission history and notes.",
	})

	return actions
}

func (c *Calculator) stageOrder() map[string]int {
	order := make(map[string]int, len(c.roadmap.Stages))
	for _, stage := range c.roadmap.Stages {
		order[stage.ID] = stage.Order
	}
	return order
}

func (c *Calculator) dependentCounts(problemID int, solved map[int]bool) (directCount int, indirectCount int) {
	graph := c.roadmap.Graph
	for _, p := range graph.Problems {
		for _, prereq := range p.Prerequisites {
			if prereq == problemID {
				if !solved[p.ID] {
					directCount++
				}
				subDirect, subIndirect := c.dependentCounts(p.ID, solved)
				indirectCount += subDirect + subIndirect
			}
		}
	}
	return
}

func difficultyRank(d roadmap.Difficulty) int {
	switch d {
	case roadmap.DifficultyEasy:
		return 0
	case roadmap.DifficultyMedium:
		return 1
	case roadmap.DifficultyHard:
		return 2
	default:
		return 0
	}
}
