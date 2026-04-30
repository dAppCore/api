// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"math"
	"testing"
)

// TestBridge_schemaInt_overflow_Bad verifies that uint/uint64 values exceeding
// math.MaxInt return (0, false) instead of silently wrapping to a negative int.
//
// G115 (gosec): integer overflow on coercion would let attacker-controlled
// JSON numbers >= 2^63 wrap to negative values, which downstream feeds into
// range checks / slice indices / array sizes with wrong sign.
func TestBridge_schemaInt_overflow_Bad(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "uint64 max", value: uint64(math.MaxUint64)},
		{name: "uint64 over MaxInt", value: uint64(math.MaxInt) + 1},
		{name: "uint over MaxInt", value: uint(math.MaxInt) + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := schemaInt(tt.value)
			if ok {
				t.Errorf("schemaInt(%v) returned ok=true; expected false on overflow", tt.value)
			}
			if got != 0 {
				t.Errorf("schemaInt(%v) returned %d; expected 0 on overflow", tt.value, got)
			}
		})
	}
}

// TestBridge_schemaInt_inrange_Good verifies that valid values still convert.
func TestBridge_schemaInt_inrange_Good(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{name: "uint zero", value: uint(0), want: 0},
		{name: "uint small", value: uint(42), want: 42},
		{name: "uint64 small", value: uint64(100), want: 100},
		{name: "uint64 maxint", value: uint64(math.MaxInt), want: math.MaxInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := schemaInt(tt.value)
			if !ok {
				t.Errorf("schemaInt(%v) returned ok=false; expected true", tt.value)
			}
			if got != tt.want {
				t.Errorf("schemaInt(%v) = %d; want %d", tt.value, got, tt.want)
			}
		})
	}
}

// TestBridge_schemaInt_boundary_Ugly tests the exact MaxInt boundary —
// MaxInt itself must succeed, MaxInt+1 must fail.
func TestBridge_schemaInt_boundary_Ugly(t *testing.T) {
	// uint64(MaxInt) — boundary, must succeed
	if got, ok := schemaInt(uint64(math.MaxInt)); !ok || got != math.MaxInt {
		t.Errorf("schemaInt(uint64(MaxInt)) = (%d, %v); want (MaxInt, true)", got, ok)
	}
	// uint64(MaxInt)+1 — one over boundary, must fail
	if _, ok := schemaInt(uint64(math.MaxInt) + 1); ok {
		t.Error("schemaInt(uint64(MaxInt)+1) returned ok=true; expected false (boundary)")
	}
}
