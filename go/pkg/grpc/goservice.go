// SPDX-License-Identifier: EUPL-1.2

// Package grpc implements the CoreGO side of the CoreGO ↔ CoreDeno
// sidecar contract defined in code/core/go/api/RFC.grpc.md.
//
// Go hosts a gRPC server (GoService) that Deno calls for sandboxed
// I/O, KV state, process execution, and core:// scheme resolution.
// Go also acts as a client (DenoClient) to drive the Deno TypeScript
// runtime's lifecycle and rendering.
//
// The service wires to real Core subsystems through narrow consumer-
// side interfaces (IOMedium, KVStore, ProcRunner, SchemeResolver) so
// the api module never hard-imports go-store. The concrete types from
// go-io (io.Medium), go-store (*store.Store), and go-process
// (*process.Service) satisfy these interfaces structurally at the
// call site. This keeps the dependency arrow pointing the right way
// (AX principle 8: lib never imports consumer).
//
//	svc := grpc.NewGoService(io.Local, kvStore, procRunner)
//	srv, _ := grpc.NewGRPCServer(
//	    grpc.WithGRPCPort(50051),
//	    grpc.WithGRPCServices(svc),
//	)
//	defer srv.Stop()
package grpc

import (
	"context"
	"io/fs"

	core "dappco.re/go"
	sidecarpb "dappco.re/go/api/pkg/proto/gen"
)

// IOMedium is the file-I/O surface the GoService consumes. It is the
// subset of go-io's Medium interface the sidecar needs, so api never
// imports go-io directly. io.Medium satisfies this structurally.
//
//	var m IOMedium = io.Local
type IOMedium interface {
	// Read returns the full content of a sandbox-relative path.
	Read(path string) (string, error)
	// WriteMode writes content at the given file mode.
	WriteMode(path, content string, mode fs.FileMode) error
	// List returns directory entries under a sandbox-relative path.
	List(path string) ([]fs.DirEntry, error)
}

// KVStore is the key/value surface the GoService consumes. It matches
// go-store's *Store (group, key) shape returning a core.Result, so api
// never imports go-store directly. *store.Store satisfies it.
//
//	var s KVStore = storeInstance
type KVStore interface {
	// Get returns the value for (group, key). The Result is failed
	// when the backend errors; an absent key yields ("", ok-Result).
	Get(group, key string) (string, core.Result)
	// Set writes value under (group, key).
	Set(group, key, value string) core.Result
}

// ProcRunner is the process-execution surface the GoService consumes.
// It matches go-process's Service.RunWithOptions contract returning a
// core.Result whose Value is the captured output string on success.
//
//	var r ProcRunner = processService
type ProcRunner interface {
	// RunWithOptions executes a command and blocks until it exits.
	// On success the Result.Value is the combined output (string).
	RunWithOptions(ctx context.Context, opts RunOptions) core.Result
}

// RunOptions describes a process to execute. It mirrors the fields the
// sidecar exposes over the wire; the gateway adapts it to the concrete
// go-process RunOptions.
type RunOptions struct {
	// Command is the executable to run.
	Command string
	// Args are the command arguments.
	Args []string
	// Dir is the optional working directory.
	Dir string
	// Env carries KEY=VALUE environment overrides.
	Env []string
}

// SchemeResolver resolves a core:// URL to a value. It matches the
// SchemeRegistry.Resolve contract from RFC.core-scheme.md §4 so the
// GoService can answer ResolveScheme without importing the registry
// implementation directly.
//
//	var r SchemeResolver = registry
type SchemeResolver interface {
	// Resolve parses a core:// URL and returns the resolution value.
	Resolve(url string) (any, error)
}

// ModuleGuard authorises a single GoService rpc on behalf of the module
// that initiated it (the request's module_code). It returns a failed
// core.Result to deny the call; the rpc then reports that error in its
// response Error field instead of touching a subsystem. A nil guard
// allows every call, so module_code is recorded but not enforced.
//
//	guard := func(rpc, moduleCode string) core.Result {
//	    if moduleCode == "" {
//	        return core.Fail(core.E("perm", "anonymous call denied", nil))
//	    }
//	    return core.Ok(nil)
//	}
type ModuleGuard func(rpc, moduleCode string) core.Result

// GoService implements the sidecarpb.GoServiceServer rpcs against the
// injected Core subsystems. Any subsystem may be nil; the matching
// rpcs then report a Core "unavailable" error rather than panicking.
//
// Every GoService request carries a module_code naming the module that
// initiated the call. The optional ModuleGuard (see WithModuleScope)
// turns that field into a per-call permission check; without a guard
// the field is passed through for downstream auditing only.
//
//	svc := grpc.NewGoService(io.Local, kvStore, procRunner)
//	svc.WithScheme(registry)
type GoService struct {
	sidecarpb.UnimplementedGoServiceServer

	medium   IOMedium
	store    KVStore
	runner   ProcRunner
	resolver SchemeResolver
	guard    ModuleGuard
}

