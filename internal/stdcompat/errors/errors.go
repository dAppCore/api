// SPDX-License-Identifier: EUPL-1.2

package errors

import core "dappco.re/go"

func New(text string) error         { return core.NewError(text) }
func Is(err, target error) bool     { return core.Is(err, target) }
func As(err error, target any) bool { return core.As(err, target) }
func Join(errs ...error) error      { return core.ErrorJoin(errs...) }

func Unwrap(err error) error {
	type unwrapper interface{ Unwrap() error }
	if err == nil {
		return nil
	}
	if wrapped, ok := err.(unwrapper); ok {
		return wrapped.Unwrap()
	}
	return nil
}
