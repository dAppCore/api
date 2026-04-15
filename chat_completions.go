// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"dappco.re/go/core"
	inference "dappco.re/go/core/inference"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const defaultChatCompletionsPath = "/v1/chat/completions"

const (
	chatDefaultTemperature = 1.0
	chatDefaultTopP        = 0.95
	chatDefaultTopK        = 64
	chatDefaultMaxTokens   = 2048
)

const channelMarker = "<|channel>"

// ChatCompletionRequest is the OpenAI-compatible request body.
//
//	body := ChatCompletionRequest{
//	    Model:    "lemer",
//	    Messages: []ChatMessage{{Role: "user", Content: "What is 2+2?"}},
//	    Stream:   true,
//	}
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float32      `json:"temperature,omitempty"`
	TopP        *float32      `json:"top_p,omitempty"`
	TopK        *int          `json:"top_k,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Stop        []string      `json:"stop,omitempty"` // Stop sequences, excluded from the final completion text.
	User        string        `json:"user,omitempty"`
}

// ChatMessage is a single turn in a conversation.
//
//	msg := ChatMessage{Role: "user", Content: "Hello"}
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse is the OpenAI-compatible response body.
//
//	resp.Choices[0].Message.Content // "4"
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
	Thought *string      `json:"thought,omitempty"`
}

// ChatChoice is a single response option.
//
//	choice.Message.Content  // The generated text
//	choice.FinishReason     // "stop", "length", or "error"
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage reports token consumption for the request.
//
//	usage.TotalTokens // PromptTokens + CompletionTokens
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionChunk is a single SSE chunk during streaming.
//
//	chunk.Choices[0].Delta.Content // Partial token text
type ChatCompletionChunk struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []ChatChunkChoice `json:"choices"`
	Thought *string           `json:"thought,omitempty"`
}

// ChatChunkChoice is a streaming delta.
//
//	delta.Content // New token(s) in this chunk
type ChatChunkChoice struct {
	Index        int              `json:"index"`
	Delta        ChatMessageDelta `json:"delta"`
	FinishReason *string          `json:"finish_reason"`
}

// ChatMessageDelta is the incremental content within a streaming chunk.
//
//	delta.Content // "" on first chunk (role-only), then token text
type ChatMessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// MarshalJSON preserves the OpenAI-style priming chunk shape while still
// omitting empty deltas for terminal chunks.
//
// The first streaming chunk carries the assistant role and an explicit empty
// content string. A terminal chunk, by contrast, carries neither field.
func (d ChatMessageDelta) MarshalJSON() ([]byte, error) {
	if d.Role == "" && d.Content == "" {
		return []byte("{}"), nil
	}

	payload := struct {
		Role    *string `json:"role,omitempty"`
		Content *string `json:"content,omitempty"`
	}{}

	if d.Role != "" {
		role := d.Role
		content := d.Content
		payload.Role = &role
		payload.Content = &content
	} else if d.Content != "" {
		content := d.Content
		payload.Content = &content
	}

	return json.Marshal(payload)
}

type chatCompletionError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code"`
}

type chatCompletionErrorResponse struct {
	Error chatCompletionError `json:"error"`
}

type modelResolutionError struct {
	code  string
	param string
	msg   string
}

func (e *modelResolutionError) Error() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// ModelResolver resolves model names to loaded inference.TextModel instances.
//
// Resolution order:
//
//  1. Exact cache hit
//  2. ~/.core/models.yaml path mapping
//  3. discovery by architecture via inference.Discover()
type ModelResolver struct {
	mu           sync.RWMutex
	loadedByName map[string]inference.TextModel
	loadedByPath map[string]inference.TextModel
	discovery    map[string]string
}

// NewModelResolver constructs a ModelResolver with empty caches. The returned
// resolver is safe for concurrent use — ResolveModel serialises cache updates
// through an internal sync.RWMutex.
//
//	resolver := api.NewModelResolver()
//	engine, _ := api.New(api.WithChatCompletions(resolver))
func NewModelResolver() *ModelResolver {
	return &ModelResolver{
		loadedByName: make(map[string]inference.TextModel),
		loadedByPath: make(map[string]inference.TextModel),
		discovery:    make(map[string]string),
	}
}

