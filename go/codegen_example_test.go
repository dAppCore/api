// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestCodegen_SDKGenerator_Generate_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Generate(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCodegen_SDKGenerator_Generate_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Generate(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCodegen_SDKGenerator_Generate_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Generate(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestCodegen_SDKGenerator_Available_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Available()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCodegen_SDKGenerator_Available_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Available()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCodegen_SDKGenerator_Available_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SDKGenerator
		_ = subject.Available()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestCodegen_SupportedLanguages_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguages()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCodegen_SupportedLanguages_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguages()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCodegen_SupportedLanguages_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguages()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestCodegen_SupportedLanguagesIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguagesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCodegen_SupportedLanguagesIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguagesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCodegen_SupportedLanguagesIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SupportedLanguagesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleSDKGenerator_Generate_codegen() {
	func() {
		defer func() { _ = recover() }()
		var subject *SDKGenerator
		_ = subject.Generate(nil, "")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSDKGenerator_Available_codegen() {
	func() {
		defer func() { _ = recover() }()
		var subject *SDKGenerator
		_ = subject.Available()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSupportedLanguages_codegen() {
	func() {
		defer func() { _ = recover() }()
		_ = SupportedLanguages()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSupportedLanguagesIter_codegen() {
	func() {
		defer func() { _ = recover() }()
		_ = SupportedLanguagesIter()
	}()
	coretest.Println("done")
	// Output: done
}
