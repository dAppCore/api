// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bufio"
	"encoding/json"
	"io"

	core "dappco.re/go"
)

type ollamaAdapter struct{}

// OllamaAdapter maps OpenAI chat completions to/from Ollama's native /api/chat
// (JSON request with an "options" block; newline-delimited JSON stream).
func OllamaAdapter() ChatFormatAdapter { return ollamaAdapter{} }

func (ollamaAdapter) Name() string         { return "ollama" }
func (ollamaAdapter) UpstreamPath() string { return "/api/chat" }

func (ollamaAdapter) BuildRequest(req ChatCompletionRequest) ([]byte, map[string]string, error) {
	msgs := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	options := map[string]any{}
	if req.Temperature != nil {
		options["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		options["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		options["top_k"] = *req.TopK
	}
	if req.MaxTokens != nil {
		options["num_predict"] = *req.MaxTokens
	}
	// Ollama's native /api/chat reads stop sequences inside the options block; a
	// top-level "stop" is silently ignored.
	if len(req.Stop) > 0 {
		options["stop"] = []string(req.Stop)
	}
	body := map[string]any{
		"model":    req.Model,
		"messages": msgs,
		"stream":   req.Stream,
	}
	if len(options) > 0 {
		body["options"] = options
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, nil, core.E("ollama", "marshal request", err)
	}
	return raw, map[string]string{"Content-Type": "application/json"}, nil
}

type ollamaResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done            bool   `json:"done"`
	DoneReason      string `json:"done_reason"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func ollamaFinish(doneReason string) string {
	if doneReason == "length" {
		return "length"
	}
	return "stop"
}

func (ollamaAdapter) DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error) {
	var or ollamaResponse
	if err := json.Unmarshal(upstream, &or); err != nil {
		return ChatCompletionResponse{}, core.E("ollama", "decode response", err)
	}
	return ChatCompletionResponse{
		ID:      newChatCompletionID(),
		Object:  "chat.completion",
		Model:   model,
		Choices: []ChatChoice{{Index: 0, Message: ChatMessage{Role: "assistant", Content: or.Message.Content}, FinishReason: ollamaFinish(or.DoneReason)}},
		Usage:   ChatUsage{PromptTokens: or.PromptEvalCount, CompletionTokens: or.EvalCount, TotalTokens: or.PromptEvalCount + or.EvalCount},
	}, nil
}

func (ollamaAdapter) Transcoder() ChatStreamTranscoder { return ollamaTranscoder{} }

type ollamaTranscoder struct{}

func (ollamaTranscoder) Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error {
	scanner := bufio.NewScanner(upstream)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := core.Trim(scanner.Text())
		if line == "" {
			continue
		}
		var or ollamaResponse
		if err := json.Unmarshal([]byte(line), &or); err != nil {
			continue // skip malformed line
		}
		if or.Done {
			if or.Message.Content != "" || first {
				delta := ChatMessageDelta{Content: or.Message.Content}
				if first {
					delta.Role = "assistant"
					first = false
				}
				writeChatChunk(w, flush, ChatCompletionChunk{
					ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
					Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: nil}},
				})
			}
			fr := ollamaFinish(or.DoneReason)
			writeChatChunk(w, flush, ChatCompletionChunk{
				ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
				Choices: []ChatChunkChoice{{Index: 0, Delta: ChatMessageDelta{}, FinishReason: &fr}},
			})
			break
		}
		delta := ChatMessageDelta{Content: or.Message.Content}
		if first {
			delta.Role = "assistant"
			first = false
		}
		writeChatChunk(w, flush, ChatCompletionChunk{
			ID: meta.ID, Object: "chat.completion.chunk", Created: meta.Created, Model: meta.Model,
			Choices: []ChatChunkChoice{{Index: 0, Delta: delta, FinishReason: nil}},
		})
	}
	if err := scanner.Err(); err != nil {
		return err // truncated stream signals incomplete; do NOT emit [DONE]
	}
	writeSSEDone(w, flush)
	return nil
}
