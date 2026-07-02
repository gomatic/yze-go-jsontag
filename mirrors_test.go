package jsontag

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// TestCalleeNameLeavesOddCallShapesAnonymous covers call shapes that are
// neither an identifier nor a selector — an immediately invoked literal.
func TestCalleeNameLeavesOddCallShapesAnonymous(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.FuncLit{Type: &ast.FuncType{}}}
	assert.Empty(t, calleeName(call))
}

// TestAppendRootSkipsUntypedArguments covers the guard for an argument the
// type info has no entry for (a synthetic AST outside any checked package).
func TestAppendRootSkipsUntypedArguments(t *testing.T) {
	pass := &analysis.Pass{TypesInfo: &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}}
	call := &ast.CallExpr{Args: []ast.Expr{&ast.Ident{NamePos: token.NoPos, Name: "x"}}}
	assert.Empty(t, appendRoot(pass, nil, call, 0))
}
