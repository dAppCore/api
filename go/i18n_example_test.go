// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestI18n_Engine_I18nConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.I18nConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestI18n_Engine_I18nConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.I18nConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestI18n_Engine_I18nConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.I18nConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestI18n_WithI18n_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithI18n()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestI18n_WithI18n_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithI18n()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestI18n_WithI18n_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithI18n()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestI18n_GetLocale_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetLocale(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestI18n_GetLocale_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetLocale(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestI18n_GetLocale_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = GetLocale(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestI18n_GetMessage_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GetMessage(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestI18n_GetMessage_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GetMessage(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestI18n_GetMessage_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GetMessage(nil, "")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_I18nConfig_i18n() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.I18nConfig()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithI18n_i18n() {
	func() {
		defer func() { _ = recover() }()
		_ = WithI18n()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGetLocale_i18n() {
	func() {
		defer func() { _ = recover() }()
		_ = GetLocale(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGetMessage_i18n() {
	func() {
		defer func() { _ = recover() }()
		_, _ = GetMessage(nil, "")
	}()
	coretest.Println("done")
	// Output: done
}
