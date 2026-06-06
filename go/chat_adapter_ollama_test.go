// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	api "dappco.re/go/api"
)

func TestOllamaAdapter_BuildRequest_Good(t *testing.T) {
	a := api.OllamaAdapter()
	mt := 64
	body, hdrs, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "llama3", Messages: []api.ChatMessage{{Role: "user", Content: "hi"}}, MaxTokens: &mt, Stream: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if hdrs["Content-Type"] != "application/json" {
		t.Errorf("missing content-type header")
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if got["model"] != "llama3" || got["stream"] != true {
		t.Errorf("bad ollama body: %s", body)
	}
	opts, _ := got["options"].(map[string]any)
	if opts["num_predict"].(float64) != 64 {
		t.Errorf("max_tokens not mapped to num_predict: %s", body)
	}
}

func TestOllamaAdapter_BuildRequest_Stop_Good(t *testing.T) {
	a := api.OllamaAdapter()
	body, _, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "llama3", Messages: []api.ChatMessage{{Role: "user", Content: "hi"}}, Stop: []string{"\n\n", "END"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	// Ollama reads stop INSIDE options; a top-level "stop" is silently ignored.
	if _, ok := got["stop"]; ok {
		t.Errorf("top-level stop key present (Ollama ignores it): %s", body)
	}
	opts, _ := got["options"].(map[string]any)
	stop, ok := opts["stop"].([]any)
	if !ok || len(stop) != 2 || stop[0] != "\n\n" || stop[1] != "END" {
		t.Errorf("stop not placed inside options: %s", body)
	}
}

func TestOllamaAdapter_DecodeResponse_Good(t *testing.T) {
	a := api.OllamaAdapter()
	out, err := a.DecodeResponse("llama3", []byte(`{"message":{"role":"assistant","content":"4"},"done":true,"done_reason":"stop","prompt_eval_count":3,"eval_count":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Choices[0].Message.Content != "4" || out.Choices[0].FinishReason != "stop" {
		t.Errorf("bad decode: %+v", out)
	}
	if out.Usage.PromptTokens != 3 || out.Usage.CompletionTokens != 1 {
		t.Errorf("bad usage: %+v", out.Usage)
	}
}

func TestOllamaAdapter_Transcode_Good(t *testing.T) {
	a := api.OllamaAdapter()
	stream := strings.Join([]string{
		`{"message":{"role":"assistant","content":"He"},"done":false}`,
		`{"message":{"role":"assistant","content":"llo"},"done":false}`,
		`{"message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}`,
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "llama3", Created: 1})
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

func TestOllamaAdapter_Transcode_EmptyStream_Good(t *testing.T) {
	a := api.OllamaAdapter()
	stream := `{"done":true,"done_reason":"stop"}`
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "llama3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"role":"assistant"`) {
		t.Errorf("missing role-priming chunk: %s", got)
	}
	if !strings.Contains(got, `"finish_reason":"stop"`) || !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing finish/[DONE]: %s", got)
	}
}

func TestOllamaAdapter_Transcode_MalformedLineSkipped_Ugly(t *testing.T) {
	a := api.OllamaAdapter()
	stream := strings.Join([]string{
		`{"message":{"role":"assistant","content":"He"},"done":false}`,
		`{not json`,
		`{"message":{"role":"assistant","content":"llo"},"done":false}`,
		`{"message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}`,
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "llama3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"He"`) || !strings.Contains(got, `"content":"llo"`) {
		t.Errorf("malformed line aborted stream, deltas missing: %s", got)
	}
	if !strings.Contains(got, "data: [DONE]") {
		t.Errorf("missing [DONE] after malformed line skip: %s", got)
	}
}

func TestOllamaAdapter_Transcode_DoneWithContent_Good(t *testing.T) {
	a := api.OllamaAdapter()
	stream := strings.Join([]string{
		`{"message":{"role":"assistant","content":"Hi"},"done":false}`,
		`{"message":{"role":"assistant","content":"!"},"done":true,"done_reason":"stop"}`,
	}, "\n")
	var buf bytes.Buffer
	err := a.Transcoder().Transcode(&buf, func() {}, strings.NewReader(stream), api.ChatStreamMeta{ID: "id", Model: "llama3", Created: 1})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"content":"!"`) {
		t.Errorf("trailing content on done line dropped: %s", got)
	}
	finishIdx := strings.Index(got, `"finish_reason":"stop"`)
	contentIdx := strings.Index(got, `"content":"!"`)
	if finishIdx < 0 || contentIdx < 0 || contentIdx > finishIdx {
		t.Errorf("trailing content must precede finish chunk: %s", got)
	}
}

func TestOllamaAdapter_DecodeResponse_Length_Good(t *testing.T) {
	a := api.OllamaAdapter()
	out, err := a.DecodeResponse("llama3", []byte(`{"message":{"role":"assistant","content":"x"},"done":true,"done_reason":"length"}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Choices[0].FinishReason != "length" {
		t.Errorf("done_reason length not mapped: %s", out.Choices[0].FinishReason)
	}
}

func TestOllamaAdapter_BuildRequest_NoOptions_Good(t *testing.T) {
	a := api.OllamaAdapter()
	body, _, err := a.BuildRequest(api.ChatCompletionRequest{
		Model: "llama3", Messages: []api.ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(body, &got)
	if _, ok := got["options"]; ok {
		t.Errorf("options key present when no sampling params set: %s", body)
	}
}
