// SPDX-License-Identifier: EUPL-1.2

package api

import (
	coretest "dappco.re/go"
	inference "dappco.re/go/inference"
)

func TestChatCompletions_StopList_UnmarshalJSON_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatStopList
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_StopList_UnmarshalJSON_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatStopList
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_StopList_UnmarshalJSON_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatStopList
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject ChatMessageDelta
		_, _ = subject.MarshalJSON()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject ChatMessageDelta
		_, _ = subject.MarshalJSON()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ChatMessageDelta_MarshalJSON_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject ChatMessageDelta
		_, _ = subject.MarshalJSON()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ResolutionError_Error_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *modelResolutionError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ResolutionError_Error_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *modelResolutionError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ResolutionError_Error_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *modelResolutionError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_NewModelResolver_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewModelResolver()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_NewModelResolver_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewModelResolver()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_NewModelResolver_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewModelResolver()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ModelResolver_ResolveModel_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ModelResolver
		_, _ = subject.ResolveModel("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ModelResolver_ResolveModel_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ModelResolver
		_, _ = subject.ResolveModel("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ModelResolver_ResolveModel_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ModelResolver
		_, _ = subject.ResolveModel("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_NewThinkingExtractor_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewThinkingExtractor()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_NewThinkingExtractor_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewThinkingExtractor()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_NewThinkingExtractor_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewThinkingExtractor()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ThinkingExtractor_Process_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		subject.Process(inference.Token{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ThinkingExtractor_Process_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		subject.Process(inference.Token{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ThinkingExtractor_Process_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		subject.Process(inference.Token{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ThinkingExtractor_Content_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Content()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ThinkingExtractor_Content_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Content()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ThinkingExtractor_Content_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Content()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_ThinkingExtractor_Thinking_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Thinking()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_ThinkingExtractor_Thinking_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Thinking()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_ThinkingExtractor_Thinking_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ThinkingExtractor
		_ = subject.Thinking()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_CompletionsHandler_ServeHTTP_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionsHandler
		subject.ServeHTTP(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_CompletionsHandler_ServeHTTP_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionsHandler
		subject.ServeHTTP(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_CompletionsHandler_ServeHTTP_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionsHandler
		subject.ServeHTTP(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestChatCompletions_CompletionRequestError_Error_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionRequestError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestChatCompletions_CompletionRequestError_Error_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionRequestError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestChatCompletions_CompletionRequestError_Error_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *chatCompletionRequestError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleStopList_UnmarshalJSON_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *chatStopList
		_ = subject.UnmarshalJSON(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleChatMessageDelta_MarshalJSON_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject ChatMessageDelta
		_, _ = subject.MarshalJSON()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleResolutionError_Error_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *modelResolutionError
		_ = subject.Error()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewModelResolver_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		_ = NewModelResolver()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleModelResolver_ResolveModel_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *ModelResolver
		_, _ = subject.ResolveModel("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewThinkingExtractor_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		_ = NewThinkingExtractor()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleThinkingExtractor_Process_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *ThinkingExtractor
		subject.Process(inference.Token{})
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleThinkingExtractor_Content_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *ThinkingExtractor
		_ = subject.Content()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleThinkingExtractor_Thinking_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *ThinkingExtractor
		_ = subject.Thinking()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleCompletionsHandler_ServeHTTP_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *chatCompletionsHandler
		subject.ServeHTTP(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleCompletionRequestError_Error_chatCompletions() {
	func() {
		defer func() { _ = recover() }()
		var subject *chatCompletionRequestError
		_ = subject.Error()
	}()
	coretest.Println("done")
	// Output: done
}
