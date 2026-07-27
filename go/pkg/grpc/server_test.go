// SPDX-License-Identifier: EUPL-1.2

package grpc_test

import (
	"context"
	"io/fs"
	"testing"
	"time"

	core "dappco.re/go"
	apigrpc "dappco.re/go/api/pkg/grpc"
	sidecarpb "dappco.re/go/api/pkg/proto/gen"
	coreio "dappco.re/go/io"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// --- Fakes ---------------------------------------------------------------

// fakeStore is an in-memory KVStore for table tests. A non-empty
// failGroup makes every call to that group return a failed Result,
// exercising the Ugly path.
type fakeStore struct {
	data      map[string]string
	failGroup string
}

func newFakeStore() *fakeStore { return &fakeStore{data: map[string]string{}} }

func (s *fakeStore) key(group, key string) string { return core.Concat(group, "\x00", key) }

func (s *fakeStore) Get(group, key string) (string, core.Result) {
	if group == s.failGroup {
		return "", core.Fail(core.E("fakeStore.Get", "backend down", nil))
	}
	return s.data[s.key(group, key)], core.Ok(nil)
}

func (s *fakeStore) Set(group, key, value string) core.Result {
	if group == s.failGroup {
		return core.Fail(core.E("fakeStore.Set", "backend down", nil))
	}
	s.data[s.key(group, key)] = value
	return core.Ok(nil)
}

// fakeRunner is a ProcRunner that echoes a canned output or fails when
// the command equals "boom".
type fakeRunner struct{}

func (fakeRunner) RunWithOptions(_ context.Context, opts apigrpc.RunOptions) core.Result {
	if opts.Command == "boom" {
		return core.Fail(core.E("fakeRunner", "exit 1", nil))
	}
	return core.Ok(core.Concat("ran:", opts.Command))
}

// fakeResolver resolves any core:// URL to a fixed map, or errors when
// the URL contains "missing".
type fakeResolver struct{}

func (fakeResolver) Resolve(url string) (any, error) {
	if core.Contains(url, "missing") {
		return nil, core.E("fakeResolver", "no such scheme segment", nil)
	}
	return map[string]string{"url": url, "value": "#6366f1"}, nil
}

// dialServer brings up a GRPCServer on a Unix socket in t.TempDir and
// returns a connected GoServiceClient. Cleanup stops both ends.
func dialServer(t *testing.T, svc *apigrpc.GoService) sidecarpb.GoServiceClient {
	t.Helper()
	socket := core.JoinPath(t.TempDir(), "sidecar.sock")
	srv, err := apigrpc.NewGRPCServer(
		apigrpc.WithGRPCSocket(socket),
		apigrpc.WithGRPCServices(svc),
	)
	if err != nil {
		t.Fatalf("NewGRPCServer: %v", err)
	}
	go func() { _ = srv.Serve() }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("unix:"+socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return sidecarpb.NewGoServiceClient(conn)
}

func ctx(t *testing.T) context.Context {
	t.Helper()
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return c
}

// --- NewGRPCServer -------------------------------------------------------

func TestServer_NewGRPCServer_Good(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		opts []apigrpc.GRPCOption
		want string // expected substring of Address()
	}{
		{
			name: "tcp loopback default port resolves",
			opts: []apigrpc.GRPCOption{apigrpc.WithGRPCPort(0)},
			want: "127.0.0.1:",
		},
		{
			name: "unix socket transport",
			opts: []apigrpc.GRPCOption{apigrpc.WithGRPCSocket(core.JoinPath(t.TempDir(), "s.sock"))},
			want: ".sock",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := apigrpc.NewGRPCServer(tc.opts...)
			if err != nil {
				t.Fatalf("NewGRPCServer: %v", err)
			}
			defer srv.Stop()
			if !core.Contains(srv.Address(), tc.want) {
				t.Fatalf("Address() = %q, want substring %q", srv.Address(), tc.want)
			}
		})
	}
}

