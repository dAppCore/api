// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"

	api "dappco.re/go/api"
)

func ExampleWithUpstreamRouter() {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("127.0.0.0/8", "10.0.0.0/8"))
	_ = reg.Set("lemma",
		api.Upstream{URL: "http://10.0.0.5:8000", Weight: 2},
		api.Upstream{URL: "http://10.0.0.6:8000", Weight: 1},
	)
	_ = reg.SetDefault(api.Upstream{URL: "http://127.0.0.1:11434"})

	engine, err := api.New(api.WithUpstreamRouter(reg))
	if err != nil {
		panic(err)
	}
	fmt.Println(engine.Addr())
	// Output: :8080
}
