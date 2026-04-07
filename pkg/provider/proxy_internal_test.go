// SPDX-License-Identifier: EUPL-1.2

package provider

import "testing"

func TestStripBasePath_Good_ExactBoundary(t *testing.T) {
	got := stripBasePath("/api/v1/cool-widget/items", "/api/v1/cool-widget")
	if got != "/items" {
		t.Fatalf("expected stripped path %q, got %q", "/items", got)
	}
}

func TestStripBasePath_Good_RootPath(t *testing.T) {
	got := stripBasePath("/api/v1/cool-widget", "/api/v1/cool-widget")
	if got != "/" {
		t.Fatalf("expected stripped root path %q, got %q", "/", got)
	}
}

func TestStripBasePath_Good_DoesNotTrimPartialPrefix(t *testing.T) {
	got := stripBasePath("/api/v1/cool-widget-2/items", "/api/v1/cool-widget")
	if got != "/api/v1/cool-widget-2/items" {
		t.Fatalf("expected partial prefix to remain unchanged, got %q", got)
	}
}
