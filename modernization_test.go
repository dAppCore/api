// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"slices"
	"testing"

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
