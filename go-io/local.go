// SPDX-License-Identifier: EUPL-1.2

package io

import "os"

// LocalFS provides simple local filesystem helpers used by the API module.
var Local localFS

type localFS struct{}

// EnsureDir creates the directory path if it does not already exist.
func (localFS) EnsureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}
	return os.MkdirAll(path, 0o755)
}

// Delete removes the named file, ignoring missing files.
func (localFS) Delete(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
