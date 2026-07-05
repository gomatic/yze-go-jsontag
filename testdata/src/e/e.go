// Package e exercises test-file neutrality: _test.go usage must neither
// revoke a production decode-only exemption (a round-trip test Marshal) nor
// grant one (a decode that happens only in a test).
package e

// Unmarshal stands in for encoding/json.Unmarshal (matched by name+shape).
func Unmarshal(data []byte, v any) error { _, _ = data, v; return nil }

// Marshal stands in for encoding/json.Marshal.
func Marshal(v any) ([]byte, error) { _ = v; return nil, nil }

// wire is decode-only in production; the round-trip Marshal in e_test.go must
// not cancel the exemption.
type wire struct {
	FromLinter string `json:"FromLinter"`
}

// testDecoded is decoded only in e_test.go, which grants nothing: flagged.
type testDecoded struct {
	BadKey string `json:"BadKey"` // want `struct tag key "BadKey" is not snake_case`
}

func use() error {
	var w wire
	return Unmarshal(nil, &w)
}
