package leetcode

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var ErrUnsupportedBrowser = errors.New("automated LeetCode Session setup requires Chrome or another Chromium-based browser; Firefox is not supported yet")

type browser struct {
	Name string
	Path string
}

type browserDetector struct {
	goos     string
	lookPath func(string) (string, error)
	getenv   func(string) string
	stat     func(string) (os.FileInfo, error)
}

func detectChromiumBrowser() (*browser, error) {
	return detectChromiumBrowserWith(browserDetector{
		goos:     runtime.GOOS,
		lookPath: exec.LookPath,
		getenv:   os.Getenv,
		stat:     os.Stat,
	})
}

func detectChromiumBrowserWith(d browserDetector) (*browser, error) {
	if d.lookPath == nil {
		d.lookPath = exec.LookPath
	}
	if d.getenv == nil {
		d.getenv = os.Getenv
	}
	if d.stat == nil {
		d.stat = os.Stat
	}

	switch d.goos {
	case "linux":
		return detectLinuxBrowser(d)
	case "windows":
		return detectWindowsBrowser(d)
	default:
		return nil, ErrUnsupportedBrowser
	}
}

func detectLinuxBrowser(d browserDetector) (*browser, error) {
	candidates := []browser{
		{Name: "Chrome", Path: "google-chrome"},
		{Name: "Chrome", Path: "google-chrome-stable"},
		{Name: "Chromium", Path: "chromium"},
		{Name: "Chromium", Path: "chromium-browser"},
		{Name: "Brave", Path: "brave-browser"},
		{Name: "Edge", Path: "microsoft-edge"},
		{Name: "Vivaldi", Path: "vivaldi"},
	}

	for _, candidate := range candidates {
		path, err := d.lookPath(candidate.Path)
		if err == nil {
			candidate.Path = path
			return &candidate, nil
		}
	}
	return nil, ErrUnsupportedBrowser
}

func detectWindowsBrowser(d browserDetector) (*browser, error) {
	programFiles := d.getenv("ProgramFiles")
	programFilesX86 := d.getenv("ProgramFiles(x86)")
	localAppData := d.getenv("LocalAppData")

	pathCandidates := []browser{
		{Name: "Chrome", Path: filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe")},
		{Name: "Chrome", Path: filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe")},
		{Name: "Chrome", Path: filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe")},
		{Name: "Edge", Path: filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe")},
		{Name: "Edge", Path: filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe")},
		{Name: "Edge", Path: filepath.Join(localAppData, "Microsoft", "Edge", "Application", "msedge.exe")},
		{Name: "Brave", Path: filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "Application", "brave.exe")},
		{Name: "Brave", Path: filepath.Join(programFiles, "BraveSoftware", "Brave-Browser", "Application", "brave.exe")},
		{Name: "Vivaldi", Path: filepath.Join(localAppData, "Vivaldi", "Application", "vivaldi.exe")},
	}
	for _, candidate := range pathCandidates {
		if candidate.Path == "" {
			continue
		}
		info, err := d.stat(candidate.Path)
		if err == nil && !info.IsDir() {
			return &candidate, nil
		}
	}

	pathCommands := []browser{
		{Name: "Chrome", Path: "chrome.exe"},
		{Name: "Edge", Path: "msedge.exe"},
		{Name: "Brave", Path: "brave.exe"},
		{Name: "Vivaldi", Path: "vivaldi.exe"},
	}
	for _, candidate := range pathCommands {
		path, err := d.lookPath(candidate.Path)
		if err == nil {
			candidate.Path = path
			return &candidate, nil
		}
	}
	return nil, ErrUnsupportedBrowser
}

func unsupportedBrowserMessage() string {
	return fmt.Sprintf("%v. Install Chrome, Chromium, Brave, Edge, or Vivaldi, then run `leetgo auth`.", ErrUnsupportedBrowser)
}
