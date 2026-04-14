// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"

	inference "dappco.re/go/core/inference"
)

// TestChatCompletions_chatResolvedFloat_Good_ReturnsDefaultWhenNil verifies the
// calibrated default wins when the caller omits the parameter (pointer is nil).
//
// Spec §11.2 — "When a parameter is omitted (nil), the server applies the
// calibrated default."
func TestChatCompletions_chatResolvedFloat_Good_ReturnsDefaultWhenNil(t *testing.T) {
	got := chatResolvedFloat(nil, chatDefaultTemperature)
	if got != chatDefaultTemperature {
		t.Fatalf("expected default %v, got %v", chatDefaultTemperature, got)
	}
}

// TestChatCompletions_chatResolvedFloat_Good_HonoursExplicitZero verifies that
// an explicitly-set zero value overrides the default.
//
// Spec §11.2 — "When explicitly set (including 0.0), the server honours the
// caller's value."
func TestChatCompletions_chatResolvedFloat_Good_HonoursExplicitZero(t *testing.T) {
	zero := float32(0.0)
	got := chatResolvedFloat(&zero, chatDefaultTemperature)
	if got != 0.0 {
		t.Fatalf("expected explicit 0.0 to be honoured, got %v", got)
	}
}

// TestChatCompletions_chatResolvedInt_Good_ReturnsDefaultWhenNil mirrors the
// float variant for integer sampling parameters (top_k, max_tokens).
func TestChatCompletions_chatResolvedInt_Good_ReturnsDefaultWhenNil(t *testing.T) {
	got := chatResolvedInt(nil, chatDefaultTopK)
	if got != chatDefaultTopK {
		t.Fatalf("expected default %d, got %d", chatDefaultTopK, got)
	}
}

// TestChatCompletions_chatResolvedInt_Good_HonoursExplicitZero verifies that
// an explicitly-set zero integer overrides the default.
func TestChatCompletions_chatResolvedInt_Good_HonoursExplicitZero(t *testing.T) {
	zero := 0
	got := chatResolvedInt(&zero, chatDefaultTopK)
	if got != 0 {
		t.Fatalf("expected explicit 0 to be honoured, got %d", got)
	}
}

