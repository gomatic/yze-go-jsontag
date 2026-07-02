// Package c exercises the decode-only mirror exemption: struct graphs that
// only flow into Unmarshal/Decode reproduce an external producer's keys.
package c

// Unmarshal stands in for encoding/json.Unmarshal (matched by name+shape).
func Unmarshal(data []byte, v any) error { _, _ = data, v; return nil }

// Marshal stands in for encoding/json.Marshal.
func Marshal(v any) ([]byte, error) { _ = v; return nil, nil }

// wireReport is decoded only: its PascalCase keys mirror the producer.
type wireReport struct {
	Issues []wireIssue `json:"Issues"`
}

// wireIssue is reached through wireReport's graph: also exempt.
type wireIssue struct {
	FromLinter string `json:"FromLinter"`
}

// ownDoc is decoded AND encoded: our own document, keys are ours — flagged.
type ownDoc struct {
	BadKey string `json:"BadKey"` // want `struct tag key "BadKey" is not snake_case`
}

// plain is never serialized through calls but declares a bad key: flagged.
type plain struct {
	AlsoBad string `json:"AlsoBad"` // want `struct tag key "AlsoBad" is not snake_case`
}

func use() error {
	var w wireReport
	if err := Unmarshal(nil, &w); err != nil {
		return err
	}
	var d ownDoc
	if err := Unmarshal(nil, &d); err != nil {
		return err
	}
	_, err := Marshal(d)
	_ = plain{}
	return err
}

// decoder exercises the Decode(target) shape and map/array containers.
type decoder struct{}

// Decode stands in for (*json.Decoder).Decode.
func (decoder) Decode(v any) error { _ = v; return nil }

// wireEntry arrives through a map value inside an array: still exempt.
type wireEntry struct {
	RawName string `json:"RawName"`
}

// wireBatch is the Decode target graph.
type wireBatch struct {
	Groups [2]map[string]wireEntry `json:"Groups"`
}

func useDecoder() error {
	var b wireBatch
	if err := (decoder{}).Decode(&b); err != nil {
		return err
	}
	// Unmarshal with the wrong arity establishes nothing.
	return UnmarshalInto(nil)
}

// UnmarshalInto has the Unmarshal NAME-mismatch shape: one arg, so decodeArg
// yields nothing for it even though the callee parses.
func UnmarshalInto(v any) error { _ = v; return nil }

// oneArgUnmarshal exercises the Unmarshal name with wrong arity.
func oneArgUnmarshal() error { return unmarshalVariadic() }

func unmarshalVariadic(extra ...any) error { _ = extra; return nil }

// wrongArity drives the Unmarshal/Decode arity guards: a one-arg Unmarshal
// selector and a two-arg Decode establish nothing.
type oddCodec struct{}

func (oddCodec) Unmarshal(v any) error { _ = v; return nil }

func (oddCodec) Decode(a, b any) error { _, _ = a, b; return nil }

func useOdd() error {
	if err := (oddCodec{}).Unmarshal(nil); err != nil {
		return err
	}
	return (oddCodec{}).Decode(nil, nil)
}
