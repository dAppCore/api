// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"sync"
	"testing"

	api "dappco.re/go/api"
)

func TestUpstreamRegistry_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	if err := reg.Set("lemma", api.Upstream{URL: "https://a.example.com:8000", Weight: 2}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := reg.Add("lemma", api.Upstream{URL: "https://b.example.com"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := reg.SetDefault(api.Upstream{URL: "https://fallback.example.com"}); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}
	keys := reg.Keys()
	if len(keys) != 1 || keys[0] != "lemma" {
		t.Fatalf("Keys = %v, want [lemma]", keys)
	}
}

func TestUpstreamRegistry_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	cases := map[string]string{
		"scheme":   "ftp://a.example.com",
		"no-host":  "http://",
		"bad-port": "http://a.example.com:99999",
		"creds":    "http://user:pass@a.example.com",
		"loopback": "http://127.0.0.1:11434",
		"private":  "http://10.0.0.5:8000",
		"metadata": "http://169.254.169.254",
	}
	for name, raw := range cases {
		if err := reg.Set("k", api.Upstream{URL: raw}); err == nil {
			t.Errorf("%s: Set(%q) = nil error, want rejection", name, raw)
		}
	}
}

func TestUpstreamRegistry_AllowPrivate_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8"))
	if err := reg.Set("local", api.Upstream{URL: "http://127.0.0.1:11434"}); err != nil {
		t.Fatalf("Set loopback with allow-list: %v", err)
	}
	// Metadata stays hard-blocked even with a broad allow-list.
	reg2 := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("0.0.0.0/0"))
	if err := reg2.Set("meta", api.Upstream{URL: "http://169.254.169.254"}); err == nil {
		t.Fatal("metadata host accepted under broad allow-list, want rejection")
	}
}

func TestUpstreamRegistry_Remove_Good(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	if err := reg.Set("k", api.Upstream{URL: "https://a.example.com"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	reg.Remove("k")
	if keys := reg.Keys(); len(keys) != 0 {
		t.Fatalf("Keys after Remove = %v, want []", keys)
	}
}

func TestUpstreamRegistry_BadCIDR_Bad(t *testing.T) {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("not-a-cidr"))
	// The recorded cidrErr must surface on every subsequent write, even for an
	// otherwise-valid public upstream.
	if err := reg.Set("k", api.Upstream{URL: "https://a.example.com"}); err == nil {
		t.Fatal("Set with recorded bad-CIDR error = nil, want rejection")
	}
	if err := reg.Add("k", api.Upstream{URL: "https://b.example.com"}); err == nil {
		t.Fatal("Add with recorded bad-CIDR error = nil, want rejection")
	}
}

func TestUpstreamRegistry_Ugly_ConcurrentWriteSnapshot(t *testing.T) {
	reg := api.NewUpstreamRegistry()
	_ = reg.Set("k", api.Upstream{URL: "https://a.example.com"})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = reg.Add("k", api.Upstream{URL: "https://b.example.com"}) }()
		go func() { defer wg.Done(); _ = reg.Keys() }()
	}
	wg.Wait()
}
