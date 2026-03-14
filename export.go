// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ExportSpec generates the OpenAPI spec and writes it to w.
// Format must be "json" or "yaml".
func ExportSpec(w io.Writer, format string, builder *SpecBuilder, groups []RouteGroup) error {
	data, err := builder.Build(groups)
	if err != nil {
		return fmt.Errorf("build spec: %w", err)
	}

	switch format {
	case "json":
		_, err = w.Write(data)
		return err
	case "yaml":
		// Unmarshal JSON then re-marshal as YAML.
		var obj any
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("unmarshal spec: %w", err)
		}
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(obj); err != nil {
			return fmt.Errorf("encode yaml: %w", err)
		}
		return enc.Close()
	default:
		return fmt.Errorf("unsupported format %q: use \"json\" or \"yaml\"", format)
	}
}

// ExportSpecToFile writes the spec to the given path.
// The parent directory is created if it does not exist.
func ExportSpecToFile(path, format string, builder *SpecBuilder, groups []RouteGroup) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	return ExportSpec(f, format, builder, groups)
}
