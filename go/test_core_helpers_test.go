// SPDX-License-Identifier: EUPL-1.2

package api_test

import core "dappco.re/go"

func coreResultError(r core.Result) error {
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
		return nil, coreResultError(r)
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func coreJSONUnmarshal(data []byte, target any) error {
	return coreResultError(core.JSONUnmarshal(data, target))
}

func coreJSONDecode(reader core.Reader, target any) error {
	r := core.ReadAll(reader)
	if !r.OK {
		return coreResultError(r)
	}
	text, _ := r.Value.(string)
	return coreJSONUnmarshal([]byte(text), target)
}

func coreWriteFile(path string, data []byte, mode core.FileMode) error {
	return coreResultError(core.WriteFile(path, data, mode))
}

func coreReadFile(path string) ([]byte, error) {
	r := core.ReadFile(path)
	if !r.OK {
		return nil, coreResultError(r)
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func coreStat(path string) (core.FsFileInfo, error) {
	r := core.Stat(path)
	if !r.OK {
		return nil, coreResultError(r)
	}
	info, _ := r.Value.(core.FsFileInfo)
	return info, nil
}

func coreStringRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	b := core.NewBuilder()
	for i := 0; i < count; i++ {
		b.WriteString(s)
	}
	return b.String()
}

func coreCutPrefix(s, prefix string) (string, bool) {
	if !core.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

func coreBytesRepeat(b []byte, count int) []byte {
	if count <= 0 {
		return nil
	}
	out := make([]byte, 0, len(b)*count)
	for i := 0; i < count; i++ {
		out = append(out, b...)
	}
	return out
}
