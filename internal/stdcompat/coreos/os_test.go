// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import (
	"syscall"

	coretest "dappco.re/go"
)

func uniquePath(t *coretest.T, suffix string) string {
	t.Helper()
	f, err := CreateTemp("", "stdcompat-os-*")
	coretest.AssertNoError(t, err)
	name := f.Name()
	coretest.AssertNoError(t, f.Close())
	coretest.AssertNoError(t, RemoveAll(name))
	return name + suffix
}

func TestOs_Getenv_Good(t *coretest.T) {
	key := "CODEX_STDCOMPAT_GETENV_GOOD"
	coretest.AssertNoError(t, Setenv(key, "value"))
	defer func() { _ = Unsetenv(key) }()
	coretest.AssertEqual(t, "value", Getenv(key))
}

func TestOs_Getenv_Bad(t *coretest.T) {
	key := "CODEX_STDCOMPAT_GETENV_BAD"
	coretest.AssertNoError(t, Unsetenv(key))
	coretest.AssertEqual(t, "", Getenv(key))
}

func TestOs_Getenv_Ugly(t *coretest.T) {
	got := Getenv("")
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestOs_LookupEnv_Good(t *coretest.T) {
	key := "CODEX_STDCOMPAT_LOOKUP_GOOD"
	coretest.AssertNoError(t, Setenv(key, "value"))
	defer func() { _ = Unsetenv(key) }()
	got, ok := LookupEnv(key)
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "value", got)
}

func TestOs_LookupEnv_Bad(t *coretest.T) {
	key := "CODEX_STDCOMPAT_LOOKUP_BAD"
	coretest.AssertNoError(t, Unsetenv(key))
	got, ok := LookupEnv(key)
	coretest.AssertFalse(t, ok)
	coretest.AssertEqual(t, "", got)
}

func TestOs_LookupEnv_Ugly(t *coretest.T) {
	key := "CODEX_STDCOMPAT_LOOKUP_UGLY"
	coretest.AssertNoError(t, Setenv(key, ""))
	defer func() { _ = Unsetenv(key) }()
	got, ok := LookupEnv(key)
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "", got)
}

func TestOs_Setenv_Good(t *coretest.T) {
	key := "CODEX_STDCOMPAT_SETENV_GOOD"
	err := Setenv(key, "value")
	defer func() { _ = Unsetenv(key) }()
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "value", Getenv(key))
}

func TestOs_Setenv_Bad(t *coretest.T) {
	err := Setenv("", "value")
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", Getenv(""))
}

func TestOs_Setenv_Ugly(t *coretest.T) {
	key := "CODEX_STDCOMPAT_SETENV_UGLY"
	err := Setenv(key, "")
	defer func() { _ = Unsetenv(key) }()
	got, ok := LookupEnv(key)
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "", got)
}

func TestOs_Unsetenv_Good(t *coretest.T) {
	key := "CODEX_STDCOMPAT_UNSETENV_GOOD"
	coretest.AssertNoError(t, Setenv(key, "value"))
	err := Unsetenv(key)
	_, ok := LookupEnv(key)
	coretest.AssertNoError(t, err)
	coretest.AssertFalse(t, ok)
}

func TestOs_Unsetenv_Bad(t *coretest.T) {
	key := "CODEX_STDCOMPAT_UNSETENV_BAD"
	err := Unsetenv(key)
	_, ok := LookupEnv(key)
	coretest.AssertNoError(t, err)
	coretest.AssertFalse(t, ok)
}

func TestOs_Unsetenv_Ugly(t *coretest.T) {
	err := Unsetenv("")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "", Getenv(""))
}

func TestOs_Exit_Good(t *coretest.T) {
	oldExit := exit
	called := -1
	exit = func(code int) { called = code }
	defer func() { exit = oldExit }()
	Exit(0)
	coretest.AssertEqual(t, 0, called)
}

func TestOs_Exit_Bad(t *coretest.T) {
	oldExit := exit
	called := -1
	exit = func(code int) { called = code }
	defer func() { exit = oldExit }()
	Exit(2)
	coretest.AssertEqual(t, 2, called)
}

func TestOs_Exit_Ugly(t *coretest.T) {
	oldExit := exit
	called := 99
	exit = func(code int) { called = code }
	defer func() { exit = oldExit }()
	Exit(-1)
	coretest.AssertEqual(t, -1, called)
}

