// SPDX-License-Identifier: EUPL-1.2

package filepath

import core "dappco.re/go"

const Separator = core.PathSeparator

func Join(elem ...string) string { return core.PathJoin(elem...) }
func Dir(p string) string        { return core.PathDir(p) }
func Clean(p string) string      { return core.CleanPath(p, string(core.PathSeparator)) }
func IsAbs(p string) bool        { return core.PathIsAbs(p) }

func Abs(p string) (string, error) {
	r := core.PathAbs(p)
	if !r.OK {
		err, _ := r.Value.(error)
		return "", err
	}
	out, _ := r.Value.(string)
	return out, nil
}

func Rel(base, target string) (string, error) {
	r := core.PathRel(base, target)
	if !r.OK {
		err, _ := r.Value.(error)
		return "", err
	}
	out, _ := r.Value.(string)
	return out, nil
}

func EvalSymlinks(p string) (string, error) {
	r := core.PathEvalSymlinks(p)
	if !r.OK {
		err, _ := r.Value.(error)
		return "", err
	}
	out, _ := r.Value.(string)
	return out, nil
}
