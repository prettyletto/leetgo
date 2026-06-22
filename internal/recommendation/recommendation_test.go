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
