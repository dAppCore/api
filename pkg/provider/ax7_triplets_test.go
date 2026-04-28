// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	coretest "dappco.re/go"
	"dappco.re/go/api"

	"github.com/gin-gonic/gin"
)

type ax7Provider struct {
	name     string
	basePath string
	channels []string
	element  ElementSpec
	specFile string
	upstream string
}

func (p *ax7Provider) Name() string     { return p.name }
func (p *ax7Provider) BasePath() string { return p.basePath }
func (p *ax7Provider) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/status", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}
func (p *ax7Provider) Channels() []string { return p.channels }
func (p *ax7Provider) Describe() []api.RouteDescription {
	return []api.RouteDescription{{Method: http.MethodGet, Path: "/status", Tags: []string{p.name}}}
}
func (p *ax7Provider) Element() ElementSpec { return p.element }
func (p *ax7Provider) SpecFile() string     { return p.specFile }
func (p *ax7Provider) Upstream() string     { return p.upstream }

func ax7ProviderOne() *ax7Provider {
	return &ax7Provider{
		name:     "alpha",
		basePath: "/api/alpha",
		channels: []string{"alpha.ready"},
		element:  ElementSpec{Tag: "core-alpha", Source: "/assets/alpha.js"},
		specFile: "/tmp/alpha.yaml",
		upstream: "http://1.1.1.1",
	}
}

func ax7WriteManifest(t *coretest.T, dir, name, upstream string) {
	t.Helper()
	coretest.RequireNoError(t, os.MkdirAll(dir, 0755))
	coretest.RequireNoError(t, os.WriteFile(filepath.Join(dir, name+".yaml"), []byte("name: "+name+"\nbasePath: /api/"+name+"\nupstream: "+upstream+"\n"), 0644))
}

func TestAX7_ProviderUpstreamBlockedError_Error_Good(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{Reason: "loopback IP"}
	text := err.Error()
	coretest.AssertContains(t, text, ErrProviderUpstreamBlocked.Error())
	coretest.AssertContains(t, text, "loopback IP")
}

func TestAX7_ProviderUpstreamBlockedError_Error_Bad(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{}
	text := err.Error()
	coretest.AssertEqual(t, ErrProviderUpstreamBlocked.Error(), text)
	coretest.AssertNotContains(t, text, ":")
}

func TestAX7_ProviderUpstreamBlockedError_Error_Ugly(t *coretest.T) {
	var err *ProviderUpstreamBlockedError
	text := err.Error()
	coretest.AssertEqual(t, ErrProviderUpstreamBlocked.Error(), text)
	coretest.AssertNotEmpty(t, text)
}

func TestAX7_ProviderUpstreamBlockedError_Is_Good(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{Reason: "metadata host"}
	ok := errors.Is(err, ErrProviderUpstreamBlocked)
	coretest.AssertTrue(t, ok)
	coretest.AssertTrue(t, err.Is(ErrProviderUpstreamBlocked))
}

func TestAX7_ProviderUpstreamBlockedError_Is_Bad(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{Reason: "metadata host"}
	ok := errors.Is(err, errors.New("other"))
	coretest.AssertFalse(t, ok)
	coretest.AssertFalse(t, err.Is(errors.New("other")))
}

func TestAX7_ProviderUpstreamBlockedError_Is_Ugly(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{}
	ok := errors.Is(err, ErrProviderUpstreamBlocked)
	coretest.AssertTrue(t, ok)
	coretest.AssertTrue(t, err.Is(ErrProviderUpstreamBlocked))
}

func TestAX7_ProviderUpstreamBlockedError_Unwrap_Good(t *coretest.T) {
	cause := errors.New("dns failed")
	err := &ProviderUpstreamBlockedError{Cause: cause}
	got := err.Unwrap()
	coretest.AssertEqual(t, cause, got)
	coretest.AssertErrorIs(t, err, cause)
}

func TestAX7_ProviderUpstreamBlockedError_Unwrap_Bad(t *coretest.T) {
	err := &ProviderUpstreamBlockedError{}
	got := err.Unwrap()
	coretest.AssertNil(t, got)
	coretest.AssertFalse(t, errors.Is(err, errors.New("missing")))
}

