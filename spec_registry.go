// SPDX-License-Identifier: EUPL-1.2

package api

import "sync"

// specRegistry stores RouteGroups that should be included in CLI-generated
// OpenAPI documents. Packages can register their groups during init and the
// API CLI will pick them up when building specs or SDKs.
var specRegistry struct {
	mu     sync.RWMutex
	groups []RouteGroup
}

// RegisterSpecGroups adds route groups to the package-level spec registry.
// Nil groups are ignored. Registered groups are returned by RegisteredSpecGroups
// in the order they were added.
func RegisterSpecGroups(groups ...RouteGroup) {
	specRegistry.mu.Lock()
	defer specRegistry.mu.Unlock()

	for _, group := range groups {
		if group == nil {
			continue
		}
		specRegistry.groups = append(specRegistry.groups, group)
	}
}

// RegisteredSpecGroups returns a copy of the route groups registered for
// CLI-generated OpenAPI documents.
func RegisteredSpecGroups() []RouteGroup {
	specRegistry.mu.RLock()
	defer specRegistry.mu.RUnlock()

	out := make([]RouteGroup, len(specRegistry.groups))
	copy(out, specRegistry.groups)
	return out
}
