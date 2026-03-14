<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Front\Api\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * API Version Middleware.
 *
 * Parses the API version from the request and sets it on the request attributes.
 * Supports version extraction from:
 *
 * 1. URL path prefix: /api/v1/users, /api/v2/users
 * 2. Accept-Version header: Accept-Version: v1, Accept-Version: 2
 * 3. Accept header with vendor type: Accept: application/vnd.hosthub.v1+json
 *
 * The resolved version is stored in request attributes and can be accessed via:
 * - $request->attributes->get('api_version') - returns integer (e.g., 1, 2)
 * - $request->attributes->get('api_version_string') - returns string (e.g., 'v1', 'v2')
 *
 * ## Configuration
 *
 * Configure in config/api.php:
 * ```php
 * 'versioning' => [
 *     'default' => 1,           // Default version when none specified
 *     'current' => 1,           // Current/latest version
 *     'supported' => [1],       // List of supported versions
 *     'deprecated' => [],       // List of deprecated (but still supported) versions
 *     'sunset' => [],           // Versions with sunset dates: [1 => '2025-06-01']
 * ],
 * ```
 *
 * ## Usage in Routes
 *
 * ```php
 * // Apply to specific routes
 * Route::middleware('api.version')->group(function () {
 *     Route::get('/users', [UserController::class, 'index']);
 * });
 *
 * // Or with version constraint
 * Route::middleware('api.version:2')->group(function () {
 *     // Only accepts v2 requests
 * });
 * ```
 *
 * ## Deprecation Headers
 *
 * When a request uses a deprecated API version, the response includes:
 * - Deprecation: true
 * - Sunset: <date> (if configured)
 * - X-API-Warn: "API version X is deprecated..."
 *
 * @see ApiVersionService For programmatic version checks
 */
class ApiVersion
{
    /**
     * Handle an incoming request.
     *
     * @param  int|null  $requiredVersion  Minimum version required (optional)
     */
    public function handle(Request $request, Closure $next, ?int $requiredVersion = null): Response
    {
        $version = $this->resolveVersion($request);
        $versionConfig = config('api.versioning', []);

        $default = $versionConfig['default'] ?? 1;
        $current = $versionConfig['current'] ?? 1;
        $supported = $versionConfig['supported'] ?? [1];
        $deprecated = $versionConfig['deprecated'] ?? [];
        $sunset = $versionConfig['sunset'] ?? [];

        // Use default if no version specified
        if ($version === null) {
            $version = $default;
        }

        // Validate version is supported
        if (! in_array($version, $supported, true)) {
            return $this->unsupportedVersion($version, $supported, $current);
        }

        // Check minimum version requirement
        if ($requiredVersion !== null && $version < $requiredVersion) {
            return $this->versionTooLow($version, $requiredVersion);
        }

        // Store version in request
        $request->attributes->set('api_version', $version);
        $request->attributes->set('api_version_string', "v{$version}");

        /** @var Response $response */
        $response = $next($request);

        // Add version header to response
        $response->headers->set('X-API-Version', (string) $version);

        // Add deprecation headers if applicable
        if (in_array($version, $deprecated, true)) {
            $response->headers->set('Deprecation', 'true');
            $response->headers->set('X-API-Warn', "API version {$version} is deprecated. Please upgrade to v{$current}.");

            // Add Sunset header if configured
            if (isset($sunset[$version])) {
                $sunsetDate = $sunset[$version];
                // Convert to HTTP date format if not already
                if (! str_contains($sunsetDate, ',')) {
                    $sunsetDate = (new \DateTimeImmutable($sunsetDate))->format(\DateTimeInterface::RFC7231);
                }
                $response->headers->set('Sunset', $sunsetDate);
            }
        }

        return $response;
    }

    /**
     * Resolve the API version from the request.
     *
     * Priority order:
     * 1. URL path (/api/v1/...)
     * 2. Accept-Version header
     * 3. Accept header vendor type
     */
    protected function resolveVersion(Request $request): ?int
    {
        // 1. Check URL path for version prefix
        $version = $this->versionFromPath($request);
        if ($version !== null) {
            return $version;
        }

        // 2. Check Accept-Version header
        $version = $this->versionFromHeader($request);
        if ($version !== null) {
            return $version;
        }

        // 3. Check Accept header for vendor type
        return $this->versionFromAcceptHeader($request);
    }

    /**
     * Extract version from URL path.
     *
     * Matches: /api/v1/..., /api/v2/...
     */
    protected function versionFromPath(Request $request): ?int
    {
        $path = $request->path();

        // Match /api/v{n}/ or /v{n}/ at the start
        if (preg_match('#^(?:api/)?v(\d+)(?:/|$)#', $path, $matches)) {
            return (int) $matches[1];
        }

        return null;
    }

    /**
     * Extract version from Accept-Version header.
     *
     * Accepts: v1, v2, 1, 2
     */
    protected function versionFromHeader(Request $request): ?int
    {
        $header = $request->header('Accept-Version');

        if ($header === null) {
            return null;
        }

        // Strip 'v' prefix if present
        $version = ltrim($header, 'vV');

        if (is_numeric($version)) {
            return (int) $version;
        }

        return null;
    }

    /**
     * Extract version from Accept header vendor type.
     *
     * Matches: application/vnd.hosthub.v1+json
     */
    protected function versionFromAcceptHeader(Request $request): ?int
    {
        $accept = $request->header('Accept', '');

        // Match vendor media type: application/vnd.{name}.v{n}+json
        if (preg_match('#application/vnd\.[^.]+\.v(\d+)\+#', $accept, $matches)) {
            return (int) $matches[1];
        }

        return null;
    }

    /**
     * Return 400 response for unsupported version.
     *
     * @param  array<int>  $supported
     */
    protected function unsupportedVersion(int $requested, array $supported, int $current): Response
    {
        return response()->json([
            'error' => 'unsupported_api_version',
            'message' => "API version {$requested} is not supported.",
            'requested_version' => $requested,
            'supported_versions' => $supported,
            'current_version' => $current,
            'hint' => 'Use Accept-Version header or URL prefix (e.g., /api/v1/) to specify version.',
        ], 400, [
            'X-API-Version' => (string) $current,
        ]);
    }

    /**
     * Return 400 response when version is too low.
     */
    protected function versionTooLow(int $requested, int $required): Response
    {
        return response()->json([
            'error' => 'api_version_too_low',
            'message' => "This endpoint requires API version {$required} or higher.",
            'requested_version' => $requested,
            'minimum_version' => $required,
        ], 400, [
            'X-API-Version' => (string) $requested,
        ]);
    }
}
