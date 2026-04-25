// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json" // Note: AX-6 - json.RawMessage preserves envelope data while remapping DTOs.
	"net/http"      // Note: AX-6 - transformer response wrappers emit HTTP status codes.

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

func (cfg *transformerRouteConfig) applyResponse(c *gin.Context, recorder *toolResponseRecorder) {
	if recorder.Status() < 200 || recorder.Status() >= 300 {
		recorder.commit()
		return
	}

	body, err := transformResponseEnvelope(c, recorder.body, cfg.out)
	if err != nil {
		writeTransformerResponseError(recorder, "Response body could not be transformed", err)
		return
	}
	if cfg.responseValidator != nil {
		if err := cfg.responseValidator.ValidateResponse(body); err != nil {
			writeTransformerResponseError(recorder, "Response body does not match the declared route schema", err)
			return
		}
	}

	recorder.body = body
	recorder.headers.Del("Content-Length")
	recorder.commit()
}

func wrapTransformerOutHandler(handler gin.HandlerFunc, pipeline []compiledTransformer) gin.HandlerFunc {
	return func(c *gin.Context) {
		recorder := newToolResponseRecorder(c.Writer)
		c.Writer = recorder

		handler(c)

		if recorder.Status() >= 200 && recorder.Status() < 300 {
			body, err := transformResponseEnvelope(c, recorder.body, pipeline)
			if err != nil {
				writeTransformerResponseError(recorder, "Response body could not be transformed", err)
				return
			}
			recorder.body = body
			recorder.headers.Del("Content-Length")
		}

		recorder.commit()
	}
}

func transformResponseEnvelope(c *gin.Context, body []byte, pipeline []compiledTransformer) ([]byte, error) {
	if len(pipeline) == 0 || core.Trim(string(body)) == "" {
		return body, nil
	}

	var envelope map[string]json.RawMessage
	if err := unmarshalEnvelope(body, &envelope); err != nil {
		return nil, err
	}

	rawSuccess, ok := envelope["success"]
	if !ok {
		return nil, core.E("TransformerOut", "response is missing a success field", nil)
	}

	var success bool
	if err := unmarshalEnvelope(rawSuccess, &success); err != nil {
		return nil, core.E("TransformerOut", "decode success field", err)
	}
	if !success {
		return body, nil
	}

	rawData, ok := envelope["data"]
	if !ok {
		return body, nil
	}

	transformed, err := runTransformerPipeline(c, rawData, pipeline)
	if err != nil {
		return nil, err
	}
	envelope["data"] = json.RawMessage(transformed)

	encoded := core.JSONMarshal(envelope)
	if !encoded.OK {
		if err, ok := encoded.Value.(error); ok {
			return nil, core.E("TransformerOut", "encode response envelope", err)
		}
		return nil, core.E("TransformerOut", "encode response envelope", nil)
	}

	data, ok := encoded.Value.([]byte)
	if !ok {
		return nil, core.E("TransformerOut", "encoded response envelope was not bytes", nil)
	}
	return data, nil
}

func unmarshalEnvelope(data []byte, target any) error {
	result := core.JSONUnmarshal(data, target)
	if result.OK {
		return nil
	}
	if err, ok := result.Value.(error); ok {
		return err
	}
	return core.E("TransformerOut", "decode response envelope", nil)
}

func writeTransformerResponseError(recorder *toolResponseRecorder, message string, err error) {
	recorder.reset()
	recorder.writeErrorResponse(http.StatusInternalServerError, FailWithDetails(
		"invalid_response_body",
		message,
		map[string]any{"error": err.Error()},
	))
}
