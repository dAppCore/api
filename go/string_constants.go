// SPDX-License-Identifier: EUPL-1.2

package api

// Shared string constants to reduce duplication.

const (
	hdrContentType     = "Content-Type"
	hdrContentEncoding = "Content-Encoding"
	mimeJSON           = "application/json"

	errBridgeValidate         = "ToolBridge.Validate"
	errBridgeValidateResp     = "ToolBridge.ValidateResponse"
	errBridgeValidateSchema   = "ToolBridge.ValidateSchema"
	errClientCall             = "OpenAPIClient.Call"
	errClientLoadSpec         = "OpenAPIClient.loadSpec"
	errClientBuildURL         = "OpenAPIClient.buildURL"
	errClientValidateSchema   = "OpenAPIClient.validateOpenAPISchema"
	errClientValidateResponse = "OpenAPIClient.validateOpenAPIResponse"
	errSDKGenerate            = "SDKGenerator.Generate"

	msgBadRequest      = "Bad request"
	msgTooManyRequests = "Too many requests"
	msgGatewayTimeout  = "Gateway timeout"
	msgInternalSrvErr  = "Internal server error"

	toolResponsePrefix = "ToolBridge.ValidateResponse"
	toolSchemaPrefix   = "ToolBridge.ValidateSchema"
)
