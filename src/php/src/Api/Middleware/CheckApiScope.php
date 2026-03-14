<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Core\Api\Models\ApiKey;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Check that the API key has required scopes for the request.
 *
 * Usage in routes:
 * Route::middleware(['auth.api', 'api.scope:write'])->post('/resource', ...);
 * Route::middleware(['auth.api', 'api.scope:read,write'])->put('/resource', ...);
 *
 * Register in bootstrap/app.php:
 *   ->withMiddleware(function (Middleware $middleware) {
 *       $middleware->alias([
 *           'api.scope' => \App\Http\Middleware\Api\CheckApiScope::class,
 *       ]);
 *   })
 */
class CheckApiScope
{
    public function handle(Request $request, Closure $next, string ...$scopes): Response
    {
        $apiKey = $request->attributes->get('api_key');

        // If not authenticated via API key, allow through
        // (Sanctum auth handles its own scopes)
        if (! $apiKey instanceof ApiKey) {
            return $next($request);
        }

        // Check all required scopes
        foreach ($scopes as $scope) {
            if (! $apiKey->hasScope($scope)) {
                return response()->json([
                    'error' => 'forbidden',
                    'message' => "API key missing required scope: {$scope}",
                    'required_scopes' => $scopes,
                    'key_scopes' => $apiKey->scopes,
                ], 403);
            }
        }

        return $next($request);
    }
}