func TestAX7_ProviderUpstreamBlockedError_Unwrap_Ugly(t *coretest.T) {
	var err *ProviderUpstreamBlockedError
	got := err.Unwrap()
	coretest.AssertNil(t, got)
	coretest.AssertNotPanics(t, func() { _ = err.Unwrap() })
}

func TestAX7_Discover_Good(t *coretest.T) {
	dir := filepath.Join(t.TempDir(), "providers")
	ax7WriteManifest(t, dir, "alpha", "http://1.1.1.1")
	providers, err := Discover(dir)
	coretest.RequireNoError(t, err)
	coretest.AssertLen(t, providers, 1)
	coretest.AssertEqual(t, "alpha", providers[0].Name())
}

func TestAX7_Discover_Bad(t *coretest.T) {
	providers, err := Discover(filepath.Join(t.TempDir(), "missing"))
	coretest.RequireNoError(t, err)
	coretest.AssertNil(t, providers)
	coretest.AssertEmpty(t, providers)
}

func TestAX7_Discover_Ugly(t *coretest.T) {
	dir := filepath.Join(t.TempDir(), "providers")
	coretest.RequireNoError(t, os.MkdirAll(dir, 0755))
	coretest.RequireNoError(t, os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: bad\n"), 0644))
	providers, err := Discover(dir)
	coretest.AssertError(t, err)
	coretest.AssertNil(t, providers)
}

func TestAX7_DiscoverDefault_Good(t *coretest.T) {
	t.Chdir(t.TempDir())
	dir := filepath.Join(DefaultProvidersDir)
	ax7WriteManifest(t, dir, "defaulted", "http://1.1.1.1")
	providers, err := DiscoverDefault()
	coretest.RequireNoError(t, err)
	coretest.AssertLen(t, providers, 1)
	coretest.AssertEqual(t, "defaulted", providers[0].Name())
}

func TestAX7_DiscoverDefault_Bad(t *coretest.T) {
	t.Chdir(t.TempDir())
	providers, err := DiscoverDefault()
	coretest.RequireNoError(t, err)
	coretest.AssertNil(t, providers)
	coretest.AssertEmpty(t, providers)
}

func TestAX7_DiscoverDefault_Ugly(t *coretest.T) {
	t.Chdir(t.TempDir())
	coretest.RequireNoError(t, os.MkdirAll(DefaultProvidersDir, 0755))
	coretest.RequireNoError(t, os.WriteFile(filepath.Join(DefaultProvidersDir, "broken.yaml"), []byte("name: broken\n"), 0644))
	providers, err := DiscoverDefault()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, providers)
}

func TestAX7_NewProxy_Good(t *coretest.T) {
	t.Setenv(providerUpstreamAllowEnv, "")
	proxy := NewProxy(ProxyConfig{Name: "public", BasePath: "/api/public", Upstream: "http://1.1.1.1"})
	coretest.AssertNotNil(t, proxy)
	coretest.AssertNoError(t, proxy.Err())
	coretest.AssertEqual(t, "public", proxy.Name())
}

func TestAX7_NewProxy_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "bad", BasePath: "/api/bad", Upstream: "://not-a-url"})
	coretest.AssertNotNil(t, proxy)
	coretest.AssertError(t, proxy.Err())
	coretest.AssertEqual(t, "bad", proxy.Name())
}

func TestAX7_NewProxy_Ugly(t *coretest.T) {
	t.Setenv(providerUpstreamAllowEnv, "127.0.0.0/8")
	proxy := NewProxy(ProxyConfig{Name: "loopback", BasePath: "/api/loopback", Upstream: "http://127.0.0.1:8080"})
	coretest.AssertNotNil(t, proxy)
	coretest.AssertNoError(t, proxy.Err())
	coretest.AssertEqual(t, "http://127.0.0.1:8080", proxy.Upstream())
}

func TestAX7_ProxyProvider_Err_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "public", BasePath: "/api/public", Upstream: "http://1.1.1.1"})
	err := proxy.Err()
	coretest.AssertNoError(t, err)
	coretest.AssertNil(t, err)
}

func TestAX7_ProxyProvider_Err_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "bad", BasePath: "/api/bad", Upstream: "not-a-host"})
	err := proxy.Err()
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "scheme")
}

