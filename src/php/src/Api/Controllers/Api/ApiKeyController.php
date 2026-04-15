<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Models\ApiKey;
use Core\Api\Resources\ApiKeyResource;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;
use Illuminate\Validation\Rule;
use Mod\Api\Services\ApiKeyService;

/**
 * API key management endpoints.
 */
class ApiKeyController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;

    public function __construct(protected ApiKeyService $service)
    {
    }

    /**
     * List API keys for the current workspace.
     *
     * GET /api/v1/api-keys
     */
    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $keys = ApiKey::query()
            ->forWorkspace($workspace->id)
            ->latest()
            ->get();

        return response()->json([
            'workspace_id' => $workspace->id,
            'api_keys' => $keys->map(fn (ApiKey $key) => (new ApiKeyResource($key))->toArray($request))->values()->all(),
            'count' => $keys->count(),
        ]);
    }

    /**
     * Create an API key for the current workspace.
     *
     * POST /api/v1/api-keys
     */
    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'scopes' => ['sometimes', 'array'],
            'scopes.*' => ['string', Rule::in(ApiKey::ALL_SCOPES)],
            'expires_at' => ['sometimes', 'nullable', 'date'],
            'server_scopes' => ['sometimes', 'nullable', 'array'],
            'server_scopes.*' => ['string', 'max:64', 'regex:'.ApiKey::SERVER_ID_PATTERN],
        ]);

        $result = $this->service->create(
            workspaceId: $workspace->id,
            userId: (int) $request->user()->id,
            name: $data['name'],
            scopes: $data['scopes'] ?? [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE],
            expiresAt: isset($data['expires_at']) && $data['expires_at'] !== null
                ? Carbon::parse($data['expires_at'])
                : null,
            serverScopes: $data['server_scopes'] ?? null,
        );

        return response()->json([
            'api_key' => ApiKeyResource::withPlainKey($result['api_key'], $result['plain_key'])->toArray($request),
        ], 201);
    }

    /**
     * Revoke an API key.
     *
     * DELETE /api/v1/api-keys/{id}
     */
    public function destroy(Request $request, string $id): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $apiKey = ApiKey::query()
            ->forWorkspace($workspace->id)
            ->find($id);

        if (! $apiKey instanceof ApiKey) {
            return $this->notFoundResponse('API key');
        }

        $this->service->revoke($apiKey);

        return $this->successResponse('API key revoked successfully.');
    }
}
