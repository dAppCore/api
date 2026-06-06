// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

func TestAnthropicAdapter_BuildRequest_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	body, hdrs, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "claude-3", Messages: []api.ChatMessage{{Role: "system", Content: "be terse"}, {Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if hdrs["anthropic-version"] == "" {
		t.Errorf("missing anthropic-version header")
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if got["system"] != "be terse" {
		t.Errorf("system not extracted: %s", body)
	}
	msgs, _ := got["messages"].([]any)
	if len(msgs) != 1 { // system removed from messages
		t.Errorf("system not removed from messages: %s", body)
	}
	if _, ok := got["max_tokens"]; !ok {
		t.Errorf("max_tokens (mandatory) missing: %s", body)
	}
}

func TestAnthropicAdapter_DecodeResponse_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	out, err := a.DecodeResponse("claude-3", []byte(`{"content":[{"type":"text","text":"Hi"},{"type":"text","text":" there"}],"stop_reason":"max_tokens","usage":{"input_tokens":5,"output_tokens":2}}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Choices[0].Message.Content != "Hi there" {
		t.Errorf("text blocks not concatenated: %q", out.Choices[0].Message.Content)
	}
	if out.Choices[0].FinishReason != "length" {
		t.Errorf("max_tokens not mapped to length: %s", out.Choices[0].FinishReason)
	}
	if out.Usage.PromptTokens != 5 || out.Usage.CompletionTokens != 2 {
		t.Errorf("bad usage: %+v", out.Usage)
	}
}

func TestAnthropicAdapter_Transcode_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	// Minimal Anthropic event stream.
	stream := strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"usage":{"input_tokens":5}}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"He"}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"llo"}}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "claude-3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"He"`) || !strings.Contains(got, `"content":"llo"`) {
		t.Errorf("missing deltas: %s", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) || !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing terminal/[DONE]: %s", got)
	}
}
