package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type editorLaunchMode int

const (
	editorLaunchAttached editorLaunchMode = iota
	editorLaunchDetached
)

func editorCommand(editor string, paths []string, workingDir string, mode editorLaunchMode) *exec.Cmd {
	mode = editorEffectiveLaunchMode(editor, mode)
	if strings.TrimSpace(editor) == "" {
		cmd := exec.Command("xdg-open", workingDir)
		cmd.Dir = workingDir
		if mode == editorLaunchDetached {
			detachEditorCommand(cmd)
		}
		return cmd
	}

	parts := strings.Fields(editor)
	args := append([]string{}, parts[1:]...)
	args = append(args, paths...)

	if mode == editorLaunchDetached && shouldWrapInTerminal(parts[0]) {
		if cmd := xdgTerminalExecCmd(parts[0], args); cmd != nil {
			cmd.Dir = workingDir
			return cmd
		}
		if cmd := fallbackTerminalCmd(parts[0], args); cmd != nil {
			cmd.Dir = workingDir
			return cmd
		}
	}

	cmd := exec.Command(parts[0], args...)
	cmd.Dir = workingDir
	if mode == editorLaunchDetached {
		detachEditorCommand(cmd)
	}
	return cmd
}

func editorEffectiveLaunchMode(editor string, mode editorLaunchMode) editorLaunchMode {
	if mode != editorLaunchAttached {
		return mode
	}
	parts := strings.Fields(editor)
	if len(parts) == 0 || !shouldWrapInTerminal(parts[0]) {
		return editorLaunchDetached
	}
	return editorLaunchAttached
}

func xdgTerminalExecCmd(editor string, editorArgs []string) *exec.Cmd {
	path, err := exec.LookPath("xdg-terminal-exec")
	if err != nil {
		return nil
	}
	shellArgs := buildShellEditorCmd(editor, editorArgs)
	cmd := exec.Command(path, shellArgs...)
	detachEditorCommand(cmd)
	return cmd
}

func fallbackTerminalCmd(editor string, editorArgs []string) *exec.Cmd {
	terminal := directTerminalLauncher()
	if len(terminal) == 0 {
		return nil
	}
	shellArgs := buildShellEditorCmd(editor, editorArgs)
	allArgs := append([]string{}, terminal[1:]...)
	allArgs = append(allArgs, shellArgs...)
	cmd := exec.Command(terminal[0], allArgs...)
	detachEditorCommand(cmd)
	return cmd
}

func detachEditorCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func startEditorCommand(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func buildShellEditorCmd(editor string, editorArgs []string) []string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmdLine := editorShellCommandLine(editor, editorArgs)
	return []string{shell, "-i", "-c", cmdLine}
}

func editorShellCommandLine(editor string, editorArgs []string) string {
	parts := make([]string, 0, len(editorArgs)+1)
	parts = append(parts, shellEscape(editor))
	for _, a := range editorArgs {
		parts = append(parts, shellEscape(a))
	}
	return strings.Join(parts, " ")
}

func shellEscape(s string) string {
	if s == "" {
		return "''"
	}
	for _, ch := range s {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || strings.ContainsRune(".-_/:=+@%", ch)) {
			goto quoted
		}
	}
	return s
quoted:
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}

func directTerminalLauncher() []string {
	if terminal := strings.TrimSpace(os.Getenv("TERMINAL")); terminal != "" {
		return terminalCommand(terminal)
	}
	for _, terminal := range []string{"ghostty", "alacritty", "kitty", "foot", "wezterm"} {
		path, err := exec.LookPath(terminal)
		if err == nil {
			return terminalCommand(path)
		}
	}
	return nil
}

func shouldWrapInTerminal(command string) bool {
	switch filepath.Base(command) {
	case "nvim", "vim", "vi", "nano", "emacs", "micro", "helix", "hx":
		return true
	default:
		return false
	}
}

func terminalCommand(terminal string) []string {
	switch filepath.Base(terminal) {
	case "alacritty", "ghostty", "kitty", "foot":
		return []string{terminal, "-e"}
	case "wezterm":
		return []string{terminal, "start", "--"}
	default:
		return []string{terminal, "-e"}
	}
}
