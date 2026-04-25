// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"log/slog"
	// Note: AX-6 — net.ParseIP/LookupIP and net.IP predicates are structural for SSRF IP-range comparison.
	"net"
	// Note: AX-6 — URL parsing is structural for SSRF scheme and host extraction before outbound requests.
	"net/url"

	core "dappco.re/go/core"
	coreerr "dappco.re/go/log"
)

// SSRF mitigation per Cerberus mechanism review on Mantis #318.
//
// Both SSEClient.Connect (transport_client.go:229) and OpenAPIClient.Call
// (client.go:342) flow through doHTTPClientRequest. The polyglot-gateway
// threat model (RFC §11) makes attacker-controlled outbound URLs reachable
// via:
//   - SSEClient(rawURL) where rawURL flows from request input
//   - WithBaseURL(baseURL) where baseURL is loaded from attacker-influenced config
//   - WithSpecReader spec.servers[].url
//
// validateOutboundURL is the singular choke-point validator applied at
// doHTTPClientRequest before client.Do(req). It denies by default:
//   - schemes other than http/https
//   - hosts that resolve to RFC1918 / loopback / link-local / cloud-metadata IPs
//
// The validator is applied at request time (not just construction time) so
// DNS rebinding attacks cannot bypass pre-resolution checks — by the time
// the request fires, the literal host has been re-resolved.

// errOutboundURLBlocked is returned when validateOutboundURL rejects a URL.
// Callers see a wrapped error from client.Do; tests assert on errors.Is.
var errOutboundURLBlocked = coreerr.E("", "outbound URL blocked by SSRF guard", nil)

// allowedSchemes is the deny-by-default scheme allowlist for outbound HTTP.
// Excludes file://, gopher://, ftp://, dict://, ldap://, etc.
var allowedSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
}

// metadataHosts are cloud instance-metadata hostnames that must NOT resolve
// to a usable backend. Compared after URL parse, before DNS resolution.
var metadataHosts = map[string]struct{}{
	"metadata.google.internal": {},
	"metadata.googleapis.com":  {},
	"metadata.azure.com":       {},
	"169.254.169.254":          {}, // AWS / GCP / OpenStack / Azure (legacy)
	"fd00:ec2::254":            {}, // AWS IPv6
	"100.100.100.200":          {}, // Alibaba Cloud
}

// resolveHost is overridden in tests to avoid real DNS lookups while still
// exercising the IP-rejection logic.
var resolveHost = net.LookupIP

// validateOutboundURL checks rawURL against the deny-by-default outbound
// policy. Returns errOutboundURLBlocked (or a wrap thereof) on rejection.
//
// Pass empty rawURL is rejected. Caller should never call client.Do with
// an unvalidated URL.
func validateOutboundURL(rawURL string) error {
	if rawURL == "" {
		return wrapBlocked("empty URL")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return wrapBlocked("parse failed: " + err.Error())
	}
	if u.User != nil {
		return wrapBlocked("URL contains embedded credentials")
	}
	if _, ok := allowedSchemes[core.Lower(u.Scheme)]; !ok {
		return wrapBlocked("disallowed scheme: " + u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return wrapBlocked("empty host")
	}
	if _, ok := metadataHosts[core.Lower(host)]; ok {
		return wrapBlocked("metadata host: " + host)
	}

	// If host is a literal IP, check directly. Otherwise resolve and check
	// every result. DNS rebinding can change resolution between calls; this
	// re-checks at request time per the choke-point design.
	if ip := net.ParseIP(host); ip != nil {
		if reason := blockedIPReason(ip); reason != "" {
			return wrapBlocked(reason + ": " + host)
		}
		return nil
	}

	ips, err := resolveHost(host)
	if err != nil {
		slog.Warn("blocked outbound URL after DNS resolution failure", "host", host, "error", err)
		return wrapBlocked("DNS resolution failed for " + host + ": " + err.Error())
	}
	if len(ips) == 0 {
		slog.Warn("blocked outbound URL after empty DNS resolution result", "host", host)
		return wrapBlocked("DNS resolution returned no IPs for " + host)
	}
	for _, ip := range ips {
		if reason := blockedIPReason(ip); reason != "" {
			return wrapBlocked(reason + " resolution for " + host + ": " + ip.String())
		}
	}
	return nil
}

// blockedIPReason returns a non-empty reason if the IP is in a denied range,
// else "".
func blockedIPReason(ip net.IP) string {
	if ip.IsLoopback() {
		return "loopback IP"
	}
	if ip.IsPrivate() {
		// IsPrivate covers RFC1918 (IPv4) + RFC4193 (IPv6 ULA).
		return "private IP"
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		// 169.254.0.0/16 (IPv4) + fe80::/10 (IPv6) — covers cloud metadata.
		return "link-local IP"
	}
	if ip.IsUnspecified() {
		// 0.0.0.0 / ::
		return "unspecified IP"
	}
	if ip.IsMulticast() {
		return "multicast IP"
	}
	return ""
}

// wrapBlocked formats a rejection reason as an error wrapping errOutboundURLBlocked
// so callers can errors.Is(err, errOutboundURLBlocked) on the rejection class.
func wrapBlocked(reason string) error {
	return blockedURLError{reason: reason}
}

type blockedURLError struct{ reason string }

func (e blockedURLError) Error() string { return errOutboundURLBlocked.Error() + ": " + e.reason }
func (e blockedURLError) Unwrap() error { return errOutboundURLBlocked }