func TestAX7_ProxyProvider_Err_Ugly(t *coretest.T) {
	var proxy *ProxyProvider
	err := proxy.Err()
	coretest.AssertNoError(t, err)
	coretest.AssertNil(t, err)
}

func TestAX7_ProxyProvider_Name_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", BasePath: "/api/alpha", Upstream: "http://1.1.1.1"})
	name := proxy.Name()
	coretest.AssertEqual(t, "alpha", name)
	coretest.AssertNotEmpty(t, name)
}

func TestAX7_ProxyProvider_Name_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{BasePath: "/api/alpha", Upstream: "http://1.1.1.1"})
	name := proxy.Name()
	coretest.AssertEqual(t, "", name)
	coretest.AssertEmpty(t, name)
}

func TestAX7_ProxyProvider_Name_Ugly(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: " spaced ", BasePath: "/api/alpha", Upstream: "http://1.1.1.1"})
	name := proxy.Name()
	coretest.AssertEqual(t, " spaced ", name)
	coretest.AssertContains(t, name, " ")
}

func TestAX7_ProxyProvider_BasePath_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", BasePath: "/api/alpha", Upstream: "http://1.1.1.1"})
	path := proxy.BasePath()
	coretest.AssertEqual(t, "/api/alpha", path)
	coretest.AssertTrue(t, coretest.HasPrefix(path, "/"))
}

func TestAX7_ProxyProvider_BasePath_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1"})
	path := proxy.BasePath()
	coretest.AssertEqual(t, "", path)
	coretest.AssertEmpty(t, path)
}

func TestAX7_ProxyProvider_BasePath_Ugly(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", BasePath: "/api/alpha/", Upstream: "http://1.1.1.1"})
	path := proxy.BasePath()
	coretest.AssertEqual(t, "/api/alpha/", path)
	coretest.AssertContains(t, path, "alpha")
}

func TestAX7_ProxyProvider_RegisterRoutes_Good(t *coretest.T) {
	t.Setenv(providerUpstreamAllowEnv, "127.0.0.0/8")
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) }))
	defer upstream.Close()
	proxy := NewProxy(ProxyConfig{Name: "proxy", BasePath: "/api/proxy", Upstream: upstream.URL})
	router := gin.New()
	proxy.RegisterRoutes(&router.RouterGroup)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	coretest.AssertEqual(t, http.StatusAccepted, rec.Code)
}

func TestAX7_ProxyProvider_RegisterRoutes_Bad(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	proxy := NewProxy(ProxyConfig{Name: "bad", BasePath: "/api/bad", Upstream: "://bad"})
	router := gin.New()
	proxy.RegisterRoutes(&router.RouterGroup)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	coretest.AssertEqual(t, http.StatusInternalServerError, rec.Code)
}

func TestAX7_ProxyProvider_RegisterRoutes_Ugly(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	var proxy *ProxyProvider
	router := gin.New()
	coretest.AssertNotPanics(t, func() {
		proxy.RegisterRoutes(&router.RouterGroup)
	})
	coretest.AssertNil(t, proxy)
}

func TestAX7_ProxyProvider_Element_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1", Element: ElementSpec{Tag: "core-alpha", Source: "/alpha.js"}})
	element := proxy.Element()
	coretest.AssertEqual(t, "core-alpha", element.Tag)
	coretest.AssertEqual(t, "/alpha.js", element.Source)
}

func TestAX7_ProxyProvider_Element_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1"})
	element := proxy.Element()
	coretest.AssertEqual(t, "", element.Tag)
	coretest.AssertEqual(t, "", element.Source)
}

func TestAX7_ProxyProvider_Element_Ugly(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1", Element: ElementSpec{Tag: "x", Source: ""}})
	element := proxy.Element()
	coretest.AssertEqual(t, "x", element.Tag)
	coretest.AssertEmpty(t, element.Source)
}

func TestAX7_ProxyProvider_SpecFile_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1", SpecFile: "/tmp/spec.yaml"})
	specFile := proxy.SpecFile()
	coretest.AssertEqual(t, "/tmp/spec.yaml", specFile)
	coretest.AssertContains(t, specFile, "spec")
}

