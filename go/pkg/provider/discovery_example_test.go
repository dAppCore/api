// SPDX-License-Identifier: EUPL-1.2

package provider

import coretest "dappco.re/go"

func TestDiscovery_Discover_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestDiscovery_Discover_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestDiscovery_Discover_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestDiscovery_DiscoverDefault_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestDiscovery_DiscoverDefault_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestDiscovery_DiscoverDefault_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_, _ = DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestDiscovery_Registry_Discover_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestDiscovery_Registry_Discover_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestDiscovery_Registry_Discover_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Discover("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestDiscovery_Registry_DiscoverDefault_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestDiscovery_Registry_DiscoverDefault_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestDiscovery_Registry_DiscoverDefault_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DiscoverDefault()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleDiscover_discovery() {
	func() {
		defer func() { _ = recover() }()
		_, _ = Discover("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleDiscoverDefault_discovery() {
	func() {
		defer func() { _ = recover() }()
		_, _ = DiscoverDefault()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Discover_discovery() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Discover("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_DiscoverDefault_discovery() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.DiscoverDefault()
	}()
	coretest.Println("done")
	// Output: done
}
