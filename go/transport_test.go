// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"
	"testing"
)

func TestTransport_TransportConfig_Ugly_NilEngineReturnsZeroValue(t *testing.T) {
	var e *Engine

	if got := e.TransportConfig(); !reflect.DeepEqual(got, TransportConfig{}) {
		t.Fatalf("expected zero-value transport config for nil engine, got %+v", got)
	}
}
