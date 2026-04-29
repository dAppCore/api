// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"dappco.re/go/api/internal/stdcompat/errors"
	"dappco.re/go/api/internal/stdcompat/strings"
	"net"
	"testing"
)

// TestSSRFBlocksMetadata — Cerberus mechanism review
// recommendation per Mantis #318. AWS/GCP/Azure metadata endpoints must be
// rejected by literal-host match before DNS resolution.
func TestSSRFBlocksMetadata(t *testing.T) {
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

// TestSSRFBlocksLoopback — localhost variants.
func TestSSRFBlocksLoopback(t *testing.T) {
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

// TestSSRFBlocksRFC1918 — internal-network IP ranges.
func TestSSRFBlocksRFC1918(t *testing.T) {
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

// TestSSRFBlocksDisallowedScheme — non-http(s) schemes.
func TestSSRFBlocksDisallowedScheme(t *testing.T) {
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

// TestSSRFBlocksEmbeddedCredentials — URL userinfo can leak
// into logs/proxies and is rejected at the outbound boundary.
func TestSSRFBlocksEmbeddedCredentials(t *testing.T) {
	badCases := []string{
		"https://user:pass@example.com/path",
		"https://user@example.com/path",
	}
	for _, raw := range badCases {
		t.Run(raw, func(t *testing.T) {
			err := validateOutboundURL(raw)
			if err == nil {
				t.Errorf("validateOutboundURL(%q) returned nil; expected credential block", raw)
				return
			}
			if !errors.Is(err, errOutboundURLBlocked) {
				t.Errorf("expected errOutboundURLBlocked; got %v", err)
			}
			if !strings.Contains(err.Error(), "URL contains embedded credentials") {
				t.Errorf("expected embedded credentials error; got %v", err)
			}
		})
	}

	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		return []net.IP{net.IPv4(93, 184, 216, 34)}, nil
	}
	if err := validateOutboundURL("https://example.com/path"); err != nil {
		t.Errorf("validateOutboundURL(%q) blocked unexpectedly: %v", "https://example.com/path", err)
	}
}

// TestSSRFAllowsHTTPS — sanity that public HTTPS still works.
// We override resolveHost to return a public IP so we don't depend on real DNS.
func TestSSRFAllowsHTTPS(t *testing.T) {
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

// TestSSRFBlocksDNSResolveToPrivate — DNS-rebinding-style:
// a public-looking hostname that resolves to an RFC1918 IP must still be
// blocked by the post-resolution check.
func TestSSRFBlocksDNSResolveToPrivate(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		// Attacker's domain that resolves to a private IP.
		return []net.IP{net.IPv4(10, 0, 0, 1)}, nil
	}

	err := validateOutboundURL("https://attacker.example.com/")
	if err == nil {
		t.Fatal("expected post-resolution private-IP block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
	if !strings.Contains(err.Error(), "10.0.0.1") {
		t.Errorf("expected error to mention resolved IP; got %v", err)
	}
}

// TestSSRFEmptyURL — defensive case.
func TestSSRFEmptyURL(t *testing.T) {
	err := validateOutboundURL("")
	if err == nil {
		t.Fatal("expected empty-URL block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
}

// TestSSRFBlocksResolverFailure — DNS resolution failure must
// fail closed so split-resolver mismatches cannot bypass the IP blocklist.
func TestSSRFBlocksResolverFailure(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		return nil, errors.New("DNS failure")
	}

	err := validateOutboundURL("https://nonexistent.example.invalid/")
	if err == nil {
		t.Fatal("expected resolver failure to block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
	if !strings.Contains(err.Error(), "DNS failure") {
		t.Errorf("expected error to mention DNS failure; got %v", err)
	}
}

// TestSSRFBlocksEmptyResolverResult — an empty DNS answer is
// equivalent to no usable IP for SSRF validation and must fail closed.
func TestSSRFBlocksEmptyResolverResult(t *testing.T) {
	prev := resolveHost
	defer func() { resolveHost = prev }()
	resolveHost = func(host string) ([]net.IP, error) {
		return []net.IP{}, nil
	}

	err := validateOutboundURL("https://empty.example.invalid/")
	if err == nil {
		t.Fatal("expected empty resolver result to block; got nil")
	}
	if !errors.Is(err, errOutboundURLBlocked) {
		t.Errorf("expected errOutboundURLBlocked; got %v", err)
	}
	if !strings.Contains(err.Error(), "no IPs") {
		t.Errorf("expected error to mention empty DNS result; got %v", err)
	}
}
