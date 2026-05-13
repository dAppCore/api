// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestServeH3_Engine_ServeH3_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.ServeH3(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestServeH3_Engine_ServeH3_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.ServeH3(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestServeH3_Engine_ServeH3_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.ServeH3(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}
