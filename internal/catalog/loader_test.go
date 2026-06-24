package catalog

import (
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	g, err := Load()
	require.NoError(t, err)
	assert.NotEmpty(t, g.Problems)

	twoSum := g.Problems[1]
	require.NotNil(t, twoSum)
	assert.Equal(t, "Two Sum", twoSum.Title)
	assert.Empty(t, twoSum.Prerequisites)

	groupAnagrams := g.Problems[49]
	require.NotNil(t, groupAnagrams)
	assert.Contains(t, groupAnagrams.Prerequisites, 242)
}

func TestLoadRoadmap(t *testing.T) {
	rm, err := LoadRoadmap("interview-sprint")
	require.NoError(t, err)
	assert.Equal(t, "interview-sprint", rm.ID)
	assert.Equal(t, "Interview Sprint", rm.Title)
	assert.NotEmpty(t, rm.Stages)
	assert.NotEmpty(t, rm.Graph.Problems)
}

func TestListRoadmaps(t *testing.T) {
	roadmaps, err := ListRoadmaps()
	require.NoError(t, err)
	ids := make([]string, len(roadmaps))
	for i, rm := range roadmaps {
		ids[i] = rm.ID
	}
	assert.Contains(t, ids, "from-zero-to-hero")
	assert.Contains(t, ids, "interview-sprint")
	assert.Contains(t, ids, "hard-mode")
}

func TestParse(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
    prerequisites: []
  - id: 15
    title: "3Sum"
    slug: "3sum"
    difficulty: medium
    category: "two-pointers"
    prerequisites: [1]
`)
	g, err := Parse(yaml)
	require.NoError(t, err)
	assert.Len(t, g.Problems, 2)
	assert.Equal(t, []int{1}, g.Problems[15].Prerequisites)
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte("not: [valid: yaml"))
	assert.Error(t, err)
}

func TestParseRoadmap_DuplicateProblemID(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
  - id: 1
    title: "Two Sum Again"
    slug: "two-sum-again"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "duplicate problem id")
}

func TestParseRoadmap_UnknownStage(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
stages:
  - id: foundations
    title: "Foundations"
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
    stage: missing
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "unknown stage")
}

func TestParseRoadmap_UnknownPrerequisite(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
    prerequisites: [999]
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "unknown prerequisite")
}

func TestParseRoadmap_MissingTagline(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "tagline is required")
}

func TestParseRoadmap_MissingAudience(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "audience is required")
}

func TestParseRoadmap_MissingPromise(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "promise is required")
}

func TestParseRoadmap_InvalidHighlightsCount(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "highlights must contain 2-3 items")
}

func TestParseRoadmap_InvalidDifficultyMix(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
difficulty_mix:
  easy: 50
  medium: 30
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "difficulty_mix must add to 100")
}

func TestParseRoadmap_ValidMetadata(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test tagline"
audience: "test audience"
promise: "test promise"
recommended: true
estimated_hours: 50
roadmap_time_estimate: "4-8 weeks at 3-5 Problems/week"
difficulty_mix:
  easy: 40
  medium: 40
  hard: 20
highlights:
  - "highlight one"
  - "highlight two"
  - "highlight three"
next_roadmaps:
  - "from-zero-to-hero"
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
    practice_focus: "Hash table lookups"
    problem_time_estimate: "15-30 minutes"
    summary: "Find two numbers that sum to a target."
    why_now: "Introduces hash tables which are fundamental."
    unlock_impact: "Prepares for 3Sum and Two Sum II."
    hints:
      - "Consider using a hash map to store visited numbers."
      - "Think about what the complement of each number would be."
`)
	rm, err := ParseRoadmap(yaml)
	require.NoError(t, err)
	assert.Equal(t, "test", rm.ID)
	assert.Equal(t, "test tagline", rm.Tagline)
	assert.Equal(t, "test audience", rm.Audience)
	assert.Equal(t, "test promise", rm.Promise)
	assert.True(t, rm.Recommended)
	assert.Equal(t, 50, rm.EstimatedHours)
	assert.Equal(t, 40, rm.DifficultyMix[roadmap.DifficultyEasy])
	assert.Equal(t, 40, rm.DifficultyMix[roadmap.DifficultyMedium])
	assert.Equal(t, 20, rm.DifficultyMix[roadmap.DifficultyHard])
	assert.Len(t, rm.Highlights, 3)
	assert.Equal(t, "4-8 weeks at 3-5 Problems/week", rm.RoadmapTimeEstimate)

	p := rm.Graph.Problems[1]
	require.NotNil(t, p)
	assert.Equal(t, "Hash table lookups", p.PracticeFocus)
	assert.Equal(t, "15-30 minutes", p.ProblemTimeEstimate)
	assert.Equal(t, "Find two numbers that sum to a target.", p.Summary)
	assert.Equal(t, "Introduces hash tables which are fundamental.", p.WhyNow)
	assert.Equal(t, "Prepares for 3Sum and Two Sum II.", p.UnlockImpact)
	assert.Len(t, p.Hints, 2)
	assert.Equal(t, "Consider using a hash map to store visited numbers.", p.Hints[0])
	assert.Equal(t, "Think about what the complement of each number would be.", p.Hints[1])
}

func TestParseRoadmap_EstimatedHoursZero(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
estimated_hours: 0
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "estimated_hours must be positive")
}

func TestParseRoadmap_EstimatedHoursNotSet(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	rm, err := ParseRoadmap(yaml)
	require.NoError(t, err)
	assert.Equal(t, 0, rm.EstimatedHours)
}

