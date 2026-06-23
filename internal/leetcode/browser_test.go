package leetcode

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectChromiumBrowser_LinuxPrefersChrome(t *testing.T) {
	d := fakeDetector("linux", map[string]string{
		"google-chrome": "/usr/bin/google-chrome",
		"chromium":      "/usr/bin/chromium",
	}, nil, nil)

	b, err := detectChromiumBrowserWith(d)

	require.NoError(t, err)
	assert.Equal(t, "Chrome", b.Name)
	assert.Equal(t, "/usr/bin/google-chrome", b.Path)
}

func TestDetectChromiumBrowser_LinuxFallback(t *testing.T) {
	d := fakeDetector("linux", map[string]string{
		"brave-browser": "/usr/bin/brave-browser",
	}, nil, nil)

	b, err := detectChromiumBrowserWith(d)

	require.NoError(t, err)
	assert.Equal(t, "Brave", b.Name)
}

func TestDetectChromiumBrowser_WindowsPrefersChromeInstallPath(t *testing.T) {
	programFiles := filepath.Join("C:", "Program Files")
	chromePath := filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe")
	edgePath := filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe")
	d := fakeDetector("windows", nil, map[string]string{
		"ProgramFiles": programFiles,
	}, map[string]bool{
		chromePath: true,
		edgePath:   true,
	})

	b, err := detectChromiumBrowserWith(d)

	require.NoError(t, err)
	assert.Equal(t, "Chrome", b.Name)
	assert.Equal(t, chromePath, b.Path)
}

func TestDetectChromiumBrowser_WindowsPathFallback(t *testing.T) {
	d := fakeDetector("windows", map[string]string{
		"msedge.exe": filepath.Join("C:", "Windows", "msedge.exe"),
	}, nil, nil)

	b, err := detectChromiumBrowserWith(d)

	require.NoError(t, err)
	assert.Equal(t, "Edge", b.Name)
}

func TestDetectChromiumBrowser_NotFound(t *testing.T) {
	d := fakeDetector("linux", nil, nil, nil)

	b, err := detectChromiumBrowserWith(d)

	assert.Nil(t, b)
	assert.ErrorIs(t, err, ErrUnsupportedBrowser)
}

func fakeDetector(goos string, paths map[string]string, env map[string]string, files map[string]bool) browserDetector {
	return browserDetector{
		goos: goos,
		lookPath: func(file string) (string, error) {
			if path, ok := paths[file]; ok {
				return path, nil
			}
			return "", os.ErrNotExist
		},
		getenv: func(key string) string {
			return env[key]
		},
		stat: func(name string) (os.FileInfo, error) {
			if files[name] {
				return fakeFileInfo{}, nil
			}
			return nil, os.ErrNotExist
		},
	}
}

type fakeFileInfo struct{}

func (fakeFileInfo) Name() string       { return "browser" }
func (fakeFileInfo) Size() int64        { return 1 }
func (fakeFileInfo) Mode() os.FileMode  { return 0o755 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() any           { return nil }
