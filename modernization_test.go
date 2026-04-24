// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"slices"
	"testing"

	api "dappco.re/go/api"
)

func TestEngine_GroupsIter(t *testing.T) {
	e, _ := api.New()
	g1 := &healthGroup{}
	e.Register(g1)

	var groups []api.RouteGroup
	for g := range e.GroupsIter() {
		groups = append(groups, g)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name() != "health-extra" {
		t.Errorf("expected group name 'health-extra', got %q", groups[0].Name())
	}
}

func TestEngine_GroupsIter_Good_SnapshotsCurrentGroups(t *testing.T) {
	e, _ := api.New()
	g1 := &healthGroup{}
	g2 := &stubGroup{}
	e.Register(g1)

	iter := e.GroupsIter()
	e.Register(g2)

	var groups []api.RouteGroup
	for g := range iter {
		groups = append(groups, g)
	}

	if len(groups) != 1 {
		t.Fatalf("expected iterator snapshot to contain 1 group, got %d", len(groups))
	}
	if groups[0].Name() != "health-extra" {
		t.Fatalf("expected snapshot to preserve original group, got %q", groups[0].Name())
	}
}

type streamGroupStub struct {
	healthGroup
	channels []string
}

func (s *streamGroupStub) Channels() []string { return s.channels }

func TestEngine_ChannelsIter(t *testing.T) {
	e, _ := api.New()
	g1 := &streamGroupStub{channels: []string{"ch1", "ch2"}}
	g2 := &streamGroupStub{channels: []string{"ch3"}}
	e.Register(g1)
	e.Register(g2)

	var channels []string
	for ch := range e.ChannelsIter() {
		channels = append(channels, ch)
	}

	expected := []string{"ch1", "ch2", "ch3"}
	if !slices.Equal(channels, expected) {
		t.Fatalf("expected channels %v, got %v", expected, channels)
	}
}

func TestEngine_ChannelsIter_Good_SnapshotsCurrentChannels(t *testing.T) {
	e, _ := api.New()
	g1 := &streamGroupStub{channels: []string{"ch1", "ch2"}}
	g2 := &streamGroupStub{channels: []string{"ch3"}}
	e.Register(g1)

	iter := e.ChannelsIter()
	e.Register(g2)

	var channels []string
	for ch := range iter {
		channels = append(channels, ch)
	}

	expected := []string{"ch1", "ch2"}
	if !slices.Equal(channels, expected) {
		t.Fatalf("expected snapshot channels %v, got %v", expected, channels)
	}
}

func TestEngine_AuthentikConfig_Good_SnapshotsCurrentSettings(t *testing.T) {
	e, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		Issuer:       "https://auth.example.com",
		ClientID:     "client",
		TrustedProxy: true,
		PublicPaths:  []string{"/public", "/docs"},
	}))

	cfg := e.AuthentikConfig()
	if cfg.Issuer != "https://auth.example.com" {
		t.Fatalf("expected issuer https://auth.example.com, got %q", cfg.Issuer)
	}
	if cfg.ClientID != "client" {
		t.Fatalf("expected client ID client, got %q", cfg.ClientID)
	}
	if !cfg.TrustedProxy {
		t.Fatal("expected trusted proxy to be enabled")
	}
	if !slices.Equal(cfg.PublicPaths, []string{"/public", "/docs"}) {
		t.Fatalf("expected public paths [/public /docs], got %v", cfg.PublicPaths)
	}
}

func TestEngine_AuthentikConfig_Good_ClonesPublicPaths(t *testing.T) {
	publicPaths := []string{"/public", "/docs"}
	e, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		Issuer:      "https://auth.example.com",
		PublicPaths: publicPaths,
	}))

	cfg := e.AuthentikConfig()
	publicPaths[0] = "/mutated"

	if cfg.PublicPaths[0] != "/public" {
		t.Fatalf("expected snapshot to preserve original public paths, got %v", cfg.PublicPaths)
	}
}

