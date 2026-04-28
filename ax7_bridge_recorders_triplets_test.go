// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"io"
	"net/http"
	"net/http/httptest"

	coretest "dappco.re/go"
	"github.com/andybalholm/brotli"

	"github.com/gin-gonic/gin"
)

func ax7ToolSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []any{"name"},
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
}

func ax7BrotliWriter() *brotliWriter {
	_, rec := ax7GinContext()
	ctx, _ := gin.CreateTestContext(rec)
	writer := brotli.NewWriter(io.Discard)
	return &brotliWriter{ResponseWriter: ctx.Writer, writer: writer}
}

func ax7GinWriter() gin.ResponseWriter {
	ctx, _ := ax7GinContext()
	return ctx.Writer
}

func TestAX7_Handler_Handle_Good(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Request.Header.Set("Accept-Encoding", "br")
	handler := newBrotliHandler(BrotliDefaultCompression)
	handler.Handle(ctx)
	coretest.AssertEqual(t, "br", rec.Header().Get("Content-Encoding"))
}

func TestAX7_Handler_Handle_Bad(t *coretest.T) {
	ctx, rec := ax7GinContext()
	handler := newBrotliHandler(BrotliDefaultCompression)
	handler.Handle(ctx)
	coretest.AssertEqual(t, "", rec.Header().Get("Content-Encoding"))
}

func TestAX7_Handler_Handle_Ugly(t *coretest.T) {
	ctx, rec := ax7GinContext()
	ctx.Request.Header.Set("Accept-Encoding", "gzip, br;q=0")
	handler := newBrotliHandler(BrotliBestCompression + 100)
	handler.Handle(ctx)
	coretest.AssertEqual(t, "", rec.Header().Get("Content-Encoding"))
}

