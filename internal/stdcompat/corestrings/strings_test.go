// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import coretest "dappco.re/go"

func TestStrings_Contains_Good(t *coretest.T) {
	got := Contains("api gateway", "gate")
	coretest.AssertTrue(t, got)
	coretest.AssertContains(t, "api gateway", "gate")
}

func TestStrings_Contains_Bad(t *coretest.T) {
	got := Contains("api gateway", "proxy")
	coretest.AssertFalse(t, got)
	coretest.AssertNotContains(t, "api gateway", "proxy")
}

func TestStrings_Contains_Ugly(t *coretest.T) {
	got := Contains("", "")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "", TrimSpace(""))
}

func TestStrings_HasPrefix_Good(t *coretest.T) {
	got := HasPrefix("api-gateway", "api")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "gateway", TrimPrefix("api-gateway", "api-"))
}

func TestStrings_HasPrefix_Bad(t *coretest.T) {
	got := HasPrefix("api-gateway", "gate")
	coretest.AssertFalse(t, got)
	coretest.AssertEqual(t, "api-gateway", TrimPrefix("api-gateway", "gate"))
}

func TestStrings_HasPrefix_Ugly(t *coretest.T) {
	got := HasPrefix("", "")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "", TrimPrefix("", ""))
}

func TestStrings_HasSuffix_Good(t *coretest.T) {
	got := HasSuffix("api-gateway", "way")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "api-gate", TrimSuffix("api-gateway", "way"))
}

func TestStrings_HasSuffix_Bad(t *coretest.T) {
	got := HasSuffix("api-gateway", "api")
	coretest.AssertFalse(t, got)
	coretest.AssertEqual(t, "api-gateway", TrimSuffix("api-gateway", "api"))
}

func TestStrings_HasSuffix_Ugly(t *coretest.T) {
	got := HasSuffix("", "")
	coretest.AssertTrue(t, got)
	coretest.AssertEqual(t, "", TrimSuffix("", ""))
}

func TestStrings_TrimSpace_Good(t *coretest.T) {
	got := TrimSpace("  api  ")
	coretest.AssertEqual(t, "api", got)
	coretest.AssertEqual(t, 3, len(got))
}

func TestStrings_TrimSpace_Bad(t *coretest.T) {
	got := TrimSpace("api")
	coretest.AssertEqual(t, "api", got)
	coretest.AssertEqual(t, "api", TrimSpace(got))
}

func TestStrings_TrimSpace_Ugly(t *coretest.T) {
	got := TrimSpace("\n\t api \r\n")
	coretest.AssertEqual(t, "api", got)
	coretest.AssertFalse(t, Contains(got, "\n"))
}

func TestStrings_TrimSuffix_Good(t *coretest.T) {
	got := TrimSuffix("token.json", ".json")
	coretest.AssertEqual(t, "token", got)
	coretest.AssertTrue(t, HasSuffix("token.json", ".json"))
}

func TestStrings_TrimSuffix_Bad(t *coretest.T) {
	got := TrimSuffix("token.json", ".yaml")
	coretest.AssertEqual(t, "token.json", got)
	coretest.AssertFalse(t, HasSuffix("token.json", ".yaml"))
}

func TestStrings_TrimSuffix_Ugly(t *coretest.T) {
	got := TrimSuffix("token", "")
	coretest.AssertEqual(t, "token", got)
	coretest.AssertTrue(t, HasSuffix("token", ""))
}

func TestStrings_TrimPrefix_Good(t *coretest.T) {
	got := TrimPrefix("Bearer token", "Bearer ")
	coretest.AssertEqual(t, "token", got)
	coretest.AssertTrue(t, HasPrefix("Bearer token", "Bearer "))
}

func TestStrings_TrimPrefix_Bad(t *coretest.T) {
	got := TrimPrefix("Basic token", "Bearer ")
	coretest.AssertEqual(t, "Basic token", got)
	coretest.AssertFalse(t, HasPrefix("Basic token", "Bearer "))
}

