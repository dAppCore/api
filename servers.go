// SPDX-License-Identifier: EUPL-1.2

package api

import (
	core "dappco.re/go/core"
)

// normaliseServers trims whitespace, removes empty entries, and preserves
// the first occurrence of each server URL.
func normaliseServers(servers []string) []string {
	if len(servers) == 0 {
		return nil
	}

	cleaned := make([]string, 0, len(servers))
	seen := make(map[string]struct{}, len(servers))

	for _, server := range servers {
		server = normaliseServer(server)
		if server == "" {
			continue
		}
		if _, ok := seen[server]; ok {
			continue
		}
		seen[server] = struct{}{}
		cleaned = append(cleaned, server)
	}

	if len(cleaned) == 0 {
		return nil
	}

	return cleaned
}

// normaliseServer trims surrounding whitespace and removes a trailing slash
// from non-root server URLs so equivalent metadata collapses to one entry.
func normaliseServer(server string) string {
	server = core.Trim(server)
	if server == "" {
		return ""
	}
	if server == "/" {
		return server
	}

	server = trimTrailingSlashes(server)
	if server == "" {
		return "/"
	}

	return server
}

func trimSlashes(value string) string {
	for core.HasPrefix(value, "/") {
		value = core.TrimPrefix(value, "/")
	}
	return trimTrailingSlashes(value)
}

func trimTrailingSlashes(value string) string {
	for core.HasSuffix(value, "/") {
		value = core.TrimSuffix(value, "/")
	}
	return value
}
