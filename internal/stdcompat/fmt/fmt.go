// SPDX-License-Identifier: EUPL-1.2

package fmt

import core "dappco.re/go"

func Sprint(args ...any) string                 { return core.Sprint(args...) }
func Sprintf(format string, args ...any) string { return core.Sprintf(format, args...) }
func Errorf(format string, args ...any) error   { return core.Errorf(format, args...) }

func Printf(format string, args ...any) (int, error) {
	text := core.Sprintf(format, args...)
	r := core.WriteString(core.Stdout(), text)
	if !r.OK {
		err, _ := r.Value.(error)
		return 0, err
	}
	n, _ := r.Value.(int)
	return n, nil
}
