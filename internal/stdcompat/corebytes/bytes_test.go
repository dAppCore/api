// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

func TestBytes_NewBufferString_Good(t *coretest.T) {
	buf := NewBufferString("payload")
	coretest.AssertEqual(t, "payload", buf.String())
	coretest.AssertEqual(t, 7, buf.Len())
}

func TestBytes_NewBufferString_Bad(t *coretest.T) {
	buf := NewBufferString("")
	coretest.AssertEqual(t, "", buf.String())
	coretest.AssertEqual(t, 0, buf.Len())
}

func TestBytes_NewBufferString_Ugly(t *coretest.T) {
	buf := NewBufferString("line\n")
	coretest.AssertEqual(t, "line\n", buf.String())
	coretest.AssertTrue(t, Contains(buf.Bytes(), []byte("\n")))
}

func TestBytes_NewBuffer_Good(t *coretest.T) {
	buf := NewBuffer([]byte("payload"))
	coretest.AssertEqual(t, "payload", buf.String())
	coretest.AssertTrue(t, Equal([]byte("payload"), buf.Bytes()))
}

func TestBytes_NewBuffer_Bad(t *coretest.T) {
	source := []byte("payload")
	buf := NewBuffer(source)
	source[0] = 'P'
	coretest.AssertEqual(t, "payload", buf.String())
	coretest.AssertFalse(t, Equal(source, buf.Bytes()))
}

func TestBytes_NewBuffer_Ugly(t *coretest.T) {
	buf := NewBuffer(nil)
	coretest.AssertEqual(t, "", buf.String())
	coretest.AssertEqual(t, 0, buf.Len())
}

func TestBytes_NewReader_Good(t *coretest.T) {
	reader := NewReader([]byte("payload"))
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "payload", r.Value)
}

func TestBytes_NewReader_Bad(t *coretest.T) {
	reader := NewReader(nil)
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "", r.Value)
}

func TestBytes_NewReader_Ugly(t *coretest.T) {
	reader := NewReader([]byte("line\n"))
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "line\n", r.Value)
}

func TestBytes_Buffer_Write_Good(t *coretest.T) {
	buf := NewBufferString("a")
	n, err := buf.Write([]byte("b"))
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 1, n)
	coretest.AssertEqual(t, "ab", buf.String())
}

func TestBytes_Buffer_Write_Bad(t *coretest.T) {
	buf := NewBufferString("a")
	n, err := buf.Write(nil)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 0, n)
	coretest.AssertEqual(t, "a", buf.String())
}

func TestBytes_Buffer_Write_Ugly(t *coretest.T) {
	buf := NewBuffer(nil)
	n, err := buf.Write([]byte{0})
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 1, n)
	coretest.AssertTrue(t, Equal([]byte{0}, buf.Bytes()))
}

func TestBytes_Buffer_WriteString_Good(t *coretest.T) {
	buf := NewBufferString("a")
	n, err := buf.WriteString("bc")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 2, n)
	coretest.AssertEqual(t, "abc", buf.String())
}

func TestBytes_Buffer_WriteString_Bad(t *coretest.T) {
	buf := NewBufferString("a")
	n, err := buf.WriteString("")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 0, n)
	coretest.AssertEqual(t, "a", buf.String())
}

func TestBytes_Buffer_WriteString_Ugly(t *coretest.T) {
	buf := NewBuffer(nil)
	n, err := buf.WriteString("\n")
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 1, n)
	coretest.AssertEqual(t, "\n", buf.String())
}

func TestBytes_Buffer_Read_Good(t *coretest.T) {
	buf := NewBufferString("abc")
	dst := make([]byte, 2)
	n, err := buf.Read(dst)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 2, n)
	coretest.AssertEqual(t, "ab", string(dst))
}

func TestBytes_Buffer_Read_Bad(t *coretest.T) {
	buf := NewBufferString("")
	dst := make([]byte, 1)
	n, err := buf.Read(dst)
	coretest.AssertError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestBytes_Buffer_Read_Ugly(t *coretest.T) {
	buf := NewBufferString("abc")
	dst := make([]byte, 8)
	n, err := buf.Read(dst)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 3, n)
	coretest.AssertEqual(t, "abc", string(dst[:n]))
}

func TestBytes_Buffer_String_Good(t *coretest.T) {
	buf := NewBufferString("payload")
	got := buf.String()
	coretest.AssertEqual(t, "payload", got)
	coretest.AssertEqual(t, 7, len(got))
}

