// SPDX-License-Identifier: EUPL-1.2

package os

import (
	"syscall"

	coretest "dappco.re/go"
)

func ExampleGetenv() {
	_ = Setenv("CODEX_STDCOMPAT_EXAMPLE_GETENV", "value")
	defer func() { _ = Unsetenv("CODEX_STDCOMPAT_EXAMPLE_GETENV") }()
	coretest.Println(Getenv("CODEX_STDCOMPAT_EXAMPLE_GETENV"))
	// Output: value
}

func ExampleLookupEnv() {
	_ = Setenv("CODEX_STDCOMPAT_EXAMPLE_LOOKUP", "value")
	defer func() { _ = Unsetenv("CODEX_STDCOMPAT_EXAMPLE_LOOKUP") }()
	got, ok := LookupEnv("CODEX_STDCOMPAT_EXAMPLE_LOOKUP")
	coretest.Println(ok, got)
	// Output: true value
}

func ExampleSetenv() {
	_ = Setenv("CODEX_STDCOMPAT_EXAMPLE_SETENV", "value")
	defer func() { _ = Unsetenv("CODEX_STDCOMPAT_EXAMPLE_SETENV") }()
	coretest.Println(Getenv("CODEX_STDCOMPAT_EXAMPLE_SETENV"))
	// Output: value
}

func ExampleUnsetenv() {
	_ = Setenv("CODEX_STDCOMPAT_EXAMPLE_UNSETENV", "value")
	_ = Unsetenv("CODEX_STDCOMPAT_EXAMPLE_UNSETENV")
	_, ok := LookupEnv("CODEX_STDCOMPAT_EXAMPLE_UNSETENV")
	coretest.Println(ok)
	// Output: false
}

func ExampleExit() {
	oldExit := exit
	called := -1
	exit = func(code int) { called = code }
	defer func() { exit = oldExit }()
	Exit(3)
	coretest.Println(called)
	// Output: 3
}

func ExampleReadFile() {
	path := osExamplePath()
	defer func() { _ = RemoveAll(path) }()
	_ = WriteFile(path, []byte("payload"), 0o600)
	data, _ := ReadFile(path)
	coretest.Println(string(data))
	// Output: payload
}

func ExampleWriteFile() {
	path := osExamplePath()
	defer func() { _ = RemoveAll(path) }()
	err := WriteFile(path, []byte("payload"), 0o600)
	coretest.Println(err == nil)
	// Output: true
}

func ExampleMkdirAll() {
	path := osExamplePath() + ".dir"
	defer func() { _ = RemoveAll(path) }()
	err := MkdirAll(path, 0o755)
	coretest.Println(err == nil)
	// Output: true
}

func ExampleRemoveAll() {
	path := osExamplePath()
	_ = WriteFile(path, []byte("payload"), 0o600)
	err := RemoveAll(path)
	coretest.Println(err == nil)
	// Output: true
}

func ExampleStat() {
	path := osExamplePath()
	defer func() { _ = RemoveAll(path) }()
	_ = WriteFile(path, []byte("payload"), 0o600)
	info, err := Stat(path)
	coretest.Println(err == nil, info.IsDir())
	// Output: true false
}

func ExampleLstat() {
	path := osExamplePath()
	defer func() { _ = RemoveAll(path) }()
	_ = WriteFile(path, []byte("payload"), 0o600)
	info, err := Lstat(path)
	coretest.Println(err == nil, info.IsDir())
	// Output: true false
}

func ExampleIsNotExist() {
	coretest.Println(IsNotExist(syscall.ENOENT))
	// Output: true
}

func ExampleIsPermission() {
	coretest.Println(IsPermission(syscall.EACCES))
	// Output: true
}

func ExampleSymlink() {
	target := osExamplePath()
	link := target + ".link"
	defer func() { _ = RemoveAll(target); _ = RemoveAll(link) }()
	_ = WriteFile(target, []byte("payload"), 0o600)
	err := Symlink(target, link)
	coretest.Println(err == nil)
	// Output: true
}

func ExampleCreateTemp() {
	f, err := CreateTemp("", "stdcompat-*")
	defer func() {
		if f != nil {
			_ = RemoveAll(f.Name())
			_ = f.Close()
		}
	}()
	coretest.Println(err == nil, f != nil)
	// Output: true true
}

func osExamplePath() string {
	f, err := CreateTemp("", "stdcompat-example-*")
	if err != nil {
		return coretest.PathJoin(coretest.TempDir(), "stdcompat-example-fallback")
	}
	name := f.Name()
	_ = f.Close()
	_ = RemoveAll(name)
	return name
}
