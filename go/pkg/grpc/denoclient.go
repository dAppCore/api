// SPDX-License-Identifier: EUPL-1.2

package grpc

import (
	"context"

	core "dappco.re/go"
	sidecarpb "dappco.re/go/api/pkg/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const denoClientScope = "grpc.DenoClient"

// DenoClient is the Go-side client for a running Deno sidecar. It
// dials the Deno-hosted DenoService and exposes lifecycle, render, and
// eval calls. The address may be a TCP host:port or a Unix socket path
// prefixed with "unix:".
//
//	client, _ := grpc.NewDenoClient("localhost:50052")
//	defer client.Close()
//	_ = client.OnConfigChange(ctx, "theme.accent", "#6366f1")
//	html, _ := client.Render(ctx, "dashboard", `{"user":"snider"}`)
type DenoClient struct {
	conn   *grpc.ClientConn
	client sidecarpb.DenoServiceClient
}

// NewDenoClient dials the Deno sidecar at addr and returns a ready
// client. TLS is optional for localhost; this constructor uses an
// insecure transport suitable for loopback or Unix-socket sidecars.
// A "unix:" prefix selects Unix-socket transport.
//
//	client, err := grpc.NewDenoClient("localhost:50052")
//	defer client.Close()
func NewDenoClient(addr string) (*DenoClient, error) {
	if core.Trim(addr) == "" {
		return nil, core.E(denoClientScope, "empty address", nil)
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, core.E(denoClientScope, core.Concat("dial failed: ", addr), err)
	}
	return &DenoClient{
		conn:   conn,
		client: sidecarpb.NewDenoServiceClient(conn),
	}, nil
}

// NewDenoClientTLS dials the Deno sidecar over a TLS-secured transport
// using the supplied transport credentials.
//
//	creds := credentials.NewTLS(tlsConfig)
//	client, _ := grpc.NewDenoClientTLS("deno.internal:50052", creds)
func NewDenoClientTLS(addr string, creds credentials.TransportCredentials) (*DenoClient, error) {
	if core.Trim(addr) == "" {
		return nil, core.E(denoClientScope, "empty address", nil)
	}
	if creds == nil {
		return nil, core.E(denoClientScope, "nil transport credentials", nil)
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, core.E(denoClientScope, core.Concat("dial failed: ", addr), err)
	}
	return &DenoClient{
		conn:   conn,
		client: sidecarpb.NewDenoServiceClient(conn),
	}, nil
}

// OnStart notifies the Deno runtime that a lifecycle start has
// occurred. reason is optional human context.
//
//	_ = client.OnStart(ctx, "boot")
func (c *DenoClient) OnStart(ctx context.Context, reason string) error {
	if c == nil || c.client == nil {
		return core.E(denoClientScope, "client not initialised", nil)
	}
	_, err := c.client.OnStart(ctx, &sidecarpb.LifecycleEvent{Phase: "start", Reason: reason})
	if err != nil {
		return core.E(denoClientScope, "OnStart failed", err)
	}
	return nil
}

// OnStop notifies the Deno runtime that a lifecycle stop has occurred.
//
//	_ = client.OnStop(ctx, "shutdown")
func (c *DenoClient) OnStop(ctx context.Context, reason string) error {
	if c == nil || c.client == nil {
		return core.E(denoClientScope, "client not initialised", nil)
	}
	_, err := c.client.OnStop(ctx, &sidecarpb.LifecycleEvent{Phase: "stop", Reason: reason})
	if err != nil {
		return core.E(denoClientScope, "OnStop failed", err)
	}
	return nil
}

// OnConfigChange notifies the Deno runtime of a settings change.
//
//	_ = client.OnConfigChange(ctx, "theme.accent", "#6366f1")
func (c *DenoClient) OnConfigChange(ctx context.Context, key, value string) error {
	if c == nil || c.client == nil {
		return core.E(denoClientScope, "client not initialised", nil)
	}
	_, err := c.client.OnConfigChange(ctx, &sidecarpb.ConfigChangeEvent{Key: key, Value: value})
	if err != nil {
		return core.E(denoClientScope, "OnConfigChange failed", err)
	}
	return nil
}

// Render asks the Deno runtime to server-side render a component with
// the given JSON props, returning the rendered HTML.
//
//	html, _ := client.Render(ctx, "dashboard", `{"user":"snider"}`)
func (c *DenoClient) Render(ctx context.Context, component, props string) (string, error) {
	if c == nil || c.client == nil {
		return "", core.E(denoClientScope, "client not initialised", nil)
	}
	resp, err := c.client.Render(ctx, &sidecarpb.RenderRequest{Component: component, Props: props})
	if err != nil {
		return "", core.E(denoClientScope, "Render failed", err)
	}
	if resp.GetError() != "" {
		return "", core.E(denoClientScope, resp.GetError(), nil)
	}
	return resp.GetHtml(), nil
}

// Eval asks the Deno runtime to evaluate a TypeScript expression,
// returning the JSON-encoded result.
//
//	resultJSON, _ := client.Eval(ctx, "1 + 1")
func (c *DenoClient) Eval(ctx context.Context, expression string) (string, error) {
	if c == nil || c.client == nil {
		return "", core.E(denoClientScope, "client not initialised", nil)
	}
	resp, err := c.client.Eval(ctx, &sidecarpb.EvalRequest{Expression: expression})
	if err != nil {
		return "", core.E(denoClientScope, "Eval failed", err)
	}
	if resp.GetError() != "" {
		return "", core.E(denoClientScope, resp.GetError(), nil)
	}
	return resp.GetResultJson(), nil
}

// Close releases the client connection. Safe to call on a nil client.
//
//	defer client.Close()
func (c *DenoClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return core.E(denoClientScope, "close failed", err)
	}
	return nil
}