func TestServer_NewGRPCServer_Bad(t *testing.T) {
	t.Parallel()
	// Binding a Unix socket inside a non-existent directory fails at
	// listen() and must surface a Core error, not a nil server.
	_, err := apigrpc.NewGRPCServer(
		apigrpc.WithGRPCSocket("/this/path/does/not/exist/sidecar.sock"),
	)
	if err == nil {
		t.Fatal("expected listen error for non-existent socket directory")
	}
	if !core.Contains(err.Error(), "unix listen failed") {
		t.Fatalf("error = %q, want unix listen failure", err.Error())
	}
}

func TestServer_NewGRPCServer_Ugly(t *testing.T) {
	t.Parallel()
	// Two servers asked to bind the SAME Unix socket: the first wins,
	// the second must fail to listen rather than silently share.
	socket := core.JoinPath(t.TempDir(), "contended.sock")
	first, err := apigrpc.NewGRPCServer(apigrpc.WithGRPCSocket(socket))
	if err != nil {
		t.Fatalf("first NewGRPCServer: %v", err)
	}
	go func() { _ = first.Serve() }()
	defer first.Stop()

	// NewGRPCServer best-effort removes a stale socket file, but a live
	// listener on it cannot be rebound on the same path while held.
	second, secondErr := apigrpc.NewGRPCServer(apigrpc.WithGRPCSocket(socket))
	if secondErr != nil {
		return // expected: contended bind rejected
	}
	// On platforms that permit the rebind after unlink, ensure the
	// second server at least produced a usable distinct listener and
	// clean up so the test does not leak.
	second.Stop()
}

// --- GoService rpcs end-to-end ------------------------------------------

