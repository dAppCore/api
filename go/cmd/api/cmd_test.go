// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"testing"

	core "dappco.re/go"
)

// TestCmd_AddAPICommands_Good_RegistersBothCommandGroups verifies the root
// command registrar wires the spec and SDK command groups onto the Core
// command tree.
func TestCmd_AddAPICommands_Good_RegistersBothCommandGroups(t *testing.T) {
	c := core.New()

	AddAPICommands(c)

	for _, path := range []string{"api/spec", "build/spec", "api/sdk", "build/sdk"} {
		r := c.Command(path)
		if !r.OK {
			t.Fatalf("expected %s command to be registered", path)
		}
		cmd, ok := r.Value.(*core.Command)
		if !ok {
			t.Fatalf("expected *core.Command for %s, got %T", path, r.Value)
		}
		if cmd.Action == nil {
			t.Fatalf("expected non-nil Action on %s", path)
		}
		if cmd.Description == "" {
			t.Fatalf("expected Description on %s", path)
		}
	}
}
