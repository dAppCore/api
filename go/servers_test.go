// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"
)

func TestServers_normaliseServer_Good_TrimsAndStripsTrailingSlash(t *testing.T) {
	got := normaliseServer(" https://api.example.com/ ")
	if got != "https://api.example.com" {
		t.Fatalf("expected trimmed server URL, got %q", got)
	}
}

func TestServers_normaliseServer_Bad_ReturnsEmptyForBlankInput(t *testing.T) {
	if got := normaliseServer("   "); got != "" {
		t.Fatalf("expected blank input to normalise to empty string, got %q", got)
	}
	if got := normaliseServer(""); got != "" {
		t.Fatalf("expected empty input to normalise to empty string, got %q", got)
	}
}

func TestServers_normaliseServer_Ugly_CollapsesRepeatedSlashesToRoot(t *testing.T) {
	if got := normaliseServer("///"); got != "/" {
		t.Fatalf("expected repeated slashes to collapse to root, got %q", got)
	}
}

func TestServers_normaliseServers_Good_DeduplicatesAndPreservesOrder(t *testing.T) {
	got := normaliseServers([]string{
		" https://api.example.com/ ",
		"https://api.example.com",
		" / ",
		"https://docs.example.com/",
	})

	want := []string{
		"https://api.example.com",
		"/",
		"https://docs.example.com",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d servers, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestServers_normaliseServers_Bad_ReturnsNilWhenOnlyBlankEntries(t *testing.T) {
	if got := normaliseServers([]string{"", "   ", "\t"}); got != nil {
		t.Fatalf("expected blank-only input to return nil, got %v", got)
	}
	if got := normaliseServers(nil); got != nil {
		t.Fatalf("expected nil input to return nil, got %v", got)
	}
}

func TestServers_normaliseServers_Ugly_CollapsesEquivalentForms(t *testing.T) {
	got := normaliseServers([]string{
		"///",
		" / ",
		"https://api.example.com//",
		"https://api.example.com",
	})

	want := []string{"/", "https://api.example.com"}

	if len(got) != len(want) {
		t.Fatalf("expected %d servers, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}
