// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"fmt"
	"os"

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
		licenseName string
		licenseURL  string
		servers     string
	)

	cmd := cli.NewCommand("spec", "Generate OpenAPI specification", "", func(cmd *cli.Command, args []string) error {
		// Build spec from all route groups registered for CLI generation.
		builder := &goapi.SpecBuilder{
			Title:       title,
			Description: description,
			Version:     version,
			Servers:     parseServers(servers),
			LicenseName: licenseName,
			LicenseURL:  licenseURL,
		}

		bridge := goapi.NewToolBridge("/tools")
		groups := append(goapi.RegisteredSpecGroups(), bridge)

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
	cli.StringFlag(cmd, &licenseName, "license-name", "", "", "OpenAPI licence name in spec")
	cli.StringFlag(cmd, &licenseURL, "license-url", "", "", "OpenAPI licence URL in spec")
	cli.StringFlag(cmd, &servers, "server", "S", "", "Comma-separated OpenAPI server URL(s)")

	parent.AddCommand(cmd)
}

func parseServers(raw string) []string {
	return splitUniqueCSV(raw)
}
