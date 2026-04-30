// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	"syscall"

	. "dappco.re/go"
)

func coreResultError(r Result) error {
	if r.OK {
		return nil
	}
	if err, ok := r.Value.(error); ok {
		return err
	}
	return NewError("core operation failed")
}

func coreJSONUnmarshal(data []byte, target any) error {
	return coreResultError(JSONUnmarshal(data, target))
}

func coreJSONEncode(writer Writer, v any) error {
	r := JSONMarshal(v)
	if !r.OK {
		return coreResultError(r)
	}
	data, _ := r.Value.([]byte)
	data = append(data, '\n')
	_, err := writer.Write(data)
	return err
}

func coreWriteFile(path string, data []byte, mode FileMode) error {
	return coreResultError(WriteFile(path, data, mode))
}

func coreMkdirAll(path string, mode FileMode) error {
	return coreResultError(MkdirAll(path, mode))
}

func coreSetenv(key, value string) error {
	return coreResultError(Setenv(key, value))
}

func coreUnsetenv(key string) error {
	return coreResultError(Unsetenv(key))
}

func corePathEvalSymlinks(path string) (string, error) {
	r := PathEvalSymlinks(path)
	if !r.OK {
		return "", coreResultError(r)
	}
	resolved, _ := r.Value.(string)
	return resolved, nil
}

func coreSymlink(oldname, newname string) error {
	return syscall.Symlink(oldname, newname)
}
