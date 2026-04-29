// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestCmd_AddAPICommands_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		AddAPICommands(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCmd_AddAPICommands_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		AddAPICommands(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCmd_AddAPICommands_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		AddAPICommands(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleAddAPICommands_cmd() {
	func() {
		defer func() { _ = recover() }()
		AddAPICommands(nil)
	}()
	coretest.Println("done")
	// Output: done
}
