// SPDX-License-Identifier: EUPL-1.2

package stream

import (
	coretest "dappco.re/go"

	"github.com/gin-gonic/gin"
)

func ax7StreamHandler(*gin.Context) {}

type ax7Registrar struct {
	method string
	path   string
	count  int
}

func (r *ax7Registrar) Handle(method, path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r.method = method
	r.path = path
	r.count += len(handlers)
	return nil
}

func TestAX7_NewGroup_Good(t *coretest.T) {
	group := NewGroup(" events ", SSE("updates", ax7StreamHandler))
	handlers := group.Handlers()
	coretest.AssertEqual(t, "events", group.Name())
	coretest.AssertLen(t, handlers, 1)
}

func TestAX7_NewGroup_Bad(t *coretest.T) {
	group := NewGroup("", Handler{})
	handlers := group.Handlers()
	coretest.AssertEqual(t, "", group.Name())
	coretest.AssertNil(t, handlers)
}

func TestAX7_NewGroup_Ugly(t *coretest.T) {
	group := NewGroup("mixed", Handler{Protocol: "ws", Path: "socket", Handle: ax7StreamHandler})
	handlers := group.Handlers()
	coretest.AssertLen(t, handlers, 1)
	coretest.AssertEqual(t, ProtocolWebSocket, handlers[0].Protocol)
}

func TestAX7_Group_Name_Good(t *coretest.T) {
	group := NewGroup("events")
	name := group.Name()
	coretest.AssertEqual(t, "events", name)
	coretest.AssertNotEqual(t, "", name)
}

func TestAX7_Group_Name_Bad(t *coretest.T) {
	var group *Group
	name := group.Name()
	coretest.AssertEqual(t, "", name)
	coretest.AssertEmpty(t, name)
}

func TestAX7_Group_Name_Ugly(t *coretest.T) {
	group := NewGroup("  spaced  ")
	name := group.Name()
	coretest.AssertEqual(t, "spaced", name)
	coretest.AssertNotContains(t, name, " ")
}

func TestAX7_Group_Handlers_Good(t *coretest.T) {
	group := NewGroup("events", SSE("/updates", ax7StreamHandler))
	handlers := group.Handlers()
	coretest.AssertLen(t, handlers, 1)
	coretest.AssertEqual(t, "/updates", handlers[0].Path)
}

func TestAX7_Group_Handlers_Bad(t *coretest.T) {
	var group *Group
	handlers := group.Handlers()
	coretest.AssertNil(t, handlers)
	coretest.AssertEmpty(t, handlers)
}

func TestAX7_Group_Handlers_Ugly(t *coretest.T) {
	group := NewGroup("events", SSE("/updates", ax7StreamHandler))
	handlers := group.Handlers()
	handlers[0].Path = "/mutated"
	coretest.AssertEqual(t, "/updates", group.Handlers()[0].Path)
}

func TestAX7_Group_Register_Good(t *coretest.T) {
	group := NewGroup("events", WebSocket("/socket", ax7StreamHandler))
	reg := &ax7Registrar{}
	group.Register(reg)
	coretest.AssertEqual(t, "GET", reg.method)
	coretest.AssertEqual(t, "/socket", reg.path)
}

func TestAX7_Group_Register_Bad(t *coretest.T) {
	var group *Group
	reg := &ax7Registrar{}
	group.Register(reg)
	coretest.AssertEqual(t, 0, reg.count)
	coretest.AssertEqual(t, "", reg.path)
}

func TestAX7_Group_Register_Ugly(t *coretest.T) {
	group := NewGroup("events", Handler{Protocol: ProtocolSSE, Path: "/events"})
	reg := &ax7Registrar{}
	group.Register(reg)
	coretest.AssertEqual(t, 0, reg.count)
	coretest.AssertNil(t, group.Handlers())
}

func TestAX7_SSE_Good(t *coretest.T) {
	handler := SSE("/events", ax7StreamHandler)
	coretest.AssertEqual(t, ProtocolSSE, handler.Protocol)
	coretest.AssertEqual(t, "GET", handler.Method)
	coretest.AssertEqual(t, "/events", handler.Path)
}

func TestAX7_SSE_Bad(t *coretest.T) {
	handler := SSE("", nil)
	coretest.AssertEqual(t, ProtocolSSE, handler.Protocol)
	coretest.AssertEqual(t, "", handler.Path)
	coretest.AssertNil(t, handler.Handle)
}

func TestAX7_SSE_Ugly(t *coretest.T) {
	group := NewGroup("events", SSE("///events///", ax7StreamHandler))
	handler := group.Handlers()[0]
	coretest.AssertEqual(t, ProtocolSSE, handler.Protocol)
	coretest.AssertEqual(t, "/events", handler.Path)
}

func TestAX7_WebSocket_Good(t *coretest.T) {
	handler := WebSocket("/socket", ax7StreamHandler)
	coretest.AssertEqual(t, ProtocolWebSocket, handler.Protocol)
	coretest.AssertEqual(t, "GET", handler.Method)
	coretest.AssertEqual(t, "/socket", handler.Path)
}

func TestAX7_WebSocket_Bad(t *coretest.T) {
	handler := WebSocket("", nil)
	coretest.AssertEqual(t, ProtocolWebSocket, handler.Protocol)
	coretest.AssertEqual(t, "", handler.Path)
	coretest.AssertNil(t, handler.Handle)
}

func TestAX7_WebSocket_Ugly(t *coretest.T) {
	group := NewGroup("socket", WebSocket("///ws///", ax7StreamHandler))
	handler := group.Handlers()[0]
	coretest.AssertEqual(t, ProtocolWebSocket, handler.Protocol)
	coretest.AssertEqual(t, "/ws", handler.Path)
}
