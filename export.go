// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
)

// ExportSpec generates the OpenAPI spec and writes it to w.
// Format must be "json" or "yaml".
//
// Example:
//
//	_ = api.ExportSpec(os.Stdout, "yaml", builder, engine.Groups())
func ExportSpec(w io.Writer, format string, builder *SpecBuilder, groups []RouteGroup) error {
	data, err := builder.Build(groups)
	if err != nil {
		return coreerr.E("ExportSpec", "build spec", err)
	}

	switch format {
	case "json":
		_, err = w.Write(data)
		return err
	case "yaml":
		// Unmarshal JSON then re-marshal as YAML.
		var obj any
		if err := json.Unmarshal(data, &obj); err != nil {
			return coreerr.E("ExportSpec", "unmarshal spec", err)
		}
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(obj); err != nil {
			return coreerr.E("ExportSpec", "encode yaml", err)
		}
		return enc.Close()
	default:
		return coreerr.E("ExportSpec", "unsupported format "+format+": use \"json\" or \"yaml\"", nil)
	}
}

// ExportSpecToFile writes the spec to the given path.
// The parent directory is created if it does not exist.
//
// Example:
//
//	_ = api.ExportSpecToFile("./api/openapi.yaml", "yaml", builder, engine.Groups())
func ExportSpecToFile(path, format string, builder *SpecBuilder, groups []RouteGroup) error {
	if err := coreio.Local.EnsureDir(filepath.Dir(path)); err != nil {
		return coreerr.E("ExportSpecToFile", "create directory", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return coreerr.E("ExportSpecToFile", "create file", err)
	}
	defer f.Close()
	return ExportSpec(f, format, builder, groups)
}
