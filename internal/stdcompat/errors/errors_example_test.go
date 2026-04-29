// SPDX-License-Identifier: EUPL-1.2

package errors

import coretest "dappco.re/go"

func ExampleNew() {
	coretest.Println(New("api failed").Error())
	// Output: api failed
}

func ExampleIs() {
	target := New("sentinel")
	err := coretest.Wrap(target, "api.Run", "failed")
	coretest.Println(Is(err, target))
	// Output: true
}

func ExampleAs() {
	err := coretest.E("api.Run", "failed", nil)
	var target *coretest.Err
	coretest.Println(As(err, &target), target.Operation)
	// Output: true api.Run
}

func ExampleJoin() {
	err := Join(New("one"), nil)
	coretest.Println(err.Error())
	// Output: one
}

func ExampleUnwrap() {
	cause := New("cause")
	err := coretest.Wrap(cause, "api.Run", "failed")
	coretest.Println(Unwrap(err).Error())
	// Output: cause
}
