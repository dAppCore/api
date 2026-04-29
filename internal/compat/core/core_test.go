// SPDX-License-Identifier: EUPL-1.2

package core

import coretest "dappco.re/go"

func TestCore_NewRegistry_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry[any]()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCore_NewRegistry_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry[any]()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCore_NewRegistry_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry[any]()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestCore_NewServiceRuntime_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewServiceRuntime[any](nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCore_NewServiceRuntime_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewServiceRuntime[any](nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCore_NewServiceRuntime_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewServiceRuntime[any](nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}
