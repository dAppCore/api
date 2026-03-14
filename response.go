// SPDX-License-Identifier: EUPL-1.2

package api

// Response is the standard envelope for all API responses.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error describes a failed API request.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta carries pagination and request metadata.
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Total     int    `json:"total,omitempty"`
}

// OK wraps data in a successful response envelope.
func OK[T any](data T) Response[T] {
	return Response[T]{
		Success: true,
		Data:    data,
	}
}

// Fail creates an error response with the given code and message.
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
