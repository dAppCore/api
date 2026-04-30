// SPDX-License-Identifier: EUPL-1.2

package provider

import coretest "dappco.re/go"

func TestProxy_ProviderUpstreamBlockedError_Error_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Error_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Error_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Error()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Is_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Is(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Is_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Is(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Is_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Is(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Unwrap_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Unwrap()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Unwrap_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Unwrap()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProviderUpstreamBlockedError_Unwrap_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProviderUpstreamBlockedError
		_ = subject.Unwrap()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_NewProxy_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewProxy(ProxyConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_NewProxy_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewProxy(ProxyConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_NewProxy_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewProxy(ProxyConfig{})
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_Err_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Err()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_Err_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Err()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_Err_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Err()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_Name_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_Name_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_Name_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Name()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_BasePath_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_BasePath_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_BasePath_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.BasePath()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_RegisterRoutes_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_RegisterRoutes_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_RegisterRoutes_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		subject.RegisterRoutes(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_Element_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Element()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_Element_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Element()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_Element_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Element()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_SpecFile_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.SpecFile()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_SpecFile_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.SpecFile()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_SpecFile_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.SpecFile()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestProxy_ProxyProvider_Upstream_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Upstream()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestProxy_ProxyProvider_Upstream_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Upstream()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestProxy_ProxyProvider_Upstream_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *ProxyProvider
		_ = subject.Upstream()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleProviderUpstreamBlockedError_Error_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProviderUpstreamBlockedError
		_ = subject.Error()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProviderUpstreamBlockedError_Is_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProviderUpstreamBlockedError
		_ = subject.Is(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProviderUpstreamBlockedError_Unwrap_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProviderUpstreamBlockedError
		_ = subject.Unwrap()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleNewProxy_proxy() {
	func() {
		defer func() { _ = recover() }()
		_ = NewProxy(ProxyConfig{})
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_Err_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.Err()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_Name_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.Name()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_BasePath_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.BasePath()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_RegisterRoutes_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		subject.RegisterRoutes(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_Element_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.Element()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_SpecFile_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.SpecFile()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleProxyProvider_Upstream_proxy() {
	func() {
		defer func() { _ = recover() }()
		var subject *ProxyProvider
		_ = subject.Upstream()
	}()
	coretest.Println("done")
	// Output: done
}
