// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestCache_Writer_Write_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCache_Writer_Write_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCache_Writer_Write_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestCache_Writer_WriteString_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCache_Writer_WriteString_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCache_Writer_WriteString_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *cacheWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleWriter_Write_cache() {
	func() {
		defer func() { _ = recover() }()
		var subject *cacheWriter
		_, _ = subject.Write(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_WriteString_cache() {
	func() {
		defer func() { _ = recover() }()
		var subject *cacheWriter
		_, _ = subject.WriteString("")
	}()
	coretest.Println("done")
	// Output: done
}
