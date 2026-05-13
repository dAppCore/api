// SPDX-License-Identifier: EUPL-1.2

package provider_test

import (
	. "dappco.re/go"
	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
	"github.com/gin-gonic/gin"
)

// -- Test helpers (minimal providers) -----------------------------------------

type stubProvider struct{}

func (s *stubProvider) Name() string                       { return "stub" }
func (s *stubProvider) BasePath() string                   { return "/api/stub" }
func (s *stubProvider) RegisterRoutes(rg *gin.RouterGroup) {}

type streamableProvider struct{ stubProvider }

func (s *streamableProvider) Channels() []string { return []string{"stub.event"} }

type describableProvider struct{ stubProvider }

func (d *describableProvider) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{Method: "GET", Path: "/items", Summary: "List items", Tags: []string{"stub"}},
	}
}

type renderableProvider struct{ stubProvider }

func (r *renderableProvider) Element() provider.ElementSpec {
	return provider.ElementSpec{Tag: "core-stub-panel", Source: "/assets/stub.js"}
}

type specFileProvider struct {
	stubProvider
	specFile string
}

func (s *specFileProvider) SpecFile() string { return s.specFile }

type fullProvider struct {
	streamableProvider
}

func (f *fullProvider) Name() string     { return "full" }
func (f *fullProvider) BasePath() string { return "/api/full" }
func (f *fullProvider) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{Method: "GET", Path: "/status", Summary: "Status", Tags: []string{"full"}},
	}
}
func (f *fullProvider) Element() provider.ElementSpec {
	return provider.ElementSpec{Tag: "core-full-panel", Source: "/assets/full.js"}
}

// -- Tests --------------------------------------------------------------------

func TestRegistry_Add_Good(t *T) {
	reg := provider.NewRegistry()
	AssertEqual(t, 0, reg.Len())

	reg.Add(&stubProvider{})
	AssertEqual(t, 1, reg.Len())

	reg.Add(&streamableProvider{})
	AssertEqual(t, 2, reg.Len())
}

func TestRegistry_Get_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})

	p := reg.Get("stub")
	AssertNotNil(t, p)
	AssertEqual(t, "stub", p.Name())
}

func TestRegistry_Get_Bad(t *T) {
	reg := provider.NewRegistry()
	p := reg.Get("nonexistent")
	AssertNil(t, p)
}

func TestRegistry_List_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	list := reg.List()
	AssertLen(t, list, 2)
}

func TestRegistry_MountAll_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	engine, err := api.New()
	RequireNoError(t, err)

	reg.MountAll(engine)
	AssertLen(t, engine.Groups(), 2)
}

func TestRegistry_Streamable_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})       // not streamable
	reg.Add(&streamableProvider{}) // streamable

	s := reg.Streamable()
	AssertLen(t, s, 1)
	AssertEqual(t, []string{"stub.event"}, s[0].Channels())
}

func TestRegistry_StreamableIter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	var streamables []provider.Streamable
	for s := range reg.StreamableIter() {
		streamables = append(streamables, s)
	}

	AssertLen(t, streamables, 1)
	AssertEqual(t, []string{"stub.event"}, streamables[0].Channels())
}

func TestRegistry_StreamableIter_Good_SnapshotCurrentProviders(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&streamableProvider{})

	iter := reg.StreamableIter()
	reg.Add(&streamableProvider{})

	var streamables []provider.Streamable
	for s := range iter {
		streamables = append(streamables, s)
	}

	AssertLen(t, streamables, 1)
	AssertEqual(t, []string{"stub.event"}, streamables[0].Channels())
}

func TestRegistry_Describable_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})        // not describable
	reg.Add(&describableProvider{}) // describable

	d := reg.Describable()
	AssertLen(t, d, 1)
	AssertLen(t, d[0].Describe(), 1)
}

func TestRegistry_DescribableIter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&describableProvider{})

	var describables []provider.Describable
	for d := range reg.DescribableIter() {
		describables = append(describables, d)
	}

	AssertLen(t, describables, 1)
	AssertLen(t, describables[0].Describe(), 1)
}

func TestRegistry_DescribableIter_Good_SnapshotCurrentProviders(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&describableProvider{})

	iter := reg.DescribableIter()
	reg.Add(&describableProvider{})

	var describables []provider.Describable
	for d := range iter {
		describables = append(describables, d)
	}

	AssertLen(t, describables, 1)
	AssertLen(t, describables[0].Describe(), 1)
}

