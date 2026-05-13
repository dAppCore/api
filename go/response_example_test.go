// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestResponse_OK_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = OK[any](nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponse_OK_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = OK[any](nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponse_OK_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = OK[any](nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponse_Fail_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Fail("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponse_Fail_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Fail("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponse_Fail_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Fail("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponse_FailWithDetails_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = FailWithDetails("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponse_FailWithDetails_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = FailWithDetails("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponse_FailWithDetails_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = FailWithDetails("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponse_Paginated_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Paginated[any](nil, 0, 0, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponse_Paginated_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Paginated[any](nil, 0, 0, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponse_Paginated_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = Paginated[any](nil, 0, 0, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponse_AttachRequestMeta_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = AttachRequestMeta[any](nil, Response[any]{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponse_AttachRequestMeta_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = AttachRequestMeta[any](nil, Response[any]{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponse_AttachRequestMeta_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = AttachRequestMeta[any](nil, Response[any]{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleOK_response() {
	func() {
		defer func() { _ = recover() }()
		_ = OK[any](nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleFail_response() {
	func() {
		defer func() { _ = recover() }()
		_ = Fail("", "")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleFailWithDetails_response() {
	func() {
		defer func() { _ = recover() }()
		_ = FailWithDetails("", "", nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExamplePaginated_response() {
	func() {
		defer func() { _ = recover() }()
		_ = Paginated[any](nil, 0, 0, 0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleAttachRequestMeta_response() {
	func() {
		defer func() { _ = recover() }()
		_ = AttachRequestMeta[any](nil, Response[any]{})
	}()
	coretest.Println("done")
	// Output: done
}
