package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func editorCommand(editor string, paths []string, fallbackDir string) *exec.Cmd {
	if strings.TrimSpace(editor) == "" {
		return exec.Command("xdg-open", fallbackDir)
	}

	parts := strings.Fields(editor)
	args := append([]string{}, parts[1:]...)
	args = append(args, paths...)

	if shouldWrapInTerminal(parts[0]) {
		terminal := terminalLauncher()
		if len(terminal) > 0 {
			termArgs := append([]string{}, terminal[1:]...)
			termArgs = append(termArgs, parts[0])
			termArgs = append(termArgs, args...)
			return exec.Command(terminal[0], termArgs...)
		}
	}

	return exec.Command(parts[0], args...)
}

func shouldWrapInTerminal(command string) bool {
	switch filepath.Base(command) {
	case "nvim", "vim", "vi", "nano", "emacs", "micro", "helix", "hx":
		return true
	default:
		return false
	}
}

func terminalLauncher() []string {
	if terminal := strings.TrimSpace(os.Getenv("TERMINAL")); terminal != "" {
		return terminalCommand(terminal)
	}

	for _, terminal := range []string{"xdg-terminal-exec", "alacritty", "ghostty", "kitty", "foot", "wezterm"} {
		path, err := exec.LookPath(terminal)
		if err == nil {
			return terminalCommand(path)
		}
	}
	return nil
}

func terminalCommand(terminal string) []string {
	switch filepath.Base(terminal) {
	case "xdg-terminal-exec":
		return []string{terminal}
	case "alacritty", "ghostty", "kitty", "foot":
		return []string{terminal, "-e"}
	case "wezterm":
		return []string{terminal, "start", "--"}
	default:
		return []string{terminal, "-e"}
	}
}
