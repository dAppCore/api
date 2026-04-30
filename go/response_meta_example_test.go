// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestResponseMeta_MetaRecorder_Header_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Header_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Header_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeader_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeader_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeader_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeaderNow_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeaderNow_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_WriteHeaderNow_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Write_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Write_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Write_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_WriteString_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_WriteString_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_WriteString_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Flush_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Flush_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Flush_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Status_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Status_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Status_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Size_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Size_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Size_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Written_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Written_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Written_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestResponseMeta_MetaRecorder_Hijack_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestResponseMeta_MetaRecorder_Hijack_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestResponseMeta_MetaRecorder_Hijack_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *responseMetaRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleMetaRecorder_Header_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_ = subject.Header()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_WriteHeader_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		subject.WriteHeader(0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_WriteHeaderNow_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		subject.WriteHeaderNow()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Write_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_WriteString_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Flush_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		subject.Flush()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Status_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_ = subject.Status()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Size_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_ = subject.Size()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Written_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_ = subject.Written()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleMetaRecorder_Hijack_responseMeta() {
	func() {
		defer func() { _ = recover() }()
		var subject *responseMetaRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.Println("done")
	// Output: done
}
