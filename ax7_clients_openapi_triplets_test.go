// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	coretest "dappco.re/go"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ax7RoundTrip func(*http.Request) (*http.Response, error)

func (f ax7RoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type ax7SpanExporter struct{}

func (ax7SpanExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (ax7SpanExporter) Shutdown(context.Context) error                             { return nil }

func ax7PublicDNS(t *coretest.T) {
	t.Helper()
	old := resolveHost
	resolveHost = func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("1.1.1.1")}, nil
	}
	t.Cleanup(func() { resolveHost = old })
}

func ax7OpenAPISpec(server string) string {
	return `{"openapi":"3.1.0","info":{"title":"AX7","version":"1"},"servers":[{"url":"` + server + `"}],"paths":{"/health":{"get":{"operationId":"getHealth","responses":{"200":{"description":"ok"}}}}}}`
}

func ax7OpenAPIClient(t *coretest.T) *OpenAPIClient {
	t.Helper()
	ax7PublicDNS(t)
	client := &http.Client{Transport: ax7RoundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Request:    r,
		}, nil
	})}
	return NewOpenAPIClient(
		WithSpecReader(strings.NewReader(ax7OpenAPISpec("http://public.test"))),
		WithHTTPClient(client),
	)
}

func TestAX7_WithSpec_Good(t *coretest.T) {
	client := NewOpenAPIClient(WithSpec("/tmp/openapi.json"))
	coretest.AssertEqual(t, "/tmp/openapi.json", client.specPath)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithSpec_Bad(t *coretest.T) {
	client := NewOpenAPIClient(WithSpec(""))
	coretest.AssertEqual(t, "", client.specPath)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithSpec_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithSpec("  spec.yaml  "))
	coretest.AssertEqual(t, "  spec.yaml  ", client.specPath)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithSpecReader_Good(t *coretest.T) {
	reader := strings.NewReader(ax7OpenAPISpec("http://public.test"))
	client := NewOpenAPIClient(WithSpecReader(reader))
	coretest.AssertNotNil(t, client.specReader)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithSpecReader_Bad(t *coretest.T) {
	client := NewOpenAPIClient(WithSpecReader(nil))
	coretest.AssertNil(t, client.specReader)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithSpecReader_Ugly(t *coretest.T) {
	reader := strings.NewReader("")
	client := NewOpenAPIClient(WithSpecReader(reader), WithSpec("/tmp/ignored.json"))
	coretest.AssertNotNil(t, client.specReader)
	coretest.AssertEqual(t, "/tmp/ignored.json", client.specPath)
}

func TestAX7_WithBaseURL_Good(t *coretest.T) {
	client := NewOpenAPIClient(WithBaseURL("https://api.example.com"))
	coretest.AssertEqual(t, "https://api.example.com", client.baseURL)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithBaseURL_Bad(t *coretest.T) {
	client := NewOpenAPIClient(WithBaseURL(""))
	coretest.AssertEqual(t, "", client.baseURL)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithBaseURL_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithBaseURL(" https://api.example.com "))
	coretest.AssertEqual(t, " https://api.example.com ", client.baseURL)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithBearerToken_Good(t *coretest.T) {
	client := NewOpenAPIClient(WithBearerToken("secret"))
	coretest.AssertEqual(t, "secret", client.bearerToken)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithBearerToken_Bad(t *coretest.T) {
	client := NewOpenAPIClient(WithBearerToken(""))
	coretest.AssertEqual(t, "", client.bearerToken)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithBearerToken_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithBearerToken(" token with spaces "))
	coretest.AssertEqual(t, " token with spaces ", client.bearerToken)
	coretest.AssertNotNil(t, client.httpClient)
}

func TestAX7_WithHTTPClient_Good(t *coretest.T) {
	httpClient := &http.Client{Timeout: time.Second}
	client := NewOpenAPIClient(WithHTTPClient(httpClient))
	coretest.AssertEqual(t, httpClient, client.httpClient)
}

func TestAX7_WithHTTPClient_Bad(t *coretest.T) {
	client := NewOpenAPIClient(WithHTTPClient(nil))
	coretest.AssertEqual(t, http.DefaultClient, client.httpClient)
	coretest.AssertEqual(t, "", client.baseURL)
}

func TestAX7_WithHTTPClient_Ugly(t *coretest.T) {
	httpClient := &http.Client{}
	client := NewOpenAPIClient(WithHTTPClient(httpClient), WithHTTPClient(nil))
	coretest.AssertEqual(t, http.DefaultClient, client.httpClient)
}

func TestAX7_NewOpenAPIClient_Good(t *coretest.T) {
	client := NewOpenAPIClient(WithBaseURL("https://api.example.com"))
	coretest.AssertNotNil(t, client)
	coretest.AssertEqual(t, "https://api.example.com", client.baseURL)
}

func TestAX7_NewOpenAPIClient_Bad(t *coretest.T) {
	client := NewOpenAPIClient()
	coretest.AssertNotNil(t, client)
	coretest.AssertEqual(t, http.DefaultClient, client.httpClient)
}

func TestAX7_NewOpenAPIClient_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithHTTPClient(nil))
	coretest.AssertNotNil(t, client)
	coretest.AssertEqual(t, http.DefaultClient, client.httpClient)
}

func TestAX7_OpenAPIClient_Operations_Good(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	ops, err := client.Operations()
	coretest.RequireNoError(t, err)
	coretest.AssertLen(t, ops, 1)
	coretest.AssertEqual(t, "getHealth", ops[0].OperationID)
}

func TestAX7_OpenAPIClient_Operations_Bad(t *coretest.T) {
	client := NewOpenAPIClient()
	ops, err := client.Operations()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, ops)
}

func TestAX7_OpenAPIClient_Operations_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithSpecReader(strings.NewReader(`{`)))
	ops, err := client.Operations()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, ops)
}

