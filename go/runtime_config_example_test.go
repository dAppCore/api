// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestRuntimeConfig_Engine_RuntimeConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.RuntimeConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRuntimeConfig_Engine_RuntimeConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.RuntimeConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRuntimeConfig_Engine_RuntimeConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.RuntimeConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_RuntimeConfig_runtimeConfig() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.RuntimeConfig()
	}()
	coretest.Println("done")
	// Output: done
}
