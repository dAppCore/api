// SPDX-License-Identifier: EUPL-1.2

package os

import (
	"syscall"
	"time"

	c "dappco.re/go"
)

type File = c.OSFile
type FileInfo = c.FsFileInfo
type FileMode = c.FileMode
type Signal interface {
	String() string
	Signal()
}

const (
	PathSeparator     = c.PathSeparator
	PathListSeparator = c.PathListSeparator
	ModeSymlink       = c.ModeSymlink
)

var (
	Stdout           = c.Stdout()
	Stderr           = c.Stderr()
	Args             = c.Args()
	Interrupt Signal = syscall.SIGINT
	exit             = c.Exit
)

func Getenv(key string) string            { return c.Getenv(key) }
func LookupEnv(key string) (string, bool) { return c.LookupEnv(key) }
func Setenv(key, value string) error      { return resultErr(c.Setenv(key, value)) }
func Unsetenv(key string) error           { return resultErr(c.Unsetenv(key)) }
func Exit(code int)                       { exit(code) }
func ReadFile(p string) ([]byte, error)   { r := c.ReadFile(p); return resultBytes(r) }
func WriteFile(p string, data []byte, mode FileMode) error {
	return resultErr(c.WriteFile(p, data, mode))
}
func MkdirAll(p string, mode FileMode) error { return resultErr(c.MkdirAll(p, mode)) }
func RemoveAll(p string) error               { return resultErr(c.RemoveAll(p)) }
func Stat(p string) (FileInfo, error)        { return resultInfo(c.Stat(p)) }
func Lstat(p string) (FileInfo, error)       { return resultInfo(c.Lstat(p)) }
func IsNotExist(err error) bool              { return c.IsNotExist(err) }
func IsPermission(err error) bool            { return c.IsPermission(err) }
func Symlink(oldname, newname string) error  { return syscall.Symlink(oldname, newname) }

func CreateTemp(dir, pattern string) (*File, error) {
	if dir == "" {
		dir = c.TempDir()
	}
	prefix, suffix := splitPattern(pattern)
	for i := 0; i < 100; i++ {
		name := c.PathJoin(dir, prefix+c.Sprintf("%d", time.Now().UnixNano()+int64(i))+suffix)
		r := c.OpenFile(name, c.O_RDWR|c.O_CREATE|c.O_EXCL, 0o600)
		if r.OK {
			f, _ := r.Value.(*File)
			return f, nil
		}
	}
	return nil, c.NewError("create temp failed")
}

func splitPattern(pattern string) (string, string) {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return pattern[:i], pattern[i+1:]
		}
	}
	return pattern, ""
}

func resultErr(r c.Result) error {
	if r.OK {
		return nil
	}
	err, _ := r.Value.(error)
	return err
}

func resultBytes(r c.Result) ([]byte, error) {
	if !r.OK {
		return nil, resultErr(r)
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func resultInfo(r c.Result) (FileInfo, error) {
	if !r.OK {
		return nil, resultErr(r)
	}
	info, _ := r.Value.(FileInfo)
	return info, nil
}
