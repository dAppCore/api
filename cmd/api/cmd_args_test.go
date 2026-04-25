// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"
	"testing"
)

// TestCmdArgs_SplitUniqueCSV_Good_TrimsAndDeduplicates verifies comma-
// separated values are trimmed, deduplicated, and returned in first-seen
// order.
func TestCmdArgs_SplitUniqueCSV_Good_TrimsAndDeduplicates(t *testing.T) {
	got := splitUniqueCSV(" go, python ,go, ruby ")
	want := []string{"go", "python", "ruby"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// TestCmdArgs_SplitUniqueCSV_Bad_EmptyInputReturnsNil verifies empty input
// produces no values.
func TestCmdArgs_SplitUniqueCSV_Bad_EmptyInputReturnsNil(t *testing.T) {
	if got := splitUniqueCSV(""); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// TestCmdArgs_SplitUniqueCSV_Ugly_IgnoresBlankSegments verifies separator
// noise and repeated whitespace do not leak empty entries into the result.
func TestCmdArgs_SplitUniqueCSV_Ugly_IgnoresBlankSegments(t *testing.T) {
	got := splitUniqueCSV(" , , go, , python,go , ")
	want := []string{"go", "python"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// TestCmdArgs_NormalisePublicPaths_Good_AddsLeadingSlashAndDeduplicates
// verifies relative paths are promoted to absolute route paths and repeated
// entries are skipped.
func TestCmdArgs_NormalisePublicPaths_Good_AddsLeadingSlashAndDeduplicates(t *testing.T) {
	got := normalisePublicPaths([]string{"docs", "/docs", " api "})
	want := []string{"/docs", "/api"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// TestCmdArgs_NormalisePublicPaths_Bad_EmptyInputReturnsNil verifies that
// empty path lists stay empty.
func TestCmdArgs_NormalisePublicPaths_Bad_EmptyInputReturnsNil(t *testing.T) {
	if got := normalisePublicPaths(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// TestCmdArgs_NormalisePublicPaths_Ugly_BlankEntriesReturnNil verifies that
// a slice containing only whitespace collapses to no public paths.
func TestCmdArgs_NormalisePublicPaths_Ugly_BlankEntriesReturnNil(t *testing.T) {
	if got := normalisePublicPaths([]string{" ", "\t", ""}); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// TestCmdArgs_NormalisePublicPaths_Ugly_NormalisesRootAndTrailingSlashes
// verifies awkward path forms still collapse to stable public path values.
func TestCmdArgs_NormalisePublicPaths_Ugly_NormalisesRootAndTrailingSlashes(t *testing.T) {
	got := normalisePublicPaths([]string{" / ", "/status/", " status// ", "/status"})
	want := []string{"/", "/status"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
