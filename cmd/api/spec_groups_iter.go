// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"
	"reflect"

	"github.com/gin-gonic/gin"

	goapi "dappco.re/go/api"
)

// specGroupsIter snapshots the registered spec groups and appends one optional
// extra group. It keeps the command paths iterator-backed while preserving the
// existing ordering guarantees.
func specGroupsIter(extra goapi.RouteGroup) iter.Seq[goapi.RouteGroup] {
	return goapi.SpecGroupsIter(extra)
}

// specToolBridge returns a tool bridge populated from any registered
// ToolBridge instances so CLI-generated specs can describe the actual bundled
// tool routes rather than an empty placeholder bridge.
func specToolBridge(basePath string) *goapi.ToolBridge {
	bridge := goapi.NewToolBridge(basePath)
	seen := map[string]struct{}{}

	for _, group := range goapi.RegisteredSpecGroups() {
		if isNilRouteGroup(group) {
			continue
		}
		source, ok := group.(interface {
			ToolsIter() iter.Seq[goapi.ToolDescriptor]
		})
		if !ok {
			continue
		}
		for desc := range source.ToolsIter() {
			key := desc.Name + "\x00" + desc.Group
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			bridge.Add(desc, noopToolHandler)
		}
	}

	return bridge
}

func noopToolHandler(*gin.Context) {}

func isNilRouteGroup(group goapi.RouteGroup) bool {
	if group == nil {
		return true
	}

	value := reflect.ValueOf(group)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
