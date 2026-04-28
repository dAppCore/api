// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io"       // Note: AX-6 - request body reads and resets are HTTP stream boundaries.
	"net/http" // Note: AX-6 - transformer middleware emits HTTP status codes.

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

type transformerRouteConfig struct {
	in                []compiledTransformer
	out               []compiledTransformer
	requestValidator  *toolInputValidator
	responseValidator *toolInputValidator
}

func transformerMiddlewareForGroup(group RouteGroup) gin.HandlerFunc {
	if _, ok := group.(*ToolBridge); ok {
		return nil
	}

	routes := transformerRoutesForGroup(group)
	if len(routes) == 0 {
		return nil
	}

	return func(c *gin.Context) {
		cfg := routes[transformerRequestKey(c)]
		if cfg == nil {
			c.Next()
			return
		}

		if !cfg.applyRequest(c) {
			return
		}
		if len(cfg.out) == 0 && cfg.responseValidator == nil {
			c.Next()
			return
		}

		recorder := newToolResponseRecorder(c.Writer)
		c.Writer = recorder
		c.Next()
		cfg.applyResponse(c, recorder)
	}
}

func transformerRoutesForGroup(group RouteGroup) map[string]*transformerRouteConfig {
	descIter := routeDescriptions(group)
	if descIter == nil {
		return nil
	}

	routes := make(map[string]*transformerRouteConfig)
	for rd := range descIter {
		resolved := resolveRouteDescription(rd)
		cfg := transformerRouteConfigForDescription(resolved)
		if cfg == nil {
			continue
		}
		path := joinTransformerRoutePath(group.BasePath(), resolved.Path)
		routes[transformerRouteKey(resolved.Method, path)] = cfg
	}
	if len(routes) == 0 {
		return nil
	}
	return routes
}

func transformerRouteConfigForDescription(rd RouteDescription) *transformerRouteConfig {
	in, err := compileTransformerPipeline(transformerDirectionIn, rd.TransformerIn)
	if err != nil {
		panic(err)
	}
	out, err := compileTransformerPipeline(transformerDirectionOut, rd.TransformerOut)
	if err != nil {
		panic(err)
	}
	if len(in) == 0 && len(out) == 0 {
		return nil
	}

	cfg := &transformerRouteConfig{
		in:  in,
		out: out,
	}
	if len(in) > 0 {
		cfg.requestValidator = newToolInputValidator(rd.RequestBody)
	}
	if len(out) > 0 {
		cfg.responseValidator = newToolInputValidator(rd.Response)
	}
	return cfg
}

func transformerRequestKey(c *gin.Context) string {
	path := c.FullPath()
	if core.Trim(path) == "" && c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	return transformerRouteKey(c.Request.Method, path)
}

func (cfg *transformerRouteConfig) applyRequest(c *gin.Context) bool {
	if len(cfg.in) == 0 {
		return true
	}

	body, ok := readTransformerRequestBody(c)
	if !ok {
		return false
	}

	if cfg.requestValidator != nil {
		if err := cfg.requestValidator.Validate(body); err != nil {
			abortTransformerRequest(c, http.StatusBadRequest, "Request body does not match the declared route schema", err)
			return false
		}
	}

	transformed, err := runTransformerPipeline(c, body, cfg.in)
	if err != nil {
		abortTransformerRequest(c, http.StatusBadRequest, "Request body could not be transformed", err)
		return false
	}

	setTransformerRequestBody(c, transformed)
	return true
}

func wrapTransformerInHandler(handler gin.HandlerFunc, pipeline []compiledTransformer) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, ok := readTransformerRequestBody(c)
		if !ok {
			return
		}

		transformed, err := runTransformerPipeline(c, body, pipeline)
		if err != nil {
			abortTransformerRequest(c, http.StatusBadRequest, "Request body could not be transformed", err)
			return
		}

		setTransformerRequestBody(c, transformed)
		handler(c)
	}
}

func readTransformerRequestBody(c *gin.Context) ([]byte, bool) {
	limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxToolRequestBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		status := http.StatusBadRequest
		msg := "Unable to read request body"
		if err.Error() == "http: request body too large" {
			status = http.StatusRequestEntityTooLarge
			msg = "Request body exceeds the maximum allowed size"
		}
		abortTransformerRequest(c, status, msg, err)
		return nil, false
	}

	if core.Trim(string(body)) == "" {
		abortTransformerRequest(c, http.StatusBadRequest, "Request body is required", nil)
		return nil, false
	}

	return body, true
}

func setTransformerRequestBody(c *gin.Context, body []byte) {
	c.Request.Body = io.NopCloser(core.NewBuffer(body))
	c.Request.ContentLength = int64(len(body))
}

func abortTransformerRequest(c *gin.Context, status int, message string, err error) {
	details := map[string]any{}
	if err != nil {
		details["error"] = err.Error()
	}
	c.AbortWithStatusJSON(status, FailWithDetails(
		"invalid_request_body",
		message,
		details,
	))
}
