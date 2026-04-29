// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

func ExampleJoin() {
	coretest.Println(Join("api", "gateway"))
	// Output: api/gateway
}

func ExampleDir() {
	coretest.Println(Dir("api/gateway/config.json"))
	// Output: api/gateway
}

func ExampleClean() {
	coretest.Println(Clean("api/../gateway"))
	// Output: gateway
}

func ExampleIsAbs() {
	coretest.Println(IsAbs("/api"))
	// Output: true
}

func ExampleAbs() {
	got, err := Abs(".")
	coretest.Println(err == nil, IsAbs(got))
	// Output: true true
}

func ExampleRel() {
	got, err := Rel("/tmp", "/tmp/api")
	coretest.Println(err == nil, got)
	// Output: true api
}

func ExampleEvalSymlinks() {
	got, err := EvalSymlinks(".")
	coretest.Println(err == nil, got != "")
	// Output: true true
}