// ResolveModel maps a model name to a loaded inference.TextModel.
// Cached models are reused. Unknown names return an error.
func (r *ModelResolver) ResolveModel(name string) (inference.TextModel, error) {
	if r == nil {
		return nil, &modelResolutionError{
			code:  "model_not_found",
			param: "model",
			msg:   "model resolver is not configured",
		}
	}

	requested := core.Lower(strings.TrimSpace(name))
	if requested == "" {
		return nil, &modelResolutionError{
			code:  "invalid_request_error",
			param: "model",
			msg:   "model is required",
		}
	}

	r.mu.RLock()
	if cached, ok := r.loadedByName[requested]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	if path, ok := r.lookupModelPath(requested); ok {
		return r.loadByPath(requested, path)
	}

	if path, ok := r.resolveDiscoveredPath(requested); ok {
		return r.loadByPath(requested, path)
	}

	return nil, &modelResolutionError{
		code:  "model_not_found",
		param: "model",
		msg:   fmt.Sprintf("model %q not found", requested),
	}
}

func (r *ModelResolver) loadByPath(name, path string) (inference.TextModel, error) {
	cleanPath := core.Path(path)
	r.mu.Lock()
	if cached, ok := r.loadedByPath[cleanPath]; ok {
		r.loadedByName[name] = cached
		r.mu.Unlock()
		return cached, nil
	}
	r.mu.Unlock()

	loaded, err := inference.LoadModel(cleanPath)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "loading") {
			return nil, &modelResolutionError{
				code:  "model_loading",
				param: "model",
				msg:   err.Error(),
			}
		}
		return nil, &modelResolutionError{
			code:  "model_not_found",
			param: "model",
			msg:   err.Error(),
		}
	}

	r.mu.Lock()
	r.loadedByName[name] = loaded
	r.loadedByPath[cleanPath] = loaded
	r.mu.Unlock()
	return loaded, nil
}

func (r *ModelResolver) lookupModelPath(name string) (string, bool) {
	mappings, ok := r.modelsYAMLMapping()
	if !ok {
		return "", false
	}

	if path, ok := mappings[name]; ok && strings.TrimSpace(path) != "" {
		return path, true
	}
	return "", false
}

func (r *ModelResolver) modelsYAMLMapping() (map[string]string, bool) {
	configPath := core.Path(core.Env("DIR_HOME"), ".core", "models.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, false
	}

	var content any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, false
	}

	root, ok := content.(map[string]any)
	if !ok || root == nil {
		return nil, false
	}

	normalized := make(map[string]string)

	if models, ok := root["models"].(map[string]any); ok && models != nil {
		for key, raw := range models {
			if value, ok := raw.(string); ok {
				normalized[core.Lower(strings.TrimSpace(key))] = strings.TrimSpace(value)
			}
		}
	}

	for key, raw := range root {
		value, ok := raw.(string)
		if !ok {
			continue
		}
		normalized[core.Lower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}

	if len(normalized) == 0 {
		return nil, false
	}
	return normalized, true
}

func (r *ModelResolver) resolveDiscoveredPath(name string) (string, bool) {
	candidates := []string{name}
	if n := strings.IndexRune(name, ':'); n > 0 {
		candidates = append(candidates, name[:n])
	}

	r.mu.RLock()
	for _, candidate := range candidates {
		if path, ok := r.discovery[candidate]; ok {
			r.mu.RUnlock()
			return path, true
		}
	}
	r.mu.RUnlock()

	base := core.Path(core.Env("DIR_HOME"), ".core", "models")
	var discovered string
	for _, m := range discoveryModels(base) {
		modelType := strings.ToLower(strings.TrimSpace(m.ModelType))
		for _, candidate := range candidates {
			if candidate != "" && candidate == modelType {
				discovered = m.Path
				break
			}
		}
		if discovered != "" {
			break
		}
	}

	if discovered == "" {
		return "", false
	}

	r.mu.Lock()
	for _, candidate := range candidates {
		if candidate != "" {
			r.discovery[candidate] = discovered
		}
	}
	r.mu.Unlock()

	return discovered, true
}

type discoveredModel struct {
	Path      string
	ModelType string
}

