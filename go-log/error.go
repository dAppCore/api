// SPDX-License-Identifier: EUPL-1.2

package log

import "fmt"

// E wraps an operation label and message in a conventional error.
// If err is non-nil, it is wrapped with %w.
func E(op, message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %s: %w", op, message, err)
	}
	return fmt.Errorf("%s: %s", op, message)
}
