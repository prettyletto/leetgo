package gitexport

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Identity struct {
	ID     string
	Source string
}

func ResolveIdentity(repoDir, dataDir string) (*Identity, error) {
	email, err := gitEmail(repoDir)
	if err == nil && email != "" {
		return &Identity{ID: HashEmail(email), Source: "git-email"}, nil
	}

	id, err := fallbackID(dataDir)
	if err != nil {
		return nil, err
	}
	return &Identity{ID: id, Source: "local-fallback"}, nil
}

func HashEmail(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func gitEmail(repoDir string) (string, error) {
	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fallbackID(dataDir string) (string, error) {
	path := filepath.Join(dataDir, "export_identity")
	data, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read fallback export identity: %w", err)
	}

	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate fallback export identity: %w", err)
	}
	id := hex.EncodeToString(random)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(id+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write fallback export identity: %w", err)
	}
	return id, nil
}
