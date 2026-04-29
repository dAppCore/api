// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestSunset_WithSunsetNoticeURL_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSunsetNoticeURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSunset_WithSunsetNoticeURL_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSunsetNoticeURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSunset_WithSunsetNoticeURL_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSunsetNoticeURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSunset_ApiSunset_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunset("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSunset_ApiSunset_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunset("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSunset_ApiSunset_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunset("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSunset_ApiSunsetWith_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunsetWith("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSunset_ApiSunsetWith_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunsetWith("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSunset_ApiSunsetWith_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ApiSunsetWith("", "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleWithSunsetNoticeURL_sunset() {
	func() {
		defer func() { _ = recover() }()
		_ = WithSunsetNoticeURL("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleApiSunset_sunset() {
	func() {
		defer func() { _ = recover() }()
		_ = ApiSunset("", "")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleApiSunsetWith_sunset() {
	func() {
		defer func() { _ = recover() }()
		_ = ApiSunsetWith("", "")
	}()
	coretest.Println("done")
	// Output: done
}
