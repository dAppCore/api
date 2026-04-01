<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Extensions;

use Core\Api\Documentation\Extension;
use Illuminate\Routing\Route;

/**
 * Sunset Extension.
 *
 * Documents endpoint deprecation and sunset metadata for routes using
 * the `api.sunset` middleware.
 */
class SunsetExtension implements Extension
{
    /**
     * Extend the complete OpenAPI specification.
     */
    public function extend(array $spec, array $config): array
    {
        $spec['components']['headers'] = $spec['components']['headers'] ?? [];

        $spec['components']['headers']['deprecation'] = [
            'description' => 'Indicates that the endpoint is deprecated.',
            'schema' => [
                'type' => 'string',
                'enum' => ['true'],
            ],
        ];

        $spec['components']['headers']['sunset'] = [
            'description' => 'The date and time after which the endpoint will no longer be supported.',
            'schema' => [
                'type' => 'string',
                'format' => 'date-time',
            ],
        ];

        $spec['components']['headers']['link'] = [
            'description' => 'Reference to the successor endpoint, when one is provided.',
            'schema' => [
                'type' => 'string',
            ],
        ];

        $spec['components']['headers']['xapiwarn'] = [
            'description' => 'Human-readable deprecation warning for clients.',
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
        $sunset = $this->sunsetMiddlewareArguments($route);

        if ($sunset === null) {
            return $operation;
        }

        $operation['deprecated'] = true;

        foreach ($operation['responses'] as $status => &$response) {
            if (! is_numeric($status) || (int) $status < 200 || (int) $status >= 300) {
                continue;
            }

            $response['headers'] = $response['headers'] ?? [];

            $response['headers']['Deprecation'] = [
                '$ref' => '#/components/headers/deprecation',
            ];
            if ($sunset['sunsetDate'] !== null && $sunset['sunsetDate'] !== '') {
                $response['headers']['Sunset'] = [
                    '$ref' => '#/components/headers/sunset',
                ];
            }
            $response['headers']['X-API-Warn'] = [
                '$ref' => '#/components/headers/xapiwarn',
            ];

            if (
                $sunset['replacement'] !== null
                && $sunset['replacement'] !== ''
                && ! isset($response['headers']['Link'])
            ) {
                $response['headers']['Link'] = [
                    '$ref' => '#/components/headers/link',
                ];
            }
        }
        unset($response);

        return $operation;
    }

    /**
     * Extract the configured sunset middleware arguments from a route.
     *
     * Returns null when the route does not use the sunset middleware.
     *
     * @return array{sunsetDate:?string,replacement:?string}|null
     */
    protected function sunsetMiddlewareArguments(Route $route): ?array
    {
        foreach ($route->middleware() as $middleware) {
            if (! str_starts_with($middleware, 'api.sunset') && ! str_contains($middleware, 'ApiSunset')) {
                continue;
            }

            $arguments = null;

            if (str_contains($middleware, ':')) {
                [, $arguments] = explode(':', $middleware, 2);
            }

            if ($arguments === null || $arguments === '') {
                return [
                    'sunsetDate' => null,
                    'replacement' => null,
                ];
            }

            $parts = explode(',', $arguments, 2);
            $sunsetDate = trim($parts[0] ?? '');
            $replacement = isset($parts[1]) ? trim($parts[1]) : null;
            if ($replacement === '') {
                $replacement = null;
            }

            return [
                'sunsetDate' => $sunsetDate !== '' ? $sunsetDate : null,
                'replacement' => $replacement,
            ];
        }

        return null;
    }
}
