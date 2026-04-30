<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Closure;
use Core\Api\Exceptions\RateLimitExceededException;
use Core\Api\Models\ApiKey;
use Core\Api\RateLimit\RateLimit;
use Core\Api\RateLimit\RateLimitResult;
use Core\Api\RateLimit\RateLimitService;
use Illuminate\Http\Request;
use ReflectionClass;
use ReflectionMethod;
use Symfony\Component\HttpFoundation\Response;

/**
 * Rate limit API requests with granular control.
 *
 * Supports:
 * - Per-endpoint rate limits via config or #[RateLimit] attribute
 * - Per-workspace rate limits with workspace ID in key
 * - Per-API key rate limits
 * - Tier-based limits based on workspace subscription
 * - Burst allowance configuration
 * - Standard rate limit headers (X-RateLimit-*)
 *
 * Priority (highest to lowest):
 * 1. Method-level #[RateLimit] attribute
 * 2. Class-level #[RateLimit] attribute
 * 3. Per-endpoint config (api.rate_limits.endpoints.{route_name})
 * 4. Tier-based limits (api.rate_limits.tiers.{tier})
 * 5. Default authenticated limits
 * 6. Default unauthenticated limits
 *
 * Register in bootstrap/app.php:
 *   ->withMiddleware(function (Middleware $middleware) {
 *       $middleware->alias([
 *           'api.rate' => \Core\Api\Middleware\RateLimitApi::class,
 *       ]);
 *   })
 */
class RateLimitApi
{
    public function __construct(
        protected RateLimitService $rateLimitService,
    ) {}

    public function handle(Request $request, Closure $next): Response
    {
        // Check if rate limiting is enabled
        if (! config('api.rate_limits.enabled', true)) {
            return $next($request);
        }

        $rateLimitConfig = $this->resolveRateLimitConfig($request);
        $key = $this->resolveRateLimitKey($request, $rateLimitConfig);

        // Perform rate limit check and hit
        $result = $this->rateLimitService->hit(
            key: $key,
            limit: $rateLimitConfig['limit'],
            window: $rateLimitConfig['window'],
            burst: $rateLimitConfig['burst'],
        );

        if (! $result->allowed) {
            throw new RateLimitExceededException($result);
        }

        $response = $next($request);

        return $this->addRateLimitHeaders($response, $result);
    }

    /**
     * Resolve the rate limit configuration for the request.
     *
     * @return array{limit: int, window: int, burst: float, key: string|null}
     */
    protected function resolveRateLimitConfig(Request $request): array
    {
        $defaults = config('api.rate_limits.default', [
            'limit' => 60,
            'window' => 60,
            'burst' => 1.0,
        ]);

        // 1. Check for #[RateLimit] attribute on controller/method
        $attributeConfig = $this->getAttributeRateLimit($request);
        if ($attributeConfig !== null) {
            return array_merge($defaults, $attributeConfig);
        }

        // 2. Check for per-endpoint config
        $endpointConfig = $this->getEndpointRateLimit($request);
        if ($endpointConfig !== null) {
            return array_merge($defaults, $endpointConfig);
        }

        // 3. Check for tier-based limits
        $tierConfig = $this->getTierRateLimit($request);
        if ($tierConfig !== null) {
            return array_merge($defaults, $tierConfig);
        }

        // 4. Use authenticated limits if authenticated
        if ($this->isAuthenticated($request)) {
            $authenticated = config('api.rate_limits.authenticated', $defaults);

            return [
                'limit' => $authenticated['requests'] ?? $authenticated['limit'] ?? $defaults['limit'],
                'window' => ($authenticated['per_minutes'] ?? 1) * 60,
                'burst' => $authenticated['burst'] ?? $defaults['burst'] ?? 1.0,
                'key' => null,
            ];
        }

        // 5. Use default limits
        return [
            'limit' => $defaults['requests'] ?? $defaults['limit'] ?? 60,
            'window' => ($defaults['per_minutes'] ?? 1) * 60,
            'burst' => $defaults['burst'] ?? 1.0,
            'key' => null,
        ];
    }

    /**
     * Get rate limit from #[RateLimit] attribute.
     *
     * @return array{limit: int, window: int, burst: float, key: string|null}|null
     */
    protected function getAttributeRateLimit(Request $request): ?array
    {
        $route = $request->route();
        if (! $route) {
            return null;
        }

        $controller = $route->getController();
        $method = $route->getActionMethod();

        if (! $controller || ! $method) {
            return null;
        }

        try {
            // Check method-level attribute first
            $reflection = new ReflectionMethod($controller, $method);
            $attributes = $reflection->getAttributes(RateLimit::class);

            if (! empty($attributes)) {
                /** @var RateLimit $rateLimit */
                $rateLimit = $attributes[0]->newInstance();

                return [
                    'limit' => $rateLimit->limit,
                    'window' => $rateLimit->window,
                    'burst' => $rateLimit->burst,
                    'key' => $rateLimit->key,
                ];
            }

            // Check class-level attribute
            $classReflection = new ReflectionClass($controller);
            $classAttributes = $classReflection->getAttributes(RateLimit::class);

            if (! empty($classAttributes)) {
                /** @var RateLimit $rateLimit */
                $rateLimit = $classAttributes[0]->newInstance();

                return [
                    'limit' => $rateLimit->limit,
                    'window' => $rateLimit->window,
                    'burst' => $rateLimit->burst,
                    'key' => $rateLimit->key,
                ];
            }
        } catch (\ReflectionException) {
            // Controller or method doesn't exist
        }

        return null;
    }

