package gitexport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/prettyletto/leetgo/internal/store"
)

type ExportOptions struct {
	Commit bool
}

type Result struct {
	Path       string
	ExportDir  string
	Identity   *Identity
	Committed  bool
	CommitHash string
	NoChanges  bool
}

func Export(ctx context.Context, repoDir, dataDir string, s *store.SQLiteStore) (string, *Identity, error) {
	result, err := ExportWithOptions(ctx, repoDir, dataDir, s, ExportOptions{})
	if err != nil {
		return "", nil, err
	}
	return result.Path, result.Identity, nil
}

func ExportWithOptions(ctx context.Context, repoDir, dataDir string, s *store.SQLiteStore, options ExportOptions) (*Result, error) {
	identity, err := ResolveIdentity(repoDir, dataDir)
	if err != nil {
		return nil, err
	}

	data, err := s.Export(ctx)
	if err != nil {
		return nil, err
	}
	data.ExportIdentity = identity.ID
	data.ExportIdentitySource = identity.Source

	dir := filepath.Join(repoDir, "leetgo", identity.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create git export dir: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("export-%s.json", time.Now().Format("2006-01-02-150405")))
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal git export: %w", err)
	}
	if err := os.WriteFile(path, jsonData, 0o644); err != nil {
		return nil, fmt.Errorf("write git export: %w", err)
	}

	latest := filepath.Join(dir, "latest.json")
	if err := os.WriteFile(latest, jsonData, 0o644); err != nil {
		return nil, fmt.Errorf("write latest git export: %w", err)
	}

	result := &Result{Path: path, ExportDir: dir, Identity: identity}
	if options.Commit {
		commitHash, noChanges, err := Commit(repoDir, filepath.Join("leetgo", identity.ID))
		if err != nil {
			return nil, err
		}
		result.Committed = commitHash != ""
		result.CommitHash = commitHash
		result.NoChanges = noChanges
	}

	return result, nil
}

func Commit(repoDir, exportRelDir string) (commitHash string, noChanges bool, err error) {
	if err := runGit(repoDir, "add", "--", exportRelDir); err != nil {
		return "", false, fmt.Errorf("git add export files: %w", err)
	}

	staged, err := gitOutput(repoDir, "diff", "--cached", "--name-only", "--", exportRelDir)
	if err != nil {
		return "", false, fmt.Errorf("check staged export changes: %w", err)
	}
	if strings.TrimSpace(staged) == "" {
		return "", true, nil
	}

	if err := runGit(repoDir, "commit", "-m", "Export leetgo progress", "--", exportRelDir); err != nil {
		return "", false, fmt.Errorf("git commit export files: %w", err)
	}

	hash, err := gitOutput(repoDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", false, fmt.Errorf("read export commit hash: %w", err)
	}
	return strings.TrimSpace(hash), false, nil
}

func runGit(repoDir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func gitOutput(repoDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
