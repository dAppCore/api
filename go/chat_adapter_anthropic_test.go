// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

// errAfterReader yields data once, then errors — simulating a truncated upstream
// stream (e.g. connection reset mid-response).
type errAfterReader struct {
	data []byte
	err  error
	done bool
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if !r.done {
		r.done = true
		n := copy(p, r.data)
		return n, nil
	}
	return 0, r.err
}

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

func TestAnthropicAdapter_Transcode_EmptyStream_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	stream := strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"usage":{"input_tokens":5}}}`,
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
	if !strings.Contains(got, `"role":"assistant"`) {
		t.Errorf("empty stream did not prime assistant role: %s", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) || !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing finish chunk/[DONE]: %s", got)
	}
}

func TestAnthropicAdapter_Transcode_Truncated_Ugly(t *testing.T) {
	a := api.AnthropicAdapter()
	r := &errAfterReader{
		data: []byte(`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"He"}}` + "\n\n"),
		err:  errors.New("reset"),
	}
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, r, api.ChatStreamMeta{ID: "id", Model: "claude-3", Created: 1})
	if err == nil {
		t.Fatal("truncated stream: want non-nil error, got nil")
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"He"`) {
		t.Errorf("partial content delta not emitted: %s", got)
	}
	if strings.Contains(got, "data: [DONE]") {
		t.Errorf("truncated stream must NOT emit [DONE]: %s", got)
	}
}

func TestAnthropicAdapter_Transcode_MalformedLineSkipped_Ugly(t *testing.T) {
	a := api.AnthropicAdapter()
	stream := strings.Join([]string{
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"He"}}`,
		"",
		"event: content_block_delta",
		`data: {not valid json`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"llo"}}`,
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
		t.Errorf("malformed line aborted the stream — deltas lost: %s", got)
	}
	if !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing [DONE]: %s", got)
	}
}

func TestAnthropicAdapter_Transcode_UnknownEvent_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	stream := strings.Join([]string{
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"He"}}`,
		"",
		"event: ping",
		`data: {"type":"ping"}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"llo"}}`,
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
		t.Errorf("unknown event disrupted deltas: %s", got)
	}
	if !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing [DONE]: %s", got)
	}
}

func TestAnthropicAdapter_Transcode_MultiBlock_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	stream := strings.Join([]string{
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"first"}}`,
		"",
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":0}`,
		"",
		"event: content_block_start",
		`data: {"type":"content_block_start","index":1}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"second"}}`,
		"",
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":1}`,
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
	firstIdx := strings.Index(got, `"content":"first"`)
	secondIdx := strings.Index(got, `"content":"second"`)
	if firstIdx < 0 || secondIdx < 0 {
		t.Errorf("multi-block deltas missing: %s", got)
	}
	if firstIdx > secondIdx {
		t.Errorf("multi-block deltas out of order: %s", got)
	}
	if !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing [DONE]: %s", got)
	}
}

func TestAnthropicAdapter_Transcode_StopWithoutMessageDelta_Good(t *testing.T) {
	a := api.AnthropicAdapter()
	stream := strings.Join([]string{
		"event: content_block_delta",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi"}}`,
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
	if !strings.Contains(got, `"finish_reason":"stop"`) {
		t.Errorf("message_stop without message_delta should default finish_reason to stop: %s", got)
	}
	if !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing [DONE]: %s", got)
	}
}