func TestAX7_Writer_Write_Good(t *coretest.T) {
	writer := ax7BrotliWriter()
	n, err := writer.Write([]byte("payload"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_Writer_Write_Bad(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.released = true
	n, err := writer.Write([]byte("payload"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_Writer_Write_Ugly(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.status = http.StatusInternalServerError
	n, err := writer.Write([]byte("error"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("error"), n)
}

func TestAX7_Writer_WriteString_Good(t *coretest.T) {
	writer := ax7BrotliWriter()
	n, err := writer.WriteString("payload")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_Writer_WriteString_Bad(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.released = true
	n, err := writer.WriteString("payload")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_Writer_WriteString_Ugly(t *coretest.T) {
	writer := ax7BrotliWriter()
	n, err := writer.WriteString("")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestAX7_Writer_WriteHeader_Good(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.WriteHeader(http.StatusCreated)
	coretest.AssertEqual(t, http.StatusCreated, writer.status)
	coretest.AssertTrue(t, writer.statusWritten)
}

func TestAX7_Writer_WriteHeader_Bad(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.released = true
	writer.WriteHeader(http.StatusCreated)
	coretest.AssertFalse(t, writer.statusWritten)
}

func TestAX7_Writer_WriteHeader_Ugly(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.Header().Set("Content-Encoding", "br")
	writer.WriteHeader(http.StatusInternalServerError)
	coretest.AssertEqual(t, "", writer.Header().Get("Content-Encoding"))
}

func TestAX7_Writer_WriteHeaderNow_Good(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.WriteHeaderNow()
	coretest.AssertTrue(t, writer.statusWritten)
	coretest.AssertEqual(t, http.StatusOK, writer.status)
}

func TestAX7_Writer_WriteHeaderNow_Bad(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.released = true
	writer.WriteHeaderNow()
	coretest.AssertFalse(t, writer.statusWritten)
}

func TestAX7_Writer_WriteHeaderNow_Ugly(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.status = http.StatusBadRequest
	writer.WriteHeaderNow()
	coretest.AssertTrue(t, writer.statusWritten)
}

func TestAX7_Writer_Flush_Good(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.Flush()
	coretest.AssertFalse(t, writer.released)
	coretest.AssertNotNil(t, writer.writer)
}

func TestAX7_Writer_Flush_Bad(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.released = true
	writer.Flush()
	coretest.AssertTrue(t, writer.released)
}

func TestAX7_Writer_Flush_Ugly(t *coretest.T) {
	writer := ax7BrotliWriter()
	writer.writer.Close()
	writer.Flush()
	coretest.AssertNotNil(t, writer.writer)
}

func TestAX7_NewToolBridge_Good(t *coretest.T) {
	bridge := NewToolBridge("tools")
	coretest.AssertEqual(t, "/tools", bridge.BasePath())
	coretest.AssertEqual(t, "tools", bridge.Name())
}

func TestAX7_NewToolBridge_Bad(t *coretest.T) {
	bridge := NewToolBridge("")
	coretest.AssertEqual(t, "/", bridge.BasePath())
	coretest.AssertEqual(t, "tools", bridge.Name())
}

func TestAX7_NewToolBridge_Ugly(t *coretest.T) {
	bridge := NewToolBridge("///mcp///")
	coretest.AssertEqual(t, "/mcp", bridge.BasePath())
	coretest.AssertEmpty(t, bridge.Tools())
}

func TestAX7_ToolBridge_Add_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping", Description: "Ping"}, func(c *gin.Context) { c.JSON(http.StatusOK, OK("pong")) })
	coretest.AssertLen(t, bridge.Tools(), 1)
	coretest.AssertEqual(t, "ping", bridge.Tools()[0].Name)
}

func TestAX7_ToolBridge_Add_Bad(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	coretest.AssertPanics(t, func() {
		bridge.Add(ToolDescriptor{Name: "bad name"}, func(*gin.Context) {})
	})
	coretest.AssertEmpty(t, bridge.Tools())
}

func TestAX7_ToolBridge_Add_Ugly(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping", InputSchema: ax7ToolSchema()}, func(c *gin.Context) { c.JSON(http.StatusOK, OK("pong")) })
	coretest.AssertLen(t, bridge.Tools(), 1)
	coretest.AssertNotNil(t, bridge.tools[0].handler)
}

func TestAX7_ToolBridge_Name_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	name := bridge.Name()
	coretest.AssertEqual(t, "tools", name)
	coretest.AssertNotEmpty(t, name)
}

func TestAX7_ToolBridge_Name_Bad(t *coretest.T) {
	bridge := &ToolBridge{}
	name := bridge.Name()
	coretest.AssertEqual(t, "", name)
	coretest.AssertEmpty(t, name)
}

func TestAX7_ToolBridge_Name_Ugly(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.name = "custom"
	coretest.AssertEqual(t, "custom", bridge.Name())
}

func TestAX7_ToolBridge_BasePath_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	path := bridge.BasePath()
	coretest.AssertEqual(t, "/tools", path)
	coretest.AssertNotEmpty(t, path)
}

func TestAX7_ToolBridge_BasePath_Bad(t *coretest.T) {
	bridge := &ToolBridge{}
	path := bridge.BasePath()
	coretest.AssertEqual(t, "", path)
	coretest.AssertEmpty(t, path)
}

func TestAX7_ToolBridge_BasePath_Ugly(t *coretest.T) {
	bridge := NewToolBridge("///")
	path := bridge.BasePath()
	coretest.AssertEqual(t, "/", path)
	coretest.AssertNotEmpty(t, path)
}

func TestAX7_ToolBridge_RegisterRoutes_Good(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping"}, func(c *gin.Context) { c.JSON(http.StatusAccepted, OK("pong")) })
	router := gin.New()
	group := router.Group("/tools")
	bridge.RegisterRoutes(group)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tools/ping", nil))
	coretest.AssertEqual(t, http.StatusAccepted, rec.Code)
}

func TestAX7_ToolBridge_RegisterRoutes_Bad(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	bridge := NewToolBridge("/tools")
	router := gin.New()
	group := router.Group("/tools")
	bridge.RegisterRoutes(group)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tools/missing", nil))
	coretest.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestAX7_ToolBridge_RegisterRoutes_Ugly(t *coretest.T) {
	gin.SetMode(gin.TestMode)
	bridge := NewToolBridge("/tools")
	router := gin.New()
	group := router.Group("/tools")
	bridge.RegisterRoutes(group)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tools/", nil))
	coretest.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestAX7_ToolBridge_Describe_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping", Description: "Ping"}, func(*gin.Context) {})
	descs := bridge.Describe()
	coretest.AssertLen(t, descs, 2)
	coretest.AssertEqual(t, "/ping", descs[1].Path)
}

func TestAX7_ToolBridge_Describe_Bad(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	descs := bridge.Describe()
	coretest.AssertLen(t, descs, 1)
	coretest.AssertEqual(t, "/", descs[0].Path)
}

func TestAX7_ToolBridge_Describe_Ugly(t *coretest.T) {
	bridge := &ToolBridge{name: ""}
	descs := bridge.Describe()
	coretest.AssertLen(t, descs, 1)
	coretest.AssertEqual(t, []string{"tools"}, descs[0].Tags)
}

func TestAX7_ToolBridge_DescribeIter_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping"}, func(*gin.Context) {})
	count := 0
	for range bridge.DescribeIter() {
		count++
	}
	coretest.AssertEqual(t, 2, count)
}

func TestAX7_ToolBridge_DescribeIter_Bad(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	count := 0
	for range bridge.DescribeIter() {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_ToolBridge_DescribeIter_Ugly(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	iter := bridge.DescribeIter()
	bridge.Add(ToolDescriptor{Name: "later"}, func(*gin.Context) {})
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 1, count)
}

func TestAX7_ToolBridge_Tools_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping"}, func(*gin.Context) {})
	tools := bridge.Tools()
	coretest.AssertLen(t, tools, 1)
	coretest.AssertEqual(t, "ping", tools[0].Name)
}

func TestAX7_ToolBridge_Tools_Bad(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	tools := bridge.Tools()
	coretest.AssertEmpty(t, tools)
	coretest.AssertLen(t, tools, 0)
}

func TestAX7_ToolBridge_Tools_Ugly(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping"}, func(*gin.Context) {})
	tools := bridge.Tools()
	tools[0].Name = "mutated"
	coretest.AssertEqual(t, "ping", bridge.Tools()[0].Name)
}

func TestAX7_ToolBridge_ToolsIter_Good(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	bridge.Add(ToolDescriptor{Name: "ping"}, func(*gin.Context) {})
	var tools []ToolDescriptor
	for tool := range bridge.ToolsIter() {
		tools = append(tools, tool)
	}
	coretest.AssertLen(t, tools, 1)
}

func TestAX7_ToolBridge_ToolsIter_Bad(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	var tools []ToolDescriptor
	for tool := range bridge.ToolsIter() {
		tools = append(tools, tool)
	}
	coretest.AssertEmpty(t, tools)
}

func TestAX7_ToolBridge_ToolsIter_Ugly(t *coretest.T) {
	bridge := NewToolBridge("/tools")
	iter := bridge.ToolsIter()
	bridge.Add(ToolDescriptor{Name: "later"}, func(*gin.Context) {})
	count := 0
	for range iter {
		count++
	}
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_IsValidMCPServerID_Good(t *coretest.T) {
	ok := IsValidMCPServerID("server-1")
	coretest.AssertTrue(t, ok)
	coretest.AssertTrue(t, IsValidMCPServerID("Server2"))
}

func TestAX7_IsValidMCPServerID_Bad(t *coretest.T) {
	ok := IsValidMCPServerID("bad/server")
	coretest.AssertFalse(t, ok)
	coretest.AssertFalse(t, IsValidMCPServerID("../bad"))
}

func TestAX7_IsValidMCPServerID_Ugly(t *coretest.T) {
	ok := IsValidMCPServerID("")
	coretest.AssertFalse(t, ok)
	coretest.AssertFalse(t, IsValidMCPServerID("server\x00id"))
}

func TestAX7_InputValidator_Validate_Good(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.Validate([]byte(`{"name":"Ada"}`))
	coretest.AssertNoError(t, err)
}

func TestAX7_InputValidator_Validate_Bad(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.Validate([]byte(`{}`))
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "name")
}

func TestAX7_InputValidator_Validate_Ugly(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.Validate(nil)
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "required")
}

func TestAX7_InputValidator_ValidateResponse_Good(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.ValidateResponse([]byte(`{"success":true,"data":{"name":"Ada"}}`))
	coretest.AssertNoError(t, err)
}

func TestAX7_InputValidator_ValidateResponse_Bad(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.ValidateResponse([]byte(`{"success":false}`))
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "successful")
}

func TestAX7_InputValidator_ValidateResponse_Ugly(t *coretest.T) {
	validator := newToolInputValidator(ax7ToolSchema())
	err := validator.ValidateResponse([]byte(`{"success":true}`))
	coretest.AssertNoError(t, err)
}

func TestAX7_ResponseRecorder_Header_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	header := w.Header()
	header.Set("X-Test", "yes")
	coretest.AssertEqual(t, "yes", w.Header().Get("X-Test"))
}

func TestAX7_ResponseRecorder_Header_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	header := w.Header()
	coretest.AssertNotNil(t, header)
	coretest.AssertEqual(t, "", header.Get("Missing"))
}

func TestAX7_ResponseRecorder_Header_Ugly(t *coretest.T) {
	base := ax7GinWriter()
	base.Header().Set("X-Original", "yes")
	w := newToolResponseRecorder(base)
	coretest.AssertEqual(t, "yes", w.Header().Get("X-Original"))
}

func TestAX7_ResponseRecorder_WriteHeader_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusCreated)
	coretest.AssertEqual(t, http.StatusCreated, w.Status())
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_ResponseRecorder_WriteHeader_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
	coretest.AssertFalse(t, w.Written())
}

func TestAX7_ResponseRecorder_WriteHeader_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(0)
	coretest.AssertEqual(t, 0, w.Status())
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_ResponseRecorder_WriteHeaderNow_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	coretest.AssertTrue(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_ResponseRecorder_WriteHeaderNow_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	coretest.AssertFalse(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_ResponseRecorder_WriteHeaderNow_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusAccepted)
	w.WriteHeaderNow()
	coretest.AssertEqual(t, http.StatusAccepted, w.Status())
}

func TestAX7_ResponseRecorder_Write_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	n, err := w.Write([]byte("payload"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_ResponseRecorder_Write_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	n, err := w.Write(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestAX7_ResponseRecorder_Write_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusAccepted)
	n, err := w.Write([]byte("ok"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, http.StatusAccepted, w.Status())
	coretest.AssertEqual(t, 2, n)
}

func TestAX7_ResponseRecorder_WriteString_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	n, err := w.WriteString("payload")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_ResponseRecorder_WriteString_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	n, err := w.WriteString("")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestAX7_ResponseRecorder_WriteString_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusAccepted)
	n, err := w.WriteString("ok")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, http.StatusAccepted, w.Status())
	coretest.AssertEqual(t, 2, n)
}