const goServiceScope = "grpc.GoService"

// NewGoService constructs a GoService wired to the given subsystems.
// Pass nil for any subsystem the deployment does not expose.
//
//	svc := grpc.NewGoService(io.Local, kvStore, procRunner)
func NewGoService(medium IOMedium, store KVStore, runner ProcRunner) *GoService {
	return &GoService{
		medium: medium,
		store:  store,
		runner: runner,
	}
}

// WithScheme attaches a core:// scheme resolver and returns the
// service for chaining. ResolveScheme reports unavailable until set.
//
//	svc := grpc.NewGoService(io.Local, kv, proc).WithScheme(registry)
func (s *GoService) WithScheme(resolver SchemeResolver) *GoService {
	s.resolver = resolver
	return s
}

// WithModuleScope attaches a ModuleGuard so each rpc is authorised
// against the request's module_code before reaching a subsystem, and
// returns the service for chaining. Without a guard, module_code is
// accepted and passed through but never blocks a call.
//
//	svc := grpc.NewGoService(io.Local, kv, proc).WithModuleScope(guard)
func (s *GoService) WithModuleScope(guard ModuleGuard) *GoService {
	s.guard = guard
	return s
}

// scope runs the ModuleGuard for rpc on behalf of moduleCode. A nil
// guard is an allow-all pass-through, so the returned Result is OK.
//
//	if r := s.scope("ReadFile", req.GetModuleCode()); !r.OK { ... }
func (s *GoService) scope(rpc, moduleCode string) core.Result {
	if s.guard == nil {
		return core.Ok(nil)
	}
	return s.guard(rpc, moduleCode)
}

// ReadFile reads a sandbox-relative path through the go-io Medium.
//
//	resp, _ := svc.ReadFile(ctx, &sidecarpb.ReadFileRequest{Path: "config/app.yaml"})
func (s *GoService) ReadFile(_ context.Context, req *sidecarpb.ReadFileRequest) (*sidecarpb.ReadFileResponse, error) {
	if req == nil {
		return &sidecarpb.ReadFileResponse{Error: "nil request"}, nil
	}
	if r := s.scope("ReadFile", req.GetModuleCode()); !r.OK {
		return &sidecarpb.ReadFileResponse{Error: r.Error()}, nil
	}
	if s.medium == nil {
		return &sidecarpb.ReadFileResponse{Error: unavailable("io")}, nil
	}
	content, err := s.medium.Read(req.GetPath())
	if err != nil {
		return &sidecarpb.ReadFileResponse{Error: err.Error()}, nil
	}
	return &sidecarpb.ReadFileResponse{Data: []byte(content)}, nil
}

// WriteFile writes content to a sandbox-relative path. A zero mode
// selects 0644, the Medium default.
//
//	resp, _ := svc.WriteFile(ctx, &sidecarpb.WriteFileRequest{Path: "a.txt", Data: []byte("hi")})
func (s *GoService) WriteFile(_ context.Context, req *sidecarpb.WriteFileRequest) (*sidecarpb.WriteFileResponse, error) {
	if req == nil {
		return &sidecarpb.WriteFileResponse{Error: "nil request"}, nil
	}
	if r := s.scope("WriteFile", req.GetModuleCode()); !r.OK {
		return &sidecarpb.WriteFileResponse{Error: r.Error()}, nil
	}
	if s.medium == nil {
		return &sidecarpb.WriteFileResponse{Error: unavailable("io")}, nil
	}
	mode := fs.FileMode(req.GetMode())
	if mode == 0 {
		mode = 0644
	}
	if err := s.medium.WriteMode(req.GetPath(), string(req.GetData()), mode); err != nil {
		return &sidecarpb.WriteFileResponse{Error: err.Error()}, nil
	}
	return &sidecarpb.WriteFileResponse{Ok: true}, nil
}

