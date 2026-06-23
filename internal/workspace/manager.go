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

	if err := m.writeScaffolding(dir, lang); err != nil {
		return "", "", fmt.Errorf("write scaffolding: %w", err)
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

func (m *Manager) writeScaffolding(dir string, lang generator.Language) error {
	switch lang {
	case generator.LangGo:
		return writeGoMod(dir)
	case generator.LangTypeScript, generator.LangJavaScript:
		return writeNodePackage(dir, lang)
	case generator.LangJava:
		return writeMavenPOM(dir)
	case generator.LangCSharp:
		return writeCSharpProject(dir)
	default:
		return nil
	}
}

func writeGoMod(dir string) error {
	modPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(modPath); err == nil {
		return nil // already exists
	}
	return os.WriteFile(modPath, []byte("module solution\n\ngo 1.21\n"), 0o644)
}

func writeNodePackage(dir string, lang generator.Language) error {
	pkgPath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return nil
	}

	content := `{
  "name": "leetgo-problem",
  "private": true,
  "scripts": {
    "test": "vitest run"
  },
  "devDependencies": {
    "vitest": "*"
  }
}
`
	return os.WriteFile(pkgPath, []byte(content), 0o644)
}

func writeMavenPOM(dir string) error {
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		return nil
	}

	content := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.leetgo</groupId>
    <artifactId>problem</artifactId>
    <version>1.0-SNAPSHOT</version>
    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>5.10.0</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>3.1.2</version>
            </plugin>
        </plugins>
    </build>
</project>
`
	return os.WriteFile(pomPath, []byte(content), 0o644)
}

func writeCSharpProject(dir string) error {
	projPath := filepath.Join(dir, "problem.csproj")
	if _, err := os.Stat(projPath); err == nil {
		return nil
	}

	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="xunit" Version="2.6.0" />
    <PackageReference Include="xunit.runner.visualstudio" Version="2.5.0" />
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.8.0" />
  </ItemGroup>
</Project>
`
	return os.WriteFile(projPath, []byte(content), 0o644)
}

func (m *Manager) ProblemDir(p *roadmap.Problem) string {
	return filepath.Join(m.root, string(p.Category), fmt.Sprintf("%d-%s", p.ID, p.Slug))
}

func toFileName(slug string) string {
	return strings.ReplaceAll(slug, "-", "_")
}
