<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Front\Api;

use Illuminate\Http\Request;

/**
 * API Version Service.
 *
 * Provides helper methods for working with API versions in controllers
 * and other application code.
 *
 * ## Usage in Controllers
 *
 * ```php
 * use Core\Front\Api\ApiVersionService;
 *
 * class UserController
 * {
 *     public function __construct(
 *         protected ApiVersionService $versions
 *     ) {}
 *
 *     public function index(Request $request)
 *     {
 *         if ($this->versions->isV2($request)) {
 *             return $this->indexV2($request);
 *         }
 *         return $this->indexV1($request);
 *     }
 * }
 * ```
 *
 * ## Version Negotiation
 *
 * The service supports version negotiation where controllers can provide
 * different responses based on the requested version:
 *
 * ```php
 * return $this->versions->negotiate($request, [
 *     1 => fn() => $this->responseV1(),
 *     2 => fn() => $this->responseV2(),
 * ]);
 * ```
 */
class ApiVersionService
{
    /**
     * Get the current API version from the request.
     *
     * Returns null if no version middleware has processed the request.
     */
    public function current(?Request $request = null): ?int
    {
        $request ??= request();

        return $request->attributes->get('api_version');
    }

    /**
     * Get the current API version as a string (e.g., 'v1').
     */
    public function currentString(?Request $request = null): ?string
    {
        $request ??= request();

        return $request->attributes->get('api_version_string');
    }

    /**
     * Check if the request is for a specific version.
     */
    public function is(int $version, ?Request $request = null): bool
    {
        return $this->current($request) === $version;
    }

    /**
     * Check if the request is for version 1.
     */
    public function isV1(?Request $request = null): bool
    {
        return $this->is(1, $request);
    }

    /**
     * Check if the request is for version 2.
     */
    public function isV2(?Request $request = null): bool
    {
        return $this->is(2, $request);
    }

    /**
     * Check if the request version is at least the given version.
     */
    public function isAtLeast(int $version, ?Request $request = null): bool
    {
        $current = $this->current($request);

        return $current !== null && $current >= $version;
    }

    /**
     * Check if the current version is deprecated.
     */
    public function isDeprecated(?Request $request = null): bool
    {
        $current = $this->current($request);
        $deprecated = config('api.versioning.deprecated', []);

        return $current !== null && in_array($current, $deprecated, true);
    }

    /**
     * Get the configured default version.
     */
    public function defaultVersion(): int
    {
        return (int) config('api.versioning.default', 1);
    }

    /**
     * Get the current/latest version.
     */
    public function latestVersion(): int
    {
        return (int) config('api.versioning.current', 1);
    }

    /**
     * Get all supported versions.
     *
     * @return array<int>
     */
    public function supportedVersions(): array
    {
        return config('api.versioning.supported', [1]);
    }

    /**
     * Get all deprecated versions.
     *
     * @return array<int>
     */
    public function deprecatedVersions(): array
    {
        return config('api.versioning.deprecated', []);
    }

    /**
     * Get sunset dates for versions.
     *
     * @return array<int, string>
     */
    public function sunsetDates(): array
    {
        return config('api.versioning.sunset', []);
    }

    /**
     * Check if a version is supported.
     */
    public function isSupported(int $version): bool
    {
        return in_array($version, $this->supportedVersions(), true);
    }

    /**
     * Negotiate response based on API version.
     *
     * Calls the appropriate handler based on the request's API version.
     * Falls back to lower version handlers if exact match not found.
     *
     * ```php
     * return $versions->negotiate($request, [
     *     1 => fn() => ['format' => 'v1'],
     *     2 => fn() => ['format' => 'v2', 'extra' => 'field'],
     * ]);
     * ```
     *
     * @param  array<int, callable>  $handlers  Version handlers keyed by version number
     * @return mixed Result from the appropriate handler
     *
     * @throws \InvalidArgumentException If no suitable handler found
     */
    public function negotiate(Request $request, array $handlers): mixed
    {
        $version = $this->current($request) ?? $this->defaultVersion();

        // Try exact match first
        if (isset($handlers[$version])) {
            return $handlers[$version]();
        }

        // Fall back to highest version that's <= requested version
        krsort($handlers);
        foreach ($handlers as $handlerVersion => $handler) {
            if ($handlerVersion <= $version) {
                return $handler();
            }
        }

        // No suitable handler found
        throw new \InvalidArgumentException(
            "No handler found for API version {$version}. Available versions: ".implode(', ', array_keys($handlers))
        );
    }

    /**
     * Transform response data based on API version.
     *
     * Useful for removing or adding fields based on version.
     *
     * ```php
     * return $versions->transform($request, $data, [
     *     1 => fn($data) => Arr::except($data, ['new_field']),
     *     2 => fn($data) => $data,
     * ]);
     * ```
     *
     * @param  array<int, callable>  $transformers  Version transformers
     */
    public function transform(Request $request, mixed $data, array $transformers): mixed
    {
        $version = $this->current($request) ?? $this->defaultVersion();

        // Try exact match first
        if (isset($transformers[$version])) {
            return $transformers[$version]($data);
        }

        // Fall back to highest version that's <= requested version
        krsort($transformers);
        foreach ($transformers as $transformerVersion => $transformer) {
            if ($transformerVersion <= $version) {
                return $transformer($data);
            }
        }

        // No transformer, return data unchanged
        return $data;
    }
}