// ListFiles lists entries under a sandbox-relative directory.
//
//	resp, _ := svc.ListFiles(ctx, &sidecarpb.ListFilesRequest{Path: "config"})
func (s *GoService) ListFiles(_ context.Context, req *sidecarpb.ListFilesRequest) (*sidecarpb.ListFilesResponse, error) {
	if req == nil {
		return &sidecarpb.ListFilesResponse{Error: "nil request"}, nil
	}
	if r := s.scope("ListFiles", req.GetModuleCode()); !r.OK {
		return &sidecarpb.ListFilesResponse{Error: r.Error()}, nil
	}
	if s.medium == nil {
		return &sidecarpb.ListFilesResponse{Error: unavailable("io")}, nil
	}
	dirEntries, err := s.medium.List(req.GetPath())
	if err != nil {
		return &sidecarpb.ListFilesResponse{Error: err.Error()}, nil
	}
	entries := make([]*sidecarpb.FileEntry, 0, len(dirEntries))
	for _, entry := range dirEntries {
		entries = append(entries, &sidecarpb.FileEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}
	return &sidecarpb.ListFilesResponse{Entries: entries}, nil
}

// StoreGet reads a value by (group, key) through go-store.
//
//	resp, _ := svc.StoreGet(ctx, &sidecarpb.StoreGetRequest{Group: "workspace", Key: "task"})
func (s *GoService) StoreGet(_ context.Context, req *sidecarpb.StoreGetRequest) (*sidecarpb.StoreGetResponse, error) {
	if req == nil {
		return &sidecarpb.StoreGetResponse{Error: "nil request"}, nil
	}
	if r := s.scope("StoreGet", req.GetModuleCode()); !r.OK {
		return &sidecarpb.StoreGetResponse{Error: r.Error()}, nil
	}
	if s.store == nil {
		return &sidecarpb.StoreGetResponse{Error: unavailable("store")}, nil
	}
	value, result := s.store.Get(req.GetGroup(), req.GetKey())
	if !result.OK {
		return &sidecarpb.StoreGetResponse{Error: result.Error()}, nil
	}
	return &sidecarpb.StoreGetResponse{Value: value, Found: value != ""}, nil
}

// StoreSet writes value under (group, key) through go-store.
//
//	resp, _ := svc.StoreSet(ctx, &sidecarpb.StoreSetRequest{Group: "workspace", Key: "task", Value: "x"})
func (s *GoService) StoreSet(_ context.Context, req *sidecarpb.StoreSetRequest) (*sidecarpb.StoreSetResponse, error) {
	if req == nil {
		return &sidecarpb.StoreSetResponse{Error: "nil request"}, nil
	}
	if r := s.scope("StoreSet", req.GetModuleCode()); !r.OK {
		return &sidecarpb.StoreSetResponse{Error: r.Error()}, nil
	}
	if s.store == nil {
		return &sidecarpb.StoreSetResponse{Error: unavailable("store")}, nil
	}
	result := s.store.Set(req.GetGroup(), req.GetKey(), req.GetValue())
	if !result.OK {
		return &sidecarpb.StoreSetResponse{Error: result.Error()}, nil
	}
	return &sidecarpb.StoreSetResponse{Ok: true}, nil
}

// Exec runs a command through go-process and returns its output.
//
//	resp, _ := svc.Exec(ctx, &sidecarpb.ExecRequest{Cmd: "echo", Args: []string{"hi"}})
func (s *GoService) Exec(ctx context.Context, req *sidecarpb.ExecRequest) (*sidecarpb.ExecResponse, error) {
	if req == nil {
		return &sidecarpb.ExecResponse{Error: "nil request"}, nil
	}
	if r := s.scope("Exec", req.GetModuleCode()); !r.OK {
		return &sidecarpb.ExecResponse{ExitCode: 1, Error: r.Error()}, nil
	}
	if s.runner == nil {
		return &sidecarpb.ExecResponse{Error: unavailable("process")}, nil
	}
	result := s.runner.RunWithOptions(ctx, RunOptions{
		Command: req.GetCmd(),
		Args:    req.GetArgs(),
		Dir:     req.GetDir(),
		Env:     req.GetEnv(),
	})
	if !result.OK {
		// A non-zero exit still carries output on the failed Result's
		// scope; surface the error message and a non-zero code.
		return &sidecarpb.ExecResponse{ExitCode: 1, Error: result.Error()}, nil
	}
	output, _ := result.Value.(string)
	return &sidecarpb.ExecResponse{Output: []byte(output)}, nil
}

// ResolveScheme resolves a core:// URL through the attached registry,
// JSON-encoding the result for the wire.
//
//	resp, _ := svc.ResolveScheme(ctx, &sidecarpb.SchemeRequest{Uri: "core://settings/theme.accent"})
func (s *GoService) ResolveScheme(_ context.Context, req *sidecarpb.SchemeRequest) (*sidecarpb.SchemeResponse, error) {
	if req == nil {
		return &sidecarpb.SchemeResponse{Error: "nil request"}, nil
	}
	if r := s.scope("ResolveScheme", req.GetModuleCode()); !r.OK {
		return &sidecarpb.SchemeResponse{Error: r.Error()}, nil
	}
	if s.resolver == nil {
		return &sidecarpb.SchemeResponse{Error: unavailable("scheme")}, nil
	}
	value, err := s.resolver.Resolve(req.GetUri())
	if err != nil {
		return &sidecarpb.SchemeResponse{Error: err.Error()}, nil
	}
	return &sidecarpb.SchemeResponse{ResultJson: core.JSONMarshalString(value)}, nil
}

// unavailable formats the Core error message for a missing subsystem.
func unavailable(subsystem string) string {
	return core.E(goServiceScope, core.Concat(subsystem, " subsystem not configured"), nil).Error()
}

// Static assertion: GoService implements the generated server.
var _ sidecarpb.GoServiceServer = (*GoService)(nil)
