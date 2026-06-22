package roadmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologicalSort(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "Two Sum", Prerequisites: nil},
		{ID: 15, Title: "3Sum", Prerequisites: []int{1}},
		{ID: 49, Title: "Group Anagrams", Prerequisites: []int{1}},
		{ID: 128, Title: "Longest Consecutive", Prerequisites: []int{1, 49}},
	}
	g := NewGraph(problems)

	sorted, err := g.TopologicalSort()
	require.NoError(t, err)
	require.Len(t, sorted, 4)

	positions := make(map[int]int)
	for i, p := range sorted {
		positions[p.ID] = i
	}

	assert.Less(t, positions[1], positions[15])
	assert.Less(t, positions[1], positions[49])
	assert.Less(t, positions[1], positions[128])
	assert.Less(t, positions[49], positions[128])
}

func TestTopologicalSort_DeterministicOrder(t *testing.T) {
	problems := []*Problem{
		{ID: 30, Title: "C"},
		{ID: 10, Title: "A"},
		{ID: 20, Title: "B"},
		{ID: 40, Title: "D", Prerequisites: []int{10, 20}},
	}
	g := NewGraph(problems)

	for i := 0; i < 20; i++ {
		sorted, err := g.TopologicalSort()
		require.NoError(t, err)
		require.Len(t, sorted, 4)
		assert.Equal(t, 10, sorted[0].ID)
		assert.Equal(t, 20, sorted[1].ID)
		assert.Equal(t, 30, sorted[2].ID)
		assert.Equal(t, 40, sorted[3].ID)
	}
}

func TestTopologicalSort_CycleDetected(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "A", Prerequisites: []int{2}},
		{ID: 2, Title: "B", Prerequisites: []int{1}},
	}
	g := NewGraph(problems)

	_, err := g.TopologicalSort()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestTopologicalSort_UnknownPrereq(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "A", Prerequisites: []int{999}},
	}
	g := NewGraph(problems)

	_, err := g.TopologicalSort()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown prerequisite")
}
