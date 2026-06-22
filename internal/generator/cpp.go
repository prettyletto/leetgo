package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type CppTemplate struct{}

const cppStubTmpl = `#include <vector>
using namespace std;

class Solution {
public:
    vector<int> {{.FuncName}}(vector<int>& nums, int target) {
        // TODO: implement
        return {};
    }
};
`

const cppTestTmpl = `#include "solution.cpp"

int main() {
    // TODO: add test cases from LeetCode examples
    return 0;
}
`

func (c *CppTemplate) StubExt() string { return ".cpp" }
func (c *CppTemplate) TestExt() string { return "_test.cpp" }

func (c *CppTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", cppStubTmpl, cppData(p))
}

func (c *CppTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", cppTestTmpl, cppData(p))
}

func cppData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toCamelCase(p.Slug),
	}
}
