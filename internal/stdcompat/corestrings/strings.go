// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import core "dappco.re/go"

func Contains(s, substr string) bool         { return core.Contains(s, substr) }
func HasPrefix(s, prefix string) bool        { return core.HasPrefix(s, prefix) }
func HasSuffix(s, suffix string) bool        { return core.HasSuffix(s, suffix) }
func TrimSpace(s string) string              { return core.Trim(s) }
func TrimSuffix(s, suffix string) string     { return core.TrimSuffix(s, suffix) }
func TrimPrefix(s, prefix string) string     { return core.TrimPrefix(s, prefix) }
func ToLower(s string) string                { return core.Lower(s) }
func NewReader(s string) core.Reader         { return core.NewReader(s) }
func Join(parts []string, sep string) string { return core.Join(sep, parts...) }
func Split(s, sep string) []string           { return core.Split(s, sep) }

func Repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	b := core.NewBuilder()
	for i := 0; i < count; i++ {
		b.WriteString(s)
	}
	return b.String()
}

func CutPrefix(s, prefix string) (string, bool) {
	if !core.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}
