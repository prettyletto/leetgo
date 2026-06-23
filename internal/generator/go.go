package generator

import (
	"fmt"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type GoTemplate struct{}

func (g *GoTemplate) StubExt() string { return ".go" }
func (g *GoTemplate) TestExt() string { return "_test.go" }

func (g *GoTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d (%s)", p.ID, p.Slug)
	}
	return renderGoStub(spec), nil
}

func (g *GoTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	spec, ok := SpecForProblem(p)
	if !ok {
		return nil, fmt.Errorf("no spec for problem %d (%s)", p.ID, p.Slug)
	}
	return renderGoTest(spec), nil
}

func SpecForProblem(p *roadmap.Problem) (*ProblemSpec, bool) {
	s, ok := problemSpecs[p.ID]
	if !ok {
		return nil, false
	}
	s.Slug = p.Slug
	return &s, true
}