func TestAX7_ResponseRecorder_Flush_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.Flush()
	coretest.AssertFalse(t, w.Written())
}

func TestAX7_ResponseRecorder_Flush_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	w.Flush()
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_ResponseRecorder_Flush_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteString("data")
	w.Flush()
	coretest.AssertEqual(t, 4, w.Size())
}

func TestAX7_ResponseRecorder_Status_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusAccepted)
	coretest.AssertEqual(t, http.StatusAccepted, w.Status())
}

func TestAX7_ResponseRecorder_Status_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	status := w.Status()
	coretest.AssertEqual(t, http.StatusOK, status)
}

func TestAX7_ResponseRecorder_Status_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeader(0)
	coretest.AssertEqual(t, 0, w.Status())
}

func TestAX7_ResponseRecorder_Size_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 4, w.Size())
}

func TestAX7_ResponseRecorder_Size_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	size := w.Size()
	coretest.AssertEqual(t, 0, size)
}

func TestAX7_ResponseRecorder_Size_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	_, err := w.Write(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, w.Size())
}

func TestAX7_ResponseRecorder_Written_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_ResponseRecorder_Written_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	coretest.AssertFalse(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_ResponseRecorder_Written_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_ResponseRecorder_Hijack_Good(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	conn, rw, err := w.Hijack()
	coretest.AssertError(t, err)
	coretest.AssertNil(t, conn)
	coretest.AssertNil(t, rw)
}

func TestAX7_ResponseRecorder_Hijack_Bad(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	_, _, err := w.Hijack()
	coretest.AssertError(t, err)
	coretest.AssertContains(t, err.Error(), "hijacking")
}

func TestAX7_ResponseRecorder_Hijack_Ugly(t *coretest.T) {
	w := newToolResponseRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	_, _, err := w.Hijack()
	coretest.AssertError(t, err)
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_Header_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.Header().Set("X-Test", "yes")
	coretest.AssertEqual(t, "yes", w.Header().Get("X-Test"))
}

func TestAX7_MetaRecorder_Header_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	header := w.Header()
	coretest.AssertNotNil(t, header)
	coretest.AssertEqual(t, "", header.Get("Missing"))
}

