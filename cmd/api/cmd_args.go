// SPDX-License-Identifier: EUPL-1.2

package api

import core "dappco.re/go"

// splitUniqueCSV trims and deduplicates a comma-separated list while
// preserving the first occurrence of each value.
func splitUniqueCSV(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := core.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		value := core.Trim(part)
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
		path = core.Trim(path)
		if path == "" {
			continue
		}
		if !core.HasPrefix(path, "/") {
			path = "/" + path
		}
		for core.HasSuffix(path, "/") {
			path = core.TrimSuffix(path, "/")
		}
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