func TestOs_ReadFile_Good(t *coretest.T) {
	path := uniquePath(t, ".txt")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, WriteFile(path, []byte("payload"), 0o600))
	data, err := ReadFile(path)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "payload", string(data))
}

func TestOs_ReadFile_Bad(t *coretest.T) {
	path := uniquePath(t, ".missing")
	data, err := ReadFile(path)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, data)
}

func TestOs_ReadFile_Ugly(t *coretest.T) {
	path := uniquePath(t, ".empty")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, WriteFile(path, nil, 0o600))
	data, err := ReadFile(path)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 0, len(data))
}

func TestOs_WriteFile_Good(t *coretest.T) {
	path := uniquePath(t, ".txt")
	defer func() { _ = RemoveAll(path) }()
	err := WriteFile(path, []byte("payload"), 0o600)
	data, readErr := ReadFile(path)
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, readErr)
	coretest.AssertEqual(t, "payload", string(data))
}

func TestOs_WriteFile_Bad(t *coretest.T) {
	path := uniquePath(t, ".dir") + "/payload"
	err := WriteFile(path, []byte("payload"), 0o600)
	coretest.AssertError(t, err)
	coretest.AssertTrue(t, IsNotExist(err))
}

func TestOs_WriteFile_Ugly(t *coretest.T) {
	path := uniquePath(t, ".empty")
	defer func() { _ = RemoveAll(path) }()
	err := WriteFile(path, nil, 0o600)
	data, readErr := ReadFile(path)
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, readErr)
	coretest.AssertEqual(t, 0, len(data))
}

func TestOs_MkdirAll_Good(t *coretest.T) {
	path := uniquePath(t, ".dir/sub")
	defer func() { _ = RemoveAll(coretest.PathDir(path)) }()
	err := MkdirAll(path, 0o755)
	info, statErr := Stat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, statErr)
	coretest.AssertTrue(t, info.IsDir())
}

func TestOs_MkdirAll_Bad(t *coretest.T) {
	err := MkdirAll("", 0o755)
	coretest.AssertError(t, err)
	coretest.AssertTrue(t, IsNotExist(err))
}

func TestOs_MkdirAll_Ugly(t *coretest.T) {
	path := uniquePath(t, ".dir")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, MkdirAll(path, 0o755))
	err := MkdirAll(path, 0o755)
	coretest.AssertNoError(t, err)
}

func TestOs_RemoveAll_Good(t *coretest.T) {
	path := uniquePath(t, ".dir")
	coretest.AssertNoError(t, MkdirAll(path, 0o755))
	err := RemoveAll(path)
	_, statErr := Stat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, IsNotExist(statErr))
}

func TestOs_RemoveAll_Bad(t *coretest.T) {
	path := uniquePath(t, ".missing")
	err := RemoveAll(path)
	coretest.AssertNoError(t, err)
	coretest.AssertFalse(t, IsPermission(err))
}

func TestOs_RemoveAll_Ugly(t *coretest.T) {
	path := uniquePath(t, ".txt")
	coretest.AssertNoError(t, WriteFile(path, []byte("payload"), 0o600))
	err := RemoveAll(path)
	_, statErr := Stat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, IsNotExist(statErr))
}

func TestOs_Stat_Good(t *coretest.T) {
	path := uniquePath(t, ".txt")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, WriteFile(path, []byte("payload"), 0o600))
	info, err := Stat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "payload", coretest.Sprintf("%s", string(mustRead(t, path))))
	coretest.AssertEqual(t, false, info.IsDir())
}

func TestOs_Stat_Bad(t *coretest.T) {
	path := uniquePath(t, ".missing")
	info, err := Stat(path)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, info)
}

func TestOs_Stat_Ugly(t *coretest.T) {
	path := uniquePath(t, ".dir")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, MkdirAll(path, 0o755))
	info, err := Stat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, info.IsDir())
}

func TestOs_Lstat_Good(t *coretest.T) {
	path := uniquePath(t, ".txt")
	defer func() { _ = RemoveAll(path) }()
	coretest.AssertNoError(t, WriteFile(path, []byte("payload"), 0o600))
	info, err := Lstat(path)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, false, info.IsDir())
}

func TestOs_Lstat_Bad(t *coretest.T) {
	path := uniquePath(t, ".missing")
	info, err := Lstat(path)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, info)
}

