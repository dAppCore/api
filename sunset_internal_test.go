// SPDX-License-Identifier: EUPL-1.2

package api

import "testing"

func TestSunset_successorLinkTarget_Good_StripsRecognisedMethodPrefix(t *testing.T) {
	got := successorLinkTarget("POST /api/v2/billing/invoices")
	if got != "/api/v2/billing/invoices" {
		t.Fatalf("expected successor target to strip recognised method prefix, got %q", got)
	}
}

func TestSunset_successorLinkTarget_Bad_ReturnsEmptyForBlankReplacement(t *testing.T) {
	if got := successorLinkTarget("   "); got != "" {
		t.Fatalf("expected blank replacement to return empty target, got %q", got)
	}
}

func TestSunset_successorLinkTarget_Ugly_PreservesUnknownMethodPrefix(t *testing.T) {
	got := successorLinkTarget("PURGE /api/v2/billing/invoices")
	if got != "PURGE /api/v2/billing/invoices" {
		t.Fatalf("expected unknown method prefix to be preserved verbatim, got %q", got)
	}
}
