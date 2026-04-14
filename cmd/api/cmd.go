// SPDX-License-Identifier: EUPL-1.2

// Package api registers the `core api` command group on the root Core
// instance. It exposes two subcommands:
//
//   - api/spec — generate the OpenAPI specification from registered route
//     groups plus the built-in tool bridge and write it to stdout or a file.
//   - api/sdk  — run openapi-generator-cli over a generated spec to produce
//     client SDKs in the configured target languages.
//
// The commands use the Core framework's declarative Command API. Flags are
// declared via the Flags Options map and read from the incoming Options
// during the Action.
package api

import (
	"dappco.re/go/core"
	"dappco.re/go/core/cli/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddAPICommands)
}

// AddAPICommands registers the `api/spec` and `api/sdk` commands on the given
// Core instance.
//
//	core.RegisterCommands(api.AddAPICommands)
func AddAPICommands(c *core.Core) {
	addSpecCommand(c)
	addSDKCommand(c)
}
