// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
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

	return writeSpec(w, format, data, "ExportSpec")
}

// ExportSpecIter generates the OpenAPI spec from an iterator and writes it to w.
// Format must be "json" or "yaml".
//
// Example:
//
//	_ = api.ExportSpecIter(os.Stdout, "json", builder, api.RegisteredSpecGroupsIter())
func ExportSpecIter(w io.Writer, format string, builder *SpecBuilder, groups iter.Seq[RouteGroup]) error {
	data, err := builder.BuildIter(groups)
	if err != nil {
		return coreerr.E("ExportSpecIter", "build spec", err)
	}

	return writeSpec(w, format, data, "ExportSpecIter")
}

func writeSpec(w io.Writer, format string, data []byte, op string) error {
	switch format {
	case "json":
		_, err := w.Write(data)
		return err
	case "yaml":
		// Unmarshal JSON then re-marshal as YAML.
		var obj any
		if err := json.Unmarshal(data, &obj); err != nil {
			return coreerr.E(op, "unmarshal spec", err)
		}
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(obj); err != nil {
			return coreerr.E(op, "encode yaml", err)
		}
		return enc.Close()
	default:
		return coreerr.E(op, fmt.Sprintf("unsupported format %s: use %q or %q", format, "json", "yaml"), nil)
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
	if err := ExportSpec(f, format, builder, groups); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return coreerr.E("ExportSpecToFile", "close file", err)
	}
	return nil
}

// ExportSpecToFileIter writes the OpenAPI spec from an iterator to the given path.
// The parent directory is created if it does not exist.
//
// Example:
//
//	_ = api.ExportSpecToFileIter("./api/openapi.json", "json", builder, api.RegisteredSpecGroupsIter())
func ExportSpecToFileIter(path, format string, builder *SpecBuilder, groups iter.Seq[RouteGroup]) error {
	if err := coreio.Local.EnsureDir(filepath.Dir(path)); err != nil {
		return coreerr.E("ExportSpecToFileIter", "create directory", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return coreerr.E("ExportSpecToFileIter", "create file", err)
	}
	if err := ExportSpecIter(f, format, builder, groups); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return coreerr.E("ExportSpecToFileIter", "close file", err)
	}
	return nil
}