func TestOs_Lstat_Ugly(t *coretest.T) {
	target := uniquePath(t, ".target")
	link := target + ".link"
	defer func() { _ = RemoveAll(target); _ = RemoveAll(link) }()
	coretest.AssertNoError(t, WriteFile(target, []byte("payload"), 0o600))
	coretest.AssertNoError(t, Symlink(target, link))
	info, err := Lstat(link)
	coretest.AssertNoError(t, err)
	coretest.AssertTrue(t, info.Mode()&ModeSymlink != 0)
}

func TestOs_IsNotExist_Good(t *coretest.T) {
	got := IsNotExist(syscall.ENOENT)
	coretest.AssertTrue(t, got)
	coretest.AssertError(t, syscall.ENOENT)
}

func TestOs_IsNotExist_Bad(t *coretest.T) {
	got := IsNotExist(nil)
	coretest.AssertFalse(t, got)
	coretest.AssertNil(t, error(nil))
}

func TestOs_IsNotExist_Ugly(t *coretest.T) {
	got := IsNotExist(coretest.NewError("plain"))
	coretest.AssertFalse(t, got)
	coretest.AssertError(t, coretest.NewError("plain"))
}

func TestOs_IsPermission_Good(t *coretest.T) {
	got := IsPermission(syscall.EACCES)
	coretest.AssertTrue(t, got)
	coretest.AssertError(t, syscall.EACCES)
}

func TestOs_IsPermission_Bad(t *coretest.T) {
	got := IsPermission(nil)
	coretest.AssertFalse(t, got)
	coretest.AssertNil(t, error(nil))
}

func TestOs_IsPermission_Ugly(t *coretest.T) {
	got := IsPermission(coretest.NewError("plain"))
	coretest.AssertFalse(t, got)
	coretest.AssertError(t, coretest.NewError("plain"))
}

func TestOs_Symlink_Good(t *coretest.T) {
	target := uniquePath(t, ".target")
	link := target + ".link"
	defer func() { _ = RemoveAll(target); _ = RemoveAll(link) }()
	coretest.AssertNoError(t, WriteFile(target, []byte("payload"), 0o600))
	err := Symlink(target, link)
	info, statErr := Lstat(link)
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, statErr)
	coretest.AssertTrue(t, info.Mode()&ModeSymlink != 0)
}

func TestOs_Symlink_Bad(t *coretest.T) {
	target := uniquePath(t, ".target")
	link := target + ".link"
	defer func() { _ = RemoveAll(target); _ = RemoveAll(link) }()
	coretest.AssertNoError(t, WriteFile(target, []byte("payload"), 0o600))
	coretest.AssertNoError(t, Symlink(target, link))
	err := Symlink(target, link)
	coretest.AssertError(t, err)
}

func TestOs_Symlink_Ugly(t *coretest.T) {
	target := uniquePath(t, ".dir")
	link := target + ".link"
	defer func() { _ = RemoveAll(target); _ = RemoveAll(link) }()
	coretest.AssertNoError(t, MkdirAll(target, 0o755))
	err := Symlink(target, link)
	info, statErr := Lstat(link)
	coretest.AssertNoError(t, err)
	coretest.AssertNoError(t, statErr)
	coretest.AssertTrue(t, info.Mode()&ModeSymlink != 0)
}

func TestOs_CreateTemp_Good(t *coretest.T) {
	f, err := CreateTemp("", "stdcompat-*")
	coretest.AssertNoError(t, err)
	defer func() { _ = RemoveAll(f.Name()) }()
	coretest.AssertNotNil(t, f)
	coretest.AssertNoError(t, f.Close())
}

func TestOs_CreateTemp_Bad(t *coretest.T) {
	f, err := CreateTemp("/definitely/missing/stdcompat", "stdcompat-*")
	coretest.AssertError(t, err)
	coretest.AssertNil(t, f)
}

func TestOs_CreateTemp_Ugly(t *coretest.T) {
	dir := uniquePath(t, ".dir")
	defer func() { _ = RemoveAll(dir) }()
	coretest.AssertNoError(t, MkdirAll(dir, 0o755))
	f, err := CreateTemp(dir, "prefix-*.txt")
	coretest.AssertNoError(t, err)
	defer func() { _ = RemoveAll(f.Name()) }()
	coretest.AssertContains(t, f.Name(), "prefix-")
	coretest.AssertNoError(t, f.Close())
}

func mustRead(t *coretest.T, path string) []byte {
	t.Helper()
	data, err := ReadFile(path)
	coretest.AssertNoError(t, err)
	return data
}
