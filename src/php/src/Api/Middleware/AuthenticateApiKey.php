<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Core\Api\Models\ApiKey;
use Core\Api\Services\IpRestrictionService;
use Core\Api\Concerns\HasApiResponses;
use Core\Tenant\Models\UserToken;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Authenticate requests using API keys or fall back to Sanctum.
 *
 * API keys are underscore-delimited and scoped to a workspace.
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
    use HasApiResponses;

    public function handle(Request $request, Closure $next, ?string $scope = null): Response
    {
        $token = $request->bearerToken();

        // If no bearer token is present, allow Sanctum/session auth to try first.
        // This keeps browser-based dashboard requests working while preserving
        // the API-key path for SDK and server-to-server calls.
        if (! $token) {
            return $this->authenticateSanctum($request, $next, $scope);
        }

        // API keys are underscore-delimited; if a token looks like an API key,
        // it must validate as one rather than falling through to Sanctum.
        if (str_contains($token, '_')) {
            $apiKey = ApiKey::findByPlainKey($token);
            if ($apiKey instanceof ApiKey) {
                return $this->authenticateResolvedApiKey($request, $next, $apiKey, $scope);
            }

            return $this->unauthorized('Invalid API key');
        }

        // Fall back to Sanctum for OAuth tokens
        return $this->authenticateSanctum($request, $next, $scope);
    }

    /**
     * Authenticate using an API key.
     */
    protected function authenticateResolvedApiKey(
        Request $request,
        Closure $next,
        ApiKey $apiKey,
        ?string $scope
    ): Response {
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
        $request->attributes->set('principal', 'api-key:'.$apiKey->id);
        $request->attributes->set('userID', (string) $apiKey->user_id);

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
        $bearerToken = $request->bearerToken();

        if (is_string($bearerToken) && $bearerToken !== '') {
            $accessToken = UserToken::findToken($bearerToken);

            if ($accessToken instanceof UserToken && $accessToken->isValid()) {
                if ($scope !== null) {
                    return $this->forbidden("Token missing required scope: {$scope}");
                }

                $accessToken->recordUsage();

                $request->setUserResolver(fn () => $accessToken->user);
                $request->attributes->set('auth_type', 'access_token');
                $request->attributes->set('access_token', $accessToken);
                $request->attributes->set('principal', 'user-token:'.$accessToken->id);
                $request->attributes->set('userID', (string) $accessToken->user_id);

                return $next($request);
            }
        }

        // For API requests, use token authentication
        if (! $request->user()) {
            // Try to authenticate via Sanctum token
            $guard = auth('sanctum');
            if (! $guard->check()) {
                return $this->unauthorized('Invalid authentication token');
            }

            $request->setUserResolver(fn () => $guard->user());
        }

        $authType = $bearerToken ? 'sanctum' : 'session';
        $user = $request->user();

        $request->attributes->set('auth_type', $authType);
        if ($user !== null && isset($user->id)) {
            $request->attributes->set('principal', 'user:'.$user->id);
            $request->attributes->set('userID', (string) $user->id);
        }

        return $next($request);
    }

    /**
     * Return 401 Unauthorised response.
     */
    protected function unauthorized(string $message): Response
    {
        return $this->errorResponse(
            errorCode: 'unauthorized',
            message: $message,
            status: 401,
        );
    }

    /**
     * Return 403 Forbidden response.
     */
    protected function forbidden(string $message): Response
    {
        return $this->forbiddenResponse($message, status: 403);
    }
}
