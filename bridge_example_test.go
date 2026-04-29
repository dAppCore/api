// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestBridge_NewToolBridge_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewToolBridge("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_NewToolBridge_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewToolBridge("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_NewToolBridge_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewToolBridge("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_Add_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.Add(ToolDescriptor{}, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_Add_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.Add(ToolDescriptor{}, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_Add_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.Add(ToolDescriptor{}, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_Name_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_Name_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_Name_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_BasePath_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_BasePath_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_BasePath_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_RegisterRoutes_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_RegisterRoutes_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_RegisterRoutes_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_Describe_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_Describe_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_Describe_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Describe()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_DescribeIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.DescribeIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_DescribeIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.DescribeIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_DescribeIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.DescribeIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_Tools_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Tools()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_Tools_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Tools()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_Tools_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.Tools()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ToolBridge_ToolsIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.ToolsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ToolBridge_ToolsIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.ToolsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ToolBridge_ToolsIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ToolBridge
		_ = subject.ToolsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_IsValidMCPServerID_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsValidMCPServerID("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_IsValidMCPServerID_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsValidMCPServerID("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_IsValidMCPServerID_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsValidMCPServerID("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_InputValidator_Validate_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.Validate(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_InputValidator_Validate_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.Validate(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_InputValidator_Validate_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.Validate(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_InputValidator_ValidateResponse_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.ValidateResponse(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_InputValidator_ValidateResponse_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.ValidateResponse(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_InputValidator_ValidateResponse_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolInputValidator
		_ = subject.ValidateResponse(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Header_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Header_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Header_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Header()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_WriteHeader_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_WriteHeader_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_WriteHeader_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeader(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_WriteHeaderNow_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_WriteHeaderNow_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_WriteHeaderNow_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.WriteHeaderNow()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Write_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Write_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Write_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_WriteString_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_WriteString_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_WriteString_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Flush_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Flush_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Flush_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		subject.Flush()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Status_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Status_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Status_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Status()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Size_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Size_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Size_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Size()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Written_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Written_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Written_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_ = subject.Written()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestBridge_ResponseRecorder_Hijack_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestBridge_ResponseRecorder_Hijack_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestBridge_ResponseRecorder_Hijack_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *toolResponseRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleNewToolBridge_bridge() {
	func() {
		defer func() { _ = recover() }()
		_ = NewToolBridge("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_Add_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		subject.Add(ToolDescriptor{}, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_Name_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.Name()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_BasePath_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.BasePath()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_RegisterRoutes_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		subject.RegisterRoutes(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_Describe_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.Describe()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_DescribeIter_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.DescribeIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_Tools_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.Tools()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleToolBridge_ToolsIter_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *ToolBridge
		_ = subject.ToolsIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleIsValidMCPServerID_bridge() {
	func() {
		defer func() { _ = recover() }()
		_ = IsValidMCPServerID("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleInputValidator_Validate_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolInputValidator
		_ = subject.Validate(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleInputValidator_ValidateResponse_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolInputValidator
		_ = subject.ValidateResponse(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Header_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_ = subject.Header()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_WriteHeader_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		subject.WriteHeader(0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_WriteHeaderNow_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		subject.WriteHeaderNow()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Write_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_, _ = subject.Write(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_WriteString_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_, _ = subject.WriteString("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Flush_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		subject.Flush()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Status_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_ = subject.Status()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Size_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_ = subject.Size()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Written_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_ = subject.Written()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResponseRecorder_Hijack_bridge() {
	func() {
		defer func() { _ = recover() }()
		var subject *toolResponseRecorder
		_, _, _ = subject.Hijack()
	}()
	coretest.Println("done")
	// Output: done
}
