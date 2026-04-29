// SPDX-License-Identifier: EUPL-1.2

package strings

import coretest "dappco.re/go"

func ExampleContains() {
	coretest.Println(Contains("api gateway", "gate"))
	// Output: true
}

func ExampleHasPrefix() {
	coretest.Println(HasPrefix("api-gateway", "api"))
	// Output: true
}

func ExampleHasSuffix() {
	coretest.Println(HasSuffix("api-gateway", "way"))
	// Output: true
}

func ExampleTrimSpace() {
	coretest.Println(TrimSpace("  api  "))
	// Output: api
}

func ExampleTrimSuffix() {
	coretest.Println(TrimSuffix("token.json", ".json"))
	// Output: token
}

func ExampleTrimPrefix() {
	coretest.Println(TrimPrefix("Bearer token", "Bearer "))
	// Output: token
}

func ExampleToLower() {
	coretest.Println(ToLower("API"))
	// Output: api
}

func ExampleNewReader() {
	r := coretest.ReadAll(NewReader("payload"))
	coretest.Println(r.Value)
	// Output: payload
}

func ExampleJoin() {
	coretest.Println(Join([]string{"api", "gateway"}, "/"))
	// Output: api/gateway
}

func ExampleSplit() {
	coretest.Println(Split("api/gateway", "/")[0])
	// Output: api
}

func ExampleRepeat() {
	coretest.Println(Repeat("ab", 2))
	// Output: abab
}

func ExampleCutPrefix() {
	rest, ok := CutPrefix("Bearer token", "Bearer ")
	coretest.Println(ok, rest)
	// Output: true token
}
