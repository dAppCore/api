// SPDX-License-Identifier: EUPL-1.2

package exec

import (
	"context"

	coretest "dappco.re/go"
)

func TestExec_CommandContext_Good(t *coretest.T) {
	ctx := context.Background()
	cmd := CommandContext(ctx, "/bin/sh", "-c", "exit 0")
	coretest.AssertNotNil(t, cmd)
	coretest.AssertEqual(t, "/bin/sh", cmd.name)
}

func TestExec_CommandContext_Bad(t *coretest.T) {
	ctx := context.Background()
	cmd := CommandContext(ctx, "", "-c", "exit 0")
	coretest.AssertNotNil(t, cmd)
	coretest.AssertEqual(t, "", cmd.name)
}

func TestExec_CommandContext_Ugly(t *coretest.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := CommandContext(ctx, "/bin/sh", "-c", "sleep 1")
	coretest.AssertNotNil(t, cmd)
	coretest.AssertEqual(t, 2, len(cmd.args))
}

func TestExec_Cmd_Run_Good(t *coretest.T) {
	cmd := CommandContext(context.Background(), "/bin/sh", "-c", "exit 0")
	err := cmd.Run()
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "/bin/sh", cmd.name)
}

func TestExec_Cmd_Run_Bad(t *coretest.T) {
	cmd := CommandContext(context.Background(), "/bin/sh", "-c", "exit 7")
	err := cmd.Run()
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "exit status")
}

func TestExec_Cmd_Run_Ugly(t *coretest.T) {
	var cmd *Cmd
	err := cmd.Run()
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "nil command")
}
