// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	inference "dappco.re/go/core/inference"

	"github.com/gin-gonic/gin"
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

type chatModelStub struct {
	tokens  []inference.Token
	err     error
	metrics inference.GenerateMetrics
}

func (m *chatModelStub) Generate(ctx context.Context, prompt string, opts ...inference.GenerateOption) iter.Seq[inference.Token] {
	return m.seq()
}

func (m *chatModelStub) Chat(ctx context.Context, messages []inference.Message, opts ...inference.GenerateOption) iter.Seq[inference.Token] {
	return m.seq()
}

func (m *chatModelStub) seq() iter.Seq[inference.Token] {
	tokens := append([]inference.Token(nil), m.tokens...)
	return func(yield func(inference.Token) bool) {
		for _, tok := range tokens {
			if !yield(tok) {
				return
			}
		}
	}
}

func (m *chatModelStub) Classify(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error) {
	return nil, nil
}

func (m *chatModelStub) BatchGenerate(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.BatchResult, error) {
	return nil, nil
}

func (m *chatModelStub) ModelType() string { return "stub" }

func (m *chatModelStub) Info() inference.ModelInfo { return inference.ModelInfo{} }

func (m *chatModelStub) Metrics() inference.GenerateMetrics { return m.metrics }

func (m *chatModelStub) Err() error { return m.err }

func (m *chatModelStub) Close() error { return nil }

func newChatLoopbackRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newChatHandlerWithModel(model inference.TextModel) *chatCompletionsHandler {
	resolver := NewModelResolver()
	resolver.loadedByName["lemer"] = model
	return newChatCompletionsHandler(resolver)
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Good_PreservesRoleAndContent(t *testing.T) {
	delta := ChatMessageDelta{Role: "assistant", Content: ""}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := string(data), `{"role":"assistant","content":""}`; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Bad_EncodesContentOnly(t *testing.T) {
	delta := ChatMessageDelta{Content: "token"}

	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := string(data), `{"content":"token"}`; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Ugly_EncodesEmptyObject(t *testing.T) {
	data, err := json.Marshal(ChatMessageDelta{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := string(data), `{}`; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestChatCompletions_normalizedStopSequences_Good_TrimsAndPreservesOrder(t *testing.T) {
	got, err := normalizedStopSequences([]string{"  END  ", "\tSTOP\n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"END", "STOP"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestChatCompletions_normalizedStopSequences_Bad_ReturnsNilForEmptyInput(t *testing.T) {
	got, err := normalizedStopSequences(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil slice, got %v", got)
	}
}

func TestChatCompletions_normalizedStopSequences_Ugly_RejectsBlankEntries(t *testing.T) {
	if _, err := normalizedStopSequences([]string{"   "}); err == nil {
		t.Fatal("expected blank stop entry to be rejected")
	}
}

func TestChatCompletions_firstStopSequenceCut_Good_ChoosesEarliestMatch(t *testing.T) {
	cut, ok := firstStopSequenceCut("abc STOP xyz END", []string{"END", "STOP"})
	if !ok {
		t.Fatal("expected stop sequence to be found")
	}
	if cut != 4 {
		t.Fatalf("expected earliest match at index 4, got %d", cut)
	}
}

func TestChatCompletions_firstStopSequenceCut_Bad_ReturnsFalseWhenNoMatch(t *testing.T) {
	if cut, ok := firstStopSequenceCut("abcdef", []string{"STOP"}); ok || cut != 0 {
		t.Fatalf("expected no match, got cut=%d ok=%v", cut, ok)
	}
}

func TestChatCompletions_firstStopSequenceCut_Ugly_HandlesEmptyInput(t *testing.T) {
	if cut, ok := firstStopSequenceCut("", []string{"STOP"}); ok || cut != 0 {
		t.Fatalf("expected empty content to return no match, got cut=%d ok=%v", cut, ok)
	}
	if cut, ok := firstStopSequenceCut("content", nil); ok || cut != 0 {
		t.Fatalf("expected empty stop list to return no match, got cut=%d ok=%v", cut, ok)
	}
}

func TestChatCompletions_truncateAtStopSequence_Good_TruncatesOnFirstMatch(t *testing.T) {
	if got, want := truncateAtStopSequence("hello STOP world END", []string{"END", "STOP"}), "hello "; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestChatCompletions_truncateAtStopSequence_Bad_LeavesContentWhenNoMatch(t *testing.T) {
	if got, want := truncateAtStopSequence("hello world", []string{"STOP"}), "hello world"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestChatCompletions_truncateAtStopSequence_Ugly_PassesThroughEmptyStops(t *testing.T) {
	if got, want := truncateAtStopSequence("hello world", nil), "hello world"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestChatCompletions_codeOrDefault_Good_ReturnsExplicitCode(t *testing.T) {
	if got := codeOrDefault("model_not_found", "fallback"); got != "model_not_found" {
		t.Fatalf("expected explicit code to win, got %q", got)
	}
}

func TestChatCompletions_codeOrDefault_Bad_UsesFallbackWhenCodeEmpty(t *testing.T) {
	if got := codeOrDefault("", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback code, got %q", got)
	}
}

func TestChatCompletions_codeOrDefault_Ugly_UsesGenericInferenceErrorWhenBothEmpty(t *testing.T) {
	if got := codeOrDefault("", ""); got != "inference_error" {
		t.Fatalf("expected inference_error fallback, got %q", got)
	}
}

func TestChatCompletions_mapResolverError_Good_MapsKnownCodes(t *testing.T) {
	status, errType, code, param := mapResolverError(&modelResolutionError{code: "model_loading", param: "model", msg: "loading"})
	if status != http.StatusServiceUnavailable || errType != "model_loading" || code != "model_loading" || param != "model" {
		t.Fatalf("unexpected mapping for model_loading: %d %q %q %q", status, errType, code, param)
	}

	status, errType, code, param = mapResolverError(&modelResolutionError{code: "model_not_found", param: "model", msg: "missing"})
	if status != http.StatusNotFound || errType != "model_not_found" || code != "model_not_found" || param != "model" {
		t.Fatalf("unexpected mapping for model_not_found: %d %q %q %q", status, errType, code, param)
	}
}

func TestChatCompletions_mapResolverError_Bad_FallsBackForUnknownErrors(t *testing.T) {
	status, errType, code, param := mapResolverError(fmt.Errorf("boom"))
	if status != http.StatusInternalServerError || errType != "inference_error" || code != "inference_error" || param != "model" {
		t.Fatalf("unexpected fallback mapping: %d %q %q %q", status, errType, code, param)
	}
}

func TestChatCompletions_mapResolverError_Ugly_HandlesNilResolutionError(t *testing.T) {
	status, errType, code, param := mapResolverError(nil)
	if status != http.StatusInternalServerError || errType != "inference_error" || code != "inference_error" || param != "model" {
		t.Fatalf("unexpected mapping for nil error: %d %q %q %q", status, errType, code, param)
	}
}

func TestChatCompletions_chatCompletionRequestError_Error_Good_ReturnsMessage(t *testing.T) {
	err := (&chatCompletionRequestError{Message: "bad request"}).Error()
	if err != "bad request" {
		t.Fatalf("expected message to round-trip, got %q", err)
	}
}

func TestChatCompletions_chatCompletionRequestError_Error_Ugly_NilReceiverReturnsEmpty(t *testing.T) {
	var err *chatCompletionRequestError
	if got := err.Error(); got != "" {
		t.Fatalf("expected nil receiver to return empty string, got %q", got)
	}
}

func TestChatCompletions_modelResolutionError_Error_Good_ReturnsMessage(t *testing.T) {
	err := (&modelResolutionError{msg: "missing"}).Error()
	if err != "missing" {
		t.Fatalf("expected message to round-trip, got %q", err)
	}
}

func TestChatCompletions_modelResolutionError_Error_Ugly_NilReceiverReturnsEmpty(t *testing.T) {
	var err *modelResolutionError
	if got := err.Error(); got != "" {
		t.Fatalf("expected nil receiver to return empty string, got %q", got)
	}
}

func TestChatCompletions_newChatCompletionID_Good_UsesExpectedPrefix(t *testing.T) {
	if got := newChatCompletionID(); !strings.HasPrefix(got, "chatcmpl-") {
		t.Fatalf("expected chat completion ID prefix, got %q", got)
	}
}

func TestChatCompletions_ResolveModel_Good_UsesCachePathAndDiscovery(t *testing.T) {
	t.Run("exact-cache-hit", func(t *testing.T) {
		model := &chatModelStub{}
		resolver := NewModelResolver()
		resolver.loadedByName["lemer"] = model

		got, err := resolver.ResolveModel("  LEMER  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != model {
			t.Fatal("expected exact cache hit to return cached model")
		}
	})

	t.Run("discovery", func(t *testing.T) {
		model := &chatModelStub{}
		resolver := NewModelResolver()
		modelDir := filepath.Join(t.TempDir(), "gemma3")
		resolver.discovery["gemma3"] = modelDir
		resolver.loadedByPath[modelDir] = model

		got, err := resolver.ResolveModel("gemma3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != model {
			t.Fatal("expected discovery cache to reuse cached path model")
		}
		if cached := resolver.loadedByName["gemma3"]; cached != model {
			t.Fatal("expected named cache entry to be populated after discovery")
		}
	})
}

// TestChatCompletions_lookupModelPath_Bad_NeedsDirHomeSeam documents the
// missing seam for redirecting core.Env("DIR_HOME") during unit tests.
func TestChatCompletions_lookupModelPath_Bad_NeedsDirHomeSeam(t *testing.T) {
	t.Skip("missing seam: core.Env(\"DIR_HOME\") is snapshotted at init, so models.yaml lookup cannot be redirected to a temp directory in a unit test")
}

func TestChatCompletions_discoveryModels_Good_FindsValidModels(t *testing.T) {
	base := t.TempDir()
	modelDir := filepath.Join(base, "gemma3")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "config.json"), []byte(`{"model_type":"gemma3"}`), 0o600); err != nil {
		t.Fatalf("write config.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "weights.safetensors"), []byte("stub"), 0o600); err != nil {
		t.Fatalf("write safetensors: %v", err)
	}

	got := discoveryModels(base)
	if len(got) != 1 {
		t.Fatalf("expected one discovered model, got %d", len(got))
	}
	if got[0].Path != modelDir {
		t.Fatalf("expected discovered path %q, got %q", modelDir, got[0].Path)
	}
	if got[0].ModelType != "gemma3" {
		t.Fatalf("expected discovered model type gemma3, got %q", got[0].ModelType)
	}
}

func TestChatCompletions_ResolveModel_Bad_ReturnsNotFoundForUnknownModel(t *testing.T) {
	resolver := NewModelResolver()

	got, err := resolver.ResolveModel("missing-model")
	if err == nil {
		t.Fatal("expected resolution error")
	}
	if got != nil {
		t.Fatalf("expected nil model, got %v", got)
	}
	if err.Error() == "" {
		t.Fatal("expected error message")
	}
}

func TestChatCompletions_ResolveModel_Ugly_RejectsEmptyAndNilResolver(t *testing.T) {
	resolver := NewModelResolver()

	if _, err := resolver.ResolveModel("   "); err == nil {
		t.Fatal("expected empty model name to be rejected")
	}

	var nilResolver *ModelResolver
	if _, err := nilResolver.ResolveModel("lemer"); err == nil {
		t.Fatal("expected nil resolver to be rejected")
	}
}

func TestChatCompletions_ServeHTTP_Good_NonStreamingResponseIncludesThoughtAndStops(t *testing.T) {
	gin.SetMode(gin.TestMode)

	model := &chatModelStub{
		tokens: []inference.Token{
			{Text: "Hello "},
			{Text: "<|channel>thought planning... "},
			{Text: "<|channel>assistant WorldSTOP and more"},
		},
		metrics: inference.GenerateMetrics{
			PromptTokens:    7,
			GeneratedTokens: 3,
		},
	}
	handler := newChatHandlerWithModel(model)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = newChatLoopbackRequest(t, `{
		"model": "lemer",
		"messages": [{"role":"user","content":"hi"}],
		"stop": ["STOP"]
	}`)

	handler.ServeHTTP(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var resp ChatCompletionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Object != "chat.completion" {
		t.Fatalf("expected chat.completion object, got %q", resp.Object)
	}
	if resp.Model != "lemer" {
		t.Fatalf("expected model lemer, got %q", resp.Model)
	}
	if !strings.HasPrefix(resp.ID, "chatcmpl-") {
		t.Fatalf("expected chat completion ID prefix, got %q", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected single choice, got %d", len(resp.Choices))
	}
	if got := resp.Choices[0].Message.Content; got != "Hello  World" {
		t.Fatalf("expected truncated assistant content, got %q", got)
	}
	if got := resp.Choices[0].FinishReason; got != "stop" {
		t.Fatalf("expected finish_reason=stop, got %q", got)
	}
	if resp.Thought == nil || *resp.Thought != " planning... " {
		t.Fatalf("expected thought content to be captured, got %v", resp.Thought)
	}
	if resp.Usage.PromptTokens != 7 || resp.Usage.CompletionTokens != 3 || resp.Usage.TotalTokens != 10 {
		t.Fatalf("unexpected usage accounting: %+v", resp.Usage)
	}
}

func TestChatCompletions_ServeHTTP_Good_StreamingResponseEmitsSSEChunks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	model := &chatModelStub{
		tokens: []inference.Token{
			{Text: "Hello "},
			{Text: "<|channel>thought planning... "},
			{Text: "<|channel>assistant WorldSTOP and more"},
		},
		metrics: inference.GenerateMetrics{
			PromptTokens:    7,
			GeneratedTokens: 3,
		},
	}
	handler := newChatHandlerWithModel(model)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = newChatLoopbackRequest(t, `{
		"model": "lemer",
		"messages": [{"role":"user","content":"hi"}],
		"stream": true,
		"stop": ["STOP"]
	}`)

	handler.ServeHTTP(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("expected SSE content type, got %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("expected cache-control=no-cache, got %q", got)
	}
	if got := rec.Header().Get("Connection"); got != "keep-alive" {
		t.Fatalf("expected connection=keep-alive, got %q", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `data: {"id":"chatcmpl-`) {
		t.Fatalf("expected streamed completion ID, got %s", body)
	}
	if !strings.Contains(body, `"role":"assistant"`) {
		t.Fatalf("expected role priming chunk, got %s", body)
	}
	if !strings.Contains(body, `"thought":" planning... "`) {
		t.Fatalf("expected thought chunk, got %s", body)
	}
	if !strings.Contains(body, `data: [DONE]`) {
		t.Fatalf("expected stream terminator, got %s", body)
	}
}

func TestChatCompletions_ServeHTTP_Bad_StreamingModelLoadingReturnsErrorBeforeBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	model := &chatModelStub{
		err: fmt.Errorf("model is loading"),
	}
	handler := newChatHandlerWithModel(model)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = newChatLoopbackRequest(t, `{
		"model": "lemer",
		"messages": [{"role":"user","content":"hi"}],
		"stream": true
	}`)

	handler.ServeHTTP(ctx)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d (%s)", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Retry-After"); got != "10" {
		t.Fatalf("expected Retry-After=10, got %q", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON error content type, got %q", got)
	}

	var payload chatCompletionErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON error response: %v", err)
	}
	if payload.Error.Code != "model_loading" {
		t.Fatalf("expected model_loading code, got %q", payload.Error.Code)
	}
	if payload.Error.Param != "model" {
		t.Fatalf("expected param=model, got %q", payload.Error.Param)
	}
}
