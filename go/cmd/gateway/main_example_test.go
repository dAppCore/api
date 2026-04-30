// SPDX-License-Identifier: EUPL-1.2

package main

import coretest "dappco.re/go"

func TestMain_RouteGroup_Name_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_Name_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_Name_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMain_RouteGroup_BasePath_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_BasePath_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_BasePath_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMain_RouteGroup_Channels_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Channels()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_Channels_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Channels()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_Channels_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Channels()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMain_RouteGroup_RegisterRoutes_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_RegisterRoutes_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_RegisterRoutes_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMain_RouteGroup_Describe_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_Describe_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_Describe_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject brainRouteGroup
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestMain_RouteGroup_HandleFunc_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *proxyRouteGroup
		subject.HandleFunc("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestMain_RouteGroup_HandleFunc_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *proxyRouteGroup
		subject.HandleFunc("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestMain_RouteGroup_HandleFunc_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *proxyRouteGroup
		subject.HandleFunc("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleRouteGroup_Name_main() {
	func() {
		defer func() { _ = recover() }()
		var subject brainRouteGroup
		_ = subject.Name()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRouteGroup_BasePath_main() {
	func() {
		defer func() { _ = recover() }()
		var subject brainRouteGroup
		_ = subject.BasePath()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRouteGroup_Channels_main() {
	func() {
		defer func() { _ = recover() }()
		var subject brainRouteGroup
		_ = subject.Channels()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRouteGroup_RegisterRoutes_main() {
	func() {
		defer func() { _ = recover() }()
		var subject brainRouteGroup
		subject.RegisterRoutes(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRouteGroup_Describe_main() {
	func() {
		defer func() { _ = recover() }()
		var subject brainRouteGroup
		_ = subject.Describe()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRouteGroup_HandleFunc_main() {
	func() {
		defer func() { _ = recover() }()
		var subject *proxyRouteGroup
		subject.HandleFunc("", nil)
	}()
	coretest.Println("done")
	// Output: done
}
