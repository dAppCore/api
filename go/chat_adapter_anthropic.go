// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"encoding/json"
	"io"

	core "dappco.re/go"
)

const anthropicVersion = "2023-06-01"

type anthropicAdapter struct{}

// AnthropicAdapter maps OpenAI chat completions to/from Anthropic's /v1/messages
// (top-level system field, mandatory max_tokens, content blocks, SSE event stream).
func AnthropicAdapter() ChatFormatAdapter { return anthropicAdapter{} }

func (anthropicAdapter) Name() string         { return "anthropic" }
func (anthropicAdapter) UpstreamPath() string { return "/v1/messages" }

func anthropicFinish(stopReason string) string {
	switch stopReason {
	case "max_tokens":
		return "length"
	default: // end_turn, stop_sequence, etc.
		return "stop"
	}
}

func (anthropicAdapter) BuildRequest(req ChatCompletionRequest) ([]byte, map[string]string, error) {
	var system string
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			if system != "" {
				system += "\n"
			}
			system += m.Content
			continue
		}
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	maxTokens := chatDefaultMaxTokens
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	body := map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": maxTokens,
		"stream":     req.Stream,
	}
	if system != "" {
		body["system"] = system
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		body["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		body["top_k"] = *req.TopK
	}
	if len(req.Stop) > 0 {
		body["stop_sequences"] = []string(req.Stop)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, nil, core.E("anthropic", "marshal request", err)
	}
	return raw, map[string]string{"Content-Type": "application/json", "anthropic-version": anthropicVersion}, nil
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (anthropicAdapter) DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error) {
	var ar anthropicResponse
	if err := json.Unmarshal(upstream, &ar); err != nil {
		return ChatCompletionResponse{}, core.E("anthropic", "decode response", err)
	}
	var content string
	for _, b := range ar.Content {
		if b.Type == "text" {
			content += b.Text
		}
	}
	return ChatCompletionResponse{
		ID:      newChatCompletionID(),
		Object:  "chat.completion",
		Model:   model,
		Choices: []ChatChoice{{Index: 0, Message: ChatMessage{Role: "assistant", Content: content}, FinishReason: anthropicFinish(ar.StopReason)}},
		Usage:   ChatUsage{PromptTokens: ar.Usage.InputTokens, CompletionTokens: ar.Usage.OutputTokens, TotalTokens: ar.Usage.InputTokens + ar.Usage.OutputTokens},
	}, nil
}

func (anthropicAdapter) Transcoder() ChatStreamTranscoder { return anthropicTranscoder{} }

type anthropicTranscoder struct{}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
}

func (anthropicTranscoder) Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error {
	scanner := bufio.NewScanner(upstream)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	stopReason := "end_turn"
	for scanner.Scan() {
		line := core.Trim(scanner.Text())
		if !core.HasPrefix(line, "data:") {
			continue // skip "event:" and blank lines; the data line carries type
		}
		payload := core.Trim(line[len("data:"):])
		if payload == "" {
			continue
		}
		var ev anthropicStreamEvent
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type != "text_delta" || ev.Delta.Text == "" {
				continue
			}
			delta := ChatMessageDelta{Content: ev.Delta.Text}
			if first {
				delta.Role = "assistant"
				first = false
			}
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: nil}},
			})
		case "message_delta":
			if ev.Delta.StopReason != "" {
				stopReason = ev.Delta.StopReason
			}
		case "message_stop":
			fr := anthropicFinish(stopReason)
			delta := ChatMessageDelta{}
			if first { // empty/no-text stream — still prime the assistant role
				delta.Role = "assistant"
				first = false
			}
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: &fr}},
			})
			if err := scanner.Err(); err != nil {
				return err // truncated stream signals incomplete; do NOT emit [DONE]
			}
			writeSSEDone(w, flush)
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err // truncated stream signals incomplete; do NOT emit [DONE]
	}
	// Stream ended without an explicit message_stop — still terminate cleanly.
	writeSSEDone(w, flush)
	return nil
}
