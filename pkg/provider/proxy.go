// SPDX-Licence-Identifier: EUPL-1.2

package provider

// ProxyProvider will wrap polyglot (PHP/TS) providers that publish an OpenAPI
// spec and run their own HTTP handler. The Go API layer reverse-proxies to
// their endpoint.
//
// This is a Phase 3 feature. The type is declared here as a forward reference
// so the package structure is established.
//
// See the design spec SS Polyglot Providers for the full ProxyProvider contract.
