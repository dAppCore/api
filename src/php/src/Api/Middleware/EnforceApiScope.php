<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Closure;
use Core\Api\Models\ApiKey;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Automatically enforce API key scopes based on HTTP method.
 *
 * Scope mapping:
 * - GET, HEAD, OPTIONS -> read
 * - POST, PUT, PATCH   -> write
 * - DELETE             -> delete
 *
 * Usage: Add to routes alongside api.auth middleware.
 * Route::middleware(['api.auth', 'api.scope.enforce'])->group(...)
 *
 * For routes that need to override the auto-detection, use CheckApiScope:
 * Route::middleware(['api.auth', 'api.scope:read'])->post('/readonly-action', ...)
 */
class EnforceApiScope
{
    /**
     * HTTP method to required scope mapping.
     */
    protected const METHOD_SCOPES = [
        'GET' => ApiKey::SCOPE_READ,
        'HEAD' => ApiKey::SCOPE_READ,
        'OPTIONS' => ApiKey::SCOPE_READ,
        'POST' => ApiKey::SCOPE_WRITE,
        'PUT' => ApiKey::SCOPE_WRITE,
        'PATCH' => ApiKey::SCOPE_WRITE,
        'DELETE' => ApiKey::SCOPE_DELETE,
    ];

    public function handle(Request $request, Closure $next): Response
    {
        $apiKey = $request->attributes->get('api_key');

        // If not authenticated via API key, allow through
        // Session auth and Sanctum handle their own permissions
        if (! $apiKey instanceof ApiKey) {
            return $next($request);
        }

        $method = strtoupper($request->method());
        $requiredScope = self::METHOD_SCOPES[$method] ?? ApiKey::SCOPE_READ;

        if (! $apiKey->hasScope($requiredScope)) {
            return response()->json([
                'error' => 'forbidden',
                'message' => "API key missing required scope: {$requiredScope}",
                'detail' => "{$method} requests require '{$requiredScope}' scope",
                'key_scopes' => $apiKey->scopes,
            ], 403);
        }

        return $next($request);
    }
}
