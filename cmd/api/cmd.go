// SPDX-License-Identifier: EUPL-1.2

package api

import "dappco.re/go/core/cli/pkg/cli"

func init() {
	cli.RegisterCommands(AddAPICommands)
}

// AddAPICommands registers the `api` command group.
//
// Example:
//
//	root := cli.NewGroup("root", "", "")
//	api.AddAPICommands(root)
func AddAPICommands(root *cli.Command) {
	apiCmd := cli.NewGroup("api", "API specification and SDK generation", "")
	root.AddCommand(apiCmd)

	addSpecCommand(apiCmd)
	addSDKCommand(apiCmd)
}