    /**
     * Get rate limit from per-endpoint config.
     *
     * @return array{limit: int, window: int, burst: float, key: string|null}|null
     */
    protected function getEndpointRateLimit(Request $request): ?array
    {
        $route = $request->route();
        if (! $route) {
            return null;
        }

        $routeName = $route->getName();
        if (! $routeName) {
            return null;
        }

        // Try exact match first (e.g., "api.users.index")
        $config = config("api.rate_limits.endpoints.{$routeName}");

        // Try with dots replaced (e.g., "users.index" for route "api.users.index")
        if (! $config) {
            $shortName = preg_replace('/^api\./', '', $routeName);
            $config = config("api.rate_limits.endpoints.{$shortName}");
        }

        if (! $config) {
            return null;
        }

        return [
            'limit' => $config['limit'] ?? $config['requests'] ?? 60,
            'window' => $config['window'] ?? (($config['per_minutes'] ?? 1) * 60),
            'burst' => $config['burst'] ?? 1.0,
            'key' => $config['key'] ?? null,
        ];
    }

    /**
     * Get tier-based rate limit from workspace subscription.
     *
     * @return array{limit: int, window: int, burst: float, key: string|null}|null
     */
    protected function getTierRateLimit(Request $request): ?array
    {
        $workspace = $request->attributes->get('workspace');
        if (! is_object($workspace)) {
            return null;
        }

        $tier = $this->getWorkspaceTier($workspace);
        $tierConfig = config("api.rate_limits.tiers.{$tier}");

        if (! $tierConfig) {
            // Fall back to by_tier for backwards compatibility
            $tierConfig = config("api.rate_limits.by_tier.{$tier}");
        }

        if (! $tierConfig) {
            return null;
        }

        return [
            'limit' => $tierConfig['limit'] ?? $tierConfig['requests'] ?? 60,
            'window' => $tierConfig['window'] ?? (($tierConfig['per_minutes'] ?? 1) * 60),
            'burst' => $tierConfig['burst'] ?? 1.0,
            'key' => null,
        ];
    }

    /**
     * Resolve the rate limit key for the request.
     *
     * @param  array{limit: int, window: int, burst: float, key: string|null}  $config
     */
    protected function resolveRateLimitKey(Request $request, array $config): string
    {
        $parts = [];

        // Use custom key suffix if provided
        $suffix = $config['key'];

        // Add endpoint to key if per_workspace is enabled and we have a route
        $perWorkspace = config('api.rate_limits.per_workspace', true);
        $route = $request->route();

        // Build identifier based on auth context
        $apiKey = $request->attributes->get('api_key');
        $workspace = $request->attributes->get('workspace');
        $user = $request->user();

        if ($apiKey instanceof ApiKey) {
            $parts[] = "api_key:{$apiKey->id}";

            // Include workspace if per_workspace is enabled
            $workspaceId = $this->resolveWorkspaceId($workspace);
            if ($perWorkspace && $workspaceId !== null) {
                $parts[] = "ws:{$workspaceId}";
            }
        } elseif ($user !== null) {
            $parts[] = "user:{$user->id}";

            $workspaceId = $this->resolveWorkspaceId($workspace);
            if ($perWorkspace && $workspaceId !== null) {
                $parts[] = "ws:{$workspaceId}";
            }
        } else {
            $parts[] = "ip:{$request->ip()}";
        }

        // Add route name for per-endpoint isolation
        if ($route && $route->getName()) {
            $parts[] = "route:{$route->getName()}";
        }

        // Add custom suffix if provided
        if ($suffix) {
            $parts[] = $suffix;
        }

        return implode(':', $parts);
    }

    /**
     * Get workspace tier for rate limiting.
     */
    protected function getWorkspaceTier(mixed $workspace): string
    {
        if (! is_object($workspace)) {
            return 'free';
        }

        // Check if workspace has an active package/subscription
        if (method_exists($workspace, 'activePackages')) {
            $package = $workspace->activePackages()->first();

            return $package?->slug ?? 'free';
        }

        // Check for a tier attribute
        if (property_exists($workspace, 'tier')) {
            return $workspace->tier ?? 'free';
        }

        // Check for a plan attribute
        if (property_exists($workspace, 'plan')) {
            return $workspace->plan ?? 'free';
        }

        return 'free';
    }

    /**
     * Resolve a workspace identifier from the request context.
     *
     * Malformed workspace attributes are treated as absent so the middleware
     * fails closed instead of dereferencing untrusted values.
     */
    protected function resolveWorkspaceId(mixed $workspace): int|string|null
    {
        if (! is_object($workspace)) {
            return null;
        }

        if (method_exists($workspace, 'getKey')) {
            $key = $workspace->getKey();

            if (is_int($key) || is_string($key)) {
                return $key;
            }
        }

        if (property_exists($workspace, 'id')) {
            $id = $workspace->id;

            if (is_int($id) || is_string($id)) {
                return $id;
            }
        }

        return null;
    }

    /**
     * Check if the request is authenticated.
     */
    protected function isAuthenticated(Request $request): bool
    {
        return $request->attributes->get('api_key') !== null
            || $request->user() !== null;
    }

    /**
     * Add rate limit headers to response.
     */
    protected function addRateLimitHeaders(Response $response, RateLimitResult $result): Response
    {
        foreach ($result->headers() as $header => $value) {
            $response->headers->set($header, (string) $value);
        }

        return $response;
    }
}
