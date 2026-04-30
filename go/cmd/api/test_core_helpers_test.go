// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"syscall"
	"time"

	core "dappco.re/go"
)

func coreResultError(r core.Result) error {
	if r.OK {
		return nil
	}
	if err, ok := r.Value.(error); ok {
		return err
	}
	return core.NewError("core operation failed")
}

func coreJSONUnmarshal(data []byte, target any) error {
	return coreResultError(core.JSONUnmarshal(data, target))
}

func coreReadFile(path string) ([]byte, error) {
	r := core.ReadFile(path)
	if !r.OK {
		return nil, coreResultError(r)
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func coreWriteFile(path string, data []byte, mode core.FileMode) error {
	return coreResultError(core.WriteFile(path, data, mode))
}

func coreMkdirAll(path string, mode core.FileMode) error {
	return coreResultError(core.MkdirAll(path, mode))
}

func coreStat(path string) (core.FsFileInfo, error) {
	r := core.Stat(path)
	if !r.OK {
		return nil, coreResultError(r)
	}
	info, _ := r.Value.(core.FsFileInfo)
	return info, nil
}

func coreLstat(path string) (core.FsFileInfo, error) {
	r := core.Lstat(path)
	if !r.OK {
		return nil, coreResultError(r)
	}
	info, _ := r.Value.(core.FsFileInfo)
	return info, nil
}

func coreSymlink(oldname, newname string) error {
	return syscall.Symlink(oldname, newname)
}

func coreCreateTemp(dir, pattern string) (*core.OSFile, error) {
	if dir == "" {
		dir = core.TempDir()
	}
	prefix, suffix := coreSplitTempPattern(pattern)
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

func coreSplitTempPattern(pattern string) (string, string) {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return pattern[:i], pattern[i+1:]
		}
	}
	return pattern, ""
}