func TestAX7_OpenAPIClient_OperationsIter_Good(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	iter, err := client.OperationsIter()
	coretest.RequireNoError(t, err)
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_OpenAPIClient_OperationsIter_Bad(t *coretest.T) {
	client := NewOpenAPIClient()
	iter, err := client.OperationsIter()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, iter)
}

func TestAX7_OpenAPIClient_OperationsIter_Ugly(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	iter, err := client.OperationsIter()
	coretest.RequireNoError(t, err)
	client.operations = nil
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_OpenAPIClient_Servers_Good(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	servers, err := client.Servers()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, []string{"http://public.test"}, servers)
}

func TestAX7_OpenAPIClient_Servers_Bad(t *coretest.T) {
	client := NewOpenAPIClient()
	servers, err := client.Servers()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, servers)
}

func TestAX7_OpenAPIClient_Servers_Ugly(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	servers, err := client.Servers()
	coretest.RequireNoError(t, err)
	servers[0] = "mutated"
	fresh, err := client.Servers()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, "http://public.test", fresh[0])
}

func TestAX7_OpenAPIClient_ServersIter_Good(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	iter, err := client.ServersIter()
	coretest.RequireNoError(t, err)
	var servers []string
	for server := range iter {
		servers = append(servers, server)
	}
	coretest.AssertEqual(t, []string{"http://public.test"}, servers)
}

func TestAX7_OpenAPIClient_ServersIter_Bad(t *coretest.T) {
	client := NewOpenAPIClient()
	iter, err := client.ServersIter()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, iter)
}

func TestAX7_OpenAPIClient_ServersIter_Ugly(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	iter, err := client.ServersIter()
	coretest.RequireNoError(t, err)
	client.servers = nil
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_OpenAPIClient_Call_Good(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	got, err := client.Call("getHealth", nil)
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, true, got.(map[string]any)["ok"])
}

func TestAX7_OpenAPIClient_Call_Bad(t *coretest.T) {
	client := ax7OpenAPIClient(t)
	got, err := client.Call("missing", nil)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, got)
}

func TestAX7_OpenAPIClient_Call_Ugly(t *coretest.T) {
	client := NewOpenAPIClient(WithSpecReader(strings.NewReader(`{`)))
	got, err := client.Call("getHealth", nil)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, got)
}

