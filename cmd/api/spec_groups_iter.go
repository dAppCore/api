// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"iter"

	goapi "dappco.re/go/core/api"
)

// specGroupsIter snapshots the registered spec groups and appends one optional
// extra group. It keeps the command paths iterator-backed while preserving the
// existing ordering guarantees.
func specGroupsIter(extra goapi.RouteGroup) iter.Seq[goapi.RouteGroup] {
	return goapi.SpecGroupsIter(extra)
}
