// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestWebsocket_normaliseWSPath_Good_TrimsAndKeepsCustomPath(t *testing.T) {
	if got := normaliseWSPath(" /socket/ "); got != "/socket" {
		t.Fatalf("expected custom path to be trimmed, got %q", got)
	}
}

func TestWebsocket_normaliseWSPath_Bad_ReturnsDefaultForBlankInput(t *testing.T) {
	if got := normaliseWSPath("   "); got != defaultWSPath {
		t.Fatalf("expected blank input to fall back to %q, got %q", defaultWSPath, got)
	}
}

func TestWebsocket_normaliseWSPath_Ugly_ReturnsDefaultForRootInput(t *testing.T) {
	if got := normaliseWSPath("/"); got != defaultWSPath {
		t.Fatalf("expected root input to fall back to %q, got %q", defaultWSPath, got)
	}
}
