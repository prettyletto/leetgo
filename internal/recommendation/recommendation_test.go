package recommendation

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRoadmap() *roadmap.Roadmap {
	problems := []*roadmap.Problem{
		{ID: 1, Title: "Two Sum", Slug: "two-sum", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "foundations"},
		{ID: 217, Title: "Contains Duplicate", Slug: "contains-duplicate", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "foundations"},
		{ID: 49, Title: "Group Anagrams", Slug: "group-anagrams", Difficulty: roadmap.DifficultyMedium, Category: "arrays-hashing", Stage: "foundations", Prerequisites: []int{242}},
		{ID: 242, Title: "Valid Anagram", Slug: "valid-anagram", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "foundations"},
		{ID: 347, Title: "Top K Frequent", Slug: "top-k-frequent-elements", Difficulty: roadmap.DifficultyMedium, Category: "arrays-hashing", Stage: "core-patterns", Prerequisites: []int{217}},
		{ID: 15, Title: "3Sum", Slug: "3sum", Difficulty: roadmap.DifficultyMedium, Category: "two-pointers", Stage: "core-patterns", Prerequisites: []int{1}},
		{ID: 42, Title: "Trapping Rain Water", Slug: "trapping-rain-water", Difficulty: roadmap.DifficultyHard, Category: "two-pointers", Stage: "advanced", Prerequisites: []int{15}},
	}

	graph := roadmap.NewGraph(problems)

	return &roadmap.Roadmap{
		ID:    "test",
		Title: "Test Roadmap",
		Stages: []roadmap.Stage{
			{ID: "foundations", Title: "Foundations", Order: 0},
			{ID: "core-patterns", Title: "Core Patterns", Order: 1},
			{ID: "advanced", Title: "Advanced", Order: 2},
		},
		Graph: graph,
	}
}

func newTestCalculator(t *testing.T) (*Calculator, store.Store) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	rm := testRoadmap()
	return NewCalculator(s, rm), s
}

func collectActionsByKind(actions []NextAction, kind ActionKind) []NextAction {
	var result []NextAction
	for _, a := range actions {
		if a.Kind == kind {
			result = append(result, a)
		}
	}
	return result
}

func TestCalculate_EmptyProgress(t *testing.T) {
	c, _ := newTestCalculator(t)
	ctx := context.Background()

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, actions)

	continues := collectActionsByKind(actions, KindContinue)
	assert.Empty(t, continues, "no InProgress when no progress exists")

	starts := collectActionsByKind(actions, KindStart)
	assert.NotEmpty(t, starts, "should list Available problems")

	startTitles := make([]string, len(starts))
	for i, a := range starts {
		startTitles[i] = a.Title
	}

	expectedAvailable := []string{"Two Sum", "Contains Duplicate", "Valid Anagram"}
	for _, exp := range expectedAvailable {
		assert.Contains(t, startTitles, exp, "expected %s in available Start actions", exp)
	}
}

func TestCalculate_InProgressRanksBeforeAvailable(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	continues := collectActionsByKind(actions, KindContinue)
	require.Len(t, continues, 1)
	assert.Equal(t, 1, continues[0].ProblemID)
	assert.Equal(t, KindContinue, continues[0].Kind)
	assert.Equal(t, "Two Sum", continues[0].Title)

	starts := collectActionsByKind(actions, KindStart)
	assert.NotEmpty(t, starts)

	for _, a := range starts {
		if a.ProblemID == 1 {
			t.Error("InProgress Two Sum should not appear as Start action")
		}
	}

	continueIdx := -1
	startIdx := -1
	for i, a := range actions {
		if a.Kind == KindContinue && continueIdx < 0 {
			continueIdx = i
		}
		if a.Kind == KindStart && startIdx < 0 {
			startIdx = i
		}
	}
	require.NotEqual(t, -1, continueIdx)
	require.NotEqual(t, -1, startIdx)
	assert.Less(t, continueIdx, startIdx, "InProgress should rank before Available")
}

