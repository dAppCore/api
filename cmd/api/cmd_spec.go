// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"fmt"
	"os"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"

	goapi "dappco.re/go/core/api"
)

func addSpecCommand(parent *cli.Command) {
	var (
		output      string
		format      string
		title       string
		description string
		version     string
		servers     string
	)

	cmd := cli.NewCommand("spec", "Generate OpenAPI specification", "", func(cmd *cli.Command, args []string) error {
		// Build spec from registered route groups.
		// Additional groups can be added here as the platform grows.
		builder := &goapi.SpecBuilder{
			Title:       title,
			Description: description,
			Version:     version,
			Servers:     parseServers(servers),
		}

		// Start with the default tool bridge — future versions will
		// auto-populate from the MCP tool registry once the bridge
		// integration lands in the local go-ai module.
		bridge := goapi.NewToolBridge("/tools")
		groups := []goapi.RouteGroup{bridge}

		if output != "" {
			if err := goapi.ExportSpecToFile(output, format, builder, groups); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Spec written to %s\n", output)
			return nil
		}

		return goapi.ExportSpec(os.Stdout, format, builder, groups)
	})

	cli.StringFlag(cmd, &output, "output", "o", "", "Write spec to file instead of stdout")
	cli.StringFlag(cmd, &format, "format", "f", "json", "Output format: json or yaml")
	cli.StringFlag(cmd, &title, "title", "t", "Lethean Core API", "API title in spec")
	cli.StringFlag(cmd, &description, "description", "d", "Lethean Core API", "API description in spec")
	cli.StringFlag(cmd, &version, "version", "V", "1.0.0", "API version in spec")
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")

	parent.AddCommand(cmd)
}

func parseServers(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	servers := make([]string, 0, len(parts))
	for _, part := range parts {
		if server := strings.TrimSpace(part); server != "" {
			servers = append(servers, server)
		}
	}

	return servers
}
