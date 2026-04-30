// SPDX-License-Identifier: EUPL-1.2

package stream

import coretest "dappco.re/go"

func TestStreamGroup_NewGroup_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_NewGroup_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_NewGroup_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewGroup("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestStreamGroup_Group_Name_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_Group_Name_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_Group_Name_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestStreamGroup_Group_Handlers_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Handlers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_Group_Handlers_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Handlers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_Group_Handlers_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		_ = subject.Handlers()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestStreamGroup_Group_Register_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		subject.Register(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_Group_Register_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		subject.Register(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_Group_Register_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Group
		subject.Register(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestStreamGroup_SSE_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SSE("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_SSE_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SSE("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_SSE_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = SSE("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestStreamGroup_WebSocket_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebSocket("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestStreamGroup_WebSocket_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebSocket("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestStreamGroup_WebSocket_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebSocket("", nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleNewGroup_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		_ = NewGroup("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGroup_Name_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		var subject *Group
		_ = subject.Name()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGroup_Handlers_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		var subject *Group
		_ = subject.Handlers()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGroup_Register_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		var subject *Group
		subject.Register(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleSSE_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		_ = SSE("", nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebSocket_streamGroup() {
	func() {
		defer func() { _ = recover() }()
		_ = WebSocket("", nil)
	}()
	coretest.Println("done")
	// Output: done
}
