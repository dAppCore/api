// SPDX-License-Identifier: EUPL-1.2

package api

import core "dappco.re/go"

func testCoreResultError(r core.Result) error {
	if r.OK {
		return nil
	}
	if err, ok := r.Value.(error); ok {
		return err
	}
	return core.NewError("core operation failed")
}

func coreJSONMarshal(v any) ([]byte, error) {
	r := core.JSONMarshal(v)
	if !r.OK {
		return nil, testCoreResultError(r)
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func coreJSONUnmarshal(data []byte, target any) error {
	return testCoreResultError(core.JSONUnmarshal(data, target))
}

func coreWriteFile(path string, data []byte, mode core.FileMode) error {
	return testCoreResultError(core.WriteFile(path, data, mode))
}

func coreMkdirAll(path string, mode core.FileMode) error {
	return testCoreResultError(core.MkdirAll(path, mode))
}

func coreBytesEqual(a, b []byte) bool {
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
