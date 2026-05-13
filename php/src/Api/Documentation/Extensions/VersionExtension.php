<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Extensions;

use Core\Api\Documentation\Extension;
use Illuminate\Routing\Route;

/**
 * API Version Extension.
 *
 * Documents the X-API-Version response header and version-driven deprecation
 * metadata for routes using the api.version middleware.
 */
class VersionExtension implements Extension
{
    /**
     * Extend the complete OpenAPI specification.
     */
    public function extend(array $spec, array $config): array
    {
        if (! (bool) config('api.headers.include_version', true)) {
            return $spec;
        }

        $spec['components']['headers'] = $spec['components']['headers'] ?? [];
        $spec['components']['headers']['xapiversion'] = [
            'description' => 'API version used to process the request.',
            'schema' => [
                'type' => 'string',
            ],
        ];

        return $spec;
    }

    /**
     * Extend an individual operation.
     */
    public function extendOperation(array $operation, Route $route, string $method, array $config): array
    {
        $version = $this->versionMiddlewareVersion($route);
        if ($version === null) {
            return $operation;
        }

        $includeVersion = (bool) config('api.headers.include_version', true);
        $includeDeprecation = (bool) config('api.headers.include_deprecation', true);

        $deprecatedVersions = array_map('intval', config('api.versioning.deprecated', []));
        $sunsetDates = config('api.versioning.sunset', []);
        $isDeprecatedVersion = in_array($version, $deprecatedVersions, true);
        $sunsetDate = $sunsetDates[$version] ?? null;

        if ($isDeprecatedVersion) {
            $operation['deprecated'] = true;
        }

        foreach ($operation['responses'] as $status => &$response) {
            if (! is_numeric($status) || (int) $status < 200 || (int) $status >= 600) {
                continue;
            }

            $response['headers'] = $response['headers'] ?? [];

            if ($includeVersion && ! isset($response['headers']['X-API-Version'])) {
                $response['headers']['X-API-Version'] = [
                    '$ref' => '#/components/headers/xapiversion',
                ];
            }

            if (! $includeDeprecation || ! $isDeprecatedVersion) {
                continue;
            }

            $response['headers']['Deprecation'] = [
                '$ref' => '#/components/headers/deprecation',
            ];
            $response['headers']['X-API-Warn'] = [
                '$ref' => '#/components/headers/xapiwarn',
            ];

            if ($sunsetDate !== null && $sunsetDate !== '') {
                $response['headers']['Sunset'] = [
                    '$ref' => '#/components/headers/sunset',
                ];
            }
        }
        unset($response);

        return $operation;
    }

    /**
     * Extract the version number from api.version middleware.
     */
    protected function versionMiddlewareVersion(Route $route): ?int
    {
        foreach ($route->middleware() as $middleware) {
            if (! str_starts_with($middleware, 'api.version') && ! str_contains($middleware, 'ApiVersion')) {
                continue;
            }

            if (! str_contains($middleware, ':')) {
                return null;
            }

            [, $arguments] = explode(':', $middleware, 2);
            $arguments = trim($arguments);
            if ($arguments === '') {
                return null;
            }

            $parts = explode(',', $arguments, 2);
            $version = ltrim(trim($parts[0] ?? ''), 'vV');
            if ($version === '' || ! is_numeric($version)) {
                return null;
            }

            return (int) $version;
        }

        return null;
    }
}
