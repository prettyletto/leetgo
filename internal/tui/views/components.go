package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	MinSupportedWidth  = 60
	MinSupportedHeight = 18
)

type Palette struct {
	Primary lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color
	Danger  lipgloss.Color
	Muted   lipgloss.Color
	Border  lipgloss.Color
	XP      lipgloss.Color
	Review  lipgloss.Color
}

func Panel(title, body string, palette Palette, focused bool) string {
	border := lipgloss.NormalBorder()
	borderColor := palette.Border
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.Primary)
	if focused {
		border = lipgloss.RoundedBorder()
		borderColor = palette.Primary
		titleStyle = titleStyle.Copy().Foreground(palette.Primary)
	}
	content := body
	if strings.TrimSpace(title) != "" {
		titleView := titleStyle.Render(title)
		if strings.TrimSpace(body) == "" {
			content = titleView
		} else {
			content = titleView + "\n\n" + body
		}
	}
	return lipgloss.NewStyle().Border(border).BorderForeground(borderColor).Padding(0, 1).Render(content)
}

func PixelFrame(title, body string, palette Palette, focused bool) string {
	border := lipgloss.BlockBorder()
	borderColor := palette.Border
	if focused {
		border = lipgloss.ThickBorder()
		borderColor = palette.Primary
	}
	content := body
	if strings.TrimSpace(title) != "" {
		content = lipgloss.NewStyle().Bold(true).Foreground(palette.Primary).Render(title) + "\n" + body
	}
	return lipgloss.NewStyle().Border(border).BorderForeground(borderColor).Padding(0, 1).Render(content)
}

func StatusPill(label string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render("[" + label + "]")
}

func ProgressBar(done, total, width int, fill, empty string) string {
	if width <= 0 {
		return ""
	}
	if fill == "" {
		fill = "#"
	}
	if empty == "" {
		empty = "-"
	}
	if total <= 0 {
		return strings.Repeat(empty, width)
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	filled := done * width / total
	return strings.Repeat(fill, filled) + strings.Repeat(empty, width-filled)
}

func KeytipFooter(keys map[string]string, order []string, palette Palette) string {
	parts := make([]string, 0, len(order))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.Primary)
	textStyle := lipgloss.NewStyle().Foreground(palette.Muted)
	for _, key := range order {
		label, ok := keys[key]
		if !ok {
			continue
		}
		parts = append(parts, keyStyle.Render(key)+" "+textStyle.Render(label))
	}
	return strings.Join(parts, textStyle.Render("  •  "))
}

func UnsupportedSize(width, height int, palette Palette) string {
	body := fmt.Sprintf("Leetgo needs a little more room.\n\nCurrent size: %dx%d\nMinimum size: %dx%d\n\nResize your terminal, or use these CLI paths instead:\nleetgo next\nleetgo info .\nleetgo test .", width, height, MinSupportedWidth, MinSupportedHeight)
	return Panel("Unsupported Size", body, palette, true)
}
