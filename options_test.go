// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	errors "dappco.re/go/api/internal/stdcompat/coreerrors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	quichttp3 "github.com/quic-go/quic-go/http3"
)

func TestWithHTTP3_Good_ConfiguresEngine(t *testing.T) {
	e, err := New(WithHTTP3(" 127.0.0.1:9443 "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !e.http3Enabled {
		t.Fatal("expected HTTP/3 to be enabled")
	}
	if e.http3Addr != "127.0.0.1:9443" {
		t.Fatalf("expected HTTP/3 addr %q, got %q", "127.0.0.1:9443", e.http3Addr)
	}
	if got := http3AltSvcHeader(e.resolvedHTTP3Addr()); got != `h3=":9443"; ma=2592000` {
		t.Fatalf("expected Alt-Svc header for HTTP/3, got %q", got)
	}
}

func TestServeH3_Good_StartsHTTP3Listener(t *testing.T) {
	addr := reserveHTTP3UDPAddr(t)
	serverTLS, clientTLS := testHTTP3TLSConfigs(t)

	e, err := New(WithHTTP3(addr))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- e.ServeH3(ctx, serverTLS)
	}()

	transport := &quichttp3.Transport{TLSClientConfig: clientTLS}
	t.Cleanup(func() { _ = transport.Close() })
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second,
	}

	var lastErr error
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://" + addr + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP/3 health status 200, got %d", resp.StatusCode)
			}
			if resp.ProtoMajor != 3 {
				t.Fatalf("expected HTTP/3 response, got %s", resp.Proto)
			}

			cancel()
			select {
			case serveErr := <-errCh:
				if serveErr != nil {
					t.Fatalf("ServeH3 returned unexpected error: %v", serveErr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("ServeH3 did not return after context cancellation")
			}
			return
		}

		lastErr = err
		select {
		case serveErr := <-errCh:
			t.Fatalf("ServeH3 exited before serving HTTP/3 health check: %v", serveErr)
		default:
		}
		time.Sleep(50 * time.Millisecond)
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(time.Second):
	}
	t.Fatalf("HTTP/3 health request failed before deadline: %v", lastErr)
}

func TestServeH3_Bad_RequiresTLSConfig(t *testing.T) {
	e, err := New(WithHTTP3("127.0.0.1:9443"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = e.ServeH3(context.Background(), nil)
	if !errors.Is(err, ErrHTTP3TLSRequired) {
		t.Fatalf("expected ErrHTTP3TLSRequired, got %v", err)
	}
}

func TestWithHTTP3_Good_AdvertisesAltSvcOnHTTP2(t *testing.T) {
	e, err := New(WithHTTP3("127.0.0.1:9443"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv := httptest.NewUnstartedServer(e.Handler())
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("HTTP/2 health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.ProtoMajor != 2 {
		t.Fatalf("expected HTTP/2 response, got %s", resp.Proto)
	}
	if got := resp.Header.Get("Alt-Svc"); got != `h3=":9443"; ma=2592000` {
		t.Fatalf("expected HTTP/3 Alt-Svc header, got %q", got)
	}
}

func reserveHTTP3UDPAddr(t *testing.T) string {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to reserve UDP address: %v", err)
	}
	addr := conn.LocalAddr().String()
	if err := conn.Close(); err != nil {
		t.Fatalf("failed to release UDP address: %v", err)
	}
	return addr
}

func testHTTP3TLSConfigs(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate HTTP/3 test key: %v", err)
	}

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, pub, priv)
	if err != nil {
		t.Fatalf("failed to generate HTTP/3 test certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("failed to parse HTTP/3 test certificate: %v", err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(cert)

	return &tls.Config{
			Certificates: []tls.Certificate{{
				Certificate: [][]byte{certDER},
				PrivateKey:  priv,
			}},
			MinVersion: tls.VersionTLS13,
		}, &tls.Config{
			RootCAs:    roots,
			MinVersion: tls.VersionTLS13,
		}
}
