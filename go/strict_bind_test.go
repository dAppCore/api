// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"testing"
	"time"

	core "dappco.re/go"
)

// TestStrictBind_addrIsLoopback_Good asserts the loopback classifier accepts
// every address shape that binds only the loopback interface.
func TestStrictBind_addrIsLoopback_Good(t *testing.T) {
	loopback := []string{
		"127.0.0.1:8787",
		"127.0.0.1",
		"[::1]:8787",
		"::1",
		"localhost:8787",
		"localhost",
		" 127.0.0.1:8787 ",
		"127.255.255.254:80",
	}
	for _, addr := range loopback {
		if !addrIsLoopback(addr) {
			t.Errorf("addrIsLoopback(%q) = false, want true", addr)
		}
	}
}

// TestStrictBind_addrIsLoopback_Bad asserts non-loopback and all-interface
// addresses are classified as not loopback.
func TestStrictBind_addrIsLoopback_Bad(t *testing.T) {
	nonLoopback := []string{
		"0.0.0.0:8787",
		":8787",
		"::",
		"[::]:8787",
		"192.168.1.10:8787",
		"10.0.0.1:8787",
		"example.com:8787",
		"",
	}
	for _, addr := range nonLoopback {
		if addrIsLoopback(addr) {
			t.Errorf("addrIsLoopback(%q) = true, want false", addr)
		}
	}
}

// TestStrictBind_validateBind_Good asserts strict mode permits loopback binds
// and that the default (non-strict) engine never rejects any address.
func TestStrictBind_validateBind_Good(t *testing.T) {
	// Default engine: strict mode off — every address passes.
	for _, addr := range []string{"0.0.0.0:8787", ":8787", "192.168.1.10:8787"} {
		e, err := New(WithAddr(addr))
		if err != nil {
			t.Fatalf("New(%q): unexpected error: %v", addr, err)
		}
		if err := e.validateBind(); err != nil {
			t.Errorf("default engine validateBind(%q) = %v, want nil", addr, err)
		}
	}

	// Strict mode on, loopback bind — passes with no bearer required.
	e, err := New(WithAddr("127.0.0.1:8787"), WithStrictBind())
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); err != nil {
		t.Errorf("strict loopback validateBind = %v, want nil", err)
	}

	// Strict mode on, public bind, bearer present — passes.
	e, err = New(
		WithAddr("0.0.0.0:8787"),
		WithStrictBind(),
		WithPublicBind(),
		WithBearerAuth("secret"),
	)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); err != nil {
		t.Errorf("strict public+bearer validateBind = %v, want nil", err)
	}

	// WithLoopbackOnly with a loopback bind passes.
	e, err = New(WithAddr("[::1]:8787"), WithLoopbackOnly())
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); err != nil {
		t.Errorf("loopback-only validateBind = %v, want nil", err)
	}
}

// TestStrictBind_validateBind_Bad asserts strict mode rejects a non-loopback
// bind without the explicit public opt-in, and rejects a public bind that has
// no bearer credential.
func TestStrictBind_validateBind_Bad(t *testing.T) {
	// Non-loopback bind without WithPublicBind — rejected.
	e, err := New(WithAddr("0.0.0.0:8787"), WithStrictBind())
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); !core.Is(err, ErrNonLoopbackBind) {
		t.Errorf("strict public-no-flag validateBind = %v, want ErrNonLoopbackBind", err)
	}

	// WithLoopbackOnly never permits a public bind even with a bearer.
	e, err = New(
		WithAddr("0.0.0.0:8787"),
		WithLoopbackOnly(),
		WithBearerAuth("secret"),
	)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); !core.Is(err, ErrNonLoopbackBind) {
		t.Errorf("loopback-only public validateBind = %v, want ErrNonLoopbackBind", err)
	}

	// Public bind opted in but no bearer — rejected.
	e, err = New(
		WithAddr("0.0.0.0:8787"),
		WithStrictBind(),
		WithPublicBind(),
	)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); !core.Is(err, ErrPublicBindNoBearer) {
		t.Errorf("strict public-no-bearer validateBind = %v, want ErrPublicBindNoBearer", err)
	}

	// Public bind opted in with an empty bearer token — still rejected, since
	// an empty token does not configure a credential.
	e, err = New(
		WithAddr("0.0.0.0:8787"),
		WithStrictBind(),
		WithPublicBind(),
		WithBearerAuth("   "),
	)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.validateBind(); !core.Is(err, ErrPublicBindNoBearer) {
		t.Errorf("strict public-empty-bearer validateBind = %v, want ErrPublicBindNoBearer", err)
	}
}

// TestStrictBind_Serve_Ugly asserts Serve fails fast on a misconfigured strict
// engine — the listener never opens — and that a non-strict engine on the same
// non-loopback address is unaffected (it binds and shuts down cleanly).
func TestStrictBind_Serve_Ugly(t *testing.T) {
	// Strict + non-loopback + no public flag: Serve must return the sentinel
	// immediately, before binding, regardless of context cancellation.
	e, err := New(WithAddr("0.0.0.0:0"), WithStrictBind())
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.Serve(context.Background()); !core.Is(err, ErrNonLoopbackBind) {
		t.Fatalf("strict Serve = %v, want ErrNonLoopbackBind", err)
	}

	// Strict + public flag + no bearer: Serve must return ErrPublicBindNoBearer
	// without opening a listener.
	e, err = New(WithAddr("0.0.0.0:0"), WithStrictBind(), WithPublicBind())
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if err := e.Serve(context.Background()); !core.Is(err, ErrPublicBindNoBearer) {
		t.Fatalf("strict Serve = %v, want ErrPublicBindNoBearer", err)
	}

	// Default engine on a non-loopback ephemeral port: must bind and exit
	// cleanly on context cancellation — strict enforcement is opt-in only.
	e, err = New(WithAddr("127.0.0.1:0"))
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.Serve(ctx)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("default Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("default Serve did not return after cancel")
	}
}
