// SPDX-License-Identifier: EUPL-1.2

package api

import "strings"

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
	server = strings.TrimSpace(server)
	if server == "" {
		return ""
	}
	if server == "/" {
		return server
	}

	server = strings.TrimRight(server, "/")
	if server == "" {
		return "/"
	}

	return server
}
