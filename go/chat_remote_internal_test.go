// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestModelResolver_Knows_Good(t *testing.T) {
	r := NewModelResolver()
	// Seed the loaded-by-name cache directly (internal test) to simulate a known model.
	r.loadedByName["lemer"] = nil
	if !r.Knows("lemer") {
		t.Fatal("Knows(lemer) = false, want true (cache hit)")
	}
}

func TestModelResolver_Knows_Bad(t *testing.T) {
	r := NewModelResolver()
	if r.Knows("does-not-exist") {
		t.Fatal("Knows(does-not-exist) = true, want false")
	}
	if r.Knows("") {
		t.Fatal("Knows(empty) = true, want false")
	}
	var nilR *ModelResolver
	if nilR.Knows("x") {
		t.Fatal("nil resolver Knows = true, want false")
	}
}
