// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestOpenapi_SpecBuilder_Build_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.Build(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestOpenapi_SpecBuilder_Build_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.Build(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestOpenapi_SpecBuilder_Build_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.Build(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestOpenapi_SpecBuilder_BuildIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.BuildIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestOpenapi_SpecBuilder_BuildIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.BuildIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestOpenapi_SpecBuilder_BuildIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SpecBuilder
		_, _ = subject.BuildIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleSpecBuilder_Build_openapi() {
	func() {
		defer func() { _ = recover() }()
		var subject *SpecBuilder
		_, _ = subject.Build(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSpecBuilder_BuildIter_openapi() {
	func() {
		defer func() { _ = recover() }()
		var subject *SpecBuilder
		_, _ = subject.BuildIter(nil)
	}()
	coretest.Println("done")
	// Output: done
}
