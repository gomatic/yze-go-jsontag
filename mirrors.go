// Decode-only mirror detection: struct graphs that only flow into
// Unmarshal/Decode reproduce an external producer's document.
package jsontag

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// decodeOnlyTypes collects the struct types reachable from decode call targets
// (Unmarshal's second argument, Decode's first) and not from encode arguments
// (Marshal/MarshalIndent/Encode) — a decode-only graph mirrors an external
// producer's document, whose keys are not the module's to choose.
func decodeOnlyTypes(pass *analysis.Pass) []types.Type {
	var decoded, encoded []types.Type
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if ok {
				decoded = appendRoot(pass, decoded, call, decodeArg(call))
				encoded = appendRoot(pass, encoded, call, encodeArg(call))
			}
			return true
		})
	}
	return minusTypes(expandAll(decoded), expandAll(encoded))
}

// decodeArg yields the argument index carrying a decode target, or -1: the
// second of an Unmarshal(data, target), the first of a Decode(target).
func decodeArg(call *ast.CallExpr) int {
	switch calleeName(call) {
	case "Unmarshal":
		if len(call.Args) == 2 {
			return 1
		}
	case "Decode":
		if len(call.Args) == 1 {
			return 0
		}
	}
	return -1
}

// encodeArg yields the argument index carrying an encode source, or -1.
func encodeArg(call *ast.CallExpr) int {
	switch calleeName(call) {
	case "Marshal", "MarshalIndent", "Encode":
		if len(call.Args) >= 1 {
			return 0
		}
	}
	return -1
}

// calleeName is the called function or method's bare name.
func calleeName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	default:
		return ""
	}
}

// appendRoot appends the type of call's idx-th argument.
func appendRoot(pass *analysis.Pass, roots []types.Type, call *ast.CallExpr, idx int) []types.Type {
	if idx < 0 {
		return roots
	}
	if tv, ok := pass.TypesInfo.Types[call.Args[idx]]; ok && tv.Type != nil {
		return append(roots, tv.Type)
	}
	return roots
}

// expandAll expands every root through pointers, containers, named types, and
// struct fields into the full reachable struct set.
func expandAll(roots []types.Type) []types.Type {
	var out []types.Type
	seen := map[types.Type]bool{}
	for _, r := range roots {
		out = expand(r, seen, out)
	}
	return out
}

// expand walks one type, accumulating every reachable struct type.
func expand(t types.Type, seen map[types.Type]bool, out []types.Type) []types.Type {
	if t == nil || seen[t] {
		return out
	}
	seen[t] = true
	if elem := elementOf(t); elem != nil {
		return expand(elem, seen, out)
	}
	if st, ok := t.(*types.Struct); ok {
		return expandStruct(st, seen, out)
	}
	return out
}

// elementOf unwraps one container/pointer/named level, or nil at a leaf.
func elementOf(t types.Type) types.Type {
	switch u := t.(type) {
	case *types.Pointer:
		return u.Elem()
	case *types.Slice:
		return u.Elem()
	case *types.Array:
		return u.Elem()
	case *types.Map:
		return u.Elem()
	case *types.Named:
		return u.Underlying()
	default:
		return nil
	}
}

// expandStruct records a struct and walks its fields.
func expandStruct(st *types.Struct, seen map[types.Type]bool, out []types.Type) []types.Type {
	out = append(out, st)
	for f := range st.Fields() {
		out = expand(f.Type(), seen, out)
	}
	return out
}

// minusTypes returns the members of a absent from b (by type identity).
func minusTypes(a, b []types.Type) []types.Type {
	var out []types.Type
	for _, t := range a {
		if !containsType(b, t) {
			out = append(out, t)
		}
	}
	return out
}

// containsType reports membership by types.Identical.
func containsType(set []types.Type, t types.Type) bool {
	for _, s := range set {
		if types.Identical(s, t) {
			return true
		}
	}
	return false
}
