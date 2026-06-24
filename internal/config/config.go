package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const MaxDisplayNameLength = 40
const CurrentOnboardingVersion = 1

var ValidThemes = []string{
	"rpg-skill-tree",
	"clean-productivity",
	"cyber-dashboard",
}

var ValidSymbolModes = []string{
	"rich",
	"plain",
}

var ValidMotionPreferences = []string{
	"normal",
	"reduced",
	"off",
}

type Config struct {
	OnboardingComplete bool   `toml:"onboarding_complete"`
	OnboardingVersion  int    `toml:"onboarding_version"`
	DisplayName        string `toml:"display_name"`
	Workspace          string `toml:"workspace"`
	Editor             string `toml:"editor"`
	Language           string `toml:"language"`
	Roadmap            string `toml:"roadmap"`
	Theme              string `toml:"theme"`
	SymbolMode         string `toml:"symbol_mode"`
	MotionPreference   string `toml:"motion_preference"`
	GitExportEnabled   bool   `toml:"git_export_enabled"`
	GitExportRepo      string `toml:"git_export_repo"`
}

func DefaultConfig() (*Config, error) {
	workspace, err := DefaultWorkspace()
	if err != nil {
		return nil, err
	}
	return &Config{
		OnboardingComplete: false,
		OnboardingVersion:  0,
		DisplayName:        "",
		Workspace:          workspace,
		Editor:             "",
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
		SymbolMode:         "rich",
		MotionPreference:   "normal",
		GitExportEnabled:   false,
		GitExportRepo:      "",
	}, nil
}

func DefaultWorkspace() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, "leetgo-workspace"), nil
}

func (c *Config) ReadyForDashboard(languages, roadmaps []string) bool {
	if !c.OnboardingComplete || c.OnboardingVersion < CurrentOnboardingVersion {
		return false
	}
	return c.Validate(languages, roadmaps) == nil
}

func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".leetgo"), nil
}

func Load() (*Config, error) {
	dataDir, err := DataDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dataDir, "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig()
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg, err := DefaultConfig()
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.Theme == "" {
		c.Theme = "rpg-skill-tree"
	}
	if c.SymbolMode == "" {
		c.SymbolMode = "rich"
	}
	if c.MotionPreference == "" {
		c.MotionPreference = "normal"
	}
}

func (c *Config) Save() error {
	c.ApplyDefaults()

	dataDir, err := DataDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := filepath.Join(dataDir, "config.toml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) Validate(languages, roadmaps []string) error {
	c.ApplyDefaults()

	if c.Workspace == "" {
		return fmt.Errorf("workspace is required")
	}
	if !slices.Contains(languages, c.Language) {
		return fmt.Errorf("unsupported language %q", c.Language)
	}
	if !slices.Contains(roadmaps, c.Roadmap) {
		return fmt.Errorf("unknown roadmap %q", c.Roadmap)
	}
	if !slices.Contains(ValidThemes, c.Theme) {
		return fmt.Errorf("unknown theme %q", c.Theme)
	}
	if !slices.Contains(ValidSymbolModes, c.SymbolMode) {
		return fmt.Errorf("unknown symbol_mode %q", c.SymbolMode)
	}
	if !slices.Contains(ValidMotionPreferences, c.MotionPreference) {
		return fmt.Errorf("unknown motion_preference %q", c.MotionPreference)
	}
	if c.OnboardingComplete {
		trimmed := strings.TrimSpace(c.DisplayName)
		if trimmed == "" {
			return fmt.Errorf("display_name is required when onboarding is complete")
		}
		if len(trimmed) > MaxDisplayNameLength {
			return fmt.Errorf("display_name exceeds %d characters", MaxDisplayNameLength)
		}
	}
	if c.GitExportEnabled && strings.TrimSpace(c.GitExportRepo) == "" {
		return fmt.Errorf("git_export_repo is required when git_export_enabled is true")
	}
	if c.GitExportEnabled && strings.TrimSpace(c.GitExportRepo) != "" {
		info, err := os.Stat(c.GitExportRepo)
		if err != nil {
			return fmt.Errorf("git_export_repo is not accessible: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("git_export_repo is not a directory: %s", c.GitExportRepo)
		}
		if _, err := os.Stat(filepath.Join(c.GitExportRepo, ".git")); err != nil {
			return fmt.Errorf("git_export_repo is not a git repository: %s", c.GitExportRepo)
		}
	}
	return nil
}
