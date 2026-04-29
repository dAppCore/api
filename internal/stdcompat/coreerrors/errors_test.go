// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

type customError struct {
	text string
}

func (e customError) Error() string { return e.text }

func TestErrors_New_Good(t *coretest.T) {
	err := New("api failed")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "api failed", err.Error())
}

func TestErrors_New_Bad(t *coretest.T) {
	err := New("")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", err.Error())
}

func TestErrors_New_Ugly(t *coretest.T) {
	err := New("line\nbreak")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "break")
}

func TestErrors_Is_Good(t *coretest.T) {
	target := New("sentinel")
	err := coretest.Wrap(target, "api.Run", "failed")
	coretest.AssertTrue(t, Is(err, target))
	coretest.AssertError(t, err)
}

func TestErrors_Is_Bad(t *coretest.T) {
	err := New("sentinel")
	target := New("sentinel")
	coretest.AssertFalse(t, Is(err, target))
	coretest.AssertFalse(t, err == target)
}

func TestErrors_Is_Ugly(t *coretest.T) {
	coretest.AssertFalse(t, Is(nil, New("sentinel")))
	coretest.AssertFalse(t, Is(New("other"), nil))
	coretest.AssertNil(t, Unwrap(nil))
}

func TestErrors_As_Good(t *coretest.T) {
	err := coretest.E("api.Run", "failed", nil)
	var target *coretest.Err
	coretest.AssertTrue(t, As(err, &target))
	coretest.AssertEqual(t, "api.Run", target.Operation)
}

func TestErrors_As_Bad(t *coretest.T) {
	err := New("plain")
	var target customError
	coretest.AssertFalse(t, As(err, &target))
	coretest.AssertEqual(t, "", target.text)
}

func TestErrors_As_Ugly(t *coretest.T) {
	err := customError{text: "custom"}
	var target customError
	coretest.AssertTrue(t, As(err, &target))
	coretest.AssertEqual(t, "custom", target.text)
}

func TestErrors_Join_Good(t *coretest.T) {
	err := Join(New("one"), New("two"))
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "one")
	coretest.AssertContains(t, err.Error(), "two")
}

func TestErrors_Join_Bad(t *coretest.T) {
	err := Join(nil)
	coretest.AssertNil(t, err)
	coretest.AssertFalse(t, Is(err, New("one")))
}

func TestErrors_Join_Ugly(t *coretest.T) {
	err := Join(New("one"), nil)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "one", err.Error())
}

func TestErrors_Unwrap_Good(t *coretest.T) {
	cause := New("cause")
	err := coretest.Wrap(cause, "api.Run", "failed")
	got := Unwrap(err)
	coretest.AssertEqual(t, cause, got)
}

func TestErrors_Unwrap_Bad(t *coretest.T) {
	err := New("plain")
	got := Unwrap(err)
	coretest.AssertNil(t, got)
}

func TestErrors_Unwrap_Ugly(t *coretest.T) {
	got := Unwrap(nil)
	coretest.AssertNil(t, got)
	coretest.AssertFalse(t, Is(got, New("cause")))
}
