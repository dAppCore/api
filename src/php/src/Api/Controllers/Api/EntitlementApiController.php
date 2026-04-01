<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Models\ApiKey;
use Core\Api\Services\ApiUsageService;
use Core\Front\Controller;
use Core\Tenant\Models\Workspace;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Entitlements API controller.
 *
 * Returns the current workspace's plan limits and usage snapshot.
 */
class EntitlementApiController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;

    public function __construct(
        protected ApiUsageService $usageService
    ) {
    }

    /**
     * Show the current workspace entitlements.
     *
     * GET /api/entitlements
     */
    public function show(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);

        if (! $workspace instanceof Workspace) {
            return $this->noWorkspaceResponse();
        }

        $apiKey = $request->attributes->get('api_key');
        $authType = $request->attributes->get('auth_type', 'session');
        $rateLimitProfile = $this->resolveRateLimitProfile($authType);
        $activeApiKeys = ApiKey::query()
            ->forWorkspace($workspace->id)
            ->active()
            ->count();

        $usage = $this->usageService->getWorkspaceSummary($workspace->id);

        return response()->json([
            'workspace_id' => $workspace->id,
            'workspace' => [
                'id' => $workspace->id,
                'name' => $workspace->name ?? null,
            ],
            'authentication' => [
                'type' => $authType,
                'scopes' => $apiKey instanceof ApiKey ? $apiKey->scopes : null,
            ],
            'limits' => [
                'rate_limit' => $rateLimitProfile,
                'api_keys' => [
                    'active' => $activeApiKeys,
                    'maximum' => (int) config('api.keys.max_per_workspace', 10),
                    'remaining' => max(0, (int) config('api.keys.max_per_workspace', 10) - $activeApiKeys),
                ],
                'webhooks' => [
                    'maximum' => (int) config('api.webhooks.max_per_workspace', 5),
                ],
            ],
            'usage' => $usage,
            'features' => [
                'pixel' => true,
                'mcp' => true,
                'webhooks' => true,
                'usage_alerts' => (bool) config('api.alerts.enabled', true),
            ],
        ]);
    }

    /**
     * Resolve the rate limit profile for the current auth context.
     */
    protected function resolveRateLimitProfile(string $authType): array
    {
        $rateLimits = (array) config('api.rate_limits', []);
        $key = $authType === 'session' ? 'default' : 'authenticated';
        $profile = (array) ($rateLimits[$key] ?? []);

        return [
            'name' => $key,
            'limit' => (int) ($profile['limit'] ?? 0),
            'window' => (int) ($profile['window'] ?? 60),
            'burst' => (float) ($profile['burst'] ?? 1.0),
        ];
    }
}
