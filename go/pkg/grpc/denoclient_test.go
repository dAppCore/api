// SPDX-License-Identifier: EUPL-1.2

package grpc_test

import (
	"context"
	"net"
	"testing"

	core "dappco.re/go"
	apigrpc "dappco.re/go/api/pkg/grpc"
	sidecarpb "dappco.re/go/api/pkg/proto/gen"
	"google.golang.org/grpc"
)

// fakeDenoServer stands in for the Deno-hosted DenoService. It records
// the last config change and can be told to fail Render/Eval to drive
// the Ugly path.
type fakeDenoServer struct {
	sidecarpb.UnimplementedDenoServiceServer
	lastKey    string
	lastValue  string
	failRender bool
}

func (f *fakeDenoServer) OnStart(_ context.Context, _ *sidecarpb.LifecycleEvent) (*sidecarpb.Ack, error) {
	return &sidecarpb.Ack{Ok: true}, nil
}

func (f *fakeDenoServer) OnStop(_ context.Context, _ *sidecarpb.LifecycleEvent) (*sidecarpb.Ack, error) {
	return &sidecarpb.Ack{Ok: true}, nil
}

func (f *fakeDenoServer) OnConfigChange(_ context.Context, ev *sidecarpb.ConfigChangeEvent) (*sidecarpb.Ack, error) {
	f.lastKey = ev.GetKey()
	f.lastValue = ev.GetValue()
	return &sidecarpb.Ack{Ok: true}, nil
}

func (f *fakeDenoServer) Render(_ context.Context, req *sidecarpb.RenderRequest) (*sidecarpb.RenderResponse, error) {
	if f.failRender {
		return &sidecarpb.RenderResponse{Error: "render exploded"}, nil
	}
	return &sidecarpb.RenderResponse{Html: core.Concat("<div>", req.GetComponent(), "</div>")}, nil
}

func (f *fakeDenoServer) Eval(_ context.Context, req *sidecarpb.EvalRequest) (*sidecarpb.EvalResponse, error) {
	return &sidecarpb.EvalResponse{ResultJson: core.Concat(`"`, req.GetExpression(), `"`)}, nil
}

// startFakeDeno serves a fakeDenoServer on a Unix socket and returns
// the dial address ("unix:<socket>") plus the backing fake.
func startFakeDeno(t *testing.T, fake *fakeDenoServer) string {
	t.Helper()
	socket := core.JoinPath(t.TempDir(), "deno.sock")
	lis, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	sidecarpb.RegisterDenoServiceServer(srv, fake)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)
	return "unix:" + socket
}

func TestDenoClient_Lifecycle_Good(t *testing.T) {
	t.Parallel()
	fake := &fakeDenoServer{}
	addr := startFakeDeno(t, fake)

	client, err := apigrpc.NewDenoClient(addr)
	if err != nil {
		t.Fatalf("NewDenoClient: %v", err)
	}
	defer func() { _ = client.Close() }()
	c := ctx(t)

	if err := client.OnStart(c, "boot"); err != nil {
		t.Fatalf("OnStart: %v", err)
	}
	if err := client.OnConfigChange(c, "theme.accent", "#6366f1"); err != nil {
		t.Fatalf("OnConfigChange: %v", err)
	}
	if fake.lastKey != "theme.accent" || fake.lastValue != "#6366f1" {
		t.Fatalf("config not received: %q=%q", fake.lastKey, fake.lastValue)
	}

	html, err := client.Render(c, "dashboard", `{"user":"snider"}`)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if html != "<div>dashboard</div>" {
		t.Fatalf("html = %q", html)
	}

	result, err := client.Eval(c, "1+1")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if result != `"1+1"` {
		t.Fatalf("eval = %q", result)
	}

	if err := client.OnStop(c, "shutdown"); err != nil {
		t.Fatalf("OnStop: %v", err)
	}
}

func TestDenoClient_NewDenoClient_Bad(t *testing.T) {
	t.Parallel()
	// An empty address must be rejected before any dial attempt.
	if _, err := apigrpc.NewDenoClient(""); err == nil {
		t.Fatal("expected error for empty address")
	}
	// A nil transport credential to the TLS constructor is rejected.
	if _, err := apigrpc.NewDenoClientTLS("localhost:1", nil); err == nil {
		t.Fatal("expected error for nil TLS credentials")
	}
}

func TestDenoClient_Render_Ugly(t *testing.T) {
	t.Parallel()
	// The Deno side returns an application error in the response body
	// (not a transport error). The client must convert that populated
	// error field into a Go error.
	fake := &fakeDenoServer{failRender: true}
	addr := startFakeDeno(t, fake)

	client, err := apigrpc.NewDenoClient(addr)
	if err != nil {
		t.Fatalf("NewDenoClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	if _, err := client.Render(ctx(t), "dashboard", "{}"); err == nil {
		t.Fatal("expected render error surfaced from response body")
	} else if !core.Contains(err.Error(), "render exploded") {
		t.Fatalf("error = %q", err.Error())
	}

	// Calls on a nil client must not panic.
	var nilClient *apigrpc.DenoClient
	if err := nilClient.OnStart(ctx(t), "x"); err == nil {
		t.Fatal("expected error from nil client")
	}
	if err := nilClient.Close(); err != nil {
		t.Fatalf("nil Close should be a no-op, got %v", err)
	}
}
