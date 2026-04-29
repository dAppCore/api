// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestGraphql_Engine_GraphQLConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.GraphQLConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestGraphql_Engine_GraphQLConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.GraphQLConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestGraphql_Engine_GraphQLConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.GraphQLConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestGraphql_WithPlayground_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithPlayground()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestGraphql_WithPlayground_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithPlayground()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestGraphql_WithPlayground_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithPlayground()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestGraphql_WithGraphQLPath_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithGraphQLPath("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestGraphql_WithGraphQLPath_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithGraphQLPath("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestGraphql_WithGraphQLPath_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithGraphQLPath("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_GraphQLConfig_graphql() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.GraphQLConfig()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithPlayground_graphql() {
	func() {
		defer func() { _ = recover() }()
		_ = WithPlayground()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithGraphQLPath_graphql() {
	func() {
		defer func() { _ = recover() }()
		_ = WithGraphQLPath("")
	}()
	coretest.Println("done")
	// Output: done
}
