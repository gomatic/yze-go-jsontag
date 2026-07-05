// Package d exercises decode-only mirror detection through type aliases and
// generic instantiations: both must land the mirrored structs in the exempt
// set exactly as their unaliased, uninstantiated declarations.
package d

// Unmarshal stands in for encoding/json.Unmarshal (matched by name+shape).
func Unmarshal(data []byte, v any) error { _, _ = data, v; return nil }

// wire is decoded only through the alias below: still exempt.
type wire struct {
	FromLinter string `json:"FromLinter"`
}

// alias names wire; decoding through it must exempt wire itself.
type alias = wire

// payload is reached only as a generic type argument: still exempt.
type payload struct {
	RawName string `json:"RawName"`
}

// box is a generic decode container: decoding box[payload] must exempt the
// generic declaration itself, not just the instantiation.
type box[T any] struct {
	Items []T `json:"Items"`
}

// ownDoc is never serialized: its bad key is flagged, proving the analyzer
// still runs in this package.
type ownDoc struct {
	BadKey string `json:"BadKey"` // want `struct tag key "BadKey" is not snake_case`
}

func use() error {
	var w alias
	if err := Unmarshal(nil, &w); err != nil {
		return err
	}
	var b box[payload]
	_ = ownDoc{}
	return Unmarshal(nil, &b)
}