func TestAX7_MetaRecorder_Header_Ugly(t *coretest.T) {
	base := ax7GinWriter()
	base.Header().Set("X-Original", "yes")
	w := newResponseMetaRecorder(base)
	coretest.AssertEqual(t, "yes", w.Header().Get("X-Original"))
}

func TestAX7_MetaRecorder_WriteHeader_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusCreated)
	coretest.AssertEqual(t, http.StatusCreated, w.Status())
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_WriteHeader_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
	coretest.AssertFalse(t, w.Written())
}

func TestAX7_MetaRecorder_WriteHeader_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	w.WriteHeader(http.StatusAccepted)
	coretest.AssertEqual(t, http.StatusAccepted, w.Status())
}

func TestAX7_MetaRecorder_WriteHeaderNow_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	coretest.AssertTrue(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_MetaRecorder_WriteHeaderNow_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertFalse(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_MetaRecorder_WriteHeaderNow_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	w.WriteHeaderNow()
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_Write_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	n, err := w.Write([]byte("payload"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_MetaRecorder_Write_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	n, err := w.Write(nil)
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestAX7_MetaRecorder_Write_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	n, err := w.Write([]byte("pass"))
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 4, n)
}

func TestAX7_MetaRecorder_WriteString_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	n, err := w.WriteString("payload")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, len("payload"), n)
}

