// Package jsontag provides a go/analysis analyzer that reports struct field
// json/yaml tag keys that are not snake_case, per the gomatic data-format
// standard that serialized keys are snake_case. A key must match
// ^[a-z0-9]+(_[a-z0-9]+)*$ — lowercase ASCII words separated by single
// underscores. The empty key and the "-" (skip) key are never checked, and
// tag options after the first comma (",omitempty", ",string") are ignored.
//
// Keys beginning with "@" are sanctioned: they are JSON-LD keywords
// (@context, @id, @type, @value, @language, @graph, …) mandated by external
// specifications (W3C Verifiable Credentials and other JSON-LD vocabularies).
// Like the `$schema` marker below, the "@" prefix itself marks externally
// specified wire syntax rather than a naming-style choice, and external specs
// own that namespace entirely — so the remainder of an @-prefixed key is
// never snake_case-checked.
//
// Two exemptions cover struct types whose keys mirror a document defined by
// an external producer and so are not the module's to choose:
//
//   - Schema mirrors: a file declaring a struct field whose json key is
//     `$schema` — JSON Schema's self-description marker — is exempt as a
//     whole; it reproduces an externally defined document (SARIF, a
//     JSON-Schema catalog).
//
//   - Decode-only mirrors: a struct type is exempt when it is reachable —
//     through pointers, slices, arrays, map values, named types, type
//     aliases, generic instantiations (both the generic declaration and its
//     type arguments), and struct fields — from a decode call target (the
//     second argument of an Unmarshal(data, target), the sole argument of a
//     Decode(target); matched by callee name and arity) and never reachable
//     from an encode argument (Marshal, MarshalIndent, Encode). Such a graph
//     is only ever populated from a foreign producer's document. _test.go
//     files are ignored when computing this exemption: test usage neither
//     grants it (a decode only in a test exempts nothing) nor revokes it (a
//     round-trip test's Marshal does not cancel a production decode-only
//     mirror).
package jsontag

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"regexp"
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
// skipping files that mirror an external self-describing schema and struct
// graphs that are only ever DECODED (their keys belong to the producer).
func run(pass *analysis.Pass) (any, error) {
	mirrors := decodeOnlyTypes(pass)
	for _, file := range pass.Files {
		if !mirrorsExternalSchema(file) {
			checkFile(pass, file, mirrors)
		}
	}
	return nil, nil
}

// checkFile reports every non-snake_case tag key in one file's struct types,
// skipping decode-only mirror structs.
func checkFile(pass *analysis.Pass, file *ast.File, mirrors []types.Type) {
	ast.Inspect(file, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		if tv, found := pass.TypesInfo.Types[st]; found && containsType(mirrors, tv.Type) {
			return true
		}
		for _, field := range st.Fields.List {
			checkField(pass, field)
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
	// JSON-LD keyword namespace — externally mandated wire syntax, never style-checked.
	if strings.HasPrefix(name, "@") {
		return
	}
	if !isSnakeCase(serializedKey(name)) {
		pass.Reportf(pos, message, name)
	}
}

// serializedKey is the key a field serializes to: the first comma-separated element of its json/yaml tag value.
type serializedKey string

// snakeCase matches lowercase ASCII words separated by single underscores.
var snakeCase = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)

// isSnakeCase reports whether name is snake_case: ^[a-z0-9]+(_[a-z0-9]+)*$.
func isSnakeCase(name serializedKey) bool {
	return snakeCase.MatchString(string(name))
}
