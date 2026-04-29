// SPDX-License-Identifier: EUPL-1.2

package fmt

import coretest "dappco.re/go"

func ExampleSprint() {
	coretest.Println(Sprint("api", " ", "gateway"))
	// Output: api gateway
}

func ExampleSprintf() {
	coretest.Println(Sprintf("%s:%d", "api", 1))
	// Output: api:1
}

func ExampleErrorf() {
	err := Errorf("api %s", "failed")
	coretest.Println(err.Error())
	// Output: api failed
}

func ExamplePrintf() {
	n, _ := Printf("api\n")
	coretest.Println(n)
	// Output:
	// api
	// 4
}
