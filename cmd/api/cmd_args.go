// SPDX-License-Identifier: EUPL-1.2

package api

import "strings"

// splitUniqueCSV trims and deduplicates a comma-separated list while
// preserving the first occurrence of each value.
func splitUniqueCSV(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}

	return values
}

// normalisePublicPaths trims whitespace, ensures a leading slash, and removes
// duplicate entries while preserving the first occurrence of each path.
func normalisePublicPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		path = strings.TrimRight(path, "/")
		if path == "" {
			path = "/"
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
