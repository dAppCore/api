// SPDX-License-Identifier: EUPL-1.2

package grpc

import (
	"crypto/tls"
	"net"
	"sync"

	core "dappco.re/go"
	sidecarpb "dappco.re/go/api/pkg/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const serverScope = "grpc.Server"

// Default transport: a Unix domain socket is preferred in production
// (faster than TCP loopback per RFC.grpc.md §5). When no socket path
// is configured the server falls back to TCP on loopback.
const (
	defaultGRPCHost = "127.0.0.1"
	defaultGRPCPort = 50051
)

// GRPCServer hosts the GoService for Deno sidecar communication. It
// owns the underlying *grpc.Server and the listener (Unix socket or
// TCP loopback) it serves on.
//
//	srv, _ := grpc.NewGRPCServer(
//	    grpc.WithGRPCPort(50051),
//	    grpc.WithGRPCServices(grpc.NewGoService(io.Local, kv, proc)),
//	)
//	go srv.Serve()
//	defer srv.Stop()
type GRPCServer struct {
	host       string
	port       int
	socketPath string
	tlsConfig  *tls.Config
	goServices []*GoService

	server   *grpc.Server
	listener net.Listener
	mu       sync.Mutex
}

// GRPCOption configures a GRPCServer during construction.
//
//	srv, _ := grpc.NewGRPCServer(grpc.WithGRPCPort(50051))
type GRPCOption func(*GRPCServer)

// WithGRPCPort sets the TCP loopback port used when no Unix socket is
// configured.
//
//	grpc.NewGRPCServer(grpc.WithGRPCPort(50051))
func WithGRPCPort(port int) GRPCOption {
	return func(s *GRPCServer) {
		s.port = port
	}
}

// WithGRPCHost overrides the loopback host (default 127.0.0.1). The
// address must remain loopback unless TLS is configured.
//
//	grpc.NewGRPCServer(grpc.WithGRPCHost("127.0.0.1"))
func WithGRPCHost(host string) GRPCOption {
	return func(s *GRPCServer) {
		s.host = host
	}
}

// WithGRPCSocket serves over a Unix domain socket at the given path
// instead of TCP. This is the preferred production transport.
//
//	grpc.NewGRPCServer(grpc.WithGRPCSocket("/run/core/sidecar.sock"))
func WithGRPCSocket(path string) GRPCOption {
	return func(s *GRPCServer) {
		s.socketPath = path
	}
}

// WithGRPCServices registers one or more GoService implementations.
//
//	grpc.NewGRPCServer(grpc.WithGRPCServices(grpc.NewGoService(io.Local, kv, proc)))
func WithGRPCServices(services ...*GoService) GRPCOption {
	return func(s *GRPCServer) {
		for _, svc := range services {
			if svc != nil {
				s.goServices = append(s.goServices, svc)
			}
		}
	}
}

// WithGRPCTLS enables TLS for the listener. TLS is optional for
// localhost sidecar communication (RFC.grpc.md §5).
//
//	grpc.NewGRPCServer(grpc.WithGRPCTLS(cfg))
func WithGRPCTLS(cfg *tls.Config) GRPCOption {
	return func(s *GRPCServer) {
		s.tlsConfig = cfg
	}
}

// NewGRPCServer builds and binds a GRPCServer. It opens the listener
// (Unix socket if WithGRPCSocket was set, else TCP loopback) and
// registers every configured GoService on the underlying grpc.Server.
// Call Serve to begin accepting; Stop to shut down gracefully.
//
//	srv, err := grpc.NewGRPCServer(
//	    grpc.WithGRPCPort(50051),
//	    grpc.WithGRPCServices(grpc.NewGoService(io.Local, kv, proc)),
//	)
//	defer srv.Stop()
func NewGRPCServer(opts ...GRPCOption) (*GRPCServer, error) {
	s := &GRPCServer{
		host: defaultGRPCHost,
		port: defaultGRPCPort,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	listener, err := s.listen()
	if err != nil {
		return nil, err
	}

	var serverOpts []grpc.ServerOption
	if s.tlsConfig != nil {
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(s.tlsConfig)))
	}

	s.server = grpc.NewServer(serverOpts...)
	for _, svc := range s.goServices {
		sidecarpb.RegisterGoServiceServer(s.server, svc)
	}
	s.listener = listener
	return s, nil
}

// listen opens the configured transport. A Unix socket path takes
// precedence; otherwise TCP loopback on host:port is used.
func (s *GRPCServer) listen() (net.Listener, error) {
	if core.Trim(s.socketPath) != "" {
		// A stale socket file blocks bind; remove it best-effort.
		_ = core.Remove(s.socketPath)
		listener, err := net.Listen("unix", s.socketPath)
		if err != nil {
			return nil, core.E(serverScope, core.Concat("unix listen failed: ", s.socketPath), err)
		}
		return listener, nil
	}
	addr := core.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, core.E(serverScope, core.Concat("tcp listen failed: ", addr), err)
	}
	return listener, nil
}

// Address returns the address the server is bound to: the socket path
// for Unix transport, or the host:port for TCP.
//
//	addr := srv.Address() // "127.0.0.1:50051" or "/run/core/sidecar.sock"
func (s *GRPCServer) Address() string {
	if core.Trim(s.socketPath) != "" {
		return s.socketPath
	}
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return core.Sprintf("%s:%d", s.host, s.port)
}

// Serve begins accepting connections and blocks until Stop is called
// or the listener errors. Run it in its own goroutine.
//
//	go srv.Serve()
func (s *GRPCServer) Serve() error {
	s.mu.Lock()
	server := s.server
	listener := s.listener
	s.mu.Unlock()
	if server == nil || listener == nil {
		return core.E(serverScope, "server not initialised", nil)
	}
	if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
		return core.E(serverScope, "serve failed", err)
	}
	return nil
}

// Stop gracefully stops the server and releases the listener. A nil or
// already-stopped server is a no-op, so Stop is safe in a defer.
//
//	defer srv.Stop()
func (s *GRPCServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil {
		s.server.GracefulStop()
	}
	if core.Trim(s.socketPath) != "" {
		_ = core.Remove(s.socketPath)
	}
}
