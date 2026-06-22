package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type JavaTemplate struct{}

const javaStubTmpl = `class Solution {
    public {{.Return}} {{.FuncName}}({{.Params}}) {
        // TODO: implement
    }
}
`

const javaTestTmpl = `import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class SolutionTest {
    @Test
    void test{{.FuncName}}() {
        // TODO: add test cases from LeetCode examples
    }
}
`

func (j *JavaTemplate) StubExt() string { return ".java" }
func (j *JavaTemplate) TestExt() string { return "Test.java" }

func (j *JavaTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", javaStubTmpl, javaData(p))
}

func (j *JavaTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", javaTestTmpl, javaData(p))
}

func javaData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toCamelCase(p.Slug),
		"Params":   "int[] nums, int target",
		"Return":   "int[]",
	}
}
