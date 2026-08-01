package jsontag

// White-box test for the file filter that builds the decode/encode sets. It
// decides which files contribute evidence, so including the wrong ones does not
// produce a wrong diagnostic — it produces silence, because a type that is only
// decoded in production looks symmetric once a test's encode call is counted.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

// TestIsTestFileKeepsTestEvidenceOutOfTheDecodeEncodeSets names isTestFile's
// claim. The analyzer compares what a package DECODES against what it ENCODES,
// and a test file routinely does both — round-tripping a fixture is the most
// ordinary thing a test does. Counting that evidence makes every decoded type
// look like it is also encoded, and the rule stops reporting anything at all:
// a silent analyzer, with a green gate.
func TestIsTestFileKeepsTestEvidenceOutOfTheDecodeEncodeSets(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		why  string
		want bool
	}{
		{name: "thing_test.go", want: true, why: "an external or in-package test file"},
		{name: "a/b/thing_test.go", want: true, why: "the directory is irrelevant"},
		{name: "thing.go", want: false, why: "ordinary source contributes evidence"},
		{name: "test.go", want: false, why: "a file merely named test is not a test file"},
		{name: "thing_testdata.go", want: false, why: "the suffix must be exactly _test.go"},
	} {
		pass, file := passForFile(t, tc.name)
		assert.Equal(t, tc.want, isTestFile(pass, file), "isTestFile(%q): %s", tc.name, tc.why)
	}
}

// passForFile builds the minimal pass/file pair isTestFile inspects, plus the
// file itself, since the predicate reads the name from the FileSet rather than
// from the AST.
func passForFile(t *testing.T, name string) (*analysis.Pass, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, "package p\n", parser.PackageClauseOnly)
	require.NoError(t, err)
	return &analysis.Pass{Fset: fset, Files: []*ast.File{file}}, file
}

// TestCalleeNameIgnoresACallThatNamesNoFunction covers calleeName's fallback.
// A call whose callee is neither an identifier nor a selector — an immediately
// invoked function literal, or the result of another call — names no function
// this analyzer can match, and returning a stray name for it would let an
// arbitrary expression be read as an Unmarshal or Decode call.
func TestCalleeNameIgnoresACallThatNamesNoFunction(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		expr string
		want string
		why  string
	}{
		{expr: unmarshalCall, want: methodUnmarshal, why: "a bare identifier"},
		{expr: "json." + unmarshalCall, want: methodUnmarshal, why: "a selector yields the method name"},
		{expr: "func() {}()", want: "", why: "an invoked function literal names nothing"},
		{expr: "f()()", want: "", why: "and neither does the result of another call"},
	} {
		assert.Equal(t, tc.want, calleeName(callExpr(t, tc.expr)), "calleeName(%s): %s", tc.expr, tc.why)
	}
}

// TestAppendRootIgnoresAnAbsentArgument covers appendRoot's two skips. A
// negative index means the call carried no decode/encode target at all, and an
// argument with no recorded type cannot contribute one — appending in either
// case would add a nil or wrong type to the root set and make the decode/encode
// comparison compare something that was never there.
func TestAppendRootIgnoresAnAbsentArgument(t *testing.T) {
	t.Parallel()
	roots := []types.Type{types.Typ[types.String]}
	pass := &analysis.Pass{TypesInfo: &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}}

	assert.Equal(t, roots, appendRoot(pass, roots, callExpr(t, unmarshalCall), -1),
		"a negative index means there was no target argument")
	assert.Equal(t, roots, appendRoot(pass, roots, callExpr(t, unmarshalCall), 1),
		"an argument with no recorded type contributes nothing")
}

// callExpr parses a single call expression.
func callExpr(t *testing.T, src string) *ast.CallExpr {
	t.Helper()
	expr, err := parser.ParseExpr(src)
	require.NoError(t, err)
	call, ok := expr.(*ast.CallExpr)
	require.True(t, ok, "%s is not a call expression", src)
	return call
}

// unmarshalCall is the decode call these tests exercise, named so the literal
// has one definition here rather than being repeated per assertion.
const unmarshalCall = "Unmarshal(a, b)"
