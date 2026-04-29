// SPDX-License-Identifier: EUPL-1.2

package exec

import (
	"context"

	coretest "dappco.re/go"
)

func ExampleCommandContext() {
	cmd := CommandContext(context.Background(), "/bin/sh", "-c", "exit 0")
	coretest.Println(cmd != nil)
	// Output: true
}

func ExampleCmd_Run() {
	cmd := CommandContext(context.Background(), "/bin/sh", "-c", "exit 0")
	err := cmd.Run()
	coretest.Println(err == nil)
	// Output: true
}
