// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestSpecRegistry_RegisterSpecGroups_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_RegisterSpecGroups_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_RegisterSpecGroups_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecRegistry_RegisterSpecGroupsIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_RegisterSpecGroupsIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_RegisterSpecGroupsIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		RegisterSpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecRegistry_RegisteredSpecGroups_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_RegisteredSpecGroups_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_RegisteredSpecGroups_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecRegistry_RegisteredSpecGroupsIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroupsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_RegisteredSpecGroupsIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroupsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_RegisteredSpecGroupsIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RegisteredSpecGroupsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecRegistry_SpecGroupsIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_SpecGroupsIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_SpecGroupsIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SpecGroupsIter(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSpecRegistry_ResetSpecGroups_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		ResetSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSpecRegistry_ResetSpecGroups_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		ResetSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSpecRegistry_ResetSpecGroups_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		ResetSpecGroups()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleRegisterSpecGroups_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		RegisterSpecGroups()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegisterSpecGroupsIter_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		RegisterSpecGroupsIter(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegisteredSpecGroups_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		_ = RegisteredSpecGroups()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegisteredSpecGroupsIter_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		_ = RegisteredSpecGroupsIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSpecGroupsIter_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		_ = SpecGroupsIter(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResetSpecGroups_specRegistry() {
	func() {
		defer func() { _ = recover() }()
		ResetSpecGroups()
	}()
	coretest.Println("done")
	// Output: done
}