// discoveryModels enumerates locally discovered models under base and
// returns Path + ModelType pairs for name resolution.
//
//	for _, m := range discoveryModels(base) {
//	    _ = m.Path
//	}
func discoveryModels(base string) []discoveredModel {
	var out []discoveredModel
	for m := range inference.Discover(base) {
		if m.Path == "" || m.ModelType == "" {
			continue
		}
		out = append(out, discoveredModel{Path: m.Path, ModelType: m.ModelType})
	}
	return out
}

// ThinkingExtractor separates thinking channel content from response text.
// Applied as a post-processing step on the token stream.
//
//	extractor := NewThinkingExtractor()
//	for tok := range model.Chat(ctx, messages) {
//		extractor.Process(tok)
//	}
//	response := extractor.Content()   // User-facing text
//	thinking := extractor.Thinking()  // Internal reasoning (may be nil)
type ThinkingExtractor struct {
	currentChannel string
	content        strings.Builder
	thought        strings.Builder
}

// NewThinkingExtractor constructs a ThinkingExtractor that starts on the
// "assistant" channel. Tokens are routed to Content() until a
// "<|channel>thought" marker switches the stream to the thinking channel (and
// similarly back).
//
//	extractor := api.NewThinkingExtractor()
func NewThinkingExtractor() *ThinkingExtractor {
	return &ThinkingExtractor{
		currentChannel: "assistant",
	}
}

// Process feeds a single generated token into the extractor. Tokens are
// appended to the current channel buffer (content or thought), switching on
// the "<|channel>NAME" marker. Non-streaming handlers call Process in a loop
// and then read Content and Thinking when generation completes.
//
//	for tok := range model.Chat(ctx, messages) {
//	    extractor.Process(tok)
//	}
func (te *ThinkingExtractor) Process(token inference.Token) {
	te.writeDeltas(token.Text)
}

// Content returns all text accumulated on the user-facing "assistant" channel
// so far. Safe to call on a nil receiver (returns "").
//
//	text := extractor.Content()
func (te *ThinkingExtractor) Content() string {
	if te == nil {
		return ""
	}
	return te.content.String()
}

// Thinking returns all text accumulated on the internal "thought" channel so
// far or nil when no thinking tokens were produced. Safe to call on a nil
// receiver.
//
//	if thinking := extractor.Thinking(); thinking != nil {
//	    response.Thought = thinking
//	}
func (te *ThinkingExtractor) Thinking() *string {
	if te == nil {
		return nil
	}
	if te.thought.Len() == 0 {
		return nil
	}
	out := te.thought.String()
	return &out
}

// writeDeltas tokenises text into the current channel, switching channels
// whenever it encounters the "<|channel>NAME" marker. It returns the content
// and thought fragments that were added to the builders during this call so
// streaming handlers can emit only the new bytes to the wire.
//
//	contentDelta, thoughtDelta := extractor.writeDeltas(tok.Text)
func (te *ThinkingExtractor) writeDeltas(text string) (string, string) {
	if te == nil {
		return "", ""
	}

	beforeContentLen := te.content.Len()
	beforeThoughtLen := te.thought.Len()

	remaining := text
	for {
		next := strings.Index(remaining, channelMarker)
		if next < 0 {
			te.writeToCurrentChannel(remaining)
			break
		}

		te.writeToCurrentChannel(remaining[:next])
		remaining = remaining[next+len(channelMarker):]

		remaining = strings.TrimLeftFunc(remaining, unicode.IsSpace)
		if remaining == "" {
			break
		}

		chanName, consumed := parseChannelName(remaining)
		if consumed <= 0 {
			te.writeToCurrentChannel(channelMarker)
			if remaining != "" {
				te.writeToCurrentChannel(remaining)
			}
			break
		}

		if chanName == "" {
			te.writeToCurrentChannel(channelMarker)
		} else {
			te.currentChannel = chanName
		}
		remaining = remaining[consumed:]
	}

	return te.content.String()[beforeContentLen:], te.thought.String()[beforeThoughtLen:]
}

func (te *ThinkingExtractor) writeToCurrentChannel(text string) {
	if text == "" {
		return
	}

	if te.currentChannel == "thought" {
		te.thought.WriteString(text)
		return
	}
	te.content.WriteString(text)
}

func parseChannelName(s string) (string, int) {
	if s == "" {
		return "", 0
	}
	count := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			count++
			continue
		}
		break
	}
	if count == 0 {
		return "", 0
	}
	return strings.ToLower(s[:count]), count
}

