// Package jsontag provides a go/analysis analyzer that reports struct field
// json/yaml tag keys that are not snake_case, per the gomatic data-format
// standard that serialized keys are snake_case.
package jsontag

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const message = "struct tag key %q is not snake_case"

// Analyzer reports struct field json/yaml tag keys that are not snake_case.
var Analyzer = &analysis.Analyzer{
	Name:     "jsontag",
	Doc:      "reports struct field json/yaml tag keys that are not snake_case, per the gomatic data-format standard",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "jsontag",
	Categories: []goyze.Category{"data"},
	URL:        "https://docs.gomatic.dev/yze/jsontag",
	Analyzer:   Analyzer,
}

// run reports every non-snake_case json/yaml tag key in the analyzed package.
func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.StructType)(nil)}, func(n ast.Node) {
		for _, field := range n.(*ast.StructType).Fields.List {
			checkField(pass, field)
		}
	})
	return nil, nil
}

// checkField inspects one struct field's tag for json and yaml keys.
func checkField(pass *analysis.Pass, field *ast.Field) {
	if field.Tag == nil {
		return
	}
	tag, _ := strconv.Unquote(field.Tag.Value)
	st := reflect.StructTag(tag)
	for _, key := range []string{"json", "yaml"} {
		checkTagKey(pass, field.Tag.Pos(), st, key)
	}
}

// checkTagKey reports the named key of st when its name is not snake_case.
func checkTagKey(pass *analysis.Pass, pos token.Pos, st reflect.StructTag, key string) {
	value, ok := st.Lookup(key)
	if !ok {
		return
	}
	name := strings.SplitN(value, ",", 2)[0]
	if name == "" || name == "-" {
		return
	}
	if !isSnakeCase(name) {
		pass.Reportf(pos, message, name)
	}
}

// isSnakeCase reports whether name contains no uppercase ASCII letter.
func isSnakeCase(name string) bool {
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			return false
		}
	}
	return true
}