func TestCalculate_AvailableFollowsStageOrder(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)
	require.NotEmpty(t, starts)

	foundationsStage := "foundations"
	corePatternsStage := "core-patterns"

	foundLastFoundations := -1
	for i, a := range starts {
		if a.Stage == foundationsStage {
			foundLastFoundations = i
		}
		if a.Stage == corePatternsStage && foundLastFoundations >= 0 {
			assert.Less(t, foundLastFoundations, i, "foundations problems should come before core-patterns")
		}
	}

	for _, a := range starts {
		if a.ProblemID == 42 {
			t.Error("locked Trapping Rain Water should not appear")
		}
	}
}

func TestCalculate_SolveUnlocksDependents(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)

	hasGroupAnagrams := false
	for _, a := range starts {
		if a.ProblemID == 49 {
			hasGroupAnagrams = true
		}
		if a.ProblemID == 42 {
			t.Error("locked Trapping Rain Water should not appear")
		}
	}
	assert.True(t, hasGroupAnagrams, "Group Anagrams should be available after Valid Anagram solved")
}

func TestCalculate_LockedProblemsExcluded(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	for _, a := range actions {
		if a.Kind == KindStart && a.ProblemID == 42 {
			t.Error("Trapping Rain Water requires 3Sum which is locked")
		}
	}
}

func TestCalculate_AllKindsPresent(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	hasContinue := false
	hasStart := false
	hasExport := false
	hasInspect := false

	for _, a := range actions {
		switch a.Kind {
		case KindContinue:
			hasContinue = true
		case KindStart:
			hasStart = true
		case KindExport:
			hasExport = true
		case KindInspect:
			hasInspect = true
		}
	}

	assert.True(t, hasContinue, "should have at least one Continue action")
	assert.True(t, hasStart, "should have at least one Start action")
	assert.True(t, hasExport, "should have an Export maintenance action")
	assert.True(t, hasInspect, "should have an Inspect maintenance action")
}

func TestCalculate_MaintenanceActionsLast(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	lastProblemAction := -1
	firstMaintenanceAction := -1

	for i, a := range actions {
		if a.Kind == KindContinue || a.Kind == KindStart {
			lastProblemAction = i
		}
		if (a.Kind == KindExport || a.Kind == KindInspect) && firstMaintenanceAction < 0 {
			firstMaintenanceAction = i
		}
	}

	if lastProblemAction >= 0 && firstMaintenanceAction >= 0 {
		assert.Less(t, lastProblemAction, firstMaintenanceAction,
			"maintenance actions should come after problem-based actions")
	}
}

func TestCalculate_MultipleInProgress(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	continues := collectActionsByKind(actions, KindContinue)
	assert.Len(t, continues, 2, "should have two InProgress actions")
}

func TestCalculate_SolvedProblemsNotListed(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	for _, a := range actions {
		if a.ProblemID == 1 || a.ProblemID == 217 || a.ProblemID == 242 {
			assert.NotEqual(t, KindContinue, a.Kind, "solved problem should not be Continue")
			assert.NotEqual(t, KindStart, a.Kind, "solved problem should not be Start")
		}
	}
}

func TestCalculate_VerifiedProblemsNotListedAsStart(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	for _, a := range collectActionsByKind(actions, KindStart) {
		assert.NotEqual(t, 242, a.ProblemID, "verified problem should not appear in Start actions")
	}
}

func TestCalculate_InProgressExcludedFromStart(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)
	for _, a := range starts {
		assert.NotEqual(t, 1, a.ProblemID, "InProgress problem should not appear in Start actions")
	}
}

func TestCalculate_StartsWithProblemFields(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)
	require.NotEmpty(t, starts)

	first := starts[0]
	assert.Equal(t, "start-"+itoa(first.ProblemID), first.ID)
	assert.Equal(t, KindStart, first.Kind)
	assert.NotEmpty(t, first.Title)
	assert.NotEmpty(t, first.Reason)
	assert.NotEmpty(t, first.Stage)
	assert.NotEmpty(t, first.Category)
	assert.NotEmpty(t, first.Slug)
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func TestCalculate_VerifiedProblemsGenerateSubmitActions(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	submits := collectActionsByKind(actions, KindSubmit)
	require.Len(t, submits, 2)

	var problemIDs []int
	for _, a := range submits {
		problemIDs = append(problemIDs, a.ProblemID)
		assert.Equal(t, KindSubmit, a.Kind)
		assert.Contains(t, a.Reason, "Verified")
		assert.NotEmpty(t, a.Title)
	}
	assert.Contains(t, problemIDs, 1)
	assert.Contains(t, problemIDs, 242)

	for _, a := range collectActionsByKind(actions, KindStart) {
		assert.NotEqual(t, 1, a.ProblemID, "Verified Two Sum should not be in Start")
		assert.NotEqual(t, 242, a.ProblemID, "Verified Valid Anagram should not be in Start")
	}
}

