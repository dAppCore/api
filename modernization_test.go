// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"slices"
	"testing"
	"time"

	api "dappco.re/go/core/api"
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

func TestEngine_CacheConfig_Good_SnapshotsCurrentSettings(t *testing.T) {
	e, _ := api.New(api.WithCacheLimits(5*time.Minute, 10, 1024))

	cfg := e.CacheConfig()

	if !cfg.Enabled {
		t.Fatal("expected cache config to be enabled")
	}
	if cfg.TTL != 5*time.Minute {
		t.Fatalf("expected TTL %v, got %v", 5*time.Minute, cfg.TTL)
	}
	if cfg.MaxEntries != 10 {
		t.Fatalf("expected MaxEntries 10, got %d", cfg.MaxEntries)
	}
	if cfg.MaxBytes != 1024 {
		t.Fatalf("expected MaxBytes 1024, got %d", cfg.MaxBytes)
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

	// Test DescribeIter
	var descs []api.RouteDescription
	for d := range b.DescribeIter() {
		descs = append(descs, d)
	}
	if len(descs) != 1 || descs[0].Path != "/test" {
		t.Errorf("DescribeIter failed, got %v", descs)
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
	if len(descs) != 1 || descs[0].Path != "/first" {
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