func TestAX7_ProxyProvider_SpecFile_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1"})
	specFile := proxy.SpecFile()
	coretest.AssertEqual(t, "", specFile)
	coretest.AssertEmpty(t, specFile)
}

func TestAX7_ProxyProvider_SpecFile_Ugly(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1", SpecFile: "../specs/openapi.yaml"})
	specFile := proxy.SpecFile()
	coretest.AssertEqual(t, "../specs/openapi.yaml", specFile)
	coretest.AssertContains(t, specFile, "..")
}

func TestAX7_ProxyProvider_Upstream_Good(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1"})
	upstream := proxy.Upstream()
	coretest.AssertEqual(t, "http://1.1.1.1", upstream)
	coretest.AssertContains(t, upstream, "http")
}

func TestAX7_ProxyProvider_Upstream_Bad(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: ""})
	upstream := proxy.Upstream()
	coretest.AssertEqual(t, "", upstream)
	coretest.AssertEmpty(t, upstream)
}

func TestAX7_ProxyProvider_Upstream_Ugly(t *coretest.T) {
	proxy := NewProxy(ProxyConfig{Name: "alpha", Upstream: "http://1.1.1.1/path?x=1"})
	upstream := proxy.Upstream()
	coretest.AssertEqual(t, "http://1.1.1.1/path?x=1", upstream)
	coretest.AssertContains(t, upstream, "?x=1")
}

func TestAX7_NewRegistry_Good(t *coretest.T) {
	reg := NewRegistry()
	coretest.AssertNotNil(t, reg)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_NewRegistry_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Get("missing")
	coretest.AssertNil(t, got)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_NewRegistry_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	list := reg.List()
	coretest.AssertLen(t, list, 1)
	coretest.AssertNil(t, list[0])
}

func TestAX7_Registry_Add_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	list := reg.List()
	coretest.AssertLen(t, list, 1)
	coretest.AssertEqual(t, "alpha", list[0].Name())
}

func TestAX7_Registry_Add_Bad(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	list := reg.List()
	coretest.AssertLen(t, list, 1)
	coretest.AssertNil(t, list[0])
}

func TestAX7_Registry_Add_Ugly(t *coretest.T) {
	reg := NewRegistry()
	provider := ax7ProviderOne()
	reg.Add(provider)
	reg.Add(provider)
	coretest.AssertEqual(t, 2, reg.Len())
}

func TestAX7_Registry_MountAll_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	engine, err := api.New()
	coretest.RequireNoError(t, err)
	reg.MountAll(engine)
	coretest.AssertLen(t, engine.Groups(), 1)
}

func TestAX7_Registry_MountAll_Bad(t *coretest.T) {
	reg := NewRegistry()
	engine, err := api.New()
	coretest.RequireNoError(t, err)
	reg.MountAll(engine)
	coretest.AssertLen(t, engine.Groups(), 0)
}

func TestAX7_Registry_MountAll_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	engine, err := api.New()
	coretest.RequireNoError(t, err)
	reg.MountAll(engine)
	coretest.AssertLen(t, engine.Groups(), 0)
}

func TestAX7_Registry_List_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	list := reg.List()
	coretest.AssertLen(t, list, 1)
	coretest.AssertEqual(t, "alpha", list[0].Name())
}

func TestAX7_Registry_List_Bad(t *coretest.T) {
	reg := NewRegistry()
	list := reg.List()
	coretest.AssertEmpty(t, list)
	coretest.AssertLen(t, list, 0)
}

func TestAX7_Registry_List_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	list := reg.List()
	list[0] = nil
	coretest.AssertNotNil(t, reg.List()[0])
}

func TestAX7_Registry_Iter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	count := 0
	for range reg.Iter() {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_Iter_Bad(t *coretest.T) {
	reg := NewRegistry()
	count := 0
	for range reg.Iter() {
		count++
	}
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_Registry_Iter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	iter := reg.Iter()
	reg.Add(ax7ProviderOne())
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_Len_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	got := reg.Len()
	coretest.AssertEqual(t, 1, got)
	coretest.AssertGreater(t, got, 0)
}

func TestAX7_Registry_Len_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Len()
	coretest.AssertEqual(t, 0, got)
	coretest.AssertFalse(t, got > 0)
}