func TestCalculate_SubmitRanksAfterContinueBeforeStart(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	continueIdx := -1
	submitIdx := -1
	startIdx := -1
	for i, a := range actions {
		switch a.Kind {
		case KindContinue:
			if continueIdx < 0 {
				continueIdx = i
			}
		case KindSubmit:
			if submitIdx < 0 {
				submitIdx = i
			}
		case KindStart:
			if startIdx < 0 {
				startIdx = i
			}
		}
	}

	require.NotEqual(t, -1, continueIdx)
	require.NotEqual(t, -1, submitIdx)
	require.NotEqual(t, -1, startIdx)
	assert.Less(t, submitIdx, continueIdx, "Submit (Verified) should rank before Continue (InProgress)")
	assert.Less(t, continueIdx, startIdx, "Continue should rank before Start")
}

func TestCalculate_NoSubmitWhenNoVerified(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusInProgress))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	submits := collectActionsByKind(actions, KindSubmit)
	assert.Empty(t, submits, "should not generate Submit actions when no Verified problems exist")
}

func TestCalculate_ManualSolveGeneratedForVerified(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	manuals := collectActionsByKind(actions, KindManualSolve)
	require.Len(t, manuals, 1)
	assert.Equal(t, 1, manuals[0].ProblemID)
	assert.Equal(t, ReasonCompletesVerified, manuals[0].ReasonType)
	assert.Contains(t, manuals[0].Reason, "unlocks dependents")
}

func TestCalculate_ReasonTypesPresent(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, s.SetProgress(ctx, 217, roadmap.StatusInProgress))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusSolved))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	for _, a := range actions {
		if a.Kind == KindExport || a.Kind == KindInspect {
			continue
		}
		assert.NotEmpty(t, a.ReasonType, "action %s should have a ReasonType", a.ID)
		assert.NotEmpty(t, a.Reason, "action %s should have a Reason", a.ID)
	}
}

func TestCalculate_NoScoreExposed(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	for _, a := range actions {
		assert.False(t, a.Priority < 0 || a.Priority > 1000, "priority should be a simple ordering index")
	}
}

func TestCalculate_DifficultySpikeAvoided(t *testing.T) {
	rm := &roadmap.Roadmap{
		ID:    "test",
		Title: "Test Roadmap",
		Stages: []roadmap.Stage{
			{ID: "foundations", Title: "Foundations", Order: 0},
		},
		Graph: roadmap.NewGraph([]*roadmap.Problem{
			{ID: 1, Title: "Easy First", Slug: "a", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "foundations"},
			{ID: 2, Title: "Hard Second", Slug: "b", Difficulty: roadmap.DifficultyHard, Category: "arrays-hashing", Stage: "foundations"},
		}),
	}
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	ctx := context.Background()
	calc := NewCalculator(s, rm)
	actions, err := calc.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)
	require.Len(t, starts, 2)
	assert.Equal(t, 1, starts[0].ProblemID, "easier problem should rank first")
	assert.Equal(t, 2, starts[1].ProblemID)
}

