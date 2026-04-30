// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestSpecBuilderHelper_Engine_OpenAPISpecBuilder_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.OpenAPISpecBuilder()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecBuilderHelper_Engine_OpenAPISpecBuilder_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.OpenAPISpecBuilder()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecBuilderHelper_Engine_OpenAPISpecBuilder_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.OpenAPISpecBuilder()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecBuilderHelper_Engine_SwaggerConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.SwaggerConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecBuilderHelper_Engine_SwaggerConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.SwaggerConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecBuilderHelper_Engine_SwaggerConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.SwaggerConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_OpenAPISpecBuilder_specBuilderHelper() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.OpenAPISpecBuilder()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleEngine_SwaggerConfig_specBuilderHelper() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.SwaggerConfig()
	}()
	coretest.Println("done")
	// Output: done
}
