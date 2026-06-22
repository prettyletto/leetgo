package gitexport

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashEmail_NormalizesEmail(t *testing.T) {
	assert.Equal(t, HashEmail("USER@example.com"), HashEmail(" user@example.com "))
}

func TestResolveIdentity_FallbackPersists(t *testing.T) {
	isolateGitConfig(t)
	dataDir := t.TempDir()
	repoDir := t.TempDir()

	first, err := ResolveIdentity(repoDir, dataDir)
	require.NoError(t, err)
	second, err := ResolveIdentity(repoDir, dataDir)
	require.NoError(t, err)

	assert.Equal(t, "local-fallback", first.Source)
	assert.Equal(t, first.ID, second.ID)
}

func TestExport_WritesSnapshotAndLatest(t *testing.T) {
	isolateGitConfig(t)
	ctx := context.Background()
	repoDir := t.TempDir()
	dataDir := t.TempDir()
	db, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))

	path, identity, err := Export(ctx, repoDir, dataDir, db)
	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.FileExists(t, filepath.Join(repoDir, "leetgo", identity.ID, "latest.json"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), identity.ID)
	assert.Contains(t, string(data), "local-fallback")
}

func TestExportWithOptions_CommitsExportFiles(t *testing.T) {
	isolateGitConfig(t)
	ctx := context.Background()
	repoDir := initTestRepo(t)
	dataDir := t.TempDir()
	db, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "leetgo.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))

	result, err := ExportWithOptions(ctx, repoDir, dataDir, db, ExportOptions{Commit: true})
	require.NoError(t, err)
	assert.True(t, result.Committed)
	assert.NotEmpty(t, result.CommitHash)
	assert.False(t, result.NoChanges)

	tracked := gitOutputForTest(t, repoDir, "ls-files", "--", filepath.Join("leetgo", result.Identity.ID))
	assert.Contains(t, tracked, "latest.json")
}

func TestCommit_NoChanges(t *testing.T) {
	isolateGitConfig(t)
	repoDir := initTestRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "leetgo", "id"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "leetgo", "id", "latest.json"), []byte("{}\n"), 0o644))
	_, noChanges, err := Commit(repoDir, filepath.Join("leetgo", "id"))
	require.NoError(t, err)
	assert.False(t, noChanges)

	_, noChanges, err = Commit(repoDir, filepath.Join("leetgo", "id"))
	require.NoError(t, err)
	assert.True(t, noChanges)
}

func isolateGitConfig(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "missing-gitconfig"))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	repoDir := t.TempDir()
	runGitForTest(t, repoDir, "init")
	runGitForTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForTest(t, repoDir, "config", "user.name", "Leetgo Test")
	return repoDir
}

func runGitForTest(t *testing.T, repoDir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func gitOutputForTest(t *testing.T, repoDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	require.NoError(t, err)
	return string(out)
}