func TestAX7_SpecBuilder_Build_Good(t *coretest.T) {
	data, err := (&SpecBuilder{Title: "API", Version: "1"}).Build(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), `"openapi"`)
}

func TestAX7_SpecBuilder_Build_Bad(t *coretest.T) {
	var builder *SpecBuilder
	data, err := builder.Build(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), `"paths"`)
}

func TestAX7_SpecBuilder_Build_Ugly(t *coretest.T) {
	data, err := (&SpecBuilder{Servers: []string{" https://api.example.com "}}).Build([]RouteGroup{nil})
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), "https://api.example.com")
}

func TestAX7_SpecBuilder_BuildIter_Good(t *coretest.T) {
	data, err := (&SpecBuilder{Title: "API"}).BuildIter(func(yield func(RouteGroup) bool) { yield(ax7RouteGroup{name: "alpha", basePath: "/alpha"}) })
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), `"tags"`)
}

func TestAX7_SpecBuilder_BuildIter_Bad(t *coretest.T) {
	data, err := (&SpecBuilder{}).BuildIter(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), `"openapi"`)
}

func TestAX7_SpecBuilder_BuildIter_Ugly(t *coretest.T) {
	var builder *SpecBuilder
	data, err := builder.BuildIter(func(yield func(RouteGroup) bool) { yield(nil) })
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, string(data), `"paths"`)
}

func TestAX7_ExportSpec_Good(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpec(buf, "json", &SpecBuilder{Title: "API"}, nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, buf.String(), `"openapi"`)
}

func TestAX7_ExportSpec_Bad(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpec(buf, "toml", &SpecBuilder{}, nil)
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "unsupported")
}

func TestAX7_ExportSpec_Ugly(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpec(buf, " yaml ", &SpecBuilder{Title: "API"}, nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, buf.String(), "openapi")
}

func TestAX7_ExportSpecIter_Good(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpecIter(buf, "json", &SpecBuilder{Title: "API"}, nil)
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, buf.String(), `"openapi"`)
}

func TestAX7_ExportSpecIter_Bad(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpecIter(buf, "bad", &SpecBuilder{}, nil)
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "unsupported")
}

func TestAX7_ExportSpecIter_Ugly(t *coretest.T) {
	buf := coretest.NewBuffer()
	err := ExportSpecIter(buf, "yaml", &SpecBuilder{}, func(yield func(RouteGroup) bool) { yield(nil) })
	coretest.RequireNoError(t, err)
	coretest.AssertContains(t, buf.String(), "openapi")
}

func TestAX7_ExportSpecToFile_Good(t *coretest.T) {
	path := coretest.Path(t.TempDir(), "openapi.json")
	err := ExportSpecToFile(path, "json", &SpecBuilder{}, nil)
	coretest.RequireNoError(t, err)
	_, statErr := os.Stat(path)
	coretest.AssertNoError(t, statErr)
}

func TestAX7_ExportSpecToFile_Bad(t *coretest.T) {
	err := ExportSpecToFile(coretest.Path(t.TempDir(), "openapi.txt"), "bad", &SpecBuilder{}, nil)
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "unsupported")
}

func TestAX7_ExportSpecToFile_Ugly(t *coretest.T) {
	path := coretest.Path(t.TempDir(), "nested", "openapi.yaml")
	err := ExportSpecToFile(path, "yaml", &SpecBuilder{}, nil)
	coretest.RequireNoError(t, err)
	_, statErr := os.Stat(path)
	coretest.AssertNoError(t, statErr)
}

func TestAX7_ExportSpecToFileIter_Good(t *coretest.T) {
	path := coretest.Path(t.TempDir(), "openapi.json")
	err := ExportSpecToFileIter(path, "json", &SpecBuilder{}, nil)
	coretest.RequireNoError(t, err)
	_, statErr := os.Stat(path)
	coretest.AssertNoError(t, statErr)
}

