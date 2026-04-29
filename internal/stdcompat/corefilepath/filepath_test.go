// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

func TestFilepath_Join_Good(t *coretest.T) {
	got := Join("api", "gateway")
	coretest.AssertEqual(t, "api/gateway", got)
	coretest.AssertContains(t, got, "gateway")
}

func TestFilepath_Join_Bad(t *coretest.T) {
	got := Join("", "gateway")
	coretest.AssertEqual(t, "gateway", got)
	coretest.AssertFalse(t, IsAbs(got))
}

func TestFilepath_Join_Ugly(t *coretest.T) {
	got := Join("api", "..", "gateway")
	coretest.AssertEqual(t, "gateway", got)
	coretest.AssertEqual(t, Clean("api/../gateway"), got)
}

func TestFilepath_Dir_Good(t *coretest.T) {
	got := Dir("api/gateway/config.json")
	coretest.AssertEqual(t, "api/gateway", got)
	coretest.AssertContains(t, got, "gateway")
}

func TestFilepath_Dir_Bad(t *coretest.T) {
	got := Dir("config.json")
	coretest.AssertEqual(t, ".", got)
	coretest.AssertFalse(t, IsAbs(got))
}

func TestFilepath_Dir_Ugly(t *coretest.T) {
	got := Dir("/")
	coretest.AssertEqual(t, "/", got)
	coretest.AssertTrue(t, IsAbs(got))
}

func TestFilepath_Clean_Good(t *coretest.T) {
	got := Clean("api/../gateway")
	coretest.AssertEqual(t, "gateway", got)
	coretest.AssertFalse(t, IsAbs(got))
}

func TestFilepath_Clean_Bad(t *coretest.T) {
	got := Clean("")
	coretest.AssertEqual(t, ".", got)
	coretest.AssertEqual(t, Dir(got), ".")
}

func TestFilepath_Clean_Ugly(t *coretest.T) {
	got := Clean("/api//gateway/")
	coretest.AssertEqual(t, "/api/gateway", got)
	coretest.AssertTrue(t, IsAbs(got))
}

func TestFilepath_IsAbs_Good(t *coretest.T) {
	got := IsAbs("/api")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "/", Dir("/api"))
}

func TestFilepath_IsAbs_Bad(t *coretest.T) {
	got := IsAbs("api")
	coretest.AssertFalse(t, got)
	coretest.AssertEqual(t, "api", Clean("api"))
}

func TestFilepath_IsAbs_Ugly(t *coretest.T) {
	got := IsAbs("")
	coretest.AssertFalse(t, got)
	coretest.AssertEqual(t, ".", Clean(""))
}

func TestFilepath_Abs_Good(t *coretest.T) {
	got, err := Abs(".")
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, IsAbs(got))
}

func TestFilepath_Abs_Bad(t *coretest.T) {
	got, err := Abs("")
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, IsAbs(got))
}

func TestFilepath_Abs_Ugly(t *coretest.T) {
	got, err := Abs("..")
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, IsAbs(got))
}

func TestFilepath_Rel_Good(t *coretest.T) {
	got, err := Rel("/tmp", "/tmp/api")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "api", got)
}

func TestFilepath_Rel_Bad(t *coretest.T) {
	got, err := Rel("/tmp/api", "/tmp/api")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, ".", got)
}

func TestFilepath_Rel_Ugly(t *coretest.T) {
	got, err := Rel("/tmp/api", "/tmp")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "..", got)
}

func TestFilepath_EvalSymlinks_Good(t *coretest.T) {
	got, err := EvalSymlinks(".")
	coretest.AssertNoError(t, err)
	coretest.AssertNotEmpty(t, got)
}

func TestFilepath_EvalSymlinks_Bad(t *coretest.T) {
	got, err := EvalSymlinks("__missing_path_for_stdcompat__")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", got)
}

func TestFilepath_EvalSymlinks_Ugly(t *coretest.T) {
	got, err := EvalSymlinks("..")
	coretest.AssertNoError(t, err)
	coretest.AssertNotEmpty(t, got)
}
