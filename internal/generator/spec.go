package generator

// ValueKind describes the type of a parameter or return value.
type ValueKind string

const (
	KindInt              ValueKind = "int"
	KindIntSlice         ValueKind = "[]int"
	KindIntSliceSlice    ValueKind = "[][]int"
	KindBool             ValueKind = "bool"
	KindString           ValueKind = "string"
	KindStringSlice      ValueKind = "[]string"
	KindStringSliceSlice ValueKind = "[][]string"
	KindByteSlice        ValueKind = "[]byte"
	KindByteSliceSlice   ValueKind = "[][]byte"
	KindFloat64          ValueKind = "float64"
	KindListNode         ValueKind = "*ListNode"
	KindTreeNode         ValueKind = "*TreeNode"
	KindVoid             ValueKind = "void"
)

// ComparisonKind describes how test results should be compared.
type ComparisonKind string

const (
	CmpExact     ComparisonKind = "exact"
	CmpDeep      ComparisonKind = "deep"
	CmpUnordered ComparisonKind = "unordered"
	CmpBool      ComparisonKind = "bool"
	CmpSkip      ComparisonKind = "skip"
)

// ParamSpec describes a single function parameter.
type ParamSpec struct {
	Name string
	Type ValueKind
}

// ReturnSpec describes the function return type.
type ReturnSpec struct {
	Type ValueKind
}

// ExampleSpec holds one test case with rendered input values and expected output.
type ExampleSpec struct {
	Input  map[string]string // param name -> rendered value string
	Expect string            // rendered expected value
}

// ProblemSpec defines the full generation metadata for one problem.
type ProblemSpec struct {
	Slug       string
	FuncName   string // language-appropriate name set later
	Params     []ParamSpec
	Return     ReturnSpec
	Comparison ComparisonKind
	Examples   []ExampleSpec
	IsDesign   bool
	DesignNote string

	// Data structures needed by this problem
	NeedsListNode bool
	NeedsTreeNode bool
}

// GoFuncName returns the Go function name (PascalCase).
func (s *ProblemSpec) GoFuncName() string {
	return toPascalCase(s.Slug)
}

// PythonFuncName returns the Python function name (snake_case).
func (s *ProblemSpec) PythonFuncName() string {
	return toSnakeCase(s.Slug)
}

// TSFuncName returns the TypeScript/JavaScript function name (camelCase).
func (s *ProblemSpec) TSFuncName() string {
	return toCamelCase(s.Slug)
}

// JavaFuncName returns the Java method name (camelCase).
func (s *ProblemSpec) JavaFuncName() string {
	return toCamelCase(s.Slug)
}

// CppFuncName returns the C++ method name (camelCase).
func (s *ProblemSpec) CppFuncName() string {
	return toCamelCase(s.Slug)
}

// RustFuncName returns the Rust function name (snake_case).
func (s *ProblemSpec) RustFuncName() string {
	return toSnakeCase(s.Slug)
}

// CSharpFuncName returns the C# method name (PascalCase).
func (s *ProblemSpec) CSharpFuncName() string {
	return toPascalCase(s.Slug)
}

// FileBase returns the slug-derived file base name.
func (s *ProblemSpec) FileBase() string {
	return toSnakeCase(s.Slug)
}
