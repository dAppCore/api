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

func TestModelResolver_Knows_CaseInsensitive_Good(t *testing.T) {
	r := NewModelResolver()
	// Cache stores the lowercased name; Knows must mirror ResolveModel's
	// normalisation so a mixed-case request still hits the known model.
	r.loadedByName["gpt-4"] = nil
	if !r.Knows("GPT-4") {
		t.Fatal("Knows(GPT-4) = false, want true (case-insensitive cache hit)")
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
