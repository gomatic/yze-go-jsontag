package e

import "testing"

// TestRoundTrip encodes the production decode-only mirror and decodes a
// test-local target: neither call may affect the production exemption sets.
func TestRoundTrip(t *testing.T) {
	if _, err := Marshal(wire{}); err != nil {
		t.Fatal(err)
	}
	var d testDecoded
	if err := Unmarshal(nil, &d); err != nil {
		t.Fatal(err)
	}
}
