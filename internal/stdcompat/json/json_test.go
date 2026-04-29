// SPDX-License-Identifier: EUPL-1.2

package json

import (
	bytebuf "dappco.re/go/api/internal/stdcompat/bytes"

	coretest "dappco.re/go"
)

type jsonSample struct {
	Name string `json:"name"`
}

func TestJson_Marshal_Good(t *coretest.T) {
	data, err := Marshal(jsonSample{Name: "Ada"})
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, `{"name":"Ada"}`, string(data))
}

func TestJson_Marshal_Bad(t *coretest.T) {
	data, err := Marshal(func() {})
	coretest.AssertError(t, err)
	coretest.AssertNil(t, data)
}

func TestJson_Marshal_Ugly(t *coretest.T) {
	data, err := Marshal(nil)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "null", string(data))
}

func TestJson_Unmarshal_Good(t *coretest.T) {
	var out jsonSample
	err := Unmarshal([]byte(`{"name":"Ada"}`), &out)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "Ada", out.Name)
}

func TestJson_Unmarshal_Bad(t *coretest.T) {
	var out jsonSample
	err := Unmarshal([]byte(`{"name":`), &out)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", out.Name)
}

func TestJson_Unmarshal_Ugly(t *coretest.T) {
	var out []string
	err := Unmarshal([]byte(`[]`), &out)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 0, len(out))
}

func TestJson_NewEncoder_Good(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	coretest.AssertNotNil(t, encoder)
	coretest.AssertNoError(t, encoder.Encode(jsonSample{Name: "Ada"}))
}

func TestJson_NewEncoder_Bad(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	coretest.AssertNotNil(t, encoder)
	coretest.AssertNoError(t, encoder.Encode(func() {}))
}

func TestJson_NewEncoder_Ugly(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	coretest.AssertNotNil(t, encoder)
	coretest.AssertNoError(t, encoder.Encode(struct{}{}))
}

func TestJson_NewDecoder_Good(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`{"name":"Ada"}`))
	var out jsonSample
	coretest.AssertNotNil(t, decoder)
	coretest.AssertNoError(t, decoder.Decode(&out))
}

func TestJson_NewDecoder_Bad(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`{"name":`))
	var out jsonSample
	coretest.AssertNotNil(t, decoder)
	coretest.AssertError(t, decoder.Decode(&out))
}

func TestJson_NewDecoder_Ugly(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`[]`))
	var out []string
	coretest.AssertNotNil(t, decoder)
	coretest.AssertNoError(t, decoder.Decode(&out))
}

func TestJson_Encoder_Encode_Good(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	err := encoder.Encode(jsonSample{Name: "Ada"})
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "{\"name\":\"Ada\"}\n", buf.String())
}

func TestJson_Encoder_Encode_Bad(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	err := encoder.Encode(func() {})
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "{}\n", buf.String())
}

func TestJson_Encoder_Encode_Ugly(t *coretest.T) {
	buf := bytebuf.NewBuffer(nil)
	encoder := NewEncoder(buf)
	err := encoder.Encode(struct{}{})
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "{}\n", buf.String())
}

func TestJson_Decoder_Decode_Good(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`{"name":"Ada"}`))
	var out jsonSample
	err := decoder.Decode(&out)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "Ada", out.Name)
}

func TestJson_Decoder_Decode_Bad(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`{"name":`))
	var out jsonSample
	err := decoder.Decode(&out)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, "", out.Name)
}

func TestJson_Decoder_Decode_Ugly(t *coretest.T) {
	decoder := NewDecoder(coretest.NewReader(`null`))
	var out jsonSample
	err := decoder.Decode(&out)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, "", out.Name)
}
