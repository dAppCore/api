// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestSse_NewSSEBroker_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEBroker()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSse_NewSSEBroker_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEBroker()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSse_NewSSEBroker_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewSSEBroker()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSse_SSEBroker_Publish_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Publish("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSse_SSEBroker_Publish_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Publish("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSse_SSEBroker_Publish_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Publish("", "", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSse_SSEBroker_Handler_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.Handler()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSse_SSEBroker_Handler_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.Handler()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSse_SSEBroker_Handler_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.Handler()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSse_SSEBroker_ClientCount_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.ClientCount()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSse_SSEBroker_ClientCount_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.ClientCount()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSse_SSEBroker_ClientCount_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		_ = subject.ClientCount()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestSse_SSEBroker_Drain_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Drain()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestSse_SSEBroker_Drain_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Drain()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestSse_SSEBroker_Drain_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *SSEBroker
		subject.Drain()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleNewSSEBroker_sse() {
	func() {
		defer func() { _ = recover() }()
		_ = NewSSEBroker()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEBroker_Publish_sse() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEBroker
		subject.Publish("", "", nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEBroker_Handler_sse() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEBroker
		_ = subject.Handler()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEBroker_ClientCount_sse() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEBroker
		_ = subject.ClientCount()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSEBroker_Drain_sse() {
	func() {
		defer func() { _ = recover() }()
		var subject *SSEBroker
		subject.Drain()
	}()
	coretest.Println("done")
	// Output: done
}
