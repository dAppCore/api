<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Apply declarative Cache-Control policies to API responses.
 *
 * Usage examples:
 * - Route::middleware('api.cache:ephemeral')->post('/token', ...)
 * - Route::get('/workspaces', ...)->defaults('api_cache_control', 'cacheable')
 */
class ApiCacheControl
{
    /**
     * Handle an incoming request.
     */
    public function handle(Request $request, Closure $next, ?string $profile = null): Response
    {
        $response = $next($request);

        if (! $response->isSuccessful()) {
            return $response;
        }

        if ($response->headers->has('Cache-Control')) {
            return $response;
        }

        $policy = $this->resolvePolicy($request, $profile);
        if ($policy === null) {
            return $response;
        }

        $response->headers->set('Cache-Control', $policy);

        if (str_contains($policy, 'no-cache') && ! $response->headers->has('Pragma')) {
            $response->headers->set('Pragma', 'no-cache');
        }

        return $response;
    }

    protected function resolvePolicy(Request $request, ?string $profile): ?string
    {
        $route = $request->route();
        $routeValue = $route?->defaults['api_cache_control']
            ?? $route?->getAction('api_cache_control');

        $candidate = is_string($routeValue) && trim($routeValue) !== ''
            ? trim($routeValue)
            : trim((string) $profile);

        if ($candidate === '') {
            return null;
        }

        $policies = config('api.cache_control.profiles', []);
        $policy = $policies[$candidate] ?? $candidate;

        return is_string($policy) && trim($policy) !== ''
            ? trim($policy)
            : null;
    }
}