func TestAX7_MetaRecorder_WriteString_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	n, err := w.WriteString("")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 0, n)
}

func TestAX7_MetaRecorder_WriteString_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	n, err := w.WriteString("pass")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 4, n)
}

func TestAX7_MetaRecorder_Flush_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.Flush()
	coretest.AssertTrue(t, w.passthrough)
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_Flush_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	w.Flush()
	coretest.AssertTrue(t, w.passthrough)
}

func TestAX7_MetaRecorder_Flush_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	w.Flush()
	coretest.AssertTrue(t, w.passthrough)
}

func TestAX7_MetaRecorder_Status_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.WriteHeader(http.StatusCreated)
	coretest.AssertEqual(t, http.StatusCreated, w.Status())
}

func TestAX7_MetaRecorder_Status_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
	coretest.AssertFalse(t, w.Written())
}

func TestAX7_MetaRecorder_Status_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.WriteHeader(0)
	coretest.AssertEqual(t, 0, w.Status())
}

func TestAX7_MetaRecorder_Size_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	coretest.AssertEqual(t, 4, w.Size())
}

func TestAX7_MetaRecorder_Size_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertEqual(t, 0, w.Size())
	coretest.AssertFalse(t, w.Written())
}

func TestAX7_MetaRecorder_Size_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	coretest.AssertGreaterOrEqual(t, w.Size(), 0)
}

func TestAX7_MetaRecorder_Written_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.WriteHeaderNow()
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_Written_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertFalse(t, w.Written())
	coretest.AssertEqual(t, http.StatusOK, w.Status())
}

func TestAX7_MetaRecorder_Written_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	_, err := w.WriteString("data")
	coretest.RequireNoError(t, err)
	coretest.AssertTrue(t, w.Written())
}

func TestAX7_MetaRecorder_Hijack_Good(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertPanics(t, func() {
		_, _, _ = w.Hijack()
	})
	coretest.AssertTrue(t, w.passthrough)
}

func TestAX7_MetaRecorder_Hijack_Bad(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	coretest.AssertPanics(t, func() {
		_, _, _ = w.Hijack()
	})
	coretest.AssertTrue(t, w.passthrough)
}

func TestAX7_MetaRecorder_Hijack_Ugly(t *coretest.T) {
	w := newResponseMetaRecorder(ax7GinWriter())
	w.passthrough = true
	coretest.AssertPanics(t, func() {
		_, _, _ = w.Hijack()
	})
	coretest.AssertTrue(t, w.passthrough)
}

func TestAX7_SSEBroker_ClientCount_Good(t *coretest.T) {
	broker := NewSSEBroker()
	count := broker.ClientCount()
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_SSEBroker_ClientCount_Bad(t *coretest.T) {
	broker := NewSSEBroker()
	count := broker.ClientCount()
	coretest.AssertEqual(t, 0, count)
}

func TestAX7_SSEBroker_ClientCount_Ugly(t *coretest.T) {
	broker := NewSSEBroker()
	broker.Drain()
	coretest.AssertEqual(t, 0, broker.ClientCount())
}
