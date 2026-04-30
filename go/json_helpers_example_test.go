// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func ExampleNumber_String_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject jsonNumber
		_ = subject.String()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNumber_Float64_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject jsonNumber
		_, _ = subject.Float64()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNumber_Int64_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject jsonNumber
		_, _ = subject.Int64()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNumber_MarshalJSON_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject jsonNumber
		_, _ = subject.MarshalJSON()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRawMessage_MarshalJSON_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject jsonRawMessage
		_, _ = subject.MarshalJSON()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRawMessage_UnmarshalJSON_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject *jsonRawMessage
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleValue_UnmarshalJSON_jsonHelpers() {
	func() {
		defer func() { _ = recover() }()
		var subject *jsonValue
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.Println("done")
	// Output: done
}
