// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestTransport_normaliseChatCompletionsPath_Good_TrimsAndKeepsCustomPath(t *testing.T) {
	if got := normaliseChatCompletionsPath(" /chat/ "); got != "/chat" {
		t.Fatalf("expected custom path to be trimmed, got %q", got)
	}
}

func TestTransport_normaliseChatCompletionsPath_Bad_FallsBackToDefaultWhenBlank(t *testing.T) {
	if got := normaliseChatCompletionsPath("   "); got != defaultChatCompletionsPath {
		t.Fatalf("expected blank input to fall back to %q, got %q", defaultChatCompletionsPath, got)
	}
}

func TestTransport_normaliseChatCompletionsPath_Ugly_FallsBackToDefaultWhenRoot(t *testing.T) {
	if got := normaliseChatCompletionsPath("/"); got != defaultChatCompletionsPath {
		t.Fatalf("expected root input to fall back to %q, got %q", defaultChatCompletionsPath, got)
	}
	if got := normaliseChatCompletionsPath("///"); got != defaultChatCompletionsPath {
		t.Fatalf("expected repeated root input to fall back to %q, got %q", defaultChatCompletionsPath, got)
	}
}

func TestTransport_TransportConfig_Ugly_NilEngineReturnsZeroValue(t *testing.T) {
	var e *Engine

	if got := e.TransportConfig(); got != (TransportConfig{}) {
		t.Fatalf("expected zero-value transport config for nil engine, got %+v", got)
	}
}
