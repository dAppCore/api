// SPDX-License-Identifier: EUPL-1.2

package main

import (
	bytes "dappco.re/go/api/internal/stdcompat/corebytes"
	strings "dappco.re/go/api/internal/stdcompat/corestrings"
	"io"
	"log/slog"
	"testing"

	coreapi "dappco.re/go/api"
	"github.com/gin-gonic/gin"
)

func TestMain_Help(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code 0, got %d; stderr=%s", code, stderr.String())
	}

	output := stdout.String()
	for _, expected := range []string{
		"CORE_GATEWAY_BIND",
		"CORE_GATEWAY_ENABLE",
		"brain",
		"brain-mcp",
		"scm",
		"process",
		"build",
		"miner",
		"proxy",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected help output to contain %q; output=%s", expected, output)
		}
	}
}

func TestMain_RegisterProvider_HandlesNilProvider(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine, err := coreapi.New()
	if err != nil {
		t.Fatalf("api.New failed: %v", err)
	}

	var provider *stubRouteGroup
	registered := registerProvider(logger, engine, providerSpec{Name: "nil-provider"}, func() coreapi.RouteGroup {
		return provider
	})
	if registered {
		t.Fatal("expected nil provider not to register")
	}
	if got := len(engine.Groups()); got != 0 {
		t.Fatalf("expected no registered groups, got %d", got)
	}
}

func TestMain_EnableFiltersSubset(t *testing.T) {
	selected := selectedProviders("scm,process")
	var enabled []string
	for _, spec := range gatewayProviderSpecs() {
		if providerEnabled(spec, selected) {
			enabled = append(enabled, spec.Name)
		}
	}
	if got, want := strings.Join(enabled, ","), "scm,process"; got != want {
		t.Fatalf("expected enabled providers %q, got %q", want, got)
	}
}

type stubRouteGroup struct{}

func (*stubRouteGroup) Name() string {
	return "stub"
}

func (*stubRouteGroup) BasePath() string {
	return "/stub"
}

func (*stubRouteGroup) RegisterRoutes(*gin.RouterGroup) {}
