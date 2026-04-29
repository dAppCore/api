// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

func ExampleNewBufferString() {
	coretest.Println(NewBufferString("payload").String())
	// Output: payload
}

func ExampleNewBuffer() {
	coretest.Println(NewBuffer([]byte("payload")).String())
	// Output: payload
}

func ExampleNewReader() {
	r := coretest.ReadAll(NewReader([]byte("payload")))
	coretest.Println(r.Value)
	// Output: payload
}

func ExampleBuffer_Write() {
	buf := NewBufferString("a")
	n, _ := buf.Write([]byte("b"))
	coretest.Println(n, buf.String())
	// Output: 1 ab
}

func ExampleBuffer_WriteString() {
	buf := NewBufferString("a")
	n, _ := buf.WriteString("bc")
	coretest.Println(n, buf.String())
	// Output: 2 abc
}

func ExampleBuffer_Read() {
	buf := NewBufferString("abc")
	dst := make([]byte, 2)
	n, _ := buf.Read(dst)
	coretest.Println(n, string(dst))
	// Output: 2 ab
}

func ExampleBuffer_String() {
	coretest.Println(NewBufferString("payload").String())
	// Output: payload
}

func ExampleBuffer_Bytes() {
	coretest.Println(string(NewBufferString("payload").Bytes()))
	// Output: payload
}

func ExampleBuffer_Len() {
	coretest.Println(NewBufferString("payload").Len())
	// Output: 7
}

func ExampleBuffer_Reset() {
	buf := NewBufferString("payload")
	buf.Reset()
	coretest.Println(buf.Len())
	// Output: 0
}

func ExampleEqual() {
	coretest.Println(Equal([]byte("api"), []byte("api")))
	// Output: true
}

func ExampleContains() {
	coretest.Println(Contains([]byte("api gateway"), []byte("gate")))
	// Output: true
}

func ExampleRepeat() {
	coretest.Println(string(Repeat([]byte("ab"), 2)))
	// Output: abab
}

func ExampleTrimSpace() {
	coretest.Println(string(TrimSpace([]byte("  api  "))))
	// Output: api
}
