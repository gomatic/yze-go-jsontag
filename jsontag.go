// Package jsontag provides a go/analysis analyzer that reports struct field
// json/yaml tag keys that are not snake_case, per the gomatic data-format
// standard that serialized keys are snake_case. A file that mirrors an
// external, self-describing JSON-Schema document — detected by a struct field
// tagged `$schema`, JSON Schema's self-description marker — is exempt: its
// keys reproduce the external standard (SARIF, JSON-Schema catalogs) and are
// not the module's to choose.
package jsontag

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

const message = "struct tag key %q is not snake_case"

// Analyzer reports struct field json/yaml tag keys that are not snake_case.
var Analyzer = &analysis.Analyzer{
	Name: "jsontag",
	Doc:  "reports struct field json/yaml tag keys that are not snake_case, per the gomatic data-format standard",
	Run:  run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "jsontag",
	Categories: []goyze.Category{"data"},
	URL:        "https://docs.gomatic.dev/yze/jsontag",
	Analyzer:   Analyzer,
}

// run reports every non-snake_case json/yaml tag key in the analyzed package,
// skipping files that mirror an external self-describing schema.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !mirrorsExternalSchema(file) {
			checkFile(pass, file)
		}
	}
	return nil, nil
}

// checkFile reports every non-snake_case tag key in one file's struct types.
func checkFile(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			for _, field := range st.Fields.List {
				checkField(pass, field)
			}
		}
		return true
	})
}

// mirrorsExternalSchema reports whether the file declares a struct field whose
// json key is `$schema` — JSON Schema's self-description marker. Such a file
// reproduces an externally defined document (SARIF, a schema catalog); its
// keys follow that standard, not the gomatic one.
func mirrorsExternalSchema(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return !found
		}
		for _, field := range st.Fields.List {
			if fieldKeyIsSchema(field) {
				found = true
			}
		}
		return !found
	})
	return found
}

// fieldKeyIsSchema reports whether a field's json tag key is `$schema`.
func fieldKeyIsSchema(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}
	tag, _ := strconv.Unquote(field.Tag.Value)
	value, ok := reflect.StructTag(tag).Lookup("json")
	return ok && strings.SplitN(value, ",", 2)[0] == "$schema"
}

// checkField inspects one struct field's tag for json and yaml keys.
func checkField(pass *analysis.Pass, field *ast.Field) {
	if field.Tag == nil {
		return
	}
	tag, _ := strconv.Unquote(field.Tag.Value)
	st := reflect.StructTag(tag)
	for _, key := range []string{"json", "yaml"} {
		checkTagKey(pass, field.Tag.Pos(), st, tagKey(key))
	}
}

// tagKey is a struct-tag key naming a serialization format ("json", "yaml").
type tagKey string

// checkTagKey reports the named key of st when its name is not snake_case.
func checkTagKey(pass *analysis.Pass, pos token.Pos, st reflect.StructTag, key tagKey) {
	value, ok := st.Lookup(string(key))
	if !ok {
		return
	}
	name := strings.SplitN(value, ",", 2)[0]
	if name == "" || name == "-" {
		return
	}
	if !isSnakeCase(serializedKey(name)) {
		pass.Reportf(pos, message, name)
	}
}

// serializedKey is the key a field serializes to: the first comma-separated element of its json/yaml tag value.
type serializedKey string

// isSnakeCase reports whether name contains no uppercase ASCII letter.
func isSnakeCase(name serializedKey) bool {
	for _, r := range string(name) {
		if r >= 'A' && r <= 'Z' {
			return false
		}
	}
	return true
}