// TestChatCompletions_chatRequestOptions_Good_AppliesGemmaDefaults verifies
// that an otherwise-empty request produces the Gemma 4 calibrated sampling
// defaults documented in RFC §11.2.
//
//	temperature 1.0, top_p 0.95, top_k 64, max_tokens 2048
func TestChatCompletions_chatRequestOptions_Good_AppliesGemmaDefaults(t *testing.T) {
	req := &ChatCompletionRequest{Model: "lemer"}

	opts, err := chatRequestOptions(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(opts) == 0 {
		t.Fatal("expected at least one inference option for defaults")
	}

	cfg := inference.ApplyGenerateOpts(opts)
	if cfg.Temperature != chatDefaultTemperature {
		t.Fatalf("expected default temperature %v, got %v", chatDefaultTemperature, cfg.Temperature)
	}
	if cfg.TopP != chatDefaultTopP {
		t.Fatalf("expected default top_p %v, got %v", chatDefaultTopP, cfg.TopP)
	}
	if cfg.TopK != chatDefaultTopK {
		t.Fatalf("expected default top_k %d, got %d", chatDefaultTopK, cfg.TopK)
	}
	if cfg.MaxTokens != chatDefaultMaxTokens {
		t.Fatalf("expected default max_tokens %d, got %d", chatDefaultMaxTokens, cfg.MaxTokens)
	}
}

// TestChatCompletions_chatRequestOptions_Good_HonoursExplicitSampling verifies
// that caller-supplied sampling parameters (including zero for greedy decoding)
// override the Gemma 4 calibrated defaults.
func TestChatCompletions_chatRequestOptions_Good_HonoursExplicitSampling(t *testing.T) {
	temp := float32(0.0)
	topP := float32(0.5)
	topK := 10
	maxTokens := 512

	req := &ChatCompletionRequest{
		Model:       "lemer",
		Temperature: &temp,
		TopP:        &topP,
		TopK:        &topK,
		MaxTokens:   &maxTokens,
	}

	opts, err := chatRequestOptions(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := inference.ApplyGenerateOpts(opts)
	if cfg.Temperature != 0.0 {
		t.Fatalf("expected explicit temperature 0.0, got %v", cfg.Temperature)
	}
	if cfg.TopP != 0.5 {
		t.Fatalf("expected explicit top_p 0.5, got %v", cfg.TopP)
	}
	if cfg.TopK != 10 {
		t.Fatalf("expected explicit top_k 10, got %d", cfg.TopK)
	}
	if cfg.MaxTokens != 512 {
		t.Fatalf("expected explicit max_tokens 512, got %d", cfg.MaxTokens)
	}
}

// TestChatCompletions_chatRequestOptions_Bad_RejectsMalformedStop verifies the
// stop-token parser surfaces malformed values rather than silently ignoring
// them.
func TestChatCompletions_chatRequestOptions_Bad_RejectsMalformedStop(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "lemer",
		Stop:  []string{"oops"},
	}
	if _, err := chatRequestOptions(req); err == nil {
		t.Fatal("expected malformed stop entry to produce an error")
	}
}

// TestChatCompletions_chatRequestOptions_Ugly_EmptyStopEntryRejected ensures
// an all-whitespace stop entry is treated as invalid rather than as zero.
func TestChatCompletions_chatRequestOptions_Ugly_EmptyStopEntryRejected(t *testing.T) {
	req := &ChatCompletionRequest{
		Model: "lemer",
		Stop:  []string{" "},
	}
	if _, err := chatRequestOptions(req); err == nil {
		t.Fatal("expected empty stop entry to produce an error")
	}
}

// TestChatCompletions_isTokenLengthCapReached_Good_RespectsCap documents the
// finish_reason=length contract — generation that meets the caller's
// max_tokens budget is reported as length-capped.
func TestChatCompletions_isTokenLengthCapReached_Good_RespectsCap(t *testing.T) {
	cap := 10
	if !isTokenLengthCapReached(&cap, 10) {
		t.Fatal("expected cap to be reached when generated == max_tokens")
	}
	if !isTokenLengthCapReached(&cap, 20) {
		t.Fatal("expected cap to be reached when generated > max_tokens")
	}
	if isTokenLengthCapReached(&cap, 5) {
		t.Fatal("expected cap not reached when generated < max_tokens")
	}
}

// TestChatCompletions_isTokenLengthCapReached_Ugly_NilOrZeroDisablesCap
// documents that the cap is disabled when max_tokens is unset or non-positive.
func TestChatCompletions_isTokenLengthCapReached_Ugly_NilOrZeroDisablesCap(t *testing.T) {
	if isTokenLengthCapReached(nil, 999_999) {
		t.Fatal("expected nil max_tokens to disable the cap")
	}
	zero := 0
	if isTokenLengthCapReached(&zero, 999_999) {
		t.Fatal("expected zero max_tokens to disable the cap")
	}
	neg := -1
	if isTokenLengthCapReached(&neg, 999_999) {
		t.Fatal("expected negative max_tokens to disable the cap")
	}
}

// TestChatCompletions_ThinkingExtractor_Good_SeparatesThoughtFromContent
// verifies the <|channel>thought marker routes tokens to Thinking() and
// subsequent <|channel>assistant tokens land back in Content(). Covers RFC
// §11.6.
//
// The extractor skips whitespace between the marker and the channel name
// ("<|channel>thought ...") but preserves whitespace inside channel bodies —
// so "Hello " + thought block + " World" arrives as "Hello  World" with
// both separating spaces retained.
func TestChatCompletions_ThinkingExtractor_Good_SeparatesThoughtFromContent(t *testing.T) {
	ex := NewThinkingExtractor()
	ex.Process(inference.Token{Text: "Hello"})
	ex.Process(inference.Token{Text: "<|channel>thought planning... "})
	ex.Process(inference.Token{Text: "<|channel>assistant World"})

	content := ex.Content()
	if content != "Hello World" {
		t.Fatalf("expected content %q, got %q", "Hello World", content)
	}
	thinking := ex.Thinking()
	if thinking == nil {
		t.Fatal("expected thinking content to be captured")
	}
	if *thinking != " planning... " {
		t.Fatalf("expected thinking %q, got %q", " planning... ", *thinking)
	}
}

// TestChatCompletions_ThinkingExtractor_Ugly_NilReceiverIsSafe documents the
// nil-safe accessors so middleware can call them defensively.
func TestChatCompletions_ThinkingExtractor_Ugly_NilReceiverIsSafe(t *testing.T) {
	var ex *ThinkingExtractor
	if got := ex.Content(); got != "" {
		t.Fatalf("expected empty content on nil receiver, got %q", got)
	}
	if got := ex.Thinking(); got != nil {
		t.Fatalf("expected nil thinking on nil receiver, got %v", got)
	}
}