func TestRegistry_Renderable_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})       // not renderable
	reg.Add(&renderableProvider{}) // renderable

	r := reg.Renderable()
	AssertLen(t, r, 1)
	AssertEqual(t, "core-stub-panel", r[0].Element().Tag)
}

func TestRegistry_RenderableIter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&renderableProvider{})

	var renderables []provider.Renderable
	for r := range reg.RenderableIter() {
		renderables = append(renderables, r)
	}

	AssertLen(t, renderables, 1)
	AssertEqual(t, "core-stub-panel", renderables[0].Element().Tag)
}

func TestRegistry_RenderableIter_Good_SnapshotCurrentProviders(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&renderableProvider{})

	iter := reg.RenderableIter()
	reg.Add(&renderableProvider{})

	var renderables []provider.Renderable
	for r := range iter {
		renderables = append(renderables, r)
	}

	AssertLen(t, renderables, 1)
	AssertEqual(t, "core-stub-panel", renderables[0].Element().Tag)
}

func TestRegistry_Info_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	infos := reg.Info()
	AssertLen(t, infos, 1)

	info := infos[0]
	AssertEqual(t, "full", info.Name)
	AssertEqual(t, "/api/full", info.BasePath)
	AssertEqual(t, []string{"stub.event"}, info.Channels)
	AssertNotNil(t, info.Element)
	AssertEqual(t, "core-full-panel", info.Element.Tag)
}

func TestRegistry_Info_Good_ProxyMetadata(t *T) {
	reg := provider.NewRegistry()
	reg.Add(provider.NewProxy(provider.ProxyConfig{
		Name:     "proxy",
		BasePath: "/api/proxy",
		Upstream: "http://127.0.0.1:9999",
		SpecFile: "/tmp/proxy-openapi.json",
	}))

	infos := reg.Info()
	AssertLen(t, infos, 1)

	info := infos[0]
	AssertEqual(t, "proxy", info.Name)
	AssertEqual(t, "/api/proxy", info.BasePath)
	AssertEqual(t, "/tmp/proxy-openapi.json", info.SpecFile)
	AssertEqual(t, "http://127.0.0.1:9999", info.Upstream)
}

func TestRegistry_InfoIter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	var infos []provider.ProviderInfo
	for info := range reg.InfoIter() {
		infos = append(infos, info)
	}

	AssertLen(t, infos, 1)
	info := infos[0]
	AssertEqual(t, "full", info.Name)
	AssertEqual(t, "/api/full", info.BasePath)
	AssertEqual(t, []string{"stub.event"}, info.Channels)
	AssertNotNil(t, info.Element)
	AssertEqual(t, "core-full-panel", info.Element.Tag)
}

func TestRegistry_InfoIter_Good_SnapshotCurrentProviders(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	iter := reg.InfoIter()
	reg.Add(&specFileProvider{specFile: "/tmp/later.json"})

	var infos []provider.ProviderInfo
	for info := range iter {
		infos = append(infos, info)
	}

	AssertLen(t, infos, 1)
	AssertEqual(t, "full", infos[0].Name)
}

func TestRegistry_Iter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	count := 0
	for range reg.Iter() {
		count++
	}
	AssertEqual(t, 2, count)
}

func TestRegistry_SpecFiles_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&specFileProvider{specFile: "/tmp/b.json"})
	reg.Add(&specFileProvider{specFile: "/tmp/a.yaml"})
	reg.Add(&specFileProvider{specFile: "/tmp/a.yaml"})
	reg.Add(&specFileProvider{specFile: ""})

	AssertEqual(t, []string{"/tmp/a.yaml", "/tmp/b.json"}, reg.SpecFiles())
}

func TestRegistry_SpecFilesIter_Good(t *T) {
	reg := provider.NewRegistry()
	reg.Add(&specFileProvider{specFile: "/tmp/z.json"})
	reg.Add(&specFileProvider{specFile: "/tmp/x.json"})

	var files []string
	for file := range reg.SpecFilesIter() {
		files = append(files, file)
	}

	AssertEqual(t, []string{"/tmp/x.json", "/tmp/z.json"}, files)
}
