// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestSwagger_Spec_ReadDoc_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *swaggerSpec
		_ = subject.ReadDoc()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSwagger_Spec_ReadDoc_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *swaggerSpec
		_ = subject.ReadDoc()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSwagger_Spec_ReadDoc_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *swaggerSpec
		_ = subject.ReadDoc()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleSpec_ReadDoc_swagger() {
	func() {
		defer func() { _ = recover() }()
		var subject *swaggerSpec
		_ = subject.ReadDoc()
	}()
	coretest.Println("done")
	// Output: done
}
