package generator

import (
	"github.com/prettyletto/leetgo/internal/roadmap"
)

type Language string

const (
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangTypeScript Language = "typescript"
	LangJava       Language = "java"
	LangCpp        Language = "cpp"
	LangJavaScript Language = "javascript"
	LangRust       Language = "rust"
	LangCSharp     Language = "csharp"
)

type Template interface {
	StubExt() string
	TestExt() string
	RenderStub(p *roadmap.Problem) ([]byte, error)
	RenderTest(p *roadmap.Problem) ([]byte, error)
}

type Generator struct {
	templates map[Language]Template
}

func New() *Generator {
	return &Generator{
		templates: map[Language]Template{
			LangGo:         &GoTemplate{},
			LangPython:     &PythonTemplate{},
			LangTypeScript: &TypeScriptTemplate{},
			LangJava:       &JavaTemplate{},
			LangCpp:        &CppTemplate{},
			LangJavaScript: &JavaScriptTemplate{},
			LangRust:       &RustTemplate{},
			LangCSharp:     &CSharpTemplate{},
		},
	}
}

func (g *Generator) GetTemplate(lang Language) (Template, bool) {
	t, ok := g.templates[lang]
	return t, ok
}

func (g *Generator) Languages() []Language {
	return []Language{LangGo, LangPython, LangTypeScript, LangJava, LangCpp, LangJavaScript, LangRust, LangCSharp}
}
