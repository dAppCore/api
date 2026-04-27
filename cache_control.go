// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// cacheControlPolicies snapshots declared Cache-Control hints from
// DescribableGroups so runtime handlers can advertise the same policy that the
// OpenAPI document describes.
func cacheControlPolicies(groups []RouteGroup) map[string]string {
	prepared := prepareRouteGroups(groups)
	if len(prepared) == 0 {
		return nil
	}

	policies := make(map[string]string)
	for _, group := range prepared {
		for _, rd := range group.descs {
			policy := core.Trim(rd.CacheControl)
			method := core.Upper(core.Trim(rd.Method))
			if policy == "" || method == "" {
				continue
			}

			path := openAPIPathToGinPath(joinOpenAPIPath(group.basePath, rd.Path))
			policies[method+" "+path] = policy
		}
	}
	if len(policies) == 0 {
		return nil
	}

	return policies
}

// cacheControlMiddleware applies RouteDescription.CacheControl to successful
// responses once Gin has resolved the matched route template.
func cacheControlMiddleware(policies map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(policies) == 0 {
			return
		}
		if c.Writer.Status() < http.StatusOK || c.Writer.Status() >= http.StatusMultipleChoices {
			return
		}
		if c.Writer.Header().Get("Cache-Control") != "" {
			return
		}

		fullPath := c.FullPath()
		if fullPath == "" {
			return
		}
		if policy := policies[c.Request.Method+" "+fullPath]; policy != "" {
			c.Header("Cache-Control", policy)
		}
	}
}

func openAPIPathToGinPath(path string) string {
	path = normaliseOpenAPIPath(path)
	if path == "/" {
		return path
	}

	segments := core.Split(path, "/")
	out := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = core.Trim(segment)
		if segment == "" {
			continue
		}
		if len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}' {
			name := core.Trim(segment[1 : len(segment)-1])
			if name != "" && !core.Contains(name, "/") && !core.Contains(name, "{") && !core.Contains(name, "}") {
				segment = ":" + name
			}
		}
		out = append(out, segment)
	}

	return "/" + core.Join("/", out...)
}