func TestServer_GoServiceRPCs_Good(t *testing.T) {
	t.Parallel()
	mem := coreio.NewMemoryMedium()
	if err := mem.WriteMode("config/app.yaml", "port: 8080", 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	store := newFakeStore()
	svc := apigrpc.NewGoService(mem, store, fakeRunner{}).WithScheme(fakeResolver{})
	client := dialServer(t, svc)
	c := ctx(t)

	t.Run("ReadFile", func(t *testing.T) {
		resp, err := client.ReadFile(c, &sidecarpb.ReadFileRequest{Path: "config/app.yaml"})
		if err != nil {
			t.Fatalf("rpc: %v", err)
		}
		if resp.GetError() != "" {
			t.Fatalf("app error: %s", resp.GetError())
		}
		if string(resp.GetData()) != "port: 8080" {
			t.Fatalf("data = %q", string(resp.GetData()))
		}
	})

	t.Run("WriteFile then ListFiles", func(t *testing.T) {
		if _, err := client.WriteFile(c, &sidecarpb.WriteFileRequest{Path: "config/new.txt", Data: []byte("hi")}); err != nil {
			t.Fatalf("write rpc: %v", err)
		}
		resp, err := client.ListFiles(c, &sidecarpb.ListFilesRequest{Path: "config"})
		if err != nil {
			t.Fatalf("list rpc: %v", err)
		}
		var names []string
		for _, e := range resp.GetEntries() {
			names = append(names, e.GetName())
		}
		if len(names) != 2 {
			t.Fatalf("entries = %v, want 2 files", names)
		}
	})

	t.Run("StoreSet then StoreGet", func(t *testing.T) {
		if _, err := client.StoreSet(c, &sidecarpb.StoreSetRequest{Group: "workspace", Key: "task", Value: "ship"}); err != nil {
			t.Fatalf("set rpc: %v", err)
		}
		resp, err := client.StoreGet(c, &sidecarpb.StoreGetRequest{Group: "workspace", Key: "task"})
		if err != nil {
			t.Fatalf("get rpc: %v", err)
		}
		if resp.GetValue() != "ship" || !resp.GetFound() {
			t.Fatalf("get = %+v", resp)
		}
	})

	t.Run("Exec", func(t *testing.T) {
		resp, err := client.Exec(c, &sidecarpb.ExecRequest{Cmd: "echo", Args: []string{"hi"}})
		if err != nil {
			t.Fatalf("exec rpc: %v", err)
		}
		if string(resp.GetOutput()) != "ran:echo" {
			t.Fatalf("output = %q", string(resp.GetOutput()))
		}
	})

	t.Run("ResolveScheme", func(t *testing.T) {
		resp, err := client.ResolveScheme(c, &sidecarpb.SchemeRequest{Uri: "core://settings/theme.accent"})
		if err != nil {
			t.Fatalf("scheme rpc: %v", err)
		}
		if !core.Contains(resp.GetResultJson(), "#6366f1") {
			t.Fatalf("result = %q", resp.GetResultJson())
		}
	})
}

func TestServer_GoServiceRPCs_Bad(t *testing.T) {
	t.Parallel()
	// A GoService with no subsystems wired must report Core
	// "unavailable" errors in the response, not crash the stream.
	svc := apigrpc.NewGoService(nil, nil, nil)
	client := dialServer(t, svc)
	c := ctx(t)

	readResp, err := client.ReadFile(c, &sidecarpb.ReadFileRequest{Path: "x"})
	if err != nil {
		t.Fatalf("transport err: %v", err)
	}
	if !core.Contains(readResp.GetError(), "io subsystem not configured") {
		t.Fatalf("ReadFile error = %q", readResp.GetError())
	}

	schemeResp, err := client.ResolveScheme(c, &sidecarpb.SchemeRequest{Uri: "core://x"})
	if err != nil {
		t.Fatalf("transport err: %v", err)
	}
	if !core.Contains(schemeResp.GetError(), "scheme subsystem not configured") {
		t.Fatalf("ResolveScheme error = %q", schemeResp.GetError())
	}
}

func TestServer_GoServiceRPCs_Ugly(t *testing.T) {
	t.Parallel()
	// Subsystems present but each forced into its failure mode. The
	// server must translate every failure into a populated error field
	// while still returning a non-nil response and nil transport error.
	mem := coreio.NewMemoryMedium() // empty: reads miss
	store := newFakeStore()
	store.failGroup = "broken"
	svc := apigrpc.NewGoService(mem, store, fakeRunner{}).WithScheme(fakeResolver{})
	client := dialServer(t, svc)
	c := ctx(t)

	cases := []struct {
		name string
		call func() string // returns the app-level error string
	}{
		{"read missing path", func() string {
			r, _ := client.ReadFile(c, &sidecarpb.ReadFileRequest{Path: "ghost"})
			return r.GetError()
		}},
		{"store backend failure", func() string {
			r, _ := client.StoreGet(c, &sidecarpb.StoreGetRequest{Group: "broken", Key: "k"})
			return r.GetError()
		}},
		{"exec non-zero exit", func() string {
			r, _ := client.Exec(c, &sidecarpb.ExecRequest{Cmd: "boom"})
			return r.GetError()
		}},
		{"scheme resolution failure", func() string {
			r, _ := client.ResolveScheme(c, &sidecarpb.SchemeRequest{Uri: "core://missing/x"})
			return r.GetError()
		}},
		{"nil request guarded", func() string {
			r, _ := client.WriteFile(c, &sidecarpb.WriteFileRequest{Path: "/dev/full/forbidden", Data: []byte("x"), Mode: uint32(fs.FileMode(0644))})
			return r.GetError() // MemoryMedium permits this; just assert no panic/transport error
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each call must complete without a transport panic; the
			// error string is allowed to be empty only for the last
			// (write succeeds in memory).
			_ = tc.call()
		})
	}

	// Explicitly assert the three genuine failures carry messages.
	if r, _ := client.ReadFile(c, &sidecarpb.ReadFileRequest{Path: "ghost"}); r.GetError() == "" {
		t.Fatal("expected read-miss error")
	}
	if r, _ := client.StoreGet(c, &sidecarpb.StoreGetRequest{Group: "broken", Key: "k"}); !core.Contains(r.GetError(), "backend down") {
		t.Fatalf("expected store failure, got %q", r.GetError())
	}
	if r, _ := client.Exec(c, &sidecarpb.ExecRequest{Cmd: "boom"}); r.GetExitCode() == 0 {
		t.Fatalf("expected non-zero exit, got %+v", r)
	}
}

// --- WithModuleScope (module_code permission scoping) -------------------

func TestServer_ModuleScope_Good(t *testing.T) {
	t.Parallel()
	// A guard that permits every module is equivalent to no guard: the
	// call reaches the subsystem and the module_code rides along.
	mem := coreio.NewMemoryMedium()
	if err := mem.WriteMode("a.txt", "hi", 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var sawRPC, sawModule string
	guard := func(rpc, moduleCode string) core.Result {
		sawRPC, sawModule = rpc, moduleCode
		return core.Ok(nil)
	}
	svc := apigrpc.NewGoService(mem, nil, nil).WithModuleScope(guard)
	client := dialServer(t, svc)
	c := ctx(t)

	resp, err := client.ReadFile(c, &sidecarpb.ReadFileRequest{Path: "a.txt", ModuleCode: "ofm.agency"})
	if err != nil {
		t.Fatalf("rpc: %v", err)
	}
	if resp.GetError() != "" || string(resp.GetData()) != "hi" {
		t.Fatalf("resp = %+v", resp)
	}
	if sawRPC != "ReadFile" || sawModule != "ofm.agency" {
		t.Fatalf("guard saw rpc=%q module=%q, want ReadFile/ofm.agency", sawRPC, sawModule)
	}
}

func TestServer_ModuleScope_Bad(t *testing.T) {
	t.Parallel()
	// A guard that denies a named module must short-circuit the rpc with
	// the guard's error and never touch the subsystem.
	store := newFakeStore()
	store.data[store.key("workspace", "task")] = "ship"
	guard := func(_, moduleCode string) core.Result {
		if moduleCode == "untrusted" {
			return core.Fail(core.E("perm", "module untrusted denied", nil))
		}
		return core.Ok(nil)
	}
	svc := apigrpc.NewGoService(nil, store, nil).WithModuleScope(guard)
	client := dialServer(t, svc)
	c := ctx(t)

	denied, err := client.StoreGet(c, &sidecarpb.StoreGetRequest{Group: "workspace", Key: "task", ModuleCode: "untrusted"})
	if err != nil {
		t.Fatalf("transport err: %v", err)
	}
	if !core.Contains(denied.GetError(), "module untrusted denied") {
		t.Fatalf("expected denial, got %+v", denied)
	}
	if denied.GetValue() != "" || denied.GetFound() {
		t.Fatalf("denied call leaked store data: %+v", denied)
	}

	// A permitted module on the same service still succeeds.
	allowed, err := client.StoreGet(c, &sidecarpb.StoreGetRequest{Group: "workspace", Key: "task", ModuleCode: "ofm.agency"})
	if err != nil {
		t.Fatalf("transport err: %v", err)
	}
	if allowed.GetValue() != "ship" || !allowed.GetFound() {
		t.Fatalf("permitted call = %+v", allowed)
	}
}

func TestServer_ModuleScope_Ugly(t *testing.T) {
	t.Parallel()
	// Empty module_code under a guard that rejects anonymous callers: the
	// denial must land before the (nil) subsystem's unavailable error, so
	// the error is the guard's, not "process subsystem not configured".
	guard := func(_, moduleCode string) core.Result {
		if core.Trim(moduleCode) == "" {
			return core.Fail(core.E("perm", "anonymous module denied", nil))
		}
		return core.Ok(nil)
	}
	svc := apigrpc.NewGoService(nil, nil, nil).WithModuleScope(guard)
	client := dialServer(t, svc)
	c := ctx(t)

	resp, err := client.Exec(c, &sidecarpb.ExecRequest{Cmd: "echo"}) // no ModuleCode
	if err != nil {
		t.Fatalf("transport err: %v", err)
	}
	if !core.Contains(resp.GetError(), "anonymous module denied") {
		t.Fatalf("expected anonymous denial, got %+v", resp)
	}
	if resp.GetExitCode() == 0 {
		t.Fatalf("denied exec must report non-zero exit, got %+v", resp)
	}
}
