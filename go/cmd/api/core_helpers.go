// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"time"

	core "dappco.re/go"
)

func createTempFile(dir, pattern string) (
	*core.OSFile,
	error,
) {
	if dir == "" {
		dir = core.TempDir()
	}
	prefix, suffix := splitTempPattern(pattern)
	for i := 0; i < 100; i++ {
		name := core.PathJoin(dir, prefix+core.Sprintf("%d", time.Now().UnixNano()+int64(i))+suffix)
		r := core.OpenFile(name, core.O_RDWR|core.O_CREATE|core.O_EXCL, 0o600)
		if r.OK {
			file, _ := r.Value.(*core.OSFile)
			return file, nil
		}
	}
	return nil, core.NewError("create temp failed")
}

func splitTempPattern(pattern string) (string, string) {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return pattern[:i], pattern[i+1:]
		}
	}
	return pattern, ""
}
