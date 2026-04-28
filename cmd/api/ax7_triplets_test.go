// SPDX-License-Identifier: EUPL-1.2

package api

import . "dappco.re/go"

func TestAX7_AddAPICommands_Good(t *T) {
	c := New()
	AddAPICommands(c)
	r := c.Command("api/spec")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestAX7_AddAPICommands_Bad(t *T) {
	var c *Core
	AssertPanics(t, func() {
		AddAPICommands(c)
	})
	AssertNil(t, c)
}

func TestAX7_AddAPICommands_Ugly(t *T) {
	c := New()
	AddAPICommands(c)
	AddAPICommands(c)
	r := c.Command("api/sdk")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}
