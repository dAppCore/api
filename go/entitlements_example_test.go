// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestEntitlements_NewEntitlementBridge_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewEntitlementBridge(EntitlementBridgeConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestEntitlements_NewEntitlementBridge_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewEntitlementBridge(EntitlementBridgeConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestEntitlements_NewEntitlementBridge_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewEntitlementBridge(EntitlementBridgeConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestEntitlements_EntitlementBridge_Check_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_, _ = subject.Check(nil, "", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestEntitlements_EntitlementBridge_Check_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_, _ = subject.Check(nil, "", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestEntitlements_EntitlementBridge_Check_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_, _ = subject.Check(nil, "", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestEntitlements_EntitlementBridge_Callback_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.Callback(nil, "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestEntitlements_EntitlementBridge_Callback_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.Callback(nil, "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestEntitlements_EntitlementBridge_Callback_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.Callback(nil, "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForRequest_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForRequest(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForRequest_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForRequest(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForRequest_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForRequest(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForGin_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForGin(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForGin_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForGin(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestEntitlements_EntitlementBridge_CallbackForGin_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *EntitlementBridge
		_ = subject.CallbackForGin(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleNewEntitlementBridge_entitlements() {
	func() {
		defer func() { _ = recover() }()
		_ = NewEntitlementBridge(EntitlementBridgeConfig{})
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleEntitlementBridge_Check_entitlements() {
	func() {
		defer func() { _ = recover() }()
		var subject *EntitlementBridge
		_, _ = subject.Check(nil, "", "", nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleEntitlementBridge_Callback_entitlements() {
	func() {
		defer func() { _ = recover() }()
		var subject *EntitlementBridge
		_ = subject.Callback(nil, "", nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleEntitlementBridge_CallbackForRequest_entitlements() {
	func() {
		defer func() { _ = recover() }()
		var subject *EntitlementBridge
		_ = subject.CallbackForRequest(nil, "")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleEntitlementBridge_CallbackForGin_entitlements() {
	func() {
		defer func() { _ = recover() }()
		var subject *EntitlementBridge
		_ = subject.CallbackForGin(nil, "")
	}()
	coretest.Println("done")
	// Output: done
}
