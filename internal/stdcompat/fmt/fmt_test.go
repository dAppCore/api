// SPDX-License-Identifier: EUPL-1.2

package fmt

import coretest "dappco.re/go"

func TestFmt_Sprint_Good(t *coretest.T) {
	got := Sprint("api", " ", "gateway")
	coretest.AssertEqual(t, "api gateway", got)
	coretest.AssertContains(t, got, "gateway")
}

func TestFmt_Sprint_Bad(t *coretest.T) {
	got := Sprint()
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestFmt_Sprint_Ugly(t *coretest.T) {
	got := Sprint("api", 1)
	coretest.AssertEqual(t, "api1", got)
	coretest.AssertContains(t, got, "1")
}

func TestFmt_Sprintf_Good(t *coretest.T) {
	got := Sprintf("%s:%d", "api", 1)
	coretest.AssertEqual(t, "api:1", got)
	coretest.AssertContains(t, got, ":")
}

func TestFmt_Sprintf_Bad(t *coretest.T) {
	got := Sprintf("")
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestFmt_Sprintf_Ugly(t *coretest.T) {
	got := Sprintf("%v", []int{1, 2})
	coretest.AssertEqual(t, "[1 2]", got)
	coretest.AssertContains(t, got, "2")
}

func TestFmt_Errorf_Good(t *coretest.T) {
	err := Errorf("api %s", "failed")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "failed")
}

func TestFmt_Errorf_Bad(t *coretest.T) {
	err := Errorf("")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", err.Error())
}

func TestFmt_Errorf_Ugly(t *coretest.T) {
	err := Errorf("%s:%d", "api", 1)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "api:1", err.Error())
}

func TestFmt_Printf_Good(t *coretest.T) {
	n, err := Printf("%s", "api")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 3, n)
}

func TestFmt_Printf_Bad(t *coretest.T) {
	n, err := Printf("")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestFmt_Printf_Ugly(t *coretest.T) {
	n, err := Printf("%s\n", "api")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 4, n)
}
