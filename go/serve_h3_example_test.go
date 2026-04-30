// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func ExampleEngine_ServeH3_serveH3() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.ServeH3(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}
