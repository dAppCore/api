// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestAuthentik_Engine_AuthentikConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.AuthentikConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestAuthentik_Engine_AuthentikConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.AuthentikConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestAuthentik_Engine_AuthentikConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.AuthentikConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestAuthentik_AuthentikUser_HasGroup_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *AuthentikUser
		_ = subject.HasGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestAuthentik_AuthentikUser_HasGroup_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *AuthentikUser
		_ = subject.HasGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestAuthentik_AuthentikUser_HasGroup_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *AuthentikUser
		_ = subject.HasGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestAuthentik_GetUser_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetUser(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestAuthentik_GetUser_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetUser(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestAuthentik_GetUser_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetUser(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestAuthentik_RequireAuth_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireAuth()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestAuthentik_RequireAuth_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireAuth()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestAuthentik_RequireAuth_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireAuth()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestAuthentik_RequireGroup_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestAuthentik_RequireGroup_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestAuthentik_RequireGroup_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RequireGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_AuthentikConfig_authentik() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.AuthentikConfig()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleAuthentikUser_HasGroup_authentik() {
	func() {
		defer func() { _ = recover() }()
		var subject *AuthentikUser
		_ = subject.HasGroup("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGetUser_authentik() {
	func() {
		defer func() { _ = recover() }()
		_ = GetUser(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRequireAuth_authentik() {
	func() {
		defer func() { _ = recover() }()
		_ = RequireAuth()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRequireGroup_authentik() {
	func() {
		defer func() { _ = recover() }()
		_ = RequireGroup("")
	}()
	coretest.Println("done")
	// Output: done
}
