package generator

import "github.com/prettyletto/leetgo/internal/roadmap"

type RustTemplate struct{}

const rustStubTmpl = `pub struct Solution;

impl Solution {
    pub fn {{.FuncName}}(nums: Vec<i32>, target: i32) -> Vec<i32> {
        // TODO: implement
        vec![]
    }
}
`

const rustTestTmpl = `use super::*;

#[test]
fn test_{{.FuncName}}() {
    // TODO: add test cases from LeetCode examples
}
`

func (r *RustTemplate) StubExt() string { return ".rs" }
func (r *RustTemplate) TestExt() string { return "_test.rs" }

func (r *RustTemplate) RenderStub(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("stub", rustStubTmpl, rustData(p))
}

func (r *RustTemplate) RenderTest(p *roadmap.Problem) ([]byte, error) {
	return renderTemplate("test", rustTestTmpl, rustData(p))
}

func rustData(p *roadmap.Problem) map[string]string {
	return map[string]string{
		"FuncName": toSnakeCase(p.Slug),
	}
}
