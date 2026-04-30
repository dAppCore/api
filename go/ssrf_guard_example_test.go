// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func ExampleURLError_Error_ssrfGuard() {
	func() {
		defer func() { _ = recover() }()
		var subject blockedURLError
		_ = subject.Error()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleURLError_Unwrap_ssrfGuard() {
	func() {
		defer func() { _ = recover() }()
		var subject blockedURLError
		_ = subject.Unwrap()
	}()
	coretest.Println("done")
	// Output: done
}
