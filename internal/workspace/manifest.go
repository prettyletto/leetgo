package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const ManifestFileName = ".leetgo-problem.toml"

type Manifest struct {
	ProblemID     int    `toml:"problem_id"`
	Slug          string `toml:"slug"`
	Roadmap       string `toml:"roadmap"`
	Stage         string `toml:"stage"`
	Language      string `toml:"language"`
	StubPath      string `toml:"stub_path"`
	TestsuitePath string `toml:"testsuite_path"`
}

func ReadManifest(startDir string) (*Manifest, string, error) {
	dir := startDir
	for {
		manifestPath := filepath.Join(dir, ManifestFileName)
		data, err := os.ReadFile(manifestPath)
		if err == nil {
			var m Manifest
			if err := toml.Unmarshal(data, &m); err != nil {
				return nil, "", fmt.Errorf("parse manifest %s: %w", manifestPath, err)
			}
			return &m, dir, nil
		}
		if !os.IsNotExist(err) {
			return nil, "", fmt.Errorf("read manifest: %w", err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, "", nil
		}
		dir = parent
	}
}

func WriteManifest(dir string, m *Manifest) error {
	manifestPath := filepath.Join(dir, ManifestFileName)

	if err := EnsureManifestWritable(dir, m.ProblemID); err != nil {
		return err
	}

	data, err := toml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}

func EnsureManifestWritable(dir string, problemID int) error {
	manifestPath := filepath.Join(dir, ManifestFileName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read manifest: %w", err)
	}

	var existing Manifest
	if err := toml.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}
	if existing.ProblemID != problemID {
		return fmt.Errorf("manifest exists for a different Problem ID (%d); refusing to overwrite with Problem ID %d", existing.ProblemID, problemID)
	}
	return nil
}
