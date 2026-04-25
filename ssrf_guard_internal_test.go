// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"errors"
	"net"
	"strings"
	"testing"
)

// TestSSRF_OutboundURL_BlocksMetadata_Ugly — Cerberus mechanism review
// recommendation per Mantis #318. AWS/GCP/Azure metadata endpoints must be
// rejected by literal-host match before DNS resolution.
func TestSSRF_OutboundURL_BlocksMetadata_Ugly(t *testing.T) {
	cases := []string{
		"http://169.254.169.254/latest/meta-data/iam/security-credentials/",
		"https://metadata.google.internal/computeMetadata/v1/instance/",
		"http://metadata.azure.com/",
		"http://[fd00:ec2::254]/",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			err := validateOutboundURL(raw)
			if err == nil {
				t.Errorf("validateOutboundURL(%q) returned nil; expected block", raw)
				return
			}
			if !errors.Is(err, errOutboundURLBlocked) {
				t.Errorf("expected errors.Is(err, errOutboundURLBlocked) for %q; got %v", raw, err)
			}
		})
	}
}

// TestSSRF_OutboundURL_BlocksLoopback_Ugly — localhost variants.
func TestSSRF_OutboundURL_BlocksLoopback_Ugly(t *testing.T) {
	cases := []string{
		"http://127.0.0.1/",
		"http://127.5.5.5/",
		"http://[::1]/",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			err := validateOutboundURL(raw)
			if err == nil {
				t.Errorf("validateOutboundURL(%q) returned nil; expected loopback block", raw)
				return
			}
			if !errors.Is(err, errOutboundURLBlocked) {
				t.Errorf("expected errors.Is(err, errOutboundURLBlocked); got %v", err)
			}
		})
	}
}

// TestSSRF_OutboundURL_BlocksRFC1918_Ugly — internal-network IP ranges.
func TestSSRF_OutboundURL_BlocksRFC1918_Ugly(t *testing.T) {
	cases := []string{
		"http://10.0.0.1/",
		"http://10.255.255.255/",
		"http://172.16.0.1/",
		"http://172.31.255.255/",
		"http://192.168.1.1/",
		"http://192.168.255.255/",
		"http://[fc00::1]/", // IPv6 ULA
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			err := validateOutboundURL(raw)
			if err == nil {
				t.Errorf("validateOutboundURL(%q) returned nil; expected RFC1918/ULA block", raw)
				return
			}
			if !errors.Is(err, errOutboundURLBlocked) {
				t.Errorf("expected errors.Is(err, errOutboundURLBlocked); got %v", err)
			}
		})
	}
}

// TestSSRF_OutboundURL_BlocksDisallowedScheme_Bad — non-http(s) schemes.
func TestSSRF_OutboundURL_BlocksDisallowedScheme_Bad(t *testing.T) {
	cases := []string{
		"file:///etc/passwd",
		"gopher://evil.example.com/_command",
		"ftp://example.com/",
		"dict://example.com:11211/stat",
		"ldap://example.com/",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			err := validateOutboundURL(raw)
			if err == nil {
				t.Errorf("validateOutboundURL(%q) returned nil; expected scheme block", raw)
				return
			}
			if !strings.Contains(err.Error(), "disallowed scheme") {
				t.Errorf("expected 'disallowed scheme' error; got %v", err)
			}
		})
	}
}

// TestSSRF_OutboundURL_AllowsHTTPS_Good — sanity that public HTTPS still works.
// We override resolveHost to return a public IP so we don't depend on real DNS.
func TestSSRF_OutboundURL_AllowsHTTPS_Good(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		// Pretend example.com resolves to a public IP.
		return []net.IP{net.IPv4(93, 184, 216, 34)}, nil
	}

	cases := []string{
		"https://example.com/",
		"https://example.com/path?q=1",
		"http://example.com:8080/",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			if err := validateOutboundURL(raw); err != nil {
				t.Errorf("validateOutboundURL(%q) blocked unexpectedly: %v", raw, err)
			}
		})
	}
}

// TestSSRF_OutboundURL_BlocksDNSResolveToPrivate_Ugly — DNS-rebinding-style:
// a public-looking hostname that resolves to an RFC1918 IP must still be
// blocked by the post-resolution check.
func TestSSRF_OutboundURL_BlocksDNSResolveToPrivate_Ugly(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		// Attacker's domain that resolves to a private IP.
		return []net.IP{net.IPv4(10, 0, 0, 5)}, nil
	}

	err := validateOutboundURL("https://attacker.example.com/")
	if err == nil {
		t.Fatal("expected post-resolution private-IP block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
	if !strings.Contains(err.Error(), "10.0.0.5") {
		t.Errorf("expected error to mention resolved IP; got %v", err)
	}
}

// TestSSRF_OutboundURL_EmptyURL_Bad — defensive case.
func TestSSRF_OutboundURL_EmptyURL_Bad(t *testing.T) {
	err := validateOutboundURL("")
	if err == nil {
		t.Fatal("expected empty-URL block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
}

// TestSSRF_OutboundURL_AllowsResolverFailure_Good — if DNS resolution fails,
// let net/http surface the real error rather than masking as a security block.
func TestSSRF_OutboundURL_AllowsResolverFailure_Good(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		return nil, errors.New("simulated NXDOMAIN")
	}

	if err := validateOutboundURL("https://nonexistent.example.invalid/"); err != nil {
		t.Errorf("expected nil (let net/http surface the error); got %v", err)
	}
}
