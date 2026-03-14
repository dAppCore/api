<?php

declare(strict_types=1);

namespace Core\Api\Guards;

use Core\Tenant\Models\User;
use Core\Tenant\Models\UserToken;
use Illuminate\Contracts\Auth\Factory;
use Illuminate\Http\Request;

/**
 * Custom authentication guard for API token-based authentication.
 *
 * This guard authenticates users via Bearer tokens sent in the Authorization header.
 * It's designed to work with Laravel's auth middleware system and provides
 * stateful API authentication using long-lived personal access tokens.
 *
 * Usage:
 *   Route::middleware('auth:access_token')->group(function () {
 *       // Protected API routes
 *   });
 */
class AccessTokenGuard
{
    /**
     * The authentication factory instance.
     */
    protected Factory $auth;

    /**
     * Create a new guard instance.
     */
    public function __construct(Factory $auth)
    {
        $this->auth = $auth;
    }

    /**
     * Handle the authentication for the incoming request.
     *
     * This method is called by Laravel's authentication system when using
     * the guard. It attempts to authenticate the request using the Bearer
     * token and returns the authenticated user if successful.
     *
     * @return User|null The authenticated user or null if authentication fails
     */
    public function __invoke(Request $request): ?User
    {
        $token = $this->getTokenFromRequest($request);

        if (! $token) {
            return null;
        }

        $accessToken = UserToken::findToken($token);

        if (! $this->isValidAccessToken($accessToken)) {
            return null;
        }

        // Update last used timestamp
        $accessToken->recordUsage();

        return $accessToken->user;
    }

    /**
     * Extract the Bearer token from the request.
     *
     * Looks for the token in the Authorization header in the format:
     * Authorization: Bearer {token}
     *
     * @return string|null The extracted token or null if not found
     */
    protected function getTokenFromRequest(Request $request): ?string
    {
        $token = $request->bearerToken();

        return ! empty($token) ? $token : null;
    }

    /**
     * Validate the access token.
     *
     * Checks if the token exists and hasn't expired.
     *
     * @return bool True if the token is valid, false otherwise
     */
    protected function isValidAccessToken(?UserToken $accessToken): bool
    {
        if (! $accessToken) {
            return false;
        }

        return $accessToken->isValid();
    }
}
