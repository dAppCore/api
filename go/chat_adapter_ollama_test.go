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
