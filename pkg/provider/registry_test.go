// SPDX-Licence-Identifier: EUPL-1.2

package provider_test

import (
	"testing"

	"dappco.re/go/api"
	"dappco.re/go/api/pkg/provider"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRegistry_Add_Good(t *testing.T) {
	reg := provider.NewRegistry()
	assert.Equal(t, 0, reg.Len())

	reg.Add(&stubProvider{})
	assert.Equal(t, 1, reg.Len())

	reg.Add(&streamableProvider{})
	assert.Equal(t, 2, reg.Len())
}

func TestRegistry_Get_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})

	p := reg.Get("stub")
	require.NotNil(t, p)
	assert.Equal(t, "stub", p.Name())
}

func TestRegistry_Get_Bad(t *testing.T) {
	reg := provider.NewRegistry()
	p := reg.Get("nonexistent")
	assert.Nil(t, p)
}

func TestRegistry_List_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	list := reg.List()
	assert.Len(t, list, 2)
}

func TestRegistry_MountAll_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	engine, err := api.New()
	require.NoError(t, err)

	reg.MountAll(engine)
	assert.Len(t, engine.Groups(), 2)
}

func TestRegistry_Streamable_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})       // not streamable
	reg.Add(&streamableProvider{}) // streamable

	s := reg.Streamable()
	assert.Len(t, s, 1)
	assert.Equal(t, []string{"stub.event"}, s[0].Channels())
}

func TestRegistry_StreamableIter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	var streamables []provider.Streamable
	for s := range reg.StreamableIter() {
		streamables = append(streamables, s)
	}

	assert.Len(t, streamables, 1)
	assert.Equal(t, []string{"stub.event"}, streamables[0].Channels())
}

func TestRegistry_StreamableIter_Good_SnapshotCurrentProviders(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&streamableProvider{})

	iter := reg.StreamableIter()
	reg.Add(&streamableProvider{})

	var streamables []provider.Streamable
	for s := range iter {
		streamables = append(streamables, s)
	}

	assert.Len(t, streamables, 1)
	assert.Equal(t, []string{"stub.event"}, streamables[0].Channels())
}

func TestRegistry_Describable_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})        // not describable
	reg.Add(&describableProvider{}) // describable

	d := reg.Describable()
	assert.Len(t, d, 1)
	assert.Len(t, d[0].Describe(), 1)
}

func TestRegistry_DescribableIter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&describableProvider{})

	var describables []provider.Describable
	for d := range reg.DescribableIter() {
		describables = append(describables, d)
	}

	assert.Len(t, describables, 1)
	assert.Len(t, describables[0].Describe(), 1)
}

func TestRegistry_DescribableIter_Good_SnapshotCurrentProviders(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&describableProvider{})

	iter := reg.DescribableIter()
	reg.Add(&describableProvider{})

	var describables []provider.Describable
	for d := range iter {
		describables = append(describables, d)
	}

	assert.Len(t, describables, 1)
	assert.Len(t, describables[0].Describe(), 1)
}

func TestRegistry_Renderable_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})       // not renderable
	reg.Add(&renderableProvider{}) // renderable

	r := reg.Renderable()
	assert.Len(t, r, 1)
	assert.Equal(t, "core-stub-panel", r[0].Element().Tag)
}

func TestRegistry_RenderableIter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&renderableProvider{})

	var renderables []provider.Renderable
	for r := range reg.RenderableIter() {
		renderables = append(renderables, r)
	}

	assert.Len(t, renderables, 1)
	assert.Equal(t, "core-stub-panel", renderables[0].Element().Tag)
}

func TestRegistry_RenderableIter_Good_SnapshotCurrentProviders(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&renderableProvider{})

	iter := reg.RenderableIter()
	reg.Add(&renderableProvider{})

	var renderables []provider.Renderable
	for r := range iter {
		renderables = append(renderables, r)
	}

	assert.Len(t, renderables, 1)
	assert.Equal(t, "core-stub-panel", renderables[0].Element().Tag)
}

func TestRegistry_Info_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	infos := reg.Info()
	require.Len(t, infos, 1)

	info := infos[0]
	assert.Equal(t, "full", info.Name)
	assert.Equal(t, "/api/full", info.BasePath)
	assert.Equal(t, []string{"stub.event"}, info.Channels)
	require.NotNil(t, info.Element)
	assert.Equal(t, "core-full-panel", info.Element.Tag)
}

func TestRegistry_Info_Good_ProxyMetadata(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(provider.NewProxy(provider.ProxyConfig{
		Name:     "proxy",
		BasePath: "/api/proxy",
		Upstream: "http://127.0.0.1:9999",
		SpecFile: "/tmp/proxy-openapi.json",
	}))

	infos := reg.Info()
	require.Len(t, infos, 1)

	info := infos[0]
	assert.Equal(t, "proxy", info.Name)
	assert.Equal(t, "/api/proxy", info.BasePath)
	assert.Equal(t, "/tmp/proxy-openapi.json", info.SpecFile)
	assert.Equal(t, "http://127.0.0.1:9999", info.Upstream)
}

func TestRegistry_InfoIter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	var infos []provider.ProviderInfo
	for info := range reg.InfoIter() {
		infos = append(infos, info)
	}

	require.Len(t, infos, 1)
	info := infos[0]
	assert.Equal(t, "full", info.Name)
	assert.Equal(t, "/api/full", info.BasePath)
	assert.Equal(t, []string{"stub.event"}, info.Channels)
	require.NotNil(t, info.Element)
	assert.Equal(t, "core-full-panel", info.Element.Tag)
}

func TestRegistry_InfoIter_Good_SnapshotCurrentProviders(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&fullProvider{})

	iter := reg.InfoIter()
	reg.Add(&specFileProvider{specFile: "/tmp/later.json"})

	var infos []provider.ProviderInfo
	for info := range iter {
		infos = append(infos, info)
	}

	require.Len(t, infos, 1)
	assert.Equal(t, "full", infos[0].Name)
}

func TestRegistry_Iter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&streamableProvider{})

	count := 0
	for range reg.Iter() {
		count++
	}
	assert.Equal(t, 2, count)
}

func TestRegistry_SpecFiles_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&stubProvider{})
	reg.Add(&specFileProvider{specFile: "/tmp/b.json"})
	reg.Add(&specFileProvider{specFile: "/tmp/a.yaml"})
	reg.Add(&specFileProvider{specFile: "/tmp/a.yaml"})
	reg.Add(&specFileProvider{specFile: ""})

	assert.Equal(t, []string{"/tmp/a.yaml", "/tmp/b.json"}, reg.SpecFiles())
}

func TestRegistry_SpecFilesIter_Good(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Add(&specFileProvider{specFile: "/tmp/z.json"})
	reg.Add(&specFileProvider{specFile: "/tmp/x.json"})

	var files []string
	for file := range reg.SpecFilesIter() {
		files = append(files, file)
	}

	assert.Equal(t, []string{"/tmp/x.json", "/tmp/z.json"}, files)
}