type chatCompletionsHandler struct {
	resolver *ModelResolver
}

func newChatCompletionsHandler(resolver *ModelResolver) *chatCompletionsHandler {
	return &chatCompletionsHandler{
		resolver: resolver,
	}
}

func (h *chatCompletionsHandler) ServeHTTP(c *gin.Context) {
	if h == nil || h.resolver == nil {
		writeChatCompletionError(c, http.StatusServiceUnavailable, "invalid_request_error", "model", "chat handler is not configured", "model")
		return
	}

	var req ChatCompletionRequest
	if err := decodeJSONBody(c.Request.Body, &req); err != nil {
		writeChatCompletionError(c, 400, "invalid_request_error", "body", "invalid request body", "")
		return
	}

	if err := validateChatRequest(&req); err != nil {
		chatErr, ok := err.(*chatCompletionRequestError)
		if !ok {
			writeChatCompletionError(c, http.StatusBadRequest, "invalid_request_error", "body", err.Error(), "")
			return
		}
		writeChatCompletionError(c, chatErr.Status, chatErr.Type, chatErr.Param, chatErr.Message, chatErr.Code)
		return
	}

	model, err := h.resolver.ResolveModel(req.Model)
	if err != nil {
		status, errType, errCode, errParam := mapResolverError(err)
		writeChatCompletionError(c, status, errType, errParam, err.Error(), errCode)
		return
	}

	reqForOptions := req
	reqForOptions.Stop = nil
	options, err := chatRequestOptions(&reqForOptions)
	if err != nil {
		writeChatCompletionError(c, 400, "invalid_request_error", "stop", err.Error(), "")
		return
	}
	stopSequences, err := normalizedStopSequences(req.Stop)
	if err != nil {
		writeChatCompletionError(c, 400, "invalid_request_error", "stop", err.Error(), "")
		return
	}

	messages := make([]inference.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, inference.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if req.Stream {
		h.serveStreaming(c, model, req, messages, stopSequences, options...)
		return
	}
	h.serveNonStreaming(c, model, req, messages, stopSequences, options...)
}

func (h *chatCompletionsHandler) serveNonStreaming(c *gin.Context, model inference.TextModel, req ChatCompletionRequest, messages []inference.Message, stopSequences []string, opts ...inference.GenerateOption) {
	ctx := c.Request.Context()
	created := time.Now().Unix()
	completionID := newChatCompletionID()

	extractor := NewThinkingExtractor()
	for tok := range model.Chat(ctx, messages, opts...) {
		extractor.Process(tok)
	}
	if err := model.Err(); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "loading") {
			writeChatCompletionError(c, http.StatusServiceUnavailable, "model_loading", "model", err.Error(), "")
			return
		}
		writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "")
		return
	}

	metrics := model.Metrics()
	content := truncateAtStopSequence(extractor.Content(), stopSequences)
	finishReason := "stop"
	if isTokenLengthCapReached(req.MaxTokens, metrics.GeneratedTokens) {
		finishReason = "length"
	}

	response := ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: created,
		Model:   req.Model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: ChatUsage{
			PromptTokens:     metrics.PromptTokens,
			CompletionTokens: metrics.GeneratedTokens,
			TotalTokens:      metrics.PromptTokens + metrics.GeneratedTokens,
		},
	}
	if thought := extractor.Thinking(); thought != nil {
		response.Thought = thought
	}

	c.JSON(http.StatusOK, response)
}

