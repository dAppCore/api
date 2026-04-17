// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestSSE_normaliseSSEPath_Good_TrimsAndKeepsCustomPath(t *testing.T) {
	if got := normaliseSSEPath(" /stream/ "); got != "/stream" {
		t.Fatalf("expected custom path to be trimmed, got %q", got)
	}
}

func TestSSE_normaliseSSEPath_Bad_ReturnsDefaultForBlankInput(t *testing.T) {
	if got := normaliseSSEPath("   "); got != defaultSSEPath {
		t.Fatalf("expected blank input to fall back to %q, got %q", defaultSSEPath, got)
	}
}

func TestSSE_normaliseSSEPath_Ugly_ReturnsDefaultForRootInput(t *testing.T) {
	if got := normaliseSSEPath("/"); got != defaultSSEPath {
		t.Fatalf("expected root input to fall back to %q, got %q", defaultSSEPath, got)
	}
}
