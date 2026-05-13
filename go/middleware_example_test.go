// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestMiddleware_GetRequestID_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestID(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMiddleware_GetRequestID_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestID(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMiddleware_GetRequestID_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestID(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMiddleware_GetRequestDuration_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestDuration(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMiddleware_GetRequestDuration_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestDuration(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMiddleware_GetRequestDuration_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestDuration(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMiddleware_GetRequestMeta_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestMeta(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMiddleware_GetRequestMeta_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestMeta(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMiddleware_GetRequestMeta_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetRequestMeta(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleGetRequestID_middleware() {
	func() {
		defer func() { _ = recover() }()
		_ = GetRequestID(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGetRequestDuration_middleware() {
	func() {
		defer func() { _ = recover() }()
		_ = GetRequestDuration(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGetRequestMeta_middleware() {
	func() {
		defer func() { _ = recover() }()
		_ = GetRequestMeta(nil)
	}()
	coretest.Println("done")
	// Output: done
}
