// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"
	"slices"

	core "dappco.re/go"
)

// specRegistry stores RouteGroups that should be included in CLI-generated
// OpenAPI documents. Packages can register their groups during init and the
// API CLI will pick them up when building specs or SDKs.
var specRegistry struct {
	mu     core.RWMutex
	groups []RouteGroup
}

// RegisterSpecGroups adds route groups to the package-level spec registry.
// Nil groups are ignored. Registered groups are returned by RegisteredSpecGroups
// in the order they were added.
//
// Example:
//
//	api.RegisterSpecGroups(api.NewToolBridge("/mcp"))
func RegisterSpecGroups(groups ...RouteGroup) {
	RegisterSpecGroupsIter(slices.Values(groups))
}

// RegisterSpecGroupsIter adds route groups from an iterator to the package-level
// spec registry.
//
// Nil groups are ignored. Registered groups are returned by RegisteredSpecGroups
// in the order they were added.
//
// Example:
//
//	api.RegisterSpecGroupsIter(api.RegisteredSpecGroupsIter())
func RegisterSpecGroupsIter(groups iter.Seq[RouteGroup]) {
	if groups == nil {
		return
	}

	specRegistry.mu.Lock()
	defer specRegistry.mu.Unlock()

	for group := range groups {
		if group == nil {
			continue
		}
		if specRegistryContains(group) {
			continue
		}
		specRegistry.groups = append(specRegistry.groups, group)
	}
}

// RegisteredSpecGroups returns a copy of the route groups registered for
// CLI-generated OpenAPI documents.
//
// Example:
//
//	groups := api.RegisteredSpecGroups()
func RegisteredSpecGroups() []RouteGroup {
	specRegistry.mu.RLock()
	defer specRegistry.mu.RUnlock()

	out := make([]RouteGroup, len(specRegistry.groups))
	copy(out, specRegistry.groups)
	return out
}

// RegisteredSpecGroupsIter returns an iterator over the route groups registered
// for CLI-generated OpenAPI documents.
//
// The iterator snapshots the current registry contents so callers can range
// over it without holding the registry lock.
//
// Example:
//
//	for g := range api.RegisteredSpecGroupsIter() {
//		_ = g
//	}
func RegisteredSpecGroupsIter() iter.Seq[RouteGroup] {
	specRegistry.mu.RLock()
	groups := slices.Clone(specRegistry.groups)
	specRegistry.mu.RUnlock()

	return slices.Values(groups)
}

// SpecGroupsIter returns the registered spec groups plus one optional extra
// group, deduplicated by group identity.
//
// The iterator snapshots the registry before yielding so callers can range
// over it without holding the registry lock.
//
// Example:
//
//	for g := range api.SpecGroupsIter(api.NewToolBridge("/tools")) {
//		_ = g
//	}
func SpecGroupsIter(extra RouteGroup) iter.Seq[RouteGroup] {
	return func(yield func(RouteGroup) bool) {
		seen := map[string]struct{}{}
		for group := range RegisteredSpecGroupsIter() {
			key := specGroupKey(group)
			seen[key] = struct{}{}
			if !yield(group) {
				return
			}
		}
		if extra != nil {
			if _, ok := seen[specGroupKey(extra)]; ok {
				return
			}
			if !yield(extra) {
				return
			}
		}
	}
}

// ResetSpecGroups clears the package-level spec registry.
// It is primarily intended for tests that need to isolate global state.
//
// Example:
//
//	api.ResetSpecGroups()
func ResetSpecGroups() {
	specRegistry.mu.Lock()
	defer specRegistry.mu.Unlock()

	specRegistry.groups = nil
}

func specRegistryContains(group RouteGroup) bool {
	key := specGroupKey(group)
	for _, existing := range specRegistry.groups {
		if specGroupKey(existing) == key {
			return true
		}
	}
	return false
}

func specGroupKey(group RouteGroup) string {
	if group == nil {
		return ""
	}

	return group.Name() + "\x00" + group.BasePath()
}
