// SPDX-License-Identifier: EUPL-1.2

package api

func cutString(s, sep string) (string, string, bool) {
	if sep == "" {
		return "", s, true
	}
	index := indexString(s, sep)
	if index < 0 {
		return s, "", false
	}
	return s[:index], s[index+len(sep):], true
}

func trimLeftFunc(s string, drop func(rune) bool) string {
	for i, r := range s {
		if !drop(r) {
			return s[i:]
		}
	}
	return ""
}

// indexString returns the index of the first occurrence of substr in s, or -1
// if substr is not present. A hand-rolled replacement for strings.Index (AX-6
// — strings is a banned import; no core primitive covers substring search).
func indexString(s, substr string) int {
	if substr == "" {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	max := len(s) - len(substr)
	first := substr[0]
	for i := 0; i <= max; i++ {
		if s[i] == first && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// indexByte returns the index of the first occurrence of target in s, or -1
// if target is not present. A hand-rolled replacement for strings.IndexByte
// (AX-6 — strings is a banned import; no core primitive covers byte search).
func indexByte(s string, target byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == target {
			return i
		}
	}
	return -1
}
