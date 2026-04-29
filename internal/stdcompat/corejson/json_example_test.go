// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import (
	bytebuf "dappco.re/go/api/internal/stdcompat/corebytes"

	coretest "dappco.re/go"
)

func ExampleMarshal() {
	data, _ := Marshal(jsonSample{Name: "Ada"})
	coretest.Println(string(data))
	// Output: {"name":"Ada"}
}

func ExampleUnmarshal() {
	var out jsonSample
	_ = Unmarshal([]byte(`{"name":"Ada"}`), &out)
	coretest.Println(out.Name)
	// Output: Ada
}

func ExampleNewEncoder() {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	coretest.Println(encoder != nil)
	// Output: true
}

func ExampleNewDecoder() {
	decoder := NewDecoder(coretest.NewReader(`{"name":"Ada"}`))
	coretest.Println(decoder != nil)
	// Output: true
}

func ExampleEncoder_Encode() {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	_ = encoder.Encode(jsonSample{Name: "Ada"})
	coretest.Println(buf.String() == "{\"name\":\"Ada\"}\n")
	// Output: true
}

func ExampleDecoder_Decode() {
	decoder := NewDecoder(coretest.NewReader(`{"name":"Ada"}`))
	var out jsonSample
	_ = decoder.Decode(&out)
	coretest.Println(out.Name)
	// Output: Ada
}
