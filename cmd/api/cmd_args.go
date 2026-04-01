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
