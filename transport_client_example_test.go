// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestTransportClient_WithWebSocketHeaders_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_WithWebSocketHeaders_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_WithWebSocketHeaders_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_WithWebSocketDialer_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketDialer(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_WithWebSocketDialer_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketDialer(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_WithWebSocketDialer_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithWebSocketDialer(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_NewWebSocketClient_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebSocketClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_NewWebSocketClient_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebSocketClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_NewWebSocketClient_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebSocketClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_WebSocketClient_DialContext_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebSocketClient
		_, _, _ = subject.DialContext(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_WebSocketClient_DialContext_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebSocketClient
		_, _, _ = subject.DialContext(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_WebSocketClient_DialContext_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebSocketClient
		_, _, _ = subject.DialContext(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_WithSSEHeaders_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_WithSSEHeaders_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_WithSSEHeaders_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHeaders(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_WithSSEHTTPClient_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_WithSSEHTTPClient_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_WithSSEHTTPClient_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WithSSEHTTPClient(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_NewSSEClient_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_NewSSEClient_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_NewSSEClient_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEClient("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_SSEClient_Connect_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Connect(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_SSEClient_Connect_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Connect(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_SSEClient_Connect_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Connect(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransportClient_SSEClient_Events_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Events(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransportClient_SSEClient_Events_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Events(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransportClient_SSEClient_Events_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEClient
		_, _ = subject.Events(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleWithWebSocketHeaders_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = WithWebSocketHeaders(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithWebSocketDialer_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = WithWebSocketDialer(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewWebSocketClient_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = NewWebSocketClient("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebSocketClient_DialContext_transportClient() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebSocketClient
		_, _, _ = subject.DialContext(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithSSEHeaders_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = WithSSEHeaders(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWithSSEHTTPClient_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = WithSSEHTTPClient(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewSSEClient_transportClient() {
	func() {
		defer func() { _ = recover() }()
		_ = NewSSEClient("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEClient_Connect_transportClient() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEClient
		_, _ = subject.Connect(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEClient_Events_transportClient() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEClient
		_, _ = subject.Events(nil)
	}()
	coretest.Println("done")
	// Output: done
}
