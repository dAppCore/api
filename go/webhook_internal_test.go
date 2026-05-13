// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestWebhook_mustParseWebhookCIDRs_Good_ReturnsParsedNetworks(t *testing.T) {
	nets := mustParseWebhookCIDRs("127.0.0.0/8", "fc00::/7")
	if len(nets) != 2 {
		t.Fatalf("expected 2 parsed networks, got %d", len(nets))
	}
	if got := nets[0].String(); got != "127.0.0.0/8" {
		t.Fatalf("expected first network to match input, got %q", got)
	}
	if got := nets[1].String(); got != "fc00::/7" {
		t.Fatalf("expected second network to match input, got %q", got)
	}
}

func TestWebhook_mustParseWebhookCIDRs_Bad_PanicsOnInvalidCIDR(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected invalid CIDR input to panic")
		}
	}()

	_ = mustParseWebhookCIDRs("not-a-cidr")
}

func TestWebhook_mustParseWebhookCIDRs_Ugly_ReturnsEmptySliceForNoInput(t *testing.T) {
	nets := mustParseWebhookCIDRs()
	if len(nets) != 0 {
		t.Fatalf("expected no networks for empty input, got %d", len(nets))
	}
}