func TestCalculate_DirectUnlockImpactsRanking(t *testing.T) {
	rm := &roadmap.Roadmap{
		ID:    "test",
		Title: "Test Roadmap",
		Stages: []roadmap.Stage{
			{ID: "s1", Title: "Stage 1", Order: 0},
		},
		Graph: roadmap.NewGraph([]*roadmap.Problem{
			{ID: 1, Title: "No Unlock", Slug: "a", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1"},
			{ID: 2, Title: "Unlocks Many", Slug: "b", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1"},
			{ID: 3, Title: "Dep A", Slug: "c", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1", Prerequisites: []int{2}},
			{ID: 4, Title: "Dep B", Slug: "d", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1", Prerequisites: []int{2}},
		}),
	}
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	calc := NewCalculator(s, rm)
	ctx := context.Background()
	actions, err := calc.Calculate(ctx)
	require.NoError(t, err)

	starts := collectActionsByKind(actions, KindStart)
	require.Len(t, starts, 2)
	assert.Equal(t, 2, starts[0].ProblemID, "problem that unlocks more should rank first")
}

func TestCalculate_ConnectLeetCodeWhenVerified(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, s.SetProgress(ctx, 242, roadmap.StatusVerified))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	connects := collectActionsByKind(actions, KindConnectLeetCode)
	require.Len(t, connects, 1)
	assert.Equal(t, "connect_leetcode", connects[0].ID)
	assert.Contains(t, connects[0].Reason, "Verified")
}

func TestCalculate_ReviewRecommendationDoesNotCreateCycle(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.RecordSolveProvenance(ctx, &store.SolveProvenance{ProblemID: 1, Kind: "manual"}))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	reviews := collectActionsByKind(actions, KindReview)
	require.NotEmpty(t, reviews)

	cycles, err := s.GetReviewCycles(ctx)
	require.NoError(t, err)
	assert.Empty(t, cycles, "Calculate should not mutate persistence by creating Review Cycles")
}

func TestCalculate_NoConnectLeetCodeWhenAcceptedProvPresent(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, s.RecordSolveProvenance(ctx, &store.SolveProvenance{ProblemID: 1, Kind: "accepted"}))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	connects := collectActionsByKind(actions, KindConnectLeetCode)
	assert.Empty(t, connects, "ConnectLeetCode should not appear when accepted provenance exists")
}

func TestCalculate_ViewRoadmapCompletionWhenAllSolved(t *testing.T) {
	rm := &roadmap.Roadmap{
		ID:           "test",
		Title:        "Test Roadmap",
		NextRoadmaps: []string{"next-roadmap"},
		Stages: []roadmap.Stage{
			{ID: "s1", Title: "Stage 1", Order: 0},
		},
		Graph: roadmap.NewGraph([]*roadmap.Problem{
			{ID: 1, Title: "Only Problem", Slug: "a", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1"},
		}),
	}
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	require.NoError(t, s.SetProgress(context.Background(), 1, roadmap.StatusSolved))
	require.NoError(t, s.RecordSolveProvenance(context.Background(), &store.SolveProvenance{ProblemID: 1, Kind: "accepted"}))

	calc := NewCalculator(s, rm)
	ctx := context.Background()
	actions, err := calc.Calculate(ctx)
	require.NoError(t, err)

	completions := collectActionsByKind(actions, KindViewRoadmapCompletion)
	require.Len(t, completions, 1)
	assert.Contains(t, completions[0].Reason, "next-roadmap")
}

func TestCalculate_NoCompletionWhenNotAllSolved(t *testing.T) {
	rm := &roadmap.Roadmap{
		ID:           "test",
		Title:        "Test Roadmap",
		NextRoadmaps: []string{"next-roadmap"},
		Stages: []roadmap.Stage{
			{ID: "s1", Title: "Stage 1", Order: 0},
		},
		Graph: roadmap.NewGraph([]*roadmap.Problem{
			{ID: 1, Title: "Problem 1", Slug: "a", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1"},
			{ID: 2, Title: "Problem 2", Slug: "b", Difficulty: roadmap.DifficultyEasy, Category: "arrays-hashing", Stage: "s1"},
		}),
	}
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	require.NoError(t, s.SetProgress(context.Background(), 1, roadmap.StatusSolved))

	calc := NewCalculator(s, rm)
	ctx := context.Background()
	actions, err := calc.Calculate(ctx)
	require.NoError(t, err)

	completions := collectActionsByKind(actions, KindViewRoadmapCompletion)
	assert.Empty(t, completions)
}

func TestCalculate_ContinueActionHasReasonType(t *testing.T) {
	c, s := newTestCalculator(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))

	actions, err := c.Calculate(ctx)
	require.NoError(t, err)

	continues := collectActionsByKind(actions, KindContinue)
	require.Len(t, continues, 1)
	assert.Equal(t, ReasonContinuesInProgress, continues[0].ReasonType)
}
