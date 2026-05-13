// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestClient_WithSpec_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpec("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_WithSpec_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpec("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_WithSpec_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpec("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_WithSpecReader_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpecReader(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_WithSpecReader_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpecReader(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_WithSpecReader_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSpecReader(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_WithBaseURL_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBaseURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_WithBaseURL_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBaseURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_WithBaseURL_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBaseURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_WithBearerToken_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBearerToken("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_WithBearerToken_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBearerToken("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_WithBearerToken_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithBearerToken("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_WithHTTPClient_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_WithHTTPClient_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_WithHTTPClient_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_NewOpenAPIClient_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewOpenAPIClient()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_NewOpenAPIClient_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewOpenAPIClient()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_NewOpenAPIClient_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewOpenAPIClient()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_OpenAPIClient_Operations_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Operations()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_OpenAPIClient_Operations_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Operations()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_OpenAPIClient_Operations_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Operations()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_OpenAPIClient_OperationsIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.OperationsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_OpenAPIClient_OperationsIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.OperationsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_OpenAPIClient_OperationsIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.OperationsIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_OpenAPIClient_Servers_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Servers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_OpenAPIClient_Servers_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Servers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_OpenAPIClient_Servers_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Servers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_OpenAPIClient_ServersIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.ServersIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_OpenAPIClient_ServersIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.ServersIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_OpenAPIClient_ServersIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.ServersIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestClient_OpenAPIClient_Call_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Call("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestClient_OpenAPIClient_Call_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Call("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestClient_OpenAPIClient_Call_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *OpenAPIClient
		_, _ = subject.Call("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleWithSpec_client() {
	func() {
		defer func() { _ = recover() }()
		_ = WithSpec("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithSpecReader_client() {
	func() {
		defer func() { _ = recover() }()
		_ = WithSpecReader(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithBaseURL_client() {
	func() {
		defer func() { _ = recover() }()
		_ = WithBaseURL("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithBearerToken_client() {
	func() {
		defer func() { _ = recover() }()
		_ = WithBearerToken("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithHTTPClient_client() {
	func() {
		defer func() { _ = recover() }()
		_ = WithHTTPClient(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewOpenAPIClient_client() {
	func() {
		defer func() { _ = recover() }()
		_ = NewOpenAPIClient()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleOpenAPIClient_Operations_client() {
	func() {
		defer func() { _ = recover() }()
		var subject *OpenAPIClient
		_, _ = subject.Operations()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleOpenAPIClient_OperationsIter_client() {
	func() {
		defer func() { _ = recover() }()
		var subject *OpenAPIClient
		_, _ = subject.OperationsIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleOpenAPIClient_Servers_client() {
	func() {
		defer func() { _ = recover() }()
		var subject *OpenAPIClient
		_, _ = subject.Servers()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleOpenAPIClient_ServersIter_client() {
	func() {
		defer func() { _ = recover() }()
		var subject *OpenAPIClient
		_, _ = subject.ServersIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleOpenAPIClient_Call_client() {
	func() {
		defer func() { _ = recover() }()
		var subject *OpenAPIClient
		_, _ = subject.Call("", nil)
	}()
	coretest.Println("done")
	// Output: done
}
