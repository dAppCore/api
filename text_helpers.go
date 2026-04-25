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
