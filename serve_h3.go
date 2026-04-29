// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"crypto/tls"
	errors "dappco.re/go/api/internal/stdcompat/coreerrors"
	"net"
	"net/http"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
	quichttp3 "github.com/quic-go/quic-go/http3"
)

var (
	// ErrHTTP3NotConfigured is returned when ServeH3 is called without
	// enabling HTTP/3 via WithHTTP3.
	ErrHTTP3NotConfigured = errors.New("api: HTTP/3 is not configured")

	// ErrHTTP3TLSRequired is returned when ServeH3 is called without TLS.
	ErrHTTP3TLSRequired = errors.New("api: HTTP/3 requires TLS configuration")

	// ErrNilContext is returned when ServeH3 is called with a nil context.
	ErrNilContext = errors.New("api: context is nil")
)

// ServeH3 starts the HTTP/3 QUIC server and blocks until the context is
// cancelled, then performs a graceful shutdown.
//
// ServeH3 is intentionally separate from Serve so callers can run the QUIC
// listener alongside their existing HTTP/1.1+2 server with an explicit TLS
// configuration.
func (e *Engine) ServeH3(ctx context.Context, tlsConfig *tls.Config) (
	_ error,
) {
	if e == nil || !e.http3Enabled {
		return ErrHTTP3NotConfigured
	}
	if ctx == nil {
		return ErrNilContext
	}
	if tlsConfig == nil {
		return ErrHTTP3TLSRequired
	}

	srv := &quichttp3.Server{
		Addr:        e.resolvedHTTP3Addr(),
		TLSConfig:   tlsConfig,
		Handler:     e.build(),
		IdleTimeout: serverIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	if e.sseBroker != nil {
		e.sseBroker.Drain()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return <-errCh
}

func (e *Engine) resolvedHTTP3Addr() string {
	if e == nil {
		return ""
	}
	if e.http3Addr != "" {
		return e.http3Addr
	}
	return e.addr
}

func http3AltSvcMiddleware(altSvc string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Alt-Svc", altSvc)
		c.Next()
	}
}

func http3AltSvcHeader(addr string) string {
	port, ok := http3AltSvcPort(addr)
	if !ok {
		return ""
	}
	return core.Sprintf(`%s=":%d"; ma=2592000`, quichttp3.NextProtoH3, port)
}

func http3AltSvcPort(addr string) (int, bool) {
	if addr == "" {
		return 0, false
	}

	_, portPart, err := net.SplitHostPort(addr)
	if err != nil {
		portPart = addr
	}

	port, err := net.LookupPort("tcp", portPart)
	if err != nil || port <= 0 {
		return 0, false
	}
	return port, true
}