func TestStrings_TrimPrefix_Ugly(t *coretest.T) {
	got := TrimPrefix("token", "")
	coretest.AssertEqual(t, "token", got)
	coretest.AssertTrue(t, HasPrefix("token", ""))
}

func TestStrings_ToLower_Good(t *coretest.T) {
	got := ToLower("API")
	coretest.AssertEqual(t, "api", got)
	coretest.AssertEqual(t, "api", TrimSpace(got))
}

func TestStrings_ToLower_Bad(t *coretest.T) {
	got := ToLower("api")
	coretest.AssertEqual(t, "api", got)
	coretest.AssertFalse(t, Contains(got, "API"))
}

func TestStrings_ToLower_Ugly(t *coretest.T) {
	got := ToLower("")
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestStrings_NewReader_Good(t *coretest.T) {
	reader := NewReader("payload")
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "payload", r.Value)
}

func TestStrings_NewReader_Bad(t *coretest.T) {
	reader := NewReader("")
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "", r.Value)
}

func TestStrings_NewReader_Ugly(t *coretest.T) {
	reader := NewReader("line\n")
	r := coretest.ReadAll(reader)
	coretest.AssertTrue(t, r.OK)
	coretest.AssertEqual(t, "line\n", r.Value)
}

func TestStrings_Join_Good(t *coretest.T) {
	got := Join([]string{"api", "gateway"}, "/")
	coretest.AssertEqual(t, "api/gateway", got)
	coretest.AssertContains(t, got, "/")
}

func TestStrings_Join_Bad(t *coretest.T) {
	got := Join(nil, "/")
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestStrings_Join_Ugly(t *coretest.T) {
	got := Join([]string{"api", "", "gateway"}, "/")
	coretest.AssertEqual(t, "api//gateway", got)
	coretest.AssertContains(t, got, "//")
}

func TestStrings_Split_Good(t *coretest.T) {
	got := Split("api/gateway", "/")
	coretest.AssertEqual(t, []string{"api", "gateway"}, got)
	coretest.AssertEqual(t, 2, len(got))
}

func TestStrings_Split_Bad(t *coretest.T) {
	got := Split("api", "/")
	coretest.AssertEqual(t, []string{"api"}, got)
	coretest.AssertEqual(t, 1, len(got))
}

func TestStrings_Split_Ugly(t *coretest.T) {
	got := Split("", "/")
	coretest.AssertEqual(t, []string{""}, got)
	coretest.AssertEqual(t, 1, len(got))
}

func TestStrings_Repeat_Good(t *coretest.T) {
	got := Repeat("ab", 3)
	coretest.AssertEqual(t, "ababab", got)
	coretest.AssertEqual(t, 6, len(got))
}

func TestStrings_Repeat_Bad(t *coretest.T) {
	got := Repeat("ab", 0)
	coretest.AssertEqual(t, "", got)
	coretest.AssertEqual(t, 0, len(got))
}

func TestStrings_Repeat_Ugly(t *coretest.T) {
	got := Repeat("ab", -1)
	coretest.AssertEqual(t, "", got)
	coretest.AssertFalse(t, Contains(got, "ab"))
}

func TestStrings_CutPrefix_Good(t *coretest.T) {
	rest, ok := CutPrefix("Bearer token", "Bearer ")
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "token", rest)
}

func TestStrings_CutPrefix_Bad(t *coretest.T) {
	rest, ok := CutPrefix("Basic token", "Bearer ")
	coretest.AssertFalse(t, ok)
	coretest.AssertEqual(t, "Basic token", rest)
}

func TestStrings_CutPrefix_Ugly(t *coretest.T) {
	rest, ok := CutPrefix("", "")
	coretest.AssertTrue(t, ok)
	coretest.AssertEqual(t, "", rest)
}
