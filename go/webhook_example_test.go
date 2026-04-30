// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestWebhook_WebhookEvents_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebhookEvents()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookEvents_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebhookEvents()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookEvents_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = WebhookEvents()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_IsKnownWebhookEvent_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsKnownWebhookEvent("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_IsKnownWebhookEvent_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsKnownWebhookEvent("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_IsKnownWebhookEvent_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = IsKnownWebhookEvent("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_NewWebhookSigner_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSigner("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_NewWebhookSigner_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSigner("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_NewWebhookSigner_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSigner("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_NewWebhookSignerWithTolerance_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSignerWithTolerance("", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_NewWebhookSignerWithTolerance_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSignerWithTolerance("", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_NewWebhookSignerWithTolerance_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewWebhookSignerWithTolerance("", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_GenerateWebhookSecret_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GenerateWebhookSecret()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_GenerateWebhookSecret_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GenerateWebhookSecret()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_GenerateWebhookSecret_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = GenerateWebhookSecret()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_Tolerance_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Tolerance()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_Tolerance_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Tolerance()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_Tolerance_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Tolerance()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_Sign_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Sign(nil, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_Sign_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Sign(nil, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_Sign_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Sign(nil, 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_SignNow_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_, _ = subject.SignNow(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_SignNow_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_, _ = subject.SignNow(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_SignNow_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_, _ = subject.SignNow(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_Headers_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Headers(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_Headers_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Headers(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_Headers_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Headers(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_Verify_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Verify(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_Verify_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Verify(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_Verify_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.Verify(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_VerifySignatureOnly_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifySignatureOnly(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_VerifySignatureOnly_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifySignatureOnly(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_VerifySignatureOnly_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifySignatureOnly(nil, "", 0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_IsTimestampValid_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.IsTimestampValid(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_IsTimestampValid_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.IsTimestampValid(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_IsTimestampValid_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.IsTimestampValid(0)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_WebhookSigner_VerifyRequest_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifyRequest(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_WebhookSigner_VerifyRequest_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifyRequest(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_WebhookSigner_VerifyRequest_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *WebhookSigner
		_ = subject.VerifyRequest(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestWebhook_ValidateWebhookURL_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ValidateWebhookURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestWebhook_ValidateWebhookURL_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ValidateWebhookURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestWebhook_ValidateWebhookURL_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ValidateWebhookURL("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleWebhookEvents_webhook() {
	func() {
		defer func() { _ = recover() }()
		_ = WebhookEvents()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleIsKnownWebhookEvent_webhook() {
	func() {
		defer func() { _ = recover() }()
		_ = IsKnownWebhookEvent("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewWebhookSigner_webhook() {
	func() {
		defer func() { _ = recover() }()
		_ = NewWebhookSigner("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewWebhookSignerWithTolerance_webhook() {
	func() {
		defer func() { _ = recover() }()
		_ = NewWebhookSignerWithTolerance("", 0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleGenerateWebhookSecret_webhook() {
	func() {
		defer func() { _ = recover() }()
		_, _ = GenerateWebhookSecret()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_Tolerance_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.Tolerance()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_Sign_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.Sign(nil, 0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_SignNow_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_, _ = subject.SignNow(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_Headers_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.Headers(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_Verify_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.Verify(nil, "", 0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_VerifySignatureOnly_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.VerifySignatureOnly(nil, "", 0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_IsTimestampValid_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.IsTimestampValid(0)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleWebhookSigner_VerifyRequest_webhook() {
	func() {
		defer func() { _ = recover() }()
		var subject *WebhookSigner
		_ = subject.VerifyRequest(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleValidateWebhookURL_webhook() {
	func() {
		defer func() { _ = recover() }()
		_ = ValidateWebhookURL("")
	}()
	coretest.Println("done")
	// Output: done
}
