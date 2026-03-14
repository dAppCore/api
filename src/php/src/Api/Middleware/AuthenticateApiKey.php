<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Core\Api\Models\ApiKey;
use Core\Api\Services\IpRestrictionService;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Authenticate requests using API keys or fall back to Sanctum.
 *
 * API keys are prefixed with 'hk_' and scoped to a workspace.
 *
 * Register in bootstrap/app.php:
 *   ->withMiddleware(function (Middleware $middleware) {
 *       $middleware->alias([
 *           'auth.api' => \App\Http\Middleware\Api\AuthenticateApiKey::class,
 *       ]);
 *   })
 */
class AuthenticateApiKey
{
    public function handle(Request $request, Closure $next, ?string $scope = null): Response
    {
        $token = $request->bearerToken();

        if (! $token) {
            return $this->unauthorized('API key required. Use Authorization: Bearer <api_key>');
        }

        // Check if it's an API key (prefixed with hk_)
        if (str_starts_with($token, 'hk_')) {
            return $this->authenticateApiKey($request, $next, $token, $scope);
        }

        // Fall back to Sanctum for OAuth tokens
        return $this->authenticateSanctum($request, $next, $scope);
    }

    /**
     * Authenticate using an API key.
     */
    protected function authenticateApiKey(
        Request $request,
        Closure $next,
        string $token,
        ?string $scope
    ): Response {
        $apiKey = ApiKey::findByPlainKey($token);

        if (! $apiKey) {
            return $this->unauthorized('Invalid API key');
        }

        if ($apiKey->isExpired()) {
            return $this->unauthorized('API key has expired');
        }

        // Check IP whitelist if restrictions are enabled
        if ($apiKey->hasIpRestrictions()) {
            $ipService = app(IpRestrictionService::class);
            $requestIp = $request->ip();

            if (! $ipService->isIpAllowed($requestIp, $apiKey->getAllowedIps() ?? [])) {
                return $this->forbidden('IP address not allowed for this API key');
            }
        }

        // Check scope if required
        if ($scope !== null && ! $apiKey->hasScope($scope)) {
            return $this->forbidden("API key missing required scope: {$scope}");
        }

        // Record usage (non-blocking)
        $apiKey->recordUsage();

        // Set request context
        $request->setUserResolver(fn () => $apiKey->user);
        $request->attributes->set('api_key', $apiKey);
        $request->attributes->set('workspace', $apiKey->workspace);
        $request->attributes->set('workspace_id', $apiKey->workspace_id);
        $request->attributes->set('auth_type', 'api_key');

        return $next($request);
    }

    /**
     * Fall back to Sanctum authentication for OAuth tokens.
     */
    protected function authenticateSanctum(
        Request $request,
        Closure $next,
        ?string $scope
    ): Response {
        // For API requests, use token authentication
        if (! $request->user()) {
            // Try to authenticate via Sanctum token
            $guard = auth('sanctum');
            if (! $guard->check()) {
                return $this->unauthorized('Invalid authentication token');
            }

            $request->setUserResolver(fn () => $guard->user());
        }

        $request->attributes->set('auth_type', 'sanctum');

        return $next($request);
    }

    /**
     * Return 401 Unauthorized response.
     */
    protected function unauthorized(string $message): Response
    {
        return response()->json([
            'error' => 'unauthorized',
            'message' => $message,
        ], 401);
    }

    /**
     * Return 403 Forbidden response.
     */
    protected function forbidden(string $message): Response
    {
        return response()->json([
            'error' => 'forbidden',
            'message' => $message,
        ], 403);
    }
}
