// SPDX-License-Identifier: EUPL-1.2

package api

import "github.com/gin-gonic/gin"

// Response is the standard envelope for all API responses.
//
// Example:
//
//	resp := api.OK(map[string]any{"id": 42})
//	resp.Success // true
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error describes a failed API request.
//
// Example:
//
//	err := api.Error{Code: "invalid_input", Message: "Name is required"}
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta carries pagination and request metadata.
//
// Example:
//
//	meta := api.Meta{RequestID: "req_123", Duration: "12ms"}
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Total     int    `json:"total,omitempty"`
}

// OK wraps data in a successful response envelope.
//
// Example:
//
//	c.JSON(http.StatusOK, api.OK(map[string]any{"name": "status"}))
func OK[T any](data T) Response[T] {
	return Response[T]{
		Success: true,
		Data:    data,
	}
}

// Fail creates an error response with the given code and message.
//
// Example:
//
//	c.JSON(http.StatusBadRequest, api.Fail("invalid_input", "Name is required"))
func Fail(code, message string) Response[any] {
	return Response[any]{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// FailWithDetails creates an error response with additional detail payload.
//
// Example:
//
//	c.JSON(http.StatusBadRequest, api.FailWithDetails("invalid_input", "Name is required", map[string]any{"field": "name"}))
func FailWithDetails(code, message string, details any) Response[any] {
	return Response[any]{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// Paginated wraps data in a successful response with pagination metadata.
//
// Example:
//
//	c.JSON(http.StatusOK, api.Paginated(items, 2, 50, 200))
func Paginated[T any](data T, page, perPage, total int) Response[T] {
	return Response[T]{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
	}
}

// AttachRequestMeta merges request metadata into an existing response envelope.
// Existing pagination metadata is preserved; request_id and duration are added
// when available from the Gin context.
//
// Example:
//
//	resp = api.AttachRequestMeta(c, resp)
func AttachRequestMeta[T any](c *gin.Context, resp Response[T]) Response[T] {
	meta := GetRequestMeta(c)
	if meta == nil {
		return resp
	}

	if resp.Meta == nil {
		resp.Meta = meta
		return resp
	}

	if resp.Meta.RequestID == "" {
		resp.Meta.RequestID = meta.RequestID
	}
	if resp.Meta.Duration == "" {
		resp.Meta.Duration = meta.Duration
	}

	return resp
}
