// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io" // Note: AX-6 — io.Writer is part of the public export API surface.
	"iter"

	"gopkg.in/yaml.v3"

	core "dappco.re/go"
	coreerr "dappco.re/go/log"
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
	switch core.Lower(core.Trim(format)) {
	case "json":
		_, err := w.Write(data)
		return err
	case "yaml":
		// Unmarshal JSON then re-marshal as YAML.
		var obj any
		decoded := core.JSONUnmarshal(data, &obj)
		if !decoded.OK {
			if err, ok := decoded.Value.(error); ok {
				return coreerr.E(op, "unmarshal spec", err)
			}
			return coreerr.E(op, "unmarshal spec", nil)
		}
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(obj); err != nil {
			return coreerr.E(op, "encode yaml", err)
		}
		return enc.Close()
	default:
		return coreerr.E(op, core.Sprintf("unsupported format %s: use %q or %q", format, "json", "yaml"), nil)
	}
}

// ExportSpecToFile writes the spec to the given path.
// The parent directory is created if it does not exist.
//
// Example:
//
//	_ = api.ExportSpecToFile("./api/openapi.yaml", "yaml", builder, engine.Groups())
func ExportSpecToFile(path, format string, builder *SpecBuilder, groups []RouteGroup) error {
	return exportSpecToFile(path, "ExportSpecToFile", func(w io.Writer) error {
		return ExportSpec(w, format, builder, groups)
	})
}

// ExportSpecToFileIter writes the OpenAPI spec from an iterator to the given path.
// The parent directory is created if it does not exist.
//
// Example:
//
//	_ = api.ExportSpecToFileIter("./api/openapi.json", "json", builder, api.RegisteredSpecGroupsIter())
func ExportSpecToFileIter(path, format string, builder *SpecBuilder, groups iter.Seq[RouteGroup]) error {
	return exportSpecToFile(path, "ExportSpecToFileIter", func(w io.Writer) error {
		return ExportSpecIter(w, format, builder, groups)
	})
}

func exportSpecToFile(path, op string, write func(io.Writer) error) (err error) {
	buf := core.NewBuffer()
	if writeErr := write(buf); writeErr != nil {
		return writeErr
	}

	localFS := (&core.Fs{}).NewUnrestricted()
	if result := localFS.WriteAtomic(path, buf.String()); !result.OK {
		err, _ := result.Value.(error)
		return coreerr.E(op, "write spec file", err)
	}
	return nil
}
