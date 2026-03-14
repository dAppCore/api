<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Extensions;

use Core\Api\Documentation\Extension;
use Core\Api\RateLimit\RateLimit;
use Illuminate\Routing\Route;
use ReflectionClass;

/**
 * Rate Limit Extension.
 *
 * Documents rate limit headers in API responses and extracts rate limit
 * information from the #[RateLimit] attribute.
 */
class RateLimitExtension implements Extension
{
    /**
     * Extend the complete OpenAPI specification.
     */
    public function extend(array $spec, array $config): array
    {
        $rateLimitConfig = $config['rate_limits'] ?? [];

        if (! ($rateLimitConfig['enabled'] ?? true)) {
            return $spec;
        }

        // Add rate limit headers to components
        $headers = $rateLimitConfig['headers'] ?? [
            'X-RateLimit-Limit' => 'Maximum number of requests allowed per window',
            'X-RateLimit-Remaining' => 'Number of requests remaining in the current window',
            'X-RateLimit-Reset' => 'Unix timestamp when the rate limit window resets',
        ];

        $spec['components']['headers'] = $spec['components']['headers'] ?? [];

        foreach ($headers as $name => $description) {
            $headerKey = str_replace(['-', ' '], '', strtolower($name));
            $spec['components']['headers'][$headerKey] = [
                'description' => $description,
                'schema' => [
                    'type' => 'integer',
                ],
            ];
        }

        // Add 429 response schema to components
        $spec['components']['responses']['RateLimitExceeded'] = [
            'description' => 'Rate limit exceeded',
            'headers' => [
                'X-RateLimit-Limit' => [
                    '$ref' => '#/components/headers/xratelimitlimit',
                ],
                'X-RateLimit-Remaining' => [
                    '$ref' => '#/components/headers/xratelimitremaining',
                ],
                'X-RateLimit-Reset' => [
                    '$ref' => '#/components/headers/xratelimitreset',
                ],
                'Retry-After' => [
                    'description' => 'Seconds to wait before retrying',
                    'schema' => ['type' => 'integer'],
                ],
            ],
            'content' => [
                'application/json' => [
                    'schema' => [
                        'type' => 'object',
                        'properties' => [
                            'message' => [
                                'type' => 'string',
                                'example' => 'Too Many Requests',
                            ],
                            'retry_after' => [
                                'type' => 'integer',
                                'description' => 'Seconds until rate limit resets',
                                'example' => 30,
                            ],
                        ],
                    ],
                ],
            ],
        ];

        return $spec;
    }

    /**
     * Extend an individual operation.
     */
    public function extendOperation(array $operation, Route $route, string $method, array $config): array
    {
        $rateLimitConfig = $config['rate_limits'] ?? [];

        if (! ($rateLimitConfig['enabled'] ?? true)) {
            return $operation;
        }

        // Check if route has rate limiting middleware
        if (! $this->hasRateLimiting($route)) {
            return $operation;
        }

        // Add rate limit headers to successful responses
        foreach ($operation['responses'] as $status => &$response) {
            if ((int) $status >= 200 && (int) $status < 300) {
                $response['headers'] = $response['headers'] ?? [];
                $response['headers']['X-RateLimit-Limit'] = [
                    '$ref' => '#/components/headers/xratelimitlimit',
                ];
                $response['headers']['X-RateLimit-Remaining'] = [
                    '$ref' => '#/components/headers/xratelimitremaining',
                ];
                $response['headers']['X-RateLimit-Reset'] = [
                    '$ref' => '#/components/headers/xratelimitreset',
                ];
            }
        }

        // Add 429 response
        $operation['responses']['429'] = [
            '$ref' => '#/components/responses/RateLimitExceeded',
        ];

        // Extract rate limit from attribute and add to description
        $rateLimit = $this->extractRateLimit($route);
        if ($rateLimit !== null) {
            $limitInfo = sprintf(
                '**Rate Limit:** %d requests per %d seconds',
                $rateLimit['limit'],
                $rateLimit['window']
            );

            if ($rateLimit['burst'] > 1.0) {
                $limitInfo .= sprintf(' (%.0f%% burst allowed)', ($rateLimit['burst'] - 1) * 100);
            }

            $operation['description'] = isset($operation['description'])
                ? $operation['description']."\n\n".$limitInfo
                : $limitInfo;
        }

        return $operation;
    }

    /**
     * Check if route has rate limiting.
     */
    protected function hasRateLimiting(Route $route): bool
    {
        $middleware = $route->middleware();

        foreach ($middleware as $m) {
            if (str_contains($m, 'throttle') ||
                str_contains($m, 'rate') ||
                str_contains($m, 'api.rate') ||
                str_contains($m, 'RateLimit')) {
                return true;
            }
        }

        // Also check for RateLimit attribute on controller
        $controller = $route->getController();
        if ($controller !== null) {
            $reflection = new ReflectionClass($controller);
            if (! empty($reflection->getAttributes(RateLimit::class))) {
                return true;
            }

            $action = $route->getActionMethod();
            if ($reflection->hasMethod($action)) {
                $method = $reflection->getMethod($action);
                if (! empty($method->getAttributes(RateLimit::class))) {
                    return true;
                }
            }
        }

        return false;
    }

    /**
     * Extract rate limit configuration from route.
     */
    protected function extractRateLimit(Route $route): ?array
    {
        $controller = $route->getController();

        if ($controller === null) {
            return null;
        }

        $reflection = new ReflectionClass($controller);
        $action = $route->getActionMethod();

        // Check method first
        if ($reflection->hasMethod($action)) {
            $method = $reflection->getMethod($action);
            $attrs = $method->getAttributes(RateLimit::class);
            if (! empty($attrs)) {
                $rateLimit = $attrs[0]->newInstance();

                return [
                    'limit' => $rateLimit->limit,
                    'window' => $rateLimit->window,
                    'burst' => $rateLimit->burst,
                ];
            }
        }

        // Check class
        $attrs = $reflection->getAttributes(RateLimit::class);
        if (! empty($attrs)) {
            $rateLimit = $attrs[0]->newInstance();

            return [
                'limit' => $rateLimit->limit,
                'window' => $rateLimit->window,
                'burst' => $rateLimit->burst,
            ];
        }

        return null;
    }
}
