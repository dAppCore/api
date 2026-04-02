// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"
	"slices"
	"strings"
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
	Path                    string
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
// Swagger, transport, cache, i18n, and Authentik metadata.
//
// Example:
//
//	builder := engine.OpenAPISpecBuilder()
func (e *Engine) OpenAPISpecBuilder() *SpecBuilder {
	if e == nil {
		return &SpecBuilder{}
	}

	runtime := e.RuntimeConfig()
	builder := &SpecBuilder{
		Title:                   runtime.Swagger.Title,
		Summary:                 runtime.Swagger.Summary,
		Description:             runtime.Swagger.Description,
		Version:                 runtime.Swagger.Version,
		SwaggerEnabled:          runtime.Swagger.Enabled,
		TermsOfService:          runtime.Swagger.TermsOfService,
		ContactName:             runtime.Swagger.ContactName,
		ContactURL:              runtime.Swagger.ContactURL,
		ContactEmail:            runtime.Swagger.ContactEmail,
		Servers:                 slices.Clone(runtime.Swagger.Servers),
		LicenseName:             runtime.Swagger.LicenseName,
		LicenseURL:              runtime.Swagger.LicenseURL,
		SecuritySchemes:         cloneSecuritySchemes(runtime.Swagger.SecuritySchemes),
		ExternalDocsDescription: runtime.Swagger.ExternalDocsDescription,
		ExternalDocsURL:         runtime.Swagger.ExternalDocsURL,
	}

	builder.SwaggerPath = runtime.Transport.SwaggerPath
	builder.GraphQLEnabled = runtime.GraphQL.Enabled
	builder.GraphQLPath = runtime.GraphQL.Path
	builder.GraphQLPlayground = runtime.GraphQL.Playground
	builder.WSPath = runtime.Transport.WSPath
	builder.WSEnabled = runtime.Transport.WSEnabled
	builder.SSEPath = runtime.Transport.SSEPath
	builder.SSEEnabled = runtime.Transport.SSEEnabled
	builder.PprofEnabled = runtime.Transport.PprofEnabled
	builder.ExpvarEnabled = runtime.Transport.ExpvarEnabled

	builder.CacheEnabled = runtime.Cache.Enabled
	if runtime.Cache.TTL > 0 {
		builder.CacheTTL = runtime.Cache.TTL.String()
	}
	builder.CacheMaxEntries = runtime.Cache.MaxEntries
	builder.CacheMaxBytes = runtime.Cache.MaxBytes

	builder.I18nDefaultLocale = runtime.I18n.DefaultLocale
	builder.I18nSupportedLocales = slices.Clone(runtime.I18n.Supported)
	builder.AuthentikIssuer = runtime.Authentik.Issuer
	builder.AuthentikClientID = runtime.Authentik.ClientID
	builder.AuthentikTrustedProxy = runtime.Authentik.TrustedProxy
	builder.AuthentikPublicPaths = slices.Clone(runtime.Authentik.PublicPaths)

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

	cfg := SwaggerConfig{
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

	if strings.TrimSpace(e.swaggerPath) != "" {
		cfg.Path = normaliseSwaggerPath(e.swaggerPath)
	}

	return cfg
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

func cloneRouteDescription(rd RouteDescription) RouteDescription {
	out := rd

	out.Tags = slices.Clone(rd.Tags)
	out.Security = cloneSecurityRequirements(rd.Security)
	out.Parameters = cloneParameterDescriptions(rd.Parameters)
	out.RequestBody = cloneOpenAPIObject(rd.RequestBody)
	out.RequestExample = cloneOpenAPIValue(rd.RequestExample)
	out.Response = cloneOpenAPIObject(rd.Response)
	out.ResponseExample = cloneOpenAPIValue(rd.ResponseExample)
	out.ResponseHeaders = cloneStringMap(rd.ResponseHeaders)

	return out
}

func cloneParameterDescriptions(params []ParameterDescription) []ParameterDescription {
	if params == nil {
		return nil
	}
	if len(params) == 0 {
		return []ParameterDescription{}
	}

	out := make([]ParameterDescription, len(params))
	for i, param := range params {
		out[i] = param
		out[i].Schema = cloneOpenAPIObject(param.Schema)
		out[i].Example = cloneOpenAPIValue(param.Example)
	}

	return out
}

func cloneSecurityRequirements(security []map[string][]string) []map[string][]string {
	if security == nil {
		return nil
	}
	if len(security) == 0 {
		return []map[string][]string{}
	}

	out := make([]map[string][]string, len(security))
	for i, requirement := range security {
		if len(requirement) == 0 {
			continue
		}

		cloned := make(map[string][]string, len(requirement))
		for name, scopes := range requirement {
			cloned[name] = slices.Clone(scopes)
		}
		out[i] = cloned
	}

	return out
}

func cloneOpenAPIObject(v map[string]any) map[string]any {
	if v == nil {
		return nil
	}
	if len(v) == 0 {
		return map[string]any{}
	}

	cloned, _ := cloneOpenAPIValue(v).(map[string]any)
	return cloned
}

func cloneStringMap(v map[string]string) map[string]string {
	if v == nil {
		return nil
	}
	if len(v) == 0 {
		return map[string]string{}
	}

	out := make(map[string]string, len(v))
	for key, value := range v {
		out[key] = value
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