func (h *chatCompletionsHandler) serveStreaming(c *gin.Context, model inference.TextModel, req ChatCompletionRequest, messages []inference.Message, stopSequences []string, opts ...inference.GenerateOption) {
	ctx := c.Request.Context()
	created := time.Now().Unix()
	completionID := newChatCompletionID()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(200)
	c.Writer.Flush()

	// Emit the OpenAI-style role priming chunk before any generated content.
	primingChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []ChatChunkChoice{
			{
				Index: 0,
				Delta: ChatMessageDelta{
					Role: "assistant",
				},
				FinishReason: nil,
			},
		},
	}
	if encoded, encodeErr := json.Marshal(primingChunk); encodeErr == nil {
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", encoded))
		c.Writer.Flush()
	}

	extractor := NewThinkingExtractor()
	sentAny := true
	emittedContent := ""

	for tok := range model.Chat(ctx, messages, opts...) {
		contentDelta, thoughtDelta := extractor.writeDeltas(tok.Text)
		candidateContent := emittedContent + contentDelta
		stopCut, stopHit := firstStopSequenceCut(candidateContent, stopSequences)
		if stopHit {
			if stopCut <= len(emittedContent) {
				contentDelta = ""
			} else {
				contentDelta = candidateContent[len(emittedContent):stopCut]
			}
		}

		if !stopHit && contentDelta == "" && thoughtDelta == "" {
			continue
		}

		delta := ChatMessageDelta{
			Content: contentDelta,
		}

		chunk := ChatCompletionChunk{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []ChatChunkChoice{
				{
					Index:        0,
					Delta:        delta,
					FinishReason: nil,
				},
			},
		}
		if thoughtDelta != "" {
			t := thoughtDelta
			chunk.Thought = &t
		}

		if encoded, encodeErr := json.Marshal(chunk); encodeErr == nil {
			c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", encoded))
			c.Writer.Flush()
			sentAny = true
		}
		if stopHit {
			emittedContent = candidateContent[:stopCut]
		} else {
			emittedContent = candidateContent
		}
		if stopHit {
			break
		}
	}

	if err := model.Err(); err != nil && !sentAny {
		if strings.Contains(strings.ToLower(err.Error()), "loading") {
			writeChatCompletionError(c, http.StatusServiceUnavailable, "model_loading", "model", err.Error(), "")
			return
		}
		writeChatCompletionError(c, http.StatusInternalServerError, "inference_error", "model", err.Error(), "")
		return
	}

	finishReason := "stop"
	metrics := model.Metrics()
	if err := model.Err(); err != nil {
		finishReason = "error"
	}
	if finishReason != "error" && isTokenLengthCapReached(req.MaxTokens, metrics.GeneratedTokens) {
		finishReason = "length"
	}

	finished := finishReason
	finalChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []ChatChunkChoice{
			{
				Index:        0,
				Delta:        ChatMessageDelta{},
				FinishReason: &finished,
			},
		},
	}
	if encoded, encodeErr := json.Marshal(finalChunk); encodeErr == nil {
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", encoded))
	}
	c.Writer.WriteString("data: [DONE]\n\n")
	c.Writer.Flush()
}

type chatCompletionRequestError struct {
	Status  int
	Type    string
	Code    string
	Param   string
	Message string
}

func (e *chatCompletionRequestError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func validateChatRequest(req *ChatCompletionRequest) error {
	if strings.TrimSpace(req.Model) == "" {
		return &chatCompletionRequestError{
			Status:  400,
			Type:    "invalid_request_error",
			Code:    "invalid_request_error",
			Param:   "model",
			Message: "model is required",
		}
	}

	if len(req.Messages) == 0 {
		return &chatCompletionRequestError{
			Status:  400,
			Type:    "invalid_request_error",
			Code:    "invalid_request_error",
			Param:   "messages",
			Message: "messages must be a non-empty array",
		}
	}

	for i, msg := range req.Messages {
		if strings.TrimSpace(msg.Role) == "" {
			return &chatCompletionRequestError{
				Status:  400,
				Type:    "invalid_request_error",
				Code:    "invalid_request_error",
				Param:   fmt.Sprintf("messages[%d].role", i),
				Message: "message role is required",
			}
		}
	}

	return nil
}

func chatRequestOptions(req *ChatCompletionRequest) ([]inference.GenerateOption, error) {
	opts := make([]inference.GenerateOption, 0, 5)
	opts = append(opts, inference.WithTemperature(chatResolvedFloat(req.Temperature, chatDefaultTemperature)))
	opts = append(opts, inference.WithTopP(chatResolvedFloat(req.TopP, chatDefaultTopP)))
	opts = append(opts, inference.WithTopK(chatResolvedInt(req.TopK, chatDefaultTopK)))
	opts = append(opts, inference.WithMaxTokens(chatResolvedInt(req.MaxTokens, chatDefaultMaxTokens)))
	stops, err := parsedStopTokens(req.Stop)
	if err != nil {
		return nil, err
	}
	if len(stops) > 0 {
		opts = append(opts, inference.WithStopTokens(stops...))
	}
	return opts, nil
}

// chatResolvedFloat honours an explicitly set float sampling parameter or
// falls back to the calibrated default when the pointer is nil.
//
// Spec §11.2: "When a parameter is omitted (nil), the server applies the
// calibrated default. When explicitly set (including 0.0), the server honours
// the caller's value."
//
//	temperature := chatResolvedFloat(req.Temperature, chatDefaultTemperature)
func chatResolvedFloat(v *float32, def float32) float32 {
	if v == nil {
		return def
	}
	return *v
}

// chatResolvedInt honours an explicitly set integer sampling parameter or
// falls back to the calibrated default when the pointer is nil.
//
//	topK := chatResolvedInt(req.TopK, chatDefaultTopK)
func chatResolvedInt(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}

func normalizedStopSequences(stops []string) ([]string, error) {
	if len(stops) == 0 {
		return nil, nil
	}

	out := make([]string, 0, len(stops))
	for _, raw := range stops {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, fmt.Errorf("stop entries cannot be empty")
		}
		out = append(out, raw)
	}
	return out, nil
}

