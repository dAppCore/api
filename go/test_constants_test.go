// SPDX-License-Identifier: EUPL-1.2

package api_test

// Shared test constants to avoid string literal duplication across test files.

const (
	fmtTestUnexpectedErr  = "unexpected error: %v"
	fmtTestExpected200    = "expected 200, got %d"
	fmtTestExpected400    = "expected 400, got %d"
	fmtTestUnmarshalErr   = "unmarshal error: %v"
	fmtTestRequestFailed  = "request failed: %v"
	fmtTestInvalidJSON    = "invalid JSON: %v"
	fmtTestFailedReadBody = "failed to read body: %v"
	fmtTestExpectedSuc    = "expected Success=true"
	fmtTestExpectedFail   = "expected Success=false"
	fmtTestExpectedData   = "expected Data=%q, got %q"
	fmtTestExpectedName   = "expected Name=%q, got %q"
	fmtTestExpectedTags   = "expected tags array, got %T"

	hdrContentType   = "Content-Type"
	hdrContentEnc    = "Content-Encoding"
	hdrContentDisp   = "Content-Disposition"
	hdrAcceptEnc     = "Accept-Encoding"
	hdrCacheControl  = "Cache-Control"
	hdrXRequestID    = "X-Request-ID"
	hdrXFrameOptions = "X-Frame-Options"
	hdrXCache        = "X-Cache"

	mimeJSON        = "application/json"
	mimeEventStream = "text/event-stream"

	pathHealth      = "/health"
	pathStubPing    = "/stub/ping"
	pathEvents      = "/events"
	pathChatComplet = "/v1/chat/completions"
	pathOpenAPIJSON = "/v1/openapi.json"
	pathDebugVars   = "/debug/vars"
	pathDebugPprof  = "/debug/pprof"
	pathGraphQL     = "/graphql"
	pathGraphQLPlay = "/graphql/playground"
	pathPublic      = "/public"
	pathStatus      = "/status"
	pathSwaggerDoc  = "/swagger/doc.json"

	apiKeyHeader = "X-API-Key"
	apiBaseURL   = "https://api.example.com"

	hdrRateLimit     = "X-RateLimit-Limit"
	hdrRateRemaining = "X-RateLimit-Remaining"
	hdrRateReset     = "X-RateLimit-Reset"
)