func TestAX7_Registry_Len_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	got := reg.Len()
	coretest.AssertEqual(t, 1, got)
	coretest.AssertGreaterOrEqual(t, got, 1)
}

func TestAX7_Registry_Get_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	got := reg.Get("alpha")
	coretest.AssertNotNil(t, got)
	coretest.AssertEqual(t, "alpha", got.Name())
}

func TestAX7_Registry_Get_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Get("missing")
	coretest.AssertNil(t, got)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_Registry_Get_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(&ax7Provider{name: "", basePath: "/"})
	got := reg.Get("")
	coretest.AssertNotNil(t, got)
	coretest.AssertEqual(t, "", got.Name())
}

func TestAX7_Registry_Streamable_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	got := reg.Streamable()
	coretest.AssertLen(t, got, 1)
	coretest.AssertEqual(t, []string{"alpha.ready"}, got[0].Channels())
}

func TestAX7_Registry_Streamable_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Streamable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_Streamable_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	got := reg.Streamable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_StreamableIter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	var got []Streamable
	for p := range reg.StreamableIter() {
		got = append(got, p)
	}
	coretest.AssertLen(t, got, 1)
}

func TestAX7_Registry_StreamableIter_Bad(t *coretest.T) {
	reg := NewRegistry()
	var got []Streamable
	for p := range reg.StreamableIter() {
		got = append(got, p)
	}
	coretest.AssertEmpty(t, got)
}

func TestAX7_Registry_StreamableIter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	iter := reg.StreamableIter()
	reg.Add(ax7ProviderOne())
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_Describable_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	got := reg.Describable()
	coretest.AssertLen(t, got, 1)
	coretest.AssertLen(t, got[0].Describe(), 1)
}

func TestAX7_Registry_Describable_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Describable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_Describable_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	got := reg.Describable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_DescribableIter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	var got []Describable
	for p := range reg.DescribableIter() {
		got = append(got, p)
	}
	coretest.AssertLen(t, got, 1)
}

func TestAX7_Registry_DescribableIter_Bad(t *coretest.T) {
	reg := NewRegistry()
	var got []Describable
	for p := range reg.DescribableIter() {
		got = append(got, p)
	}
	coretest.AssertEmpty(t, got)
}

func TestAX7_Registry_DescribableIter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	iter := reg.DescribableIter()
	reg.Add(ax7ProviderOne())
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_Renderable_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	got := reg.Renderable()
	coretest.AssertLen(t, got, 1)
	coretest.AssertEqual(t, "core-alpha", got[0].Element().Tag)
}

func TestAX7_Registry_Renderable_Bad(t *coretest.T) {
	reg := NewRegistry()
	got := reg.Renderable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_Renderable_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(nil)
	got := reg.Renderable()
	coretest.AssertEmpty(t, got)
	coretest.AssertLen(t, got, 0)
}

func TestAX7_Registry_RenderableIter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	var got []Renderable
	for p := range reg.RenderableIter() {
		got = append(got, p)
	}
	coretest.AssertLen(t, got, 1)
}

func TestAX7_Registry_RenderableIter_Bad(t *coretest.T) {
	reg := NewRegistry()
	var got []Renderable
	for p := range reg.RenderableIter() {
		got = append(got, p)
	}
	coretest.AssertEmpty(t, got)
}

func TestAX7_Registry_RenderableIter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	iter := reg.RenderableIter()
	reg.Add(ax7ProviderOne())
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_Info_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	info := reg.Info()
	coretest.AssertLen(t, info, 1)
	coretest.AssertEqual(t, "alpha", info[0].Name)
}

func TestAX7_Registry_Info_Bad(t *coretest.T) {
	reg := NewRegistry()
	info := reg.Info()
	coretest.AssertEmpty(t, info)
	coretest.AssertLen(t, info, 0)
}

func TestAX7_Registry_Info_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(&ax7Provider{name: "minimal", basePath: "/m"})
	info := reg.Info()
	coretest.AssertLen(t, info, 1)
	coretest.AssertNotNil(t, info[0].Element)
	coretest.AssertEqual(t, "", info[0].Element.Tag)
}

func TestAX7_Registry_InfoIter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	var info []ProviderInfo
	for item := range reg.InfoIter() {
		info = append(info, item)
	}
	coretest.AssertLen(t, info, 1)
}

