package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/roadmap"
)

type Manager struct {
	root      string
	generator *generator.Generator
}

func New(root string, gen *generator.Generator) *Manager {
	return &Manager{root: root, generator: gen}
}

func (m *Manager) Generate(p *roadmap.Problem, lang generator.Language) (stubPath, testPath string, err error) {
	tmpl, ok := m.generator.GetTemplate(lang)
	if !ok {
		return "", "", fmt.Errorf("unsupported language: %s", lang)
	}

	dir := filepath.Join(m.root, string(p.Category), fmt.Sprintf("%d-%s", p.ID, p.Slug))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create problem dir: %w", err)
	}

	stubName := toFileName(p.Slug) + tmpl.StubExt()
	testName := toFileName(p.Slug) + tmpl.TestExt()

	stubPath = filepath.Join(dir, stubName)
	testPath = filepath.Join(dir, testName)

	stub, err := tmpl.RenderStub(p)
	if err != nil {
		return "", "", fmt.Errorf("render stub: %w", err)
	}
	if err := os.WriteFile(stubPath, stub, 0o644); err != nil {
		return "", "", fmt.Errorf("write stub: %w", err)
	}

	test, err := tmpl.RenderTest(p)
	if err != nil {
		return "", "", fmt.Errorf("render test: %w", err)
	}
	if err := os.WriteFile(testPath, test, 0o644); err != nil {
		return "", "", fmt.Errorf("write test: %w", err)
	}

	return stubPath, testPath, nil
}

func (m *Manager) ProblemDir(p *roadmap.Problem) string {
	return filepath.Join(m.root, string(p.Category), fmt.Sprintf("%d-%s", p.ID, p.Slug))
}

func toFileName(slug string) string {
	return strings.ReplaceAll(slug, "-", "_")
}
