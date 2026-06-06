// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io" // Note: AX-6 — io.Writer/Reader are the transcoder stream boundary.

	core "dappco.re/go"
)

// ChatFormatAdapter maps between the OpenAI chat shape and a non-OpenAI upstream.
// OpenAI-compatible upstreams need NO adapter — passthrough is the default.
type ChatFormatAdapter interface {
	// Name identifies the adapter, e.g. "ollama", "anthropic".
	Name() string
	// UpstreamPath is the path under the upstream base URL, e.g. "/api/chat".
	UpstreamPath() string
	// BuildRequest maps the OpenAI request into the upstream body + protocol
	// headers (Content-Type, anthropic-version). Operator secrets (x-api-key)
	// belong in Upstream.Headers, not here.
	BuildRequest(req ChatCompletionRequest) (body []byte, headers map[string]string, err error)
	// DecodeResponse maps a complete (non-streaming) upstream body into the
	// OpenAI response.
	DecodeResponse(model string, upstream []byte) (ChatCompletionResponse, error)
	// Transcoder converts the upstream stream into OpenAI chunk SSE; nil means
	// the adapter supports non-streaming only.
	Transcoder() ChatStreamTranscoder
}

// ChatStreamTranscoder converts an upstream response stream into OpenAI
// chat.completion.chunk SSE events written to w (flushing via flush as it goes).
// It emits the terminating "data: [DONE]". Returns on upstream EOF or error.
type ChatStreamTranscoder interface {
	Transcode(w io.Writer, flush func(), upstream io.Reader, meta ChatStreamMeta) error
}

// ChatStreamMeta carries the OpenAI identity fields a transcoder stamps on every chunk.
type ChatStreamMeta struct {
	ID      string
	Model   string
	Created int64
}

// writeChatChunk marshals a chunk as one SSE "data:" event and flushes.
func writeChatChunk(w io.Writer, flush func(), chunk ChatCompletionChunk) {
	data := core.JSONMarshal(chunk)
	raw, ok := data.Value.([]byte)
	if !data.OK || !ok {
		return
	}
	_, _ = io.WriteString(w, "data: ")
	_, _ = w.Write(raw)
	_, _ = io.WriteString(w, "\n\n")
	if flush != nil {
		flush()
	}
}

// writeSSEDone emits the terminating sentinel.
func writeSSEDone(w io.Writer, flush func()) {
	_, _ = io.WriteString(w, "data: [DONE]\n\n")
	if flush != nil {
		flush()
	}
}
