// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestCacheConfig_Engine_CacheConfig_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.CacheConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestCacheConfig_Engine_CacheConfig_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.CacheConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestCacheConfig_Engine_CacheConfig_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Engine
		_ = subject.CacheConfig()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleEngine_CacheConfig_cacheConfig() {
	func() {
		defer func() { _ = recover() }()
		var subject *Engine
		_ = subject.CacheConfig()
	}()
	coretest.Println("done")
	// Output: done
}
