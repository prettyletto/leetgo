package roadmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph_IsUnlocked(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "Two Sum", Prerequisites: nil},
		{ID: 15, Title: "3Sum", Prerequisites: []int{1}},
		{ID: 49, Title: "Group Anagrams", Prerequisites: nil},
		{ID: 128, Title: "Longest Consecutive", Prerequisites: []int{1, 49}},
	}
	g := NewGraph(problems)

	tests := []struct {
		name     string
		id       int
		solved   map[int]bool
		expected bool
	}{
		{"no prereqs", 1, map[int]bool{}, true},
		{"prereq not solved", 15, map[int]bool{}, false},
		{"prereq solved", 15, map[int]bool{1: true}, true},
		{"multiple prereqs all solved", 128, map[int]bool{1: true, 49: true}, true},
		{"multiple prereqs partial", 128, map[int]bool{1: true}, false},
		{"unknown problem", 999, map[int]bool{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, g.IsUnlocked(tt.id, tt.solved))
		})
	}
}

func TestGraph_Available(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "Two Sum", Prerequisites: nil},
		{ID: 15, Title: "3Sum", Prerequisites: []int{1}},
		{ID: 49, Title: "Group Anagrams", Prerequisites: nil},
	}
	g := NewGraph(problems)

	solved := map[int]bool{1: true}
	available := g.Available(solved)

	ids := make([]int, len(available))
	for i, p := range available {
		ids[i] = p.ID
	}
	assert.ElementsMatch(t, []int{15, 49}, ids)
}

func TestGraph_IsUnlocked_VerifiedSatisfiesPrereq(t *testing.T) {
	problems := []*Problem{
		{ID: 1, Title: "Two Sum", Prerequisites: nil},
		{ID: 15, Title: "3Sum", Prerequisites: []int{1}},
	}
	g := NewGraph(problems)

	// Verified satisfies prerequisite
	solved := map[int]bool{1: true}
	assert.True(t, g.IsUnlocked(15, solved))
}
