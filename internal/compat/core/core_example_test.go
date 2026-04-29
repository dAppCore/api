// SPDX-License-Identifier: EUPL-1.2

package core

import coretest "dappco.re/go"

func ExampleNewRegistry_core() {
	func() {
		defer func() { _ = recover() }()
		_ = NewRegistry[any]()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewServiceRuntime_core() {
	func() {
		defer func() { _ = recover() }()
		_ = NewServiceRuntime[any](nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}
