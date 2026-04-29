// SPDX-License-Identifier: EUPL-1.2

package bytes

import core "dappco.re/go"

type Buffer struct {
	data []byte
	off  int
}

func NewBufferString(s string) *Buffer { return &Buffer{data: []byte(s)} }
func NewBuffer(b []byte) *Buffer       { return &Buffer{data: append([]byte(nil), b...)} }
func NewReader(b []byte) core.Reader   { return core.NewReader(string(b)) }

func (b *Buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *Buffer) WriteString(s string) (int, error) {
	b.data = append(b.data, s...)
	return len(s), nil
}

func (b *Buffer) Read(p []byte) (int, error) {
	if b.off >= len(b.data) {
		return 0, core.EOF
	}
	n := copy(p, b.data[b.off:])
	b.off += n
	return n, nil
}

func (b *Buffer) String() string { return string(b.data) }
func (b *Buffer) Bytes() []byte  { return append([]byte(nil), b.data...) }
func (b *Buffer) Len() int       { return len(b.data) - b.off }
func (b *Buffer) Reset()         { b.data, b.off = nil, 0 }

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Contains(b, sub []byte) bool { return core.Contains(string(b), string(sub)) }

func Repeat(b []byte, count int) []byte {
	if count <= 0 {
		return nil
	}
	out := make([]byte, 0, len(b)*count)
	for i := 0; i < count; i++ {
		out = append(out, b...)
	}
	return out
}

func TrimSpace(b []byte) []byte {
	start := 0
	for start < len(b) && isSpace(b[start]) {
		start++
	}
	end := len(b)
	for end > start && isSpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t', '\v', '\f':
		return true
	default:
		return false
	}
}
