<?php

declare(strict_types=1);

namespace Mod\Api\Services;

use Illuminate\Support\Facades\Log;
use Mod\Api\Models\ApiKey;

/**
 * API Key Service - manages API key lifecycle.
 *
 * Provides methods for creating, rotating, and managing API keys
 * with proper validation and logging.
 */
class ApiKeyService
{
    /**
     * Create a new API key for a workspace.
     *
     * @return array{api_key: ApiKey, plain_key: string}
     */
    public function create(
        int $workspaceId,
        int $userId,
        string $name,
        array $scopes = [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE],
        ?\DateTimeInterface $expiresAt = null,
        ?array $serverScopes = null
    ): array {
        // Check workspace key limit
        $maxKeys = config('api.keys.max_per_workspace', 10);
        $currentCount = ApiKey::forWorkspace($workspaceId)->active()->count();

        if ($currentCount >= $maxKeys) {
            throw new \RuntimeException(
                "Workspace has reached the maximum number of API keys ({$maxKeys})"
            );
        }

        $result = ApiKey::generate($workspaceId, $userId, $name, $scopes, $expiresAt);

        // Set server scopes if provided
        if ($serverScopes !== null) {
            $result['api_key']->update(['server_scopes' => $serverScopes]);
        }

        Log::info('API key created', [
            'key_id' => $result['api_key']->id,
            'workspace_id' => $workspaceId,
            'user_id' => $userId,
            'name' => $name,
        ]);

        return $result;
    }

    /**
     * Rotate an existing API key.
     *
     * Creates a new key with the same settings, keeping the old key
     * valid for a grace period to allow migration.
     *
     * @param  int  $gracePeriodHours  Hours the old key remains valid (default: 24)
     * @return array{api_key: ApiKey, plain_key: string, old_key: ApiKey}
     */
    public function rotate(ApiKey $apiKey, int $gracePeriodHours = ApiKey::DEFAULT_GRACE_PERIOD_HOURS): array
    {
        // Don't rotate keys that are already being rotated out
        if ($apiKey->isInGracePeriod()) {
            throw new \RuntimeException(
                'This key is already being rotated. Wait for the grace period to end or end it manually.'
            );
        }

        // Don't rotate revoked keys
        if ($apiKey->trashed()) {
            throw new \RuntimeException('Cannot rotate a revoked key.');
        }

        $result = $apiKey->rotate($gracePeriodHours);

        Log::info('API key rotated', [
            'old_key_id' => $apiKey->id,
            'new_key_id' => $result['api_key']->id,
            'workspace_id' => $apiKey->workspace_id,
            'grace_period_hours' => $gracePeriodHours,
            'grace_period_ends_at' => $apiKey->fresh()->grace_period_ends_at?->toIso8601String(),
        ]);

        return $result;
    }

    /**
     * Revoke an API key immediately.
     */
    public function revoke(ApiKey $apiKey): void
    {
        $apiKey->revoke();

        Log::info('API key revoked', [
            'key_id' => $apiKey->id,
            'workspace_id' => $apiKey->workspace_id,
        ]);
    }

    /**
     * End the grace period for a rotating key and revoke it.
     */
    public function endGracePeriod(ApiKey $apiKey): void
    {
        if (! $apiKey->isInGracePeriod()) {
            throw new \RuntimeException('This key is not in a grace period.');
        }

        $apiKey->endGracePeriod();

        Log::info('API key grace period ended', [
            'key_id' => $apiKey->id,
            'workspace_id' => $apiKey->workspace_id,
        ]);
    }

    /**
     * Clean up keys with expired grace periods.
     *
     * This should be called by a scheduled command to revoke
     * old keys after their grace period has ended.
     *
     * @return int Number of keys cleaned up
     */
    public function cleanupExpiredGracePeriods(): int
    {
        $keys = ApiKey::gracePeriodExpired()
            ->whereNull('deleted_at')
            ->get();

        $count = 0;

        foreach ($keys as $key) {
            $key->revoke();
            $count++;

            Log::info('Cleaned up API key after grace period', [
                'key_id' => $key->id,
                'workspace_id' => $key->workspace_id,
            ]);
        }

        return $count;
    }

    /**
     * Update API key scopes.
     */
    public function updateScopes(ApiKey $apiKey, array $scopes): void
    {
        // Validate scopes
        $validScopes = array_intersect($scopes, ApiKey::ALL_SCOPES);

        if (empty($validScopes)) {
            throw new \InvalidArgumentException('At least one valid scope must be provided.');
        }

        $apiKey->update(['scopes' => array_values($validScopes)]);

        Log::info('API key scopes updated', [
            'key_id' => $apiKey->id,
            'scopes' => $validScopes,
        ]);
    }

    /**
     * Update API key server scopes.
     */
    public function updateServerScopes(ApiKey $apiKey, ?array $serverScopes): void
    {
        $apiKey->update(['server_scopes' => $serverScopes]);

        Log::info('API key server scopes updated', [
            'key_id' => $apiKey->id,
            'server_scopes' => $serverScopes,
        ]);
    }

    /**
     * Rename an API key.
     */
    public function rename(ApiKey $apiKey, string $name): void
    {
        $apiKey->update(['name' => $name]);

        Log::info('API key renamed', [
            'key_id' => $apiKey->id,
            'name' => $name,
        ]);
    }

    /**
     * Get statistics for a workspace's API keys.
     */
    public function getStats(int $workspaceId): array
    {
        $keys = ApiKey::forWorkspace($workspaceId);

        return [
            'total' => (clone $keys)->count(),
            'active' => (clone $keys)->active()->count(),
            'expired' => (clone $keys)->expired()->count(),
            'in_grace_period' => (clone $keys)->inGracePeriod()->count(),
            'revoked' => ApiKey::withTrashed()
                ->forWorkspace($workspaceId)
                ->whereNotNull('deleted_at')
                ->count(),
        ];
    }
}
