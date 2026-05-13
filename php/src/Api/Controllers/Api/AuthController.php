<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Front\Controller;
use Core\Tenant\Models\User;
use Core\Tenant\Models\UserToken;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\Hash;
use Illuminate\Validation\ValidationException;

/**
 * Authentication endpoints for dashboard sessions and bearer-token clients.
 */
class AuthController extends Controller
{
    use HasApiResponses;

    /**
     * Exchange credentials for a bearer token.
     *
     * Example:
     * POST /api/v1/auth/token {"email":"agent@example.com","password":"secret","token_name":"CLI"}
     */
    public function store(Request $request): JsonResponse
    {
        $data = $request->validate([
            'email' => ['required', 'string', 'email'],
            'password' => ['required', 'string'],
            'token_name' => ['sometimes', 'string', 'max:255'],
            'expires_at' => ['sometimes', 'nullable', 'date'],
        ]);

        /** @var User|null $user */
        $user = User::query()->where('email', $data['email'])->first();

        if (! $user instanceof User || ! Hash::check($data['password'], (string) $user->password)) {
            throw ValidationException::withMessages([
                'email' => ['The provided credentials are invalid.'],
            ]);
        }

        $plainTextToken = bin2hex(random_bytes(20));
        $accessToken = new UserToken([
            'name' => $data['token_name'] ?? 'API token',
            'token' => hash('sha256', $plainTextToken),
            'expires_at' => isset($data['expires_at']) && $data['expires_at'] !== null
                ? Carbon::parse($data['expires_at'])
                : null,
        ]);
        $accessToken->user()->associate($user);
        $accessToken->save();

        return response()->json([
            'token_type' => 'Bearer',
            'access_token' => $plainTextToken,
            'expires_at' => $accessToken->expires_at?->toIso8601String(),
            'user' => $this->userPayload($user),
        ], 201);
    }

    /**
     * Revoke the current bearer token or API key.
     *
     * Example:
     * DELETE /api/v1/auth/token with Authorization: Bearer <token>
     */
    public function destroy(Request $request): JsonResponse
    {
        $accessToken = $request->attributes->get('access_token');
        if ($accessToken instanceof UserToken) {
            $accessToken->delete();

            return $this->successResponse('Access token revoked.');
        }

        $apiKey = $request->attributes->get('api_key');
        if ($apiKey instanceof \Core\Api\Models\ApiKey) {
            $apiKey->revoke();

            return $this->successResponse('API key revoked.');
        }

        return $this->notFoundResponse('Access token');
    }

    /**
     * Show the currently authenticated user.
     *
     * Example:
     * GET /api/v1/auth/me
     */
    public function show(Request $request): JsonResponse
    {
        $user = $request->user();
        if (! $user instanceof User) {
            return $this->errorResponse('unauthorized', 'Authentication required.', status: 401);
        }

        return response()->json([
            'user' => $this->userPayload($user),
            'auth_type' => $request->attributes->get('auth_type', 'session'),
            'workspace_id' => $request->attributes->get('workspace_id'),
        ]);
    }

    /**
     * Build the stable user payload returned by auth endpoints.
     */
    protected function userPayload(User $user): array
    {
        return [
            'id' => $user->id,
            'name' => $user->name,
            'email' => $user->email,
            'default_workspace_id' => $user->defaultHostWorkspace()?->id,
        ];
    }
}
