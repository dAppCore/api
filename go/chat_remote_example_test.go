// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	"fmt"

	api "dappco.re/go/api"
)

func ExampleWithChatCompletionsRemote() {
	reg := api.NewUpstreamRegistry(api.AllowPrivateUpstreams("10.0.0.0/8"))
	_ = reg.Set("llama3:70b", api.Upstream{URL: "http://10.0.0.5:11434"})
	_ = reg.SetDefault(api.Upstream{URL: "https://llm.lthn.sh"}) // OpenAI-compatible — passthrough

	engine, err := api.New(
		api.WithChatCompletionsRemote(reg,
			api.WithChatModelAdapter("llama3:70b", api.OllamaAdapter()),
		),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(engine.Addr())
	// Output: :8080
}
