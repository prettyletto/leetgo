package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type CSharpTemplate struct{}

const csharpStubTmpl = `public class Solution {
    public int[] {{.FuncName}}(int[] nums, int target) {
        // TODO: implement
        return new int[] {};
    }
}
`

const csharpTestTmpl = `using Xunit;

public class SolutionTests {
    [Fact]
    public void {{.FuncName}}_PassesExamples() {
        // TODO: add test cases from LeetCode examples
    }
}
`

func (c *CSharpTemplate) StubExt() string { return ".cs" }
func (c *CSharpTemplate) TestExt() string { return "Tests.cs" }

func (c *CSharpTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", csharpStubTmpl, csharpData(p))
}

func (c *CSharpTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", csharpTestTmpl, csharpData(p))
}

func csharpData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toPascalCase(p.Slug),
	}
}