func TestBytes_Buffer_String_Bad(t *coretest.T) {
	buf := NewBuffer(nil)
	got := buf.String()
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestBytes_Buffer_String_Ugly(t *coretest.T) {
	buf := NewBuffer([]byte{0, 'a'})
	got := buf.String()
	coretest.AssertEqual(t, string([]byte{0, 'a'}), got)
	coretest.AssertTrue(t, Contains([]byte(got), []byte{'a'}))
}

func TestBytes_Buffer_Bytes_Good(t *coretest.T) {
	buf := NewBufferString("payload")
	got := buf.Bytes()
	coretest.AssertTrue(t, Equal([]byte("payload"), got))
	coretest.AssertEqual(t, 7, len(got))
}

func TestBytes_Buffer_Bytes_Bad(t *coretest.T) {
	buf := NewBufferString("payload")
	got := buf.Bytes()
	got[0] = 'P'
	coretest.AssertEqual(t, "payload", buf.String())
	coretest.AssertFalse(t, Equal(got, buf.Bytes()))
}

func TestBytes_Buffer_Bytes_Ugly(t *coretest.T) {
	buf := NewBuffer(nil)
	got := buf.Bytes()
	coretest.AssertNil(t, got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestBytes_Buffer_Len_Good(t *coretest.T) {
	buf := NewBufferString("abc")
	got := buf.Len()
	coretest.AssertEqual(t, 3, got)
	coretest.AssertEqual(t, "abc", buf.String())
}

func TestBytes_Buffer_Len_Bad(t *coretest.T) {
	buf := NewBufferString("abc")
	dst := make([]byte, 1)
	_, err := buf.Read(dst)
	coretest.AssertNoError(t, err)
	coretest.AssertEqual(t, 2, buf.Len())
}

func TestBytes_Buffer_Len_Ugly(t *coretest.T) {
	buf := NewBuffer(nil)
	got := buf.Len()
	coretest.AssertEqual(t, 0, got)
	coretest.AssertEqual(t, "", buf.String())
}

func TestBytes_Buffer_Reset_Good(t *coretest.T) {
	buf := NewBufferString("payload")
	buf.Reset()
	coretest.AssertEqual(t, "", buf.String())
	coretest.AssertEqual(t, 0, buf.Len())
}

func TestBytes_Buffer_Reset_Bad(t *coretest.T) {
	buf := NewBuffer(nil)
	buf.Reset()
	coretest.AssertEqual(t, "", buf.String())
	coretest.AssertEqual(t, 0, buf.Len())
}

func TestBytes_Buffer_Reset_Ugly(t *coretest.T) {
	buf := NewBufferString("payload")
	dst := make([]byte, 3)
	_, err := buf.Read(dst)
	coretest.AssertNoError(t, err)
	buf.Reset()
	coretest.AssertEqual(t, 0, buf.Len())
}

func TestBytes_Equal_Good(t *coretest.T) {
	got := Equal([]byte("api"), []byte("api"))
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, 3, len([]byte("api")))
}

func TestBytes_Equal_Bad(t *coretest.T) {
	got := Equal([]byte("api"), []byte("API"))
	coretest.AssertFalse(t, got)
	coretest.AssertNotEqual(t, []byte("api"), []byte("API"))
}

func TestBytes_Equal_Ugly(t *coretest.T) {
	got := Equal(nil, nil)
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, 0, len([]byte(nil)))
}

func TestBytes_Contains_Good(t *coretest.T) {
	got := Contains([]byte("api gateway"), []byte("gate"))
	coretest.AssertTrue(t, got)
	coretest.AssertContains(t, "api gateway", "gate")
}

func TestBytes_Contains_Bad(t *coretest.T) {
	got := Contains([]byte("api gateway"), []byte("proxy"))
	coretest.AssertFalse(t, got)
	coretest.AssertNotContains(t, "api gateway", "proxy")
}

func TestBytes_Contains_Ugly(t *coretest.T) {
	got := Contains(nil, nil)
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, 0, len([]byte(nil)))
}

func TestBytes_Repeat_Good(t *coretest.T) {
	got := Repeat([]byte("ab"), 2)
	coretest.AssertTrue(t, Equal([]byte("abab"), got))
	coretest.AssertEqual(t, 4, len(got))
}

func TestBytes_Repeat_Bad(t *coretest.T) {
	got := Repeat([]byte("ab"), 0)
	coretest.AssertNil(t, got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestBytes_Repeat_Ugly(t *coretest.T) {
	got := Repeat([]byte("ab"), -1)
	coretest.AssertNil(t, got)
	coretest.AssertFalse(t, Contains(got, []byte("ab")))
}

func TestBytes_TrimSpace_Good(t *coretest.T) {
	got := TrimSpace([]byte("  api  "))
	coretest.AssertTrue(t, Equal([]byte("api"), got))
	coretest.AssertEqual(t, 3, len(got))
}

func TestBytes_TrimSpace_Bad(t *coretest.T) {
	got := TrimSpace([]byte("api"))
	coretest.AssertTrue(t, Equal([]byte("api"), got))
	coretest.AssertEqual(t, 3, len(got))
}

func TestBytes_TrimSpace_Ugly(t *coretest.T) {
	got := TrimSpace([]byte("\n\t api \r\n"))
	coretest.AssertTrue(t, Equal([]byte("api"), got))
	coretest.AssertFalse(t, Contains(got, []byte("\n")))
}