func TestAX7_ExportSpecToFileIter_Bad(t *coretest.T) {
	err := ExportSpecToFileIter(coretest.Path(t.TempDir(), "openapi.txt"), "bad", &SpecBuilder{}, nil)
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "unsupported")
}

func TestAX7_ExportSpecToFileIter_Ugly(t *coretest.T) {
	path := coretest.Path(t.TempDir(), "nested", "openapi.yaml")
	err := ExportSpecToFileIter(path, "yaml", &SpecBuilder{}, func(yield func(RouteGroup) bool) { yield(nil) })
	coretest.RequireNoError(t, err)
	_, statErr := os.Stat(path)
	coretest.AssertNoError(t, statErr)
}

func TestAX7_Spec_ReadDoc_Good(t *coretest.T) {
	spec := newSwaggerSpec(&SpecBuilder{Title: "API"}, nil)
	doc := spec.ReadDoc()
	coretest.AssertContains(t, doc, `"openapi"`)
}

func TestAX7_Spec_ReadDoc_Bad(t *coretest.T) {
	spec := newSwaggerSpec(nil, nil)
	doc := spec.ReadDoc()
	coretest.AssertContains(t, doc, `"openapi"`)
}

func TestAX7_Spec_ReadDoc_Ugly(t *coretest.T) {
	spec := newSwaggerSpec(&SpecBuilder{Title: "API"}, []RouteGroup{nil})
	first := spec.ReadDoc()
	second := spec.ReadDoc()
	coretest.AssertEqual(t, first, second)
}

func TestAX7_SDKGenerator_Generate_Good(t *coretest.T) {
	gen := &SDKGenerator{}
	err := gen.Generate(context.Background(), "go")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "spec path")
}

func TestAX7_SDKGenerator_Generate_Bad(t *coretest.T) {
	var gen *SDKGenerator
	err := gen.Generate(context.Background(), "go")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "nil")
}

func TestAX7_SDKGenerator_Generate_Ugly(t *coretest.T) {
	gen := &SDKGenerator{}
	err := gen.Generate(nil, "go")
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "context")
}

func TestAX7_SDKGenerator_Available_Good(t *coretest.T) {
	gen := &SDKGenerator{}
	available := gen.Available()
	coretest.AssertEqual(t, available, gen.Available())
}

func TestAX7_SDKGenerator_Available_Bad(t *coretest.T) {
	var gen *SDKGenerator
	available := gen.Available()
	coretest.AssertEqual(t, available, (&SDKGenerator{}).Available())
}

func TestAX7_SDKGenerator_Available_Ugly(t *coretest.T) {
	gen := &SDKGenerator{PackageName: "ignored"}
	available := gen.Available()
	coretest.AssertEqual(t, available, gen.Available())
}

func TestAX7_SupportedLanguages_Good(t *coretest.T) {
	langs := SupportedLanguages()
	coretest.AssertContains(t, langs, "go")
	coretest.AssertContains(t, langs, "python")
}

func TestAX7_SupportedLanguages_Bad(t *coretest.T) {
	langs := SupportedLanguages()
	langs[0] = "mutated"
	coretest.AssertNotEqual(t, "mutated", SupportedLanguages()[0])
}

func TestAX7_SupportedLanguages_Ugly(t *coretest.T) {
	langs := SupportedLanguages()
	coretest.AssertTrue(t, slices.IsSorted(langs))
	coretest.AssertNotEmpty(t, langs)
}

func TestAX7_SupportedLanguagesIter_Good(t *coretest.T) {
	var langs []string
	for lang := range SupportedLanguagesIter() {
		langs = append(langs, lang)
	}
	coretest.AssertContains(t, langs, "go")
}

func TestAX7_SupportedLanguagesIter_Bad(t *coretest.T) {
	count := 0
	for range SupportedLanguagesIter() {
		count++
		break
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_SupportedLanguagesIter_Ugly(t *coretest.T) {
	var langs []string
	for lang := range SupportedLanguagesIter() {
		langs = append(langs, lang)
	}
	coretest.AssertEqual(t, SupportedLanguages(), langs)
}
