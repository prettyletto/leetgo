package views

import (
	"testing"
	"time"

	"github.com/prettyletto/leetgo/internal/gamification"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestStatsBar_Render(t *testing.T) {
	stats := &store.Stats{
		TotalXP:       150,
		Level:         2,
		Streak:        5,
		LongestStreak: 10,
		Solved:        8,
		Total:         150,
	}

	bar := NewStatsBar(stats)
	view := bar.Render()
	assert.Contains(t, view, "LVL 2")
	assert.Contains(t, view, "150 XP")
	assert.Contains(t, view, "8/150 solved")
}

func TestStatsBar_NilStats(t *testing.T) {
	bar := NewStatsBar(nil)
	assert.Empty(t, bar.Render())
}

func TestHeatmapView_Render(t *testing.T) {
	days := []time.Time{
		time.Now().AddDate(0, 0, -1),
		time.Now(),
	}

	v := NewHeatmapView(days)
	view := v.Render()
	assert.Contains(t, view, "Less")
	assert.Contains(t, view, "More")
	assert.Contains(t, view, "Total days: 2")
}

func TestHeatmapView_Empty(t *testing.T) {
	v := NewHeatmapView(nil)
	view := v.Render()
	assert.Contains(t, view, "Total days: 0")
}

func TestNotificationManager_Add(t *testing.T) {
	nm := NewNotificationManager()
	nm.Add("test message")
	assert.Contains(t, nm.Render(), "test message")
}

func TestNotificationManager_AddAchievement(t *testing.T) {
	nm := NewNotificationManager()
	a := gamification.Achievements["first_solve"]
	nm.AddAchievement(a)
	assert.Contains(t, nm.Render(), "First Blood")
}

func TestNotificationManager_Prune(t *testing.T) {
	nm := NewNotificationManager()
	nm.maxAge = 1 * time.Millisecond
	nm.Add("old message")
	time.Sleep(5 * time.Millisecond)
	assert.Empty(t, nm.Render())
}
