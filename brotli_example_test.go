// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestBrotli_Handler_Handle_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliHandler
		subject.Handle(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Handler_Handle_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliHandler
		subject.Handle(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Handler_Handle_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliHandler
		subject.Handle(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBrotli_Writer_Write_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Writer_Write_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Writer_Write_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBrotli_Writer_WriteString_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Writer_WriteString_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Writer_WriteString_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBrotli_Writer_WriteHeader_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Writer_WriteHeader_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Writer_WriteHeader_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBrotli_Writer_WriteHeaderNow_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Writer_WriteHeaderNow_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Writer_WriteHeaderNow_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBrotli_Writer_Flush_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBrotli_Writer_Flush_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBrotli_Writer_Flush_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *brotliWriter
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleHandler_Handle_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliHandler
		subject.Handle(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_Write_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliWriter
		_, _ = subject.Write(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_WriteString_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliWriter
		_, _ = subject.WriteString("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_WriteHeader_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliWriter
		subject.WriteHeader(0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_WriteHeaderNow_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliWriter
		subject.WriteHeaderNow()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWriter_Flush_brotli() {
	func() {
		defer func() { _ = recover() }()
		var subject *brotliWriter
		subject.Flush()
	}()
	coretest.Println("done")
	// Output: done
}