func TestAX7_Registry_InfoIter_Bad(t *coretest.T) {
	reg := NewRegistry()
	var info []ProviderInfo
	for item := range reg.InfoIter() {
		info = append(info, item)
	}
	coretest.AssertEmpty(t, info)
}

func TestAX7_Registry_InfoIter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	iter := reg.InfoIter()
	reg.Add(ax7ProviderOne())
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_Registry_SpecFiles_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	files := reg.SpecFiles()
	coretest.AssertLen(t, files, 1)
	coretest.AssertEqual(t, "/tmp/alpha.yaml", files[0])
}

func TestAX7_Registry_SpecFiles_Bad(t *coretest.T) {
	reg := NewRegistry()
	files := reg.SpecFiles()
	coretest.AssertEmpty(t, files)
	coretest.AssertLen(t, files, 0)
}

func TestAX7_Registry_SpecFiles_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(&ax7Provider{name: "a", specFile: "/tmp/b.yaml"})
	reg.Add(&ax7Provider{name: "b", specFile: "/tmp/a.yaml"})
	files := reg.SpecFiles()
	coretest.AssertEqual(t, []string{"/tmp/a.yaml", "/tmp/b.yaml"}, files)
}

func TestAX7_Registry_SpecFilesIter_Good(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(ax7ProviderOne())
	var files []string
	for file := range reg.SpecFilesIter() {
		files = append(files, file)
	}
	coretest.AssertEqual(t, []string{"/tmp/alpha.yaml"}, files)
}

func TestAX7_Registry_SpecFilesIter_Bad(t *coretest.T) {
	reg := NewRegistry()
	var files []string
	for file := range reg.SpecFilesIter() {
		files = append(files, file)
	}
	coretest.AssertEmpty(t, files)
}

func TestAX7_Registry_SpecFilesIter_Ugly(t *coretest.T) {
	reg := NewRegistry()
	reg.Add(&ax7Provider{name: "a", specFile: "/tmp/a.yaml"})
	reg.Add(&ax7Provider{name: "b", specFile: "/tmp/a.yaml"})
	var files []string
	for file := range reg.SpecFilesIter() {
		files = append(files, file)
	}
	coretest.AssertEqual(t, []string{"/tmp/a.yaml"}, files)
}

func TestAX7_Registry_Discover_Good(t *coretest.T) {
	dir := filepath.Join(t.TempDir(), "providers")
	ax7WriteManifest(t, dir, "alpha", "http://1.1.1.1")
	reg := NewRegistry()
	err := reg.Discover(dir)
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 1, reg.Len())
}

func TestAX7_Registry_Discover_Bad(t *coretest.T) {
	reg := NewRegistry()
	err := reg.Discover(filepath.Join(t.TempDir(), "missing"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_Registry_Discover_Ugly(t *coretest.T) {
	dir := filepath.Join(t.TempDir(), "providers")
	coretest.RequireNoError(t, os.MkdirAll(dir, 0755))
	coretest.RequireNoError(t, os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: bad\n"), 0644))
	reg := NewRegistry()
	err := reg.Discover(dir)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_Registry_DiscoverDefault_Good(t *coretest.T) {
	t.Chdir(t.TempDir())
	ax7WriteManifest(t, DefaultProvidersDir, "alpha", "http://1.1.1.1")
	reg := NewRegistry()
	err := reg.DiscoverDefault()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 1, reg.Len())
}

func TestAX7_Registry_DiscoverDefault_Bad(t *coretest.T) {
	t.Chdir(t.TempDir())
	reg := NewRegistry()
	err := reg.DiscoverDefault()
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, reg.Len())
}

func TestAX7_Registry_DiscoverDefault_Ugly(t *coretest.T) {
	t.Chdir(t.TempDir())
	coretest.RequireNoError(t, os.MkdirAll(DefaultProvidersDir, 0755))
	coretest.RequireNoError(t, os.WriteFile(filepath.Join(DefaultProvidersDir, "bad.yaml"), []byte("name: bad\n"), 0644))
	reg := NewRegistry()
	err := reg.DiscoverDefault()
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, 0, reg.Len())
}
