package views

import (
	"fmt"
	"strings"
	"time"
)

type RewardMoment struct {
	Title                string
	Subject              string
	XP                   int
	Reward               string
	Duration             time.Duration
	Unlocked             []string
	Next                 string
	Reason               string
	Actions              []string
	AdditionalHighlights []string
}

func RenderRewardMoment(moment RewardMoment, palette Palette) string {
	return PixelFrame(moment.Title, renderRewardMomentBody(moment), palette)
}

func RenderCLIRewardMoment(moment RewardMoment) string {
	body := renderRewardMomentBody(moment)
	if moment.Title == "" {
		return "Reward Moment\n" + body
	}
	return "Reward Moment\n" + moment.Title + "\n" + body
}

func renderRewardMomentBody(moment RewardMoment) string {
	var lines []string
	if moment.Subject != "" {
		lines = append(lines, moment.Subject)
	}

	if moment.XP > 0 || moment.Reward != "" || moment.Duration > 0 {
		lines = append(lines, "", "Reward")
		if moment.XP > 0 {
			lines = append(lines, fmt.Sprintf("+%d XP", moment.XP))
		}
		if moment.Reward != "" {
			lines = append(lines, moment.Reward)
		}
		if moment.Duration > 0 {
			lines = append(lines, "Duration: "+moment.Duration.Round(time.Second).String())
		}
	}

	if len(moment.Unlocked) > 0 || moment.Next != "" || moment.Reason != "" {
		lines = append(lines, "", "Progress")
		if len(moment.Unlocked) > 0 {
			lines = append(lines, "Unlocked: "+strings.Join(moment.Unlocked, ", "))
		}
		if moment.Next != "" {
			lines = append(lines, "Next: "+moment.Next)
		}
		if moment.Reason != "" {
			lines = append(lines, "Why: "+moment.Reason)
		}
	}

	if len(moment.AdditionalHighlights) > 0 {
		lines = append(lines, "")
		lines = append(lines, moment.AdditionalHighlights...)
	}

	if len(moment.Actions) > 0 {
		lines = append(lines, "", "Actions")
		lines = append(lines, moment.Actions...)
	}

	return strings.Join(lines, "\n")
}