func parsedStopTokens(stops []string) ([]int32, error) {
	if len(stops) == 0 {
		return nil, nil
	}

	out := make([]int32, 0, len(stops))
	for _, raw := range stops {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, fmt.Errorf("stop entries cannot be empty")
		}
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid stop token %q", raw)
		}
		out = append(out, int32(parsed))
	}
	return out, nil
}

func firstStopSequenceCut(content string, stops []string) (int, bool) {
	if len(stops) == 0 || content == "" {
		return 0, false
	}

	cut := -1
	for _, stop := range stops {
		if stop == "" {
			continue
		}
		idx := strings.Index(content, stop)
		if idx < 0 {
			continue
		}
		if cut < 0 || idx < cut {
			cut = idx
		}
	}

	if cut < 0 {
		return 0, false
	}
	return cut, true
}

func truncateAtStopSequence(content string, stops []string) string {
	cut, ok := firstStopSequenceCut(content, stops)
	if !ok {
		return content
	}
	return content[:cut]
}

// isTokenLengthCapReached reports whether the generated token count meets or
// exceeds the caller's max_tokens budget. Nil or non-positive caps disable the
// check (streams terminate by backend signal only).
//
//	if isTokenLengthCapReached(req.MaxTokens, metrics.GeneratedTokens) {
//	    finishReason = "length"
//	}
func isTokenLengthCapReached(maxTokens *int, generated int) bool {
	if maxTokens == nil || *maxTokens <= 0 {
		return false
	}
	return generated >= *maxTokens
}

func mapResolverError(err error) (int, string, string, string) {
	resErr, ok := err.(*modelResolutionError)
	if !ok {
		return 500, "inference_error", "inference_error", "model"
	}
	switch resErr.code {
	case "model_loading":
		return http.StatusServiceUnavailable, "model_loading", "model_loading", resErr.param
	case "model_not_found":
		return 404, "model_not_found", "model_not_found", resErr.param
	default:
		return 500, "inference_error", "inference_error", resErr.param
	}
}

func writeChatCompletionError(c *gin.Context, status int, errType, param, message, code string) {
	if status <= 0 {
		status = http.StatusInternalServerError
	}
	resp := chatCompletionErrorResponse{
		Error: chatCompletionError{
			Message: message,
			Type:    errType,
			Param:   param,
			Code:    codeOrDefault(code, errType),
		},
	}
	c.Header("Content-Type", "application/json")
	if status == http.StatusServiceUnavailable {
		// Retry-After must be set BEFORE c.JSON commits headers to the
		// wire. RFC 9110 §10.2.3 allows either seconds or an HTTP-date;
		// we use seconds for simplicity and OpenAI parity.
		c.Header("Retry-After", "10")
	}
	c.JSON(status, resp)
}

func codeOrDefault(code, fallback string) string {
	if code != "" {
		return code
	}
	if fallback != "" {
		return fallback
	}
	return "inference_error"
}

func newChatCompletionID() string {
	return fmt.Sprintf("chatcmpl-%d-%06d", time.Now().Unix(), rand.Intn(1_000_000))
}

func decodeJSONBody(reader io.Reader, dest any) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}