func TestListRoadmaps_HasMetadata(t *testing.T) {
	roadmaps, err := ListRoadmaps()
	require.NoError(t, err)

	for _, rm := range roadmaps {
		assert.NotEmpty(t, rm.Tagline, "roadmap %s missing tagline", rm.ID)
		assert.NotEmpty(t, rm.Audience, "roadmap %s missing audience", rm.ID)
		assert.NotEmpty(t, rm.Promise, "roadmap %s missing promise", rm.ID)
		assert.GreaterOrEqual(t, len(rm.Highlights), 2, "roadmap %s should have at least 2 highlights", rm.ID)
		assert.LessOrEqual(t, len(rm.Highlights), 3, "roadmap %s should have at most 3 highlights", rm.ID)
		if rm.ID == "from-zero-to-hero" {
			assert.NotEmpty(t, rm.RoadmapTimeEstimate, "from-zero-to-hero should have roadmap_time_estimate")
		}
	}
}

func TestListRoadmaps_NextRoadmaps(t *testing.T) {
	roadmaps, err := ListRoadmaps()
	require.NoError(t, err)

	ids := make(map[string]bool, len(roadmaps))
	for _, rm := range roadmaps {
		ids[rm.ID] = true
	}
	for _, rm := range roadmaps {
		for _, nextID := range rm.NextRoadmaps {
			assert.True(t, ids[nextID], "roadmap %s references unknown next_roadmap %q", rm.ID, nextID)
		}
	}
}

func TestValidateRoadmapSet_BadNextRoadmap(t *testing.T) {
	rm1 := &roadmap.Roadmap{ID: "a", Recommended: false, NextRoadmaps: []string{"missing"}}
	rm2 := &roadmap.Roadmap{ID: "b", Recommended: false}
	err := ValidateRoadmapSet([]*roadmap.Roadmap{rm1, rm2})
	assert.ErrorContains(t, err, "unknown next_roadmap")
}

func TestParseRoadmap_EmptyRoadmapTimeEstimate(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
roadmap_time_estimate: ""
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	_, err := ParseRoadmap(yaml)
	assert.ErrorContains(t, err, "roadmap_time_estimate must be non-empty")
}

func TestParseRoadmap_ProblemMetadataOptional(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
  - id: 2
    title: "With Metadata"
    slug: "with-metadata"
    difficulty: medium
    category: "arrays-hashing"
    practice_focus: "Hash maps"
    problem_time_estimate: "20-30 min"
    summary: "A problem with metadata."
    why_now: "Because it matters."
    unlock_impact: "Unlocks more complexity."
    hints:
      - "Hint one"
      - "Hint two"
`)
	rm, err := ParseRoadmap(yaml)
	require.NoError(t, err)

	p1 := rm.Graph.Problems[1]
	require.NotNil(t, p1)
	assert.Empty(t, p1.PracticeFocus)
	assert.Empty(t, p1.Summary)
	assert.Empty(t, p1.Hints)

	p2 := rm.Graph.Problems[2]
	require.NotNil(t, p2)
	assert.Equal(t, "Hash maps", p2.PracticeFocus)
	assert.Equal(t, "20-30 min", p2.ProblemTimeEstimate)
	assert.Equal(t, "A problem with metadata.", p2.Summary)
	assert.Equal(t, "Because it matters.", p2.WhyNow)
	assert.Equal(t, "Unlocks more complexity.", p2.UnlockImpact)
	assert.Len(t, p2.Hints, 2)
}

func TestParseRoadmap_NextRoadmaps(t *testing.T) {
	yaml := []byte(`
id: test
title: "Test"
tagline: "test"
audience: "test"
promise: "test"
highlights: ["a", "b"]
next_roadmaps:
  - "from-zero-to-hero"
  - "hard-mode"
problems:
  - id: 1
    title: "Two Sum"
    slug: "two-sum"
    difficulty: easy
    category: "arrays-hashing"
`)
	rm, err := ParseRoadmap(yaml)
	require.NoError(t, err)
	assert.Len(t, rm.NextRoadmaps, 2)
	assert.Equal(t, "from-zero-to-hero", rm.NextRoadmaps[0])
	assert.Equal(t, "hard-mode", rm.NextRoadmaps[1])
}

func TestListRoadmaps_SingleRecommended(t *testing.T) {
	roadmaps, err := ListRoadmaps()
	require.NoError(t, err)

	recommendedCount := 0
	for _, rm := range roadmaps {
		if rm.Recommended {
			recommendedCount++
		}
	}
	assert.Equal(t, 1, recommendedCount, "exactly one roadmap should be recommended")
}

func TestValidateRoadmapSet_DuplicateRecommended(t *testing.T) {
	rm1 := &roadmap.Roadmap{ID: "a", Recommended: true}
	rm2 := &roadmap.Roadmap{ID: "b", Recommended: true}

	err := ValidateRoadmapSet([]*roadmap.Roadmap{rm1, rm2})
	assert.ErrorContains(t, err, "only one bundled roadmap may be marked recommended")
}

func TestValidateRoadmapSet_NoRecommended(t *testing.T) {
	rm1 := &roadmap.Roadmap{ID: "a", Recommended: false}
	rm2 := &roadmap.Roadmap{ID: "b", Recommended: false}

	err := ValidateRoadmapSet([]*roadmap.Roadmap{rm1, rm2})
	require.NoError(t, err)
}

func TestLoadRoadmap_HasMetadata(t *testing.T) {
	rm, err := LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	assert.Equal(t, "from-zero-to-hero", rm.ID)
	assert.NotEmpty(t, rm.Tagline)
	assert.NotEmpty(t, rm.Audience)
	assert.NotEmpty(t, rm.Promise)
	assert.True(t, rm.Recommended)
	assert.Equal(t, 80, rm.EstimatedHours)
	assert.Len(t, rm.Highlights, 3)
}
