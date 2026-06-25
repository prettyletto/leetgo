package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettyletto/leetgo/internal/gamification"
)

type Notification struct {
	Message   string
	Timestamp time.Time
}

type NotificationManager struct {
	notifications []Notification
	maxAge        time.Duration
	palette       Palette
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		maxAge: 5 * time.Second,
	}
}

func (m *NotificationManager) SetPalette(p Palette) {
	m.palette = p
}

func (m *NotificationManager) Add(msg string) {
	m.notifications = append(m.notifications, Notification{
		Message:   msg,
		Timestamp: time.Now(),
	})
}

func (m *NotificationManager) AddAchievement(a gamification.Achievement) {
	msg := fmt.Sprintf("%s Achievement Unlocked: %s", a.Icon, a.Name)
	m.Add(msg)
}

func (m *NotificationManager) Render() string {
	m.prune()
	if len(m.notifications) == 0 {
		return ""
	}

	latest := m.notifications[len(m.notifications)-1]

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Primary).
		Foreground(m.palette.Primary).
		Padding(0, 1).
		Bold(true)

	return style.Render(latest.Message)
}

func (m *NotificationManager) prune() {
	cutoff := time.Now().Add(-m.maxAge)
	var active []Notification
	for _, n := range m.notifications {
		if n.Timestamp.After(cutoff) {
			active = append(active, n)
		}
	}
	m.notifications = active
}
