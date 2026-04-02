// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"
	"slices"
)

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

	transport := e.TransportConfig()
	builder := &SpecBuilder{
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

	builder.SwaggerPath = transport.SwaggerPath
	builder.GraphQLPath = transport.GraphQLPath
	builder.GraphQLPlayground = transport.GraphQLPlayground
	builder.WSPath = transport.WSPath
	builder.SSEPath = transport.SSEPath
	builder.PprofEnabled = transport.PprofEnabled
	builder.ExpvarEnabled = transport.ExpvarEnabled

	return builder
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
