// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"
	"slices"
)

// SwaggerConfig captures the configured Swagger/OpenAPI metadata for an Engine.
//
// It is intentionally small and serialisable so callers can inspect the active
// documentation surface without rebuilding an OpenAPI document.
//
// Example:
//
//	cfg := api.SwaggerConfig{Title: "Service", Summary: "Public API"}
type SwaggerConfig struct {
	Enabled                 bool
	Title                   string
	Summary                 string
	Description             string
	Version                 string
	TermsOfService          string
	ContactName             string
	ContactURL              string
	ContactEmail            string
	Servers                 []string
	LicenseName             string
	LicenseURL              string
	SecuritySchemes         map[string]any
	ExternalDocsDescription string
	ExternalDocsURL         string
}

// OpenAPISpecBuilder returns a SpecBuilder populated from the engine's current
// Swagger and transport metadata.
//
// Example:
//
//	builder := engine.OpenAPISpecBuilder()
func (e *Engine) OpenAPISpecBuilder() *SpecBuilder {
	if e == nil {
		return &SpecBuilder{}
	}

	swagger := e.SwaggerConfig()
	transport := e.TransportConfig()
	builder := &SpecBuilder{
		Title:                   swagger.Title,
		Summary:                 swagger.Summary,
		Description:             swagger.Description,
		Version:                 swagger.Version,
		TermsOfService:          swagger.TermsOfService,
		ContactName:             swagger.ContactName,
		ContactURL:              swagger.ContactURL,
		ContactEmail:            swagger.ContactEmail,
		Servers:                 slices.Clone(swagger.Servers),
		LicenseName:             swagger.LicenseName,
		LicenseURL:              swagger.LicenseURL,
		SecuritySchemes:         cloneSecuritySchemes(swagger.SecuritySchemes),
		ExternalDocsDescription: swagger.ExternalDocsDescription,
		ExternalDocsURL:         swagger.ExternalDocsURL,
	}

	builder.SwaggerPath = transport.SwaggerPath
	builder.GraphQLPath = transport.GraphQLPath
	builder.GraphQLPlayground = transport.GraphQLPlayground
	builder.WSPath = transport.WSPath
	builder.SSEPath = transport.SSEPath
	builder.PprofEnabled = transport.PprofEnabled
	builder.ExpvarEnabled = transport.ExpvarEnabled

	return builder
}

// SwaggerConfig returns the currently configured Swagger metadata for the engine.
//
// The result snapshots the Engine state at call time and clones slices/maps so
// callers can safely reuse or modify the returned value.
//
// Example:
//
//	cfg := engine.SwaggerConfig()
func (e *Engine) SwaggerConfig() SwaggerConfig {
	if e == nil {
		return SwaggerConfig{}
	}

	return SwaggerConfig{
		Enabled:                 e.swaggerEnabled,
		Title:                   e.swaggerTitle,
		Summary:                 e.swaggerSummary,
		Description:             e.swaggerDesc,
		Version:                 e.swaggerVersion,
		TermsOfService:          e.swaggerTermsOfService,
		ContactName:             e.swaggerContactName,
		ContactURL:              e.swaggerContactURL,
		ContactEmail:            e.swaggerContactEmail,
		Servers:                 slices.Clone(e.swaggerServers),
		LicenseName:             e.swaggerLicenseName,
		LicenseURL:              e.swaggerLicenseURL,
		SecuritySchemes:         cloneSecuritySchemes(e.swaggerSecuritySchemes),
		ExternalDocsDescription: e.swaggerExternalDocsDescription,
		ExternalDocsURL:         e.swaggerExternalDocsURL,
	}
}

func cloneSecuritySchemes(schemes map[string]any) map[string]any {
	if len(schemes) == 0 {
		return nil
	}

	out := make(map[string]any, len(schemes))
	for name, scheme := range schemes {
		out[name] = cloneOpenAPIValue(scheme)
	}
	return out
}

// cloneOpenAPIValue recursively copies JSON-like OpenAPI values so callers can
// safely retain and reuse their original maps after configuring an engine.
func cloneOpenAPIValue(v any) any {
	switch value := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for k, nested := range value {
			out[k] = cloneOpenAPIValue(nested)
		}
		return out
	case []any:
		out := make([]any, len(value))
		for i, nested := range value {
			out[i] = cloneOpenAPIValue(nested)
		}
		return out
	default:
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			return nil
		}

		switch rv.Kind() {
		case reflect.Map:
			out := reflect.MakeMapWithSize(rv.Type(), rv.Len())
			for _, key := range rv.MapKeys() {
				cloned := cloneOpenAPIValue(rv.MapIndex(key).Interface())
				if cloned == nil {
					out.SetMapIndex(key, reflect.Zero(rv.Type().Elem()))
					continue
				}
				out.SetMapIndex(key, reflect.ValueOf(cloned))
			}
			return out.Interface()
		case reflect.Slice:
			if rv.IsNil() {
				return v
			}
			out := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
			for i := 0; i < rv.Len(); i++ {
				cloned := cloneOpenAPIValue(rv.Index(i).Interface())
				if cloned == nil {
					out.Index(i).Set(reflect.Zero(rv.Type().Elem()))
					continue
				}
				out.Index(i).Set(reflect.ValueOf(cloned))
			}
			return out.Interface()
		default:
			return value
		}
	}
}