func TestEngine_AuthentikConfig_Good_NormalisesPublicPaths(t *testing.T) {
	e, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		PublicPaths: []string{" /public/ ", "docs", "/public"},
	}))

	cfg := e.AuthentikConfig()
	expected := []string{"/public", "/docs"}
	if !slices.Equal(cfg.PublicPaths, expected) {
		t.Fatalf("expected normalised public paths %v, got %v", expected, cfg.PublicPaths)
	}
}

func TestEngine_AuthentikConfig_Good_BlankPublicPathsRemainNil(t *testing.T) {
	e, _ := api.New(api.WithAuthentik(api.AuthentikConfig{
		PublicPaths: []string{" ", "\t", ""},
	}))

	cfg := e.AuthentikConfig()
	if cfg.PublicPaths != nil {
		t.Fatalf("expected blank public paths to collapse to nil, got %v", cfg.PublicPaths)
	}
}

func TestEngine_Register_Good_IgnoresNilGroups(t *testing.T) {
	e, _ := api.New()

	var nilGroup *healthGroup
	e.Register(nilGroup)

	g1 := &healthGroup{}
	e.Register(g1)

	groups := e.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 registered group, got %d", len(groups))
	}
	if groups[0].Name() != "health-extra" {
		t.Fatalf("expected the original group to be preserved, got %q", groups[0].Name())
	}
}

func TestToolBridge_Iterators(t *testing.T) {
	b := api.NewToolBridge("/tools")
	desc := api.ToolDescriptor{Name: "test", Group: "g1"}
	b.Add(desc, nil)

	// Test ToolsIter
	var tools []api.ToolDescriptor
	for t := range b.ToolsIter() {
		tools = append(tools, t)
	}
	if len(tools) != 1 || tools[0].Name != "test" {
		t.Errorf("ToolsIter failed, got %v", tools)
	}

	// Test DescribeIter — emits the GET listing entry followed by each tool.
	var descs []api.RouteDescription
	for d := range b.DescribeIter() {
		descs = append(descs, d)
	}
	if len(descs) != 2 {
		t.Errorf("DescribeIter failed: expected listing + 1 tool, got %v", descs)
	}
	if descs[0].Method != "GET" || descs[0].Path != "/" {
		t.Errorf("DescribeIter failed: expected first entry to be GET / listing, got %+v", descs[0])
	}
	if descs[1].Path != "/test" {
		t.Errorf("DescribeIter failed: expected tool entry at /test, got %+v", descs[1])
	}
}

func TestToolBridge_Iterators_Good_SnapshotCurrentTools(t *testing.T) {
	b := api.NewToolBridge("/tools")
	b.Add(api.ToolDescriptor{Name: "first", Group: "g1"}, nil)

	toolsIter := b.ToolsIter()
	descsIter := b.DescribeIter()

	b.Add(api.ToolDescriptor{Name: "second", Group: "g2"}, nil)

	var tools []api.ToolDescriptor
	for tool := range toolsIter {
		tools = append(tools, tool)
	}

	var descs []api.RouteDescription
	for desc := range descsIter {
		descs = append(descs, desc)
	}

	if len(tools) != 1 || tools[0].Name != "first" {
		t.Fatalf("expected ToolsIter snapshot to contain the original tool, got %v", tools)
	}
	// DescribeIter snapshots the listing entry plus the tool registered at call time.
	if len(descs) != 2 {
		t.Fatalf("expected DescribeIter snapshot to contain listing + original tool, got %v", descs)
	}
	if descs[0].Method != "GET" || descs[0].Path != "/" {
		t.Fatalf("expected DescribeIter snapshot to start with the GET listing entry, got %+v", descs[0])
	}
	if descs[1].Path != "/first" {
		t.Fatalf("expected DescribeIter snapshot to contain the original tool, got %v", descs)
	}
}

func TestCodegen_SupportedLanguagesIter(t *testing.T) {
	var langs []string
	for l := range api.SupportedLanguagesIter() {
		langs = append(langs, l)
	}

	if !slices.Contains(langs, "go") {
		t.Errorf("SupportedLanguagesIter missing 'go'")
	}

	// Should be sorted
	if !slices.IsSorted(langs) {
		t.Errorf("SupportedLanguagesIter should be sorted, got %v", langs)
	}
}
