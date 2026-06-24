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

func TestPanel_Render(t *testing.T) {
	view := Panel("Quest Board", "Two Sum", Palette{}, false)
	assert.Contains(t, view, "Quest Board")
	assert.Contains(t, view, "Two Sum")
}

func TestPixelFrame_Render(t *testing.T) {
	view := PixelFrame("Character HUD", "LVL 2", Palette{})
	assert.Contains(t, view, "Character HUD")
	assert.Contains(t, view, "LVL 2")
}

func TestStatusPill_Render(t *testing.T) {
	assert.Contains(t, StatusPill("Solved", ""), "[Solved]")
}

func TestProgressBar_Render(t *testing.T) {
	assert.Equal(t, "##--", ProgressBar(1, 2, 4, "#", "-"))
	assert.Equal(t, "----", ProgressBar(0, 0, 4, "#", "-"))
}

func TestKeytipFooter_Render(t *testing.T) {
	view := KeytipFooter(map[string]string{"j": "down", "k": "up"}, []string{"j", "k"}, Palette{})
	assert.Contains(t, view, "j")
	assert.Contains(t, view, "down")
	assert.Contains(t, view, "k")
	assert.Contains(t, view, "up")
}

func TestUnsupportedSize_Render(t *testing.T) {
	view := UnsupportedSize(50, 12, Palette{})
	assert.Contains(t, view, "Unsupported Size")
	assert.Contains(t, view, "50x12")
	assert.Contains(t, view, "60x18")
	assert.Contains(t, view, "leetgo next")
	assert.Contains(t, view, "leetgo info .")
	assert.Contains(t, view, "leetgo test .")
}

func TestRenderRewardMoment(t *testing.T) {
	view := RenderRewardMoment(RewardMoment{
		Title:    "Problem Solved",
		Subject:  "#1 Two Sum",
		XP:       10,
		Duration: 2 * time.Second,
		Unlocked: []string{"#49 Group Anagrams"},
		Next:     "Group Anagrams",
		Reason:   "unlocks hashing practice",
		Actions:  []string{"leetgo next"},
	}, Palette{})

	assert.Contains(t, view, "Problem Solved")
	assert.Contains(t, view, "+10 XP")
	assert.Contains(t, view, "Unlocked: #49 Group Anagrams")
	assert.Contains(t, view, "Actions")
}

func TestRenderCLIRewardMoment(t *testing.T) {
	view := RenderCLIRewardMoment(RewardMoment{Title: "Manual Solve", Subject: "#1 Two Sum", Reward: "No XP awarded"})
	assert.Contains(t, view, "Reward Moment")
	assert.Contains(t, view, "Manual Solve")
	assert.Contains(t, view, "No XP awarded")
}
