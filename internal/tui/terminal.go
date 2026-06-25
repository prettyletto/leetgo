package tui

import (
	"sync"

	"github.com/muesli/termenv"
)

type terminalProfile struct {
	HasDarkBackground bool
	Profile           termenv.Profile
}

var (
	cachedTerminal terminalProfile
	detectOnce     sync.Once
)

func detectTerminal() terminalProfile {
	detectOnce.Do(func() {
		output := termenv.ColorProfile()
		cachedTerminal = terminalProfile{
			HasDarkBackground: termenv.HasDarkBackground(),
			Profile:           output,
		}
	})
	return cachedTerminal
}
