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
        if (! $this->hasSunsetMiddleware($route)) {
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
            $response['headers']['Sunset'] = [
                '$ref' => '#/components/headers/sunset',
            ];
            $response['headers']['X-API-Warn'] = [
                '$ref' => '#/components/headers/xapiwarn',
            ];

            if (! isset($response['headers']['Link'])) {
                $response['headers']['Link'] = [
                    '$ref' => '#/components/headers/link',
                ];
            }
        }
        unset($response);

        return $operation;
    }

    /**
     * Determine whether the route uses the sunset middleware.
     */
    protected function hasSunsetMiddleware(Route $route): bool
    {
        foreach ($route->middleware() as $middleware) {
            if (str_starts_with($middleware, 'api.sunset') || str_contains($middleware, 'ApiSunset')) {
                return true;
            }
        }

        return false;
    }
}
